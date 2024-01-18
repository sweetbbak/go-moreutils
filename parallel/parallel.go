package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Null    bool `short:"0" long:"zero" description:"print a NULL instead of a newline"`
	Jobs    int  `short:"j" long:"jobs" description:"number of jobs to run"`
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

const (
	newLine  = byte('\n')
	nullLine = byte('\x00')
)

func Parallel(args []string) error {
	var cmd string
	if len(args) < 1 {
		return fmt.Errorf("example usage: parallel [CMD] < find")
	} else {
		cmd = args[0]
	}

	delim := newLine
	if opts.Null {
		delim = nullLine
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	w, err := newWorkerPool(
		ctx,
		os.Stdout,
		os.Stderr,
		os.Stdin,
		delim,
		cmd,
		args[1:],
		opts.Jobs,
	)

	if err != nil {
		log.Println(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go handleSignals(c, cancel)

	w.run()

	return nil
}

type writer struct {
	writer *bufio.Writer
	mu     sync.Mutex
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	n, err = w.writer.Write(p)
	w.mu.Unlock()
	return n, err
}

func (w *writer) Flush() error {
	return w.writer.Flush()
}

func newWriter(w io.Writer) *writer {
	return &writer{writer: bufio.NewWriter(w)}
}

type queue struct {
	ch <-chan string
}

func newQueue(reader io.Reader, splitChar byte, queueBuffer int) queue {
	ch := make(chan string, queueBuffer*2) // Buffer the channel to a reasonable value

	// Build the scanner and start scanning lines into the job queue in the
	// background while we return our new queue.
	go func() {
		scanner := newScanner(reader, splitChar)

		for scanner.Scan() {
			ch <- scanner.Text()
		}
		close(ch)
	}()
	return queue{ch: ch}
}

func newScanner(reader io.Reader, splitChar byte) *bufio.Scanner {
	scanner := bufio.NewScanner(reader)
	scanner.Split(newSplitFunc(splitChar))
	return scanner
}

// This function is used to return a new `bufio.SplitFunc` splitting on
// whichever character the user specifies. The code for this is mostly just
// lifted out of `bufio.ScanLines`, replacing the newline character with a
// paramter.
func newSplitFunc(char byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, char); i >= 0 {
			// We have a full null-terminated line.
			return i + 1, data[0:i], nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
}

type WorkerPool struct {
	args        []string
	cmd         string
	concurrency int
	ctx         context.Context
	err         *writer
	out         *writer
	queue       queue
	runner      func(line string)
	start       time.Time
}

func (w *WorkerPool) startWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		case line, open := <-w.queue.ch:
			if !open {
				return
			}
			w.runner(line)
		}
	}
}

func (w *WorkerPool) runCmd(input string) {
	args := append(w.args, input)
	cmd := newCmd(w.out, w.err, w.cmd, args...)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(w.err, "failed to run command: '%v %v' %v\n", w.cmd, strings.Join(args, " "), err)
	}
}

func (w *WorkerPool) run() {
	wg := sync.WaitGroup{}

	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go w.startWorker(&wg)
	}
	wg.Wait()
	w.out.Flush()
	w.err.Flush()
}

func newWorkerPool(ctx context.Context, stdout, stderr io.Writer, reader io.Reader, splitChar byte, cmd string, args []string, concurrency int) (*WorkerPool, error) {
	var path string
	var err error
	if cmd != "" {
		path, err = exec.LookPath(cmd)
		if err != nil {
			return nil, err
		}
	}

	w := &WorkerPool{
		args:        args,
		cmd:         path,
		concurrency: concurrency,
		ctx:         ctx,
		err:         newWriter(stderr),
		out:         newWriter(stdout),
		queue:       newQueue(reader, splitChar, concurrency),
		start:       time.Now(),
	}
	w.runner = w.runCmd
	return w, nil
}

func handleSignals(c chan os.Signal, cancel context.CancelFunc) {
	<-c
	fmt.Println("Got SIGINT, exiting gracefully... [Ctrl+C] to end immediately")
	go func() {
		cancel()
		os.Exit(0)
	}()
	<-c
	os.Exit(33)
}

func init() {
	opts.Jobs = runtime.NumCPU()
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Jobs > runtime.NumCPU() {
		opts.Jobs = runtime.NumCPU()
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Parallel(args); err != nil {
		log.Fatal(err)
	}
}
