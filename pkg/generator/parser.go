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
	}

	pkgs, err := packages.Load(cfg, pkgPath)
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
	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("package has errors: %v", pkg.Errors)
	}

	pkgInfo := &PackageInfo{
		Name:      pkg.Name,
		Path:      pkgPath,
		Files:     make(map[string]*FileInfo),
		TypesInfo: pkg.TypesInfo,
	}

	// Parse each file
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

		fileInfo := &FileInfo{
			Name:    filepath.Base(filename),
			Path:    filename,
			AST:     astFile,
			Structs: []*StructInfo{},
		}

		// Extract structs from this file
		ast.Inspect(astFile, func(n ast.Node) bool {
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return true
			}

			structInfo := parseStruct(typeSpec, structType, filename, pkg.TypesInfo)
			if structInfo != nil {
				fileInfo.Structs = append(fileInfo.Structs, structInfo)
			}

			return true
		})

		pkgInfo.Files[fileInfo.Name] = fileInfo
	}

	return pkgInfo, nil
}

// parseStruct extracts struct information including fields and validation tags
func parseStruct(typeSpec *ast.TypeSpec, structType *ast.StructType, filename string, typesInfo *types.Info) *StructInfo {
	structInfo := &StructInfo{
		Name:       typeSpec.Name.Name,
		TypeSpec:   typeSpec,
		Fields:     []*FieldInfo{},
		NeedsGen:   false,
		SourceFile: filepath.Base(filename),
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
	}

	// Extract structs
	ast.Inspect(astFile, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structInfo := parseStruct(typeSpec, structType, filename, nil)
		if structInfo != nil {
			fileInfo.Structs = append(fileInfo.Structs, structInfo)
		}

		return true
	})

	return fileInfo, nil
}
