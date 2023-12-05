package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Append       bool `short:"a" long:"append" description:"append to file, do not overwrite"`
	IgnoreSignal bool `short:"i" long:"ignore" description:"ignore SIGINT and SIGKILL, ignore ctrl+c and kill commands"`
}

var (
	Overwrite = true
)

func handleSignals() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGINT)
	go func() {
		for {
			<-signalChan
			if !opts.IgnoreSignal {
				os.Exit(1)
			}
		}
	}()
}

func Tee(args []string) error {
	var files []*os.File

	if opts.Append {
		Overwrite = false
	}

	if len(args) >= 1 && args[0] == "-" || len(args) == 0 {
		files = append(files, os.Stdin)
	} else {
		for _, fi := range args {
			var f *os.File
			var err error
			if Overwrite {
				f, err = os.OpenFile(fi, os.O_WRONLY|os.O_CREATE, 0644)
			} else {
				f, err = os.OpenFile(fi, os.O_WRONLY|os.O_APPEND, 0644)
			}
			if err != nil {
				f.Close()
				continue
			}

			defer f.Close()
			files = append(files, f)
		}
	}
	handleSignals()

	buffer := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			en, ok := err.(syscall.Errno)
			// interrupted
			if ok && int(en) == int(syscall.EINTR) {
				continue
			} else {
				break
			}
		}
		os.Stdout.Write(buffer[0:n])
		for _, ff := range files {
			ff.Write(buffer[0:n])
		}
		if n <= 0 {
			break
		}
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if err := Tee(args); err != nil {
		log.Fatal(err)
	}
}
