package main

// FixedPenalty represents a fixed penalty amount
type FixedPenalty struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Currency string  `json:"currency" validate:"required"`
}

// PercentagePenalty represents a percentage-based penalty
type PercentagePenalty struct {
	Percentage float64 `json:"percentage" validate:"required,gt=0,lte=100"`
}

// Penalty demonstrates required_without validation
// This ensures that at least one of FixedPenalty or PercentagePenalty is provided
type Penalty struct {
	FixedPenalty      *FixedPenalty      `json:"fixedPenalty,omitempty"      validate:"required_without=PercentagePenalty"`
	PercentagePenalty *PercentagePenalty `json:"percentagePenalty,omitempty" validate:"required_without=FixedPenalty"`
	Reason            string             `json:"reason" validate:"required"`
}
