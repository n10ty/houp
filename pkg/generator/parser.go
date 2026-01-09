package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ParsePackage parses all Go files in the given directory
func ParsePackage(pkgPath string) (*PackageInfo, error) {
	// Load package with type information
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: pkgPath,
	}

	// Use pattern "." to load the package in the current directory
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found at %s", pkgPath)
	}

	if len(pkgs) > 1 {
		return nil, fmt.Errorf("multiple packages found at %s", pkgPath)
	}

	pkg := pkgs[0]
	// Allow type errors during generation - this is expected when generating for the first time
	// Only fail on syntax errors
	if len(pkg.Errors) > 0 {
		// Check if all errors are type errors (which is ok during initial generation)
		hasNonTypeErrors := false
		for _, err := range pkg.Errors {
			// Type errors typically contain phrases like "undefined", "has no field or method"
			// Syntax errors contain phrases like "syntax error", "expected", etc.
			// Module errors contain "outside main module" which can be ignored
			// Go version errors contain "requires newer Go version" which can be ignored
			errStr := err.Error()
			if !strings.Contains(errStr, "undefined") &&
				!strings.Contains(errStr, "has no field or method") &&
				!strings.Contains(errStr, "not used") &&
				!strings.Contains(errStr, "outside main module") &&
				!strings.Contains(errStr, "requires newer Go version") {
				hasNonTypeErrors = true
				break
			}
		}
		if hasNonTypeErrors {
			return nil, fmt.Errorf("package has errors: %v", pkg.Errors)
		}
		// Continue with type errors - they'll be fixed after generation
	}

	pkgInfo := &PackageInfo{
		Name:      pkg.Name,
		Path:      pkgPath,
		PkgPath:   pkg.PkgPath,
		Files:     make(map[string]*FileInfo),
		TypesInfo: pkg.TypesInfo,
	}

	// Parse each file
	fset := token.NewFileSet()
	for i, astFile := range pkg.Syntax {
		var filename string
		if i < len(pkg.GoFiles) {
			filename = pkg.GoFiles[i]
		} else if i < len(pkg.CompiledGoFiles) {
			filename = pkg.CompiledGoFiles[i]
		} else {
			// Fallback to file position if available
			filename = pkg.Fset.File(astFile.Pos()).Name()
		}

		// Re-parse the file with ParseComments to ensure we get doc comments
		astFileWithComments, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			// If re-parsing fails, use the original AST (without comments)
			astFileWithComments = astFile
		}

		fileInfo := &FileInfo{
			Name:    filepath.Base(filename),
			Path:    filename,
			AST:     astFileWithComments,
			Structs: []*StructInfo{},
			Skip:    hasFileSkipAnnotation(astFileWithComments),
		}

		// Extract structs from this file
		// Use file.Decls directly to preserve Doc comments
		// First, collect all type declaration positions for skip annotation detection
		var typeGenDeclPositions []token.Pos
		for _, decl := range astFileWithComments.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if ok && genDecl.Tok == token.TYPE {
				typeGenDeclPositions = append(typeGenDeclPositions, genDecl.Pos())
			}
		}

		declIndex := 0
		for _, decl := range astFileWithComments.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// Doc comments can be on either GenDecl or TypeSpec
				// If there's only one spec in the GenDecl, the comment is on GenDecl
				// If there are multiple specs, each TypeSpec has its own Doc
				if typeSpec.Doc == nil && len(genDecl.Specs) == 1 {
					typeSpec.Doc = genDecl.Doc
				}

				// Determine the range to check for skip annotations
				var prevDeclPos token.Pos = 0
				if declIndex > 0 {
					prevDeclPos = typeGenDeclPositions[declIndex-1]
				}

				structInfo := parseStruct(typeSpec, structType, filename, pkg.TypesInfo, genDecl, astFileWithComments.Comments, prevDeclPos)
				if structInfo != nil {
					fileInfo.Structs = append(fileInfo.Structs, structInfo)
				}
			}
			declIndex++
		}

		pkgInfo.Files[fileInfo.Name] = fileInfo
	}

	// Check if we actually found any files
	if len(pkgInfo.Files) == 0 {
		return nil, fmt.Errorf("no Go files found in package %s", pkgPath)
	}

	// Discover structs referenced by 'dive' tags and mark them for generation
	// This ensures that structs without validation tags but referenced by dive
	// will get empty Validate() methods generated
	discoverAndMarkDiveStructs(pkgInfo)

	return pkgInfo, nil
}

