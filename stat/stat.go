package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	JsonOut bool `short:"j" long:"json" description:"output in json"`
	Color   bool `short:"c" long:"color" description:"color output"`
}

type fStat struct {
	file     string
	filetype string
	mode     uint32
	stat     *syscall.Stat_t
	sstat    fs.FileInfo
	usr      *user.User
	grp      *user.Group
}

func printStat(fst *fStat) {
	var format string
	if opts.JsonOut {
		// format = `[{"File": "%v",\n"Type": "%v",\n"Size": "%-11v",\n"Blocks": "%-11v",\n"Device": "%-11v",\n"Inode": "%-11v",\n"Links": "%-11v",\n"Access": "(%04o/%v)",\n"Uid": "(%5v/%8v)",\n"Gid": "(%5v/%8v)",\n"AccessTime": "%v",\n"Modify": "%v",\n"Change": "%v"\n}]`
		format = `[{"File": "%v","Type": "%v","Size": "%v","Blocks": "%v","Block": "%v","Device": "%v","Inode": "%v","Links": "%v","Access": "(%04o/%v)","Uid": "(%v/%v)","Gid": "(%v/%v)","AccessTime": "%v","Modify": "%v","Change": "%v"}]`
	} else {
		format = "  File: %v\n  Type: %v\n  Size: %-11v Blocks: %-11v IO Block: %-11v\nDevice: %#-11v Inode : %-11v Links   : %-11v\nAccess: (%04o/%v)  Uid: (%5v/%8v)  Gid: (%5v/%8v)\nAccess: %v\nModify: %v\nChange: %v\n"
	}
	fmt.Printf(format, fst.file, fst.filetype, fst.sstat.Size(), fst.stat.Blocks, fst.stat.Blksize, fst.stat.Dev, fst.stat.Ino, fst.stat.Nlink, fst.mode, fst.sstat.Mode().String(), fst.stat.Uid, fst.usr.Username, fst.stat.Gid, fst.grp.Name, timespecToTime(fst.stat.Atim), timespecToTime(fst.stat.Mtim), timespecToTime(fst.stat.Ctim))
}

func stat(args []string) error {
	for _, file := range args {
		var fst fStat
		stat, err := os.Stat(file)
		if err != nil {
			log.Printf("Couldnt stat file %v: %v", file, err)
		}
		filetype := "regular file"
		if stat.IsDir() {
			filetype = "directory"
		}

		stat_t := stat.Sys().(*syscall.Stat_t)
		fileMode := uint32(stat.Mode())
		if fileMode > 2147483648 {
			fileMode -= 2147483648
		}

		username, err := user.LookupId(strconv.Itoa(int(stat_t.Uid)))
		if err != nil {
			log.Printf("Couldnt stat file owner %v: %v", username, err)
		}
		groupname, err := user.LookupGroupId(strconv.Itoa(int(stat_t.Gid)))
		if err != nil {
			log.Printf("Couldnt stat file owner %v: %v", groupname, err)
		}
		fst.file = file
		fst.filetype = filetype
		fst.stat = stat_t
		fst.sstat = stat
		fst.mode = fileMode
		fst.usr = username
		fst.grp = groupname
		printStat(&fst)
	}
	return nil
}

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		if err == flags.ErrHelp {
			os.Exit(0)
		}
		log.Fatal(err)
	}

	if len(args) < 1 {
		os.Exit(1)
	}

	if err := stat(args); err != nil {
		log.Fatal(err)
	}
}
