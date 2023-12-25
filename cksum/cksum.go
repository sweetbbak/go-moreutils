package main

import (
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Cksum(args []string) error {
	var file *os.File
	if len(args) == 0 || args[0] == "-" {
		file = os.Stdin
		if err := checksum(file, crc32.IEEE); err != nil {
			return err
		}
	}

	for _, f := range args {
		f = os.ExpandEnv(f)
		file, err := os.Open(f)
		if err != nil {
			return err
		}

		if err := checksum(file, crc32.IEEE); err != nil {
			return err
		}
	}

	return nil
}

func checksum(file *os.File, polynomial uint32) error {
	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	crc32q := crc32.MakeTable(crc32.IEEE)
	fmt.Printf("%d %d %s\n", crc32.Checksum(content, crc32q), len(content), file.Name())
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

	if err := Cksum(args); err != nil {
		log.Fatal(err)
	}
}
