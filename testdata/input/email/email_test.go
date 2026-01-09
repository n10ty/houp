package email

import "testing"

func TestDemo(t *testing.T) {
	Demo()
}

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid email",
			user: User{
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "valid complex email",
			user: User{
				Email: "test.user+tag@sub.example.co.uk",
			},
			wantErr: false,
		},
		{
			name: "invalid email - no @",
			user: User{
				Email: "invalid.email.com",
			},
			wantErr: true,
		},
		{
			name: "invalid email - no domain",
			user: User{
				Email: "test@",
			},
			wantErr: true,
		},
		{
			name: "invalid email - no local part",
			user: User{
				Email: "@example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserValidateOptionalEmail(t *testing.T) {
	validEmail := "test@example.com"
	invalidEmail := "invalid"

	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "nil optional email - should pass",
			user: User{
				Email:         "required@example.com",
				OptionalEmail: nil,
			},
			wantErr: false,
		},
		{
			name: "valid optional email",
			user: User{
				Email:         "required@example.com",
				OptionalEmail: &validEmail,
			},
			wantErr: false,
		},
		{
			name: "invalid optional email",
			user: User{
				Email:         "required@example.com",
				OptionalEmail: &invalidEmail,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContactValidate(t *testing.T) {
	validEmail1 := "valid1@example.com"
	validEmail2 := "valid2@example.com"
	invalidEmail := "invalid"

	tests := []struct {
		name    string
		contact Contact
		wantErr bool
	}{
		{
			name: "valid contact with slice of emails",
			contact: Contact{
				Name:         "John Doe",
				EmailAddress: []string{"john@example.com", "doe@example.com"},
			},
			wantErr: false,
		},
		{
			name: "valid contact with empty email slice",
			contact: Contact{
				Name:         "Jane Doe",
				EmailAddress: []string{},
			},
			wantErr: false,
		},
		{
			name: "valid contact with nil email slice",
			contact: Contact{
				Name:         "Bob Smith",
				EmailAddress: nil,
			},
			wantErr: false,
		},
		{
			name: "invalid contact - one bad email in slice",
			contact: Contact{
				Name:         "Invalid User",
				EmailAddress: []string{"good@example.com", "bad-email"},
			},
			wantErr: true,
		},
		{
			name: "invalid contact - all bad emails",
			contact: Contact{
				Name:         "All Bad",
				EmailAddress: []string{"bad1", "bad2"},
			},
			wantErr: true,
		},
		{
			name: "valid contact with pointer slice - all valid",
			contact: Contact{
				Name:         "Pointer User",
				BackupEmails: []*string{&validEmail1, &validEmail2},
			},
			wantErr: false,
		},
		{
			name: "valid contact with pointer slice - nil element",
			contact: Contact{
				Name:         "Nil Pointer",
				BackupEmails: []*string{&validEmail1, nil, &validEmail2},
			},
			wantErr: false,
		},
		{
			name: "invalid contact - pointer slice with bad email",
			contact: Contact{
				Name:         "Bad Pointer",
				BackupEmails: []*string{&validEmail1, &invalidEmail},
			},
			wantErr: true,
		},
		{
			name: "missing required name",
			contact: Contact{
				Name:         "",
				EmailAddress: []string{"valid@example.com"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contact.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Contact.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
