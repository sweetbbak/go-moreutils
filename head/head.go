package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	Chars   = flag.Int64("c", 0, "read N number of bytes only")
	Lines   = flag.Int("n", 10, "read N number of lines only")
	toLines = flag.Int("m", 1, "skip N number of lines")
	Quiet   = flag.Bool("q", false, "read n number of bytes only")

	passedChars bool
	passedLines bool
)

func byteHead(f io.Reader, n int64) error {
	lr := io.LimitReader(f, n)
	if _, err := io.Copy(os.Stdout, lr); err != nil {
		return err
	}
	return nil
}

func lineHead(f io.Reader, n int) error {
	scanner := bufio.NewScanner(f)
	line := 1
	for scanner.Scan() {
		if line < *toLines {
			line++
			continue
		}
		fmt.Println(scanner.Text())
		if line >= n {
			break
		}
		line++
	}
	return scanner.Err()
}

func Head(args []string) error {
	if len(args) == 0 || args[0] == "-" {
		if err := lineHead(os.Stdin, *Lines); err != nil {
			return err
		}
	}

	for _, file := range args {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if len(args) > 1 && !*Quiet && !passedChars {
			fmt.Printf("==> %s <==\n", file)
		}

		if passedChars {
			byteHead(f, *Chars)
		} else {
			lineHead(f, *Lines)
		}

	}
	return nil
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	flag.Parse()
	args := flag.Args()

	passedChars = isFlagPassed("c")
	passedLines = isFlagPassed("n")

	if passedChars && passedLines {
		fmt.Println("cannot pass -c [chars] and -n [lines] at the same time.")
		os.Exit(1)
	}

	if err := Head(args); err != nil {
		log.Fatal(err)
	}
}
