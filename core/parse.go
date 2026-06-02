package core

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
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

type line struct {
	n   int
	str string
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

func readLines(r io.Reader) <-chan line {
	out := make(chan line)
	scanner := bufio.NewScanner(r)

	go func() {
		defer close(out)
		n := 0
		for scanner.Scan() {
			n++
			l := line{
				n:   n,
				str: scanner.Text(),
			}
			out <- l
		}
	}()

	return out
}

func Process(r io.Reader, w io.Writer, errW io.Writer, workers int) error {
	var err error
	var wg sync.WaitGroup
	stats := make(map[string]int)

	lines := readLines(r)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for l := range lines {
				entry, err := ParseLine(l.str)
				if err != nil {
					fmt.Fprintf(errW, "line %d: %s", l.n, err)
					return
				}
				stats[entry.Level]++
			}
		}()
	}
	wg.Wait()

	fmt.Fprintf(w, "DEBUG: %d\nERROR: %d\nINFO: %d\nWARN: %d\n",
		stats[LevelDebug], stats[LevelError], stats[LevelInfo], stats[LevelWarn])

	return err
}
