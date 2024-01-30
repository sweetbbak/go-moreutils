package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/jessevdk/go-flags"
	"github.com/u-root/u-root/pkg/dhclient"
	"github.com/vishvananda/netlink"
)

var opts struct {
	Timeout  int    `short:"t" long:"timeout" description:"timeout in seconds"`
	Retry    int    `short:"r" long:"retry" description:"Max number of attempts for DHCP clients to send requests. -1 means infinite"`
	ipv4     bool   `long:"ipv4" description:"use IPV4"`
	ipv6     bool   `long:"ipv6" description:"use IPV6"`
	V6Port   int    `long:"v6-port" description:"DHCPv6 server port to send to"`
	V4Port   int    `long:"v4-port" description:"DHCPv4 server port to send to"`
	V6Server string `long:"v6-server" description:"DHCPv6 server address to send to (multicast or unicast)"`
	DryRun   bool   `short:"d" long:"dry-run" description:"Just make the DHCP requests but dont configure interfaces"`
	Verbose  []bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}
var ifname = "^e.*"

func Dhclient(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("only one re")
	}

	if len(args) > 0 {
		ifname = args[0]
	}

	filteredInterfaces, err := dhclient.Interfaces(ifname)
	if err != nil {
		return err
	}

	configureAll(filteredInterfaces)
	return nil
}

func configureAll(ifs []netlink.Link) {
	packetTimeout := time.Duration(opts.Timeout) * time.Second

	c := dhclient.Config{
		Timeout: packetTimeout,
		Retries: opts.Retry,
		V4ServerAddr: &net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: opts.V4Port,
		},
		V6ServerAddr: &net.UDPAddr{
			IP:   net.ParseIP(opts.V6Server),
			Port: opts.V6Port,
		},
	}
	if len(opts.Verbose) == 1 {
		c.LogLevel = dhclient.LogSummary
	}
	if len(opts.Verbose) > 1 {
		c.LogLevel = dhclient.LogDebug
	}
	r := dhclient.SendRequests(context.Background(), ifs, opts.ipv4, opts.ipv6, c, 30*time.Second)

	for result := range r {
		if result.Err != nil {
			log.Printf("Could not configure %s for %s: %v", result.Interface.Attrs().Name, result.Protocol, result.Err)
		} else if opts.DryRun {
			log.Printf("Dry run: would have configured %s with %s", result.Interface.Attrs().Name, result.Lease)
		} else if err := result.Lease.Configure(); err != nil {
			log.Printf("Could not configure %s for %s: %v", result.Interface.Attrs().Name, result.Protocol, err)
		} else {
			log.Printf("Configured %s with %s", result.Interface.Attrs().Name, result.Lease)
		}
	}
	log.Printf("Finished trying to configure all interfaces.")
}

func init() {
	opts.V6Port = dhcpv6.DefaultServerPort
	opts.V4Port = dhcpv4.ServerPort
	opts.Timeout = 15
	opts.Retry = 5
}

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if len(opts.Verbose) > 0 {
		Debug = log.Printf
	}

	if err := Dhclient(args); err != nil {
		log.Fatal(err)
	}
}
