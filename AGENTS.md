# Agent Coding Guidelines for houp

This document provides coding guidelines and commands for AI agents working on the houp codebase.

## Project Overview

**houp** is a Go code generator that creates validation functions for structs based on struct tags. It uses AST parsing for type-safe code generation without runtime reflection.

- **Language:** Go 1.24.7
- **Module:** `github.com/n10ty/houp`
- **Architecture:** Layered (Parser → Validator → CodeGen → Orchestrator)

## Build, Test, and Lint Commands

### Building

```bash
# Build the CLI tool
go build -o houp ./cmd/houp

# Install to GOPATH/bin
go install ./cmd/houp

# Or install directly from GitHub
go install github.com/n10ty/houp/cmd/houp@latest

# Build with verbose output
go build -v ./...
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./pkg/generator -cover

# Run tests with verbose output
go test -v ./...

# Run a single test
go test ./pkg/generator -run TestGenerateSimple

# Run a single test with verbose output
go test -v ./pkg/generator -run TestGenerateComplex

# Update golden files (when expected output changes)
go test ./pkg/generator -update

# Run tests with race detection
go test -race ./...
```

### Linting and Formatting

```bash
# Format all code (run before committing)
go fmt ./...

# Run go vet for static analysis
go vet ./...

# Check for common issues
go mod verify
go mod tidy
```

### Running the Tool

```bash
# Show version
houp --version

# Generate validators for a package
houp ./examples/demo

# Dry run (show what would be generated without writing)
houp --dry-run ./examples/demo

# Custom output suffix
houp --suffix _validators ./examples/demo

# Skip unknown validation tags
houp --unknown-tags=skip ./examples/demo
```

## Code Style Guidelines

### File Organization

- One main concern per file: `parser.go`, `validator.go`, `codegen.go`, `generator.go`
- Test files alongside implementation: `generator_test.go`
- Generated files use configurable suffix: `{source}_validate.go`
- Group related types and functions together within files

### Import Ordering

Follow standard Go conventions:

```go
import (
	// Standard library first
	"fmt"
	"go/ast"
	"strings"

	// External dependencies
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/go/packages"

	// Internal packages last
	"github.com/n10ty/houp/internal/testutil"
)
```

### Formatting

- Use **tabs for indentation** (Go standard)
- Max line length: ~100 characters (flexible, favor readability)
- Use `go fmt` before committing
- Break long function calls across multiple lines with proper alignment

### Type Definitions

```go
// Document all exported types
type ValidationRule interface {
	Name() string
	Validate(fieldType TypeInfo) error
	Generate(ctx *CodeGenContext, field *FieldInfo) (string, error)
}

// Use descriptive field names
type FieldInfo struct {
	Name  string      // Field name
	Type  ast.Expr    // Field type expression
	Tag   string      // Complete struct tag
	Rules []ValidationRule
}
```

### Naming Conventions

- **Packages:** lowercase, single word (`generator`, `testutil`)
- **Exported types:** PascalCase (`RequiredRule`, `TypeInfo`)
- **Unexported types:** camelCase (`typeKind`, `fieldInfo`)
- **Receiver variables:** Single lowercase letter from struct name
  - `r *RequiredRule`, `g *Generator`, `p *Parser`
- **Interface methods:** Descriptive verbs (`Validate`, `Generate`, `Parse`)
- **Constants:** PascalCase for exported, camelCase for unexported

### Error Handling

Use Go's standard error handling patterns:

```go
// Return errors, don't panic
func ParseFile(path string) (*StructInfo, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	// ...
}

// Wrap errors with context using %w
if err != nil {
	return nil, fmt.Errorf("failed to parse package %s: %w", pkgPath, err)
}

// Provide detailed validation errors with field names
return fmt.Errorf("field %s: min value %s is greater than max value %s", 
	field.Name, minVal, maxVal)

// Check all errors explicitly, never ignore
result, err := someFunction()
if err != nil {
	return nil, err
}
```

### Function Signatures

