package core

import (
	"bufio"
	"fmt"
	"io"
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

func Process(r io.Reader, w io.Writer, errW io.Writer, workers int) error {
	var err error
	stats := make(map[string]int)
	line := 0
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line++
		entry, err := ParseLine(scanner.Text())

		stats[entry.Level]++

		if err != nil {
			fmt.Fprintf(errW, "line %d: %s", line, err.Error())
		}
	}

	if scanner.Err() != nil {
		fmt.Fprintln(errW, scanner.Err())
	}

	fmt.Fprintf(w, "DEBUG: %d\nERROR: %d\nINFO: %d\nWARN: %d\n",
		stats[LevelDebug], stats[LevelError], stats[LevelInfo], stats[LevelWarn])

	return err
}
