package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Filter func(hdr *tar.Header) bool

type TarOpts struct {
	Filters   []Filter
	NoRecurse bool
	ChangeDir string
}

func passesFilter(hdr *tar.Header, filters []Filter) bool {
	for _, filter := range filters {
		if !filter(hdr) {
			return false
		}
	}
	return true
}

func applyToArchive(tarFile io.Reader, f func(tr *tar.Reader, hdr *tar.Header) error) error {
	tr := tar.NewReader(tarFile)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if err := f(tr, hdr); err != nil {
			return err
		}
	}
	return nil
}

func listArchive(tarfile io.Reader) error {
	return applyToArchive(tarfile, func(tr *tar.Reader, hdr *tar.Header) error {
		fmt.Println(hdr.Name)
		return nil
	})
}

func SafeFilepathJoin(path1, path2 string) (string, error) {
	relPath, err := filepath.Rel(".", path2)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("(zipslip) filepath is unsafe %q: %w", path2, err)
	}
	if path1 == "" {
		path1 = "."
	}
	return filepath.Join(path1, filepath.Join(string(filepath.Separator), relPath)), nil
}

func extractDir(tarfile io.Reader, dir string, opts *TarOpts) error {
	if opts == nil {
		opts = &TarOpts{}
	}

	if !filepath.IsAbs(dir) {
		dir = filepath.Join(opts.ChangeDir, dir)
	}

	fi, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			return fmt.Errorf("could not create directory %s: %w", dir, err)
		}
	} else if err != nil || !fi.IsDir() {
		return fmt.Errorf("could not stat directory %s: %w", dir, err)
	}

	return applyToArchive(tarfile, func(tr *tar.Reader, hdr *tar.Header) error {
		if !passesFilter(hdr, opts.Filters) {
			return nil
		}
		return createFileInRoot(hdr, tr, dir)
	})
}

func createFileInRoot(hdr *tar.Header, r io.Reader, rootDir string) error {
	fi := hdr.FileInfo()
	path, err := SafeFilepathJoin(rootDir, hdr.Name)
	if err != nil {
		log.Printf("Warning: skipping file %v due to: %v", hdr.Name, err)
		return nil
	}

	switch fi.Mode() & os.ModeType {
	case os.ModeSymlink:
		return fmt.Errorf("symlinks not supported yet")
	case os.FileMode(0):
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, r); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	case os.ModeDir:
		if err := os.MkdirAll(path, fi.Mode()&os.ModePerm); err != nil {
			return err
		}
	case os.ModeDevice:
		return fmt.Errorf("block device not supported yet: %v", path)
	case os.ModeCharDevice:
		return fmt.Errorf("char device not supported yet: %v", path)
	default:
		return fmt.Errorf("unkown file type: %v: %v", path, fi.Mode()&os.ModePerm)
	}

	if err := os.Chmod(path, fi.Mode()&os.ModePerm); err != nil {
		return fmt.Errorf("error setting mode %#o on %v: %w", fi.Mode()&os.ModePerm, path, err)
	}
	// add ownership
	return nil
}

func NoFilter(hdr *tar.Header) bool {
	return true
}

// VerboseFilter prints the name of every file.
func VerboseFilter(hdr *tar.Header) bool {
	fmt.Println(hdr.Name)
	return true
}

// VerboseLogFilter logs the name of every file.
func VerboseLogFilter(hdr *tar.Header) bool {
	log.Println(hdr.Name)
	return true
}

// SafeFilter filters out all files which are not regular and not directories.
// It also sets appropriate permissions.
func SafeFilter(hdr *tar.Header) bool {
	if hdr.Typeflag == tar.TypeDir {
		hdr.Mode = 0o770
		return true
	}
	if hdr.Typeflag == tar.TypeReg {
		hdr.Mode = 0o660
		return true
	}
	return false
}

func CreateTar(tarFile io.Writer, files []string, opts *TarOpts) error {
	if opts == nil {
		opts = &TarOpts{}
	}

	tw := tar.NewWriter(tarFile)
	for _, bFile := range files {
		// Simulate a "cd" to another directory. There are 3 parts to
		// the file path:
		// a) The path passed to ChangeDirectory
		// b) The path passed in files
		// c) The path in the current walk
		// I prefixed corresponding a/b/c onto the variable name as an
		// aid. For example abFile is the filepath of a+b.
		abFile := filepath.Join(opts.ChangeDir, bFile)
		if filepath.IsAbs(bFile) {
			// "cd" does nothing if the file is absolute.
			abFile = bFile
		}

		walk := filepath.Walk
		if opts.NoRecurse {
			// This "walk" function does not recurse.
			walk = func(root string, walkFn filepath.WalkFunc) error {
				fi, err := os.Lstat(root)
				return walkFn(root, fi, err)
			}
		}

		err := walk(abFile, func(abcPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// The record should not contain the ChangeDirectory
			// path, so we need to derive bc from abc.
			bcPath, err := filepath.Rel(opts.ChangeDir, abcPath)
			if err != nil {
				return err
			}
			if filepath.IsAbs(bFile) {
				// "cd" does nothing if the file is absolute.
				bcPath = abcPath
			}

			var symlink string
			if info.Mode()&os.ModeSymlink == os.ModeSymlink {
				if symlink, err = os.Readlink(abcPath); err != nil {
					return err
				}
			}
			hdr, err := tar.FileInfoHeader(info, symlink)
			if err != nil {
				return err
			}
			hdr.Name = bcPath
			if !passesFilter(hdr, opts.Filters) {
				return nil
			}
			switch hdr.Typeflag {
			case tar.TypeLink, tar.TypeSymlink, tar.TypeChar, tar.TypeBlock, tar.TypeDir, tar.TypeFifo:
				if err := tw.WriteHeader(hdr); err != nil {
					return err
				}
			default:
				f, err := os.Open(abcPath)
				if err != nil {
					return err
				}

				var r io.Reader = f
				if hdr.Size == 0 {
					// Some files don't report their size correctly
					// (ex: procfs), so we use an intermediary
					// buffer to determine size.
					b := &bytes.Buffer{}
					if _, err := io.Copy(b, f); err != nil {
						f.Close()
						return err
					}
					f.Close()
					hdr.Size = int64(b.Len())
					r = b
				}

				if err := tw.WriteHeader(hdr); err != nil {
					return err
				}
				if _, err := io.Copy(tw, r); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return nil
}
