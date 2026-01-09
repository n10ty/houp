package uuid

// Resource demonstrates UUID validation
type Resource struct {
	ID         string  `json:"id" validate:"required,uuid"`
	OwnerID    string  `json:"owner_id" validate:"uuid"`
	OptionalID *string `json:"optional_id" validate:"omitempty,uuid"`
	Name       string  `json:"name" validate:"required"`
}

// MultipleUUIDs tests multiple UUID fields
type MultipleUUIDs struct {
	UserID    string `json:"user_id" validate:"required,uuid"`
	SessionID string `json:"session_id" validate:"required,uuid"`
	RequestID string `json:"request_id" validate:"uuid"`
	TraceID   string `json:"trace_id" validate:"uuid"`
}
