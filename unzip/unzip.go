package main

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/ybirader/pzip"
)

var opts struct {
	Directory   string `short:"d" long:"directory" description:"directory to output unzipped files (it will be made if it doesnt exist)"`
	Force       bool   `short:"f" long:"force" description:"force overwrite existing files if they exist"`
	Suffix      string `short:"S" long:"suffix" description:"use SUF on compressed files"`
	ExFile      string `short:"j" long:"just" description:"extract specified files from an archive. [-j ark.zip file1 file2]"`
	List        bool   `short:"l" long:"list" description:"list information about the gzip archive"`
	Concurrency int    `short:"n" long:"concurrency" description:"number of concurrent workers, more is faster but uses more CPU"`
	Keep        bool   `short:"k" long:"keep" description:"keep the source GZIP archive"`
	Stdout      bool   `short:"c" long:"stdout" description:"print decompressed contents to stdout"`
	Verbose     bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func unzip2(file, dest string) error {
	extractor, err := pzip.NewExtractor(dest, pzip.ExtractorConcurrency(opts.Concurrency))
	if err != nil {
		return err
	}
	defer extractor.Close()

	err = extractor.Extract(context.Background(), file)
	if err != nil {
		return err
	}
	return nil
}

func unzip(file, destination string) error {
	r, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(destination, 0o755)
	if err != nil {
		return err
	}

	_, err = os.Stat(file)
	if err != nil {
		return err
	}

	fmt.Printf("Archive: %-2s\n", file)
	fmt.Printf("  %-11s %-7s %-7s %-7s\n", "Length", "Date", "Time", "Name")
	fmt.Printf("---------  ---------- -----   ----\n")
	var lenSize uint64

	for _, f := range r.File {
		if !opts.List {
			err := unzipWrite(destination, f)
			if err != nil {
				return err
			}
		}

		printInfo(f)
		lenSize += f.UncompressedSize64
	}
	fmt.Printf("---------                     ------\n")
	fmt.Printf("%-30v %v files\n", lenSize, len(r.File))

	return nil
}

func printInfo(fi *zip.File) {
	time := fi.Modified.UTC()
	tz := time.Format("2006-01-02 15:04")
	fmt.Printf("%-10v %-18v %-5v\n", fi.UncompressedSize64, tz, fi.Name)
}

func unzipWrite(dest string, file *zip.File) error {
	zf, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := zf.Close(); err != nil {
			panic(err)
		}
	}()

	path := filepath.Join(dest, file.Name)
	if file.FileInfo().IsDir() {
		err := os.MkdirAll(path, file.Mode())
		if err != nil {
			return err
		}
	} else { // is a file
		var creationFlags int
		creationFlags = os.O_CREATE | os.O_EXCL | os.O_WRONLY

		if !opts.Force {
			_, err := os.Stat(path)
			if err == nil {
				var rpl bool
				rpl, path, err = overwritePrompt(path)
				if err != nil {
					return err
				}

				if rpl {
					creationFlags = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
				}
			}
		}

		err := os.MkdirAll(filepath.Dir(path), file.Mode())
		if err != nil {
			return err
		}

		f, err := os.OpenFile(path, creationFlags, file.Mode())
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, zf)
		if err != nil {
			return err
		}
	}
	return nil
}

func overwritePrompt(file string) (bool, string, error) {
	s := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("replace %s? [y]es, [n]o, [A]ll, [N]one, [r]ename: ", file)

		response, err := s.ReadString('\n')
		if err != nil {
			return false, file, err
		}

		response = strings.TrimSpace(response)

		switch response {
		case "y", "yes":
			return true, file, nil
		case "n", "no":
			return false, file, nil
		case "N":
			opts.Force = false
			return false, file, nil
		case "A":
			opts.Force = true
			return true, file, nil
		case "r":
			base := filepath.Dir(file)
			fn, err := getRename(filepath.Base(file))
			if err != nil {
				return false, file, err
			}
			np := filepath.Join(base, fn)
			return false, np, err
		}
	}
}

func getRename(og string) (string, error) {
	s := bufio.NewReader(os.Stdin)
	var response string
	var err error
	for {
		fmt.Print("new name: ")
		response, err = s.ReadString('\n')
		if err != nil {
			return og, err
		} else {
			break
		}
	}
	return response, nil
}

func Unzip(args []string) error {
	for _, file := range args {
		file = os.ExpandEnv(file)
		err := unzip2(file, opts.Directory)
		if err != nil {
			return err
		}
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

	fstat, err := os.Stat(file)
	if err != nil {
		return err
	}

	for _, fi := range r.File {
		fmt.Printf("Archive: %-2s\n", file)
		fmt.Printf("%-2s %-8s %-12s %-14s", "Length", "Date", "Time", "Name")
		fmt.Printf("---------  ---------- -----   ----\n")
		fmt.Printf("%v %v %v\n", fi.CompressedSize64, fi.Modified, fi.Name)
		fmt.Printf("---------                     ------\n")
	}

	fmt.Printf("%v %-14v files\n", fstat.Size(), len(r.File))
	return nil
}

// TODO this really really needs a refactor. User input doesn't work well either
// so far unzip -l and unzip -d output asdf.zip works well
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

	if opts.Directory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		opts.Directory = cwd
	} else {
		opts.Directory = os.ExpandEnv(opts.Directory)
	}

	if err := Unzip(args); err != nil {
		log.Fatal(err)
	}
}
