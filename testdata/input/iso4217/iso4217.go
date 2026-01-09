package iso4217

// Payment demonstrates ISO 4217 currency code validation
type Payment struct {
	Currency       string  `json:"currency" validate:"required,iso4217"`
	BaseCurrency   string  `json:"base_currency" validate:"iso4217"`
	TargetCurrency *string `json:"target_currency" validate:"omitempty,iso4217"`
	Amount         float64 `json:"amount" validate:"required,gt=0"`
}

// MultiCurrencyTransaction tests multiple currency fields
type MultiCurrencyTransaction struct {
	FromCurrency string `json:"from_currency" validate:"required,iso4217"`
	ToCurrency   string `json:"to_currency" validate:"required,iso4217"`
	FeeCurrency  string `json:"fee_currency" validate:"iso4217"`
}
