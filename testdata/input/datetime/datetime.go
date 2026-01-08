package datetime

// Custom string types for datetime validation
type MetadataTimestamp string
type ISODate string

// Event demonstrates datetime validation
type Event struct {
	Name      string  `json:"name" validate:"required"`
	StartTime string  `json:"start_time" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	EndTime   string  `json:"end_time" validate:"datetime=2006-01-02T15:04:05Z07:00"`
	CreatedAt string  `json:"created_at" validate:"datetime=2006-01-02"`
	UpdatedAt *string `json:"updated_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

// DateFormats tests various date formats
type DateFormats struct {
	RFC3339    string `json:"rfc3339" validate:"datetime=2006-01-02T15:04:05Z07:00"`
	DateOnly   string `json:"date_only" validate:"datetime=2006-01-02"`
	TimeOnly   string `json:"time_only" validate:"datetime=15:04:05"`
	CustomDate string `json:"custom_date" validate:"datetime=01/02/2006"`
	UnixDate   string `json:"unix_date" validate:"datetime=Mon Jan _2 15:04:05 MST 2006"`
}

// CustomStringTypes tests datetime validation with custom string types
type CustomStringTypes struct {
	Timestamp  MetadataTimestamp  `json:"timestamp" validate:"datetime=2006-01-02T15:04:05Z07:00"`
	Date       ISODate            `json:"date" validate:"datetime=2006-01-02"`
	OptionalTs *MetadataTimestamp `json:"optional_ts" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}