// parseStruct extracts struct information including fields and validation tags
func parseStruct(typeSpec *ast.TypeSpec, structType *ast.StructType, filename string, typesInfo *types.Info, genDecl *ast.GenDecl, fileComments []*ast.CommentGroup, prevDeclPos token.Pos) *StructInfo {
	structInfo := &StructInfo{
		Name:             typeSpec.Name.Name,
		TypeSpec:         typeSpec,
		Fields:           []*FieldInfo{},
		NeedsGen:         false,
		SourceFile:       filepath.Base(filename),
		CustomValidators: []CustomValidator{},
		Skip:             hasStructSkipAnnotation(typeSpec, genDecl, fileComments, prevDeclPos),
	}

	// Parse struct-level validation comments
	if typeSpec.Doc != nil {
		for _, comment := range typeSpec.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			// Look for //validate:pkg/path:FuncName
			if strings.HasPrefix(text, "validate:") {
				validatorStr := strings.TrimPrefix(text, "validate:")
				validatorStr = strings.TrimSpace(validatorStr)

				// Parse the validator: should be in format pkg/path:FuncName
				if validator, err := parseStructValidator(validatorStr); err == nil {
					structInfo.CustomValidators = append(structInfo.CustomValidators, validator)
					structInfo.NeedsGen = true
				}
			}
		}
	}

	if structType.Fields == nil {
		return structInfo
	}

	for _, field := range structType.Fields.List {
		// Skip embedded fields for now
		if len(field.Names) == 0 {
			continue
		}

		fieldName := field.Names[0].Name

		// Skip unexported fields
		if !ast.IsExported(fieldName) {
			continue
		}

		var tag string
		if field.Tag != nil {
			tag = field.Tag.Value
			// Remove backticks
			tag = strings.Trim(tag, "`")
		}

		// Parse validation tag
		validateTag := extractTag(tag, "validate")
		if validateTag == "" {
			continue // No validation for this field
		}

		fieldInfo := &FieldInfo{
			Name:       fieldName,
			Type:       field.Type,
			TypeString: types.ExprString(field.Type),
			Tag:        tag,
			JSONName:   extractTag(tag, "json"),
		}

		// Parse validation rules
		rules, err := parseValidationRules(validateTag)
		if err != nil {
			// Skip fields with invalid validation tags for now
			// Will be handled by validator
			continue
		}

		fieldInfo.Rules = rules
		structInfo.Fields = append(structInfo.Fields, fieldInfo)
		structInfo.NeedsGen = true
	}

	return structInfo
}

// extractTag extracts a specific tag value from struct tag
func extractTag(tag, key string) string {
	structTag := reflect.StructTag(tag)
	return structTag.Get(key)
}

