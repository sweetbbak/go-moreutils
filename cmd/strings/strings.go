package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	MinLen  int    `short:"n" long:"min-len" default:"4" description:"print sequences that are [n] characters long at minimum"`
	Offset  string `short:"t" long:"radix" choice:"d" choice:"o" choice:"x" description:"Print string offset using decimal, octal, or hexadecimal (d/o/x)"`
	Verbose bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}
var errInvalidRange = errors.New("Min character range must be positive.")
var errInvalidArg = errors.New("Offset format must be one of: d, o, x.")

func offsetValue(offset, offsetOg int) string {
	off := offsetOg - offset

	switch opts.Offset {
	case "d":
		return fmt.Sprintf("%d ", off)
	case "o":
		return fmt.Sprintf("%o ", off)
	case "x":
		return fmt.Sprintf("%x ", off)
	default:
		panic("Offset parameter for flag [-t, --radix] invalid")
	}
}

func asciiIsPrint(char byte) bool {
	return char >= 32 && char <= 126
}

func stringsIO(br *bufio.Reader) error {
	var o []byte
	var offset int
	offset = 0
	for {
		b, err := br.ReadByte()
		if errors.Is(err, io.EOF) {
			if len(o) >= opts.MinLen {
				if opts.Offset != "" {
					os.Stdout.Write([]byte(offsetValue(len(o), offset)))
				}
				os.Stdout.Write(o)
				os.Stdout.Write([]byte{'\n'})
			}
			return nil
		}
		if err != nil {
			return err
		}

		if !asciiIsPrint(b) {
			if len(o) >= opts.MinLen {
				if opts.Offset != "" {
					os.Stdout.Write([]byte(offsetValue(len(o), offset)))
				}
				os.Stdout.Write(o)
				os.Stdout.Write([]byte{'\n'})
			}
			o = o[:0]
			offset++
			continue
		}
		if len(o) >= opts.MinLen+1024 {
			os.Stdout.Write(o[:1024])
			o = o[1024:]
		}
		o = append(o, b)
		offset++
	}
}

func stringsFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	br := bufio.NewReader(f)
	return stringsIO(br)
}

func Strings(args []string) error {
	if opts.MinLen < 1 {
		return fmt.Errorf("%w: %v", errInvalidRange, opts.MinLen)
	}

	if opts.Offset != "" {
		if opts.Offset != "d" && opts.Offset != "o" && opts.Offset != "x" {
			return fmt.Errorf("%w: %v", errInvalidArg, opts.Offset)
		}
	}

	if len(args) == 0 {
		br := bufio.NewReader(os.Stdin)
		if err := stringsIO(br); err != nil {
			return err
		}
	}

	for _, file := range args {
		if err := stringsFile(file); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Strings(args); err != nil {
		log.Fatal(err)
	}
}
