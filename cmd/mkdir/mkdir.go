package main

import (
	"fmt"
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"
)

var opt struct {
	Parents bool   `short:"p" long:"parents" description:"make parent directories if they dont exist"`
	Mode    uint32 `short:"m" long:"mode" description:"set file mode (as in chmod), not a=rwx - umask"`
	Verbose bool   `short:"v" long:"verbose" description:"print a message for each created directory"`
}

func Mkdirs(dirs []string) ([]string, error) {
	var (
		dir  string
		mode os.FileMode
		err  error
	)

	mode = os.FileMode(uint32(opt.Mode))
	// defaults to 0 so set it to something sane
	if mode.Perm() == 0 {
		mode = 0777
	}

	for _, dir = range dirs {
		if opt.Parents {
			if err = os.MkdirAll(dir, mode); err != nil {
				return dirs, err
			}
		} else if err = os.Mkdir(dir, mode); err != nil {
			return dirs, err
		}

		if opt.Verbose {
			_, err = fmt.Printf("Created %s\t%v\n", dir, mode)
			if err != nil {
				return dirs, err
			}
		}
	}

	return dirs, nil
}

func main() {
	args, err := flags.Parse(&opt)
	if err != nil {
		log.Fatal(err)
	}

	Mkdirs(args)
}
