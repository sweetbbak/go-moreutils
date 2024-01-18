package main

import (
	"archive/zip"
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/jessevdk/go-flags"
	"github.com/ybirader/pzip"
)

var opts struct {
	Outfile    string `short:"o" long:"output" description:"output zip archive to compress files to"`
	Force      bool   `short:"f" long:"force" description:"force overwrite existing files if they exist"`
	Recursive  bool   `short:"r" long:"recursive" description:"recursively compress directories"`
	Concurrent int    `short:"c" long:"concurrency" description:"number of workers to use, more is faster but uses more CPU."`
	List       bool   `short:"l" long:"list" description:"list information about the gzip archive"`
	Verbose    bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Zip(args []string) error {
	var targetArchive string
	var creation int

	if opts.Outfile == "" {
		if len(args) < 2 {
			return fmt.Errorf("Must provide an archive")
		} else {
			targetArchive = args[0]
			args = args[1:]
		}
	} else {
		targetArchive = opts.Outfile
	}

	if opts.Force {
		creation = os.O_CREATE | os.O_TRUNC | os.O_RDWR
	} else {
		creation = os.O_CREATE | os.O_EXCL | os.O_RDWR
	}

	arc, err := os.OpenFile(targetArchive, creation, 0o644)
	if err != nil {
		return err
	}

	archiver, err := pzip.NewArchiver(arc, pzip.ArchiverConcurrency(opts.Concurrent))
	if err != nil {
		return err
	}
	defer archiver.Close()

	err = archiver.Archive(context.Background(), args)
	if err != nil {
		return err
	}

	return nil
}

func Info(file string) error {
	r, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	fmt.Printf("Archive: %-2s\n", file)
	fmt.Printf("  %-11s %-7s %-7s %-7s\n", "Length", "Date", "Time", "Name")
	fmt.Printf("---------  ---------- -----   ----\n")
	var size uint64
	for _, fi := range r.File {
		printInfo(fi)
		size += fi.UncompressedSize64
	}
	fmt.Printf("---------                     ------\n")
	fmt.Printf("%-30v %v files\n", size, len(r.File))
	return nil
}

func printInfo(fi *zip.File) {
	time := fi.Modified.UTC()
	tz := time.Format("2006-01-02 15:04")
	fmt.Printf("%-10v %-18v %-5v\n", fi.UncompressedSize64, tz, fi.Name)
}

func init() {
	opts.Concurrent = runtime.NumCPU()
}

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if opts.List {
		for _, f := range args {
			if err := Info(f); err != nil {
				log.Println(err)
			}
		}
		os.Exit(0)
	}

	if err := Zip(args); err != nil {
		log.Fatal(err)
	}
}
