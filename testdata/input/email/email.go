package email

// User represents a user with email validation
type User struct {
	Email         string  `validate:"email"`
	OptionalEmail *string `validate:"omitempty,email"`
}

// Contact represents a contact with multiple email addresses
type Contact struct {
	Name         string    `validate:"required"`
	EmailAddress []string  `validate:"omitempty,email"`
	BackupEmails []*string `validate:"omitempty,email"`
}
