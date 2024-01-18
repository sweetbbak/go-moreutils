package main

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/jessevdk/go-flags"
)

var opt struct {
	All  bool `short:"a" long:"all" description:"Show group name and number"`
	Name bool `short:"n" long:"name" description:"show group names only"`
	IDs  bool `short:"i" long:"id" description:"show group IDs only"`
}

func legacyGroups(gids []string) error {
	groups := fmt.Sprintf("")
	for _, num := range gids {
		f, err := user.LookupGroupId(num)
		if err != nil {
		}
		groups += fmt.Sprintf("%s ", f.Name)
	}
	fmt.Println(groups)
	return nil
}

func IDs(gids []string) error {
	ids := fmt.Sprintf("")
	for _, num := range gids {
		ids += fmt.Sprintf("%s ", num)
	}
	fmt.Println(ids)
	return nil
}

func Groups(usr *user.User, args []string) error {
	gids, err := usr.GroupIds()
	if err != nil {
		return err
	}

	if !opt.All && !opt.IDs || opt.Name {
		return legacyGroups(gids)
	}

	if opt.IDs {
		return IDs(gids)
	}

	if opt.All {
		legacyGroups(gids)
		IDs(gids)
	}

	return nil
}

func main() {
	args, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(0)
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	if err := Groups(usr, args); err != nil {
		log.Fatal(err)
	}
}