// parseValidationRules parses the validation tag into individual rules
func parseValidationRules(validateTag string) ([]ValidationRule, error) {
	if validateTag == "" {
		return nil, nil
	}

	parts := strings.Split(validateTag, ",")
	rules := make([]ValidationRule, 0, len(parts))

	// Find the index of 'dive' if present
	diveIndex := -1
	for i, part := range parts {
		if strings.TrimSpace(part) == "dive" {
			diveIndex = i
			break
		}
	}

	// If dive is found, split rules into pre-dive and post-dive
	if diveIndex >= 0 {
		// Parse pre-dive rules
		for i := 0; i < diveIndex; i++ {
			part := strings.TrimSpace(parts[i])
			if part == "" {
				continue
			}

			rule, err := parseValidationRule(part)
			if err != nil {
				return nil, err
			}
			rules = append(rules, rule)
		}

		// Parse post-dive rules (rules that apply to each element)
		var elementRules []ValidationRule
		for i := diveIndex + 1; i < len(parts); i++ {
			part := strings.TrimSpace(parts[i])
			if part == "" {
				continue
			}

			rule, err := parseValidationRule(part)
			if err != nil {
				return nil, err
			}
			elementRules = append(elementRules, rule)
		}

		// Add the dive rule with element rules
		rules = append(rules, &DiveRule{ElementRules: elementRules})

		return rules, nil
	}

	// No dive tag, parse all rules normally
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		rule, err := parseValidationRule(part)
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// parseValidationRule parses a single validation rule string
func parseValidationRule(ruleStr string) (ValidationRule, error) {
	// Check if it contains '=' for parameterized rules
	parts := strings.SplitN(ruleStr, "=", 2)
	ruleName := parts[0]
	var param string
	if len(parts) == 2 {
		param = parts[1]
	}

	switch ruleName {
	case "required":
		return &RequiredRule{}, nil
	case "required_without":
		if param == "" {
			return nil, fmt.Errorf("required_without rule requires a field name parameter")
		}
		return &RequiredWithoutRule{OtherField: param}, nil
	case "eqfield":
		if param == "" {
			return nil, fmt.Errorf("eqfield rule requires a field name parameter")
		}
		return &EqFieldRule{OtherField: param}, nil
	case "omitempty":
		return &OmitEmptyRule{}, nil
	case "min":
		return &MinRule{Value: param}, nil
	case "max":
		return &MaxRule{Value: param}, nil
	case "gt":
		return &GTRule{Value: param}, nil
	case "lt":
		return &LTRule{Value: param}, nil
	case "gte":
		return &GTERule{Value: param}, nil
	case "lte":
		return &LTERule{Value: param}, nil
	case "regexp":
		return parseRegexpRule(param)
	case "unique":
		if param == "" {
			return &UniqueRule{}, nil
		}
		return &UniqueRule{FieldName: param}, nil
	case "dive":
		return &DiveRule{}, nil
	case "datetime":
		if param == "" {
			return nil, fmt.Errorf("datetime rule requires a format parameter")
		}
		return &DateTimeRule{Format: param}, nil
	case "uuid":
		return &UUIDRule{}, nil
	case "iso4217":
		return &ISO4217Rule{}, nil
	case "email":
		return &EmailRule{}, nil
	case "iso3166_1_alpha2":
		return &ISO3166_1_Alpha2Rule{}, nil
	default:
		// Check if it's a custom validator (contains ':')
		if strings.Contains(ruleStr, ":") {
			return parseCustomRule(ruleStr)
		}
		return &UnknownRule{Raw: ruleStr}, nil
	}
}

// parseRegexpRule parses regexp=pkg/path:VarName
func parseRegexpRule(param string) (ValidationRule, error) {
	if param == "" {
		return nil, fmt.Errorf("regexp rule requires parameter in format pkg/path:VarName")
	}

	parts := strings.SplitN(param, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("regexp rule must be in format pkg/path:VarName, got: %s", param)
	}

	return &RegexpRule{
		ImportPath: parts[0],
		VarName:    parts[1],
	}, nil
}

// parseCustomRule parses custom validator in format pkg/path:FuncName
func parseCustomRule(ruleStr string) (ValidationRule, error) {
	parts := strings.SplitN(ruleStr, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("custom rule must be in format pkg/path:FuncName, got: %s", ruleStr)
	}

	return &CustomRule{
		ImportPath: parts[0],
		FuncName:   parts[1],
	}, nil
}

// parseStructValidator parses struct-level validator in two formats:
// 1. pkg/path:FuncName - for validators in external packages
// 2. FuncName - for validators in the same package (no import needed)
func parseStructValidator(validatorStr string) (CustomValidator, error) {
	// Check if it contains a colon (indicating package:function format)
	if strings.Contains(validatorStr, ":") {
		// Format: pkg/path:FuncName
		parts := strings.SplitN(validatorStr, ":", 2)
		if len(parts) != 2 {
			return CustomValidator{}, fmt.Errorf("struct validator must be in format pkg/path:FuncName or just FuncName, got: %s", validatorStr)
		}

		return CustomValidator{
			ImportPath: parts[0],
			FuncName:   parts[1],
		}, nil
	}

	// Format: FuncName (same package)
	// Use empty string as ImportPath to indicate same package
	return CustomValidator{
		ImportPath: "",
		FuncName:   validatorStr,
	}, nil
}

// ResolveTypeInfo resolves type information from an AST expression
func ResolveTypeInfo(expr ast.Expr, typesInfo *types.Info) TypeInfo {
	typeInfo := TypeInfo{
		Kind:         TypeUnknown,
		UnderlyingGo: expr,
	}

	switch t := expr.(type) {
	case *ast.Ident:
		// Built-in or named type
		typeInfo.Name = t.Name
		typeInfo.Kind = getTypeKind(t.Name)

		// If it's an unknown type and we have type info, check the underlying type
		if typeInfo.Kind == TypeUnknown && typesInfo != nil {
			if obj := typesInfo.Uses[t]; obj != nil {
				if typeName, ok := obj.(*types.TypeName); ok {
					underlying := typeName.Type().Underlying()
					if basic, ok := underlying.(*types.Basic); ok {
						typeInfo.Kind = getTypeKindFromBasic(basic.Kind())
					}
				}
			}
		}

	case *ast.StarExpr:
		// Pointer type
		typeInfo.IsPointer = true
		typeInfo.Kind = TypePointer
		elem := ResolveTypeInfo(t.X, typesInfo)
		typeInfo.Elem = &elem

	case *ast.ArrayType:
		// Slice or array
		if t.Len == nil {
			// Slice
			typeInfo.IsSlice = true
			typeInfo.Kind = TypeSlice
		} else {
			typeInfo.Kind = TypeArray
		}
		elem := ResolveTypeInfo(t.Elt, typesInfo)
		typeInfo.Elem = &elem

	case *ast.StructType:
		typeInfo.Kind = TypeStruct
		typeInfo.IsStruct = true

	case *ast.SelectorExpr:
		// Qualified type from another package (e.g., pkg.Type)
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			typeInfo.PkgName = pkgIdent.Name
			typeInfo.Name = t.Sel.Name

			// Check if this is json.Number
			if typeInfo.PkgName == "json" && typeInfo.Name == "Number" {
				typeInfo.Kind = TypeJSONNumber
			} else {
				typeInfo.Kind = TypeStruct // Assume struct for now
			}

			// Try to get import path from types.Info
			if typesInfo != nil {
				if obj := typesInfo.Uses[pkgIdent]; obj != nil {
					if pkgName, ok := obj.(*types.PkgName); ok {
						typeInfo.PkgPath = pkgName.Imported().Path()
						// Double-check if it's json.Number via import path
						if typeInfo.PkgPath == "encoding/json" && typeInfo.Name == "Number" {
							typeInfo.Kind = TypeJSONNumber
						}
					}
				}
			}
		}

	case *ast.MapType:
		typeInfo.Kind = TypeMap

	case *ast.InterfaceType:
		typeInfo.Kind = TypeInterface
	}

	return typeInfo
}

