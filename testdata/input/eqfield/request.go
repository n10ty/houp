package eqfield

// Request demonstrates eqfield validation similar to the user's example
type Request struct {
	// CancelOrderId is used for full order cancellation.
	// If it is present, it must be equal OrderId.
	CancelOrderId *string `json:"cancelOrderId,omitempty" validate:"omitempty,eqfield=OrderId"`
	OrderId       string  `json:"orderId" validate:"required"`
}

// UserPasswordConfirm demonstrates eqfield with non-pointer fields
type UserPasswordConfirm struct {
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
}

// MixedPointers demonstrates eqfield with mixed pointer/non-pointer fields
type MixedPointers struct {
	Value1 *int `json:"value1" validate:"omitempty,eqfield=Value2"`
	Value2 int  `json:"value2" validate:"required"`
}

// BothPointers demonstrates eqfield with both fields as pointers
type BothPointers struct {
	Field1 *string `json:"field1" validate:"omitempty,eqfield=Field2"`
	Field2 *string `json:"field2"`
}
