package generator

import (
	"fmt"
	"strconv"
	"strings"
)

// RequiredRule validates that a field is not a zero value
type RequiredRule struct{}

func (r *RequiredRule) Name() string { return "required" }

func (r *RequiredRule) Validate(fieldType TypeInfo) error {
	// Required can be applied to any type
	return nil
}

func (r *RequiredRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Generate appropriate check based on type
	if typeInfo.IsPointer {
		return fmt.Sprintf(`	if %s.%s == nil {
		return fmt.Errorf("field %s is required")
	}`, receiverVar, field.Name, field.Name), nil
	}

	if typeInfo.IsSlice {
		return fmt.Sprintf(`	if %s.%s == nil || len(%s.%s) == 0 {
		return fmt.Errorf("field %s is required")
	}`, receiverVar, field.Name, receiverVar, field.Name, field.Name), nil
	}

	switch typeInfo.Kind {
	case TypeString:
		return fmt.Sprintf(`	if %s.%s == "" {
		return fmt.Errorf("field %s is required")
	}`, receiverVar, field.Name, field.Name), nil

	case TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeUint, TypeUint8, TypeUint16, TypeUint32, TypeUint64:
		return fmt.Sprintf(`	if %s.%s == 0 {
		return fmt.Errorf("field %s is required")
	}`, receiverVar, field.Name, field.Name), nil

	case TypeFloat32, TypeFloat64:
		return fmt.Sprintf(`	if %s.%s == 0 {
		return fmt.Errorf("field %s is required")
	}`, receiverVar, field.Name, field.Name), nil

	case TypeBool:
		// For bool, required doesn't make much sense, but check for explicit false
		return fmt.Sprintf(`	// field %s: required validation skipped for bool type`, field.Name), nil

	default:
		// For structs and other types, we can't easily check zero value
		return fmt.Sprintf(`	// field %s: required validation not implemented for this type`, field.Name), nil
	}
}

// OmitEmptyRule wraps other validations to skip if field is empty
type OmitEmptyRule struct{}

func (r *OmitEmptyRule) Name() string { return "omitempty" }

func (r *OmitEmptyRule) Validate(fieldType TypeInfo) error {
	return nil
}

func (r *OmitEmptyRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	// omitempty is handled specially in code generation
	// It wraps subsequent validations
	return "", nil
}

// MinRule validates minimum value or length
type MinRule struct {
	Value string
}

func (r *MinRule) Name() string { return "min" }

func (r *MinRule) Validate(fieldType TypeInfo) error {
	if fieldType.Kind == TypeBool {
		return fmt.Errorf("min validation not applicable to bool type")
	}
	return nil
}

