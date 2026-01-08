package dive

// Address represents an address
type Address struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	ZipCode string `json:"zipCode" validate:"required,min=5,max=10"`
}

// Contact represents contact information
type Contact struct {
	Email string `json:"email" validate:"required"`
	Phone string `json:"phone" validate:"omitempty,min=10"`
}

// Person demonstrates dive validation for nested structs
type Person struct {
	Name    string   `json:"name" validate:"required"`
	Address *Address `json:"address" validate:"required,dive"`
	Contact Contact  `json:"contact" validate:"dive"`
}

// Item represents an item in an order
type Item struct {
	Name     string  `json:"name" validate:"required"`
	Quantity int     `json:"quantity" validate:"min=1"`
	Price    float64 `json:"price" validate:"gt=0"`
}

// Order demonstrates dive validation for slices
type Order struct {
	ID    string `json:"id" validate:"required"`
	Items []Item `json:"items" validate:"required,min=1,dive"`
}

// Company demonstrates multi-level dive
type Company struct {
	Name      string   `json:"name" validate:"required"`
	Employees []Person `json:"employees" validate:"min=1,dive"`
	HQ        *Address `json:"hq" validate:"required,dive"`
}
