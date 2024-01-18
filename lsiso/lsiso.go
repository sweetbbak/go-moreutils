package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
)

func PrintIsoInfo(isoPath string) {
	disk, err := diskfs.Open(isoPath)
	if err != nil {
		log.Fatal(err)
	}
	fs, err := disk.GetFilesystem(0)
	if err != nil {
		log.Fatal(err)
	}

	err = fileInfoFor("/", fs)
	if err != nil {
		log.Fatalf("Failed to get file info: %s\n", err)
	}
}

func hlmode(s string) string {
	var str string
	str = strings.ReplaceAll(s, "r", "\x1b[33mr\x1b[0m")
	str = strings.ReplaceAll(str, "w", "\x1b[31mw\x1b[0m")
	str = strings.ReplaceAll(str, "x", "\x1b[4m\x1b[32mx\x1b[0m")
	str = strings.ReplaceAll(str, "-", "\x1b[38;2;111;111;114m-\x1b[0m")
	return str
}

func fileInfoFor(path string, fs filesystem.FileSystem) error {
	files, err := fs.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		fullPath := filepath.Join(path, file.Name())
		if file.IsDir() {
			err = fileInfoFor(fullPath, fs)
			if err != nil {
				return err
			}
			continue
		}
		isoFile, err := fs.OpenFile(fullPath, os.O_RDONLY)
		if err != nil {
			fmt.Printf("Failed to open file %s: %v\n", fullPath, err)
			continue
		}

		myFile := isoFile.(*iso9660.File)
		m := fmt.Sprintf("%s", myFile.Mode())
		mode := hlmode(m)
		str := fmt.Sprintf("%v %-11v %-11v %-5v", mode, file.Size(), myFile.Location(), fullPath)
		str = strings.ToLower(str)
		fmt.Println(str)
		// fmt.Printf("%s\n Size: %d\n Location: %d\n\n", fullPath, file.Size(), myFile.Location())
	}
	return nil
}

func printHelp() {
	fmt.Println("usage: lsiso <path-to-isos>")
	fmt.Println("\tsupports listing more than one iso at a time")
}

func main() {
	if len(os.Args) <= 1 {
		printHelp()
		os.Exit(0)
	}

	for _, file := range os.Args[1:] {
		switch file {
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		default:
			PrintIsoInfo(file)
		}
	}
}
