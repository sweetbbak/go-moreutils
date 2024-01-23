package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	EGID    bool `short:"g" long:"group" description:"print the effective Group ID for the user"`
	Groups  bool `short:"G" long:"groups" description:"print all Group IDs"`
	Name    bool `short:"n" long:"name" description:"print a name instead of an ID number"`
	User    bool `short:"u" long:"user" description:"print the effeective User ID"`
	Real    bool `short:"r" long:"real" description:"print the real ID instead of the effective ID"`
	Zero    bool `short:"z" long:"zero" description:"delimit entries with NUL characters, not whitespace"`
	Verbose bool `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}
var DELIM = " "

const (
	passwdFile = "/etc/passwd"
	groupFile  = "/etc/group"
)

func isNum(str string) bool {
	for _, i := range str {
		if i >= 0 && i <= 9 {
			continue
		} else {
			return false
		}
	}
	return true
}

func ID(name string, current bool) error {
	var usr *user.User
	var err error

	if current {
		usr, err = user.Current()
		if err != nil {
			return err
		}
	} else {
		// look up user ID and usernames
		if isNum(name) {
			usr, err = user.LookupId(name)
			if err != nil {
				return err
			}
		} else {
			usr, err = user.Lookup(name)
			if err != nil {
				return err
			}
		}
	}

	if opts.User {
		if opts.Name {
			fmt.Println(usr.Username)
			return nil
		}

		// idk how to sus this out and id -ur VS id -u is the same always
		if opts.Real {
			fmt.Println(usr.Uid)
			return nil
		} else {
			fmt.Println(usr.Uid)
			return nil
		}
	}

	if opts.EGID {
		if opts.Name {
			g, err := user.LookupGroupId(usr.Gid)
			if err != nil {
				return err
			}

			fmt.Println(g)
			return nil
		} else {
			fmt.Println(usr.Gid)
			return nil
		}
	}

	if opts.Groups {
		ids, err := usr.GroupIds()
		if err != nil {
			return err
		}

		if opts.Name {
			return printGID(ids)
		} else {
			str := strings.Join(ids, DELIM)
			if !opts.Zero {
				str += "\n"
			}

			fmt.Print(str)
			return nil
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("uid=%s(%s)", usr.Uid, usr.Username))
	sb.WriteString(" ")

	gid, err := user.LookupGroupId(usr.Gid)
	if err != nil {
		return err
	}

	sb.WriteString(fmt.Sprintf("gid=%s(%s)", usr.Gid, gid.Name))
	sb.WriteString(" ")

	ids, err := usr.GroupIds()
	if err != nil {
		return err
	}

	sb.WriteString("groups=")

	stop := len(ids) - 1

	for i, id := range ids {
		grp, err := user.LookupGroupId(id)
		if err != nil {
			return err
		}
		sb.WriteString(fmt.Sprintf("%s(%s)", grp.Gid, grp.Name))
		if i != stop {
			sb.WriteString(",")
		} else {
			sb.WriteString("\n")
		}
	}

	fmt.Print(sb.String())

	return nil
}

func printGID(ids []string) error {
	var sb strings.Builder
	for _, g := range ids {
		grp, err := user.LookupGroupId(g)
		if err != nil {
			return err
		}
		sb.WriteString(grp.Name)
		sb.WriteString(DELIM)
	}

	if !opts.Zero {
		sb.WriteString("\n")
	}

	fmt.Print(sb.String())
	sb.Reset()
	return nil
}

func verifyOptions() error {
	if opts.Groups && opts.EGID {
		return fmt.Errorf("cannot specify both group options")
	}
	if opts.User && opts.EGID || opts.User && opts.Groups {
		return fmt.Errorf("cannot specify both group options")
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

	if err := verifyOptions(); err != nil {
		log.Fatal(err)
	}

	if opts.Zero {
		DELIM = "\x00"
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if len(args) == 0 {
		if err := ID("", true); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	for _, name := range args {
		if err := ID(name, false); err != nil {
			log.Fatal(err)
		}
	}
}
