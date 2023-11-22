package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	backup = flag.Bool("b", false, "backup existing destination file")
	update = flag.Bool("u", false, "backup existing destination file")
	force  = flag.Bool("f", false, "do not prompt before overwriting")
)

func init() {
	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
}
