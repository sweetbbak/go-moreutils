package main

import (
	"github.com/pkg/term"
	"golang.org/x/sys/unix"
	"syscall"
	"unsafe"
)

type winsize struct {
	rows    uint16
	cols    uint16
	xpixels uint16
	ypixels uint16
}

func get_term_size(fd uintptr) (int, int) {
	var sz winsize
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
		fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return int(sz.cols), int(sz.rows)
}

func getChar() (ascii int, keyCode int, err error) {
	t, _ := term.Open("/dev/tty")

	term.RawMode(t)
	bytes := make([]byte, 3)

	var numRead int
	numRead, err = t.Read(bytes)
	if err != nil {
		return
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bytes[0])
	} else {
		// Two characters read??
	}
	t.Restore()
	t.Close()
	return
}

func makeRaw(fd uintptr) (*unix.Termios, error) {
	var oldState unix.Termios
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(unix.TCGETS), uintptr(unsafe.Pointer(&oldState)))
	if err != 0 {
		return nil, err
	}

	newState := oldState
	newState.Lflag &^= unix.ECHO
	newState.Lflag &= unix.ECHO

	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(unix.TCSETS), uintptr(unsafe.Pointer(&newState)))
	if err != 0 {
		return nil, err
	}

	return &oldState, nil
}

func restoreTerminal(fd uintptr, oldState *unix.Termios) {
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(unix.TCSETS), uintptr(unsafe.Pointer(oldState)))
}
