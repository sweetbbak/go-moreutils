package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/icza/backscanner"
	"github.com/jessevdk/go-flags"
)

var opts struct {
}

func reverseStdin() error {
	scanner := bufio.NewScanner(os.Stdin)
	var line []string
	for scanner.Scan() {
		line = append(line, scanner.Text())
	}
	for i := len(line) - 1; i >= 0; i-- {
		fmt.Println(line[i])
	}
	return nil
}

func reverse(f *os.File) error {
	ret, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		// return err
		ret = 2000
	}

	scanner := backscanner.New(f, int(ret))
	for {
		line, pos, err := scanner.Line()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading line at position %v: %v", pos, err)
		}
		fmt.Println(line)
	}
	return nil
}

func Tac(args []string) error {
	if len(args) == 0 {
		return reverseStdin()
	}
	for _, file := range args {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		reverse(f)
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err == flags.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}
	if len(args) == 0 {
	}

	if err := Tac(args); err != nil {
		log.Fatal(err)
	}
}
