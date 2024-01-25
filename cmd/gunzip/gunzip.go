package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	List     bool `short:"l" long:"list" description:"list information about the gzip archive"`
	Keep     bool `short:"k" long:"keep" description:"keep the source GZIP archive"`
	Force    bool `short:"f" long:"force" description:"force overwrite existing files if they exist"`
	Stdout   bool `short:"s" long:"stdout" description:"print decompressed contents to stdout"`
	Examples bool `short:"H" long:"examples" description:"print a few examples of uses"`
	Verbose  bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func printExamples() {
	fmt.Println("\x1b[32mgunzip\x1b[0m")
	fmt.Println("")
	fmt.Println("\x1b[33mExtract file(s) from a gzip (.gz) archive.\x1b[0m")
	fmt.Println("\x1b[33mMore information: https://manned.org/gunzip.\x1b[0m")
	fmt.Println("")
	fmt.Println("\x1b[32m- Extract a file from an archive, replacing the original file if it exists:\x1b[0m")
	fmt.Println("\x1b[31mgunzip archive.tar.gz\x1b[0m")
	fmt.Println("")
	fmt.Println("\x1b[32m- Extract a file to a target destination:\x1b[0m")
	fmt.Println("\x1b[31mgunzip --stdout archive.tar.gz > archive.tar\x1b[0m")
	fmt.Println("")
	fmt.Println("\x1b[32m- Extract a file and keep the archive file:\x1b[0m")
	fmt.Println("\x1b[31mgunzip --keep archive.tar.gz\x1b[0m")
	fmt.Println("")
	fmt.Println("\x1b[32m- List the contents of a compressed file:\x1b[0m")
	fmt.Println("\x1b[31mgunzip --list file.txt.gz\x1b[0m")
	fmt.Println("")
	fmt.Println("\x1b[32m- Decompress an archive from `stdin`:\x1b[0m")
	fmt.Println("\x1b[31mcat path/to/archive.gz | gunzip > archive\x1b[0m")
	os.Exit(1)
}

func Gunzip(args []string) error {
	if isOpen(os.Stdin) && len(args) == 0 {
		err := unzipStdin(os.Stdin) // unzip from stdin to stdout by default
		if err != nil {
			return err
		}
	}

	for _, file := range args {
		file = os.ExpandEnv(file)
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		if opts.List {
			gInfo(f)
		} else {
			err := unzip(f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isOpen(file *os.File) bool {
	o, _ := file.Stat()
	if (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice { //Terminal
		//Display info to the terminal
		return false
	} else { //It is not the terminal
		// Display info to a pipe
		return true
	}
}

func gInfo(file *os.File) error {
	fi, err := file.Stat()
	if err != nil {
		return err
	}

	size := fi.Size()

	dcmp, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer dcmp.Close()

	bs, err := io.Copy(io.Discard, dcmp)
	if err != nil {
		return err
	}

	ratio := PercentOf(int(size), int(bs))
	ratio = 100 - ratio

	fmt.Printf("compressed\tuncompressed\tratio\tname\n")
	fmt.Printf("%v\t\t%v\t\t%.2f\t%v\n", size, bs, ratio, file.Name())
	return nil
}

func PercentOf(part int, total int) float64 {
	return (float64(part) * float64(100)) / float64(total)
}

// man gunzip - if file is "-" stdin is decompressed to stdout
func unzipStdin(file *os.File) error {
	dcmp, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	if err == nil {
		defer dcmp.Close()
		if _, err := io.Copy(os.Stdout, dcmp); err != nil {
			return err
		}
	}
	return nil
}

func newName(filename string) (string, error) {
	ext := filepath.Ext(filename)
	switch ext {
	case ".gz", ".z", ".tgz", ".taz", ".Z":
		return strings.TrimSuffix(filename, ext), nil
	}
	if strings.HasSuffix(filename, "-gz") {
		return strings.TrimSuffix(filename, "-gz"), nil
	}
	if strings.HasSuffix(filename, "-z") {
		return strings.TrimSuffix(filename, "-z"), nil
	}
	if strings.HasSuffix(filename, "_z") {
		return strings.TrimSuffix(filename, "_z"), nil
	}
	return "", fmt.Errorf("Extension undetected")
}

// for file args
func unzip(file *os.File) error {
	dcmp, err := gzip.NewReader(file)
	outfile, err := newName(file.Name())
	if err != nil {
		return err
	}

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	// as per GNU open-time flags, create + excl will never overwrite an existing file
	// we also copy the OG files mode and all that
	var creation int
	if opts.Force {
		creation = os.O_WRONLY | os.O_CREATE
	} else {
		creation = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	}

	var out *os.File
	if opts.Stdout {
		out = os.Stdout
	} else {
		out, err = os.OpenFile(outfile, creation, fi.Mode())
		if err != nil {
			return err
		}
	}
	defer out.Close()

	if err == nil {
		defer dcmp.Close()
		if _, err := io.Copy(out, dcmp); err != nil {
			return fmt.Errorf("gunzip: %w\n", err)
		}
	} else {
		return fmt.Errorf("gunzip: %w\n", err)
	}
	return nil
}

func main() {
	// p := flags.NewParser(&opts, flags.Default)
	// p.WriteManPage(os.Stdout)
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Examples {
		printExamples()
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Gunzip(args); err != nil {
		log.Fatal(err)
	}
}
