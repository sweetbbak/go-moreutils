package internal

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
)

// Gettys spawn the n number of getty processes on tty(s)
// If persist is true, they'll be respawned if they die.
func Gettys(ctx context.Context, n int, persist bool) {
	// getty(tty) blocks, so spawn a goroutine for each one and wait
	// for them to finish with a waitgroup, respawning as necessary in the
	// goroutine if it happens to quit. (NB: if persist is true they will
	// never finish.)
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(tty string) {
			defer wg.Done()
			for {
				err := <-getty(tty)
				if err != nil {
					log.Println(err)
				}
				if !persist {
					return
				}
			}
		}("tty" + strconv.Itoa(i))
	}
	// Block until all the ttys we spawned in goroutines are finished instead of
	// returning right away (and shutting down the system.)
	wg.Wait()
}

// Spawn a single getty on tty
// /bin/getty /dev/tty0 115200 linux
func getty(tty string) chan error {
	errCh := make(chan error)

	f, err := os.Open(filepath.Join("/dev", tty))
	if err != nil {
		errCh <- err
		return errCh
	}
	defer f.Close()

	cmd := exec.Command("/bin/getty", tty, "115200", "linux")

	// If we don't Setsid, we'll get an "inappropriate ioctl for device"
	// error upon starting the login shell.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = append(cmd.ExtraFiles, f)

	errCh <- cmd.Run()
	return errCh
}
