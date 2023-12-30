package main

import (
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/araddon/dateparse"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	NoCreate  bool   `short:"c" long:"no-create" description:"do not create any new files, just touch existing files"`
	Reference string `short:"r" long:"reference" description:"use a reference file to copy (access or modtime) from"`
	Time      string `short:"T" long:"time" description:"use a time duration to set a time N (h|m|s|ns) ago ex: (2h45m)"`
	TStamp    string `short:"t" long:"timestamp" description:"use [[CC]YY]MMDDhhmm[.ss]"`
	Date      string `short:"d" long:"date" description:"set file time to an exact date time (October 31 2001)"`
	Access    bool   `short:"a" long:"access" description:"change files access time"`
	ModTime   bool   `short:"m" long:"modtime" description:"change files modtime"`
	Verbose   bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func reference(file string) (time.Time, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return time.Now(), err
	}
	return fi.ModTime(), nil
}

func parseDate(d string) (time.Time, error) {
	now := time.Now()
	t, err := dateparse.ParseAny(d)
	if err != nil {
		return now, err
	}
	return t, nil
}

func parseTime(d string, now time.Time) (time.Time, error) {
	td, err := time.ParseDuration(d)
	if err != nil {
		return now, err
	}
	return now.Add(-td), nil
}

func touch(file string, now time.Time, fi fs.FileInfo, exists bool) error {
	var modNow time.Time
	if exists {
		modNow = fi.ModTime()
	} else {
		modNow = now
	}

	if !exists && !opts.NoCreate {
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()

	}

	// counter-intuitive - ive made it so now is either the actual creation time, or any
	// number of modified times, so by defualt we need to change the file time so hence the
	// double negative bools here. Otherwise we do one or the other. Refactor later.
	if !opts.Access && !opts.ModTime {
		err := os.Chtimes(file, now, now)
		if err != nil {
			return err
		}
		return nil
	}

	if opts.Access {
		err := os.Chtimes(file, now, modNow)
		if err != nil {
			return err
		}
	}

	if opts.ModTime {
		err := os.Chtimes(file, modNow, now)
		if err != nil {
			return err
		}
	}
	return nil
}

func Touch(args []string) error {
	var now time.Time
	var err error
	now = time.Now()

	if opts.Reference != "" {
		now, err = reference(opts.Reference)
	}

	if opts.Date != "" {
		now, err = parseDate(opts.Date)
		if err != nil {
			return err
		}
	}

	Debug("touch: Base time: %v\n", now.String())

	if opts.Time != "" {
		now, err = parseTime(opts.Time, now)
		if err != nil {
			return err
		}
		Debug("touch: Base time after modification: %v\n", now.String())
	}

	for _, file := range args {
		path := os.ExpandEnv(file)
		Debug("touch: %v: time: %v\n", path, now.String())
		fi, err := os.Stat(file)
		exists := (err == nil)

		err = touch(file, now, fi, exists)
		if err != nil {
			return err
		}
	}
	return nil
}

func pathExists(file string) bool {
	_, err := os.Stat(file)
	return (err == nil)
}

func init() {
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Touch(args); err != nil {
		log.Fatal(err)
	}
}
