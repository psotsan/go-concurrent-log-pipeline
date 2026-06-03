# Concurrent Log Pipeline

Concurrent CLI tool that reads log lines (timestamp|level|message), counts occurrences of each level (DEBUG, ERROR, INFO, WARN), and prints the summary.  
Supports reading from stdin or a file, configurable worker pool, and graceful handling of invalid lines.

---

## Quick start

```bash
# Build
go build -o log-pipeline

# Run with stdin
echo "2025-01-15T10:00:01Z|INFO|Login OK" | ./log-pipeline

# Run with a file
./log-pipeline -input app.log

Benchmark
go test -bench=. ./parse -v
```

---

## Log format
- **Timestamp** - RFC3339 (UTC): 2006-01-02T15:04:05Z
- **Level** - INFO, WARN, ERROR, DEBUG (case-insensitive)
- **Message** - Any string

Lines must contain exactly three fields separated by |.
Empty lines are ignored.
Malformed lines are skipped – errors are written to stderr.

---

## Flags
|Flag|Default|Description|
|----------|----------|----------|
|-input|(none)|Read from file instead of stdin|
|-workers|4|Concurrent parsing workers|
|-help|false|Usage help|

---

## Example input
```
2025-01-15T10:00:01Z|INFO|Login OK
2025-01-15T10:00:02Z|ERROR|DB timeout
2025-01-15T10:00:02Z|WRONG_LEVEL|This line is invalid
2025-01-15T10:00:03Z|INFO|Login OK
```

## Example output
```bash
$ ./log-pipeline -input app.log
DEBUG: 0
ERROR: 1
INFO: 2
WARN: 0
```

Invalid lines produce error messages on stderr:
```
line 3: Invalid log level "WRONG_LEVEL"
```

---

## Tests and benchmarks
```bash
# Run all tests
go test -v ./...

# Run tests with race detector
go test -race ./...

# Run benchmarks (100k lines) with different worker counts
go test -bench=. ./parse
```
The benchmark generates 100 000 valid log lines and measures throughput for 1, 4 and 16 workers.

## How it works

1. Lines are read asynchronously and pushed into a channel.

2. A configurable pool of workers consumes lines from the channel, parses each line, and updates a shared statistics map protected by a mutex.

3. Invalid lines are sent to a separate error channel instead of breaking the pipeline.

4. A dedicated error arranger goroutine collects errors, sorts them by line number, and writes them to stderr deterministically.

5. After all workers finish, the final statistics are printed to stdout.

This design pursues thread safety, preserve error ordering, and to make the tool robust for large log files.