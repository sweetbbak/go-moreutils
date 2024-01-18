package main

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func setDefault(fd uintptr) (*unix.Termios, error) {
	// example of my default terminal
	// var cc [19]uint8
	// cc = [19]uint8{3, 28, 127, 21, 4, 0, 1, 0, 17, 19, 26, 0, 18, 15, 23, 22, 0, 0, 0}
	// d := unix.Termios{
	// 	Iflag:  uint32(17644),
	// 	Oflag:  uint32(5),
	// 	Cflag:  uint32(191),
	// 	Lflag:  uint32(35387),
	// 	Line:   0,
	// 	Cc:     cc,
	// 	Ispeed: uint32(0),
	// 	Ospeed: uint32(0),
	// }

	var oldState unix.Termios
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(unix.TCGETS), uintptr(unsafe.Pointer(&oldState)))
	if err != 0 {
		return nil, err
	}

	newState := oldState
	newState.Iflag |= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	newState.Oflag |= syscall.OPOST
	newState.Cflag |= syscall.CS8
	newState.Lflag |= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	newState.Cc[syscall.VMIN+1] = 0
	newState.Cc[syscall.VTIME+1] = 1

	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(unix.TCSETS), uintptr(unsafe.Pointer(&newState)))
	if err != 0 {
		return nil, err
	}
	return &oldState, nil
}

func main() {
	reset := "\033c\033(B\033[m\033[J\033[?25h"
	println(reset)
	setDefault(os.Stdin.Fd())
}
