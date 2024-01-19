package internal

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/digineo/go-dhclient"
	"github.com/google/gopacket/layers"
	"github.com/vishvananda/netlink"
)

func DHCPClient(ifname string) (*dhclient.Client, error) {
	// Setup interface to receive DHCP traffic
	iface, err := net.InterfaceByName(ifname)
	if err != nil {
		return nil, fmt.Errorf("error finding interface by name %q: %w", ifname, err)
	}

	client := &dhclient.Client{
		Iface: iface,
		OnBound: func(lease *dhclient.Lease) {
			ip := &net.IPNet{IP: lease.FixedAddress, Mask: lease.Netmask}
			log.Printf(
				"dhcp lease server=%s expires=%s ip=%s domain=%s resolvers=%s routers=%s\n",
				lease.ServerID, time.Until(lease.Expire), ip.String(),
				lease.DomainName, lease.DNS, lease.Router,
			)

			link, err := netlink.LinkByName(iface.Name)
			if err != nil {
				log.Printf("error getting link by name %q: %s", iface.Name, err)
				return
			}

			// Set address / netmask into cidr we can use to apply to interface
			cidr := net.IPNet{
				IP:   lease.FixedAddress,
				Mask: lease.Netmask,
			}
			addr, err := netlink.ParseAddr(cidr.String())
			if err != nil {
				log.Printf("error parsing address %q: %s", cidr.String(), err)
				return
			}

			if err := netlink.AddrAdd(link, addr); err != nil {
				log.Printf("error adding %s to link %s", cidr.String(), iface.Name)
			}

			var gw net.IP

			if len(lease.Router) > 0 {
				gw = lease.Router[0]
			} else {
				gw = lease.ServerID
			}

			// Apply default gateway so we can route outside
			route := netlink.Route{
				Scope: netlink.SCOPE_UNIVERSE,
				Gw:    gw,
			}
			if err := netlink.RouteAdd(&route); err != nil {
				log.Printf("error setting gateway [%v]", err)
			}
		},
	}

	// Add requests for default options
	for _, param := range dhclient.DefaultParamsRequestList {
		client.AddParamRequest(layers.DHCPOpt(param))
	}

	// Add hostname option
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("error getting hostname: %w", err)
	}
	client.AddOption(layers.DHCPOptHostname, []byte(hostname))

	client.Start()

	return client, nil
}

func InterfaceUp(ifname string) error {
	ifaceDev, err := netlink.LinkByName(ifname)
	if err != nil {
		return fmt.Errorf("error finding interface %q: %w", ifname, err)
	}

	if err := netlink.LinkSetUp(ifaceDev); err != nil {
		return fmt.Errorf("error bringing up interface %q: %w", ifname, err)
	}

	return nil
}

func SetupNetworkInterfaces() error {
	if err := InterfaceUp("lo"); err != nil {
		return err
	}

	if err := InterfaceUp("eth0"); err != nil {
		return err
	}

	return nil
}
