package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Regex      []string `short:"r" long:"regex" description:"use regex to match filenames"`
	Extensions []string `short:"e" long:"extension" description:"use regex to match filenames"`
	Names      []string `short:"n" long:"name" description:"use file globbing to match filenames (follows shell patterns)"`
	Types      []string `short:"t" long:"type" description:"use file attributes to match filenames (follows shell patterns)"`
	Root       string   `short:"R" long:"root" description:"set root directory to start searching from"`
	Relative   bool     `long:"relative" description:"print file names as paths relative to root directory"`
	Absolute   bool     `short:"a" long:"absolute" description:"print absolute file paths"`
	Color      bool     `short:"c" long:"color" description:"print paths with color"`
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
	MatchName  bool
	MatchRegex bool
	MatchExt   bool
)

func matchRegex(str string, patterns []string) (bool, error) {
	for _, p := range patterns {
		Debug("file: %v Pattern: %v\n", str, p)
		r, err := regexp.Compile(p)
		if err != nil {
			Debug("Bad regex pattern: %v\n", p)
			return false, errSkip
		}

		if r.Match([]byte(str)) {
			return true, nil
		}
	}
	return false, errSkip
}

func matchExt(str string, extensions []string) (bool, error) {
	Debug("file: %v Pattern: %v\n", str, extensions)

	for _, e := range extensions {
		if e == "" {
			continue
		}
		var mr string
		if strings.HasPrefix(e, ".") {
			mr = fmt.Sprintf(".*\\%s$", e)
		} else {
			mr = fmt.Sprintf(".*\\.%s$", e)
		}
		r := regexp.MustCompile(mr)
		if r.Match([]byte(str)) {
			return true, nil
		}
	}

	return false, errSkip
}

func matchGlob(str string, globs []string) (bool, error) {
	for _, p := range globs {
		return Glob(p, str), nil
	}
	return false, errSkip
}

func matchFileAttr(path string, info os.FileInfo) (bool, error) {
	// fileTypes := map[string]os.FileMode{
	// 	"f":         0,
	// 	"file":      0,
	// 	"d":         os.ModeDir,
	// 	"directory": os.ModeDir,
	// 	"s":         os.ModeSocket,
	// 	"p":         os.ModeNamedPipe,
	// 	"l":         os.ModeSymlink,
	// 	"c":         os.ModeCharDevice | os.ModeDevice,
	// 	"b":         os.ModeDevice,
	// }

	return false, errSkip
}

func walkFunc(path string, info os.FileInfo, err error) error {
	// allows for multiple matches of one file without double printing
	var printOut bool
	printOut = false

	var proot, ppath string
	if opts.Color {
		end := filepath.Base(path)
		beg := filepath.Dir(path)
		beg = fmt.Sprintf("%s/%s", opts.Root, beg)

		ppath = fmt.Sprintf("\x1b[38;2;219;88;100m%s\x1b[0m", end)
		proot = fmt.Sprintf("\x1b[38;2;89;182;227m%s/\x1b[0m", beg)
	}

	if opts.Absolute {
		path = filepath.Join(opts.Root, path)
	}

	if opts.Relative {
		path = filepath.Join(opts.Root, path)
		path, _ = filepath.Rel(CWD, path)
	}

	if MatchRegex && !printOut {
		m, err := matchRegex(path, opts.Regex)
		if err != nil {
			if err == errSkip {
				Debug("Skip: %v\n", path)
			}
		}
		if m {
			printOut = true
		}
	}

	if MatchExt && !printOut {
		m, err := matchExt(path, opts.Extensions)
		if err != nil {
			if err == errSkip {
				Debug("Skip: %v\n", path)
			}
		}
		if m {
			printOut = true
		}
	}

	if MatchName && !printOut {
		m, err := matchGlob(path, opts.Names)
		if err != nil {
			if err == errSkip {
				Debug("Skip: %v\n", path)
			}
		}
		if m {
			printOut = true
		}
	}

	if !MatchExt && !MatchName && !MatchRegex {
		fmt.Println(path)
		return nil
	}

	if printOut {
		if opts.Color {
			fmt.Printf("%s%s\n", proot, ppath)
		} else {
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
	if len(opts.Extensions) != 0 {
		MatchExt = true
	}
	if len(opts.Names) != 0 {
		MatchName = true
	}

	if opts.Absolute && opts.Relative {
		log.Fatalf("options error: opts --absolute and --relative are exclusive, you must specify one or the other")
	}

	if !opts.Absolute && !opts.Relative {
		opts.Absolute = true
	}

	CWD, _ = os.Getwd()

	if opts.Root == "" {
		opts.Root = CWD
	}

	if err := Find(args); err != nil {
		log.Fatal(err)
	}
}
