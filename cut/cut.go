package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Delim   string `short:"d" long:"delimiter" description:"delimiter to use to split a string"`
	Fields  string `short:"f" long:"fields" description:"show fields"`
	Verbose bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func cuttup(file *os.File, fields *Fields) error {
	delim := opts.Delim
	if delim == "" {
		delim = " "
	}
	index := 0

	Debug("%v: fields: [%v] delimiter: [%v]\n", file.Name(), opts.Fields, delim)

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := sc.Text()
		array := strings.Split(line, delim)
		if len(array) == 0 {
			continue
		}

		seen := make(map[int]string)
		for in, split := range array {
			seen[in+1] = split // "cut" starts at 1 index and not 0 :O
		}

		// for in, item := range seen {
		// 	fmt.Println(in, item)
		// }

		var sb strings.Builder
		for _, f := range fields.Field {
			item, ok := seen[f]
			if ok {
				sb.WriteString(item)
				sb.WriteString(" ")
			}
		}
		fmt.Println(sb.String())
		sb.Reset()

		// maxx := len(array)
		// for _, split := range array {
		// 	for _, i := range fields.Field {
		// 		if i > maxx {
		// 			continue
		// 		}

		// 		if i >= 0 && i <= maxx {
		// 			seen[split] = true
		// 		}

		// 		if i < 0 {
		// 			ix := len(array) - i
		// 			if ix >= 0 && ix <= maxx {

		// 			}
		// 		}
		// 	}
		// }

		index++
	}
	return nil
}

func isNum(s string) bool {
	for _, i := range s {
		if i >= '0' && i <= '9' {
			continue
		} else {
			return false
		}
	}
	return true
}

func convertNum(s string) (int, error) {
	return strconv.Atoi(s)
}

type Range struct {
	i int
	r int
}

type Fields struct {
	Field  []int
	Ranges []Range
}

func parseFields(rawFields string) (*Fields, error) {
	var fields Fields
	f := strings.Split(rawFields, ",")
	if len(f) == 0 {
		return nil, fmt.Errorf("Fields length cannot be zero.")
	}

	for _, item := range f {
		switch {
		case strings.HasPrefix(item, "-"):
			item = item[1:]
			item, err := convertNum(item)
			if err != nil {
				return nil, err
			}
			fields.Ranges = append(fields.Ranges, Range{0, item})

		case strings.HasSuffix(item, "-"):
			item = item[0 : len(item)-1]
			item, err := convertNum(item)
			if err != nil {
				return nil, err
			}
			fields.Ranges = append(fields.Ranges, Range{item, -1})
		case strings.Contains(item, "-"):
			out := strings.Split(item, "-")
			if len(out) != 2 {
				return nil, fmt.Errorf("Invalid range")
			}
			item, err := convertNum(out[0])
			if err != nil {
				return nil, err
			}
			item1, err := convertNum(out[1])
			if err != nil {
				return nil, err
			}
			fields.Ranges = append(fields.Ranges, Range{item, item1})
		default:
			if isNum(item) {
				item, err := convertNum(item)
				if err != nil {
					return nil, err
				}
				fields.Field = append(fields.Field, item)
			}
		}
	}
	return &fields, nil
}

func Cut(args []string) error {
	field, err := parseFields(opts.Fields)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return cuttup(os.Stdin, field)
	}

	for _, path := range args {
		path = os.ExpandEnv(path)
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if err := cuttup(file, field); err != nil {
			return err
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

	if err := Cut(args); err != nil {
		log.Fatal(err)
	}
}
