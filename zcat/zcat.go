package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	List    bool `short:"l" long:"list" description:"list information about the gzip archive"`
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Zcat(args []string) error {
	for _, file := range args {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		if opts.List {
			gInfo(f)
		} else {
			unzip(f)
		}
	}
	return nil
}

func isRedir() bool {
	o, _ := os.Stdout.Stat()
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
	ratio := bs / size

	fmt.Printf("compressed\tuncompressed\tratio\tname\n")
	fmt.Printf("%v %d %d %v\n", size, bs, ratio, dcmp.Name)
	return nil
}

func unzip(file *os.File) error {
	dcmp, err := gzip.NewReader(file)
	if err == nil {
		func() {
			defer dcmp.Close()
			if _, err := io.Copy(os.Stdout, dcmp); err != nil {
				fmt.Printf("zcat: %v\n", err)
			}
		}()
	} else {
		fmt.Printf("zcat: %v\n", err)
		return err
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Zcat(args); err != nil {
		log.Fatal(err)
	}
}
