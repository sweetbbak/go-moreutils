package main

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Logname(args []string) error {
	username, err := user.Current()
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, username)

	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}

	if err != nil {
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Logname(args); err != nil {
		log.Fatal(err)
	}
}
