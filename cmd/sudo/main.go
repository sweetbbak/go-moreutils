package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"syscall"
	"time"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"golang.org/x/sys/unix"
)

var (
	STDINFILENO int = 0
)

func timeout() bool {
	fi, err := os.Stat("/proc/self/exe")
	if err != nil {
		fmt.Println(err)
	}

	file_time := fi.ModTime()
	t2 := file_time.Add(time.Minute * 5)
	now := time.Now()

	if now.After(t2) {
		// ask for pass
		return true
	} else {
		return false
	}
}

func reset_modtime() {
	// os.Chtimes("/proc/self/exe", time.Now(), time.Now())
}

func restore(raw *unix.Termios) {
	err := unix.IoctlSetTermios(STDINFILENO, unix.TCSETS, raw)
	if err != nil {
		panic(err)
	}
}

func askpass() (string, error) {
	// turn off terminal echo
	raw, err := unix.IoctlGetTermios(STDINFILENO, unix.TCGETS)
	if err != nil {
		return "", err
	}

	rawState := *raw
	rawState.Lflag &^= unix.ECHO
	rawState.Lflag &^= unix.ICANON
	rawState.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	rawState.Oflag &^= unix.OPOST
	rawState.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	rawState.Cflag &^= unix.CSIZE | unix.PARENB
	rawState.Cflag |= unix.CS8
	rawState.Cc[unix.VMIN] = 1
	rawState.Cc[unix.VTIME] = 0

	err = unix.IoctlSetTermios(STDINFILENO, unix.TCSETS, &rawState)
	if err != nil {
		return "", err
	}
	var password string
	var prompt string

	// prompt = "\x1b[33m[\x1b[0m\x1b[32msuwudo\x1b[0m\x1b[33m]\x1b[0m password for %s: "
	prompt = "\x1b[38;2;111;111;111m│ \x1b[0m\x1b[38;2;124;120;254mpassword for \x1b[38;2;245;127;224m\x1b[3m%s\x1b[0m \x1b[38;2;124;120;254m>\x1b[0m "
	// prompt = "%s > "
	user := os.Getenv("USER")
	if user == "" {
		user = "user"
	}

	fmt.Fprintf(os.Stderr, "\x1b[2K")
	fmt.Fprintf(os.Stderr, "\x1b[0G")
	fmt.Fprintf(os.Stderr, prompt, user)
	fmt.Fscanf(os.Stdout, "%s", &password)

	// erase line
	fmt.Fprintf(os.Stderr, "\x1b[2K")
	fmt.Fprintf(os.Stderr, "\x1b[0G")
	defer restore(raw)
	fmt.Fprintf(os.Stderr, "\x1b[2K")
	return password, nil
}

func get_user() string {
	var name string
	uid := os.Geteuid()
	fmt.Println(uid)
	// open pass file and read the user name from it by matching the UID
	// sweet:x:1000:1000:sweet:/home/sweet:/bin/zsh - is what it looks like
	file, err := os.Open("/etc/passwd")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), fmt.Sprintf("%d", uid)) {
			name = scanner.Text()
		}
	}

	// match UID with the line and get the first field which is the users name
	splits := strings.Split(name, ":")
	name = splits[0]
	return name
}

func verify_pass(password string, uid int) (bool, error) {
	var token string
	UID := fmt.Sprintf("%d", uid)
	name, err := user.LookupId(UID)
	if err != nil {
		return false, err
	}

	// open etc shadow and find the users hash - name:$6$reallylonghash:12345:0:99999:7:::
	fi, err := os.Open("/etc/shadow")
	if err != nil {
		fmt.Println(err)
	}

	defer fi.Close()
	scanner := bufio.NewScanner(fi)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), name.Name) {
			token = scanner.Text()
		}
	}

	// hash/token is the 2nd field with a ":" delimiter
	split := strings.Split(token, ":")
	token = split[1]

	// check the hash against the password with this convenient go package
	// if err returns, pass is incorrect. simple shit.
	crypt := crypt.SHA512.New()
	err = crypt.Verify(token, []byte(password))
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func main() {
	userID := syscall.Getuid()

	// get effective user ID and set to root user
	err := syscall.Setuid(0)
	if err != nil {
		fmt.Println("Error setting user as root")
		os.Exit(1)
	}

	pass, err := askpass()
	if err != nil {
		fmt.Println(err)
	}

	passed, err := verify_pass(pass, userID)
	if err != nil {
		log.Fatal(err)
	}

	if !passed {
		fmt.Fprintln(os.Stderr, "Incorrect password")
		os.Exit(1)
	}

	cmd := strings.Join(os.Args[1:], " ")
	if cmd == "" {
		os.Exit(0)
	}

	exitCode := system(cmd)
	os.Exit(exitCode)
}

func system(cmd string) int {
	c := exec.Command("sh", "-c", cmd)
	c.Env = os.Environ()
	// c.Env = append(c.Env, env_vars)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := c.Run()
	if err == nil {
		return 0
	}

	// Figure out the exit code
	if ws, ok := c.ProcessState.Sys().(syscall.WaitStatus); ok {
		if ws.Exited() {
			return ws.ExitStatus()
		}

		if ws.Signaled() {
			return -int(ws.Signal())
		}
	}
	return -1
}
