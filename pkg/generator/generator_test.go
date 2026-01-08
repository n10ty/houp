package generator

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/n10ty/houp/internal/testutil"
)

var update = flag.Bool("update", false, "update golden files")

func TestGenerateSimple(t *testing.T) {
	testGenerate(t, "simple", "basic.go")
}

func TestGeneratePointers(t *testing.T) {
	testGenerate(t, "pointers", "pointers.go")
}

func TestGenerateSlices(t *testing.T) {
	testGenerate(t, "slices", "slices.go")
}

func TestGenerateUnique(t *testing.T) {
	testGenerate(t, "unique", "unique.go")
}

func TestGenerateDive(t *testing.T) {
	testGenerate(t, "dive", "dive.go")
}

func TestGenerateComplex(t *testing.T) {
	testGenerate(t, "complex", "complex.go")
}

func TestGenerateDateTime(t *testing.T) {
	testGenerate(t, "datetime", "datetime.go")
}

func TestGenerateUUID(t *testing.T) {
	testGenerate(t, "uuid", "uuid.go")
}

func testGenerate(t *testing.T, testDir, inputFile string) {
	t.Helper()

	// Paths
	inputPath := filepath.Join("../../testdata/input", testDir)
	goldenPath := filepath.Join("../../testdata/golden", testDir,
		inputFile[:len(inputFile)-3]+"_validate.go")

	// Generate options
	opts := &GenerateOptions{
		Suffix:         "_validate",
		Overwrite:      true,
		DryRun:         false,
		UnknownTagMode: "fail",
	}

	// Generate validation code
	if err := Generate(inputPath, opts); err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Read generated file
	generatedPath := filepath.Join(inputPath, inputFile[:len(inputFile)-3]+"_validate.go")
	generated, err := ioutil.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	// Compare with golden
	testutil.CompareWithGolden(t, goldenPath, string(generated), *update)
}

func TestUnknownTagFail(t *testing.T) {
	// Create a temporary test file with unknown tag
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package test

type TestStruct struct {
	Name string ` + "`" + `validate:"required,unknowntag"` + "`" + `
}
`
	if err := ioutil.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	opts := &GenerateOptions{
		Suffix:         "_validate",
		Overwrite:      true,
		DryRun:         false,
		UnknownTagMode: "fail",
	}

	// Should fail with unknown tag
	err := Generate(tmpDir, opts)
	if err == nil {
		t.Errorf("expected error for unknown tag, got nil")
	}
}

func TestUnknownTagSkip(t *testing.T) {
	// Create a temporary test file with unknown tag
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package test

type TestStruct struct {
	Name string ` + "`" + `validate:"required,unknowntag"` + "`" + `
	Age  int    ` + "`" + `validate:"min=1"` + "`" + `
}
`
	if err := ioutil.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	opts := &GenerateOptions{
		Suffix:         "_validate",
		Overwrite:      true,
		DryRun:         false,
		UnknownTagMode: "skip",
	}

	// Should succeed and skip unknown tag using GenerateForFiles
	if err := GenerateForFiles([]string{testFile}, opts); err != nil {
		t.Fatalf("GenerateForFiles() with skip mode failed: %v", err)
	}

	// Verify generated file exists
	genFile := filepath.Join(tmpDir, "test_validate.go")
	generated, err := ioutil.ReadFile(genFile)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	// Should contain validation for Age but not unknowntag
	genStr := string(generated)
	if !contains(genStr, "if t.Age < 1") {
		t.Errorf("generated code missing Age validation")
	}
}

func TestDryRun(t *testing.T) {
	inputPath := filepath.Join("../../testdata/input/simple")

	opts := &GenerateOptions{
		Suffix:         "_validate",
		Overwrite:      true,
		DryRun:         true,
		UnknownTagMode: "fail",
	}

	// Should succeed without writing files
	if err := Generate(inputPath, opts); err != nil {
		t.Fatalf("Generate() in dry-run mode failed: %v", err)
	}
}

func TestParseValidationRules(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantLen int
		wantErr bool
	}{
		{
			name:    "required",
			tag:     "required",
			wantLen: 1,
		},
		{
			name:    "required with min/max",
			tag:     "required,min=3,max=50",
			wantLen: 3,
		},
		{
			name:    "omitempty with validators",
			tag:     "omitempty,min=1,max=10",
			wantLen: 3,
		},
		{
			name:    "unique",
			tag:     "unique",
			wantLen: 1,
		},
		{
			name:    "unique with field",
			tag:     "unique=Email",
			wantLen: 1,
		},
		{
			name:    "dive",
			tag:     "dive",
			wantLen: 1,
		},
		{
			name:    "complex combination",
			tag:     "required,min=1,dive,unique=ID",
			wantLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := parseValidationRules(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseValidationRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(rules) != tt.wantLen {
				t.Errorf("parseValidationRules() got %d rules, want %d", len(rules), tt.wantLen)
			}
		})
	}
}

func TestTypeInfoIsNumeric(t *testing.T) {
	tests := []struct {
		kind TypeKind
		want bool
	}{
		{TypeInt, true},
		{TypeInt64, true},
		{TypeFloat32, true},
		{TypeFloat64, true},
		{TypeString, false},
		{TypeBool, false},
		{TypeSlice, false},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			ti := TypeInfo{Kind: tt.kind}
			if got := ti.IsNumeric(); got != tt.want {
				t.Errorf("TypeInfo.IsNumeric() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Add String method for TypeKind for better test output
func (k TypeKind) String() string {
	switch k {
	case TypeBool:
		return "bool"
	case TypeInt:
		return "int"
	case TypeInt8:
		return "int8"
	case TypeInt16:
		return "int16"
	case TypeInt32:
		return "int32"
	case TypeInt64:
		return "int64"
	case TypeUint:
		return "uint"
	case TypeUint8:
		return "uint8"
	case TypeUint16:
		return "uint16"
	case TypeUint32:
		return "uint32"
	case TypeUint64:
		return "uint64"
	case TypeFloat32:
		return "float32"
	case TypeFloat64:
		return "float64"
	case TypeString:
		return "string"
	case TypeSlice:
		return "slice"
	case TypeArray:
		return "array"
	case TypeMap:
		return "map"
	case TypeStruct:
		return "struct"
	case TypePointer:
		return "pointer"
	case TypeInterface:
		return "interface"
	default:
		return "unknown"
	}
}
