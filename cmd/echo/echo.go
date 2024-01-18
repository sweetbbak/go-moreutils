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

var (
	noNewline                 bool = false
	help                      bool = false
	interpretEscapes          bool = false
	interpretBackslashEscapes bool = false
	specialFormatter          bool = false
)

var usage = `echo <options> [strings]...
Usage:
	-e    interpret escape sequences
	-n    suppress newlines
	-E    disable interpretation of backslash sequences
	-f    use special formatting replacement strings for hex colors "{#1e1e1e}Hello{clear}"
Examples:
	echo -f "{#ff11aa}Hello{clear}{#DDD123}World{clear}\n\t:)"
`

func escapeStr(s string) (string, error) {
	if len(s) < 1 {
		return "", nil
	}

	s = strings.Split(s, "\\c")[0]
	s = strings.Replace(s, "\\0", "\\", -1)
	s = fmt.Sprintf("\"%s\"", s)

	_, err := fmt.Sscanf(s, "%q", &s)
	if err != nil {
		return "", nil
	}

	return s, nil
}

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
	r := regexp.MustCompile(`\{#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})\}`)

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

func echo(w io.Writer, noNewline, escape, backslash bool, s ...string) error {
	var err error

	if backslash {
		escape = false
	}

	line := strings.Join(s, " ")
	if escape {
		line, err = escapeStr(line)
		if err != nil {
			return err
		}
	}

	if specialFormatter {
		line = replaceColor(line)
	}

	format := "%s"
	if !noNewline {
		format += "\n"
	}

	_, err = fmt.Fprintf(w, format, line)
	return err
}

func parseargs() []string {

	var args []string
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "-h" || os.Args[i] == "--help" {
			fmt.Println(usage)
			os.Exit(0)
		}

		// fmt.Println(os.Args[i])
		if os.Args[i][0] == '-' {
			// fmt.Println("flag start")
			tacs := os.Args[i]
			for x := 0; x < len(tacs); x++ {
				if tacs[x] == ' ' {
					break
				}
				switch tacs[x] {
				case ' ':
					// fmt.Println("Space character - flags ended")
				case 'n':
					// fmt.Println("no newline")
					noNewline = true
				case 'f':
					// fmt.Println("format")
					specialFormatter = true
				case 'e':
					// fmt.Println("escapes")
					interpretEscapes = true
				case 'E':
					// fmt.Println("no escapes")
					interpretBackslashEscapes = true
				}
			}
		} else if i != 0 {
			args = append(args, os.Args[i])
		}
	}
	return args
}

func main() {
	args := parseargs()
	err := echo(os.Stdout, noNewline, interpretEscapes, interpretBackslashEscapes, args...)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
