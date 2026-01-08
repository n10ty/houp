package testutil

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// CompareWithGolden compares generated output with a golden file
func CompareWithGolden(t *testing.T, goldenPath string, got string, update bool) {
	t.Helper()

	if update {
		// Update golden file
		if err := ioutil.WriteFile(goldenPath, []byte(got), 0644); err != nil {
			t.Fatalf("failed to update golden file %s: %v", goldenPath, err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	// Read golden file
	want, err := ioutil.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}

	// Compare
	if diff := cmp.Diff(string(want), got); diff != "" {
		t.Errorf("Generated code differs from golden file %s (-want +got):\n%s",
			filepath.Base(goldenPath), diff)
		t.Logf("To update golden files, run: go test -update")
	}
}

// ReadTestData reads a test input file
func ReadTestData(t *testing.T, path string) string {
	t.Helper()

	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read test data %s: %v", path, err)
	}

	return string(data)
}

// WriteTestOutput writes test output to a file
func WriteTestOutput(t *testing.T, path string, content string) {
	t.Helper()

	if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test output %s: %v", path, err)
	}
}
