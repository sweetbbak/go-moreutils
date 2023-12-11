package main

// This package is a substitute for TPUT but a little more sane

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	output = flag.String("o", "stdout", "options: [stdout/stderr] - which stream to print to")
	List   = flag.Bool("l", false, "list available commands")
)

var commands = map[string]string{
	"clear":       "\x1b[2H\x1b[H",
	"hidecursor":  "\x1b[?25l",
	"showcursor":  "\x1b[?25h",
	"restore":     "\x1b[?47l",
	"save":        "\x1b[?47h",
	"rmaltscreen": "\x1b[?1049l",
	"alt":         "\x1b[?1049h",
}

func Ansi(w io.Writer, args []string) error {
	for _, arg := range args {
		_, exists := commands[arg]
		if exists {
			fmt.Fprintf(w, commands[arg])
		} else {
			return fmt.Errorf("ANSI command '%v' doesnt exist", arg)
		}
	}
	return nil
}

func main() {
	flag.Parse()

	if *List {
		for k, _ := range commands {
			fmt.Println(k)
		}
		os.Exit(0)
	}

	var out *os.File
	switch strings.ToLower(*output) {
	case "stdout":
		out = os.Stdout
	case "stderr":
		out = os.Stderr
	}

	if err := Ansi(out, flag.Args()); err != nil {
		log.Fatal(err)
	}
}
