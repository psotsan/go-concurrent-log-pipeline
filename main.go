package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/psotsan/go-concurrent-log-pipeline/core"
)

const defaultWorkers = 4

type config struct {
	path    string
	workers int
	help    bool
}

func parseArgs(c *config, args []string, errW io.Writer) (*flag.FlagSet, error) {
	name := args[0]
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

func run(args []string, r io.Reader, w io.Writer, errW io.Writer) int {
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
		r, err = os.Open(c.path)
		if err != nil {
			e := fmt.Errorf("Could not open file %s", c.path)
			fmt.Fprintln(errW, e)
			return 1
		}
	}

	err = core.Process(r, w, errW, c.workers)
	if err != nil {
		return 1
	}
	return 0
}

func main() {
	args := os.Args[1:]
	r := os.Stdin
	w := os.Stdout
	errW := os.Stderr

	exit := run(args, r, w, errW)

	os.Exit(exit)
}
