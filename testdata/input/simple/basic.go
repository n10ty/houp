package simple

// BasicTypes demonstrates validation for basic scalar types
type BasicTypes struct {
	Name   string  `json:"name" validate:"required,min=3,max=50"`
	Age    int     `json:"age" validate:"gte=0,lte=150"`
	Email  string  `json:"email" validate:"required"`
	Score  float64 `json:"score" validate:"gt=0,lt=100"`
	Active bool    `json:"active"`
}

// MinMaxValidation tests min/max on different types
type MinMaxValidation struct {
	Username string `json:"username" validate:"min=3,max=20"`
	Count    int    `json:"count" validate:"min=1,max=1000"`
	Rating   int32  `json:"rating" validate:"min=1,max=5"`
}

// RequiredOnly tests simple required validation
type RequiredOnly struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}
