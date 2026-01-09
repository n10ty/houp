package eqfield

import (
	"testing"
)

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     Request
		wantErr bool
	}{
		{
			name: "valid - CancelOrderId matches OrderId",
			req: Request{
				CancelOrderId: strPtr("12345"),
				OrderId:       "12345",
			},
			wantErr: false,
		},
		{
			name: "valid - CancelOrderId is nil",
			req: Request{
				CancelOrderId: nil,
				OrderId:       "12345",
			},
			wantErr: false,
		},
		{
			name: "invalid - CancelOrderId does not match OrderId",
			req: Request{
				CancelOrderId: strPtr("12345"),
				OrderId:       "67890",
			},
			wantErr: true,
		},
		{
			name: "invalid - OrderId is empty",
			req: Request{
				CancelOrderId: nil,
				OrderId:       "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Request.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserPasswordConfirmValidation(t *testing.T) {
	tests := []struct {
		name    string
		user    UserPasswordConfirm
		wantErr bool
	}{
		{
			name: "valid - passwords match",
			user: UserPasswordConfirm{
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: false,
		},
		{
			name: "invalid - passwords do not match",
			user: UserPasswordConfirm{
				Password:        "password123",
				ConfirmPassword: "password456",
			},
			wantErr: true,
		},
		{
			name: "invalid - password too short",
			user: UserPasswordConfirm{
				Password:        "pass",
				ConfirmPassword: "pass",
			},
			wantErr: true,
		},
		{
			name: "invalid - password is empty",
			user: UserPasswordConfirm{
				Password:        "",
				ConfirmPassword: "password123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("UserPasswordConfirm.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMixedPointersValidation(t *testing.T) {
	tests := []struct {
		name    string
		mixed   MixedPointers
		wantErr bool
	}{
		{
			name: "valid - pointer value matches non-pointer",
			mixed: MixedPointers{
				Value1: intPtr(42),
				Value2: 42,
			},
			wantErr: false,
		},
		{
			name: "valid - pointer is nil",
			mixed: MixedPointers{
				Value1: nil,
				Value2: 42,
			},
			wantErr: false,
		},
		{
			name: "invalid - pointer value does not match",
			mixed: MixedPointers{
				Value1: intPtr(10),
				Value2: 42,
			},
			wantErr: true,
		},
		{
			name: "invalid - Value2 is zero",
			mixed: MixedPointers{
				Value1: nil,
				Value2: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mixed.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MixedPointers.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBothPointersValidation(t *testing.T) {
	tests := []struct {
		name    string
		both    BothPointers
		wantErr bool
	}{
		{
			name: "valid - both pointers match",
			both: BothPointers{
				Field1: strPtr("test"),
				Field2: strPtr("test"),
			},
			wantErr: false,
		},
		{
			name: "valid - Field1 is nil",
			both: BothPointers{
				Field1: nil,
				Field2: strPtr("test"),
			},
			wantErr: false,
		},
		{
			name: "valid - both are nil",
			both: BothPointers{
				Field1: nil,
				Field2: nil,
			},
			wantErr: false,
		},
		{
			name: "invalid - values do not match",
			both: BothPointers{
				Field1: strPtr("test1"),
				Field2: strPtr("test2"),
			},
			wantErr: true,
		},
		{
			name: "invalid - Field2 is nil but Field1 is not",
			both: BothPointers{
				Field1: strPtr("test"),
				Field2: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.both.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BothPointers.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