// getTypeKind returns the TypeKind for a built-in type name
func getTypeKind(name string) TypeKind {
	switch name {
	case "bool":
		return TypeBool
	case "int":
		return TypeInt
	case "int8":
		return TypeInt8
	case "int16":
		return TypeInt16
	case "int32":
		return TypeInt32
	case "int64":
		return TypeInt64
	case "uint":
		return TypeUint
	case "uint8", "byte":
		return TypeUint8
	case "uint16":
		return TypeUint16
	case "uint32":
		return TypeUint32
	case "uint64":
		return TypeUint64
	case "float32":
		return TypeFloat32
	case "float64":
		return TypeFloat64
	case "string":
		return TypeString
	default:
		return TypeUnknown
	}
}

// getTypeKindFromBasic converts types.BasicKind to TypeKind
func getTypeKindFromBasic(kind types.BasicKind) TypeKind {
	switch kind {
	case types.Bool:
		return TypeBool
	case types.Int:
		return TypeInt
	case types.Int8:
		return TypeInt8
	case types.Int16:
		return TypeInt16
	case types.Int32:
		return TypeInt32
	case types.Int64:
		return TypeInt64
	case types.Uint:
		return TypeUint
	case types.Uint8:
		return TypeUint8
	case types.Uint16:
		return TypeUint16
	case types.Uint32:
		return TypeUint32
	case types.Uint64:
		return TypeUint64
	case types.Float32:
		return TypeFloat32
	case types.Float64:
		return TypeFloat64
	case types.String:
		return TypeString
	default:
		return TypeUnknown
	}
}

// ParseFile parses a single Go file (useful for testing)
func ParseFile(filename string) (*FileInfo, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	fileInfo := &FileInfo{
		Name:    filepath.Base(filename),
		Path:    filename,
		AST:     astFile,
		Structs: []*StructInfo{},
		Skip:    hasFileSkipAnnotation(astFile),
	}

	// Extract structs - use file.Decls to get GenDecl for skip annotation detection
	// First, collect all type declaration positions
	var typeGenDeclPositions []token.Pos
	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if ok && genDecl.Tok != token.TYPE {
			typeGenDeclPositions = append(typeGenDeclPositions, genDecl.Pos())
		}
	}

	declIndex := 0
	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Doc comments can be on either GenDecl or TypeSpec
			if typeSpec.Doc == nil && len(genDecl.Specs) == 1 {
				typeSpec.Doc = genDecl.Doc
			}

			// Determine the range to check for skip annotations
			var prevDeclPos token.Pos = 0
			if declIndex > 0 {
				prevDeclPos = typeGenDeclPositions[declIndex-1]
			}

			structInfo := parseStruct(typeSpec, structType, filename, nil, genDecl, astFile.Comments, prevDeclPos)
			if structInfo != nil {
				fileInfo.Structs = append(fileInfo.Structs, structInfo)
			}
		}
		declIndex++
	}

	return fileInfo, nil
}

