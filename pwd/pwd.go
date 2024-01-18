package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	_        = flag.Bool("L", true, "don't follow symlinks")
	physical = flag.Bool("P", false, "follow all symlinks")
	pretty   = flag.Bool("p", false, "pretty colored output")
)

func pwd(followSym bool) (string, error) {
	path, err := os.Getwd()
	if err == nil && followSym {
		path, err = filepath.EvalSymlinks(path)
	}

	if *pretty {
		path = fmt.Sprintf("\x1b[35m%s", path)
	}

	return path, err
}

func main() {
	flag.Parse()
	path, err := pwd(*physical)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(path)
}
