package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	pretty = flag.Bool("p", false, "Pretty print uptime")
)

func uptime(contents string) (*time.Time, error) {
	uptimeArray := strings.Fields(contents)
	if len(uptimeArray) == 0 {
		return nil, errors.New("error:The contents of /proc/uptime are empty")
	}

	uptimeDuration, err := time.ParseDuration(string(uptimeArray[0]) + "s")
	if err != nil {
		return nil, err
	}

	// uptime := time.Time{}.Add(uptimeDuration.Abs())
	uptime := time.Time{}.Add(uptimeDuration)
	return &uptime, nil
}

func loadAvgCompute(contents string) (loadaverage string, err error) {
	loadav := strings.Fields(contents)
	if len(loadav) < 3 {
		return "", fmt.Errorf("error:invalid contents of /proc/loadavg")
	}

	var str string
	if *pretty {
		// str = fmt.Sprintf("\x1b[35m%s\x1b[0m, \x1b[35m%s\x1b[0m, \x1b[35m%s\x1b[0m", loadav[0], loadav[1], loadav[2])
		str = fmt.Sprintf("CPU Load AVG:\n1min: \x1b[35m%s\x1b[0m\n5min: \x1b[35m%s\x1b[0m\n15min: \x1b[35m%s\x1b[0m", loadav[0], loadav[1], loadav[2])
	} else {
		// str = fmt.Sprintf("CPU Load AVG:\n1min: %s, %s, %s", loadav[0], loadav[1], loadav[2])
		str = fmt.Sprintf("CPU Load AVG:\n1min: %s\n5min: %s\n15min: %s", loadav[0], loadav[1], loadav[2])
	}

	return str, nil
}

func loadavg() string {
	procLoadavg, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		log.Fatalf("error reading /proc/loadavg: %v\n", err)
	}
	loadAvg, err := loadAvgCompute(string(procLoadavg))
	return loadAvg
}

func main() {
	flag.Parse()

	procUptime, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Fatalf("error reading /proc/uptime: %v\n", err)
	}

	uptime, err := uptime(string(procUptime))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	load := loadavg()

	if *pretty {
		fmt.Printf("\x1b[35m%s\x1b[0m up \x1b[35m%d\x1b[0m days, \x1b[35m%d\x1b[0m hours, \x1b[35m%d\x1b[0m min\n", time.Now().Format("15:33:01"), (uptime.Day() - 1), uptime.Hour(), uptime.Minute())
		fmt.Printf("%s\n", load)
	} else {
		fmt.Printf("%s up %d days, %d hours, %d min\n", time.Now().Format("15:33:01"), (uptime.Day() - 1), uptime.Hour(), uptime.Minute())
		fmt.Printf("\n%s\n", load)
	}

}
