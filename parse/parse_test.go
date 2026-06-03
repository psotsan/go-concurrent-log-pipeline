package parse

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func genLargeData() string {
	var ret strings.Builder
	for range 100_000 {
		ret.WriteString("2025-01-15T10:00:01Z|INFO|Login OK\n")
	}

	return ret.String()
}

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
		want    logEntry
		wantErr error
	}{
		{
			name: "valid line",
			line: "2025-01-15T10:00:01Z|INFO|Login OK",
			want: logEntry{
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
			got, err := parseLine(tt.line)

			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("ParseLine (%q): expected %q; got %q", tt.name, tt.wantErr.Error(), err.Error())
			}

			if got != tt.want {
				t.Fatalf("ParseLine(%q): expected %v; got %v", tt.name, tt.want, got)
			}
		})
	}
}

func TestProcess(t *testing.T) {
	ps := []struct {
		name    string
		r       strings.Reader
		w       strings.Builder
		errW    strings.Builder
		want    string
		wantErr string
	}{
		{
			name: "valid line",
			r:    *strings.NewReader("2025-01-15T10:00:01Z|INFO|Login OK"),
			want: `DEBUG: 0
ERROR: 0
INFO: 1
WARN: 0
`,
		},
		{
			name: "multiple valid lines",
			r: *strings.NewReader(`2025-01-15T10:00:01Z|INFO|Login OK
2025-01-15T10:00:02Z|ERROR|DB timeout
2025-01-15T10:00:03Z|INFO|Login OK`),
			want: `DEBUG: 0
ERROR: 1
INFO: 2
WARN: 0
`,
		},
		{
			name: "multiple valid lines with empty lines",
			r: *strings.NewReader(`2025-01-15T10:00:01Z|INFO|Login OK

2025-01-15T10:00:02Z|ERROR|DB timeout

2025-01-15T10:00:03Z|INFO|Login OK`),
			want: `DEBUG: 0
ERROR: 1
INFO: 2
WARN: 0
`,
		},
		{
			name: "wrong line",
			r: *strings.NewReader(`2025-01-15T10:00:01Z|INFO|Login OK
2025-01-15T10:00:02Z|ERROR|DB timeout
2025-01-15T10:00:02Z|WRONG_LEVEL|DB timeout
2025-01-15T10:00:03Z|INFO|Login OK`),
			want: `DEBUG: 0
ERROR: 1
INFO: 2
WARN: 0
`,
			wantErr: "line 3: Invalid log level \"WRONG_LEVEL\"\n",
		},
	}

	for _, tt := range ps {
		t.Run(tt.name, func(t *testing.T) {
			Process(&tt.r, &tt.w, &tt.errW, 4)

			if tt.errW.String() != tt.wantErr {
				t.Fatalf("Process (%q): expected error output: %q; got %q", tt.name, tt.wantErr, &tt.errW)
			}

			if tt.w.String() != tt.want {
				t.Fatalf("Process (%q): expected:\n%s\ngot:\n%s ", tt.name, tt.want, tt.w.String())
			}
		})
	}
}

func BenchmarkProcess(b *testing.B) {
	largeData := genLargeData()
	cases := []struct {
		name    string
		workers int
		data    string
	}{
		{"workers=1", 1, largeData},
		{"workers=4", 4, largeData},
		{"workers=16", 16, largeData},
	}

	for _, tb := range cases {
		b.Run(tb.name, func(b *testing.B) {
			r := strings.NewReader(tb.data)
			var w, errW strings.Builder

			for range b.N {
				r.Reset(tb.data)
				w.Reset()
				errW.Reset()
				_ = Process(r, &w, &errW, tb.workers)
			}
		})
	}
}
