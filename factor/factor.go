package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Factor(args []string) error {
	// if 0 arguments and stdin is a pipe append those numbers to args
	if len(args) == 0 && isOpen() {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := sc.Text()
			if strings.Contains(line, " ") {
				ll := strings.Split(line, " ")
				for _, subarg := range ll {
					args = append(args, subarg)
				}
			} else {
				args = append(args, line)
			}
		}
	}

	if len(args) == 0 {
		return interactiveMode()
	}

	for _, n := range args {
		number, err := convNumber(n)
		if err != nil {
			return err
		}

		factors := calculateFactors(number)
		fmt.Printf("%d: %s\n", number, factorsString(&factors))
	}

	return nil
}

func calculateFactors(number int) []int {
	var factors []int
	for index := 2; index <= number; index += 2 {
		if number%index == 0 {
			factors = append(factors, index)
			number /= index
			index = 0
		} else if index*index > number {
			factors = append(factors, number)
			break
		} else if index*index == number {
			factors = append(factors, index)
			factors = append(factors, index)
			break
		}
		if index == 2 {
			index = 1
		}
	}
	return factors
}

func factorsString(nums *[]int) string {
	var buf bytes.Buffer
	for _, n := range *nums {
		buf.WriteString(" " + strconv.Itoa(n))
	}
	return buf.String()
}

func handleSigs() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println()
		fmt.Println("Goodbye!")
		os.Exit(0)
	}()
}

func interactiveMode() error {
	var number int
	var a string
	go handleSigs()

	fmt.Println("input number: ")
	for {
		var err error

		fmt.Scan(&a)
		if a == "exit" {
			os.Exit(0)
		}

		number, err = convNumber(a)
		if err != nil {
			fmt.Println("Invalid number")
		}

		factors := calculateFactors(number)
		fmt.Printf("%d:%s\n", number, factorsString(&factors))
	}
}

func isOpen() bool {
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice == os.ModeCharDevice {
		return false
	} else {
		return true
	}
}

func convNumber(n string) (int, error) {
	num, err := strconv.Atoi(n)
	if err != nil {
		return -1, err
	}
	return num, nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Factor(args); err != nil {
		log.Fatal(err)
	}
}
