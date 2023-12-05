package main

import (
	"bufio"
	"log"
	"os"
)

func parseFlags() {
	args := os.Args[1:]
	if len(args) == 1 && args[0] == "-h" || args[0] == "--help" {
		println("usage: ts")
		println("")
		println("ts reads from stdin and appends a timestamp to the beginning of each line of input")
		println("examples:\n\r\techo $MY_VARIABLE | ts [outputs] > 2023/12/1 12:00 my_variable")
		println("\tts < file")
		os.Exit(0)
	}
}

func main() {
	parseFlags()
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		log.Printf("%v ", s.Text())
	}
}
