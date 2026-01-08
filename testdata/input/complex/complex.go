package complex

// ComplexValidation demonstrates combination of multiple validators
type ComplexValidation struct {
	// Required with min/max
	Username string `json:"username" validate:"required,min=3,max=20"`

	// Omitempty with range
	Age *int `json:"age" validate:"omitempty,gte=18,lte=100"`

	// Slice with unique and min
	Tags []string `json:"tags" validate:"required,min=1,max=10,unique"`

	// Nested struct with dive
	Profile *Profile `json:"profile" validate:"required,dive"`

	// Slice of structs with dive and unique
	Items []Item `json:"items" validate:"min=1,dive,unique=Code"`
}

// Profile is a nested struct
type Profile struct {
	Bio       string `json:"bio" validate:"required,max=500"`
	Website   string `json:"website" validate:"omitempty,min=10"`
	AvatarURL string `json:"avatarUrl" validate:"omitempty"`
}

// Item represents an item with a unique code
type Item struct {
	Code        string `json:"code"`
	Description string `json:"description" validate:"required"`
	Price       int    `json:"price" validate:"gt=0"`
}
