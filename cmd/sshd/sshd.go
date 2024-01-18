package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/jessevdk/go-flags"
	"github.com/u-root/u-root/pkg/pty"
	"golang.org/x/crypto/ssh"
)

var opts struct {
	Keys       string `short:"k" long:"keys" description:"Path to the authorized keys file"`
	PrivateKey string `short:"P" long:"private-key" default:"id_rsa" description:"Name of private key file"`
	IP         string `short:"i" long:"ip-address" default:"0.0.0.0" description:"IP address to listen on"`
	Port       string `short:"p" long:"port" default:"2022" description:"Port to listen on"`
	Verbose    bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

type ptyReq struct {
	TERM   string // env variable
	Col    uint32
	Row    uint32
	Xpixel uint32
	Ypixel uint32
	Modes  string // terminal modes
}

type execReq struct {
	Command string
}

type exitStatusReq struct {
	ExitStatus uint32
}

var shells = []string{"sh", "bash", "zsh", "dash", "fish", "elvish"}
var shell string

func readKeyFile(keyfile string) ([]byte, error) {
	authBytes, err := os.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}
	return authBytes, nil
}

func newPTY(b []byte) (*pty.Pty, error) {
	ptyReq := &ptyReq{}
	err := ssh.Unmarshal(b, ptyReq)
	Debug("New PTY: %q", ptyReq)
	if err != nil {
		return nil, err
	}
	p, err := pty.New()
	if err != nil {
		return nil, err
	}
	ws, err := p.TTY.GetWinSize()
	if err != nil {
		return nil, err
	}
	ws.Row = uint16(ptyReq.Row)
	ws.Ypixel = uint16(ptyReq.Ypixel)
	ws.Col = uint16(ptyReq.Col)
	ws.Xpixel = uint16(ptyReq.Xpixel)
	Debug("New PTY: Set window sizes to: %v", ws)
	if err := p.TTY.SetWinSize(ws); err != nil {
		return nil, err
	}
	Debug("New PTY: Set TERM to: %v", ptyReq.TERM)
	if err := os.Setenv("TERM", ptyReq.TERM); err != nil {
		return nil, err
	}
	return p, nil
}

func init() {
	for _, s := range shells {
		if _, err := exec.LookPath(s); err == nil {
			shell = s
		}
	}
}

func runCommand(c ssh.Channel, p *pty.Pty, cmd string, args ...string) error {
	var ps *os.ProcessState
	defer c.Close()

	if p != nil {
		log.Printf("Executing PTY command %s %v", cmd, args)
		p.Command(cmd, args...)
		if err := p.C.Start(); err != nil {
			Debug("Failed to execute: %v", err)
			return err
		}

		defer p.C.Wait()
		go io.Copy(p.Ptm, c)
		go io.Copy(c, p.Ptm)
		ps, _ = p.C.Process.Wait()
	} else {
		e := exec.Command(cmd, args...)
		e.Stdin, e.Stdout, e.Stderr = c, c, c
		log.Printf("Executing non-PTY command %s %v", cmd, args)
		if err := e.Start(); err != nil {
			Debug("Failed to execute: %v", err)
			return err
		}
		ps, _ = e.Process.Wait()
	}
	if ps.Exited() {
		code := uint32(ps.ExitCode())
		Debug("Exit status: %v\n", code)
		c.SendRequest("exit-status", false, ssh.Marshal(exitStatusReq{code}))
	}
	return nil
}

func session(chans <-chan ssh.NewChannel) {
	var p *pty.Pty
	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unkown channel type")
			continue
		}

		channel, requests, err := newChan.Accept()
		if err != nil {
			log.Printf("Could not accept channel: %v", err)
			continue
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				Debug("%v\n", req.Type)
				switch req.Type {
				case "shell":
					err := runCommand(channel, p, shell)
					req.Reply(true, []byte(fmt.Sprintf("%v", err)))
				case "exec":
					e := &execReq{}
					if err := ssh.Unmarshal(req.Payload, e); err != nil {
						log.Printf("sshd: %v", err)
						break
					}
					err := runCommand(channel, p, shell, "-c", e.Command)
					req.Reply(true, []byte(fmt.Sprintf("%v", err)))
				case "pty-req":
					p, err = newPTY(req.Payload)
					req.Reply(err == nil, nil)
				default:
					log.Printf("Not handling request %v %q", req, string(req.Payload))
					req.Reply(false, nil)
				}
			}
		}(requests)

	}
}

func Sshd(args []string) error {
	authBytes, err := readKeyFile(opts.Keys)
	if err != nil {
		return err
	}

	authMap := map[string]bool{}
	for len(authBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authBytes)
		if err != nil {
			return err
		}

		Debug("%v\n", string(pubKey.Marshal()))
		authMap[string(pubKey.Marshal())] = true
		authBytes = rest
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if authMap[string(key.Marshal())] {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(key),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", conn.User())
		},
	}

	privateBytes, err := os.ReadFile(opts.PrivateKey)
	if err != nil {
		return err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return err
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", net.JoinHostPort(opts.IP, opts.Port))
	if err != nil {
		return err
	}

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept incoming connection: %s", err)
			continue
		}

		conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			log.Printf("failed to handshake: %s", err)
			continue
		}

		log.Printf("%v logged in with key %s", conn.RemoteAddr(), conn.Permissions.Extensions["pubkey-fp"])

		go ssh.DiscardRequests(reqs)
		go session(chans)
	}
}

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Sshd(args); err != nil {
		log.Fatal(err)
	}
}
