package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/jessevdk/go-flags"
)

const (
	// BytesRequiredForGPTPartitionEntries is the total bytes required to store the GPT partition
	// entries. 128 bytes are required per partition, total no of partition supported by GPT = 128
	// Therefore, total bytes = 128*128
	BytesRequiredForGPTPartitionEntries = 16384

	// GPTPartitionStartByte is the byte on the disk at which the first partition starts.
	// Normally partition starts at 1MiB, (as done by fdisk utility). This is done to
	// align the partition start to physical block sizes on the disk.
	GPTPartitionStartByte = 1048576

	// NoOfLogicalBlocksForGPTHeader is the no. of logical blocks for the GPT header.
	NoOfLogicalBlocksForGPTHeader = 1
)

var opts struct {
	Label   string `short:"l" long:"label" description:"filesystem label"`
	FsType  string `short:"f" long:"filesystem" description:"filesystem type [fat32|squashfs|iso]"`
	Verbose bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Mkfs(args []string) error {
	d, err := diskfs.Open(args[0])
	if err != nil {
		return fmt.Errorf("error opening disk %s: %v", args[0], err)
	}

	fspec := disk.FilesystemSpec{Partition: 0, VolumeLabel: opts.Label}

	switch strings.ToLower(opts.FsType) {
	case "fat", "fat32", "vfat":
		fspec.FSType = filesystem.TypeFat32
	case "iso", "iso9660":
		fspec.FSType = filesystem.TypeISO9660
	case "squash", "squashfs":
		fspec.FSType = filesystem.TypeSquashfs
	default:
		return fmt.Errorf("unsupported file system type: %v", opts.FsType)
	}

	if _, err := d.CreateFilesystem(fspec); err != nil {
		return fmt.Errorf("error creating filesystem: %w", err)
	}

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

	if err := Mkfs(args); err != nil {
		log.Fatal(err)
	}
}
