package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type FormatError struct {
	bad string
}

func (err FormatError) Error() string {
	return fmt.Sprintf("%q: invalid directive", err.bad)
}

func Fprintf(w io.Writer, format string, s []string) (n int, err error) {
	format = formatReplace(format)
	a, err := StrToIF(format, s)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Fprintf(w, format, a...)
}

func Printf(format string, s []string) (n int, err error) {
	return Fprintf(os.Stdout, format, s)
}

func StrToIF(f string, s []string) ([]interface{}, error) {
	a := []interface{}{}
	runes := []rune(f)
	argNum := 0
	percent := false
	verb := ""
	for _, r := range runes {
		if percent {
			verb += string(r)
			switch r {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			case '+', '-', '.':
			case '%':
				if verb != "%%" {
					return a, FormatError{verb}
				} else { // %%
					percent = false
					continue
				}
			case 's', 'q':
				if argNum < len(s) {
					a = append(a, s[argNum])
					argNum++
				} else {
					a = append(a, "")
				}
				percent = false
			case 'd', 'o', 'x', 'X':
				var i int64
				var err error
				if argNum < len(s) {
					i, err = strconv.ParseInt(s[argNum], 0, 64)
					if err != nil {
						i = 0
					}
					argNum++
				} else {
					i = 0
				}
				a = append(a, i)
				percent = false
			case 'f', 'F', 'e', 'E':
				var fl float64
				var err error
				if argNum < len(s) {
					fl, err = strconv.ParseFloat(s[argNum], 64)
					if err != nil {
						fl = 0.0
					}
					argNum++
				} else {
					fl = 0.0
				}
				a = append(a, fl)
				percent = false
			default:
				return a, FormatError{verb}
			}
		}
		if r == '%' {
			if percent == false {
				percent = true
				verb = "%"
			}
		}
	}
	if percent {
		return a, FormatError{verb}
	}
	return a, nil
}

func formatReplace(f string) string {
	f = strings.Replace(f, "\\a", "\a", -1)
	f = strings.Replace(f, "\\b", "\b", -1)
	f = strings.Replace(f, "\\f", "\f", -1)
	f = strings.Replace(f, "\\n", "\n", -1)
	f = strings.Replace(f, "\\r", "\r", -1)
	f = strings.Replace(f, "\\t", "\t", -1)
	f = strings.Replace(f, "\\v", "\v", -1)
	f = strings.Replace(f, "\\'", "'", -1)
	f = strings.Replace(f, "\\\"", "\"", -1)

	return f
}

func main() {
	if len(os.Args) < 2 {
		Fprintf(os.Stderr, "%s: not enough arguments\n", []string{os.Args[0]})
	} else if len(os.Args) == 2 {
		f := os.Args[1]
		Printf(f, []string{})
		//Printf(f, []string{})
	} else {
		f := os.Args[1]
		s := os.Args[2:]
		Printf(f, s)
	}
}
