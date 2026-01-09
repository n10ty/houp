package api

import (
	"github.com/n10ty/houp/testdata/input/dive_cross_package/models"
)

// ErrorRs uses dive to validate nested struct from different package
type ErrorRs struct {
	Errors   []models.Error `json:"errors" validate:"required,dive"`
	Metadata string         `json:"metadata"`
}
