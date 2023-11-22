package main

import (
	"fmt"
	"os"
)

func main() {
	var args []string
	for _, x := range os.Args[1:] {
		args = append(args, x)
	}

	fm := args[1]
	fm = string(fm)
	for _, a := range args {
		fmt.Printf(fm, a)
	}
}
