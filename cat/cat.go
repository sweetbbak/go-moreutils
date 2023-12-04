package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	CountNonBlank bool `short:"b" long:"blank" description:"Number non-blank lines starting at 1"`
	NumberOut     bool `short:"n" long:"number" description:"Number all lines starting at 1"`
	Squeeze       bool `short:"s" long:"squeeze" description:"Squeeze multiple empty lines into a single empty line"`
	Verbose       bool `short:"v" long:"verbose" description:"Print exactly what the program is doing"`
}

// wrap open file with the ability to open Unix sockets
func openfile(s string) (io.ReadWriteCloser, error) {
	file, err := os.Stat(s)
	if err != nil {
		return nil, err
	}

	// bitwise operation
	if file.Mode()&os.ModeSocket != 0 {
		return net.Dial("unix", s)
	}
	return os.Open(s)
}

func concatenate(w io.Writer, r io.Reader) (n int64, err error) {
	var lastline, line string
	br := bufio.NewReader(r)
	nr := 0
	for {
		line, err = br.ReadString('\n')
		if err != nil {
			return
		}

		// if last line and this line are empty, pass
		if opts.Squeeze && lastline == "\n" && line == "\n" {
			continue
		}
		// count non-blank dont print a number, just print an empty line
		if opts.CountNonBlank && line == "\n" || line == "\n" {
			fmt.Fprint(w, line)
		} else if opts.CountNonBlank || opts.NumberOut {
			// else print line + number + increment counter
			nr++
			fmt.Fprintf(w, "%6d\t%s", nr, line)
		} else {
			fmt.Fprint(w, line)
		}
		// assign lastline to line after we are done operating on it
		lastline = line
	}
}

func Cat(args []string) error {
	var byte_count int64
	mycopy := io.Copy
	if opts.CountNonBlank || opts.NumberOut || opts.Squeeze {
		mycopy = concatenate
	}

	if len(args) < 1 {
		mycopy(os.Stdout, os.Stdin)
	}

	for _, filename := range args {
		// typical cat behavior in a pipe or when called alone
		if filename == "-" {
			mycopy(os.Stdout, os.Stdin)
		} else {
			f, err := openfile(filename)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			by, err := mycopy(os.Stdout, f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			if opts.Verbose {
				byte_count += by
			}
			f.Close()
		}
	}

	if opts.Verbose {
		fmt.Fprintln(os.Stderr, byte_count)
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		if flags.WroteHelp(err) {
			fmt.Println("")
		}
		os.Exit(0)
	}

	if err := Cat(args); err != nil {
		log.Fatal(err)
	}
}
