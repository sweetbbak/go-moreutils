package main

import (
	"errors"
	"os"
	"syscall"
)

const (
	ioprioWhoProcess = 1 + iota
	ioprioWhoProcessGroup
	ioprioWhoUser
)

type Niceness int

const (
	VeryLow Niceness = iota
	Low
	Standard
	High
	VeryHigh
)

func SetIoPriority(niceness int) error {
	return setIoPriority(niceness)
}

// func SetIoPriority(niceness Niceness) error {
// 	return setIoPriority(niceness)
// }

func GetIoPriority() (Niceness, error) {
	return getIoPriority()
}

// Mapping from Niceness to values for ioprio_set. Lower values to ioprio_set mean a higher priority.
var priorityMapping = map[Niceness]int{
	VeryLow:  7,
	Low:      6,
	Standard: 4,
	High:     2,
	VeryHigh: 0,
}

func nicenessToInt(n Niceness) (int, error) {
	priority, isValid := priorityMapping[n]
	if !isValid {
		return -1, errors.New("invalid niceness specified")
	}
	return priority, nil
}

func setIoPriority(priority int) error {
	// priority, isValid := priorityMapping[niceness]
	// if !isValid {
	// 	return errors.New("invalid niceness specified")
	// }
	r1, _, e1 := syscall.Syscall(syscall.SYS_IOPRIO_SET, ioprioWhoProcess, uintptr(os.Getpid()), uintptr(ioprioPrioValue(ioprioClassBe, priority)))
	if int(r1) == -1 {
		return e1
	}
	return nil
}

func getIoPriority() (Niceness, error) {
	ioprio, _, e1 := syscall.Syscall(syscall.SYS_IOPRIO_GET, ioprioWhoProcess, uintptr(os.Getpid()), 0)
	if int(ioprio) == -1 {
		return 0, e1
	}
	ioprioClass, ioprioData := ioprioPrioClass(int(ioprio)), ioprioPrioData(int(ioprio))
	switch ioprioClass {
	case ioprioClassNone:
		return Standard, nil
	case ioprioClassRt:
		return VeryHigh, nil
	case ioprioClassIdle:
		return VeryLow, nil
	case ioprioClassBe:
		// Find closest matching niceness
		for _, niceness := range []Niceness{VeryLow, Low, Standard, High, VeryHigh} {
			if ioprioData >= priorityMapping[niceness] {
				return niceness, nil
			}
		}
		// Never reached - ioprioData has maximum of 7, which is VeryHigh
		panic("invalid ioprio data")
	}
	return 0, errors.New("unknown ioprio priority class")
}

// Priority code from https://sources.debian.org/src/linux/5.19.6-1/include/uapi/linux/ioprio.h/

/*
 * Gives us 8 prio classes with 13-bits of data for each class
 */
const (
	ioprioClassShift = 13
	ioprioClassMask  = 0x07
	ioprioPrioMask   = (1 << ioprioClassShift) - 1
)

func ioprioPrioClass(ioprio int) int {
	return (ioprio >> ioprioClassShift) & ioprioClassMask
}

func ioprioPrioData(ioprio int) int {
	return ioprio & ioprioPrioMask
}

func ioprioPrioValue(class int, data int) int {
	return ((class & ioprioClassMask) << ioprioClassShift) | (data & ioprioPrioMask)
}

const (
	ioprioClassNone = iota
	ioprioClassRt   // realtime class, it always gets premium service. For ATA disks supporting NCQ IO priority, RT class IOs will be processed using high priority NCQ commands.
	ioprioClassBe   // best-effort scheduling class, the default for any process
	ioprioClassIdle // idle scheduling class, it is only served when no one else is using the disk
)

/*
 * The RT and BE priority classes both support up to 8 priority levels.
 */
const ioprioNrLevels = 8
const ioprioBeNr = ioprioNrLevels
