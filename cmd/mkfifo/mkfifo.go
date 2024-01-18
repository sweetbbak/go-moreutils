package main

import (
	"flag"
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

const (
	defaultMode = 0o660 | unix.S_IFIFO // 0x1000 - 4096
)

var mode = flag.Int("mode", defaultMode, "Mode to create the fifo")

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: mkfifo <path>")
		fmt.Println("example: mkfifo -m 0o666 my-fifo")
		log.Fatal("please provide a path, or multiple, to create a fifo")
	}

	for _, fi := range flag.Args() {
		if err := unix.Mkfifo(fi, uint32(*mode)); err != nil {
			log.Fatalf("error while creating fifo: %v", err)
		}
	}
}
