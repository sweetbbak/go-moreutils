package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Count      int    `short:"n" long:"head-count" default:"0" description:"output at most N number of lines"`
	Outfile    string `short:"o" long:"output" description:"write results to FILE instead of stdout"`
	InputRange string `short:"i" long:"input-range" description:"treat a range of numbers as input, must be two ints separated by a '-' ie 1-100 [--input-range=LO-HI]"`
	Echo       bool   `short:"e" long:"echo" description:"treat CLI arguments as input to be shuffled ie: (shuf a b c 1 2 4)"`
	NonEmpty   bool   `short:"b" long:"no-blank" description:"dont treat blank lines or lines that are all spaces as input"`
	Repeat     bool   `short:"r" long:"repeat" description:"output lines can be repeated"`
	Zero       bool   `short:"z" long:"zero-terminated" description:"line delimiter is NUL, not a newline"`
	Verbose    bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func randomShuf(file *os.File) error {
	var lines [][]byte
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, []byte(scanner.Text()))
	}

	Debug("length of input: %v\n", len(lines))

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	rand.Shuffle(len(lines), func(i, j int) {
		lines[i], lines[j] = lines[j], lines[i]
	})

	delim := "\n"
	if opts.Zero {
		delim = "\x00"
	}

	var f *os.File
	var err error

	if opts.Outfile != "" {
		f, err = os.Open(opts.Outfile)
		if err != nil {
			return err
		}
	} else {
		f = os.Stdout
	}

	if opts.Count > 0 {
		if opts.Count > len(lines) {
			opts.Count = len(lines)
		}
		// this was a "simple" way to only print N amount of lines
		// but it falls apart with non-repeats and no empty lines
		// lines = lines[:opts.Count]
	}

	if !opts.Repeat {
		m := make(map[string]bool)
		for _, x := range lines {
			m[string(x)] = false
		}

		if opts.Count == 0 {
			opts.Count = len(m) - 1
		}

		var printCount int = 0
		for k := range m {
			if printCount == opts.Count {
				break
			}

			// check if our string -> bool map has been printed yet, if not flip the str to true and print
			exists := m[k]
			if !exists {
				if !opts.NonEmpty {
					if k != "" && k != "\n" && len(strings.TrimSpace(k)) != 0 {
						fmt.Fprintf(f, "%s%s", k, delim)
						printCount++
					} else {
						continue
					}
				} else {
					fmt.Fprintf(f, "%s%s", k, delim)
					printCount++
				}
				m[k] = true
			}
		}
	}

	if opts.Repeat {
		var printCount int = 0
		for _, line := range lines {
			if printCount == opts.Count {
				break
			}

			if !opts.NonEmpty {
				if line != nil && len(bytes.TrimSpace(line)) != 0 {
					fmt.Fprintf(f, "%s%s", line, delim)
					Debug("%x\n", line)
					printCount++
				}
			} else {
				Debug("%d\n", line)
				fmt.Fprintf(f, "%s%s", line, delim)
				printCount++
			}
		}
	}
	return nil
}

func isatty() bool {
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice == os.ModeCharDevice {
		return true // data is being piped in
	} else {
		return false // data is NOT being piped in
	}
}

func Shuf(args []string) error {
	if !isatty() {
		Debug("STDIN is open\n")
		randomShuf(os.Stdin)
	}

	for _, file := range args {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		Debug("Shuffling: %v\n", f.Name())
		randomShuf(f)
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) || err == flags.ErrHelp {
		os.Exit(0)
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Shuf(args); err != nil {
		log.Fatal(err)
	}
}
