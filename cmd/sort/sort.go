package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Blanks     bool `short:"b" long:"ignore-leading-blanks" description:"ignore leading blanks"`
	IgnoreCase bool `short:"i" long:"ignore" description:"fold lower case to upper case characters"`
	Reverse    bool `short:"r" long:"reverse" description:"reverse the results of comparisons"`
	Unique     bool `short:"u" long:"unique" description:"print only unique"`
	Numeric    bool `short:"n" long:"numeric" description:"compare and sort strings according to numerical value"`
	Verbose    bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

type Sorter []string

func (a Sorter) Len() int           { return len(a) }
func (a Sorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Sorter) Less(i, j int) bool { return strings.ToUpper(a[i]) < strings.ToUpper(a[j]) }
func (a Sorter) Less2(i, j int) bool {
	l := strings.TrimLeftFunc(a[i], unicode.IsSpace)
	r := strings.TrimLeftFunc(a[j], unicode.IsSpace)
	if l == r {
		return len(a[i]) >= len(a[j])
	}
	return l < r
}

func uniq(s string) string {
	sl := strings.Split(s, "\n")
	var sb strings.Builder

	m := make(map[string]bool)
	for _, line := range sl {
		if m[line] == true {
			continue
		} else {
			m[line] = true
			sb.WriteString(line + "\n")
		}
	}
	return sb.String()
}

func sortAlgo(s string) string {
	if len(s) == 0 {
		return ""
	}

	lines := strings.Split(s, "\n")
	var si sort.Interface
	si = Sorter(lines)
	sort.Sort(si)
	return strings.Join(lines, "\n") + "\n"
}

func Sort(args []string) error {
	var from []io.ReadCloser
	for _, v := range args {
		f, err := os.Open(v)
		if err != nil {
			return err
		}
		defer f.Close()

		from = append(from, f)
	}

	if len(args) == 0 {
		from = append(from, os.Stdin)
	}

	fileContents := []string{}
	for _, f := range from {
		bytes, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		s := string(bytes)
		fileContents = append(fileContents, s)

		// add newline
		if len(s) > 0 && s[len(s)-1] != '\n' {
			fileContents = append(fileContents, "\n")
			s = s[:len(s)-1]
		}
	}

	s := strings.Join(fileContents, "")
	// remove newline
	if len(s) > 0 && s[len(s)-1] != '\n' {
		s = s[:len(s)-1]
	}

	out := sortAlgo(s)
	if opts.Unique {
		out = uniq(out)
	}

	fmt.Print(out)
	return nil
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

	if err := Sort(args); err != nil {
		log.Fatal(err)
	}
}
