package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/jessevdk/go-flags"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

var (
	command string
)

var opts struct {
	Command string `short:"c" long:"command" description:"command to run"`
}

func sh(w io.Writer, cmd string, args []string) error {
	err := runAll(cmd, args)
	if e, ok := interp.IsExitStatus(err); ok {
		os.Exit(int(e))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func runAll(cmd string, args []string) error {
	r, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		return err
	}
	if command != "" {
		return run(r, strings.NewReader(command), "")
	}
	if len(args) == 0 {
		if term.IsTerminal(int(os.Stdin.Fd())) {
			return runInteractive(r, os.Stdin, os.Stdout, os.Stderr)
		}
		return run(r, os.Stdin, "")
	}
	for _, path := range args {
		if err := runPath(r, path); err != nil {
			return err
		}
	}
	return nil
}

func run(r *interp.Runner, reader io.Reader, name string) error {
	prog, err := syntax.NewParser().Parse(reader, name)
	if err != nil {
		return err
	}
	r.Reset()
	ctx := context.Background()
	return r.Run(ctx, prog)
}

func runPath(r *interp.Runner, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return run(r, f, path)
}

func runInteractive(r *interp.Runner, stdin io.Reader, stdout, stderr io.Writer) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	parser := syntax.NewParser()
	fmt.Fprintf(stdout, "$ ")
	var runErr error
	fn := func(stmts []*syntax.Stmt) bool {
		if parser.Incomplete() {
			fmt.Fprintf(stdout, "> ")
			return true
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			select {
			case <-signals:
				cancel()
				return
			case <-ctx.Done():
				return
			}
		}()
		for _, stmt := range stmts {
			runErr = r.Run(ctx, stmt)
			if r.Exited() {
				return false
			}
		}
		fmt.Fprintf(stdout, "$ ")
		return true
	}
	if err := parser.Interactive(stdin, fn); err != nil {
		return err
	}
	return runErr
}

func Shell(stdout io.Writer, args []string) error {
	return sh(stdout, command, args)
}

func main() {
	args, err := flags.Parse(&opts)
	if err == flags.ErrHelp {
		os.Exit(0)
	}

	if err != nil {
		log.Fatal(err)
	}

	command = opts.Command

	if err := Shell(os.Stdout, args); err != nil {
		log.Fatal(err)
	}
}
