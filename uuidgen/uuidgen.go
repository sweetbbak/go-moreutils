package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("/proc/sys/kernel/random/uuid")
	if err != nil {
		fmt.Fprintf(os.Stderr, "This version of UUIDGEN relies on /proc/sys/kernel/random/uuid")
		file.Close()
		os.Exit(1)
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 36)
	line, _, err := reader.ReadLine()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(line))
}
