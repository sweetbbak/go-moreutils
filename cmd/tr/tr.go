package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"unicode"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Delete  bool `short:"d" long:"delete" description:"delete characters in SET"`
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

var escapeCharacters = map[rune]rune{
	'\\': '\\',
	'a':  '\a',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
}

type Set string

const (
	ALPHA Set = "[:alpha:]"
	DIGIT Set = "[:digit:]"
	GRAPH Set = "[:graph:]"
	CNTRL Set = "[:cntrl:]"
	PUNCT Set = "[:punct:]"
	SPACE Set = "[:space:]"
	ALNUM Set = "[:alnum:]"
	LOWER Set = "[:lower:]"
	UPPER Set = "[:upper:]"
)

var sets = map[Set]func(r rune) bool{
	ALNUM: func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	},

	ALPHA: unicode.IsLetter,
	DIGIT: unicode.IsDigit,
	GRAPH: unicode.IsGraphic,
	CNTRL: unicode.IsControl,
	PUNCT: unicode.IsPunct,
	SPACE: unicode.IsSpace,
	LOWER: unicode.IsLower,
	UPPER: unicode.IsUpper,
}

type transformer struct {
	transform func(r rune) rune
}

func unescape(s Set) ([]rune, error) {
	var out []rune
	var escape bool
	for _, r := range s {
		if escape {
			v, ok := escapeCharacters[r]
			if !ok {
				return nil, fmt.Errorf("unknown escape sequence '\\%c'", r)
			}
			out = append(out, v)
			escape = false
			continue
		}

		if r == '\\' {
			escape = true
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func lowerToUpper() *transformer {
	return &transformer{
		transform: func(r rune) rune {
			return unicode.ToUpper(r)
		},
	}
}

func upperToLower() *transformer {
	return &transformer{
		transform: func(r rune) rune {
			return unicode.ToLower(r)
		},
	}
}

func setToRune(s Set, outRune rune) *transformer {
	check := sets[s]
	return &transformer{
		transform: func(r rune) rune {
			if check(r) {
				return outRune
			}
			return r
		},
	}
}

func runesToRunes(in []rune, out ...rune) *transformer {
	convs := make(map[rune]rune)
	l := len(out)
	for i, r := range in {
		ind := i
		if i > l-1 {
			ind = l - 1
		}
		convs[r] = out[ind]
	}
	return &transformer{
		transform: func(r rune) rune {
			if outRune, ok := convs[r]; ok {
				return outRune
			}
			return r
		},
	}
}

func parse(args []string) (*transformer, error) {
	narg := len(args)

	switch {
	case narg == 0 || (narg == 1 && !opts.Delete):
		return nil, fmt.Errorf("missing operand")
	case narg > 1 && opts.Delete:
		return nil, fmt.Errorf("extra operand after %v", args[0])
	case narg > 2:
		return nil, fmt.Errorf("extra operand after %v", args[1])
	}

	set1 := Set(args[0])
	arg1, err := unescape(set1)
	if err != nil {
		return nil, err
	}

	var set2 Set
	if opts.Delete {
		set2 = Set(unicode.ReplacementChar)
	} else {
		set2 = Set(args[1])
	}

	if set1 == LOWER && set2 == UPPER {
		return lowerToUpper(), nil
	}
	if set1 == UPPER && set2 == LOWER {
		return upperToLower(), nil
	}
	if (set2 == LOWER || set2 == UPPER) && (set1 != LOWER && set1 != UPPER) ||
		(set1 == LOWER && set2 == LOWER) || (set1 == UPPER && set2 == UPPER) {
		return nil, fmt.Errorf("misaligned [:upper:] and/or [:lower:] construct")
	}
	if _, ok := sets[set2]; ok {
		return nil, fmt.Errorf(`the only character classes that may appear in SET2 are 'upper' and 'lower'`)
	}
	arg2, err := unescape(set2)
	if err != nil {
		return nil, err
	}
	if len(arg2) == 0 {
		return nil, fmt.Errorf("SET 2 must be non-empty")
	}
	if _, ok := sets[set1]; ok {
		return setToRune(set1, arg2[0]), nil
	}
	return runesToRunes(arg1, arg2...), nil
}

func Truncate(args []string) error {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	tr, err := parse(args)
	if err != nil {
		return err
	}

	defer out.Flush()

	for {
		inRune, size, err := in.ReadRune()
		if inRune == unicode.ReplacementChar {
			in.UnreadRune()

			b, err := in.ReadByte()
			if err != nil {
				return err
			}

			if err := out.WriteByte(b); err != nil {
				return err
			}
		} else if size > 0 {
			if outRune := tr.transform(inRune); outRune != unicode.ReplacementChar {
				if _, err := out.WriteRune(outRune); err != nil {
					return err
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Truncate(args); err != nil {
		log.Fatal(err)
	}
}