```go
// Return (result, error) pattern
func Generate(opts *GenerateOptions) ([]byte, error)

// Use context structs for complex state
func (r *MinRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error)

// Use options structs for configuration
type GenerateOptions struct {
	Suffix         string
	Overwrite      bool
	DryRun         bool
	UnknownTagMode string
}
```

### Testing

Use table-driven tests and golden files:

```go
func TestParseValidationRules(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		want    []string
		wantErr bool
	}{
		{
			name: "single rule",
			tag:  `validate:"required"`,
			want: []string{"required"},
		},
		// ... more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test implementation
		})
	}
}

// Mark helper functions
func testGenerate(t *testing.T, testDir, inputFile string) {
	t.Helper()
	// ...
}
```

### Comments and Documentation

```go
// Document all exported functions, types, and packages
// Use complete sentences with proper punctuation.

// Package generator provides code generation for struct validation.
package generator

// RequiredRule validates that a field is not a zero value.
// It supports all Go types including pointers, slices, and basic types.
type RequiredRule struct{}

// Generate creates validation code for the required rule.
// Returns generated code as a string or an error if generation fails.
func (r *RequiredRule) Generate(ctx *CodeGenContext, field *FieldInfo) (string, error) {
	// Implementation
}
```

### Generated Code Format

Generated files must include:

```go
// Code generated by houp. DO NOT EDIT.
package demo

import (
	"fmt"
)

// Validate validates the User struct based on validation tags.
// Returns an error if validation fails, nil otherwise.
func (u *User) Validate() error {
	// Validation code
}
```

## Common Patterns

### AST Parsing Pattern

```go
// Use go/packages for robust package loading
cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedSyntax}
pkgs, err := packages.Load(cfg, dir)

// Walk AST to find structs
ast.Inspect(file, func(n ast.Node) bool {
	// Pattern matching on AST nodes
})
```

### Validation Rule Implementation

All validation rules implement the `ValidationRule` interface:

1. **Name()** - Returns the tag name (e.g., "required", "min", "max")
2. **Validate()** - Checks if rule is applicable to field type
3. **Generate()** - Generates validation code for the rule

### Type Resolution

Use `ResolveTypeInfo()` to handle complex types:

```go
typeInfo := ResolveTypeInfo(field.Type, nil)
if typeInfo.IsPointer { /* ... */ }
if typeInfo.IsSlice { /* ... */ }
if typeInfo.IsNumeric() { /* ... */ }
```

## Special Considerations

### Golden File Testing

- Golden files are in `testdata/golden/`
- Update with `-update` flag when expected output changes
- Always review diffs before committing golden file changes

### Unknown Validation Tags

Support two modes:
- **fail** (default): Return error on unknown tags
- **skip**: Ignore unknown tags and continue

### Import Management

Track required imports in `CodeGenContext.Imports`:

```go
ctx.Imports["fmt"] = "fmt"
ctx.Imports["regexp"] = "regexp"
```

## Testing Requirements

- All new validation rules must have tests
- Use golden files for code generation tests
- Maintain test coverage above 60%
- Test both success and error cases
- Test edge cases (empty slices, nil pointers, zero values)

## Files to Never Modify

- `testdata/input/**/*.go` - Test input files (unless adding new test cases)
- `testdata/golden/**/*.go` - Only update via `-update` flag
- `go.sum` - Managed by `go mod`
- Generated files with header `// Code generated by houp. DO NOT EDIT.`

## Common Tasks

### Adding a New Validation Rule

1. Create struct implementing `ValidationRule` in `validator.go`
2. Add to `parseValidationRules()` function
3. Add test input file in `testdata/input/`
4. Run tests with `-update` to create golden file
5. Add documentation to README.md

### Modifying Code Generation

1. Edit relevant function in `codegen.go`
2. Run `go test ./pkg/generator -update` to update golden files
3. Review all golden file diffs carefully
4. Ensure generated code compiles and passes validation

### Debugging Generated Code

Use `--dry-run` flag to see what would be generated:
```bash
houp --dry-run ./examples/demo
```

## Dependencies

- `golang.org/x/tools` - AST parsing and package loading
- `github.com/google/go-cmp` - Test assertions and comparisons

Keep dependencies minimal. Prefer standard library when possible.
