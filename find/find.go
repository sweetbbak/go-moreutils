package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Regex      []string `short:"r" long:"regex" description:"use regex to match filenames"`
	Extensions []string `short:"e" long:"extension" description:"use regex to match filenames"`
	Root       string   `short:"R" long:"root" description:"set root directory to start searching from"`
	Relative   bool     `long:"relative" description:"print file names as paths relative to root directory"`
	Absolute   bool     `short:"a" long:"absolute" description:"print absolute file paths"`
	Verbose    bool     `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}
var errSkip = errors.New("Skip")
var CWD string

type finder struct {
	root       string
	pattern    []string
	extension  []string
	match      func(pattern []string, name string) (bool, error)
	mode       os.FileMode
	mask       os.FileMode
	files      chan *File
	sendErrors bool
}

type File struct {
	Name string
	os.FileInfo
	Err error
}

var (
	MatchPattern bool
	MatchRegex   bool
)

func matchRegex(str string, patterns []string) (bool, error) {
	for _, p := range patterns {
		Debug("file: %v Pattern: %v\n", str, p)
		r := regexp.MustCompile(p)
		if r.Match([]byte(str)) {
			return true, nil
		}
	}
	return false, errSkip
}

func walkFunc(path string, info os.FileInfo, err error) error {
	// fmt.Println(path)
	if opts.Absolute {
		path = filepath.Join(opts.Root, path)
	}

	if opts.Relative {
		path = filepath.Join(opts.Root, path)
		path, _ = filepath.Rel(CWD, path)
	}

	if MatchRegex {
		m, err := matchRegex(path, opts.Regex)
		if err != nil {
		}
		if m {
			fmt.Println(path)
		}
	}
	return nil
}

func Find(args []string) error {
	root := "/"
	if opts.Root != "" {
		root = opts.Root
	}

	err := Walk(root, walkFunc)
	return err
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		if err == flags.ErrHelp {
			os.Exit(0)
		}
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if len(opts.Regex) != 0 {
		MatchRegex = true
	}

	if opts.Absolute && opts.Relative {
		log.Fatalf("--absolute and --relative are exclusive, you must specify one or the other")
	}

	if !opts.Absolute && !opts.Relative {
		opts.Absolute = true
	}

	CWD, _ = os.Getwd()

	if err := Find(args); err != nil {
		log.Fatal(err)
	}
}
