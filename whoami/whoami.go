package main

import (
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"os"
	"os/user"
)

var opts struct {
	All    bool `short:"a" long:"all" description:"show group"`
	Pretty bool `short:"p" long:"pretty" description:"pretty print whoami"`
	Group  bool `short:"g" long:"group" description:"show group"`
	GID    bool `short:"G" long:"gid" description:"show group"`
	UID    bool `short:"U" long:"uid" description:"show group"`
	User   bool `short:"u" long:"user" description:"show group"`
	Home   bool `short:"H" long:"home" description:"show group"`
}

type users struct {
	Username string
	GID      string
	UID      string
	Homedir  string
	groups   []string
}

func (u *users) Print() {
	if opts.Pretty {

	}

	if opts.All {
		opts.User = true
		opts.GID = true
		opts.UID = true
		opts.Home = true
		opts.Group = true
	}

	if opts.User {
		if opts.Pretty {
			fmt.Printf("\x1b[35mUser\x1b[0m: \x1b[33m%s\x1b[0m\n", u.Username)
		} else {
			fmt.Printf("User: %s\n", u.Username)
		}
	}

	if opts.GID {
		if opts.Pretty {
			fmt.Printf("\x1b[35mGID\x1b[0m: \x1b[33m%s\x1b[0m\n", u.GID)
		} else {
			fmt.Printf("GID: %s\n", u.GID)
		}
	}

	if opts.UID {
		if opts.Pretty {
			fmt.Printf("\x1b[35mUID\x1b[0m: \x1b[33m%s\x1b[0m\n", u.UID)
		} else {
			fmt.Printf("UID: %s\n", u.UID)
		}
	}

	if opts.Home {
		if opts.Pretty {
			fmt.Printf("\x1b[35mHome\x1b[0m: \x1b[33m%s\x1b[0m\n", u.Homedir)
		} else {
			fmt.Printf("Home: %s\n", u.Homedir)
		}
	}

	if opts.Group {
		str := ""
		for x := range u.groups {
			f := u.groups[x]
			str += fmt.Sprintf("%s ", f)
		}
		if opts.Pretty {
			fmt.Printf("\x1b[35mGroups\x1b[0m: \x1b[31m%s\x1b[0m\n", str)
		} else {
			fmt.Printf("Groups: %s\n", str)
		}
	}

}

func Whoami() {
	var U users

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user: ", err)
		os.Exit(1)
	}

	if len(os.Args) == 1 {
		fmt.Println(usr.Username)
		os.Exit(0)
	}

	gids, err := usr.GroupIds()
	if err != nil {
		fmt.Println(err)
	}

	U.Username = usr.Username
	U.GID = usr.Gid
	U.UID = usr.Uid
	U.Homedir = usr.HomeDir
	U.groups = gids

	U.Print()
	os.Exit(0)
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Println(err)
	}

	Whoami()
}
