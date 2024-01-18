package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/u-root/u-root/pkg/cpio"
)

// TODO write my own CPIO libs but damn that is extremely extra right now
// like that is a lot of work right there for a tool that is barely used outside
// of creating an initramfs
// I have to do a lot more research on this one to fully understand what even needs to be done

var opts struct {
	Format  string `short:"H" long:"format" default:"newc" description:"cpio archive format"`
	Null    bool   `short:"0" long:"null" description:"files on stdin are delimited by a null byte [DONT USE]"`
	Extract bool   `short:"i" long:"extract" description:"Extract files from an archive"`
	Create  bool   `short:"o" long:"create" description:"Create an archive"`
	List    bool   `short:"t" long:"list" description:"Print a table of contents of the input"`
	Verbose bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func extract(archiver cpio.RecordFormat) error {
	var inums map[uint64]string
	inums = make(map[uint64]string)
	rr, err := archiver.NewFileReader(os.Stdin)
	if err != nil {
		return err
	}
	for {
		rec, err := rr.ReadRecord()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading record: %v", err)
		}

		Debug("%v: ino %s\n", rec.Name, rec.Ino)
		// record with ino 0 could be a hardlink
		if rec.Ino != 0 {
			switch rec.Mode & cpio.S_IFMT {
			case cpio.S_IFDIR:
			default:
				if rec.FileSize != 0 {
					break
				}
				ino, ok := inums[rec.Ino]
				if !ok {
					break
				}

				Debug("hard linking %v to %v\n", ino, rec.Name)
				err := os.Link(ino, rec.Name)
				if err != nil {
					return err
				}
				continue
			}
			inums[rec.Ino] = rec.Name
		}
		Debug("Creating: %v\n", rec.Name)
		if err := cpio.CreateFile(rec); err != nil {
			log.Printf("Creating %q failed: %v", rec.Name, err)
		}
	}
	return nil
}

func create(archiver cpio.RecordFormat) error {
	rw := archiver.Writer(os.Stdout)
	cr := cpio.NewRecorder()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		name := scanner.Text()
		rec, err := cr.GetRecord(name)
		if err != nil {
			return fmt.Errorf("Getting record of %q failed: %w", name, err)
		}

		if err := rw.WriteRecord(rec); err != nil {
			return fmt.Errorf("Writing record of %q failed: %w", name, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from stdin: %v", err)
	}
	if err := cpio.WriteTrailer(rw); err != nil {
		return fmt.Errorf("Error writing trailer record: %w", err)
	}
	return nil
}

func list(archiver cpio.RecordFormat) error {
	rr, err := archiver.NewFileReader(os.Stdin)
	if err != nil {
		return err
	}
	for {
		rec, err := rr.ReadRecord()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading records: %w", err)
		}
		fmt.Fprintln(os.Stdout, rec)
	}
	return nil
}

func Cpio(args []string) error {
	archiver, err := cpio.Format(opts.Format)
	if err != nil {
		return fmt.Errorf("Format %v not supported: %v", opts.Format, err)
	}

	var counter []bool
	counter = append(counter, opts.Create, opts.Extract, opts.List)
	var i int8
	for _, b := range counter {
		if b {
			i++
		}
	}

	if i != 1 {
		return fmt.Errorf("Cannot specify multiple operations at once")
	}

	if opts.Extract {
		return extract(archiver)
	}
	if opts.Create {
		return create(archiver)
	}
	if opts.List {
		return list(archiver)
	}
	return fmt.Errorf("Invalid arguments")
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

	if err := Cpio(args); err != nil {
		log.Fatal(err)
	}
}
