// thanks to: https://github.com/moxtsuan/printf
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Hex string
type Ansi string
type RGB struct {
	R int
	G int
	B int
}

func HextoRGB(hex Hex) RGB {
	if hex[0:1] == "#" {
		hex = hex[1:]
	}

	r := string(hex)[0:2]
	g := string(hex)[2:4]
	b := string(hex)[4:6]

	R, _ := strconv.ParseInt(r, 16, 0)
	G, _ := strconv.ParseInt(g, 16, 0)
	B, _ := strconv.ParseInt(b, 16, 0)

	return RGB{int(R), int(G), int(B)}

}

func HextoAnsi(hex Hex) Ansi {
	rgb := HextoRGB(hex)
	str := "\x1b[38;2;" + strconv.FormatInt(int64(rgb.R), 10) + ";" + strconv.FormatInt(int64(rgb.G), 10) + ";" + strconv.FormatInt(int64(rgb.B), 10) + "m"
	return Ansi(str)
}

func replaceColor(line string) string {
	// matches 3 hex color and 6 hex color
	// r := regexp.MustCompile(`\{#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})\}`)
	r := regexp.MustCompile(`\{#(?:[0-9a-fA-F]{6})\}`)

	line = strings.ReplaceAll(line, "{clr}", "\x1b[0m")
	line = strings.ReplaceAll(line, "{clear}", "\x1b[0m")

	if r.Match([]byte(line)) {
		line = r.ReplaceAllStringFunc(line, func(line string) string {
			str := line
			// str = strings.ReplaceAll(str, "{clr}", "\x1b[0m")
			// str = strings.ReplaceAll(str, "{clr}", "")
			// str = strings.ReplaceAll(str, "{clear}", "")
			str = strings.ReplaceAll(str, "{", "")
			str = strings.ReplaceAll(str, "}", "")

			hex := HextoAnsi(Hex(str))
			return string(hex)
		})
		return line
	}
	return line
}

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
	//
	str := fmt.Sprintf(format, a...)
	if err != nil {
		fmt.Println(err)
	}
	str = replaceColor(str)
	return fmt.Fprintf(w, "%s", str)
	//

	// return fmt.Fprintf(w, format, a...)
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
	f = strings.Replace(f, "\\e", "\x1b", -1)

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
