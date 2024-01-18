package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/squashfs"
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

func makeem(ff filesystem.Type, args []string) error {
	fspec := disk.FilesystemSpec{Partition: 0, VolumeLabel: opts.Label}
	fspec.FSType = ff

	d, err := diskfs.Open(args[0])
	if err != nil {
		return fmt.Errorf("error opening disk %s: %v", args[0], err)
	}

	fspec.FSType = filesystem.TypeFat32
	if _, err := d.CreateFilesystem(fspec); err != nil {
		return fmt.Errorf("error creating filesystem: %w", err)
	}
	return nil
}

func Mkfs(args []string) error {
	switch strings.ToLower(opts.FsType) {

	case "fat", "fat32", "vfat":
		makeem(filesystem.TypeFat32, args)

	case "iso", "iso9660":
		makeem(filesystem.TypeISO9660, args)

	case "squash", "squashfs":
		return CreateSquashfs(args[0], args[1:])

	default:
		return fmt.Errorf("unsupported file system type: %v", opts.FsType)
	}
	return nil
}

func CreateSquashfs(diskImg string, contents []string) error {
	if diskImg == "" {
		log.Fatal("must have a valid path for diskImg")
	}
	var diskSize int64 = 10 * 1024 * 1024 // 10 MB
	mydisk, err := diskfs.Create(diskImg, diskSize, diskfs.Raw, diskfs.SectorSizeDefault)
	if err != nil {
		return err
	}

	fspec := disk.FilesystemSpec{Partition: 0, FSType: filesystem.TypeSquashfs, VolumeLabel: opts.Label}
	fs, err := mydisk.CreateFilesystem(fspec)
	if err != nil {
		return err
	}

	for _, spfile := range contents {
		content, err := os.ReadFile(spfile)
		if err != nil {
			fmt.Println(err)
		}

		rw, err := fs.OpenFile(spfile, os.O_CREATE|os.O_RDWR)
		_, err = rw.Write(content)
		if err != nil {
			return err
		}
	}

	// rw, err := fs.OpenFile("demo.txt", os.O_CREATE|os.O_RDWR)
	// content := []byte("demo")
	// _, err = rw.Write(content)
	// if err != nil {
	// 	return err
	// }

	sqs, ok := fs.(*squashfs.FileSystem)
	if !ok {
		if err != nil {
			return fmt.Errorf("not a squashfs filesystem")
		}
	}

	err = sqs.Finalize(squashfs.FinalizeOptions{})
	if err != nil {
		return err
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
