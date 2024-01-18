package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	gzip "github.com/klauspost/pgzip"
)

var opts struct {
	Decompress  bool   `short:"d" long:"decompress" description:"decompress a gzip archive"`
	Force       bool   `short:"f" long:"force" description:"force overwrite existing files if they exist"`
	Procs       int    `short:"p" long:"procs" description:"number of CPU threads to use for processing"`
	CompressLvl int    `long:"level" description:"level of compression N [1-9] where 1 is fast and 9 is intensive & higher compression"`
	BlockSize   int    `short:"b" long:"blocks" description:"set compression block size in KiB, default is 128"`
	Suffix      string `short:"S" long:"suffix" description:"use SUF on compressed files"`
	List        bool   `short:"l" long:"list" description:"list information about the gzip archive"`
	Keep        bool   `short:"k" long:"keep" description:"keep the source GZIP archive"`
	Stdout      bool   `short:"c" long:"stdout" description:"print decompressed contents to stdout"`
	Verbose     bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`

	One   bool `short:"1" hidden:"true" description:"compression level"`
	Two   bool `short:"2" hidden:"true" description:"compression level"`
	Three bool `short:"3" hidden:"true" description:"compression level"`
	Four  bool `short:"4" hidden:"true" description:"compression level"`
	Five  bool `short:"5" hidden:"true" description:"compression level"`
	Six   bool `short:"6" hidden:"true" description:"compression level"`
	Seven bool `short:"7" hidden:"true" description:"compression level"`
	Eight bool `short:"8" hidden:"true" description:"compression level"`
	Nine  bool `short:"9" hidden:"true" description:"compression level"`
}

var Debug = func(string, ...interface{}) {}

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

func unzip(file *os.File) error {
	var outfile string
	var err error

	if opts.Decompress {
		outfile, err = newName(file.Name())
		if err != nil {
			return err
		}

		if opts.Suffix != "" {
			outfile = outfile + opts.Suffix
		}
	} else {
		if opts.Suffix != "" {
			outfile = file.Name() + opts.Suffix
		} else {
			outfile = file.Name() + ".gz"
		}
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

	if opts.Decompress {
		return Decompress(file, out, opts.BlockSize, opts.Procs)
	} else {
		return Compress(file, out, opts.CompressLvl, opts.BlockSize, opts.Procs)
	}
}

func Gzip(args []string) error {
	if isOpen(os.Stdin) && len(args) == 0 {
		if opts.Decompress {
			err := Decompress(os.Stdin, os.Stdout, opts.BlockSize, opts.Procs)
			if err != nil {
				return err
			}
		} else {
			err := Compress(os.Stdin, os.Stdout, opts.CompressLvl, opts.BlockSize, opts.Procs)
			if err != nil {
				return err
			}
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

func unzipStdin(file *os.File) error {
	return Decompress(file, os.Stdout, opts.BlockSize, opts.Procs)
}

func Decompress(r io.Reader, w io.Writer, blocksize int, procs int) error {
	zr, err := gzip.NewReaderN(r, blocksize*1024, procs)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, zr); err != nil {
		zr.Close()
		return err
	}

	return zr.Close()
}

func Compress(r io.Reader, w io.Writer, level int, blocksize int, processes int) error {
	zw, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		return err
	}

	if err := zw.SetConcurrency(blocksize*1024, processes); err != nil {
		zw.Close()
		return err
	}

	if _, err := io.Copy(zw, r); err != nil {
		zw.Close()
		return err
	}

	return zw.Close()
}

func compressionLevel(args []string) int {
	if opts.CompressLvl != 0 {
		if opts.CompressLvl >= 1 && opts.CompressLvl <= 9 {
			return opts.CompressLvl
		}
	}
	return 6
}

func compLevel() {
	if opts.One {
		opts.CompressLvl = 1
	}
	if opts.Two {
		opts.CompressLvl = 2
	}
	if opts.Three {
		opts.CompressLvl = 3
	}
	if opts.Four {
		opts.CompressLvl = 4
	}
	if opts.Five {
		opts.CompressLvl = 5
	}
	if opts.Six {
		opts.CompressLvl = 6
	}
	if opts.Seven {
		opts.CompressLvl = 7
	}
	if opts.Eight {
		opts.CompressLvl = 8
	}
	if opts.Nine {
		opts.CompressLvl = 9
	}
}

func main() {
	opts.Procs = 0
	opts.CompressLvl = 6
	opts.BlockSize = 128

	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if opts.Procs == 0 {
		opts.Procs = runtime.NumCPU()
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	// process short opts -9 etc, then verify compression level is [1-9]. if err return 6 as default
	compLevel()
	opts.CompressLvl = compressionLevel(args)

	if err := Gzip(args); err != nil {
		log.Fatal(err)
	}
}
