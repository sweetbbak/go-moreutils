package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Format     string `short:"f" long:"format" description:"format to use following printf style. (ex: %v)"`
	EqualWidth bool   `short:"w" long:"width" description:"equalize width by padding with leading zeros"`
	Separator  string `short:"s" long:"separator" description:"character to use to seperate sequenced numbers (default \n)"`
	Verbose    bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

// seq FIRST INC LAST || seq FIRST LAST
func Seq(args []string) error {
	var (
		width int
		end   float64
		step  = 1.0
		stt   = 1.0
	)

	argc := len(args)
	if argc < 1 || argc > 4 {
		return fmt.Errorf("incorrect number of arguments: args should be greater or equal to 1 and less than or equal to 3: got %v", argc)
	}

	if argc == 3 {
		_, err := fmt.Sscanf(args[1], "%v", &step)
		if step-float64(int(step)) > 0 && opts.Format == "%v" {
			d := len(fmt.Sprintf("%v", step-float64(int(step)))) - 2
			opts.Format = fmt.Sprintf("%%.%df", d)
			Debug("format: %v\n", opts.Format)
		}

		if step == 0.0 {
			return errors.New("step value should not be 0")
		}

		if err != nil {
			return err
		}
	}

	if argc >= 2 {
		if _, err := fmt.Sscanf(args[0]+" "+args[argc-1], "%v %v", &stt, &end); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Sscanf(args[0], "%v", &end); err != nil {
			return err
		}
	}

	opts.Format = strings.Replace(opts.Format, "%", "%0*", 1)
	if opts.EqualWidth {
		width = len(fmt.Sprintf(opts.Format, 0, end))
	}

	defer fmt.Printf("\n")
	for stt <= end {
		fmt.Printf(opts.Format, width, stt)
		stt += step
		if stt <= end {
			fmt.Print(opts.Separator)
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

	if opts.Separator == "" {
		opts.Separator = "\n"
	}

	if opts.Format == "" {
		opts.Format = "%v"
	}

	if err := Seq(args); err != nil {
		log.Fatal(err)
	}
}
