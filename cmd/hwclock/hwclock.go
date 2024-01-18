package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

var (
	write      = flag.Bool("w", false, "set the hwclock from system clock in UTC time")
	unixFormat = flag.Bool("x", false, "get hardware clock time in a Unix epoch format")
)

type RTC struct {
	file *os.File
	syscalls
}

type syscalls interface {
	ioctlGetRTCTime(int) (*unix.RTCTime, error)
	ioctlSetRTCTime(int, *unix.RTCTime) error
}

type realSyscalls struct{}

// wrapper around syscall that sets and gets the hardware times
func (sc realSyscalls) ioctlGetRTCTime(fd int) (*unix.RTCTime, error) {
	return unix.IoctlGetRTCTime(fd)
}

func (sc realSyscalls) ioctlSetRTCTime(fd int, time *unix.RTCTime) error {
	return unix.IoctlSetRTCTime(fd, time)
}

// Read implements Read for the Linux RTC
func (r *RTC) Read() (time.Time, error) {
	rt, err := r.ioctlGetRTCTime(int(r.file.Fd()))
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(int(rt.Year)+1900,
		time.Month(rt.Mon+1),
		int(rt.Mday),
		int(rt.Hour),
		int(rt.Min),
		int(rt.Sec),
		0,
		time.UTC), nil
}

// Set implements Set for the Linux RTC
func (r *RTC) Set(tu time.Time) error {
	rt := unix.RTCTime{
		Sec:   int32(tu.Second()),
		Min:   int32(tu.Minute()),
		Hour:  int32(tu.Hour()),
		Mday:  int32(tu.Day()),
		Mon:   int32(tu.Month() - 1),
		Year:  int32(tu.Year() - 1900),
		Wday:  int32(0),
		Yday:  int32(0),
		Isdst: int32(0),
	}

	return r.ioctlSetRTCTime(int(r.file.Fd()), &rt)
}

func openRTC() (*RTC, error) {
	devices := []string{
		"/dev/rtc",
		"/dev/rtc0",
		"/dev/misc/rtc0",
	}

	for _, dev := range devices {
		f, err := os.Open(dev)
		if err == nil {
			return &RTC{f, realSyscalls{}}, err
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return nil, errors.New("no RTC device was found")
}

func (r *RTC) Close() error {
	return r.file.Close()
}

func main() {
	r, err := openRTC()
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	if *write {
		tnow := time.Now().UTC()
		if err := r.Set(tnow); err != nil {
			log.Fatal(err)
		}
	}

	t, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}

	x := t.Local().Format("Mon Jan_2 15:04:05 MST 2006")
	if *unixFormat {
		fmt.Println(x)
	} else {
		fmt.Println(t.Local().Format("Mon 2 Jan 2006 15:04:05 AM MST"))
	}
}
