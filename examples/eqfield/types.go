package main

// UserRegistration demonstrates password confirmation with eqfield
type UserRegistration struct {
	Username        string `json:"username" validate:"required,min=3,max=20"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
}

// OrderModificationRequest demonstrates order cancellation with eqfield
type OrderModificationRequest struct {
	// CancelOrderId is used for full order cancellation.
	// If it is present, it must be equal OrderId to confirm the cancellation.
	CancelOrderId *string `json:"cancelOrderId,omitempty" validate:"omitempty,eqfield=OrderId"`
	OrderId       string  `json:"orderId" validate:"required"`
	Reason        string  `json:"reason,omitempty"`
}
