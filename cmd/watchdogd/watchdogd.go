package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/u-root/u-root/pkg/watchdogd"
)

var opts struct {
	Dev        string        `short:"d" long:"dev" default:"/dev/watchdog" description:"specify watchdog device"`
	Timeout    time.Duration `short:"t" long:"timeout" description:"duration before timeout"`
	preTimeout time.Duration `short:"p" long:"pre-timeout" description:"duration before pre-timeout"`
	KeepAlive  time.Duration `short:"k" long:"keep-alive" description:"duration between issuing keepalive signal"`
	Monitors   string        `short:"m" long:"monitors" description:"comma separated list of monitors"`
	UDS        string        `short:"u" long:"uds" description:"unix domain socket path for the daemon"`
	Verbose    bool          `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func WatchDog(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No input")
	}

	cmd, args := args[0], args[1:]

	var timeout *time.Duration
	timeout = &opts.Timeout
	var preTimeout *time.Duration
	preTimeout = &opts.preTimeout

	switch cmd {
	case "run":
		if opts.Timeout == -1 {
			timeout = nil
		}
		if opts.preTimeout == -1 {
			preTimeout = nil
		}
		monitorFuncs := []func() error{}
		for _, m := range strings.Split(opts.Monitors, ",") {
			if m == "oops" {
				monitorFuncs = append(monitorFuncs, watchdogd.MonitorOops)
			} else {
				return fmt.Errorf("unrecognized monitors: %v", m)
			}
		}
		return watchdogd.Run(context.Background(), &watchdogd.DaemonOpts{
			Dev:        opts.Dev,
			Timeout:    timeout,
			PreTimeout: preTimeout,
			KeepAlive:  opts.KeepAlive,
			Monitors:   monitorFuncs,
			UDS:        opts.UDS,
		})
	default:
		if len(args) != 0 {
			return fmt.Errorf("Unknown extra arguments after: %v unknown: [%v]", cmd, args[0:])
		}

		d, err := watchdogd.NewClient()
		if err != nil {
			return fmt.Errorf("could not dial watchdog daemon: %w", err)
		}
		f, ok := map[string]func() error{
			"stop":     d.Stop,
			"continue": d.Continue,
			"arm":      d.Arm,
			"disarm":   d.Disarm,
		}[cmd]
		if !ok {
			return fmt.Errorf("unrecognized command %q", cmd)
		}
		return f()
	}
}

const usage = `SUBCOMMAND:
watchdogd run
	Run the watchdogd daemon as a child process (does not daemonize)
watchdogd stop
	Send a signal to arm the running watchdogd.
watchdogd continue
	Send a signal to disarm the running watchdogd.
watchdogd arm
	Send a signal to arm the running watchdogd.
watchdogd disarm
	Send a signal to disarm the running watchdogd.
`

func init() {
	opts.KeepAlive = time.Second * 5
}
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

	if err := WatchDog(args); err != nil {
		log.Fatal(err)
	}
}
