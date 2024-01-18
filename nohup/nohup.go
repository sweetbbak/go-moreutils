package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	// "strings"
	"syscall"
	"time"
)

// Exit Status
// 125 - command failed
// 127 command not found
// - the command exit status

var (
	STDOUT_FILENO      = 1
	STDERR_FILENO      = 2
	ignore_input       bool
	redirecting_stdout bool
	redirecting_stderr bool
	stdin_closed       bool
)

func exit_internal_failure() int {
	// posix internal failure status requires 127 exit code VS 1-125
	posix := os.Getenv("POSIXLY_CORRECT")
	if posix == "" {
		return 127
	} else {
		return 127
	}
}

func isatty(file *os.File) bool {
	// pass os.Stdout or os.Stdin
	o, _ := file.Stat()
	if (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice { //Terminal
		return true
	} else { //It is not the terminal
		return false
	}
}

func Dup2(oldfd int, newfd int) error {
	return syscall.Dup2(oldfd, newfd)
}

func fd_reopen(fd2 int, file string, flags int, mode fs.FileMode) int {
	fd, err := os.OpenFile(file, flags, mode)
	if err != nil {
	}
	if int(fd.Fd()) == fd2 || int(fd.Fd()) < 0 {
		return int(fd.Fd())
	} else {
		err := Dup2(int(fd.Fd()), fd2)
		if err != nil {
			log.Println(err)
		}
		return fd2
	}
}

func CatchHup(sigs chan os.Signal) {
	for {
		signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		sig := <-sigs
		switch sig {
		case syscall.SIGINT:
			log.Println("recieved int: ", sig)
			signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
			// continue
		}
		<-sigs
	}
}

func nohup(cmd []string) int {
	ignore_input = isatty(os.Stdin)
	redirecting_stdout = isatty(os.Stdout)
	redirecting_stderr = isatty(os.Stderr)

	output, err := os.OpenFile("nohup.out", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}

	// cmdstr := strings.Join(cmd, " ")
	// p, err := Start(output, cmdstr)
	// if err != nil {
	// 	log.Println(err)
	// }
	// time.Sleep(2 * time.Second)
	// p.Signal(syscall.SIGHUP)

	exitCode := System(cmd, output, true)
	return exitCode
}

// func System(cmd string, out *os.File, ignoreStdin bool) int {
func System(cmd []string, out *os.File, ignoreStdin bool) int {
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdout = out
	c.Stderr = out
	err := c.Run()
	logOut := fmt.Sprintf("startd process of PID [%v]", c.Process.Pid)
	out.WriteString(logOut)

	if err == nil {
		return 0
	}

	// Figure out the exit code
	if ws, ok := c.ProcessState.Sys().(syscall.WaitStatus); ok {
		if ws.Exited() {
			return ws.ExitStatus()
		}

		if ws.Signaled() {
			return -int(ws.Signal())
		}
	}
	return -1
}

// Janitor will watch a file and make sure it doesn't go over the specified size
func Janitor(file string, size int) {
	for {
		info, err := os.Stat(file)
		if err == nil {
			if int64(size) <= int64(float64(0.000001)*float64(info.Size())) {
				// we should do something about the file size!
				os.Truncate(file, 0)
			}
		}
		<-time.After(1 * time.Second)
	}
}

func Start(out *os.File, args ...string) (p *os.Process, err error) {
	if args[0], err = exec.LookPath(args[0]); err == nil {
		var procAttr os.ProcAttr
		sys := syscall.SysProcAttr{
			Setsid: true,
		}

		cwd, _ := os.Getwd()

		procAttr = os.ProcAttr{
			Dir:   cwd,
			Env:   os.Environ(),
			Sys:   &sys,
			Files: []*os.File{os.Stdin, out, out},
		}

		p, err := os.StartProcess(args[0], args, &procAttr)
		if err == nil {
			return p, nil
		}
	}
	return nil, err
}

func shup() {
	// channel
	sigs := make(chan os.Signal)
	// notify channel
	if !signal.Ignored(syscall.SIGHUP) {
		signal.Ignore(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	} else {
		fmt.Println("signal is ignored")
	}
	// catch the SIGHUP signal (Hangup - often from terminal close)
	CatchHup(sigs)
}

func main() {
	// channel
	// sigs := make(chan os.Signal, 5)
	// notify channel
	// signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	// catch the SIGHUP signal (Hangup - often from terminal close)
	// go CatchHup(sigs)
	// go shup()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			switch sig {
			case syscall.SIGINT:
				fmt.Println(sig)
			}
		}
	}()

	cmd := os.Args[1:]
	// Start(cmd...)

	ex := nohup(cmd)
	os.Exit(ex)
}
