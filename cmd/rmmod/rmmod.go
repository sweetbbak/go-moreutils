package main

import (
	"log"
	"os"
	"syscall"

	"mybox/pkg/kmodule"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("rmmod: missing module name\n")
	}

	for _, modname := range os.Args[1:] {
		if err := kmodule.Delete(modname, syscall.O_NONBLOCK); err != nil {
			log.Fatalf("rmmod: %v", err)
		}
	}
}
