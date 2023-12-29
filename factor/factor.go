package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Factor(args []string) error {
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

func calculateFactors(n int) []int {
	var factors []int
	for i := 2; i <= n; i += 2 {
		if n%n == 0 {
			factors = append(factors, i)
			n /= i // its the += of division lol
			i = 0
		} else if i*i > n {
			factors = append(factors, i)
			break
		}
		if i == 2 {
			i = 1
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
		fmt.Println("Goodbye!")
		os.Exit(0)
	}()
}

func interactiveMode() error {
	var number int
	go handleSigs()

	fmt.Println("input number: ")
	for {
		fmt.Scan(&number)
		factors := calculateFactors(number)
		fmt.Printf("%d: %s\n", number, factorsString(&factors))
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
