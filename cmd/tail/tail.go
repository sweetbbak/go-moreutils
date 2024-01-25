package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Follow   bool `short:"f" long:"follow" description:"follow the end of a file"`
	Numlines uint `short:"n" long:"lines" description:"print N number of lines"`
	Verbose  bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

type readAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

func getBlocksize(numlines uint) int64 {
	// 81 is an estimation of the average line length
	return 81 * int64(numlines)
}

func lastNLines(buf []byte) []byte {
	slice := buf
	var data []byte
	if len(slice) != 0 {
		if slice[len(slice)-1] == '\n' {
			slice = slice[:len(slice)-1]
		}
		var foundLines uint
		var idx int

		for foundLines < opts.Numlines {
			idx = bytes.LastIndexByte(slice, '\n')
			if idx == -1 {
				break
			}
			foundLines++
			if len(slice) > 1 && slice[idx-1] == '\n' {
				slice = slice[:idx]
			} else {
				slice = slice[:idx-1]
			}
		}
		if idx == -1 {
			data = buf
		} else {
			data = buf[idx+1:] // +1 to skip the newline to beginning of previous line
		}
	}
	return data
}

func readLastLines(file readAtSeeker, writer io.Writer) error {
	blksize := getBlocksize(opts.Numlines)
	lastPos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	// read block by block in reverse until we get numLines
	readData := make([]byte, 0)
	buf := make([]byte, blksize)
	pos := lastPos
	var foundLines uint

	// for each block, count lines
	for pos != 0 {
		var thisChunkSize int64
		if pos < blksize {
			thisChunkSize = pos
		} else {
			thisChunkSize = blksize
		}
		pos -= thisChunkSize
		n, err := file.ReadAt(buf, pos)
		if err != nil && err != io.EOF {
			return err
		}

		// merge this block to what was read so far
		readData = append(buf[:n], readData...)
		// count lines and stop if done
		foundLines += uint(bytes.Count(buf[:n], []byte{'\n'}))
		if foundLines >= opts.Numlines {
			break
		}
	}

	data := lastNLines(readData)
	if _, err := writer.Write(data); err != nil {
		return err
	}
	_, err = file.Seek(lastPos, io.SeekStart)
	return err
}

func readLinesBeginning(input io.ReadSeeker, writer io.Writer) error {
	blksize := getBlocksize(opts.Numlines)
	buf := make([]byte, blksize)
	var slice []byte
	var foundLines uint

	for {
		n, err := io.ReadFull(input, buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if err != io.ErrUnexpectedEOF {
				return err
			}
		}

		foundLines += uint(bytes.Count(buf[:n], []byte{'\n'}))
		slice = append(slice, buf[:n]...)
		slice = lastNLines(slice)
	}
	if _, err := writer.Write(slice); err != nil {
		return err
	}
	return nil
}

func isTruncated(file *os.File) (bool, error) {
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}

	fi, err := file.Stat()
	if err != nil {
		return false, err
	}
	return currentPos > fi.Size(), nil
}

func tailFile(file *os.File, writer io.Writer) error {
	retryFromBeginning := false
	err := readLastLines(file, os.Stdout)
	if err != nil {
		// if it fails from being unable to seek, retry from beginning
		if patherr, ok := err.(*os.PathError); ok && patherr.Err == syscall.ESPIPE {
			retryFromBeginning = true
		} else {
			return err
		}
	}

	if retryFromBeginning {
		err = readLinesBeginning(file, os.Stdout)
		if err != nil {
			return err
		}
	}

	if opts.Follow {
		fi, err := file.Stat()
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeNamedPipe != 0 {
			return fifoReader(file, os.Stdout)
		} else {
			return followFile(file, os.Stdout)
		}
	}
	return nil
}

func followFile(file *os.File, writer io.Writer) error {
	blksize := getBlocksize(1)
	buf := make([]byte, blksize)
	for {
		n, err := file.Read(buf)
		if errors.Is(err, io.EOF) {
			// without this sleep you would hogg the CPU
			time.Sleep(50 * time.Millisecond)
			// truncated ?
			truncated, errTruncated := isTruncated(file)
			if errTruncated != nil {
				break
			}
			if truncated {
				// seek from start
				_, errSeekStart := file.Seek(0, io.SeekStart)
				if errSeekStart != nil {
					break
				}
			}
			continue
		}
		if err == nil {
			_, err := writer.Write(buf[:n])
			if err != nil {
				return err
			}
			continue
		}
		break
	}
	return nil
}

func handleSignals(stop chan os.Signal) {
	for {
		<-stop
		fmt.Fprintln(os.Stderr, "Exiting...")
		os.Exit(0)
	}
}

func fifoReader(pipe *os.File, writer io.Writer) error {
	reader := bufio.NewReader(pipe)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, os.Kill)

	go handleSignals(stop)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			} else {
				return err
			}
		}
		_, err = writer.Write(line)
		if err != nil {
			return err
		}
	}
}

func Tail(args []string) error {
	if opts.Numlines < 0 {
		return fmt.Errorf("number of lines cannot be negative")
	}

	for _, file := range args {
		infile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer infile.Close()

		if err := tailFile(infile, os.Stdout); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	opts.Numlines = 10
}

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Tail(args); err != nil {
		log.Fatal(err)
	}
}
