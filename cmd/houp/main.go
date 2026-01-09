package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/n10ty/houp/pkg/generator"
)

const version = "0.1.0"

func main() {
	// Define flags
	var (
		suffix         = flag.String("suffix", "_validation.gen", "Suffix for the generated validation file (generates validation.gen.go)")
		overwrite      = flag.Bool("overwrite", true, "Overwrite existing generated files")
		dryRun         = flag.Bool("dry-run", false, "Show what would be generated without writing files")
		unknownTagMode = flag.String("unknown-tags", "fail", "How to handle unknown validation tags: 'fail' or 'skip'")
		multiError     = flag.Bool("multi-error", false, "Collect all validation errors (not yet implemented)")
		showVersion    = flag.Bool("version", false, "Show version information")
		help           = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Printf("houp version %s\n", version)
		os.Exit(0)
	}

	if *help {
		usage()
		os.Exit(0)
	}

	// Validate unknown-tags flag
	if *unknownTagMode != "fail" && *unknownTagMode != "skip" {
		fmt.Fprintf(os.Stderr, "Error: --unknown-tags must be 'fail' or 'skip', got: %s\n", *unknownTagMode)
		os.Exit(1)
	}

	// Get package paths from args
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no package path specified\n\n")
		usage()
		os.Exit(1)
	}

	// Create options
	opts := &generator.GenerateOptions{
		Suffix:         *suffix,
		Overwrite:      *overwrite,
		DryRun:         *dryRun,
		UnknownTagMode: *unknownTagMode,
		MultiError:     *multiError,
	}

	// Run generator for each package path
	hasErrors := false
	for _, pkgPath := range args {
		if err := generator.Generate(pkgPath, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating validation for %s: %v\n", pkgPath, err)
			hasErrors = true
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `houp - Static validation generator for Go structs

Usage:
  houp [options] <package-path> [package-path...]

Options:
  --suffix string
        Suffix for generated file (default "_validation.gen")
        Note: Generates a single validation.gen.go file per package

  --overwrite
        Overwrite existing generated files (default true)

  --dry-run
        Show what would be generated without writing files (default false)

  --unknown-tags string
        How to handle unknown validation tags (default "fail")
        Values: "fail" - exit with error
                "skip" - log warning and continue

  --multi-error
        Collect all validation errors instead of returning on first error
        (not yet fully implemented) (default false)

  --version
        Show version information

  --help
        Show this help message

Examples:
  # Show version
  houp --version

  # Generate validation for a package (creates validation.gen.go)
  houp ./models

  # Generate validation for multiple packages
  houp ./models ./api ./services

  # Dry run to see what would be generated
  houp --dry-run ./models

  # Skip unknown validation tags instead of failing
  houp --unknown-tags=skip ./models

  # Use custom suffix for generated file
  houp --suffix=_validate ./models

  # Generate for multiple packages with options
  houp --dry-run --unknown-tags=skip ./models ./api

Output:
  Generates a single validation.gen.go file per package containing all
  Validate() methods for structs with validation tags. This consolidates
  all validation code in one place, avoiding multiple files.

Supported Validation Tags:
  required              Field must not be zero value
  omitempty             Skip validation if field is empty
  min=N                 Minimum value/length (numbers, strings, slices)
  max=N                 Maximum value/length (numbers, strings, slices)
  gt=N                  Greater than (numbers only)
  lt=N                  Less than (numbers only)
  gte=N                 Greater than or equal (numbers only)
  lte=N                 Less than or equal (numbers only)
  regexp=pkg:Var        Match against imported regexp variable
  unique                Values must be unique (slices of scalars)
  unique=Field          Field values must be unique (slices of structs, field must be string)
  dive                  Recursively validate nested structs
  pkg/path:FuncName     Custom validator function

Tag Examples:
  validate:"required"
  validate:"required,min=3,max=50"
  validate:"omitempty,min=1"
  validate:"gt=0,lt=100"
  validate:"regexp=github.com/myorg/validators:EmailPattern"
  validate:"unique=Email"
  validate:"required,dive"
  validate:"github.com/myorg/validators:CustomValidate"

For more information, visit: https://github.com/n10ty/houp
`)
}
