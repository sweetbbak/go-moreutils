package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Canon        bool   `short:"e" long:"canonicalize-existing" description:"all components of the path must exist"`
	Logical      bool   `short:"L" long:"logical" description:"resolve .. before symlink"`
	Physcal      bool   `short:"P" long:"physical" description:"resolve symlinks"`
	Quiet        bool   `short:"q" long:"quiet" description:"suppress error messages"`
	Relative     string `short:"r" long:"relative" description:"print the path in relation to the given path"`
	RelativeBase string `short:"b" long:"relative-base" description:"print absolute paths unless the paths are below the given DIR"`
	Zero         bool   `short:"0" long:"zero" description:"end each line with a NUL byte, not a new line"`
	Strip        bool   `short:"s" long:"strip" description:"dont expand symlinks"`
}

func checkLink(path string) string {
	fi, err := os.Lstat(path)
	if err != nil {
		return path
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		p, err := os.Readlink(path)
		if err != nil {
		}
		return p
	}
	return path
}
func isSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink == os.ModeSymlink
}

func realpath(args []string) error {
	for _, p := range args {

		pp := checkLink(p)
		path, err := filepath.Abs(pp)
		if err != nil {
			return err
		}

		if opts.Relative != "" {
			if opts.Relative == "." {
				opts.Relative, _ = os.Getwd()
			}
			path, _ = filepath.Rel(opts.Relative, path)
		}
		fmt.Println(path)
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}
	if err := realpath(args); err != nil {
		log.Fatal(err)
	}
}
