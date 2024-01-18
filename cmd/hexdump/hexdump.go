package main

import (
	"encoding/hex"
	"flag"
	"io"
	"log"
	"os"
)

func Hexdump(args []string) error {
	var readers []io.Reader
	if len(args) == 0 {
		readers = []io.Reader{os.Stdin}
	} else {
		readers = make([]io.Reader, 0, len(args))

		for _, filename := range args {
			f, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer f.Close()
			readers = append(readers, f)
		}
	}

	r := io.MultiReader(readers...)
	w := hex.Dumper(os.Stdout)
	defer w.Close()

	if _, err := io.Copy(w, r); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	if err := Hexdump(flag.Args()); err != nil {
		log.Fatal(err)
	}
}
