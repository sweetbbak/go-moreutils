package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"mybox/pkg/mount/block"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Blkid(args []string, out io.ReadWriter) error {
	dev, err := block.GetBlockDevices()
	if err != nil {
		return err
	}

	for _, device := range dev {
		fmt.Print(device.DevicePath())
		if device.FsUUID != "" {
			fmt.Fprintf(out, ` UUID="%s"`, device.FsUUID)
		}
		if device.FSType != "" {
			fmt.Fprintf(out, ` TYPE="%s"`, device.FSType)
		}
		fmt.Println()
	}
	return nil
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

	if err := Blkid(args, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
