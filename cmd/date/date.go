package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Reference string `short:"r" long:"reference" description:"display modification time of FILE"`
	Univeral  bool   `short:"u" long:"utc" description:"Coordinated Universal Time (UTC)"`
	Verbose   bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (r RealClock) Now() time.Time {
	return time.Now()
}

var (
	fmtMap = map[string]string{
		"%a": "Mon",
		"%A": "Monday",
		"%b": "Jan",
		"%h": "Jan",
		"%B": "January",
		"%c": time.UnixDate,
		"%d": "02",
		"%e": "_2",
		"%H": "15",
		"%I": "03",
		"%m": "1",
		"%M": "04",
		"%p": "PM",
		"%S": "05",
		"%y": "06",
		"%Y": "2006",
		"%z": "-0700",
		"%Z": "MST",
	}
)

func formatParser(args string) []string {
	r := regexp.MustCompile("%[a-zA-Z]")
	match := r.FindAll([]byte(args), -1)
	var results []string
	for _, m := range match {
		results = append(results, string(m[:]))
	}
	return results
}

func date(t time.Time, z *time.Location) string {
	return t.In(z).Format(time.UnixDate)
}

func dateMap(t time.Time, z *time.Location, format string) string {
	d := t.In(z)
	var toReplace string
	for _, match := range formatParser(format) {
		tl, exists := fmtMap[match]
		switch {
		case exists:
			toReplace = d.Format(tl)
		case match == "%C":
			toReplace = strconv.Itoa(d.Year() / 100)
		case match == "%D":
			toReplace = dateMap(t, z, "%m/%d/%y")
		case match == "%j":
			year, weekyear := d.ISOWeek()
			firstWeekDay := time.Date(year, 1, 1, 1, 1, 1, 1, time.UTC).Weekday()
			weekDay := d.Weekday()
			dayYear := int(weekyear)*7 - (int(firstWeekDay) - 1) + int(weekDay)
			toReplace = strconv.Itoa(dayYear)
		case match == "%n":
			// A <newline>.
			toReplace = "\n"
		case match == "%r":
			// 12-hour clock time [01,12] using the AM/PM notation;
			// in the POSIX locale, this shall be equivalent to %I : %M : %S %p.
			toReplace = dateMap(t, z, "%I:%M:%S %p")
		case match == "%t":
			// A <tab>.
			toReplace = "\t"
		case match == "%T":
			toReplace = dateMap(t, z, "%H:%M:%S")
		case match == "%W":
			// Week of the year (Sunday as the first day of the week)
			// as a decimal number [00,53]. All days in a new year preceding
			// the first Sunday shall be considered to be in week 0.
			_, weekYear := d.ISOWeek()
			weekDay := int(d.Weekday())
			isNotSunday := 1
			if weekDay == 0 {
				isNotSunday = 0
			}
			toReplace = strconv.Itoa(weekYear - (isNotSunday))
		case match == "%w":
			// Weekday as a decimal number [0,6] (0=Sunday).
			toReplace = strconv.Itoa(int(d.Weekday()))
		case match == "%V":
			// Week of the year (Monday as the first day of the week)
			// as a decimal number [01,53]. If the week containing January 1
			// has four or more days in the new year, then it shall be
			// considered week 1; otherwise, it shall be the last week
			// of the previous year, and the next week shall be week 1.
			_, weekYear := d.ISOWeek()
			toReplace = strconv.Itoa(int(weekYear))
		case match == "%x":
			// Locale's appropriate date representation.
			toReplace = dateMap(t, z, "%m/%d/%y") // TODO: decision algorithm
		case match == "%F":
			// Date yyyy-mm-dd defined by GNU implementation
			toReplace = dateMap(t, z, "%Y-%m-%d")
		case match == "%X":
			// Locale's appropriate time representation.
			toReplace = dateMap(t, z, "%I:%M:%S %p") // TODO: decision algorithm
		default:
			continue
		}

		format = strings.Replace(format, match, toReplace, 1)
	}
	return format
}

func ints(s string, i ...*int) error {
	var err error
	for _, p := range i {
		if *p, err = strconv.Atoi(s[0:2]); err != nil {
			return err
		}
		s = s[2:]
	}
	return nil
}

func getTime(z *time.Location, s string, clock Clock) (t time.Time, err error) {
	var MM, DD, hh, mm int
	YY := clock.Now().Year() % 100
	CC := clock.Now().Year() / 100
	SS := clock.Now().Second()
	if err = ints(s, &MM, &DD, &hh, &mm); err != nil {
		return
	}
	s = s[8:]
	switch len(s) {
	case 0:
	case 2:
		err = ints(s, &YY)
	case 3:
		err = ints(s[1:], &SS)
	case 4:
		err = ints(s, &CC, &YY)
	case 5:
		s = s[0:2] + s[3:]
		err = ints(s, &YY, &SS)
	case 7:
		s = s[0:4] + s[5:]
		err = ints(s, &CC, &YY, &SS)
	default:
		err = fmt.Errorf("optional string is %v instead of [[CC]YY][.ss]", s)
	}

	if err != nil {
		return
	}

	YY = YY + CC*100
	t = time.Date(YY, time.Month(MM), DD, hh, mm, SS, 0, z)
	return

}

func setDate(d string, z *time.Location, clock Clock) error {
	t, err := getTime(z, d, clock)
	if err != nil {
		log.Fatalf("%v: %v", d, err)
	}
	tv := syscall.NsecToTimeval(t.UnixNano())
	return syscall.Settimeofday(&tv)
}

func Date(args []string) error {
	rclock := RealClock{}
	t := rclock.Now()
	z := time.Local

	if opts.Univeral {
		z = time.UTC
	}

	if opts.Reference != "" {
		stat, err := os.Stat(opts.Reference)
		if err != nil {
			return fmt.Errorf("unable to stat file: %v", opts.Reference)
		}
		t = stat.ModTime()
	}

	switch len(args) {
	case 0:
		fmt.Printf("%v\n", date(t, z))
	case 1:
		if strings.HasPrefix(args[0], "+") {
			fmt.Printf("%v\n", dateMap(t, z, args[0][1:]))
		} else {
			if err := setDate(args[0], z, rclock); err != nil {
				return fmt.Errorf("%v: %v", args[0], err)
			}
		}
	default:
		return nil
	}
	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Date(args); err != nil {
		log.Fatal(err)
	}
}
