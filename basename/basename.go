package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	suffix   = flag.String("s", "", "suffix to trim ie: /a/b.png -> b")
	zero     = flag.Bool("z", false, "print null byte instead of new line")
	multiple = flag.Bool("a", false, "support multiple arguments and treat each as a name")
)

func Basename(w io.Writer, args []string) {
	var names []string
	if *multiple {
		for x := range args {
			f := args[x]
			f = filepath.Base(f)
			if *suffix != "" {
				f = strings.TrimSuffix(f, *suffix)
			}
			names = append(names, f)
		}
	} else {
		f := filepath.Base(args[0])
		if *suffix != "" {
			f = strings.TrimSuffix(f, *suffix)
		}
		names = append(names, f)
	}
	for k := range names {
		if *zero {
			fmt.Fprintf(w, "%s\x00", names[k])
		} else {
			fmt.Fprintf(w, "%s\n", names[k])
		}
	}
}

func main() {
	flag.Parse()
	Basename(os.Stdout, flag.Args())
}
