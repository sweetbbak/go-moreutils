package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/u-root/u-root/pkg/watchdog"
)

var opts struct {
	Dev     string `short:"d" long:"dev" default:"/dev/watchdog" description:"specify watchdog device"`
	Verbose bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Dog(args []string) error {
	wd, err := watchdog.Open(opts.Dev)
	if err != nil {
		return err
	}
	defer func() {
		if err := wd.Close(); err != nil {
			log.Printf("Failed to close watchdog: %v\n", err)
		}
	}()

	switch args[0] {
	case "keepalive":
		if err := wd.KeepAlive(); err != nil {
			return err
		}
	case "settimeout":
		if len(args) < 2 {
			return fmt.Errorf("Must include DURATION in format (1h2m33s)")
		}
		d, err := time.ParseDuration(args[1])
		if err != nil {
			return err
		}
		if err := wd.SetTimeout(d); err != nil {
			return err
		}
	case "setpretimeout":
		if len(args) < 2 {
			return fmt.Errorf("Must include DURATION in format (1h2m33s)")
		}
		d, err := time.ParseDuration(args[1])
		if err != nil {
			return err
		}
		if err := wd.SetPreTimeout(d); err != nil {
			return err
		}
	case "gettimeout":
		i, err := wd.PreTimeout()
		if err != nil {
			return err
		}
		fmt.Println(i)
	case "gettimeleft":
		i, err := wd.TimeLeft()
		if err != nil {
			return err
		}
		fmt.Println(i)
	default:
		return fmt.Errorf("unrecognized command: %q", args[0])
	}
	return nil
}

const usage = `SUBCOMMAND:
watchdog keepalive
		Pet the watchdog. This resets the time left back to the timeout.
watchdog set[pre]timeout DURATION
		Set the watchdog timeout or pretimeout.
watchdog gettimeleft
		Print the amount of time left.
	`

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		fmt.Print(usage)
		os.Exit(0)
	}

	if err != nil {
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Dog(args); err != nil {
		log.Fatal(err)
	}
}
