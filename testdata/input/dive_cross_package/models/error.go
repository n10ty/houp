package models

// Error represents an error from external package with validation
type Error struct {
	Code     string `json:"code,omitempty"`
	LangCode string `json:"langCode" validate:"required"`
	TypeCode string `json:"typeCode" validate:"required"`
}
