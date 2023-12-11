package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Timer   bool   `short:"t" description:"Dont print header"`
	Sec     string `short:"n" long:"num" default:"2s" description:"loop every N duration"`
	NoColor bool   `short:"c" long:"no-color" description:"Verbose output"`
	Verbose bool   `short:"v" long:"verbose" description:"Verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Watch(args []string) error {
	d, err := time.ParseDuration(opts.Sec)
	if err != nil {
		return err
	}
	Debug("Duration: %v\n", d)

	for {
		fmt.Print("\x1b[0;0H\x1b[J")
		if !opts.Timer {
			if opts.NoColor {
				fmt.Printf("Every [%v]: %v\n\n", d.String(), args)
			} else {
				fmt.Printf("Every \x1b[33m[%v]\x1b[0m: %v\n\n", d.String(), args)
			}
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		Debug("Running command: %v\n", cmd)
		if err := cmd.Run(); err != nil {
			if strings.Contains(err.Error(), "executable file not found") {
				fmt.Print(err)
			}
		}
		Debug("Sleeping: %v\n", d)
		time.Sleep(d)
	}
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		fmt.Println("Time duration format: [1h|1m|1s|1ms|1ns|1us]")
		fmt.Println("examples: 1.5s - 100ms - 99ns")
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Watch(args); err != nil {
		log.Fatal(err)
	}
}
