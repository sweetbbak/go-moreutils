package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

const clear = "\x1b[0m"

func monthLength(year, month int) int {
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	var days int

	t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	m := t.Month()
	for t.Month() == m {
		days++
		t = t.AddDate(0, 0, 1)
	}
	Debug("%v\n", days)
	return days
}

func leapyear(year int) int {
	//Return 1 if leapyear, 0 if not
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		return 1
	}
	return 0
}

func monthlen(month int, year int) int {
	//Return length of month in days
	switch month {
	case 1:
		return 31
	case 2:
		return 28 + leapyear(year)
	case 3:
		return 31
	case 4:
		return 30
	case 5:
		return 31
	case 6:
		return 30
	case 7:
		return 31
	case 8:
		return 31
	case 9:
		return 30
	case 10:
		return 31
	case 11:
		return 30
	case 12:
		return 31
	}
	return 0
}

func calendar(month, year int) {
	t := time.Now()
	t2 := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	var highlightDay int = -1
	var format string = ""

	if t.Month() == t2.Month() && t.Year() == t2.Year() {
		highlightDay = t.Day()
		format = "\x1b[7;32m"
	}

	weekday := int(t2.Weekday())
	fmt.Printf("%11s %d\n", t2.Month().String(), year)
	fmt.Println("Su Mo Tu We Th Fr Sa")
	for i := 0; i < weekday; i++ {
		fmt.Print("   ")
	}

	for day := 1; day <= monthlen(month, year); day++ {
		if weekday == 6 {
			if highlightDay != -1 && day == highlightDay {
				fmt.Printf("%s%2d%s\n", format, day, clear)
			} else {
				fmt.Printf("%2d\n", day)
			}

			weekday = 0
		} else {
			if highlightDay != -1 && day == highlightDay {
				fmt.Printf("%s%2d%s ", format, day, clear)
			} else {
				fmt.Printf("%2d ", day)
			}
			weekday++
		}
	}

	if weekday != 7 {
		fmt.Printf("\n")
	}
}

// func printDay()

func Cal(args []string) error {
	Debug("length args: %v\n", len(args))
	switch len(args) {
	case 0:
		year := int(time.Now().Year())
		month := int(time.Now().Month())
		calendar(month, year)
	case 1:
		year, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("error parsing month %v: %v", args[0], err)
		}
		for month := 1; month <= 12; month++ {
			calendar(month, year)
			fmt.Println()
		}
	case 2:
		month, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("error parsing month %v: %v", args[0], err)
		}

		year, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("error parsing month %v: %v", args[0], err)
		}
		calendar(month, year)
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

	if err := Cal(args); err != nil {
		log.Fatal(err)
	}
}
