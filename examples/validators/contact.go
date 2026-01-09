package main

type ContactInfo struct {
	Email   string `validate:"email"`
	Country string `validate:"iso3166_1_alpha2"`
}
