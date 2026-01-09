package dive_primitives

import "testing"

func TestEmailListValidation(t *testing.T) {
	// Valid emails
	valid := &EmailList{
		Emails: []string{"test@example.com", "user@domain.org"},
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("Valid emails failed: %v", err)
	}

	// Invalid email
	invalid := &EmailList{
		Emails: []string{"test@example.com", "invalid-email"},
	}
	if err := invalid.Validate(); err == nil {
		t.Error("Invalid email should fail")
	}
}

func TestNumberListValidation(t *testing.T) {
	// Valid numbers
	valid := &NumberList{
		Numbers: []int{1, 50, 100},
	}
	if err := valid.Validate(); err != nil {
		t.Errorf("Valid numbers failed: %v", err)
	}

	// Number too small
	tooSmall := &NumberList{
		Numbers: []int{1, 0, 50},
	}
	if err := tooSmall.Validate(); err == nil {
		t.Error("Number too small should fail")
	}

	// Number too large
	tooLarge := &NumberList{
		Numbers: []int{1, 50, 101},
	}
	if err := tooLarge.Validate(); err == nil {
		t.Error("Number too large should fail")
	}
}
