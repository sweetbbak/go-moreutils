package main

import (
	"fmt"
	"os"
	"syscall"
)

func Sethostname(name string) error {
	if err := syscall.Sethostname([]byte(name)); err != nil {
		return err
	}
	return nil
}

func userSetHostname(hostname string) error {
	if err := os.WriteFile("/etc/hostname", []byte(hostname), 0o644); err != nil {
		return fmt.Errorf("error writing hostname file: %w", err)
	}

	if err := Sethostname(hostname); err != nil {
		return fmt.Errorf("error setting hostname: %w", err)
	}
	return nil
}
