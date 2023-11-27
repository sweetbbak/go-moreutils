package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 1 {
		fmt.Println("which <binary_name>")
		os.Exit(0)
	}

	path, err := exec.LookPath(os.Args[1])
	if err != nil {
		fmt.Printf("%v not found\n", os.Args[1])
	} else {
		fmt.Printf("%v\n", path)
	}

}
