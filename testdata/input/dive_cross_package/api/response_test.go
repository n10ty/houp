package api

import (
	"testing"

	"github.com/n10ty/houp/testdata/input/dive_cross_package/models"
)

func TestErrorRs_Validate(t *testing.T) {
	t.Run("Valid error list", func(t *testing.T) {
		validError := models.Error{
			Code:     "ERR001",
			LangCode: "en",
			TypeCode: "ERROR",
		}
		errorRs := ErrorRs{
			Errors:   []models.Error{validError},
			Metadata: "test",
		}

		err := errorRs.Validate()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid error item missing LangCode", func(t *testing.T) {
		invalidError := models.Error{
			Code:     "ERR001",
			LangCode: "",
			TypeCode: "ERROR",
		}
		errorRs := ErrorRs{
			Errors:   []models.Error{invalidError},
			Metadata: "test",
		}

		err := errorRs.Validate()
		if err == nil {
			t.Error("Expected validation error for missing LangCode, got nil")
		}
	})

	t.Run("Invalid error item missing TypeCode", func(t *testing.T) {
		invalidError := models.Error{
			Code:     "ERR001",
			LangCode: "en",
			TypeCode: "",
		}
		errorRs := ErrorRs{
			Errors:   []models.Error{invalidError},
			Metadata: "test",
		}

		err := errorRs.Validate()
		if err == nil {
			t.Error("Expected validation error for missing TypeCode, got nil")
		}
	})

	t.Run("Empty errors list", func(t *testing.T) {
		errorRs := ErrorRs{
			Errors:   []models.Error{},
			Metadata: "test",
		}

		err := errorRs.Validate()
		if err == nil {
			t.Error("Expected validation error for empty errors list, got nil")
		}
	})
}
