package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// refresh reloads data associated with the process
func (p *UnixProcess) Refresh() error {
	statPath := fmt.Sprintf("/proc/%d/stat", p.pid)
	dataBytes, err := os.ReadFile(statPath)
	if err != nil {
		return err
	}

	// parse out image name
	data := string(dataBytes)
	binStart := strings.IndexRune(data, '(') + 1
	binEnd := strings.IndexRune(data[binStart:], ')')
	p.binary = data[binStart : binStart+binEnd]
	p.mem = p.Mem()

	// move past image name and start parsing the rest
	data = data[binStart+binEnd+2:]
	// info for this sscanf comes from the man page - man 'proc(5)' under the proc/[pid]/stat section
	// it specifies: 1 %d pid - 2 %s comm - 3 %c state - 4 %d parent 5 %d proc group 6 %d session etc...
	// more info exists about each field there. As well as the short hand IDs for state. There are 52 fields
	_, err = fmt.Sscanf(data, "%c %d %d %d", &p.state, &p.ppid, &p.pgrp, &p.sid)
	return err
}

func GetMemory(pid int) (string, error) {
	if exists, _ := findProcess(pid); exists != nil {
		procStat, err := os.Open("/proc/" + strconv.Itoa(pid) + "/status")
		if err != nil {
			return "", err
		}
		match := "VmSize:"
		scanner := bufio.NewScanner(procStat)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), match) {
				splits := strings.SplitN(scanner.Text(), ":", 2)
				if len(splits) == 2 {
					str := strings.TrimSpace(splits[1])
					if strings.Contains("kB", str) {
						s := strings.Split(str, " ")
						str = s[0]
					}
					return str, nil
				}
				break
			}
		}
	}
	return "", nil
}

// Returns start time of process, in number of clock ticks after
// system boot. See "man 5 proc" -> /proc/[pid]/stat -> field 22
// for details
func StartTime(pid int) (int, error) {
	// pid := p.pid
	if exists, _ := findProcess(pid); exists != nil {
		procStat, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
		if err != nil {
			return 0, err
		}

		statData := strings.Split(string(procStat), " ")
		startTime, err := strconv.Atoi(statData[21])
		if err != nil {
			return 0, err
		}

		return startTime, nil
	}
	return 0, nil
}

func Cmdline(pid int) (string, error) {
	if exists, _ := findProcess(pid); exists != nil {
		procStat, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
		if err != nil {
			return "", err
		}
		procStat = bytes.ReplaceAll(procStat, []byte("\x00"), []byte(" "))
		return string(procStat), nil
	}
	return "", nil
}
