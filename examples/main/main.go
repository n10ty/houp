package main

import (
	"fmt"

	"github.com/n10ty/houp/examples/demo"
)

func main() {
	fmt.Println("ValidGen Demo")
	fmt.Println("=============\n")

	// Valid user
	validUser := &demo.User{
		ID:       "user-123",
		Username: "johndoe",
		Email:    "john@example.com",
		Age:      25,
		Tags:     []string{"golang", "developer"},
		Profile: &demo.Profile{
			Bio:     "Software engineer",
			Website: "https://example.com",
		},
	}

	fmt.Println("Validating valid user...")
	if err := validUser.Validate(); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("✓ Valid user passed validation\n")
	}

	// Invalid user - missing required field
	invalidUser1 := &demo.User{
		ID:       "",
		Username: "johndoe",
		Email:    "john@example.com",
		Age:      25,
		Tags:     []string{"golang"},
		Profile: &demo.Profile{
			Bio: "Software engineer",
		},
	}

	fmt.Println("Validating user with missing ID...")
	if err := invalidUser1.Validate(); err != nil {
		fmt.Printf("✓ Expected error: %v\n\n", err)
	}

	// Invalid user - duplicate tags
	invalidUser2 := &demo.User{
		ID:       "user-123",
		Username: "johndoe",
		Email:    "john@example.com",
		Age:      25,
		Tags:     []string{"golang", "developer", "golang"}, // Duplicate
		Profile: &demo.Profile{
			Bio: "Software engineer",
		},
	}

	fmt.Println("Validating user with duplicate tags...")
	if err := invalidUser2.Validate(); err != nil {
		fmt.Printf("✓ Expected error: %v\n\n", err)
	}

	fmt.Println("All validation tests completed!")
}
