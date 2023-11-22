package main

import (
	"log"

	"golang.org/x/sys/unix"
)

func main() {
	if err := unix.Reboot(unix.LINUX_REBOOT_CMD_POWER_OFF); err != nil {
		log.Fatal(err)
	}
}
