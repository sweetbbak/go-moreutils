package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
)

func Sethostname(name string) error {
	if err := syscall.Sethostname([]byte(name)); err != nil {
		return err
	}
	return nil
}

func Hostname(stdout io.Writer, args []string) error {
	switch len(args) {
	case 2:
		if err := Sethostname(args[1]); err != nil {
			return fmt.Errorf("could not set hostname: %v", err)
		}
	case 1:
		host, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("could not get hostname: %v", err)
		}
		_, err = fmt.Fprintln(stdout, host)
		return err
	default:
		return fmt.Errorf("usage: hostname [HOSTNAME]")
	}
	return nil
}

func main() {
	if err := Hostname(os.Stdout, os.Args); err != nil {
		log.Fatal(err)
	}
}
