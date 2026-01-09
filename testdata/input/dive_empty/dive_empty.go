package dive_empty

// CabinComponent has NO validation tags on any field
type CabinComponent struct {
	CabinComponentCode string                 `json:"cabinComponentCode"`
	ColumnIds          []string               `json:"columnIds"`
	FirstRowNumber     int                    `json:"firstRowNumber"`
	LastRowNumber      int                    `json:"lastRowNumber"`
	PositionCode       *ComponentPositionCode `json:"positionCode"`
}

// ComponentPositionCode is a simple type alias with no validation
type ComponentPositionCode string

// SeatColumn also has NO validation tags
type SeatColumn struct {
	Code string `json:"code"`
}

// SeatRow also has NO validation tags
type SeatRow struct {
	Number int `json:"number"`
}

// CabinCompartment references structs with dive but those structs have no validation
type CabinCompartment struct {
	CabinComponent []CabinComponent `json:"cabinComponent" validate:"omitempty,dive"`
	FirstRowNumber int              `json:"firstRowNumber,omitempty"`
	LastRowNumber  int              `json:"lastRowNumber,omitempty"`
	DeckCode       string           `json:"deckCode,omitempty"`
	SeatColumns    []SeatColumn     `json:"seatColumns" validate:"required,dive"`
	SeatRows       []SeatRow        `json:"seatRows" validate:"required,dive"`
}
