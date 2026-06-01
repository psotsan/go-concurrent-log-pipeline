package core

import (
	"fmt"
	"strings"
	"time"
)

const (
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
	LevelDebug = "DEBUG"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Msg       string
}

func validateLevel(level string) (string, error) {
	level = strings.ToUpper(level)
	if level != LevelInfo && level != LevelWarn && level != LevelError && level != LevelDebug {
		return "", fmt.Errorf("Invalid log level %q", level)
	}
	return level, nil
}

func ParseLine(line string) (LogEntry, error) {
	var entry LogEntry
	var err error

	fields := strings.Split(line, "|")

	if len(fields) != 3 {
		return LogEntry{}, fmt.Errorf("%d fields found, expected 3", len(fields))
	}

	entry.Timestamp, err = time.Parse("2006-01-02T15:04:05Z", strings.Trim(fields[0], " "))
	if err != nil {
		return LogEntry{}, err
	}

	entry.Level, err = validateLevel(strings.Trim(fields[1], " "))
	if err != nil {
		return LogEntry{}, err
	}

	entry.Msg = strings.Trim(fields[2], " ")

	return entry, nil
}
