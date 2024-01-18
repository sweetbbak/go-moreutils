package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"text/tabwriter"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	all       bool `short:"a" long:"all" description:""`
	human     bool `short:"h" long:"human" description:""`
	directory bool `short:"d" long:"directory" description:""`
	long      bool `short:"l" long:"long" description:""`
	quoted    bool `short:"Q" long:"quote-name" description:""`
	recurse   bool `short:"R" long:"recursive" description:""`
	classify  bool `short:"F" long:"classify" description:""`
	size      bool `short:"s" long:"size" description:""`
}

type file struct {
	path string
	osfi os.FileInfo
	// lsfi ls.FileInfo
	err error
}

func listFiles(d string) error {
	var dir []string
	stat, err := os.Lstat(d)
	if err != nil {
		fmt.Println(err)
	}

	if stat.IsDir() {
		dir = append(dir, d)
	}

	for _, x := range dir {
		fs, err := os.ReadDir(x)
		if err != nil {
			fmt.Println(err)
		}
		for _, f := range fs {
			fmt.Println(f.Name())
		}
	}

	return nil
}

func list(w io.Writer, args []string) error {
	if len(args) == 0 {
		args = []string{"."}
	}
	// write output using tabwriter
	tw := &tabwriter.Writer{}
	tw.Init(w, 0, 0, 1, ' ', 0)
	defer tw.Flush()
	for _, d := range args {
		if err := listFiles(d); err != nil {
			return fmt.Errorf("error while listing %q: %w", d, err)
		}
	}

	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if err := list(os.Stdout, args); err != nil {
		log.Fatal(err)
	}
}
