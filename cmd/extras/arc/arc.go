package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Compress   bool   `short:"c" long:"compress" description:"compress an archive"`
	Decompress bool   `short:"d" long:"decompress" description:"decompress an archive"`
	Append     bool   `short:"a" long:"add" description:"add specified files to an archive"`
	Type       string `short:"t" long:"type" description:"archive type to use [7z|bzip2|rar|tar|tar.gz|zip] default is inferred by extension"`
	Outfile    string `short:"o" long:"out" description:"path to [de]compressed output archive/dir depending on operation"`
	List       bool   `short:"l" long:"list" description:"list all supported compression|decompression formats"`
	Verbose    bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Unarc(args []string) error {
	if len(args) < 1 && opts.Outfile == "" {
		return fmt.Errorf("must provide an output archive and files")
	}

	var output string
	var err error
	if len(args) == 1 && opts.Outfile == "" {
		output, err = os.Getwd()
	} else if opts.Outfile == "" {
		output = args[len(args)-1]
		args = args[:len(args)-1]
	}

	err = os.MkdirAll(output, 0o755)
	if err != nil {
		return err
	}

	for _, archive := range args {
		if err := Extract7z(archive, output); err != nil {
			return err
		}
	}

	return nil
}

func Arc(args []string) error {
	if len(args) < 2 && opts.Outfile == "" {
		return fmt.Errorf("must provide an output archive and files")
	}

	var archive string

	if opts.Outfile != "" {
		archive = opts.Outfile
	} else {
		archive = args[0]
		args = args[1:]
	}

	if err := CompressTo7z(archive, args); err != nil {
		return err
	}

	return nil
}

func extensions(file string) {
	switch file {
	case "tar.br", ".tbr", ".br":
	case "tar.bz2", ".tbz2", ".bz2":
	case ".gz", ".tar.gz":
	case ".tar.lz4", ".tlz4", ".lz4":
	case ".tar.sz", ".sz", "tsz":
	case "tar.xz", ".txz", ".xz":
	case "tar.zst", ".zst":
	case ".zip":
	default:

	}
}

var arcTypes = `Archive Formats:
  Format  Extensions
  ------  -----------
  gzip,   [ .tar.gz  | .tgz  | .gz  ]
  bzip2,  [ .tar.bz2 | .tbz2 | .bz2 ]
  xz,     [ .tar.xz  | .txz  | .xz  ]
  lzma,   [ .tar.lz4 | .tlz4 | .lz4 ]
  snappy, [ .tar.sz  | .tsz  | .sz  ]
  brotli, [ .tar.br  | .tbr  | .br  ]
  zst,    [ .tar.zst | .zst ]
  zip,    [ .zip ]`

func List() {
	fmt.Println(arcTypes)
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
		List()
	}

	if opts.Compress {
		if err := Arc(args); err != nil {
			log.Fatal(err)
		}
	}

	if opts.Decompress {
		if err := Unarc(args); err != nil {
			log.Fatal(err)
		}
	}
}
