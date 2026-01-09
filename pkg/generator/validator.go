package generator

import (
	"fmt"
	"go/types"
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
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
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

// EqFieldRule validates that a field equals another field
type EqFieldRule struct {
	OtherField string
}

func (r *EqFieldRule) Name() string { return "eqfield" }

func (r *EqFieldRule) Validate(fieldType TypeInfo) error {
	// Can be applied to any comparable type
	return nil
}

func (r *EqFieldRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Find the other field to get its type
	var otherFieldInfo *FieldInfo
	for _, f := range ctx.Struct.Fields {
		if f.Name == r.OtherField {
			otherFieldInfo = f
			break
		}
	}

	// If we can't find the other field in Fields, it might not have validation tags
	// We need to check the struct definition anyway
	var otherFieldTypeInfo TypeInfo
	if otherFieldInfo != nil {
		otherFieldTypeInfo = ResolveTypeInfo(otherFieldInfo.Type, ctx.TypesInfo)
	} else {
		// We'll try to compare anyway - compilation will catch type mismatches
		otherFieldTypeInfo = typeInfo
	}

	// Build field references
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	otherFieldRef := fmt.Sprintf("%s.%s", receiverVar, r.OtherField)

	// Handle pointer types - need to compare dereferenced values
	if typeInfo.IsPointer && otherFieldTypeInfo.IsPointer {
		// Both pointers - check if both non-nil and equal, or handle nil mismatch
		return fmt.Sprintf(`	if %s != nil && %s != nil {
		if *%s != *%s {
			return fmt.Errorf("field %s must equal field %s")
		}
	} else if (%s == nil) != (%s == nil) {
		return fmt.Errorf("field %s must equal field %s")
	}`, fieldRef, otherFieldRef, fieldRef, otherFieldRef, field.Name, r.OtherField,
			fieldRef, otherFieldRef, field.Name, r.OtherField), nil
	}

	if typeInfo.IsPointer && !otherFieldTypeInfo.IsPointer {
		// Current field is pointer, other is not
		return fmt.Sprintf(`	if %s != nil {
		if *%s != %s {
			return fmt.Errorf("field %s must equal field %s")
		}
	} else {
		return fmt.Errorf("field %s must equal field %s (pointer is nil)")
	}`, fieldRef, fieldRef, otherFieldRef, field.Name, r.OtherField,
			field.Name, r.OtherField), nil
	}

	if !typeInfo.IsPointer && otherFieldTypeInfo.IsPointer {
		// Other field is pointer, current is not
		return fmt.Sprintf(`	if %s != nil {
		if %s != *%s {
			return fmt.Errorf("field %s must equal field %s")
		}
	} else {
		return fmt.Errorf("field %s must equal field %s (comparison field is nil)")
	}`, otherFieldRef, fieldRef, otherFieldRef, field.Name, r.OtherField,
			field.Name, r.OtherField), nil
	}

	// Neither is a pointer - simple comparison
	return fmt.Sprintf(`	if %s != %s {
		return fmt.Errorf("field %s must equal field %s")
	}`, fieldRef, otherFieldRef, field.Name, r.OtherField), nil
}

// RequiredWithoutRule validates that a field is not zero when another field is zero
type RequiredWithoutRule struct {
	OtherField string
}

func (r *RequiredWithoutRule) Name() string { return "required_without" }

func (r *RequiredWithoutRule) Validate(fieldType TypeInfo) error {
	// Can be applied to any type
	return nil
}

func (r *RequiredWithoutRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Find the other field to get its type
	var otherFieldInfo *FieldInfo
	for _, f := range ctx.Struct.Fields {
		if f.Name == r.OtherField {
			otherFieldInfo = f
			break
		}
	}

	// If we can't find the other field in Fields, it might not have validation tags
	// We need to check the struct definition anyway
	var otherFieldTypeInfo TypeInfo
	if otherFieldInfo != nil {
		otherFieldTypeInfo = ResolveTypeInfo(otherFieldInfo.Type, ctx.TypesInfo)
	} else {
		// Default to assuming pointer type (common for optional fields)
		otherFieldTypeInfo = TypeInfo{IsPointer: true}
	}

	// Generate condition to check if other field is zero/empty
	var otherFieldIsEmpty string
	if otherFieldTypeInfo.IsPointer {
		otherFieldIsEmpty = fmt.Sprintf("%s.%s == nil", receiverVar, r.OtherField)
	} else if otherFieldTypeInfo.IsSlice {
		otherFieldIsEmpty = fmt.Sprintf("(%s.%s == nil || len(%s.%s) == 0)", receiverVar, r.OtherField, receiverVar, r.OtherField)
	} else if otherFieldTypeInfo.Kind == TypeString {
		otherFieldIsEmpty = fmt.Sprintf("%s.%s == \"\"", receiverVar, r.OtherField)
	} else if otherFieldTypeInfo.IsNumeric() {
		otherFieldIsEmpty = fmt.Sprintf("%s.%s == 0", receiverVar, r.OtherField)
	} else {
		// For unknown types, assume pointer
		otherFieldIsEmpty = fmt.Sprintf("%s.%s == nil", receiverVar, r.OtherField)
	}

	// Generate condition to check if current field is zero/empty
	var currentFieldIsEmpty string
	if typeInfo.IsPointer {
		currentFieldIsEmpty = fmt.Sprintf("%s.%s == nil", receiverVar, field.Name)
	} else if typeInfo.IsSlice {
		currentFieldIsEmpty = fmt.Sprintf("(%s.%s == nil || len(%s.%s) == 0)", receiverVar, field.Name, receiverVar, field.Name)
	} else if typeInfo.Kind == TypeString {
		currentFieldIsEmpty = fmt.Sprintf("%s.%s == \"\"", receiverVar, field.Name)
	} else if typeInfo.IsNumeric() {
		currentFieldIsEmpty = fmt.Sprintf("%s.%s == 0", receiverVar, field.Name)
	} else {
		// For unknown types, skip validation
		return fmt.Sprintf("\t// field %s: required_without validation not implemented for this type", field.Name), nil
	}

	// Generate validation: if other field is empty, then this field is required
	return fmt.Sprintf(`	if %s && %s {
		return fmt.Errorf("field %s is required when %s is not provided")
	}`, otherFieldIsEmpty, currentFieldIsEmpty, field.Name, r.OtherField), nil
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
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
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

	case TypeJSONNumber:
		// For json.Number, convert to float64 and compare
		if needsDeref {
			fieldRef = fmt.Sprintf("*%s.%s", receiverVar, field.Name)
		}
		// Use unique variable name to avoid redeclaration
		ctx.VarCounter++
		varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
		return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s < %s {
		return fmt.Errorf("field %s must be at least %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil

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
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
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

	case TypeJSONNumber:
		// For json.Number, convert to float64 and compare
		if needsDeref {
			fieldRef = fmt.Sprintf("*%s.%s", receiverVar, field.Name)
		}
		// Use unique variable name to avoid redeclaration
		ctx.VarCounter++
		varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
		return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s > %s {
		return fmt.Errorf("field %s must be at most %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil

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
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Handle pointer types
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if typeInfo.IsPointer {
		if typeInfo.Elem != nil && typeInfo.Elem.Kind == TypeJSONNumber {
			// Pointer to json.Number
			fieldRef = fmt.Sprintf("(*%s.%s)", receiverVar, field.Name)
			// Use unique variable name to avoid redeclaration
			ctx.VarCounter++
			varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
			return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s <= %s {
		return fmt.Errorf("field %s must be greater than %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil
		}
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	// Handle json.Number
	if typeInfo.Kind == TypeJSONNumber {
		// Use unique variable name to avoid redeclaration
		ctx.VarCounter++
		varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
		return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s <= %s {
		return fmt.Errorf("field %s must be greater than %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil
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
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Handle pointer types
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if typeInfo.IsPointer {
		if typeInfo.Elem != nil && typeInfo.Elem.Kind == TypeJSONNumber {
			// Pointer to json.Number
			fieldRef = fmt.Sprintf("(*%s.%s)", receiverVar, field.Name)
			// Use unique variable name to avoid redeclaration
			ctx.VarCounter++
			varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
			return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s >= %s {
		return fmt.Errorf("field %s must be less than %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil
		}
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	// Handle json.Number
	if typeInfo.Kind == TypeJSONNumber {
		// Use unique variable name to avoid redeclaration
		ctx.VarCounter++
		varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
		return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s >= %s {
		return fmt.Errorf("field %s must be less than %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil
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
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Handle pointer types
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if typeInfo.IsPointer {
		if typeInfo.Elem != nil && typeInfo.Elem.Kind == TypeJSONNumber {
			// Pointer to json.Number
			fieldRef = fmt.Sprintf("(*%s.%s)", receiverVar, field.Name)
			// Use unique variable name to avoid redeclaration
			ctx.VarCounter++
			varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
			return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s < %s {
		return fmt.Errorf("field %s must be at least %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil
		}
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	// Handle json.Number
	if typeInfo.Kind == TypeJSONNumber {
		// Use unique variable name to avoid redeclaration
		ctx.VarCounter++
		varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
		return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s < %s {
		return fmt.Errorf("field %s must be at least %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil
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
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Handle pointer types
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)
	if typeInfo.IsPointer {
		if typeInfo.Elem != nil && typeInfo.Elem.Kind == TypeJSONNumber {
			// Pointer to json.Number
			fieldRef = fmt.Sprintf("(*%s.%s)", receiverVar, field.Name)
			// Use unique variable name to avoid redeclaration
			ctx.VarCounter++
			varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
			return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s > %s {
		return fmt.Errorf("field %s must be at most %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil
		}
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	// Handle json.Number
	if typeInfo.Kind == TypeJSONNumber {
		// Use unique variable name to avoid redeclaration
		ctx.VarCounter++
		varName := fmt.Sprintf("%sFloat%d", field.Name, ctx.VarCounter)
		return fmt.Sprintf(`	%s, err := %s.Float64()
	if err != nil {
		return fmt.Errorf("field %s must be a valid number: %%w", err)
	}
	if %s > %s {
		return fmt.Errorf("field %s must be at most %s")
	}`, varName, fieldRef, field.Name, varName, r.Value, field.Name, r.Value), nil
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
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)

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
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)

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
type DiveRule struct {
	// ElementRules are validation rules to apply to each element
	// These are the rules that come AFTER the dive tag
	ElementRules []ValidationRule
}

func (r *DiveRule) Name() string { return "dive" }

func (r *DiveRule) Validate(fieldType TypeInfo) error {
	return nil
}

func (r *DiveRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)
	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	if typeInfo.IsSlice {
		// Dive into slice elements
		if typeInfo.Elem == nil {
			return "", fmt.Errorf("cannot dive into slice: element type unknown")
		}

		elemType := *typeInfo.Elem

		// Check if element is a struct type (or pointer to struct)
		isStructElem := false
		if elemType.IsPointer && elemType.Elem != nil {
			isStructElem = elemType.Elem.Kind == TypeStruct || elemType.Elem.Kind == TypeUnknown
		} else {
			isStructElem = elemType.Kind == TypeStruct || elemType.Kind == TypeUnknown
		}

		// If we have element-specific validation rules AND element is primitive
		if len(r.ElementRules) > 0 && !isStructElem {
			// Generate validation for primitive slice elements with custom rules
			return r.generateSliceElementValidation(ctx, field, elemType, receiverVar)
		}

		// Check if element type is from an external package
		isExternalType := r.isExternalType(elemType)

		// For struct elements, we need to:
		// 1. Call .Validate() on each element
		// 2. Apply any element rules (like unique) that work on the struct level
		if len(r.ElementRules) > 0 && isStructElem {
			// Generate both Validate() calls and struct-level rules like unique
			return r.generateStructSliceValidation(ctx, field, elemType, receiverVar, isExternalType)
		}

		// Skip generating Validate() calls for external types without validation tags
		if isExternalType {
			return fmt.Sprintf("\t// Skipping dive validation for external type without validation tags"), nil
		}

		// No element rules - just call Validate() on struct elements
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

	// Check if type is from an external package
	isExternalType := r.isExternalType(typeInfo)

	// Skip generating Validate() calls for external types
	if isExternalType {
		return fmt.Sprintf("\t// Skipping dive validation for external type without validation tags"), nil
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

// isExternalType checks if a type is from an external package
func (r *DiveRule) isExternalType(typeInfo TypeInfo) bool {
	// Check if the type has a package path (indicating it's from another package)
	if typeInfo.PkgPath != "" {
		return true
	}

	// For pointer types, check the underlying element type
	if typeInfo.IsPointer && typeInfo.Elem != nil {
		return r.isExternalType(*typeInfo.Elem)
	}

	return false
}

// generateStructSliceValidation handles dive on slice of structs with additional element rules
func (r *DiveRule) generateStructSliceValidation(ctx *CodeGenContext, field *FieldInfo, elemType TypeInfo, receiverVar string, isExternalType bool) (string, error) {
	var code strings.Builder

	// Only call Validate() on each element if it's not an external type
	if !isExternalType {
		if elemType.IsPointer {
			code.WriteString(fmt.Sprintf(`	for i := range %s.%s {
		if %s.%s[i] == nil {
			continue
		}
		if err := %s.%s[i].Validate(); err != nil {
			return fmt.Errorf("field %s[%%d] validation failed: %%w", i, err)
		}
	}`, receiverVar, field.Name, receiverVar, field.Name, receiverVar, field.Name, field.Name))
		} else {
			code.WriteString(fmt.Sprintf(`	for i := range %s.%s {
		if err := %s.%s[i].Validate(); err != nil {
			return fmt.Errorf("field %s[%%d] validation failed: %%w", i, err)
		}
	}`, receiverVar, field.Name, receiverVar, field.Name, field.Name))
		}
	} else {
		// Add a comment indicating we're skipping validation for external types
		code.WriteString(fmt.Sprintf("\t// Skipping Validate() call for external type %s in field %s\n", elemType.Name, field.Name))
	}

	// Now apply struct-level rules (like unique)
	for _, rule := range r.ElementRules {
		ruleCode, err := rule.Generate(ctx, field)
		if err != nil {
			return "", fmt.Errorf("failed to generate dive element rule %s: %w", rule.Name(), err)
		}

		if ruleCode != "" {
			code.WriteString("\n")
			code.WriteString(ruleCode)
		}
	}

	return code.String(), nil
}

// generateSliceElementValidation generates validation code for slice elements with custom rules
func (r *DiveRule) generateSliceElementValidation(ctx *CodeGenContext, field *FieldInfo, elemType TypeInfo, receiverVar string) (string, error) {
	// Create a temporary FieldInfo for the element
	// This allows us to reuse existing rule generation logic
	elemField := &FieldInfo{
		Name:  "elem",
		Type:  elemType.UnderlyingGo,
		Rules: r.ElementRules,
	}

	// Generate validation code for all element rules first to see if we have any valid code
	var validationLines []string
	for _, rule := range r.ElementRules {
		// Generate the rule code
		ruleCode, err := rule.Generate(ctx, elemField)
		if err != nil {
			return "", fmt.Errorf("failed to generate dive element rule %s: %w", rule.Name(), err)
		}

		if ruleCode != "" {
			// Fix up the generated code to work in the loop context
			// 1. Replace receiver.elem with just elem (the loop variable)
			ruleCode = strings.ReplaceAll(ruleCode, receiverVar+".elem", "elem")

			// 2. Update error messages to include array index
			ruleCode = strings.ReplaceAll(ruleCode, `"field elem`, fmt.Sprintf(`"field %s[%%d]`, field.Name))

			// 3. Add index parameter to fmt.Errorf calls
			// Only replace the closing ) in fmt.Errorf lines
			lines := strings.Split(strings.TrimSpace(ruleCode), "\n")
			var fixedLines []string
			for _, line := range lines {
				if strings.Contains(line, "fmt.Errorf") && !strings.Contains(line, ", i)") {
					// Add ", i" before the last ")"
					lastParen := strings.LastIndex(line, ")")
					if lastParen > 0 {
						line = line[:lastParen] + ", i" + line[lastParen:]
					}
				}
				fixedLines = append(fixedLines, line)
			}
			validationLines = append(validationLines, fixedLines...)
		}
	}

	// If no validation code was generated, don't create an empty loop
	if len(validationLines) == 0 {
		return "", nil
	}

	var code strings.Builder

	// Start loop
	code.WriteString(fmt.Sprintf("\tfor i, elem := range %s.%s {\n", receiverVar, field.Name))

	// Add validation lines
	for _, line := range validationLines {
		code.WriteString("\t\t")
		code.WriteString(line)
		code.WriteString("\n")
	}

	code.WriteString("\t}")

	return code.String(), nil
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

// UUIDRule validates that a string field is a valid UUID
type UUIDRule struct{}

func (r *UUIDRule) Name() string { return "uuid" }

func (r *UUIDRule) Validate(fieldType TypeInfo) error {
	// Handle pointer to string
	if fieldType.IsPointer && fieldType.Elem != nil && fieldType.Elem.Kind == TypeString {
		return nil
	}

	if fieldType.Kind != TypeString {
		return fmt.Errorf("uuid validation only applicable to string types")
	}
	return nil
}

func (r *UUIDRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)

	// Skip non-string types
	if typeInfo.Kind != TypeString {
		if typeInfo.IsPointer && typeInfo.Elem != nil && typeInfo.Elem.Kind != TypeString {
			return "", fmt.Errorf("uuid validation only applicable to string types")
		}
		if !typeInfo.IsPointer {
			return "", fmt.Errorf("uuid validation only applicable to string types")
		}
	}

	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Add regexp package import
	ctx.AddImport("regexp", "regexp")

	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)

	// UUID regex pattern (matches UUID v1-v5)
	uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`

	if typeInfo.IsPointer {
		// For pointer to string, dereference
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	// Get or create package-level regexp variable
	regexpVar := ctx.AddRegexpVar(uuidPattern, "uuidRegexp")

	return fmt.Sprintf(`	if !%s.MatchString(%s) {
		return fmt.Errorf("field %s must be a valid UUID")
	}`, regexpVar, fieldRef, field.Name), nil
}

// ISO4217Rule validates that a string field is a valid ISO 4217 currency code
type ISO4217Rule struct{}

func (r *ISO4217Rule) Name() string { return "iso4217" }

func (r *ISO4217Rule) Validate(fieldType TypeInfo) error {
	// Handle pointer to string
	if fieldType.IsPointer && fieldType.Elem != nil && fieldType.Elem.Kind == TypeString {
		return nil
	}

	if fieldType.Kind != TypeString {
		return fmt.Errorf("iso4217 validation only applicable to string types")
	}
	return nil
}

func (r *ISO4217Rule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)

	// Skip non-string types
	if typeInfo.Kind != TypeString {
		if typeInfo.IsPointer && typeInfo.Elem != nil && typeInfo.Elem.Kind != TypeString {
			return "", fmt.Errorf("iso4217 validation only applicable to string types")
		}
		if !typeInfo.IsPointer {
			return "", fmt.Errorf("iso4217 validation only applicable to string types")
		}
	}

	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)

	if typeInfo.IsPointer {
		// For pointer to string, dereference
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	// Use unique variable name to avoid redeclaration
	ctx.VarCounter++
	mapVar := fmt.Sprintf("iso4217Codes%d", ctx.VarCounter)

	// Generate the validation code with an inline map
	return fmt.Sprintf(`	%s := map[string]struct{}{
		"AFN": {}, "EUR": {}, "ALL": {}, "DZD": {}, "USD": {},
		"AOA": {}, "XCD": {}, "ARS": {}, "AMD": {}, "AWG": {},
		"AUD": {}, "AZN": {}, "BSD": {}, "BHD": {}, "BDT": {},
		"BBD": {}, "BYN": {}, "BZD": {}, "XOF": {}, "BMD": {},
		"INR": {}, "BTN": {}, "BOB": {}, "BOV": {}, "BAM": {},
		"BWP": {}, "NOK": {}, "BRL": {}, "BND": {}, "BGN": {},
		"BIF": {}, "CVE": {}, "KHR": {}, "XAF": {}, "CAD": {},
		"KYD": {}, "CLP": {}, "CLF": {}, "CNY": {}, "COP": {},
		"COU": {}, "KMF": {}, "CDF": {}, "NZD": {}, "CRC": {},
		"CUP": {}, "CZK": {}, "DKK": {}, "DJF": {}, "DOP": {},
		"EGP": {}, "SVC": {}, "ERN": {}, "SZL": {}, "ETB": {},
		"FKP": {}, "FJD": {}, "XPF": {}, "GMD": {}, "GEL": {},
		"GHS": {}, "GIP": {}, "GTQ": {}, "GBP": {}, "GNF": {},
		"GYD": {}, "HTG": {}, "HNL": {}, "HKD": {}, "HUF": {},
		"ISK": {}, "IDR": {}, "XDR": {}, "IRR": {}, "IQD": {},
		"ILS": {}, "JMD": {}, "JPY": {}, "JOD": {}, "KZT": {},
		"KES": {}, "KPW": {}, "KRW": {}, "KWD": {}, "KGS": {},
		"LAK": {}, "LBP": {}, "LSL": {}, "ZAR": {}, "LRD": {},
		"LYD": {}, "CHF": {}, "MOP": {}, "MKD": {}, "MGA": {},
		"MWK": {}, "MYR": {}, "MVR": {}, "MRU": {}, "MUR": {},
		"XUA": {}, "MXN": {}, "MXV": {}, "MDL": {}, "MNT": {},
		"MAD": {}, "MZN": {}, "MMK": {}, "NAD": {}, "NPR": {},
		"NIO": {}, "NGN": {}, "OMR": {}, "PKR": {}, "PAB": {},
		"PGK": {}, "PYG": {}, "PEN": {}, "PHP": {}, "PLN": {},
		"QAR": {}, "RON": {}, "RUB": {}, "RWF": {}, "SHP": {},
		"WST": {}, "STN": {}, "SAR": {}, "RSD": {}, "SCR": {},
		"SLE": {}, "SGD": {}, "XSU": {}, "SBD": {}, "SOS": {},
		"SSP": {}, "LKR": {}, "SDG": {}, "SRD": {}, "SEK": {},
		"CHE": {}, "CHW": {}, "SYP": {}, "TWD": {}, "TJS": {},
		"TZS": {}, "THB": {}, "TOP": {}, "TTD": {}, "TND": {},
		"TRY": {}, "TMT": {}, "UGX": {}, "UAH": {}, "AED": {},
		"USN": {}, "UYU": {}, "UYI": {}, "UYW": {}, "UZS": {},
		"VUV": {}, "VES": {}, "VED": {}, "VND": {}, "YER": {},
		"ZMW": {}, "ZWG": {}, "XBA": {}, "XBB": {}, "XBC": {},
		"XBD": {}, "XCG": {}, "XTS": {}, "XXX": {}, "XAU": {},
		"XPD": {}, "XPT": {}, "XAG": {},
	}
	if _, ok := %s[%s]; !ok {
		return fmt.Errorf("field %s must be a valid ISO 4217 currency code")
	}`, mapVar, mapVar, fieldRef, field.Name), nil
}

// EmailRule validates that a string field is a valid email address
type EmailRule struct{}

func (r *EmailRule) Name() string { return "email" }

func (r *EmailRule) Validate(fieldType TypeInfo) error {
	// Handle pointer to string
	if fieldType.IsPointer && fieldType.Elem != nil && fieldType.Elem.Kind == TypeString {
		return nil
	}

	// Handle slice of strings
	if fieldType.IsSlice && fieldType.Elem != nil && fieldType.Elem.Kind == TypeString {
		return nil
	}

	// Handle slice of pointer to strings
	if fieldType.IsSlice && fieldType.Elem != nil && fieldType.Elem.IsPointer && fieldType.Elem.Elem != nil && fieldType.Elem.Elem.Kind == TypeString {
		return nil
	}

	if fieldType.Kind != TypeString {
		return fmt.Errorf("email validation only applicable to string types")
	}
	return nil
}

func (r *EmailRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)

	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Add regexp package import
	ctx.AddImport("regexp", "regexp")

	// Basic email regex pattern - intentionally broad
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Get or create package-level regexp variable
	regexpVar := ctx.AddRegexpVar(emailPattern, "emailRegexp")

	// Handle slice of strings
	if typeInfo.IsSlice {
		if typeInfo.Elem == nil {
			return "", fmt.Errorf("cannot validate slice: element type unknown")
		}

		elemType := *typeInfo.Elem

		// Handle slice of pointer to strings
		if elemType.IsPointer {
			return fmt.Sprintf(`	for i, email := range %s.%s {
		if email == nil {
			continue
		}
		if !%s.MatchString(*email) {
			return fmt.Errorf("field %s[%%d] must be a valid email address", i)
		}
	}`, receiverVar, field.Name, regexpVar, field.Name), nil
		}

		// Handle slice of strings
		if elemType.Kind == TypeString {
			return fmt.Sprintf(`	for i, email := range %s.%s {
		if !%s.MatchString(email) {
			return fmt.Errorf("field %s[%%d] must be a valid email address", i)
		}
	}`, receiverVar, field.Name, regexpVar, field.Name), nil
		}

		return "", fmt.Errorf("email validation only applicable to string types")
	}

	// Skip non-string types
	if typeInfo.Kind != TypeString {
		if typeInfo.IsPointer && typeInfo.Elem != nil && typeInfo.Elem.Kind != TypeString {
			return "", fmt.Errorf("email validation only applicable to string types")
		}
		if !typeInfo.IsPointer {
			return "", fmt.Errorf("email validation only applicable to string types")
		}
	}

	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)

	if typeInfo.IsPointer {
		// For pointer to string, dereference
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	return fmt.Sprintf(`	if !%s.MatchString(%s) {
		return fmt.Errorf("field %s must be a valid email address")
	}`, regexpVar, fieldRef, field.Name), nil
}

// ISO3166_1_Alpha2Rule validates that a string field is a valid ISO 3166-1 alpha-2 country code
type ISO3166_1_Alpha2Rule struct{}

func (r *ISO3166_1_Alpha2Rule) Name() string { return "iso3166_1_alpha2" }

func (r *ISO3166_1_Alpha2Rule) Validate(fieldType TypeInfo) error {
	// Handle pointer to string
	if fieldType.IsPointer && fieldType.Elem != nil && fieldType.Elem.Kind == TypeString {
		return nil
	}

	if fieldType.Kind != TypeString {
		return fmt.Errorf("iso3166_1_alpha2 validation only applicable to string types")
	}
	return nil
}

func (r *ISO3166_1_Alpha2Rule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)

	// Skip non-string types
	if typeInfo.Kind != TypeString {
		if typeInfo.IsPointer && typeInfo.Elem != nil && typeInfo.Elem.Kind != TypeString {
			return "", fmt.Errorf("iso3166_1_alpha2 validation only applicable to string types")
		}
		if !typeInfo.IsPointer {
			return "", fmt.Errorf("iso3166_1_alpha2 validation only applicable to string types")
		}
	}

	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))
	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)

	if typeInfo.IsPointer {
		// For pointer to string, dereference
		fieldRef = fmt.Sprintf("*%s", fieldRef)
	}

	// Use unique variable name to avoid redeclaration
	ctx.VarCounter++
	mapVar := fmt.Sprintf("iso3166_1_alpha2Codes%d", ctx.VarCounter)

	// Generate the validation code with an inline map
	return fmt.Sprintf(`	%s := map[string]struct{}{
		"AF": {}, "AX": {}, "AL": {}, "DZ": {}, "AS": {},
		"AD": {}, "AO": {}, "AI": {}, "AQ": {}, "AG": {},
		"AR": {}, "AM": {}, "AW": {}, "AU": {}, "AT": {},
		"AZ": {}, "BS": {}, "BH": {}, "BD": {}, "BB": {},
		"BY": {}, "BE": {}, "BZ": {}, "BJ": {}, "BM": {},
		"BT": {}, "BO": {}, "BQ": {}, "BA": {}, "BW": {},
		"BV": {}, "BR": {}, "IO": {}, "BN": {}, "BG": {},
		"BF": {}, "BI": {}, "KH": {}, "CM": {}, "CA": {},
		"CV": {}, "KY": {}, "CF": {}, "TD": {}, "CL": {},
		"CN": {}, "CX": {}, "CC": {}, "CO": {}, "KM": {},
		"CG": {}, "CD": {}, "CK": {}, "CR": {}, "CI": {},
		"HR": {}, "CU": {}, "CW": {}, "CY": {}, "CZ": {},
		"DK": {}, "DJ": {}, "DM": {}, "DO": {}, "EC": {},
		"EG": {}, "SV": {}, "GQ": {}, "ER": {}, "EE": {},
		"ET": {}, "FK": {}, "FO": {}, "FJ": {}, "FI": {},
		"FR": {}, "GF": {}, "PF": {}, "TF": {}, "GA": {},
		"GM": {}, "GE": {}, "DE": {}, "GH": {}, "GI": {},
		"GR": {}, "GL": {}, "GD": {}, "GP": {}, "GU": {},
		"GT": {}, "GG": {}, "GN": {}, "GW": {}, "GY": {},
		"HT": {}, "HM": {}, "VA": {}, "HN": {}, "HK": {},
		"HU": {}, "IS": {}, "IN": {}, "ID": {}, "IR": {},
		"IQ": {}, "IE": {}, "IM": {}, "IL": {}, "IT": {},
		"JM": {}, "JP": {}, "JE": {}, "JO": {}, "KZ": {},
		"KE": {}, "KI": {}, "KP": {}, "KR": {}, "KW": {},
		"KG": {}, "LA": {}, "LV": {}, "LB": {}, "LS": {},
		"LR": {}, "LY": {}, "LI": {}, "LT": {}, "LU": {},
		"MO": {}, "MK": {}, "MG": {}, "MW": {}, "MY": {},
		"MV": {}, "ML": {}, "MT": {}, "MH": {}, "MQ": {},
		"MR": {}, "MU": {}, "YT": {}, "MX": {}, "FM": {},
		"MD": {}, "MC": {}, "MN": {}, "ME": {}, "MS": {},
		"MA": {}, "MZ": {}, "MM": {}, "NA": {}, "NR": {},
		"NP": {}, "NL": {}, "NC": {}, "NZ": {}, "NI": {},
		"NE": {}, "NG": {}, "NU": {}, "NF": {}, "MP": {},
		"NO": {}, "OM": {}, "PK": {}, "PW": {}, "PS": {},
		"PA": {}, "PG": {}, "PY": {}, "PE": {}, "PH": {},
		"PN": {}, "PL": {}, "PT": {}, "PR": {}, "QA": {},
		"RE": {}, "RO": {}, "RU": {}, "RW": {}, "BL": {},
		"SH": {}, "KN": {}, "LC": {}, "MF": {}, "PM": {},
		"VC": {}, "WS": {}, "SM": {}, "ST": {}, "SA": {},
		"SN": {}, "RS": {}, "SC": {}, "SL": {}, "SG": {},
		"SX": {}, "SK": {}, "SI": {}, "SB": {}, "SO": {},
		"ZA": {}, "GS": {}, "SS": {}, "ES": {}, "LK": {},
		"SD": {}, "SR": {}, "SJ": {}, "SZ": {}, "SE": {},
		"CH": {}, "SY": {}, "TW": {}, "TJ": {}, "TZ": {},
		"TH": {}, "TL": {}, "TG": {}, "TK": {}, "TO": {},
		"TT": {}, "TN": {}, "TR": {}, "TM": {}, "TC": {},
		"TV": {}, "UG": {}, "UA": {}, "AE": {}, "GB": {},
		"US": {}, "UM": {}, "UY": {}, "UZ": {}, "VU": {},
		"VE": {}, "VN": {}, "VG": {}, "VI": {}, "WF": {},
		"EH": {}, "YE": {}, "ZM": {}, "ZW": {}, "XK": {},
	}
	if _, ok := %s[%s]; !ok {
		return fmt.Errorf("field %s must be a valid ISO 3166-1 alpha-2 country code")
	}`, mapVar, mapVar, fieldRef, field.Name), nil
}

// DateTimeRule validates that a string field matches a Go time format
type DateTimeRule struct {
	Format string
}

func (r *DateTimeRule) Name() string { return "datetime" }

func (r *DateTimeRule) Validate(fieldType TypeInfo) error {
	// Handle pointer to string or custom string type
	if fieldType.IsPointer && fieldType.Elem != nil {
		// Check if element is string or string-based
		if fieldType.Elem.Kind == TypeString || fieldType.Elem.Kind == TypeUnknown {
			return nil
		}
		return fmt.Errorf("datetime validation only applicable to string types")
	}

	// Handle direct string or custom string types (which have Kind == TypeString after resolution)
	if fieldType.Kind != TypeString && fieldType.Kind != TypeUnknown {
		return fmt.Errorf("datetime validation only applicable to string types")
	}
	return nil
}

func (r *DateTimeRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	typeInfo := ResolveTypeInfo(field.Type, ctx.TypesInfo)

	// Check if this is a valid type for datetime validation
	// Valid types: string, custom string types (TypeUnknown with string underlying), pointer to either
	isValidType := typeInfo.Kind == TypeString || typeInfo.Kind == TypeUnknown
	if typeInfo.IsPointer && typeInfo.Elem != nil {
		isValidType = typeInfo.Elem.Kind == TypeString || typeInfo.Elem.Kind == TypeUnknown
	}

	if !isValidType {
		return "", fmt.Errorf("datetime validation only applicable to string types")
	}

	receiverVar := strings.ToLower(string(ctx.Struct.Name[0]))

	// Add time package import
	ctx.AddImport("time", "time")

	fieldRef := fmt.Sprintf("%s.%s", receiverVar, field.Name)

	if typeInfo.IsPointer {
		// For pointer to string or custom string type
		if typeInfo.Elem != nil {
			// Check if the element is a custom string type
			elemNeedsCast := typeInfo.Elem.Name != "" && typeInfo.Elem.Name != "string"
			if elemNeedsCast {
				fieldRef = fmt.Sprintf("string(*%s)", fieldRef)
			} else {
				fieldRef = fmt.Sprintf("*%s", fieldRef)
			}
		} else {
			fieldRef = fmt.Sprintf("*%s", fieldRef)
		}
	} else {
		// For non-pointer types, check if it's a custom string type
		needsCast := typeInfo.Name != "" && typeInfo.Name != "string"
		if needsCast {
			fieldRef = fmt.Sprintf("string(%s)", fieldRef)
		}
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
func ValidateRules(field *FieldInfo, unknownTagMode string, typesInfo *types.Info) error {
	typeInfo := ResolveTypeInfo(field.Type, typesInfo)

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
