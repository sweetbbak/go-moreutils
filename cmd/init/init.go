package main

import (
	"context"
	"fmt"
	"log"
	"mybox/pkg/xinit"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "golang.org/x/sys/unix"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Config   string `short:"c" long:"conf" description:"path to config file"`
	Hostname string `long:"hostname" description:"set default hostname"`
	Verbose  bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

const (
	CONFIG_FILE = ""
	CONFIG_DIR  = ""
	LOG_FILE    = ""
	HOSTNAME    = ""
	REBOOT_CMD  = 0
	defaultPath = "/sbin:/usr/sbin:/bin:/usr/bin"
	defaultHome = "/root"
	defaultTerm = "linux"
)

func reapZombies() {
	for {
		var status syscall.WaitStatus

		// this many zombie processes indicates an issue
		for i := 0; i < 10; i++ {
			syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
		}

		// block and wait for zombies
		syscall.Wait4(-1, &status, 0, nil)
	}
}

func sysInit(start time.Time) error {
	os.Setenv("PATH", defaultPath)

	if err := baseMounts(); err != nil {
		return err
	}

	if err := internal.CreateDevices(); err != nil {
		return err
	}

	if err := internal.ScanDevices(context.Background()); err != nil {
		return err
	}

	if err := internal.SetupNetworkInterfaces(); err != nil {
		return err
	}

	if err := userSetHostname(opts.Hostname); err != nil {
		return err
	}

	services, err := internal.ParseServiceConfigs(CONFIG_DIR)
	if err != nil {
		return err
	}

	ctx, stop := context.WithCancel(context.Background())

	go reapZombies()
	go internal.WatchDevices(ctx)
	go internal.StartServices(services)
	go internal.Gettys(ctx, 1, true)

	go func() {
		_, err := internal.DHCPClient("eth0")
		if err != nil {
			log.Panicln(err)
			return
		}

		go internal.StartSSHServer(ctx)
	}()

	fmt.Printf("Boot time: %v\n", time.Since(start))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
	s := <-sig
	stop()

	fmt.Println("System shutting down...")

	switch s {
	case syscall.SIGINT, syscall.SIGTERM:
		Exit(REBOOT_RESTART, services)
	case syscall.SIGUSR1:
		Exit(REBOOT_HALT, services)
	case syscall.SIGUSR2:
		Exit(REBOOT_POWEROFF, services)
	}

	return nil
}

func main() {
	_, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	start := time.Now()

	if os.Getpid() != 1 {
		fmt.Fprintln(os.Stderr, "Must be ran os PID 1")
		os.Exit(0)
	}

	if opts.Config == "" {
		opts.Config = CONFIG_FILE
	}
	if opts.Hostname == "" {
		opts.Hostname = HOSTNAME
	}

	w, err := os.OpenFile(opts.Config, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Panicln(err.Error())
	}
	defer w.Close()
	log.SetOutput(w)

	if err := sysInit(start); err != nil {
		log.Panicln(err)
	}
}
