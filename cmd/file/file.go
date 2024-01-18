package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	m, err := Open(MAGIC_NONE)
	if err != nil {
		log.Fatal(err)
	}

	defer m.Close()

	for _, file := range os.Args[1:] {
		res, err := m.File(file)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", file, res)
	}
}
