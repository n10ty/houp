# Required Without Validation Example

This example demonstrates the `required_without` validation tag in houp.

## Overview

The `required_without` tag ensures that at least one of two fields must be provided. It's useful for scenarios where you have mutually exclusive options, but at least one must be present.

## Example Scenario

In this example, we have a `Penalty` struct that can represent either:
- A **FixedPenalty** - a fixed amount in a specific currency
- A **PercentagePenalty** - a percentage of some base amount
- Or both

The validation ensures that at least one penalty type is provided.

## Usage

```go
type Penalty struct {
    FixedPenalty      *FixedPenalty      `validate:"required_without=PercentagePenalty"`
    PercentagePenalty *PercentagePenalty `validate:"required_without=FixedPenalty"`
    Reason            string             `validate:"required"`
}
```

## How It Works

- `FixedPenalty` has `required_without=PercentagePenalty` - meaning if `PercentagePenalty` is nil/empty, then `FixedPenalty` is required
- `PercentagePenalty` has `required_without=FixedPenalty` - meaning if `FixedPenalty` is nil/empty, then `PercentagePenalty` is required
- This ensures at least one must be provided, but allows both to be provided as well

## Running the Example

```bash
# Generate validation code
houp .

# Run the example
go run .
```

## Expected Output

The example shows:
1. ✓ Valid penalty with only FixedPenalty
2. ✓ Valid penalty with only PercentagePenalty  
3. ✓ Valid penalty with both types
4. ❌ Invalid penalty with neither type
5. ❌ Invalid penalty missing required Reason field
