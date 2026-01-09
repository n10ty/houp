# EqField Validator Example

This example demonstrates the `eqfield` validator tag, which validates that a field equals another field in the same struct.

## Use Cases

1. **Password Confirmation**: Ensure password and confirm password fields match
2. **Order Cancellation**: Require explicit confirmation by matching order IDs
3. **Data Integrity**: Validate that related fields have matching values

## Running the Example

```bash
# Generate validation code (already done)
go run ../../cmd/houp --suffix _validate .

# Run the example
go run .
```

## Expected Output

```
=== Testing UserRegistration ===
✓ Valid user registration passed
✓ Invalid user correctly rejected: field ConfirmPassword must equal field Password

=== Testing OrderModificationRequest ===
✓ Valid order cancellation passed
✓ Invalid cancellation correctly rejected: field CancelOrderId must equal field OrderId
✓ Valid order modification (no cancellation) passed
```

## Features Demonstrated

### 1. Simple Field Equality
```go
type UserRegistration struct {
    Password        string `validate:"required,min=8"`
    ConfirmPassword string `validate:"required,eqfield=Password"`
}
```

Both fields are non-pointer strings, compared directly.

### 2. Pointer Field Equality with Optional Check
```go
type OrderModificationRequest struct {
    CancelOrderId *string `validate:"omitempty,eqfield=OrderId"`
    OrderId       string  `validate:"required"`
}
```

When `CancelOrderId` is nil, validation passes (due to `omitempty`).
When `CancelOrderId` is not nil, it must equal `OrderId`.

## Validation Logic

The `eqfield` validator handles:
- **Non-pointer to non-pointer**: Direct comparison
- **Pointer to pointer**: Dereferences both and compares, handles nil cases
- **Pointer to non-pointer**: Dereferences pointer and compares with non-pointer value
- **Non-pointer to pointer**: Dereferences pointer and compares with non-pointer value

## Generated Code

The generator creates type-safe comparison code that:
1. Handles nil pointer checks
2. Dereferences pointer values when needed
3. Provides clear error messages indicating which fields must match
