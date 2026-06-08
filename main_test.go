package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	helpStr := "usage: " + os.Args[0] + " [options]\nOptions:\n  -help\n    \tshow help\n  -input string\n    \tfile path\n  -workers int\n    \tconcurrent parsing workers (default 4)\n"
	rs := []struct {
		name       string
		args       []string
		r          strings.Reader
		w          strings.Builder
		errW       strings.Builder
		openFn     openFunc
		want       string
		wantErr    string
		wantReturn int
	}{
		{
			name:       "help flag",
			args:       []string{"-help"},
			wantErr:    helpStr,
			wantReturn: 0,
		},
		{
			name:       "valid line",
			r:          *strings.NewReader("2025-01-15T10:00:01Z|INFO|Login OK"),
			want:       "DEBUG: 0\nERROR: 0\nINFO: 1\nWARN: 0\n",
			wantReturn: 0,
		},
		{
			name: "simulated input file",
			args: []string{"-input=fake.log"},
			openFn: func(path string) (io.ReadCloser, error) {
				if path != "fake.log" {
					return nil, fmt.Errorf("unexpected path")
				}
				content := "2025-01-15T10:00:01Z|INFO|Login OK\n"
				return io.NopCloser(strings.NewReader(content)), nil
			},
			want:       "DEBUG: 0\nERROR: 0\nINFO: 1\nWARN: 0\n",
			wantReturn: 0,
		},
		{
			name: "input file not found",
			args: []string{"-input=missing.log"},
			openFn: func(path string) (io.ReadCloser, error) {
				return nil, os.ErrNotExist
			},
			wantErr:    "Could not open file missing.log\n",
			wantReturn: 1,
		},
		{
			name:       "workers flag used",
			args:       []string{"-workers=6"},
			r:          *strings.NewReader("2025-01-15T10:00:01Z|INFO|Login OK"),
			want:       "DEBUG: 0\nERROR: 0\nINFO: 1\nWARN: 0\n",
			wantReturn: 0,
		},
		{
			name:    "invalid lines",
			r:       *strings.NewReader("2025-01-15T10:00:01Z|DEBUG|Testing DB\n2025-01-15T10:00:01Z|DBG|Testing DB\n2025-01-15T10:00:01Z|INFO|Login OK\n2025-01-15T10:00:01Z|INFORM|Login OK\n"),
			want:    "DEBUG: 1\nERROR: 0\nINFO: 1\nWARN: 0\n",
			wantErr: "line 2: Invalid log level \"DBG\"\nline 4: Invalid log level \"INFORM\"\n",
		},
		{
			name:       "invalid argument",
			args:       []string{"-argument"},
			wantErr:    "flag provided but not defined: -argument\n" + helpStr,
			wantReturn: 1,
		},
		{
			name:       "no input and no flags",
			args:       []string{},
			r:          *strings.NewReader(""),
			want:       "DEBUG: 0\nERROR: 0\nINFO: 0\nWARN: 0\n",
			wantReturn: 0,
		},
		{
			name:       "0 workers",
			args:       []string{"-workers=0"},
			r:          *strings.NewReader("2025-01-15T10:00:01Z|INFO|Login OK"),
			want:       "DEBUG: 0\nERROR: 0\nINFO: 1\nWARN: 0\n",
			wantErr:    "workers must be > 0, falling back to 4 workers\n",
			wantReturn: 0,
		},
	}

	for _, tt := range rs {
		t.Run(tt.name, func(t *testing.T) {
			ret := run(tt.args, &tt.r, &tt.w, &tt.errW, tt.openFn)
			if ret != tt.wantReturn {
				t.Fatalf("run (%q): expected return %d; returned %d", tt.name, tt.wantReturn, ret)
			}
			if tt.errW.String() != tt.wantErr {
				t.Fatalf("run (%q): expected error:\n%s\ngot error:\n%s", tt.name, tt.wantErr, &tt.errW)
			}
			if tt.w.String() != tt.want {
				t.Fatalf("run (%q): expected:\n%s\ngot:\n%s", tt.name, tt.want, &tt.w)
			}
		})
	}
}
