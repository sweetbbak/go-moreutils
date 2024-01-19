package internal

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type runState uint8

// Types of valid runState
const (
	notStarted = runState(iota)
	starting
	started
	stopped
	errored
)

func (rs runState) String() string {
	switch rs {
	case notStarted:
		return "not started"
	case starting:
		return "being started"
	case started:
		return "already started"
	case stopped:
		return "stopped"
	case errored:
		return "errored"
	default:
		return "in an invalid state"
	}
}

type ServiceName string
type ServiceType string

type Command string

func (c Command) String() string {
	return string(c)
}

func parseLine(line string, s *Service) error {
	if strings.HasPrefix(line, "Needs:") {
		specified := strings.Split(strings.TrimPrefix(line, "Needs:"), ",")
		for _, nd := range specified {
			s.Needs = append(s.Needs, ServiceType(strings.TrimSpace(nd)))
		}
	} else if strings.HasPrefix(line, "Provides:") {
		specified := strings.Split(strings.TrimPrefix(line, "Provides:"), ",")
		for _, nd := range specified {
			s.Provides = append(s.Provides, ServiceType(strings.TrimSpace(nd)))
		}
	} else if strings.HasPrefix(line, "Startup:") {
		if s.Startup != "" {
			return fmt.Errorf("startup already set")
		}
		s.Startup = Command(strings.TrimSpace(strings.TrimPrefix(line, "Startup:")))
	} else if strings.HasPrefix(line, "Shutdown:") {
		if s.Shutdown != "" {
			return fmt.Errorf("shutdown already set")
		}
		s.Shutdown = Command(strings.TrimSpace(strings.TrimPrefix(line, "Shutdown:")))
	} else if strings.HasPrefix(line, "# ") {
		if s.Name == "" {
			s.Name = ServiceName(strings.TrimSpace(strings.TrimPrefix(line, "# ")))
		}
	}
	return nil
}

// Parses a single config file into the services it provides
func ParseConfig(r io.Reader) (Service, error) {
	s := Service{}
	var line string
	var err error
	scanner := bufio.NewReader(r)

	for {
		line, err = scanner.ReadString('\n')
		switch err {
		case io.EOF:
			if err := parseLine(line, &s); err != nil {
				log.Println(err)
			}
			return s, nil
		case nil:
			if err := parseLine(line, &s); err != nil {
				log.Println(err)
			}
		default:
			return Service{}, err
		}
	}
}

// Parses all the config in directory dir return a map of
// providers of ServiceTypes from that directory.
func ParseServiceConfigs(dir string) (map[ServiceType][]*Service, error) {
	providers := make(map[ServiceType][]*Service)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, fstat := range files {
		if fstat.IsDir() {
			// Mostly to skip "." and ".."
			continue
		}
		f, err := os.Open(dir + "/" + fstat.Name())
		if err != nil {
			log.Println(err)
			continue
		}
		s, err := ParseConfig(f)
		f.Close()
		if err != nil {
			log.Println(err)
			continue
		}
		for _, t := range s.Provides {
			providers[t] = append(providers[t], &s)
		}

	}
	return providers, nil
}

type Service struct {
	Name     ServiceName
	Startup  Command
	Shutdown Command
	Provides []ServiceType
	Needs    []ServiceType

	state runState
}

func StartServices(services map[ServiceType][]*Service) {
	wg := sync.WaitGroup{}

	var startedMu *sync.RWMutex = &sync.RWMutex{}
	startedTypes := make(map[ServiceType]bool)
	for _, services := range services {
		wg.Add(len(services))
		for _, s := range services {
			go func(s *Service) {
				// TODO: This should ensure that Needs are satisfiable instead of getting into an
				// infinite loop when they're not.
				// (TODO(2): Prove N=NP in order to do the above efficiently.)
				for satisfied, tries := false, 0; !satisfied && tries < 60; tries++ {
					satisfied = s.NeedsSatisfied(startedTypes, startedMu)
					time.Sleep(2 * time.Second)

				}
				if s.state == notStarted {
					if err := s.Start(); err != nil {
						log.Println(err)
					}

				}

				startedMu.Lock()
				for _, t := range s.Provides {
					startedTypes[t] = true
				}
				startedMu.Unlock()
				wg.Done()
			}(s)
		}
	}
	wg.Wait()
}

// Starts the Service(s)
func (s *Service) Start() error {
	if s.state != notStarted {
		return fmt.Errorf("Service %v is %v", s.Name, s.state.String())
	}
	s.state = starting
	cmd := exec.Command("/bin/sh", "-c", s.Startup.String())
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		s.state = errored
		return err
	}
	s.state = started
	return nil
}

func (s *Service) Stop() error {
	if s.state == notStarted {
		return fmt.Errorf("Service %v is %v", s.Name, s.state.String())
	}

	cmd := exec.Command("/bin/sh", "-c", s.Shutdown.String())
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		s.state = errored
		return err
	}
	s.state = stopped
	return nil
}

// Checks if all of s's needs are satified by the passed list of provided types
func (s Service) NeedsSatisfied(started map[ServiceType]bool, mu *sync.RWMutex) bool {
	mu.RLock()
	defer mu.RUnlock()
	for _, st := range s.Needs {
		if !started[st] {
			return false
		}
	}
	return true
}
