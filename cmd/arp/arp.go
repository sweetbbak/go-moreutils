package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Iname   string `short:"i" long:"interface" description:"print info from a specific interface (ensp6s0)"`
	Verbose bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

type Arp struct {
	addr   string
	hwType string
	flag   string
	hwAddr string
	mask   string
	iname  string
}

type ArpTable map[string]Arp

func getArp(args []string) error {
	table, err := arp()
	if err != nil {
		return err
	}
	fmt.Printf("%-16s %-17s %s\n", "address", "HW Address", "Interface")
	for _, arp := range table {
		if opts.Iname != "" && opts.Iname != arp.iname {
			continue
		}
		fmt.Printf("%-16s %s %s\n", arp.addr, arp.hwAddr, arp.iname)
	}
	return nil
}

func arp() (ArpTable, error) {
	f, err := os.Open("/proc/net/arp")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip a field
	table := make(map[string]Arp)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		a := Arp{
			addr:   fields[0],
			hwType: fields[1],
			flag:   fields[2],
			hwAddr: fields[3],
			mask:   fields[4],
			iname:  fields[5],
		}
		table[a.iname] = a
	}

	return table, nil
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := getArp(args); err != nil {
		log.Fatal(err)
	}
}
