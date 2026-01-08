package slices

// SliceValidation demonstrates validation for slice types
type SliceValidation struct {
	Tags       []string `json:"tags" validate:"required,min=1,max=10"`
	Categories []string `json:"categories" validate:"omitempty,max=5"`
	Numbers    []int    `json:"numbers" validate:"min=1"`
}

// SliceOfPointers has slices of pointer types
type SliceOfPointers struct {
	Items []*string `json:"items" validate:"required,min=1"`
	IDs   []*int    `json:"ids" validate:"omitempty"`
}
