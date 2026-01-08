package datetime

import (
	"testing"
)

func TestEventValidation(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
	}{
		{
			name: "valid event",
			event: Event{
				Name:      "Conference",
				StartTime: "2024-01-15T09:00:00Z",
				EndTime:   "2024-01-15T17:00:00Z",
				CreatedAt: "2024-01-01",
			},
			wantErr: false,
		},
		{
			name: "invalid start time format",
			event: Event{
				Name:      "Conference",
				StartTime: "2024-01-15",
				EndTime:   "2024-01-15T17:00:00Z",
				CreatedAt: "2024-01-01",
			},
			wantErr: true,
		},
		{
			name: "invalid created at format",
			event: Event{
				Name:      "Conference",
				StartTime: "2024-01-15T09:00:00Z",
				EndTime:   "2024-01-15T17:00:00Z",
				CreatedAt: "01/15/2024",
			},
			wantErr: true,
		},
		{
			name: "missing required name",
			event: Event{
				Name:      "",
				StartTime: "2024-01-15T09:00:00Z",
				EndTime:   "2024-01-15T17:00:00Z",
				CreatedAt: "2024-01-01",
			},
			wantErr: true,
		},
		{
			name: "valid with pointer field",
			event: Event{
				Name:      "Conference",
				StartTime: "2024-01-15T09:00:00Z",
				EndTime:   "2024-01-15T17:00:00Z",
				CreatedAt: "2024-01-01",
				UpdatedAt: stringPtr("2024-01-02T10:00:00Z"),
			},
			wantErr: false,
		},
		{
			name: "invalid pointer field",
			event: Event{
				Name:      "Conference",
				StartTime: "2024-01-15T09:00:00Z",
				EndTime:   "2024-01-15T17:00:00Z",
				CreatedAt: "2024-01-01",
				UpdatedAt: stringPtr("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateFormatsValidation(t *testing.T) {
	tests := []struct {
		name    string
		df      DateFormats
		wantErr bool
	}{
		{
			name: "valid formats",
			df: DateFormats{
				RFC3339:    "2024-01-15T09:00:00Z",
				DateOnly:   "2024-01-15",
				TimeOnly:   "09:30:45",
				CustomDate: "01/15/2024",
				UnixDate:   "Mon Jan 15 09:00:00 UTC 2024",
			},
			wantErr: false,
		},
		{
			name: "invalid RFC3339",
			df: DateFormats{
				RFC3339:    "not a date",
				DateOnly:   "2024-01-15",
				TimeOnly:   "09:30:45",
				CustomDate: "01/15/2024",
				UnixDate:   "Mon Jan 15 09:00:00 UTC 2024",
			},
			wantErr: true,
		},
		{
			name: "invalid custom date",
			df: DateFormats{
				RFC3339:    "2024-01-15T09:00:00Z",
				DateOnly:   "2024-01-15",
				TimeOnly:   "09:30:45",
				CustomDate: "2024-01-15",
				UnixDate:   "Mon Jan 15 09:00:00 UTC 2024",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.df.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DateFormats.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestCustomStringTypesValidation(t *testing.T) {
	tests := []struct {
		name    string
		cst     CustomStringTypes
		wantErr bool
	}{
		{
			name: "valid custom string types",
			cst: CustomStringTypes{
				Timestamp: "2024-01-15T09:00:00Z",
				Date:      "2024-01-15",
			},
			wantErr: false,
		},
		{
			name: "invalid timestamp format",
			cst: CustomStringTypes{
				Timestamp: "2024-01-15",
				Date:      "2024-01-15",
			},
			wantErr: true,
		},
		{
			name: "invalid date format",
			cst: CustomStringTypes{
				Timestamp: "2024-01-15T09:00:00Z",
				Date:      "01/15/2024",
			},
			wantErr: true,
		},
		{
			name: "valid with optional pointer",
			cst: CustomStringTypes{
				Timestamp:  "2024-01-15T09:00:00Z",
				Date:       "2024-01-15",
				OptionalTs: metadataTimestampPtr("2024-01-16T10:00:00Z"),
			},
			wantErr: false,
		},
		{
			name: "invalid optional pointer",
			cst: CustomStringTypes{
				Timestamp:  "2024-01-15T09:00:00Z",
				Date:       "2024-01-15",
				OptionalTs: metadataTimestampPtr("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cst.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CustomStringTypes.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func metadataTimestampPtr(s string) *MetadataTimestamp {
	ts := MetadataTimestamp(s)
	return &ts
}
