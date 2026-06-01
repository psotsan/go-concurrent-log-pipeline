package core

import (
	"fmt"
	"testing"
	"time"
)

func TestValidateLevel(t *testing.T) {
	vls := []struct {
		name    string
		level   string
		want    string
		wantErr error
	}{
		{
			name:  "level info",
			level: "INFO",
			want:  "INFO",
		},
		{
			name:  "level warn",
			level: "WARN",
			want:  "WARN",
		},
		{
			name:  "level error",
			level: "ERROR",
			want:  "ERROR",
		},
		{
			name:  "level debug",
			level: "DEBUG",
			want:  "DEBUG",
		},
		{
			name:  "lower case",
			level: "info",
			want:  "INFO",
		},
		{
			name:    "non existant level",
			level:   "ULTRACRITICAL",
			wantErr: fmt.Errorf("Invalid log level %q", "ULTRACRITICAL"),
		},
	}

	for _, tt := range vls {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateLevel(tt.level)
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("ValidateLevel (%q): got error %q; expected %q", tt.name, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ValidateLevel (%q): got %q; expected %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestParseLine(t *testing.T) {
	validTime, _ := time.Parse("2006-01-02T15:04:05Z", "2025-01-15T10:00:01Z")

	pls := []struct {
		name    string
		line    string
		want    LogEntry
		wantErr error
	}{
		{
			name: "valid line",
			line: "2025-01-15T10:00:01Z|INFO|Login OK",
			want: LogEntry{
				Timestamp: validTime,
				Level:     LevelInfo,
				Msg:       "Login OK",
			},
		},
		{
			name:    "4 fields",
			line:    "2025-01-15T10:00:01Z|INFO|Login OK|extra field",
			wantErr: fmt.Errorf("4 fields found, expected 3"),
		},
		{
			name:    "2 fields",
			line:    "2025-01-15T10:00:01Z|INFO",
			wantErr: fmt.Errorf("2 fields found, expected 3"),
		},
		{
			name:    "invalid time",
			line:    "2025/01-15T10:00:01Z|INFO|Login OK",
			wantErr: fmt.Errorf("parsing time \"2025/01-15T10:00:01Z\" as \"2006-01-02T15:04:05Z\": cannot parse \"/01-15T10:00:01Z\" as \"-\""),
		},
		{
			name:    "invalid level",
			line:    "2025-01-15T10:00:01Z|INFORMATION|Login OK",
			wantErr: fmt.Errorf("Invalid log level \"INFORMATION\""),
		},
	}

	for _, tt := range pls {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLine(tt.line)

			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("ParseLine (%q): expected %q; got %q", tt.name, tt.wantErr.Error(), err.Error())
			}

			if got != tt.want {
				t.Fatalf("ParseLine(%q): expected %v; got %v", tt.name, tt.want, got)
			}
		})
	}
}
