package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/jessevdk/go-flags"
	"golang.org/x/sys/unix"
)

var opts struct {
	Lazy    bool `short:"l" long:"lazy" description:"lazy umount"`
	Force   bool `short:"f" long:"force" description:"force umount"`
	Unmount bool `short:"u" long:"unmount" description:"unmount device"`
	Verbose bool `short:"v" long:"verbose" description:"use verbose output"`
}

var Debug = func(string, ...interface{}) {}

const (
	CommFD  = "_FUSE_COMMFD"
	fuseDev = "/dev/fuse"
)

var fileSystemUID, fileSystemGID int

func dropPrivs() error {
	uid := unix.Getuid()
	if uid == 0 {
		return nil
	}

	var err error
	fileSystemUID, err = unix.SetfsuidRetUid(uid)
	if err != nil {
		return err
	}
	fileSystemGID, err = unix.SetfsgidRetGid(unix.Getgid())
	return err
}

func restorePrivs() {
	if unix.Getuid() == 0 {
		return
	}
	// We're exiting, if there's an error, not much to do.
	unix.Setfsuid(fileSystemUID)
	unix.Setfsgid(fileSystemGID)
}

func preMount() error {
	// I guess this umask is the thing to do.
	unix.Umask(0o33)
	return nil
}

func umount(name string) error {
	options := 0
	if opts.Lazy {
		options |= unix.MNT_DETACH
	}

	return unix.Unmount(name, options)
}

func openFuse() (int, error) {
	return unix.Open("/dev/fuse", unix.O_RDWR, 0)
}
func MountPointOK(mpt string) error {
	// We wait until we can drop privs to test the mpt
	// parameter, since ability to walk the path can
	// differ for root and the real user id.
	if err := dropPrivs(); err != nil {
		return err
	}
	defer restorePrivs()
	mpt = filepath.Clean(mpt)
	r, err := filepath.EvalSymlinks(mpt)
	if err != nil {
		return err
	}
	if r != mpt {
		return fmt.Errorf("resolved path %q and mountpoint %q are not the same", r, mpt)
	}
	// I'm not sure why fusermount wants to open the mountpoint, so let's mot for now.
	// And, for now, directories only? We don't see a current need to mount
	// FUSE on any other type of file.
	if err := os.Chdir(mpt); err != nil {
		return err
	}

	return nil
}

func getCommFD() (int, error) {
	commfd, ok := os.LookupEnv(CommFD)
	if !ok {
		return -1, fmt.Errorf(CommFD + "was not set and this program can't be used interactively")
	}
	Debug("CommFD %v", commfd)

	cfd, err := strconv.Atoi(commfd)
	if err != nil {
		return -1, fmt.Errorf("%s: %v", CommFD, err)
	}
	Debug("CFD is %v", cfd)
	var st unix.Stat_t
	if err := unix.Fstat(cfd, &st); err != nil {
		return -1, fmt.Errorf("_FUSE_COMMFD: %d: %v", cfd, err)
	}
	Debug("cfd stat is %v", st)

	return cfd, nil
}

func doMount(fd int) error {
	flags := uintptr(unix.MS_NODEV | unix.MS_NOSUID)
	// From the kernel:
	// if (!d->fd_present || !d->rootmode_present ||
	//	!d->user_id_present || !d->group_id_present)
	//		return 0;
	// Yeah. You get EINVAL if any one of these is not set.
	// Docs? what? Docs?
	return unix.Mount("nodev", ".", "fuse", flags, fmt.Sprintf("rootmode=%o,user_id=0,group_id=0,fd=%d", unix.S_IFDIR, fd))
}

// returnResult returns the result from earlier operations.
// It is called with the control fd, a FUSE fd, and an error.
// If the error is not nil, then we are shutting down the cfd;
// If it is nil then we try to send the fd back.
// We return either e or the error result and e
func returnResult(cfd, ffd int, e error) error {
	if e != nil {
		if err := unix.Shutdown(cfd, unix.SHUT_RDWR); err != nil {
			return fmt.Errorf("shutting down after failed mount with %v: %v", e, err)
		}
		return e
	}
	oob := unix.UnixRights(int(ffd))
	if err := unix.Sendmsg(cfd, []byte(""), oob, nil, 0); err != nil {
		return fmt.Errorf("%s: %d: %v", CommFD, cfd, err)
	}
	return nil
}

func Fusermount(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Must provide at least 1 argument")
	}

	mpt := args[0]
	Debug("Mountpoint: %v\n", mpt)

	fusefd, err := openFuse()
	if err != nil {
		log.Printf("%v", err)
		os.Exit(int(syscall.ENOENT))
	}

	Debug("Fuse fd: %v\n", fusefd)

	if opts.Lazy && !opts.Unmount {
		return fmt.Errorf("--lazy cannot be used with --unmount")
	}

	if opts.Unmount {
		if err := umount(mpt); err != nil {
			return fmt.Errorf("fusermount: unmount: %v", err)
		}
	}

	if err := MountPointOK(mpt); err != nil {
		return err
	}

	if err := preMount(); err != nil {
		return err
	}

	cfd, err := getCommFD()
	if err != nil {
		return err
	}

	if err := doMount(fusefd); err != nil {
		return err
	}

	if err := returnResult(cfd, fusefd, err); err != nil {
		return err
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

	if err := Fusermount(args); err != nil {
		log.Fatal(err)
	}
}
