package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

const (
	defaultBind               = ":22"
	defaultAuthorizedKeysFile = "/root/.ssh/authorized_keys"
	defaultHostKeyFile        = "/etc/ssh/host_rsa"
	defaultEnvPath            = "/sbin:/usr/sbin:/bin:/usr/bin"
	defaultEnvHome            = "/root"
	defaultEnvTerm            = "linux"
	defaultEntrypoint         = "/bin/sh"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

type option func(*server) error

func WithHostKeyFile(fn string) option {
	return func(s *server) error {
		if !fileExists(fn) {
			if err := generateHostKey(fn); err != nil {
				return err
			}
		}

		s.hostKeyFile = fn

		return nil
	}
}

type server struct {
	bind           string
	authorizedkeys []ssh.PublicKey
	hostKeyFile    string
}

func NewSSHServer(bind, keys string, opts ...option) (*server, error) {
	var authorizedkeys []ssh.PublicKey

	if strings.HasPrefix(keys, "file://") || fileExists(keys) {
		keys = strings.TrimPrefix(keys, "file://")
		data, err := os.ReadFile(keys)
		if err != nil {
			return nil, fmt.Errorf("error reading keys file: %w", err)
		}

		for _, key := range bytes.Split(data, []byte("\n")) {
			publickey, _, _, _, _ := ssh.ParseAuthorizedKey(key)
			authorizedkeys = append(authorizedkeys, publickey)
		}
	}

	s := &server{
		bind:           bind,
		authorizedkeys: authorizedkeys,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	return s, nil
}

func (s *server) sessionHandler(sess ssh.Session) {
	stime := time.Now()
	log.Printf(
		"New session @%s %s (%d)",
		sess.User(), sess.RemoteAddr(), stime.Unix(),
	)
	defer func() {
		etime := time.Now()
		dtime := etime.Sub(stime)
		log.Printf(
			"Session ended @%s %s (%d) [%s]",
			sess.User(), sess.RemoteAddr(), etime.Unix(), dtime,
		)
	}()

	var (
		entrypoint string
		args       []string
	)

	log.Printf("User wants to execute: %v", sess.Command())

	if len(sess.Command()) > 0 {
		entrypoint = sess.Command()[0]
		if !filepath.IsAbs(entrypoint) {
			path, err := exec.LookPath(entrypoint)
			if err != nil {
				log.Printf("error looking up entrypoint %q: %s", entrypoint, err)
				sess.Exit(1)
				return
			}
			entrypoint = path
		}
		args = sess.Command()[1:]
	} else {
		entrypoint = defaultEntrypoint
	}

	log.Printf("Executing cmd=%s args=%v", entrypoint, args)

	cmd := exec.Command(entrypoint, args...)

	cmd.Dir = defaultEnvHome
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("PATH=%s", defaultEnvPath),
		fmt.Sprintf("HOME=%s", defaultEnvHome),
		fmt.Sprintf("TERM=%s", defaultEnvTerm),
	}...)

	ptyReq, winCh, isPty := sess.Pty()
	if isPty {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))

		f, err := pty.StartWithAttrs(
			cmd,
			&pty.Winsize{
				Rows: uint16(ptyReq.Window.Height),
				Cols: uint16(ptyReq.Window.Width),
			},
			&syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 1},
		)
		if err != nil {
			log.Printf("error executing command %q: %s", cmd.String(), err.Error())
			sess.Exit(1)
			return
		}
		go func() {
			for win := range winCh {
				setWinsize(f, win.Width, win.Height)
			}
		}()
		go io.Copy(f, sess) // stdin
		go io.Copy(sess, f) // stdout
	} else {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Printf("error creating stdin pipe: %s", err.Error())
			sess.Exit(1)
			return
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("error creating stdout pipe: %s", err.Error())
			sess.Exit(1)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Printf("error creating stderr pipe: %s", err.Error())
			sess.Exit(1)
			return
		}

		go func() {
			defer stdin.Close()
			io.Copy(stdin, sess) // stdin
		}()

		go io.Copy(sess, stdout)          // stdout
		go io.Copy(sess.Stderr(), stderr) // stderr

		if err := cmd.Start(); err != nil {
			log.Printf("error executing command %q: %s", cmd.String(), err.Error())
			sess.Exit(1)
			return
		}
	}

	if err := cmd.Wait(); err == nil {
		sess.Exit(0)
	} else {
		if exitError, ok := err.(*exec.ExitError); ok {
			sess.Exit(exitError.ExitCode())
			return
		}
		log.Printf("error waiting on command %q: %s", cmd.String(), err.Error())
		sess.Exit(255)
	}
}

func (s *server) Shutdown() (err error) {
	return
}

func (s *server) Run(ctx context.Context) (err error) {
	sshServer := &ssh.Server{
		Addr:    s.bind,
		Handler: s.sessionHandler,
	}

	sshServer.SetOption(ssh.HostKeyFile(s.hostKeyFile))
	sshServer.SetOption(
		ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			user := ctx.User()

			for _, publickey := range s.authorizedkeys {
				if ssh.KeysEqual(key, publickey) {
					log.Printf("User %s authorized", user)
					return true
				}
			}
			log.Printf("User %s denied", user)
			return false
		}),
	)

	go func() {
		<-ctx.Done()
		log.Printf("Shutdown server")
		sshServer.Close()
	}()

	if err := sshServer.ListenAndServe(); err != nil {
		return s.Shutdown()
	}
	return
}

func StartSSHServer(ctx context.Context) {
	svr, err := NewSSHServer(
		defaultBind, defaultAuthorizedKeysFile,
		WithHostKeyFile(defaultHostKeyFile),
	)
	if err != nil {
		log.Panicln(err.Error())
		return
	}
	if err := svr.Run(ctx); err != nil {
		log.Printf("error running ssh server: %s\n", err)
	}
}
