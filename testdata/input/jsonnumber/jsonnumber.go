package jsonnumber

import "encoding/json"

// JSONNumberValidation demonstrates validation for json.Number type
type JSONNumberValidation struct {
	Price    json.Number `json:"price" validate:"gte=0,lte=999999"`
	Quantity json.Number `json:"quantity" validate:"min=1,max=1000"`
	Discount json.Number `json:"discount" validate:"gt=0,lt=100"`
	Rating   json.Number `json:"rating" validate:"gte=1,lte=5"`
}

// JSONNumberPointer tests validation for pointer to json.Number
type JSONNumberPointer struct {
	Amount *json.Number `json:"amount" validate:"gte=0"`
}
