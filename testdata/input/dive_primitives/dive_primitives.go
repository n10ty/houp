package dive_primitives

// EmailList demonstrates dive with email validation on string slice
type EmailList struct {
	Emails []string `json:"emails" validate:"required,dive,email"`
}

// NumberList demonstrates dive with min validation on int slice
type NumberList struct {
	Numbers []int `json:"numbers" validate:"omitempty,dive,min=1,max=100"`
}
