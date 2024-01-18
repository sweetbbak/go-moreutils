package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Dir     bool   `short:"d" long:"directory" description:"make a temp directory"`
	DryRun  bool   `short:"u" long:"dry-run" description:"do not create anything, merely print a name"`
	Quiet   bool   `short:"q" long:"quiet" description:"Show no errors"`
	Prefix  string `short:"P" long:"prefix" description:""`
	Suffix  string `short:"s" long:"suffix" description:"Show no errors"`
	TempDir string `short:"p" long:"tmpdir" description:"directory prefix to use, example: /tmp"`
}

func Mktemp(args []string) error {
	if len(args) == 1 {
		opts.Prefix = opts.Prefix + strings.Split(args[0], "X")[0] + opts.Suffix
	}

	if opts.TempDir == "" {
		opts.TempDir = os.TempDir()
	}

	var printOut string
	if opts.Dir {
		d, err := os.MkdirTemp(opts.TempDir, opts.Prefix)
		if err != nil {
			return err
		}
		printOut = d
	} else {
		f, err := os.CreateTemp(opts.TempDir, opts.Prefix)
		if err != nil {
			return err
		}
		printOut = f.Name()
	}

	if !opts.Quiet {
		fmt.Println(printOut)
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if err := Mktemp(args); err != nil {
		log.Fatal(err)
	}
}
