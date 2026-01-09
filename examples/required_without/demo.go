package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== Required Without Validation Examples ===")

	// Example 1: Valid with FixedPenalty
	fmt.Println("\n1. Valid penalty with fixed amount:")
	fixedPenalty := &Penalty{
		FixedPenalty: &FixedPenalty{
			Amount:   50.0,
			Currency: "USD",
		},
		Reason: "Late payment",
	}
	if err := fixedPenalty.Validate(); err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Valid penalty with fixed amount: $%.2f %s\n",
			fixedPenalty.FixedPenalty.Amount, fixedPenalty.FixedPenalty.Currency)
	}

	// Example 2: Valid with PercentagePenalty
	fmt.Println("\n2. Valid penalty with percentage:")
	percentagePenalty := &Penalty{
		PercentagePenalty: &PercentagePenalty{
			Percentage: 5.0,
		},
		Reason: "Early cancellation",
	}
	if err := percentagePenalty.Validate(); err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Valid penalty with percentage: %.1f%%\n",
			percentagePenalty.PercentagePenalty.Percentage)
	}

	// Example 3: Valid with both
	fmt.Println("\n3. Valid penalty with both fixed and percentage:")
	bothPenalty := &Penalty{
		FixedPenalty: &FixedPenalty{
			Amount:   100.0,
			Currency: "EUR",
		},
		PercentagePenalty: &PercentagePenalty{
			Percentage: 10.0,
		},
		Reason: "Service violation",
	}
	if err := bothPenalty.Validate(); err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Valid penalty with both types\n")
	}

	// Example 4: Invalid - neither provided
	fmt.Println("\n4. Invalid penalty - neither fixed nor percentage:")
	invalidPenalty := &Penalty{
		Reason: "Some reason",
	}
	if err := invalidPenalty.Validate(); err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✓ Valid penalty")
	}

	// Example 5: Invalid - missing reason
	fmt.Println("\n5. Invalid penalty - missing reason:")
	noReasonPenalty := &Penalty{
		FixedPenalty: &FixedPenalty{
			Amount:   25.0,
			Currency: "GBP",
		},
	}
	if err := noReasonPenalty.Validate(); err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✓ Valid penalty")
	}

	fmt.Println("\n=== End of Examples ===")
}
