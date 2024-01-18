package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Recursive        bool   `short:"r" long:"recursive" description:"recursively copy a file tree"`
	Ask              bool   `short:"i" long:"interactive" description:"ask before overwriting files"`
	Force            bool   `short:"f" long:"force" description:"forcefully overwrite files [use caution]"`
	noFollowSymlinks bool   `short:"P" long:"no-dereference" default:"true" description:"dont follow symlinks"`
	Target           string `short:"t" long:"target-dir" description:"directory to copy files into"`
	Verbose          bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}
var errSkip = errors.New("skip")

func stat(path string) (os.FileInfo, error) {
	if opts.noFollowSymlinks {
		return os.Lstat(path)
	}
	return os.Stat(path)
}

func copyRegularFile(src, dest string, srcInfo os.FileInfo) error {
	srcf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcf.Close()

	destfile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return err
	}
	defer destfile.Close()

	_, err = io.Copy(destfile, srcf)
	return err
}

func copyFile(src, dest string, srcInfo os.FileInfo) error {
	m := srcInfo.Mode()
	switch {
	case m.IsDir():
		return os.MkdirAll(dest, srcInfo.Mode().Perm())
	case m.IsRegular():
		return copyRegularFile(src, dest, srcInfo)
	case m&os.ModeSymlink == os.ModeSymlink:
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dest)

	default:
		return &os.PathError{
			Op:   "copy",
			Path: src,
			Err:  fmt.Errorf("unsupported file mode %s", m),
		}
	}
}

func copyPrep(src, dest string) error {
	srcInfo, err := stat(src)
	if err != nil {
		return err
	}

	if err := checkPreOperation(src, dest, srcInfo); err == errSkip {
		return nil
	} else if err != nil {
		return err
	}

	if err := copyFile(src, dest, srcInfo); err != nil {
		return err
	}

	if opts.Verbose {
		Debug("%v => %v\n", src, dest)
	}

	return nil
}

func copyTree(src, dest string) error {
	return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		return copyPrep(path, filepath.Join(dest, rel))
	})
}

func checkPreOperation(src, dest string, srcinfo os.FileInfo) error {
	if !opts.Recursive && srcinfo.IsDir() {
		fmt.Printf("cp: -r not specified, omitting directory %s\n", src)
		return errSkip
	}

	destInfo, err := os.Stat(dest)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("cp: %v: cant handle error %v\n", dest, err)
		return errSkip
	} else if err != nil {
		return nil
	}

	if os.SameFile(srcinfo, destInfo) {
		fmt.Printf("cp: %v and %v are the same file\n", src, dest)
		return errSkip
	}

	if opts.Ask && !opts.Force {
		overwrite, err := confirmOverwrite(dest)
		if err != nil {
			return err
		}

		if !overwrite {
			return errSkip
		}
	}
	return nil
}

func confirmOverwrite(dest string) (bool, error) {
	input := bufio.NewReader(os.Stdin)
	fmt.Printf("cp: overwrite %v? ", dest)
	answer, err := input.ReadString('\n')
	if err != nil {
		return false, err
	}

	if strings.ToLower(answer)[0] != 'y' {
		return false, nil
	}

	return true, nil
}

func Copy(args []string) error {
	toDir := false

	var from []string
	var to string

	if opts.Target != "" {
		from, to = args, opts.Target
	} else {
		from, to = args[:len(args)-1], args[len(args)-1]
	}

	toStat, err := os.Stat(to)
	if err == nil {
		toDir = toStat.IsDir()
	}

	if len(args) > 2 && !toDir {
		return fmt.Errorf("No target directory for multiple files")
	}

	var lastErr error
	for _, file := range from {
		dest := to
		if toDir {
			dest = filepath.Join(dest, filepath.Base(file))
		}

		if opts.Recursive {
			// recursive copy func
			lastErr = copyTree(file, dest)
		} else {
			// normal copy func
			lastErr = copyPrep(file, dest)
		}
	}

	return lastErr
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Copy(args); err != nil {
		log.Fatal(err)
	}
}
