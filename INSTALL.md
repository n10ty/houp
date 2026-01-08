# Installation Guide

## Local Installation

ValidGen is now installed locally and ready to use!

### Installation Steps

1. **Install the tool**:
   ```bash
   go install ./cmd/validgen
   ```

2. **Verify installation**:
   ```bash
   validgen --help
   ```

3. **Check location**:
   ```bash
   which validgen
   # Output: /Users/andrii.mazurian/go/bin/validgen
   ```

### Usage

Now you can use `validgen` from anywhere:

```bash
# Generate validation for any package
validgen ./path/to/your/package

# Examples
validgen ./models
validgen ./api/dto
validgen --dry-run ./services
```

### Ensure GOPATH/bin is in PATH

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your PATH:

```bash
# Check if it's in PATH
echo $PATH | grep -o "$HOME/go/bin"

# If not, add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export PATH="$PATH:$HOME/go/bin"
```

## Alternative: Use Local Binary

If you prefer not to install globally, use the local binary:

```bash
# Build local binary
go build -o validgen ./cmd/validgen

# Use it directly
./validgen ./models
```

## Quick Start

1. **Add validation tags to your structs**:

   ```go
   package models
   
   type User struct {
       ID    string `validate:"required"`
       Email string `validate:"required,min=5"`
       Age   int    `validate:"gte=18,lte=100"`
   }
   ```

2. **Generate validation**:

   ```bash
   validgen ./models
   ```

3. **Use generated validation**:

   ```go
   user := &User{ID: "123", Email: "test@example.com", Age: 25}
   if err := user.Validate(); err != nil {
       log.Printf("Validation failed: %v", err)
   }
   ```

## Uninstall

To remove the installed binary:

```bash
rm $(which validgen)
# Or
rm ~/go/bin/validgen
```

## Build from Source

```bash
# Clone or navigate to project
cd /Users/andrii.mazurian/dev/hope

# Install dependencies
go mod download

# Build and install
go install ./cmd/validgen

# Or just build
go build -o validgen ./cmd/validgen
```

## Run Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/generator -v

# With coverage
go test ./pkg/generator -cover
```

## Example Project

Try the demo:

```bash
cd examples/main
go run main.go
```

Output:
```
ValidGen Demo
=============

Validating valid user...
✓ Valid user passed validation

Validating user with missing ID...
✓ Expected error: field ID is required

Validating user with duplicate tags...
✓ Expected error: field Tags has duplicate value at index 2

All validation tests completed!
```

## Next Steps

1. **Read the README**: See `README.md` for comprehensive documentation
2. **Check examples**: See `examples/` for working code
3. **Run tests**: Verify everything works with `go test ./...`
4. **Try it out**: Generate validation for your own structs!

## Support

- View help: `validgen --help`
- See examples in `examples/` directory
- Read full documentation in `README.md`
- Check test cases in `testdata/` for more examples
