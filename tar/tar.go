package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Create      bool   `short:"c" long:"create" description:"create a tar archive"`
	Extract     bool   `short:"x" long:"extract" description:"extract a tar archive"`
	File        string `short:"f" long:"file" description:"extract a tar archive"`
	List        bool   `short:"l" long:"list" description:"list contents of a tar archive"`
	NoRecursion bool   `long:"no-recursion" description:"do not recurse into directories"`
	Verbose     bool   `short:"v" long:"verbose" description:"print file names and operations"`
}

func checkOptions(args []string) error {
	if opts.Create && opts.Extract {
		return fmt.Errorf("Cannot create and extract an archive at the same time")
	}
	if opts.Create && opts.List {
		return fmt.Errorf("Cannot create and list an archive at the same time")
	}
	if opts.Extract && opts.List {
		return fmt.Errorf("Cannot extract and list an archive at the same time")
	}
	if opts.Extract && len(args) != 1 {
		return fmt.Errorf("extract needs an argument of a [file]")
	}
	if !opts.Extract && !opts.List && !opts.Create {
		return fmt.Errorf("Must include an operation [extract|create|list]")
	}
	if opts.File == "" {
		return fmt.Errorf("Must include a file")
	}
	return nil
}

func Tar(args []string) error {
	toptions := &TarOpts{
		NoRecurse: opts.NoRecursion,
	}

	if opts.Verbose {
		toptions.Filters = []Filter{VerboseFilter}
	}

	switch {
	case opts.Create:
		f, err := os.Create(opts.File)
		if err != nil {
			return err
		}
		if err := CreateTar(f, args, toptions); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}

	case opts.Extract:
		f, err := os.Open(opts.File)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := extractDir(f, args[0], toptions); err != nil {
			return err
		}

	case opts.List:
		f, err := os.Open(opts.File)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := listArchive(f); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		if err == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Fatal(err)
		}
	}

	if err := checkOptions(args); err != nil {
		log.Fatal(err)
	}

	if err := Tar(args); err != nil {
		log.Fatal(err)
	}
}