// discoverAndMarkDiveStructs finds all structs referenced by 'dive' tags
// and marks them as NeedsGen even if they don't have their own validation tags.
// This ensures empty Validate() methods are generated for them.
func discoverAndMarkDiveStructs(pkgInfo *PackageInfo) {
	// Build a map of all struct names to StructInfo
	allStructs := make(map[string]*StructInfo)
	for _, fileInfo := range pkgInfo.Files {
		for _, structInfo := range fileInfo.Structs {
			allStructs[structInfo.Name] = structInfo
		}
	}

	// Find all structs referenced by dive tags
	referencedStructs := make(map[string]bool)
	for _, fileInfo := range pkgInfo.Files {
		for _, structInfo := range fileInfo.Structs {
			for _, field := range structInfo.Fields {
				for _, rule := range field.Rules {
					if _, ok := rule.(*DiveRule); ok {
						// Extract type name from field
						typeInfo := ResolveTypeInfo(field.Type, pkgInfo.TypesInfo)

						var typeName string
						if typeInfo.IsPointer && typeInfo.Elem != nil {
							typeName = typeInfo.Elem.Name
						} else if typeInfo.IsSlice && typeInfo.Elem != nil {
							if typeInfo.Elem.IsPointer && typeInfo.Elem.Elem != nil {
								typeName = typeInfo.Elem.Elem.Name
							} else {
								typeName = typeInfo.Elem.Name
							}
						} else {
							typeName = typeInfo.Name
						}

						if typeName != "" {
							referencedStructs[typeName] = true
						}
					}
				}
			}
		}
	}

	// Mark referenced structs as needing generation
	for typeName := range referencedStructs {
		if structInfo, exists := allStructs[typeName]; exists {
			structInfo.NeedsGen = true
		}
	}
}

// hasFileSkipAnnotation checks if a file has //validate:skip annotation in the package comments
func hasFileSkipAnnotation(file *ast.File) bool {
	// Check File.Doc first (comments directly attached to package declaration)
	if file.Doc != nil {
		for _, comment := range file.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if text == "validate:skip" {
				return true
			}
		}
	}

	// Also check all comments in the file that appear before the package declaration
	// These would be in File.Comments but not File.Doc
	if file.Comments != nil {
		for _, commentGroup := range file.Comments {
			// Only check comments that appear before the package declaration
			if commentGroup.Pos() < file.Package {
				for _, comment := range commentGroup.List {
					text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
					if text == "validate:skip" {
						return true
					}
				}
			}
		}
	}

	return false
}

// hasStructSkipAnnotation checks if a struct has //validate:skip annotation in its doc comments
func hasStructSkipAnnotation(typeSpec *ast.TypeSpec, genDecl *ast.GenDecl, fileComments []*ast.CommentGroup, prevDeclPos token.Pos) bool {
	// Check TypeSpec.Doc first
	if typeSpec.Doc != nil {
		for _, comment := range typeSpec.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if text == "validate:skip" {
				return true
			}
		}
	}

	// If there's only one spec in GenDecl, also check GenDecl.Doc
	if len(genDecl.Specs) == 1 && genDecl.Doc != nil {
		for _, comment := range genDecl.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if text == "validate:skip" {
				return true
			}
		}
	}

	// Check all comment groups that appear between the previous declaration and this one
	// This handles cases where //validate:skip is separated by blank lines from the struct
	if fileComments != nil {
		genDeclPos := genDecl.Pos()

		// Find any comment group between prevDeclPos and genDeclPos that contains //validate:skip
		for _, commentGroup := range fileComments {
			if commentGroup.Pos() > prevDeclPos && commentGroup.End() < genDeclPos {
				for _, comment := range commentGroup.List {
					text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
					if text == "validate:skip" {
						return true
					}
				}
			}
		}
	}

	return false
}
