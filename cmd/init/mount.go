package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

type mount struct {
	source  string
	target  string
	fstype  string
	options string
}

type MountOption struct {
	Source string
	Target string
	Type   string
	Option string
	Flag   uintptr
}

type Unmounter func() error

func Mount(mountOpts ...MountOption) (Unmounter, error) {
	unmounter := func() error {
		for _, p := range mountOpts {
			if err := syscall.Unmount(p.Target, 0); err != nil {
				return fmt.Errorf("uneable to mount %q: %w", p.Target, err)
			}
		}
		return nil
	}

	for _, p := range mountOpts {
		if stat, err := os.Stat(p.Source); err == nil {
			if stat.IsDir() {
				if err := os.MkdirAll(p.Target, 0755); err != nil {
					return unmounter, fmt.Errorf("uneable to unable to mount %s to %s: %w", p.Source, p.Target, err)
				}
			} else {
				if err := os.MkdirAll(filepath.Dir(p.Target), 0755); err != nil {
					return unmounter, fmt.Errorf("unable to mount %s to %s: %w", p.Source, p.Target, err)
				}
				if err := os.WriteFile(p.Target, []byte{}, 0644); err != nil {
					return unmounter, fmt.Errorf("unable to mount %s to %s: %w", p.Source, p.Target, err)
				}
			}
		}

		if err := syscall.Mount(p.Source, p.Target, p.Type, p.Flag, p.Option); err != nil {
			return unmounter, fmt.Errorf("unable to mount %s to %s: %w", p.Source, p.Target, err)
		}
	}

	return unmounter, nil
}

// /etc/mtab is a symlink to /proc/mount or /proc/self/mount
func parseMtab() ([]mount, error) {
	b, err := os.ReadFile("/etc/mtab")
	if err != nil {
		return nil, err
	}

	d := string(b)
	lines := strings.Split(d, "\n")

	var mounts []mount
	for _, line := range lines {
		fields := strings.Fields(line)
		mnt := mount{
			source:  fields[0],
			target:  fields[1],
			fstype:  fields[2],
			options: fields[3],
		}
		mounts = append(mounts, mnt)
	}
	return mounts, nil
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func baseMounts() error {
	log.Println("Mounting: /dev /proc and /sys")
	mpoints := []string{"/dev", "/proc", "/sys"}
	for _, dir := range mpoints {
		if err := ensureDir(dir); err != nil {
			return err
		}
	}

	mountPoints := []MountOption{
		{
			Source: "proc",
			Target: "/proc",
			Type:   "proc",
			Flag:   syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID,
		},
		{
			Source: "sysfs",
			Target: "/sys",
			Type:   "sysfs",
			Flag:   syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID,
		},
		{
			Source: "dev",
			Target: "/dev",
			Type:   "tmpfs",
			Flag:   syscall.MS_NOSUID,
		},
		// mount -t cgroup2 -o nosuid,nodev,noexec,relatime,nsdelegate,memory_recursiveprot cgroup2 /sys/fs/cgorup
		{
			Source: "cgroup2",
			Target: "/sys/fs/cgroup",
			Type:   "cgroup2",
			Flag:   unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOEXEC | unix.MS_RELATIME,
			Option: "nsdelegate,memory_recursiveprot",
		},
	}

	if _, err := Mount(mountPoints...); err != nil {
		return err
	}

	for _, dir := range []string{"/dev/pts", "/dev/shm"} {
		if err := ensureDir(dir); err != nil {
			return err
		}
	}

	mountPoints = []MountOption{
		{Source: "devpts", Target: "/dev/pts", Type: "devpts"},
		{Source: "shm", Target: "/dev/shm", Type: "tmpfs"},
	}

	if _, err := Mount(mountPoints...); err != nil {
		return err
	}

	if err := os.Symlink("/dev/pts/ptmx", "/dev/ptms"); err != nil {
		return err
	}

	if err := os.Symlink("/proc/mounts", "/etc/mtab"); err != nil {
		return err
	}

	return nil
}
