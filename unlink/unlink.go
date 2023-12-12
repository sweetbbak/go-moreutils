package main

import (
	"flag"
	"log"
	"syscall"
)

func main() {
	flag.Parse()
	flag.Usage = func() {
		println("Usage: unlink <files>")
		println("Will remove files. This operation is destructive.")
	}

	args := flag.Args()

	for _, item := range args {
		err := syscall.Unlink(item)
		if err != nil {
			log.Println(err)
		}
	}
}
