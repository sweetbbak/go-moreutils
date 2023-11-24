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
	// if flags.Args() exists, then assume we will lookup those variables and exit
	Lookup bool
)

var (
	White  string = "\x1b[38;2;255;255;255m"
	Orange string = "\x1b[38;2;230;219;116m"
	Pink   string = "\x1b[38;2;249;38;114m"
	Clear  string = "\x1b[0m"
)

// iterate and lookup env vars that the user asks
func lookupVars(e []string) {
	for x := range e {
		env, envbool := os.LookupEnv(e[x])
		if envbool {
			fmt.Fprintf(os.Stdout, "%v\n", env)
		}
	}
	os.Exit(0)
}

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
	// split vars on the left most '=' color left side white, right side yellow/orange
	// and then re-join those args with a pink/red colored '='
	for _, x := range e {
		strs := strings.SplitN(x, "=", -1)
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

func init() {
	flag.Parse()
	if len(flag.Args()) > 0 {
		Lookup = true
	}
}

// i didn't know that printenv can be used like this `printenv BROWSER DIFFPAGER MANPAGER`
// and it will look up these keys lol I thought it just printed the env vars
func main() {
	// flag.Args() is the remaining arguments after flags
	if Lookup {
		lookupVars(flag.Args())
	} else {
		printenv(os.Stdout)
	}
}
