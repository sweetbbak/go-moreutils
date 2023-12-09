package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Line   bool `short:"l" description:"count lines"`
	Words  bool `short:"w" description:"count words"`
	Runes  bool `short:"r" description:"count runes"`
	Broken bool `short:"b" description:"count broken characters"`
	Chars  bool `short:"c" description:"count characters"`
}

var (
	StdinOpen = false
)

type Count struct {
	lines, words, runes, badRunes, chars int64
}

func stdOpen(std *os.File) bool {
	stat, _ := std.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// fmt.Println("data is being piped to stdin")
		return true
	} else {
		// fmt.Println("stdin is from a terminal")
		return false
	}
}

// A modified version of utf8.Valid()
func invalidCount(p []byte) (n int64) {
	i := 0
	for i < len(p) {
		if p[i] < utf8.RuneSelf {
			i++
		} else {
			_, size := utf8.DecodeRune(p[i:])
			if size == 1 {
				// All valid runes of size 1 (those
				// below RuneSelf) were handled above.
				// This muse be a RuneError.
				n++
			}
			i += size
		}
	}
	return
}

func count(in io.Reader, fname string) Count {
	// b := bufio.NewReaderSize(in, 8192)
	b := bufio.NewReader(in)
	counted := false
	count := Count{}
	for !counted {
		line, err := b.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				counted = true
			} else {
				fmt.Fprintf(os.Stderr, "wc: %s: %v\n", fname, err)
				return Count{}
			}
		}
		if !counted {
			count.lines++
		}
		count.words += int64(len(bytes.Fields(line)))
		count.runes += int64(utf8.RuneCount(line))
		count.chars += int64(len(line))
		count.badRunes += invalidCount(line)
	}
	return count
}

func printCount(c Count, fname string) {
	fields := []string{}
	if opts.Line {
		fields = append(fields, fmt.Sprintf("%d", c.lines))
	}
	if opts.Words {
		fields = append(fields, fmt.Sprintf("%d", c.words))
	}
	if opts.Runes {
		fields = append(fields, fmt.Sprintf("%d", c.runes))
	}
	if opts.Broken {
		fields = append(fields, fmt.Sprintf("%d", c.badRunes))
	}
	if opts.Chars {
		fields = append(fields, fmt.Sprintf("%d", c.chars))
	}
	if fname != "" {
		fields = append(fields, fname)
	}
	fmt.Fprintln(os.Stdout, strings.Join(fields, " "))
}

func wordcount(args []string) error {
	var total Count
	if !opts.Broken && !opts.Chars && !opts.Line && !opts.Runes && !opts.Words {
		opts.Line, opts.Words, opts.Chars = true, true, true
	}

	if stdOpen(os.Stdin) || len(args) == 0 {
		c := count(os.Stdin, "")
		printCount(c, "")
		return nil
	}

	for _, fn := range args {
		f, err := os.Open(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wc: %s: %v", fn, err)
			continue
		}
		result := count(f, fn)

		total.lines += result.lines
		total.words += result.words
		total.runes += result.runes
		total.chars += result.chars
		total.badRunes += result.badRunes
		printCount(result, fn)
	}

	if len(args) > 1 {
		printCount(total, "total")
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if err := wordcount(args); err != nil {
		log.Fatal(err)
	}
}
