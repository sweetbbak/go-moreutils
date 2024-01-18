package main

import (
	archiver "github.com/mholt/archiver/v3"
)

func CompressTo7z(filename string, filesToCompress []string) error {
	err := archiver.Archive(filesToCompress, filename)
	if err != nil {
		return err
	}

	return nil
}

func Extract7z(filename, destFoldername string) error {
	err := archiver.Unarchive(filename, destFoldername)
	if err != nil {
		return err
	}

	return nil
}
