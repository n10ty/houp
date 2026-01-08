package pointers

// PointerFields demonstrates validation for pointer types
type PointerFields struct {
	Name  *string `json:"name" validate:"required"`
	Age   *int    `json:"age" validate:"omitempty,gt=0,lt=120"`
	Email *string `json:"email" validate:"omitempty"`
}

// MixedPointers has both pointer and non-pointer fields
type MixedPointers struct {
	ID       string  `json:"id" validate:"required"`
	Optional *string `json:"optional" validate:"omitempty,min=5"`
	Count    *int    `json:"count" validate:"omitempty,gte=1"`
}
