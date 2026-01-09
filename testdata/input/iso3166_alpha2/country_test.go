package iso3166_alpha2

import "testing"

func TestAddressValidate(t *testing.T) {
	tests := []struct {
		name    string
		address Address
		wantErr bool
	}{
		{
			name: "valid country code - US",
			address: Address{
				Country: "US",
			},
			wantErr: false,
		},
		{
			name: "valid country code - GB",
			address: Address{
				Country: "GB",
			},
			wantErr: false,
		},
		{
			name: "valid country code - UA",
			address: Address{
				Country: "UA",
			},
			wantErr: false,
		},
		{
			name: "valid country code - XK (Kosovo)",
			address: Address{
				Country: "XK",
			},
			wantErr: false,
		},
		{
			name: "invalid country code - lowercase",
			address: Address{
				Country: "us",
			},
			wantErr: true,
		},
		{
			name: "invalid country code - not exists",
			address: Address{
				Country: "XX",
			},
			wantErr: true,
		},
		{
			name: "invalid country code - too long",
			address: Address{
				Country: "USA",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.address.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Address.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddressValidateOptionalCountry(t *testing.T) {
	validCountry := "CA"
	invalidCountry := "ZZ"

	tests := []struct {
		name    string
		address Address
		wantErr bool
	}{
		{
			name: "nil optional country - should pass",
			address: Address{
				Country:         "US",
				OptionalCountry: nil,
			},
			wantErr: false,
		},
		{
			name: "valid optional country",
			address: Address{
				Country:         "US",
				OptionalCountry: &validCountry,
			},
			wantErr: false,
		},
		{
			name: "invalid optional country",
			address: Address{
				Country:         "US",
				OptionalCountry: &invalidCountry,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.address.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Address.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
