package uuid

import (
	"testing"
)

func TestResourceValidation(t *testing.T) {
	tests := []struct {
		name    string
		res     Resource
		wantErr bool
	}{
		{
			name: "valid resource",
			res: Resource{
				ID:      "123e4567-e89b-12d3-a456-426614174000",
				OwnerID: "550e8400-e29b-41d4-a716-446655440000",
				Name:    "Test Resource",
			},
			wantErr: false,
		},
		{
			name: "invalid UUID format - missing dashes",
			res: Resource{
				ID:      "123e4567e89b12d3a456426614174000",
				OwnerID: "550e8400-e29b-41d4-a716-446655440000",
				Name:    "Test Resource",
			},
			wantErr: true,
		},
		{
			name: "invalid UUID - wrong version",
			res: Resource{
				ID:      "123e4567-e89b-02d3-a456-426614174000", // version 0 is invalid
				OwnerID: "550e8400-e29b-41d4-a716-446655440000",
				Name:    "Test Resource",
			},
			wantErr: true,
		},
		{
			name: "invalid UUID - wrong variant",
			res: Resource{
				ID:      "123e4567-e89b-12d3-1456-426614174000", // variant 1 is invalid
				OwnerID: "550e8400-e29b-41d4-a716-446655440000",
				Name:    "Test Resource",
			},
			wantErr: true,
		},
		{
			name: "missing required ID",
			res: Resource{
				ID:      "",
				OwnerID: "550e8400-e29b-41d4-a716-446655440000",
				Name:    "Test Resource",
			},
			wantErr: true,
		},
		{
			name: "missing required name",
			res: Resource{
				ID:      "123e4567-e89b-12d3-a456-426614174000",
				OwnerID: "550e8400-e29b-41d4-a716-446655440000",
				Name:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Resource.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceOptionalID(t *testing.T) {
	validUUID := "123e4567-e89b-12d3-a456-426614174000"
	invalidUUID := "not-a-uuid"

	tests := []struct {
		name    string
		res     Resource
		wantErr bool
	}{
		{
			name: "nil optional ID is valid",
			res: Resource{
				ID:         "123e4567-e89b-12d3-a456-426614174000",
				OwnerID:    "550e8400-e29b-41d4-a716-446655440000",
				OptionalID: nil,
				Name:       "Test",
			},
			wantErr: false,
		},
		{
			name: "valid optional ID",
			res: Resource{
				ID:         "123e4567-e89b-12d3-a456-426614174000",
				OwnerID:    "550e8400-e29b-41d4-a716-446655440000",
				OptionalID: &validUUID,
				Name:       "Test",
			},
			wantErr: false,
		},
		{
			name: "invalid optional ID",
			res: Resource{
				ID:         "123e4567-e89b-12d3-a456-426614174000",
				OwnerID:    "550e8400-e29b-41d4-a716-446655440000",
				OptionalID: &invalidUUID,
				Name:       "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Resource.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMultipleUUIDsValidation(t *testing.T) {
	tests := []struct {
		name    string
		m       MultipleUUIDs
		wantErr bool
	}{
		{
			name: "all valid UUIDs",
			m: MultipleUUIDs{
				UserID:    "123e4567-e89b-12d3-a456-426614174000",
				SessionID: "550e8400-e29b-41d4-a716-446655440000",
				RequestID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				TraceID:   "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			},
			wantErr: false,
		},
		{
			name: "invalid UserID",
			m: MultipleUUIDs{
				UserID:    "invalid-uuid",
				SessionID: "550e8400-e29b-41d4-a716-446655440000",
				RequestID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				TraceID:   "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MultipleUUIDs.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
