package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func runYes(w io.Writer, count uint64, args ...string) error {
	yes := "y"
	if len(args) > 0 {
		yes = strings.Join(args, " ")
	}

	for {
		if _, err := fmt.Fprintf(w, "%s\n", yes); err != nil {
			return err
		}

		if count > 1 {
			count--
		} else if count == 1 {
			break
		}
	}

	return nil
}

func main() {
	flag.Parse()
	if err := runYes(os.Stdout, 0, flag.Args()...); err != nil {
		log.Fatal(err)
	}
}
