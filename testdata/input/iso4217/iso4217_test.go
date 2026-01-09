package iso4217

import (
	"testing"
)

func TestPaymentValidation(t *testing.T) {
	tests := []struct {
		name    string
		payment Payment
		wantErr bool
	}{
		{
			name: "valid payment with USD",
			payment: Payment{
				Currency:     "USD",
				BaseCurrency: "EUR",
				Amount:       100.50,
			},
			wantErr: false,
		},
		{
			name: "valid payment with JPY",
			payment: Payment{
				Currency:     "JPY",
				BaseCurrency: "GBP",
				Amount:       10000,
			},
			wantErr: false,
		},
		{
			name: "invalid currency code",
			payment: Payment{
				Currency:     "XXZ",
				BaseCurrency: "EUR",
				Amount:       100.50,
			},
			wantErr: true,
		},
		{
			name: "invalid base currency code",
			payment: Payment{
				Currency:     "USD",
				BaseCurrency: "INVALID",
				Amount:       100.50,
			},
			wantErr: true,
		},
		{
			name: "lowercase currency code",
			payment: Payment{
				Currency:     "usd",
				BaseCurrency: "EUR",
				Amount:       100.50,
			},
			wantErr: true,
		},
		{
			name: "missing required currency",
			payment: Payment{
				Currency:     "",
				BaseCurrency: "EUR",
				Amount:       100.50,
			},
			wantErr: true,
		},
		{
			name: "valid XCD (East Caribbean Dollar)",
			payment: Payment{
				Currency:     "XCD",
				BaseCurrency: "EUR",
				Amount:       100.50,
			},
			wantErr: false,
		},
		{
			name: "valid CHF (Swiss Franc)",
			payment: Payment{
				Currency:     "CHF",
				BaseCurrency: "EUR",
				Amount:       100.50,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payment.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Payment.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentOptionalTargetCurrency(t *testing.T) {
	validCurrency := "EUR"
	invalidCurrency := "INVALID"

	tests := []struct {
		name    string
		payment Payment
		wantErr bool
	}{
		{
			name: "nil target currency is valid",
			payment: Payment{
				Currency:       "USD",
				BaseCurrency:   "EUR",
				TargetCurrency: nil,
				Amount:         100.50,
			},
			wantErr: false,
		},
		{
			name: "valid target currency",
			payment: Payment{
				Currency:       "USD",
				BaseCurrency:   "GBP",
				TargetCurrency: &validCurrency,
				Amount:         100.50,
			},
			wantErr: false,
		},
		{
			name: "invalid target currency",
			payment: Payment{
				Currency:       "USD",
				BaseCurrency:   "GBP",
				TargetCurrency: &invalidCurrency,
				Amount:         100.50,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payment.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Payment.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMultiCurrencyTransactionValidation(t *testing.T) {
	tests := []struct {
		name    string
		trans   MultiCurrencyTransaction
		wantErr bool
	}{
		{
			name: "all valid currencies",
			trans: MultiCurrencyTransaction{
				FromCurrency: "USD",
				ToCurrency:   "EUR",
				FeeCurrency:  "GBP",
			},
			wantErr: false,
		},
		{
			name: "invalid from currency",
			trans: MultiCurrencyTransaction{
				FromCurrency: "INVALID",
				ToCurrency:   "EUR",
				FeeCurrency:  "GBP",
			},
			wantErr: true,
		},
		{
			name: "invalid to currency",
			trans: MultiCurrencyTransaction{
				FromCurrency: "USD",
				ToCurrency:   "BADCODE",
				FeeCurrency:  "GBP",
			},
			wantErr: true,
		},
		{
			name: "exotic currencies",
			trans: MultiCurrencyTransaction{
				FromCurrency: "XAU", // Gold
				ToCurrency:   "XAG", // Silver
				FeeCurrency:  "XPT", // Platinum
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trans.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MultiCurrencyTransaction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
