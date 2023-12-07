package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
)

var opt struct {
	Force           bool   `short:"f" long:"force" description:"ignore nonexistent files and arguments, never prompt"`
	Update          bool   `short:"u" long:"update" description:"move only when the SOURCE file is newer than the DESTINATION file"`
	Interactive     bool   `short:"i" long:"interactive" description:"Ask before removing a file"`
	InteractiveOnce bool   `short:"I" description:"prompt once before removing multiple files, or when recursively removing files. Less intrusive than -i"`
	NoClobber       bool   `short:"n" long:"no-clobber" description:"do not overwrite existing files"`
	Destination     string `short:"d" long:"destination" description:"Specify file destination"`
	Verbose         bool   `short:"v" description:"explain what is being done"`
}

var dest string

func moveFile(source string, dest string) error {
	if opt.NoClobber {
		_, err := os.Lstat(dest)
		if !os.IsNotExist(err) {
			// this is either a real error or something happened with lstat
			return err
		}
	}

	if opt.Update {
		sourceInfo, err := os.Lstat(source)
		if err != nil {
			return err
		}

		destInfo, err := os.Lstat(dest)
		if err != nil {
			return err
		}

		// check if dest exists and is newer than source file
		if destInfo.ModTime().After(sourceInfo.ModTime()) {
			// source is older than dest file
			return nil
		}
	}

	if err := os.Rename(source, dest); err != nil {
		return err
	}
	return nil
}

func mv(files []string, todir bool, dest string) error {
	// rename or move a file
	if len(files) == 2 && !todir {
		if err := moveFile(files[0], files[1]); err != nil {
			return err
		}
	} else {
		// if dest is set not set explicitly we assume the last argument is the dest and exclude it from mv
		if opt.Destination == "" {
			files = files[:len(files)-1]
		}
		// destdir := files[len(files)-1]
		for _, f := range files {
			newPath := filepath.Join(dest, filepath.Base(f))
			if err := moveFile(f, newPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func move(files []string) error {
	var todir bool
	// explicit set destination or infer last dir/file as dest
	if opt.Destination != "" {
		dest = opt.Destination
	} else {
		dest = files[len(files)-1]
	}

	// stat that last file to check if is dir, if so, thats our destination
	if destdir, err := os.Lstat(dest); err == nil {
		todir = destdir.IsDir()
	}

	if len(files) > 2 && !todir {
		return fmt.Errorf("not a directory: %s", dest)
	}
	return mv(files, todir, dest)
}

func main() {
	args, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(0)
	}

	if len(args) < 2 {
		os.Exit(1)
	}

	if err := move(args); err != nil {
		log.Fatal(err)
	}

}
