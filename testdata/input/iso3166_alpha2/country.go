package iso3166_alpha2

// Address represents an address with country code validation
type Address struct {
	Country         string  `validate:"iso3166_1_alpha2"`
	OptionalCountry *string `validate:"omitempty,iso3166_1_alpha2"`
}