func (r *MinRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Track if we need to dereference
	needsDeref := typeInfo.IsPointer && typeInfo.Elem != nil

	// Handle pointer types
	if typeInfo.IsPointer && typeInfo.Elem != nil {
		typeInfo = *typeInfo.Elem
	}

	// Build field reference
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if needsDeref && typeInfo.Kind == TypeString {
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	if typeInfo.IsSlice {
		return fmt.Sprintf(`	if len(%s.%s) < %s {
		return fmt.Errorf("field %s must have at least %s elements")
	}`, receiverVar, field.Name, r.Value, field.Name, r.Value), nil
	}

	switch typeInfo.Kind {
	case TypeString:
		return fmt.Sprintf(`	if len(%s) < %s {
		return fmt.Errorf("field %s must be at least %s characters")
	}`, fieldRef, r.Value, field.Name, r.Value), nil

	case TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeUint, TypeUint8, TypeUint16, TypeUint32, TypeUint64,
		TypeFloat32, TypeFloat64:
		if needsDeref {
			fieldRef = fmt.Sprintf("*%s.%s", receiverVar, field.Name)
		}
		return fmt.Sprintf(`	if %s < %s {
		return fmt.Errorf("field %s must be at least %s")
	}`, fieldRef, r.Value, field.Name, r.Value), nil

	default:
		return "", fmt.Errorf("min validation not supported for type %s", typeInfo.Name)
	}
}

// MaxRule validates maximum value or length
type MaxRule struct {
	Value string
}

func (r *MaxRule) Name() string { return "max" }

func (r *MaxRule) Validate(fieldType TypeInfo) error {
	if fieldType.Kind == TypeBool {
		return fmt.Errorf("max validation not applicable to bool type")
	}
	return nil
}

func (r *MaxRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Track if we need to dereference
	needsDeref := typeInfo.IsPointer && typeInfo.Elem != nil

	// Handle pointer types
	if typeInfo.IsPointer && typeInfo.Elem != nil {
		typeInfo = *typeInfo.Elem
	}

	// Build field reference
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if needsDeref && typeInfo.Kind == TypeString {
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	if typeInfo.IsSlice {
		return fmt.Sprintf(`	if len(%s.%s) > %s {
		return fmt.Errorf("field %s must have at most %s elements")
	}`, receiverVar, field.Name, r.Value, field.Name, r.Value), nil
	}

	switch typeInfo.Kind {
	case TypeString:
		return fmt.Sprintf(`	if len(%s) > %s {
		return fmt.Errorf("field %s must be at most %s characters")
	}`, fieldRef, r.Value, field.Name, r.Value), nil

	case TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeUint, TypeUint8, TypeUint16, TypeUint32, TypeUint64,
		TypeFloat32, TypeFloat64:
		if needsDeref {
			fieldRef = fmt.Sprintf("*%s.%s", receiverVar, field.Name)
		}
		return fmt.Sprintf(`	if %s > %s {
		return fmt.Errorf("field %s must be at most %s")
	}`, fieldRef, r.Value, field.Name, r.Value), nil

	default:
		return "", fmt.Errorf("max validation not supported for type %s", typeInfo.Name)
	}
}

// GTRule validates greater than (exclusive)
type GTRule struct {
	Value string
}

func (r *GTRule) Name() string { return "gt" }

func (r *GTRule) Validate(fieldType TypeInfo) error {
	if !fieldType.IsNumeric() && fieldType.Kind != TypePointer {
		return fmt.Errorf("gt validation only applicable to numeric types")
	}
	return nil
}

func (r *GTRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Handle pointer types
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if typeInfo.IsPointer {
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	return fmt.Sprintf(`	if %s <= %s {
		return fmt.Errorf("field %s must be greater than %s")
	}`, fieldRef, r.Value, field.Name, r.Value), nil
}

// LTRule validates less than (exclusive)
type LTRule struct {
	Value string
}

func (r *LTRule) Name() string { return "lt" }

func (r *LTRule) Validate(fieldType TypeInfo) error {
	if !fieldType.IsNumeric() && fieldType.Kind != TypePointer {
		return fmt.Errorf("lt validation only applicable to numeric types")
	}
	return nil
}

func (r *LTRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Handle pointer types
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if typeInfo.IsPointer {
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	return fmt.Sprintf(`	if %s >= %s {
		return fmt.Errorf("field %s must be less than %s")
	}`, fieldRef, r.Value, field.Name, r.Value), nil
}

// GTERule validates greater than or equal (inclusive)
type GTERule struct {
	Value string
}

func (r *GTERule) Name() string { return "gte" }

func (r *GTERule) Validate(fieldType TypeInfo) error {
	if !fieldType.IsNumeric() && fieldType.Kind != TypePointer {
		return fmt.Errorf("gte validation only applicable to numeric types")
	}
	return nil
}

func (r *GTERule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Handle pointer types
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if typeInfo.IsPointer {
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	return fmt.Sprintf(`	if %s < %s {
		return fmt.Errorf("field %s must be at least %s")
	}`, fieldRef, r.Value, field.Name, r.Value), nil
}

// LTERule validates less than or equal (inclusive)
type LTERule struct {
	Value string
}

func (r *LTERule) Name() string { return "lte" }

func (r *LTERule) Validate(fieldType TypeInfo) error {
	if !fieldType.IsNumeric() && fieldType.Kind != TypePointer {
		return fmt.Errorf("lte validation only applicable to numeric types")
	}
	return nil
}

func (r *LTERule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Handle pointer types
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if typeInfo.IsPointer {
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	return fmt.Sprintf(`	if %s > %s {
		return fmt.Errorf("field %s must be at most %s")
	}`, fieldRef, r.Value, field.Name, r.Value), nil
}

// RegexpRule validates using an imported regexp variable
type RegexpRule struct {
	ImportPath string
	VarName    string
}

func (r *RegexpRule) Name() string { return "regexp" }

func (r *RegexpRule) Validate(fieldType TypeInfo) error {
	// Handle pointer to string
	if fieldType.IsPointer && fieldType.Elem != nil && fieldType.Elem.Kind == TypeString {
		return nil
	}

	if fieldType.Kind != TypeString {
		// Silently skip non-string types as per requirements
		return nil
	}
	return nil
}

func (r *RegexpRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)

	// Skip non-string types
	if typeInfo.Kind != TypeString {
		if typeInfo.IsPointer && typeInfo.Elem != nil && typeInfo.Elem.Kind != TypeString {
			return "", nil
		}
		if !typeInfo.IsPointer {
			return "", nil
		}
	}

	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Add import
	parts := strings.Split(r.ImportPath, "/")
	pkgName := parts[len(parts)-1]
	alias := ctx.AddImport(r.ImportPath, pkgName)

	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)

	if typeInfo.IsPointer {
		// For pointer to string, dereference
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	return fmt.Sprintf(`	if !%s.%s.MatchString(%s) {
		return fmt.Errorf("field %s does not match required pattern")
	}`, alias, r.VarName, fieldRef, field.Name), nil
}

// UniqueRule validates uniqueness within a slice
type UniqueRule struct {
	FieldName string // empty for scalar slices
}

func (r *UniqueRule) Name() string { return "unique" }

func (r *UniqueRule) Validate(fieldType TypeInfo) error {
	if !fieldType.IsSlice {
		// Silently skip non-slice types as per requirements
		return nil
	}
	return nil
}

func (r *UniqueRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)

	// Skip non-slice types
	if !typeInfo.IsSlice {
		return "", nil
	}

	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))
	mapVar := fmt.Sprintf("seen%s", field.Name)

	if r.FieldName != "" {
		mapVar = fmt.Sprintf("seen%s%s", field.Name, r.FieldName)
	}

	var code strings.Builder

	// Generate map initialization
	code.WriteString(fmt.Sprintf("\t%s := make(map[string]bool, len(%s.%s))\n",
		mapVar, receiverVar, field.Name))

	// Generate loop
	if r.FieldName == "" {
		// Scalar slice - check each element directly
		// For non-string types, we need to convert to string for the map key
		needsConversion := typeInfo.Elem != nil && typeInfo.Elem.Kind != TypeString

		if needsConversion {
			code.WriteString(fmt.Sprintf(`	for i, item := range %s.%s {
		key := fmt.Sprintf("%%v", item)
		if %s[key] {
			return fmt.Errorf("field %s has duplicate value at index %%d", i)
		}
		%s[key] = true
	}`, receiverVar, field.Name, mapVar, field.Name, mapVar))
		} else {
			code.WriteString(fmt.Sprintf(`	for i, item := range %s.%s {
		if %s[item] {
			return fmt.Errorf("field %s has duplicate value at index %%d", i)
		}
		%s[item] = true
	}`, receiverVar, field.Name, mapVar, field.Name, mapVar))
		}
	} else {
		// Struct slice - check specific field
		// Need to determine if slice of pointers or values
		if typeInfo.Elem != nil && typeInfo.Elem.IsPointer {
			// Slice of pointers
			code.WriteString(fmt.Sprintf(`	for i, item := range %s.%s {
		if item == nil {
			continue
		}
		if %s[item.%s] {
			return fmt.Errorf("field %s has duplicate %s at index %%d", i)
		}
		%s[item.%s] = true
	}`, receiverVar, field.Name, mapVar, r.FieldName, field.Name, r.FieldName, mapVar, r.FieldName))
		} else {
			// Slice of values
			code.WriteString(fmt.Sprintf(`	for i, item := range %s.%s {
		if %s[item.%s] {
			return fmt.Errorf("field %s has duplicate %s at index %%d", i)
		}
		%s[item.%s] = true
	}`, receiverVar, field.Name, mapVar, r.FieldName, field.Name, r.FieldName, mapVar, r.FieldName))
		}
	}

	return code.String(), nil
}

// DiveRule validates nested structures
type DiveRule struct{}

func (r *DiveRule) Name() string { return "dive" }

func (r *DiveRule) Validate(fieldType TypeInfo) error {
	return nil
}

func (r *DiveRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	if typeInfo.IsSlice {
		// Dive into slice elements
		if typeInfo.Elem == nil {
			return "", fmt.Errorf("cannot dive into slice: element type unknown")
		}

		elemType := *typeInfo.Elem

		// Handle slice of pointers vs values
		if elemType.IsPointer {
			return fmt.Sprintf(`	for i := range %s.%s {
		if %s.%s[i] == nil {
			continue
		}
		if err := %s.%s[i].Validate(); err != nil {
			return fmt.Errorf("field %s[%%d] validation failed: %%w", i, err)
		}
	}`, receiverVar, field.Name, receiverVar, field.Name, receiverVar, field.Name, field.Name), nil
		}

		return fmt.Sprintf(`	for i := range %s.%s {
		if err := %s.%s[i].Validate(); err != nil {
			return fmt.Errorf("field %s[%%d] validation failed: %%w", i, err)
		}
	}`, receiverVar, field.Name, receiverVar, field.Name, field.Name), nil
	}

	if typeInfo.IsPointer {
		// Dive into pointer to struct
		return fmt.Sprintf(`	if %s.%s != nil {
		if err := %s.%s.Validate(); err != nil {
			return fmt.Errorf("field %s validation failed: %%w", err)
		}
	}`, receiverVar, field.Name, receiverVar, field.Name, field.Name), nil
	}

	// Dive into struct field
	return fmt.Sprintf(`	if err := %s.%s.Validate(); err != nil {
		return fmt.Errorf("field %s validation failed: %%w", err)
	}`, receiverVar, field.Name, field.Name), nil
}

// CustomRule calls a custom validation function
type CustomRule struct {
	ImportPath string
	FuncName   string
}

func (r *CustomRule) Name() string { return "custom" }

func (r *CustomRule) Validate(fieldType TypeInfo) error {
	return nil
}

func (r *CustomRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Add import
	parts := strings.Split(r.ImportPath, "/")
	pkgName := parts[len(parts)-1]
	alias := ctx.AddImport(r.ImportPath, pkgName)

	return fmt.Sprintf(`	if err := %s.%s(%s.%s); err != nil {
		return fmt.Errorf("field %s custom validation failed: %%w", err)
	}`, alias, r.FuncName, receiverVar, field.Name, field.Name), nil
}

// DateTimeRule validates that a string field matches a Go time format
type DateTimeRule struct {
	Format string
}

func (r *DateTimeRule) Name() string { return "datetime" }

func (r *DateTimeRule) Validate(fieldType TypeInfo) error {
	// Handle pointer to string
	if fieldType.IsPointer && fieldType.Elem != nil && fieldType.Elem.Kind == TypeString {
		return nil
	}

	if fieldType.Kind != TypeString {
		return fmt.Errorf("datetime validation only applicable to string types")
	}
	return nil
}

func (r *DateTimeRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, nil)

	// Skip non-string types
	if typeInfo.Kind != TypeString {
		if typeInfo.IsPointer && typeInfo.Elem != nil && typeInfo.Elem.Kind != TypeString {
			return "", fmt.Errorf("datetime validation only applicable to string types")
		}
		if !typeInfo.IsPointer {
			return "", fmt.Errorf("datetime validation only applicable to string types")
		}
	}

	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Add time package import
	ctx.AddImport("time", "time")

	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)

	if typeInfo.IsPointer {
		// For pointer to string, dereference
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	return fmt.Sprintf(`	if _, err := time.Parse("%s", %s); err != nil {
		return fmt.Errorf("field %s must be a valid datetime in format %s: %%w", err)
	}`, r.Format, fieldRef, field.Name, r.Format), nil
}

// UnknownRule represents an unknown validation tag
type UnknownRule struct {
	Raw string
}

func (r *UnknownRule) Name() string { return "unknown" }

func (r *UnknownRule) Validate(fieldType TypeInfo) error {
	return fmt.Errorf("unknown validation rule: %s", r.Raw)
}

func (r *UnknownRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	return "", fmt.Errorf("cannot generate code for unknown rule: %s", r.Raw)
}

// ValidateRules checks all rules for a field and returns errors for unknown/invalid rules
func ValidateRules(field *FieldInfo, unknownTagMode string) error {
	typeInfo := ResolveTypeInfo(field.Type, nil)

	for _, rule := range field.Rules {
		if unknownRule, ok := rule.(*UnknownRule); ok {
			if unknownTagMode == "fail" {
				return fmt.Errorf("unknown validation tag '%s' on field '%s'", unknownRule.Raw, field.Name)
			}
			// skip mode - just log warning (caller should handle)
			continue
		}

		if err := rule.Validate(typeInfo); err != nil {
			return fmt.Errorf("validation rule '%s' not applicable to field '%s': %w", rule.Name(), field.Name, err)
		}
	}

	return nil
}

// HasOmitEmpty checks if the field has omitempty rule
func HasOmitEmpty(rules []ValidationRule) bool {
	for _, rule := range rules {
		if _, ok := rule.(*OmitEmptyRule); ok {
			return true
		}
	}
	return false
}

// GetNonOmitEmptyRules returns all rules except omitempty
func GetNonOmitEmptyRules(rules []ValidationRule) []ValidationRule {
	result := make([]ValidationRule, 0, len(rules))
	for _, rule := range rules {
		if _, ok := rule.(*OmitEmptyRule); !ok {
			result = append(result, rule)
		}
	}
	return result
}

// ValidateUniqueFieldType validates that the field used in unique validation is a string
func ValidateUniqueFieldType(sliceElemType TypeInfo, fieldName string) error {
	if fieldName == "" {
		// Scalar slice - no field to check
		return nil
	}

	// For now, we'll assume string type - full type checking would require AST inspection
	// This will be caught at compile time if wrong
	return nil
}

// Helper function to parse numeric value from string
func parseNumeric(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
