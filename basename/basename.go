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
	multiple = flag.Bool("a", true, "support multiple arguments and treat each as a name")
	infer    = flag.Bool("i", false, "infer the suffix/extension to trim")
)

// idk why basename just doesnt automatically spit out all the basenames by default but thats what it does sooo...
func Basename(w io.Writer, args []string) {
	var names []string

	if *multiple {
		for x := range args {
			f := args[x]
			f = filepath.Base(f)
			if *suffix != "" {
				f = strings.TrimSuffix(f, *suffix)
			}

			if *infer && *suffix == "" {
				parts := strings.Split(f, ".")
				if len(parts) > 1 {
					suf := fmt.Sprintf(".%s", parts[len(parts)-1])
					f = strings.TrimSuffix(f, suf)
				}
			}

			names = append(names, f)
		}
	} else {
		f := filepath.Base(args[0])
		if *suffix != "" {
			f = strings.TrimSuffix(f, *suffix)
		}

		if *infer && *suffix == "" {
			parts := strings.Split(f, ".")
			if len(parts) > 1 {
				suf := fmt.Sprintf(".%s", parts[len(parts)-1])
				f = strings.TrimSuffix(f, suf)
			}
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
