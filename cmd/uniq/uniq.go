package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Count      bool `short:"c" long:"count" description:"print the count of matched lines"`
	Unique     bool `short:"u" long:"unique" description:"print unique lines"`
	Repeats    bool `short:"r" long:"repeat" description:"print only duplicate lines"`
	IgnoreCase bool `short:"i" long:"ignore-case" description:"ignore upper and lower case differences"`
	Verbose    bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func printout(count int, last string, duplicate bool) {
	if duplicate {
		if opts.Repeats {
			if opts.Count {
				fmt.Printf("%-10d %s\n", count, last)
			} else {
				fmt.Println(last)
			}
		}
	} else {
		if opts.Unique {
			if opts.Count {
				fmt.Printf("%-10d %s\n", count, last)
			} else {
				fmt.Println(last)
			}
		}
	}
}

func uniqFile(f io.Reader) error {
	scanner := bufio.NewScanner(f)
	var count int
	var duplicate bool

	// get first line to inlcusively compare it to the rest
	scanner.Scan()
	last := scanner.Text()
	count = 1

	for scanner.Scan() {
		if opts.IgnoreCase {
			duplicate = strings.EqualFold(last, scanner.Text())
		} else {
			duplicate = last == scanner.Text() // bool
		}

		printout(count, last, duplicate)
		count += 1
		last = scanner.Text()
	}

	if last != "" {
		printout(count, last, duplicate)
	}
	return nil
}

func Uniq(args []string) error {
	if len(args) == 0 {
		uniqFile(os.Stdin)
	}

	for _, file := range args {
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		uniqFile(f)
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if !opts.Repeats && !opts.Unique {
		opts.Unique = true
	}

	if err := Uniq(args); err != nil {
		log.Fatal(err)
	}
}
