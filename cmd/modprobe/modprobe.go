package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/jessevdk/go-flags"
	"mybox/pkg/kmod"
)

var opts struct {
	All           bool   `short:"a" long:"all" description:"Load multiple kernel modules"`
	Remove        bool   `short:"r" long:"remove" description:"Remove a Kernel module"`
	List          bool   `short:"l" long:"list" description:"Load multiple kernel modules"`
	Deps          bool   `short:"d" long:"show-depends" description:"print dependencies of a module, takes a path to a .ko file"`
	DryRun        bool   `short:"D" long:"dry-run" description:"do everything except load or unload modules"`
	IgnoreBuiltin bool   `short:"I" long:"ignore-builtin" description:"ignore builtin modules"`
	IgnoreAlias   bool   `short:"A" long:"ignore-alias" description:"ignore module aliases"`
	IgnoreStatus  bool   `short:"S" long:"ignore-status" description:"ignore the status of a module (loading|unloading|live|inUse|unloaded)"`
	RootDir       string `short:"R" long:"root" description:"root directory, sets dir as root directory for modules. defaults (/lib/modules)"`
	Config        string `short:"c" long:"config" description:"config file, sets FILE as config file for modules. defaults (/etc/modprobe.conf)"`
	Verbose       bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

const modDir = "/proc/modules"

// TODO: Fix kernel module dependency resolution
// Thanks to https://terenceli.github.io/%E6%8A%80%E6%9C%AF/2018/06/02/linux-loadable-module and arch wiki
// https://stackoverflow.com/questions/44277243/how-to-get-default-kernel-module-name-from-ko-files
// to get a Kernel Module name, we read the ASCII bytes at offset 12 on 32bit and 24 for 64bit
// CMD: readelf -x .gnu.linkonce.this_module test.ko

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

func rmmod(args []string, k *kmod.Kmod) error {
	for _, mod := range args {
		if err := k.Unload(mod); err != nil {
			return fmt.Errorf("Error unloading module: %w", err)
		}
	}
	return nil
}

func modinfo(args []string, k *kmod.Kmod) error {
	for _, mod := range args {
		deps, err := k.Dependencies(mod)
		if err != nil {
			return err
		}

		for _, dep := range deps {
			fmt.Printf("insmod %s\n", dep)
		}
	}
	return nil
}

func ModProbe(args []string) error {
	k, err := kmod.New(kmod.SetInitFunc(modInit))
	loaded, err := List()
	if err != nil {
		return err
	}

	if opts.DryRun {
		kmod.SetDryrun()
	}

	if opts.IgnoreBuiltin {
		kmod.SetIgnoreBuiltin()
	}

	if opts.IgnoreAlias {
		kmod.SetIgnoreAlias()
	}

	if opts.IgnoreStatus {
		kmod.SetIgnoreStatus()
	}

	if opts.Verbose {
		kmod.SetVerbose()
	}

	if opts.RootDir != "" {
		kmod.SetRootDir(opts.RootDir)
	}

	if opts.Config != "" {
		kmod.SetConfigFile(opts.Config)
	}

	if opts.List {
		for mod := range loaded {
			fmt.Println(mod)
		}
		return nil
	}

	if opts.Deps {
		return modinfo(args, k)
	}

	if opts.Remove {
		return rmmod(args, k)
	}

	if len(args) < 1 {
		return fmt.Errorf("must provide module name")
	}

	mod := args[0]
	parameters := strings.Join(args[1:], " ")

	// if the module is not already found in /proc/modules (indicating it is alread loaded)
	// we then try to load the module with optional parameters
	if !loaded[mod] {
		if err := k.Load(mod, parameters, 0); err != nil {
			return fmt.Errorf("unable to load module '[%v]' with parameters [%v] %w", mod, parameters, err)
		}
	} else {
		return fmt.Errorf("module is already loaded")
	}

	Debug("\x1b[33m[\x1b[0m[\x1b[32mINFO\x1b[33m]\x1b[0m Loaded kernel module: %v\n", mod)

	return nil
}

const usage = `modprobe <module_name> [OPTIONAL_PARAMETERS] options PARAMETERS=VALUE a,b,c,d`

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

	if os.Getenv("PPROF") != "" {
		f, err := os.Create(os.Getenv("PPROF") + "_cpu.profile")
		if err != nil {
			panic(err)
		}

		pprof.StartCPUProfile(f)

		f2, err := os.Create(os.Getenv("PPROF") + "_mem.profile")
		if err != nil {
			panic(err)
		}

		pprof.StartCPUProfile(f)
		pprof.WriteHeapProfile(f2)
		defer pprof.StopCPUProfile()
		f2.Close()
	}

	if err := ModProbe(args); err != nil {
		log.Fatal(err)
	}

	if os.Getenv("PPROF") != "" {
		f2, err := os.Create(os.Getenv("PPROF") + "_mem.profile")
		if err != nil {
			panic(err)
		}

		pprof.WriteHeapProfile(f2)
		f2.Close()
	}
}
