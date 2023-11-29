// thanks to liz rice
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Print    bool     `short:"p" long:"print" description:"print syscalls as they are made VS waiting until the end"`
	Disallow []string `short:"d" long:"disallow" description:"syscalls to block"`
	Verbose  bool     `short:"v" long:"verbose" description:"describe what program is doing"`
}

func strace(args []string) error {
	var regs syscall.PtraceRegs
	var ss syscallCounter

	ss = ss.init()

	fmt.Printf("Run %v\n", args[0:])

	// Uncommenting this will cause the open syscall to return with Operation Not Permitted error
	if len(opts.Disallow) > 0 {
		for _, sc := range opts.Disallow {
			// ex: disallow("read")
			err := disallow(sc)
			if err != nil {
				return err
			}
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace: true,
	}

	cmd.Start()
	err := cmd.Wait()
	if err != nil {
		fmt.Printf("Wait returned: %v\n", err)
	}

	pid := cmd.Process.Pid
	exit := true

	for {
		if exit {
			err = syscall.PtraceGetRegs(pid, &regs)
			if err != nil {
				break
			}

			if opts.Print {
				// Uncomment to show each syscall as it's called
				name := ss.getName(regs.Orig_rax)
				fmt.Printf("%s\n", name)
			}
			ss.inc(regs.Orig_rax)
		}

		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			return err
		}

		_, err = syscall.Wait4(pid, nil, 0, nil)
		if err != nil {
			return err
		}

		exit = !exit
	}

	ss.print()
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}
	if err := strace(args); err != nil {
		log.Fatal(err)
	}
}
