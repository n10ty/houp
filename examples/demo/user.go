package demo

// User demonstrates basic validation
type User struct {
	ID       string   `json:"id" validate:"required"`
	Username string   `json:"username" validate:"required,min=3,max=20"`
	Email    string   `json:"email" validate:"required"`
	Age      int      `json:"age" validate:"gte=18,lte=100"`
	Tags     []string `json:"tags" validate:"min=1,max=5,unique"`
	Profile  *Profile `json:"profile" validate:"required,dive"`
}

// Profile is a nested struct
type Profile struct {
	Bio       string `json:"bio" validate:"required,max=500"`
	Website   string `json:"website" validate:"omitempty,min=10"`
	AvatarURL string `json:"avatarUrl" validate:"omitempty"`
}
