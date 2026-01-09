package email

import (
	"fmt"
)

func Demo() {
	// Test 1: Valid contact with slice of emails
	contact1 := Contact{
		Name:         "John Doe",
		EmailAddress: []string{"john@example.com", "johndoe@example.org"},
	}

	if err := contact1.Validate(); err != nil {
		fmt.Printf("Error validating contact1: %v\n", err)
	} else {
		fmt.Println("✓ Contact1 validated successfully!")
	}

	// Test 2: Invalid contact - one bad email in slice
	contact2 := Contact{
		Name:         "Jane Smith",
		EmailAddress: []string{"jane@example.com", "bad-email"},
	}

	if err := contact2.Validate(); err != nil {
		fmt.Printf("✗ Contact2 validation failed (as expected): %v\n", err)
	} else {
		fmt.Println("Contact2 validated successfully (unexpected!)")
	}

	// Test 3: Valid contact with pointer slice
	email1 := "backup1@example.com"
	email2 := "backup2@example.com"
	contact3 := Contact{
		Name:         "Bob Johnson",
		EmailAddress: []string{"bob@example.com"},
		BackupEmails: []*string{&email1, nil, &email2}, // nil is allowed
	}

	if err := contact3.Validate(); err != nil {
		fmt.Printf("Error validating contact3: %v\n", err)
	} else {
		fmt.Println("✓ Contact3 validated successfully (with nil in pointer slice)!")
	}

	// Test 4: Invalid contact - bad email in pointer slice
	badEmail := "not-an-email"
	contact4 := Contact{
		Name:         "Alice Brown",
		EmailAddress: []string{"alice@example.com"},
		BackupEmails: []*string{&email1, &badEmail},
	}

	if err := contact4.Validate(); err != nil {
		fmt.Printf("✗ Contact4 validation failed (as expected): %v\n", err)
	} else {
		fmt.Println("Contact4 validated successfully (unexpected!)")
	}
}
