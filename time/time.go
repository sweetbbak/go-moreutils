package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func printProcessState(c *exec.Cmd, rt time.Duration) {
	if c.ProcessState == nil {
		return
	}
	rts := fmt.Sprintf("%s %.03f", "real", rt.Seconds())
	us := fmt.Sprintf("%s %s", "user", c.ProcessState.UserTime())
	sys := fmt.Sprintf("%s %s", "sys", c.ProcessState.SystemTime())
	fmt.Printf("%v\n%v\n%v\n", rts, us, sys)
}

func run(args []string) error {
	start := time.Now()
	if len(args) == 0 {
		return errors.New("no command")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%q:%w", args, err)
	}

	rt := time.Since(start)
	printProcessState(cmd, rt)

	return nil
}

func main() {
	flag.Parse()
	if err := run(flag.Args()); err != nil {
		log.Fatalf("time: %v", err)
	}
}
