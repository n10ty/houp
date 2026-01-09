package main

import (
	"fmt"
)

func main() {
	// Valid contact info
	validContact := ContactInfo{
		Email:   "john.doe@example.com",
		Country: "US",
	}
	if err := validContact.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Valid contact info passed validation")
	}

	// Invalid email
	invalidEmail := ContactInfo{
		Email:   "not-an-email",
		Country: "US",
	}
	if err := invalidEmail.Validate(); err != nil {
		fmt.Printf("Expected error for invalid email: %v\n", err)
	}

	// Invalid country code
	invalidCountry := ContactInfo{
		Email:   "john.doe@example.com",
		Country: "XX",
	}
	if err := invalidCountry.Validate(); err != nil {
		fmt.Printf("Expected error for invalid country: %v\n", err)
	}

	// Various valid country codes
	countries := []string{"US", "GB", "CA", "AU", "DE", "FR", "JP", "CN", "UA", "XK"}
	fmt.Println("\nTesting various country codes:")
	for _, code := range countries {
		contact := ContactInfo{
			Email:   "test@example.com",
			Country: code,
		}
		if err := contact.Validate(); err != nil {
			fmt.Printf("  %s: FAILED - %v\n", code, err)
		} else {
			fmt.Printf("  %s: OK\n", code)
		}
	}
}
