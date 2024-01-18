package main

import (
	"bufio"
	"bytes"
	crand "crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Unlink   bool   `short:"u" long:"remove" description:"unlink/remove a file after renaming it"`
	Passes   int64  `short:"n" long:"iterations" default:"3" description:"number of passes"`
	Force    bool   `short:"f" long:"force" description:"change file permissions to allow writing"`
	Quiet    bool   `short:"q" long:"quiet" description:"dont print any output"`
	Secure   bool   `short:"s" long:"secure" description:"use random bytes that are genuinely more random"`
	Zero     bool   `short:"z" long:"zero" description:"add a final overwrite with zeros to hide shredding"`
	RandFile string `long:"random-source" description:"get random bytes from FILE (no-op for now)"`
	Verbose  bool   `short:"v" long:"verbose" description:"verbose mode"`
}

// default "chunk" size we will use to overwrite a file
const defualtBufSize = 4096

const (
	RED     = "\x1b[31m"
	GREEN   = "\x1b[32m"
	BLUE    = "\x1b[33m"
	CLR     = "\x1b[0m"
	BOLD    = "\x1b[1m"
	ITALIC  = "\x1b[3m"
	UNDRLNE = "\x1b[4m"
)

var Debug = func(string, ...interface{}) {}

func shredFile(f *os.File, fi fs.FileInfo, bytesWritten *int64) error {
	var r io.Reader

	if opts.Secure {
		r = crand.Reader
	} else {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	wr := bufio.NewWriter(f)
	buf := make([]byte, defualtBufSize)

	// var bytesWritten int64
	// progress := New(os.Stderr, "xfer", &bytesWritten)
	// progress.Begin()

	// write data in blocks of 4096 until we reach the end and then we seek to the end of the file.
	for sizeLeft := fi.Size(); sizeLeft > 0; sizeLeft -= defualtBufSize {
		if sizeLeft < defualtBufSize {
			buf = make([]byte, sizeLeft)
		}
		r.Read(buf)
		nn, err := wr.Write(buf)
		if err != nil {
			if err == io.EOF {
				continue
			}
			return err
		}

		if !opts.Quiet {
			atomic.AddInt64(bytesWritten, int64(nn))
		}
	}

	// progress.End()

	wr.Flush()
	f.Sync()
	// f.Close()
	return nil
}

func zeroFile(f *os.File, fi fs.FileInfo) error {
	f.Seek(0, 0)
	buf := bytes.Repeat([]byte{0x00}, defualtBufSize)
	wr := bufio.NewWriter(f)

	// write data in blocks of 4096 until we reach the end and then we seek to the end of the file.
	for sizeLeft := fi.Size(); sizeLeft > 0; sizeLeft -= defualtBufSize {
		if sizeLeft < defualtBufSize {
			buf = bytes.Repeat([]byte{0x00}, int(sizeLeft))
		}
		wr.Write(buf)
	}

	wr.Flush()
	f.Sync()
	f.Close()
	return nil
}

const LettersAscii = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = LettersAscii[rand.Intn(len(LettersAscii))]
	}
	return string(b)
}

func renameFile(filename string) error {
	for i := len(filename); i > 1; i-- {
		newname := randomString(i)
		os.Rename(filename, newname)
		filename = newname
	}

	return syscall.Unlink(filename)
}

func Shred(args []string) error {
	for _, file := range args {
		Debug("%s[INFO]%s Begin shredding file: %s%v%s\n", GREEN, CLR, ITALIC, file, CLR)
		fi, err := os.Stat(file)
		if err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return fmt.Errorf("Cannot shred special files or directories.")
		}

		f, err := os.OpenFile(file, os.O_WRONLY, 0o660)
		if err != nil {
			if os.IsPermission(err) {
				Debug("%s[ERR]%s shred: %v\n", RED, CLR, err)
				if opts.Force {
					if err := os.Chmod(file, fi.Mode()|syscall.S_IWUSR); err != nil {
						Debug("%s[ERR]%s shred: %v\n", RED, CLR, err)
						return err
					}
				}
			} else {
				return err
			}
		}

		if opts.Passes < 0 {
			return fmt.Errorf("Passes cannot be negative.")
		}

		var bytesWritten int64 = 0
		for i := int64(0); i < int64(opts.Passes); i++ {
			Debug("%s[INFO]%s shred: shredding %v: pass %s%v%s\n", GREEN, CLR, file, BOLD, i+1, CLR)

			progress := New(os.Stderr, "progress", &bytesWritten)
			progress.Begin()

			f, err := os.OpenFile(file, os.O_WRONLY, 0o660)
			if err != nil {
				return err
			}
			shredFile(f, fi, &bytesWritten)
			f.Close()

			progress.End()
		}

		if opts.Zero {
			f, err := os.OpenFile(file, os.O_WRONLY, 0o660)
			if err != nil {
				return err
			}
			Debug("%s[INFO]%s shred: %szeroing%s %v\n", GREEN, CLR, ITALIC, CLR, file)
			zeroFile(f, fi)
			f.Close()
		}

		if opts.Unlink {
			Debug("%s[INFO]%s shred: %sremoving%s %v\n", GREEN, CLR, ITALIC, CLR, file)
			renameFile(f.Name())
		}
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Shred(args); err != nil {
		log.Fatal(err)
	}
}
