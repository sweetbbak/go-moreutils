package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"mybox/pkg/go-modprobe"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	All     bool `short:"a" long:"all" description:"Load multiple kernel modules"`
	Remove  bool `short:"r" long:"remove" description:"Remove a Kernel module"`
	List    bool `short:"l" long:"list" description:"Load multiple kernel modules"`
	Deps    bool `short:"d" long:"show-depends" description:"print dependencies of a module, takes a path to a .ko file"`
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

const modDir = "/proc/modules"

func List() (map[string]bool, error) {
	loaded := map[string]bool{}
	f, err := os.Open(modDir)
	if err != nil {
		return loaded, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		loaded[strings.SplitN(sc.Text(), " ", 2)[0]] = true
	}

	return loaded, err
}

func rmmod(args []string) error {
	for _, mod := range args {
		if err := modprobe.Remove(mod); err != nil {
			return fmt.Errorf("Error unloading module: %v", err)
		}
	}
	return nil
}

func modinfo(args []string) error {
	for _, mname := range args {
		modname, err := modprobe.ResolveName(mname)
		if err != nil {
			// return err
		}

		Debug("Module path resolved to: %v\n", modname)

		deps, err := modprobe.Dependencies(modname)
		if err != nil {
			return err
		}

		for _, dep := range deps {
			fmt.Println(dep)
		}
	}
	return nil
}

func ModProbe(args []string) error {
	loaded, err := List()
	if err != nil {
		return err
	}

	if opts.List {
		for mod := range loaded {
			fmt.Println(mod)
		}
		return nil
	}

	if opts.Deps {
		return modinfo(args)
	}

	if opts.Remove {
		return rmmod(args)
	}

	if len(args) < 1 {
		return fmt.Errorf("must provide module name")
	}

	mod := args[0]
	parameters := strings.Join(args[1:], " ")

	// if the module is not already found in /proc/modules (indicating it is alread loaded)
	// we then try to load the module with optional parameters
	if !loaded[mod] {
		if err := modprobe.Load(mod, parameters); err != nil {
			return fmt.Errorf("unable to load module '[%v]' with parameters [%v] %v", mod, parameters, err)
		}

		Debug("\x1b[33m[\x1b[0m[\x1b[32mINFO\x1b[33m]\x1b[0m Loaded kernel module: %v\n", mod)
	} else {
		return fmt.Errorf("module is already loaded")
	}

	return nil
}

const usage = `modprobe <module_name> [OPTIONAL] PARAMETERS=VALUE`

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		fmt.Println(usage)
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := ModProbe(args); err != nil {
		log.Fatal(err)
	}
}
