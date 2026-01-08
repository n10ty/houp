package unique

// User represents a user with unique fields
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Product represents a product
type Product struct {
	SKU  string `json:"sku"`
	Name string `json:"name"`
}

// UniqueValidation demonstrates unique constraint validation
type UniqueValidation struct {
	// Slice of structs with unique Email
	Users []User `json:"users" validate:"required,min=1,unique=Email"`

	// Slice of pointers with unique SKU
	Products []*Product `json:"products" validate:"unique=SKU"`

	// Slice of scalars - unique values
	Tags []string `json:"tags" validate:"unique"`

	// Slice of ints - unique values
	CategoryIDs []int `json:"categoryIds" validate:"min=1,unique"`
}
