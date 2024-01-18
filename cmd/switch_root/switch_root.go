package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	"golang.org/x/sys/unix"
)

var opt struct {
	NewRoot string `short:"r" long:"root" description:"new root to change into"`
	Init    string `short:"r" long:"root" description:"init system to run"`
}

// SameFilesystem returns true if both paths reside in the same filesystem.
// This is achieved by comparing Stat_t.Dev, which contains the fs device's
// major/minor numbers.
func SameFilesystem(path1, path2 string) (bool, error) {
	var stat1, stat2 unix.Stat_t
	if err := unix.Stat(path1, &stat1); err != nil {
		return false, err
	}
	if err := unix.Stat(path2, &stat2); err != nil {
		return false, err
	}
	return stat1.Dev == stat2.Dev, nil
}

// addSpecialMounts moves the 'special' mounts to the given target path
//
// 'special' in this context refers to the following non-blockdevice backed
// mounts that are almost always used: /dev, /proc, /sys, and /run.
// This function will create the target directories, if necessary.
// If the target directories already exist, they must be empty.
// This function skips missing mounts.
func addSpecialMounts(newRoot string) error {
	mounts := []string{"/dev", "/proc", "/sys", "/run"}

	for _, mount := range mounts {
		path := filepath.Join(newRoot, mount)
		// Skip all mounting if the directory does not exist.
		if _, err := os.Stat(mount); os.IsNotExist(err) {
			log.Printf("switch_root: Skipping %q as the dir does not exist", mount)
			continue
		} else if err != nil {
			return err
		}
		// Also skip if not currently a mount point
		if same, err := SameFilesystem("/", mount); err != nil {
			return err
		} else if same {
			log.Printf("switch_root: Skipping %q as it is not a mount", mount)
			continue
		}
		// Make sure the target dir exists.
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
		if err := MoveMount(mount, path); err != nil {
			return err
		}
	}
	return nil
}

func MoveMount(oldPath string, newPath string) error {
	return unix.Mount(oldPath, newPath, "", unix.MS_MOVE, "")
}

// newRoot is the "first half" of SwitchRoot - that is, it creates special mounts
// in newRoot, chroot's there, and RECURSIVELY DELETES everything in the old root.
func newRoot(newRootDir string) error {
	log.Printf("switch_root: moving mounts")
	if err := addSpecialMounts(newRootDir); err != nil {
		return fmt.Errorf("switch_root: moving mounts failed %v", err)
	}

	log.Printf("switch_root: Changing directory")
	if err := unix.Chdir(newRootDir); err != nil {
		return fmt.Errorf("switch_root: failed change directory to new_root %v", err)
	}

	// Open "/" now, we need the file descriptor later.
	oldRoot, err := os.Open("/")
	if err != nil {
		return err
	}
	defer oldRoot.Close()

	log.Printf("switch_root: Moving /")
	if err := MoveMount(newRootDir, "/"); err != nil {
		return err
	}

	log.Printf("switch_root: Changing root!")
	if err := unix.Chroot("."); err != nil {
		return fmt.Errorf("switch_root: fatal chroot error %v", err)
	}

	log.Printf("switch_root: Deleting old /")
	return recursiveDelete(int(oldRoot.Fd()))
}

func SwitchRoot(newRootDir string, init string) error {
	err := newRoot(newRootDir)
	if err != nil {
		return err
	}
	return execInit(init)
}

// execInit is generally only useful as part of SwitchRoot or similar.
// It exec's the given binary in place of the current binary, necessary so that
// the new binary can be pid 1.
func execInit(init string) error {
	log.Printf("switch_root: executing init")
	if err := unix.Exec(init, []string{init}, []string{}); err != nil {
		return fmt.Errorf("switch_root: exec failed %v", err)
	}
	return nil
}

func recursiveDelete(fd int) error {
	parentDev, err := getDev(fd)
	if err != nil {
		log.Printf("warn: unable to get underlying dev for dir: %v", err)
		return nil
	}

	// The file descriptor is already open, but allocating a os.File
	// here makes reading the files in the dir so much nicer.
	dir := os.NewFile(uintptr(fd), "__ignored__")
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		log.Printf("warn: unable to read dir %s: %v", dir.Name(), err)
		return nil
	}

	for _, name := range names {
		// Loop here, but handle loop in separate function to make defer work as expected.
		if err := recusiveDeleteInner(fd, parentDev, name); err != nil {
			return err
		}
	}
	return nil
}
func recusiveDeleteInner(parentFd int, parentDev uint64, childName string) error {
	// O_DIRECTORY and O_NOFOLLOW make this open fail for all files and all symlinks (even when pointing to a dir).
	// We need to filter out symlinks because getDev later follows them.
	childFd, err := unix.Openat(parentFd, childName, unix.O_DIRECTORY|unix.O_NOFOLLOW, unix.O_RDWR)
	if err != nil {
		// childName points to either a file or a symlink, delete in any case.
		if err := unix.Unlinkat(parentFd, childName, 0); err != nil {
			log.Printf("warn: unable to remove file %s: %v", childName, err)
		}
	} else {
		// Open succeeded, which means childName points to a real directory.
		defer unix.Close(childFd)

		// Don't descend into other file systems.
		if childFdDev, err := getDev(childFd); err != nil {
			log.Printf("warn: unable to get underlying dev for dir: %s: %v", childName, err)
			return nil
		} else if childFdDev != parentDev {
			// This means continue in recursiveDelete.
			return nil
		}

		if err := recursiveDelete(childFd); err != nil {
			return err
		}
		// Back from recursion, the directory is now empty, delete.
		if err := unix.Unlinkat(parentFd, childName, unix.AT_REMOVEDIR); err != nil {
			log.Printf("warn: unable to remove dir %s: %v", childName, err)
		}
	}
	return nil
}

// getDev returns the device (as returned by the FSTAT syscall) for the given file descriptor.
func getDev(fd int) (dev uint64, err error) {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return 0, err
	}
	return uint64(stat.Dev), nil
}

func main() {
	args, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(0)
	}

	var newRoot string
	if opt.NewRoot != "" {
		newRoot = args[0]
	} else {
		newRoot = opt.NewRoot
	}
	var init string
	if opt.Init != "" {
		init = args[1]
	} else {
		init = opt.Init
	}
	if err := SwitchRoot(newRoot, init); err != nil {
		log.Fatalf("switch_root failed %v\n", err)
	}

}
