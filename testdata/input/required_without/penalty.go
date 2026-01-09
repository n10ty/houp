package required_without

// FixedPenalty represents a fixed penalty amount
type FixedPenalty struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Currency string  `json:"currency" validate:"required,iso4217"`
}

// PercentagePenalty represents a percentage-based penalty
type PercentagePenalty struct {
	Percentage float64 `json:"percentage" validate:"required,gt=0,lte=100"`
}

// Penalty demonstrates required_without validation
// Either FixedPenalty OR PercentagePenalty must be provided
type Penalty struct {
	FixedPenalty      *FixedPenalty      `json:"fixedPenalty,omitempty"      validate:"required_without=PercentagePenalty"`
	PercentagePenalty *PercentagePenalty `json:"percentagePenalty,omitempty" validate:"required_without=FixedPenalty"`
}

// Payment demonstrates multiple required_without scenarios
type Payment struct {
	CreditCard  *string `json:"creditCard,omitempty"  validate:"required_without=BankAccount"`
	BankAccount *string `json:"bankAccount,omitempty" validate:"required_without=CreditCard"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
}
