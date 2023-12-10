package main

import (
	"bufio"
	"fmt"
	"os"
)

var lines = 10
var (
	wscol int
	wsrow int
)

func navigate() {
	a, _, err := getChar()
	if err != nil {
		fmt.Println(err)
	}

	switch a {
	case int('j'):
	case int('k'):
	}
}

func More(args []string) error {
	oldState, err := makeRaw(os.Stdin.Fd())
	if err != nil {
		return err
	}
	defer restoreTerminal(os.Stdin.Fd(), oldState)
	fmt.Print("\x1b[H\x1b[2J") // clear screen

	stdout := os.Stdout
	for _, file := range args {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for i := 0; scanner.Scan(); i++ {
			// our screen buffer size
			if (i+1)%lines == 0 {
				fmt.Fprint(stdout, scanner.Text())
				c := make([]byte, 1)
				// We expect the OS to echo the newline character.
				if _, err := os.Stdin.Read(c); err != nil {
					return err
				}
			} else {
				fmt.Fprintln(stdout, scanner.Text())
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	return nil
}

// PercentOf - calculate what percent [number1] is of [number2].
// ex. 300 is 12.5% of 2400
func PercentOf(part int, total int) float64 {
	return (float64(part) * float64(100)) / float64(total)
}

func PercentageChange(old, new int) (delta float64) {
	diff := float64(new)
	delta = (diff / float64(old)) * 100
	return
}

func More2(args []string) error {
	oldState, err := makeRaw(os.Stdin.Fd())
	if err != nil {
		return err
	}

	defer restoreTerminal(os.Stdin.Fd(), oldState)
	fmt.Print("\x1b[H\x1b[2J") // clear screen

	// stdout := os.Stdout
	for _, file := range args {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		// alt screen
		fmt.Printf("\x1b[?1049h")

		// frameBuf := lines + 10
		buf := []string{}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			buf = append(buf, line)
		}

		n := 0
		a := 100
		for a != int('q') {
			if n+lines < len(buf) {
				for _, item := range buf[n : n+lines] {
					fmt.Println(item)
				}
			} else {
				for _, item := range buf[n:] {
					fmt.Println(item)
				}
			}
			pc := PercentOf(n, len(buf))
			status := fmt.Sprintf("%0.1f -- %s   [q/j/k]", pc, file)
			// fmt.Printf("\x1b[7m\x1b[1;%vH -- %s\x1b[0m\n", lines+2, status)
			fmt.Printf("\x1b[7m \x1b[%v;1H -- %s \x1b[0m", lines+2, status)

			a, _, _ = getChar()
			switch a {
			case int('j'):
				if n < len(buf)-1 {
					n++
				}
			case int('k'):
				if n > 0 {
					n--
				}
			case int('q'), 27:
				break
			default:
				continue
			}

			fmt.Print("\x1b[H\x1b[2J") // clear screen
		}

	}
	return nil
}

func main() {
	wscol, wsrow = get_term_size(os.Stdin.Fd())
	lines = wsrow - 1
	args := os.Args[1:]
	More2(args)
}
