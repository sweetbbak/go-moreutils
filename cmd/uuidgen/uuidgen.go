package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Kernel  bool `short:"k" long:"kernel" description:"use the kernels native UUID functionality"`
	Crypt   bool `short:"r" long:"random" description:"use cryptographically random numbers to generate a UUID [default]"`
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

type UUID [16]byte

// create a new uuid v4
func NewUUID() *UUID {
	u := &UUID{}
	_, err := rand.Read(u[:16])
	if err != nil {
		panic(err)
	}

	u[8] = (u[8] | 0x80) & 0xBf
	u[6] = (u[6] | 0x40) & 0x4f
	return u
}

func (u *UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func UUIDFromKernel() error {
	file, err := os.Open("/proc/sys/kernel/random/uuid")
	if err != nil {
		return fmt.Errorf("Error getting kernel uuid: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 36)
	line, _, err := reader.ReadLine()
	if err != nil {
		return err
	}

	fmt.Println(string(line))
	return nil
}

func Run(args []string) error {
	if !opts.Crypt && !opts.Kernel || opts.Crypt {
		u := NewUUID()
		fmt.Println(u.String())
	}

	if opts.Kernel {
		return UUIDFromKernel()
	}

	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Run(args); err != nil {
		log.Fatal(err)
	}
}
