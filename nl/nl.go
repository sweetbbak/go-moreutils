package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Width     int    `short:"w" long:"width" description:"width of line numbers"`
	Separator string `short:"s" long:"separator" description:"use string as a separator"`
	Verbose   bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func nl(file *os.File, start int, width int, sep string) (int, error) {
	sc := bufio.NewScanner(file)
	n := start
	for sc.Scan() {
		fmt.Printf("%*d%s %s\n", width, n, sep, sc.Text())
		n++
	}

	if err := sc.Err(); err != nil {
		return -1, err
	}
	return n, nil
}

func NumberLines(args []string) error {
	if len(args) == 0 {
		if _, err := nl(os.Stdin, 1, opts.Width, opts.Separator); err != nil {
			return fmt.Errorf("error processing input: %v", err)
		}
		return nil
	}

	start := 1
	for _, file := range args {
		path := os.ExpandEnv(file)

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		n, err := nl(f, start, opts.Width, opts.Separator)
		if err != nil {
			return fmt.Errorf("error processing input: %v", err)
		}
		start = n
	}
	return nil
}

func main() {
	opts.Width = 4
	opts.Separator = " "

	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := NumberLines(args); err != nil {
		log.Fatal(err)
	}
}
