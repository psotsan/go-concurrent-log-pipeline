package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/psotsan/go-concurrent-log-pipeline/parse"
)

const defaultWorkers = 4

type openFunc func(string) (io.ReadCloser, error)

type config struct {
	path    string
	workers int
	help    bool
}

func parseArgs(c *config, args []string, errW io.Writer) (*flag.FlagSet, error) {
	name := os.Args[0]
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(errW)

	fs.StringVar(&c.path, "input", "", "file path")
	fs.IntVar(&c.workers, "workers", defaultWorkers, "concurrent parsing workers")
	fs.BoolVar(&c.help, "help", false, "show help")

	fs.Usage = func() {
		fmt.Fprintf(errW, "usage: %s [options]\n", name)
		fmt.Fprintln(errW, "Options:")
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	return fs, err
}

func run(args []string, r io.Reader, w io.Writer, errW io.Writer, openFn openFunc) int {
	var c config

	fs, err := parseArgs(&c, args, errW)
	if err != nil {
		return 1
	}

	if c.help {
		fs.Usage()
		return 0
	}

	if c.path != "" {
		if openFn == nil {
			openFn = func(path string) (io.ReadCloser, error) {
				return os.Open(path)
			}
		}
		file, err := openFn(c.path)
		if err != nil {
			e := fmt.Errorf("Could not open file %s", c.path)
			fmt.Fprintln(errW, e)
			return 1
		}
		defer file.Close()
		r = file
	}

	if c.workers <= 0 {
		c.workers = defaultWorkers
		fmt.Fprintf(errW, "workers must be > 0, falling back to %d workers\n", c.workers)
	}

	err = parse.Process(r, w, errW, c.workers)
	if err != nil {
		return 1
	}
	return 0
}

func main() {
	args := make([]string, 0)

	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	exit := run(args, os.Stdin, os.Stdout, os.Stderr, nil)

	os.Exit(exit)
}
