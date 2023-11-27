package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) == 0 {
		fmt.Println("which <binary_name>")
		os.Exit(0)
	}

	for _, arg := range os.Args[1:] {
		path, err := exec.LookPath(arg)
		if err != nil {
			fmt.Printf("%v not found\n", arg)
		} else {
			fmt.Printf("%v\n", path)
		}
	}

}
