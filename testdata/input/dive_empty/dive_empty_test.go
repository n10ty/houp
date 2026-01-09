package dive_empty

import (
	"testing"
)

func TestCabinCompartmentValidation(t *testing.T) {
	// Test valid data
	cabin := &CabinCompartment{
		CabinComponent: []CabinComponent{
			{
				CabinComponentCode: "COMP1",
				ColumnIds:          []string{"A", "B"},
				FirstRowNumber:     1,
				LastRowNumber:      10,
			},
		},
		SeatColumns: []SeatColumn{
			{Code: "A"},
			{Code: "B"},
		},
		SeatRows: []SeatRow{
			{Number: 1},
			{Number: 2},
		},
	}

	if err := cabin.Validate(); err != nil {
		t.Errorf("Validation should pass: %v", err)
	}

	// Test missing required field
	invalidCabin := &CabinCompartment{
		CabinComponent: []CabinComponent{},
		SeatColumns:    []SeatColumn{},
		SeatRows:       []SeatRow{},
	}

	if err := invalidCabin.Validate(); err == nil {
		t.Error("Should have failed validation for missing SeatColumns")
	}
}
