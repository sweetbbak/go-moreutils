package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Recursive bool   `short:"R" long:"recursive" description:"recursively change mode of files [use caution]"`
	Reference string `short:"r" long:"reference" description:"copy the file mode from an existing reference file"`
	Debugging bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

const (
	special = 99999
)

func changeMode(path string, mode os.FileMode, octval uint64, mask uint64) (fs.FileMode, error) {
	if mask == special {
		if err := os.Chmod(path, mode); err != nil {
			return 0, err
		}
		return mode, nil
	}

	var info os.FileInfo
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	mode = info.Mode() & os.FileMode(mask)
	mode = mode | os.FileMode(octval)
	Debug("Changing mode: [%v] [%v]\n", path, mode)

	if err := os.Chmod(path, mode); err != nil {
		return 0, err
	}
	return mode, nil
}

func setuid(path string, mode os.FileMode, octval uint64, mask uint64) {
	// Get current file permissions
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	// Add setuid and sticky bit to the file mode
	newMode := fileInfo.Mode() | os.ModeSetuid | os.ModeSticky
	// Set the new file mode
	err = os.Chmod(path, newMode)
	if err != nil {
		log.Fatal(err)
	}
}

func calculateMode(modestr string) (mode os.FileMode, octval uint64, mask uint64, err error) {
	octval, err = strconv.ParseUint(modestr, 8, 32)
	if err == nil {
		// if octval > 0o777 {
		if octval > 4777 {
			return mode, octval, mask, fmt.Errorf("%w: invalid octal value %0o. Value should be less than or equal to 0777", strconv.ErrRange, octval)
		}
		mask = special
		mode = os.FileMode(octval)

		if octval == 4777 {
			mode = mode | os.ModeSetuid | os.ModeSticky
		}
		return
	}

	reMode := regexp.MustCompile("^([ugoa]+)([-+=](.*))")
	m := reMode.FindStringSubmatch(modestr)
	reMode = regexp.MustCompile("^[rwx]*$")
	if len(m) < 3 || !reMode.MatchString(m[3]) {
		return mode, octval, mask, fmt.Errorf("%w: unable to decode mode %q. Please use an octal value.", strconv.ErrRange, modestr)
	}

	var octvalDigit uint64
	if strings.Contains(m[3], "r") {
		octvalDigit += 4
	}
	if strings.Contains(m[3], "w") {
		octvalDigit += 2
	}
	if strings.Contains(m[3], "x") {
		octvalDigit++
	}
	if strings.Contains(m[3], "s") {
	}

	operator := m[2]
	mask = 0o777
	if operator == "-" {
		octvalDigit = 7 - octvalDigit
	}
	if strings.Contains(m[1], "o") || strings.Contains(m[1], "a") {
		octval += octvalDigit
		mask = mask & 0o770
	}
	if strings.Contains(m[1], "g") || strings.Contains(m[1], "a") {
		octval += octvalDigit << 3
		mask = mask & 0o707
	}
	if strings.Contains(m[1], "u") || strings.Contains(m[1], "a") {
		octval += octvalDigit << 6
		mask = mask & 0o077
	}
	if operator == "+" {
		mask = 0o777
	}

	if operator == "=" && strings.Contains(m[1], "a") {
		mask = special
		mode = os.FileMode(octval)
	}
	return mode, octval, mask, nil
}

func chmod(args []string) (os.FileMode, error) {
	var mode os.FileMode
	if len(args) < 1 {
		return mode, fmt.Errorf("Must provide mode and file [chmod +x my-file]")
	}

	if len(args) < 2 && opts.Reference == "" {
		return mode, fmt.Errorf("Must provide mode and file [chmod +x my-file]")
	}

	var (
		err          error
		octval, mask uint64
		filelist     []string
	)

	if opts.Reference != "" {
		fi, err := os.Stat(opts.Reference)
		if err != nil {
			return 0, fmt.Errorf("bad reference file: %w", err)
		}
		mask = special
		mode = fi.Mode()
		filelist = args
		Debug("Reference file mask && mode: [%v] [%v]\n", mask, mode)
		Debug("File list: [%v]\n", filelist)
	} else {
		var err error
		if mode, octval, mask, err = calculateMode(args[0]); err != nil {
			return mode, err
		}
		filelist = args[1:]
		Debug("File list: [%v]\n", filelist)
	}

	for _, name := range filelist {
		if opts.Recursive {
			err := filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
				mode, err = changeMode(path, mode, octval, mask)
				Debug("Recursive chmod: [%v] mode [%v]\n", path, mode)
				return err
			})
			if err != nil {
				return mode, err
			}
		} else {
			if mode, err = changeMode(name, mode, octval, mask); err != nil {
				return mode, err
			}
		}
	}
	return mode, err
}

func init() {
	if len(os.Args) < 3 {
		os.Exit(1)
	} else {
		for i, arg := range os.Args {
			switch arg {
			case "-r", "--reference":
			case "-R", "--recursive":
			case "-v", "--verbose":
			default:
				// this will allow us to pass [chmod +x file] vs [chmod a+x file] which is implied in the OG chmod
				// this 'a' is implied by default
				if strings.HasPrefix(arg, "+") || strings.HasPrefix(arg, "-") {
					arg = fmt.Sprintf("a%v", arg)
					os.Args[i] = arg
				}
			}
		}
	}
}

func main() {
	Defaults := flags.IgnoreUnknown | flags.HelpFlag | flags.PassDoubleDash
	parser := flags.NewParser(&opts, flags.Options(Defaults))
	args, err := parser.Parse()
	if err != nil {
		if err == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println(err)
		}
	}

	if opts.Debugging {
		Debug = log.Printf
	}
	Debug("arguments: %v\n", args)

	if _, err := chmod(args); err != nil {
		log.Fatal(err)
	}
}
