package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opt struct {
}

func md5Sum(r io.Reader) ([]byte, error) {
	md5gen := md5.New()
	if _, err := io.Copy(md5gen, r); err != nil {
		return nil, err
	}
	return md5gen.Sum(nil), nil
}

func md(w io.Writer, r io.Reader, args ...string) error {
	var err error
	if len(args) == 0 {
		h, err := md5Sum(r)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "%x\n", h)
		if err != nil {
			return err
		}
	} else {
		fileDesc, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer fileDesc.Close()
		h, err := md5Sum(fileDesc)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "%x %s\n", h, args[0])
		if err != nil {
			return err
		}
	}
	return err
}

func main() {
	args, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(0)
	}

	if err := md(os.Stdout, os.Stdin, args...); err != nil {
		log.Fatal(err)
	}
}
