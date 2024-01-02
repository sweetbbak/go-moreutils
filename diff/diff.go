package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var opts struct {
	Unified bool `short:"u" long:"unified" description:"unified output as used by git diff"`
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Diff(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("must ")
	}

	b1, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	b2, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}

	t1 := string(b1)
	t2 := string(b2)

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(t1, t2, false)
	fmt.Println(dmp.DiffPrettyText(diffs))

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

	if err := Diff(args); err != nil {
		log.Fatal(err)
	}
}
