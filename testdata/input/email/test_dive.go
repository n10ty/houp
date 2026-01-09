package email

// TestStruct demonstrates different validation approaches for slices
type TestStruct struct {
	// Without dive - validates each email in the slice
	EmailsNoDive []string `validate:"omitempty,email"`

	// With dive - this would be an error since dive is for structs
	// EmailsWithDive []string `validate:"omitempty,dive,email"`
}
