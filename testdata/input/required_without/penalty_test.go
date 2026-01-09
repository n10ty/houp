package required_without

import (
	"testing"
)

func TestPenaltyValidation(t *testing.T) {
	tests := []struct {
		name    string
		penalty Penalty
		wantErr bool
	}{
		{
			name: "valid with FixedPenalty",
			penalty: Penalty{
				FixedPenalty: &FixedPenalty{
					Amount:   100.0,
					Currency: "USD",
				},
			},
			wantErr: false,
		},
		{
			name: "valid with PercentagePenalty",
			penalty: Penalty{
				PercentagePenalty: &PercentagePenalty{
					Percentage: 10.0,
				},
			},
			wantErr: false,
		},
		{
			name: "valid with both penalties",
			penalty: Penalty{
				FixedPenalty: &FixedPenalty{
					Amount:   100.0,
					Currency: "USD",
				},
				PercentagePenalty: &PercentagePenalty{
					Percentage: 10.0,
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid with neither penalty",
			penalty: Penalty{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.penalty.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Penalty.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentValidation(t *testing.T) {
	creditCard := "1234-5678-9012-3456"
	bankAccount := "DE89370400440532013000"

	tests := []struct {
		name    string
		payment Payment
		wantErr bool
	}{
		{
			name: "valid with CreditCard",
			payment: Payment{
				CreditCard: &creditCard,
				Amount:     100.0,
			},
			wantErr: false,
		},
		{
			name: "valid with BankAccount",
			payment: Payment{
				BankAccount: &bankAccount,
				Amount:      100.0,
			},
			wantErr: false,
		},
		{
			name: "valid with both",
			payment: Payment{
				CreditCard:  &creditCard,
				BankAccount: &bankAccount,
				Amount:      100.0,
			},
			wantErr: false,
		},
		{
			name: "invalid with neither",
			payment: Payment{
				Amount: 100.0,
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
