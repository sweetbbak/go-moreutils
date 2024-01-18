package main

import (
	"math"
	"os"
	"syscall"
	"time"
)

type FileInfo struct {
	Name          string
	Mode          os.FileMode
	Rdev          uint64
	UID, GID      uint32
	Size          int64
	MTime         time.Time
	SymlinkTarget string
}

func FromOSFileInfo(path string, fi os.FileInfo) FileInfo {
	var link string
	UID, GID, rdev := uint32(math.MaxUint32), uint32(math.MaxUint32), uint64(math.MaxUint64)
	if s, ok := fi.Sys().(*syscall.Stat_t); ok {
		UID, GID, rdev = s.Uid, s.Gid, uint64(s.Rdev)
	}

	if fi.Mode()&os.ModeType == os.ModeSymlink {
		if l, err := os.Readlink(path); err != nil {
			link = err.Error()
		} else {
			link = l
		}
	}
	return FileInfo{
		Name:          fi.Name(),
		Mode:          fi.Mode(),
		Rdev:          rdev,
		UID:           UID,
		GID:           GID,
		Size:          fi.Size(),
		MTime:         fi.ModTime(),
		SymlinkTarget: link,
	}
}
