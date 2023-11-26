package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	flags "github.com/jessevdk/go-flags"
)

var opt struct {
	Recursive       bool `short:"r" long:"recursive" description:"Recursively remove a file system node"`
	Force           bool `short:"f" long:"force" description:"ignore nonexistent files and arguments, never prompt"`
	Interactive     bool `short:"i" long:"interactive" description:"Ask before removing a file"`
	InteractiveOnce bool `short:"I" description:"prompt once before removing multiple files, or when recursively removing files. Less intrusive than -i"`
	Directories     bool `short:"d" long:"dir" description:"remove empty directories"`
	OneFS           bool `long:"one-file-system" description:"When removing a hierarchy recursively, skip any directory that is on a different file system"`
	NoPreserveRoot  bool `long:"no-preserve-root" description:"dont treat '/' specially - Dangerous"`
	Verbose         bool `short:"v" description:"explain what is being done"`
}

// Confirm file deletion, True is confirmed. Uses Scanln and prints to stderr
// TODO fprintf to stdout or stderr based on whether stdout is a terminal
func Confirm(file string) (bool, error) {
	var answer string
	fmt.Fprintf(os.Stderr, "Remove this file '%s': [y/N] ", file)
	_, err := fmt.Scanln(&answer)
	if err != nil {
		return false, errors.New("Not confirmed")
	}

	if answer == "y" || answer == "yes" || answer == "Y" {
		// confirmed with no errors
		return true, nil
	} else {
		return false, errors.New("Not confirmed")
	}
}

// remove a normal file - ask for confirmation if necessary
func RemoveFile(file string) error {
	if opt.Interactive {
		_, err := Confirm(file)
		if err != nil {
			return err
		}
	}

	if opt.Force {
		opt.Interactive = false
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	toRemove := file
	if !path.IsAbs(file) {
		toRemove = filepath.Join(cwd, file)
	}

	fmt.Println("Removed: ", toRemove)
	// remove func here
	//
	return nil
}

func RemoveDir(dir string) error {
	// -r must be specified to rm a directory
	if !opt.Recursive {
		return errors.New("Is a directory")
	}

	if opt.Force {
		opt.Interactive = false
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// should be false if -I is specified
	if opt.Interactive {
		_, err := Confirm(dir)
		if err != nil {
			return err
		}
	}

	toRemove := dir
	if !path.IsAbs(dir) {
		toRemove = filepath.Join(cwd, dir)
	}

	fmt.Println("Removed: ", toRemove)
	// remove func here
	//

	return nil
}

// generic confirmation - used to confirm once for arg "-I"
// removes the need to confirm rm for every single file
func ConfrimOnce(prompt string) (bool, error) {
	fmt.Fprintf(os.Stdout, prompt)
	input := bufio.NewReader(os.Stdin)
	answer, err := input.ReadString('\n')

	// if answer string first character isnt equal to "y" then no confirmation provided
	if err != nil || strings.ToLower(answer)[0] != 'y' {
		return false, errors.New("No Confirmation")
	} else {
		return true, nil
	}
}

// main function for removing a slice of files
func RemoveFiles(files []string) error {
	if len(files) < 1 {
		return fmt.Errorf("%v", flags.ErrHelp)
	}

	// Confirm remove ALL files once instead of individually
	if opt.InteractiveOnce {
		prompt := fmt.Sprintf("Remove %d files? [y/N] ", len(files))
		conf, err := ConfrimOnce(prompt)
		if err != nil {
			return err
		} else if conf {
			opt.Interactive = false
		}
	}

	for _, file := range files {
		// stat the files to determine file type and handle path/not-exists errors
		fi, err := os.Lstat(file)
		if err != nil {
			if err, ok := err.(*os.PathError); ok && (os.IsNotExist(err.Err) || err.Err == syscall.ENOTDIR) {
				continue
			}
			continue
		}

		// if item is a file we can just remove it and move on
		if !fi.IsDir() {
			err := RemoveFile(file)
			if err != nil {
				fmt.Printf("Error removing: %v\n", err)
			}
		}

		// check if -r is specified and handle other flags for recursive deletion
		if fi.IsDir() {
			err := RemoveDir(file)
			if err != nil {
				fmt.Printf("cannot remove %s: %v\n", file, err)
			}
		}
	}
	return nil
}

func main() {
	// args will be our files to be removed
	args, err := flags.Parse(&opt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(&opt)

	if opt.Verbose {
		fmt.Println("Removing: ")
		for i := range args {
			fmt.Println(args[i])
		}
	}

	if err := RemoveFiles(args); err != nil {
		log.Fatal(err)
	}
}
