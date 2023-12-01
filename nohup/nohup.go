package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
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
		sig := <-sigs
		log.Println("recieved: ", sig)
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

	// shup()
	cmdstr := strings.Join(cmd, " ")
	Start(output, cmdstr)
	// exitCode := System(cmd, output, true)
	// return exitCode
	return 0
}

// func System(cmd string, out *os.File, ignoreStdin bool) int {
func System(cmd []string, out *os.File, ignoreStdin bool) int {
	go shup()
	c := exec.Command(cmd[0], cmd[1:]...)
	// c.Stdin
	c.Stdout = out
	c.Stderr = out
	err := c.Run()
	// err := c.Start()
	logOut := fmt.Sprintf("startd process of PID [%v]", c.Process.Pid)
	out.WriteString(logOut)
	// c.Wait()

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
		// procAttr.Files = []*os.File{
		// 	os.Stdin,
		// 	os.Stdout,
		// 	os.Stderr,
		// }

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

		// procAttr.Sys.Ptrace = true

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
		signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	} else {
		fmt.Println("signal is ignored")
	}
	// catch the SIGHUP signal (Hangup - often from terminal close)
	go CatchHup(sigs)
}

func main() {
	// channel
	sigs := make(chan os.Signal)
	// notify channel
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	// catch the SIGHUP signal (Hangup - often from terminal close)
	go CatchHup(sigs)

	cmd := os.Args[1:]
	// Start(cmd...)

	ex := nohup(cmd)
	os.Exit(ex)
}
