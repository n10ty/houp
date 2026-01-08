package generator

import (
	"go/ast"
	"go/types"
)

// GenerateOptions contains configuration for the generator
type GenerateOptions struct {
	// Suffix for generated files (default: "_validate")
	Suffix string

	// Whether to collect all validation errors or return on first error
	MultiError bool

	// Whether to overwrite existing files
	Overwrite bool

	// DryRun mode - don't write files, just report what would be generated
	DryRun bool

	// UnknownTagMode determines behavior when unknown validation tags are encountered
	// "fail" - exit with error (default)
	// "skip" - log warning and continue
	UnknownTagMode string
}

// PackageInfo represents a parsed Go package
type PackageInfo struct {
	Name      string
	Path      string
	Files     map[string]*FileInfo // filename -> FileInfo
	TypesInfo *types.Info
}

// FileInfo represents a single Go source file
type FileInfo struct {
	Name    string
	Path    string
	AST     *ast.File
	Structs []*StructInfo
}

// StructInfo represents a struct with validation requirements
type StructInfo struct {
	Name       string
	TypeSpec   *ast.TypeSpec
	Fields     []*FieldInfo
	NeedsGen   bool // true if any field has validation tags
	SourceFile string
}

// FieldInfo represents a struct field with validation metadata
type FieldInfo struct {
	Name       string
	Type       ast.Expr
	TypeString string // string representation of the type
	Tag        string // full struct tag
	Rules      []ValidationRule
	JSONName   string // extracted from json tag
}

// ValidationRule represents a single validation constraint
type ValidationRule interface {
	// Name returns the rule name (e.g., "required", "min", "max")
	Name() string

	// Validate checks if this rule is applicable to the given field type
	// Returns error if rule cannot be applied to this type
	Validate(fieldType TypeInfo) error

	// Generate produces the validation code for this rule
	Generate(ctx *CodeGenContext, field *FieldInfo) (string, error)
}

// TypeInfo provides information about a field's type
type TypeInfo struct {
	Kind         TypeKind
	Name         string    // type name
	Elem         *TypeInfo // for pointers, slices, arrays
	PkgPath      string    // import path for named types from other packages
	PkgName      string    // package name for imports
	IsPointer    bool
	IsSlice      bool
	IsStruct     bool
	UnderlyingGo ast.Expr // original AST expression
}

// TypeKind represents the kind of type
type TypeKind int

const (
	TypeUnknown TypeKind = iota
	TypeBool
	TypeInt
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUint
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	TypeFloat32
	TypeFloat64
	TypeString
	TypeJSONNumber // encoding/json.Number
	TypeSlice
	TypeArray
	TypeMap
	TypeStruct
	TypePointer
	TypeInterface
)

// IsNumeric returns true if the type is a numeric type
func (t TypeInfo) IsNumeric() bool {
	return (t.Kind >= TypeInt && t.Kind <= TypeFloat64) || t.Kind == TypeJSONNumber
}

// IsInteger returns true if the type is an integer type
func (t TypeInfo) IsInteger() bool {
	return t.Kind >= TypeInt && t.Kind <= TypeUint64
}

// IsFloat returns true if the type is a float type
func (t TypeInfo) IsFloat() bool {
	return t.Kind == TypeFloat32 || t.Kind == TypeFloat64
}

// CodeGenContext holds context for code generation
type CodeGenContext struct {
	Struct     *StructInfo
	Imports    map[string]string // import path -> alias
	Buffer     []string          // lines of generated code
	Options    *GenerateOptions
	VarCounter int         // counter for generating unique variable names
	TypesInfo  *types.Info // type information for resolving underlying types
}

// AddImport adds an import to the context and returns the alias to use
func (ctx *CodeGenContext) AddImport(pkgPath, preferredAlias string) string {
	if alias, exists := ctx.Imports[pkgPath]; exists {
		return alias
	}

	// Find a unique alias
	alias := preferredAlias
	counter := 1
	for {
		// Check if this alias is already used for a different package
		alreadyUsed := false
		for path, existingAlias := range ctx.Imports {
			if existingAlias == alias && path != pkgPath {
				alreadyUsed = true
				break
			}
		}
		if !alreadyUsed {
			break
		}
		alias = preferredAlias + string(rune('0'+counter))
		counter++
	}

	ctx.Imports[pkgPath] = alias
	return alias
}

// UniqueVarName generates a unique variable name
func (ctx *CodeGenContext) UniqueVarName(prefix string) string {
	ctx.VarCounter++
	return prefix + string(rune('0'+ctx.VarCounter))
}

// Import represents an import statement
type Import struct {
	Path  string
	Alias string
}
