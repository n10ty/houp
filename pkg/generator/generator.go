package generator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Generate processes a Go package and generates validation code
func Generate(pkgPath string, opts *GenerateOptions) error {
	// Set defaults
	if opts.Suffix == "" {
		opts.Suffix = "_validate"
	}
	if opts.UnknownTagMode == "" {
		opts.UnknownTagMode = "fail"
	}

	// Parse the package
	pkgInfo, err := ParsePackage(pkgPath)
	if err != nil {
		return fmt.Errorf("failed to parse package: %w", err)
	}

	if len(pkgInfo.Files) == 0 {
		return fmt.Errorf("no Go files found in package %s", pkgPath)
	}

	// Track if we generated any files
	generated := 0

	// Process each file
	for _, fileInfo := range pkgInfo.Files {
		// Skip test files
		if strings.HasSuffix(fileInfo.Name, "_test.go") {
			continue
		}

		// Skip already generated files
		if strings.HasSuffix(fileInfo.Name, opts.Suffix+".go") {
			continue
		}

		// Check if file has any structs needing validation
		hasValidation := false
		for _, structInfo := range fileInfo.Structs {
			if structInfo.NeedsGen {
				hasValidation = true
				break
			}
		}

		if !hasValidation {
			continue
		}

		// Generate validation code
		code, err := GenerateFileValidation(fileInfo, pkgInfo.Name, opts, pkgInfo.TypesInfo)
		if err != nil {
			return fmt.Errorf("failed to generate validation for file %s (package %s): %w", fileInfo.Name, pkgInfo.Name, err)
		}

		if code == "" {
			continue
		}

		// Determine output filename
		baseName := strings.TrimSuffix(fileInfo.Name, ".go")
		outputName := baseName + opts.Suffix + ".go"
		outputPath := filepath.Join(filepath.Dir(fileInfo.Path), outputName)

		// Check if file exists and we shouldn't overwrite
		if !opts.Overwrite {
			if _, err := os.Stat(outputPath); err == nil {
				fmt.Printf("Skipping %s (already exists, use --overwrite to replace)\n", outputPath)
				continue
			}
		}

		// Dry run mode
		if opts.DryRun {
			fmt.Printf("Would generate: %s\n", outputPath)
			continue
		}

		// Write generated code
		if err := ioutil.WriteFile(outputPath, []byte(code), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", outputPath, err)
		}

		fmt.Printf("Generated: %s\n", outputPath)
		generated++
	}

	if generated == 0 {
		fmt.Println("No validation code generated (no structs with validation tags found)")
	} else {
		fmt.Printf("Successfully generated %d validation file(s)\n", generated)
	}

	return nil
}

// GenerateForFiles generates validation for specific files
func GenerateForFiles(files []string, opts *GenerateOptions) error {
	// Set defaults
	if opts.Suffix == "" {
		opts.Suffix = "_validate"
	}
	if opts.UnknownTagMode == "" {
		opts.UnknownTagMode = "fail"
	}

	for _, filePath := range files {
		// Parse single file
		fileInfo, err := ParseFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", filePath, err)
		}

		// Check if file has structs needing validation
		hasValidation := false
		for _, structInfo := range fileInfo.Structs {
			if structInfo.NeedsGen {
				hasValidation = true
				break
			}
		}

		if !hasValidation {
			fmt.Printf("Skipping %s (no validation tags found)\n", filePath)
			continue
		}

		// Extract package name from AST
		pkgName := fileInfo.AST.Name.Name

		// Generate validation code
		code, err := GenerateFileValidation(fileInfo, pkgName, opts, nil)
		if err != nil {
			return fmt.Errorf("failed to generate validation for file %s (package %s): %w", filePath, pkgName, err)
		}

		if code == "" {
			continue
		}

		// Determine output filename
		dir := filepath.Dir(filePath)
		baseName := strings.TrimSuffix(filepath.Base(filePath), ".go")
		outputName := baseName + opts.Suffix + ".go"
		outputPath := filepath.Join(dir, outputName)

		// Check if file exists
		if !opts.Overwrite {
			if _, err := os.Stat(outputPath); err == nil {
				fmt.Printf("Skipping %s (already exists, use --overwrite to replace)\n", outputPath)
				continue
			}
		}

		// Dry run mode
		if opts.DryRun {
			fmt.Printf("Would generate: %s\n", outputPath)
			continue
		}

		// Write generated code
		if err := ioutil.WriteFile(outputPath, []byte(code), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", outputPath, err)
		}

		fmt.Printf("Generated: %s\n", outputPath)
	}

	return nil
}

// DiscoverStructsWithDive finds all structs that are referenced with 'dive' tag
// This is useful for generating empty Validate() methods for structs without their own validation
func DiscoverStructsWithDive(pkgInfo *PackageInfo) map[string]bool {
	referenced := make(map[string]bool)

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
							referenced[typeName] = true
						}
					}
				}
			}
		}
	}

	return referenced
}
