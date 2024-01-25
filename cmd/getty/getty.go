package main

import (
	"fmt"
	"log"
	"mybox/pkg/termios"
	"os"
	"os/exec"
	// "strconv"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Baud     int    `short:"b" long:"baud" description:"set baud rate of terminal"`
	Port     string `short:"p" long:"port" description:"set port of terminal"`
	Terminal string `short:"t" long:"terminal" description:"set terminal environment variable"`
	Shell    string `short:"s" long:"shell" default:"/bin/sh" description:"set shell for new tty. defaults to /bin/sh"`
}

func Getty(args []string) error {
	port := opts.Port
	baud := opts.Baud
	// port := args[0]
	// baud, err := strconv.Atoi(args[1])
	// if err != nil {
	// 	baud = 0
	// }

	var term string
	if opts.Terminal != "" {
		term = opts.Terminal
	}

	// if len(args) > 2 {
	// 	term = args[2]
	// }

	ttys, err := termios.NewTTYS(port)
	if err != nil {
		return fmt.Errorf("Error opening port %s: %w", port, err)
	}

	if _, err := ttys.Serial(baud); err != nil {
		return fmt.Errorf("error configuring baud rate for %s to %d: %w", port, baud, err)
	}

	if term != "" {
		err = os.Setenv("TERM", term)
		if err != nil {
			log.Printf("error setting TERM environment variable: TERM=%s: %v", term, err)
		}
	}

	shell := opts.Shell
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell)
	cmd.Env = os.Environ()
	ttys.Ctty(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting shell: %w", err)
	}

	if err := cmd.Process.Release(); err != nil {
		log.Printf("Error releasing process: %v", err)
	}

	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		fmt.Println("Example: getty --terminal=linux --port=ttys0 --shell=/bin/sh")
		os.Exit(0)
	}

	if err := Getty(args); err != nil {
		log.Fatal(err)
	}
}
