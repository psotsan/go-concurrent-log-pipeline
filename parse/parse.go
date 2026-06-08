package parse

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"slices"
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

type logEntry struct {
	Timestamp time.Time
	Level     string
	Msg       string
}

type line struct {
	n   int
	str string
}

type stats struct {
	s  map[string]int
	mu sync.Mutex
}

type errStruct struct {
	str  string
	line int
}

func validateLevel(level string) (string, error) {
	level = strings.ToUpper(level)
	if level != LevelInfo && level != LevelWarn && level != LevelError && level != LevelDebug {
		return "", fmt.Errorf("invalid log level %q", level)
	}
	return level, nil
}

func parseLine(line string) (logEntry, error) {
	var entry logEntry
	var err error

	fields := strings.Split(line, "|")

	if len(fields) != 3 {
		return logEntry{}, fmt.Errorf("%d fields found, expected 3", len(fields))
	}

	entry.Timestamp, err = time.Parse("2006-01-02T15:04:05Z", strings.Trim(fields[0], " "))
	if err != nil {
		return logEntry{}, err
	}

	entry.Level, err = validateLevel(strings.Trim(fields[1], " "))
	if err != nil {
		return logEntry{}, err
	}

	entry.Msg = strings.Trim(fields[2], " ")

	return entry, nil
}

func readLines(r io.Reader) <-chan line {
	// reads lines from the input concurrently with parsing already read lines
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

func processLines(lines <-chan line, errors chan<- errStruct, stats *stats, wg *sync.WaitGroup) {
	// worker function to parse the input lines concurrently
	defer wg.Done()
	for l := range lines {
		if strings.TrimSpace(l.str) == "" {
			continue
		}
		entry, err := parseLine(l.str)
		if err != nil {
			errors <- errStruct{
				str:  fmt.Sprintf("line %d: %s", l.n, err),
				line: l.n,
			}
			continue
		}
		stats.mu.Lock()
		stats.s[entry.Level]++
		stats.mu.Unlock()
	}
}

func arrangeErrors(errors <-chan errStruct, errW io.Writer, done chan<- error) {
	// concurrently reads errors from parser workers to write the errors deterministically ordered
	var err error
	defer func() {
		done <- err
	}()

	errSlice := make([]errStruct, 0)
	for e := range errors {
		errSlice = append(errSlice, e)
	}

	slices.SortFunc(errSlice, func(a, b errStruct) int {
		return cmp.Compare(a.line, b.line)
	})

	for _, e := range errSlice {
		if _, err = fmt.Fprintln(errW, e.str); err != nil {
			return
		}
	}
}

// Process orchestrates the reading from the input, the output and error output. It waits until arrangeErrors ends to return.
func Process(r io.Reader, w io.Writer, errW io.Writer, workers int) error {
	var err error
	var wg sync.WaitGroup

	st := stats{
		s: make(map[string]int),
	}

	errors := make(chan errStruct)
	arrangeDone := make(chan struct{})

	lines := readLines(r)

	go arrangeErrors(errors, errW, arrangeDone)

	for range workers {
		wg.Add(1)
		go processLines(lines, errors, &st, &wg)
	}
	wg.Wait()
	close(errors)
	<-arrangeDone

	_, err = fmt.Fprintf(w, "DEBUG: %d\nERROR: %d\nINFO: %d\nWARN: %d\n",
		st.s[LevelDebug], st.s[LevelError], st.s[LevelInfo], st.s[LevelWarn])

	return err
}
