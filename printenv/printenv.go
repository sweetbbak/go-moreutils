package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	pretty = flag.Bool("p", false, "colorize output of print-env")
)

func printenv(w io.Writer) {
	e := os.Environ()

	if *pretty {
		prettyPrint(e)
	} else {
		for _, x := range e {
			fmt.Fprintf(w, "%v\n", x)
		}
	}
}

func prettyPrint(e []string) {
	for _, x := range e {
		strs := strings.SplitN(x, "=", -1)
		// strs[0] = fmt.Sprintf("%v%v%v", White, strs[0], Clear)
		// strs[1] = fmt.Sprintf("%v%v%v", Orange, strs[1], Clear)

		// fmt.Println(strs)

		// for i = range strs[1:] {
		for i := 0; i < len(strs); i++ {

			switch i {
			case 0:
				strs[i] = fmt.Sprintf("%v%v%v", White, strs[i], Clear)
			default:
				strs[i] = fmt.Sprintf("%v%v%v", Orange, strs[i], Clear)
			}
		}

		sep := fmt.Sprintf("%v%v%v", Pink, "=", Clear)
		s := strings.Join(strs, sep)
		fmt.Println(s)
	}
}

func white(s string) string {
	return fmt.Sprintf("%x%v", White, s)
}

var (
	White  string = "\x1b[38;2;255;255;255m"
	Orange string = "\x1b[38;2;230;219;116m"
	Pink   string = "\x1b[38;2;249;38;114m"
	Clear  string = "\x1b[0m"
)

func init() {
	flag.Parse()
}

func main() {
	printenv(os.Stdout)
}
