package main

import (
	"flag"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

var (
	Ypos int
)

func wrapLines(text string, width int) []string {
	output := []string{}
	buffer := ""
	for _, letter := range text {
		if letter == '\n' {
			output = append(output, buffer)
			buffer = ""
			continue
		}
		buffer += string(letter)
		if len(buffer) >= width {
			output = append(output, buffer)
			buffer = ""
		}
	}
	if len(buffer) != 0 {
		output = append(output, buffer)
	}
	return output
}

func Less(args []string) error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	if err := screen.Init(); err != nil {
		return err
	}

	width, height := screen.Size()
	screen.ShowCursor(1, height-1)

	fileName := os.Stdin.Name()
	if len(args) >= 1 {
		fileName = args[0]
	}

	filedata, err := os.ReadFile(fileName)
	if err != nil {
		screen.Fini()
		return err
	}

	lines := wrapLines(string(filedata), width)

	for {
		screen.Clear()
		breakOut := false
		width, height = screen.Size()
		if Ypos >= len(lines) {
			Ypos = len(lines) - 1
		} else if Ypos < 0 {
			Ypos = 0
		}

		currentLines := lines[Ypos:]
		for index, line := range currentLines {
			for x := 0; x < len(line); x++ {
				if index >= height-1 {
					breakOut = true
					break
				}
				screen.SetContent(x, index, []rune(string(line[x]))[0], nil, tcell.StyleDefault)
			}
			if breakOut {
				break
			}
		}
		screen.SetContent(0, height-1, ':', nil, tcell.StyleDefault)
		screen.Show()

		switch event := screen.PollEvent().(type) {
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventKey:
			if event.Key() == tcell.KeyEscape {
				screen.Fini()
				return nil
			} else if event.Key() == tcell.KeyRune {
				switch event.Rune() {
				case 'q':
					screen.Fini()
					return nil
				case 'g':
					Ypos = 0
				case 'G':
					Ypos = len(lines) - 1
				case 'k':
					Ypos -= 1
				case 'j':
					Ypos += 1
				case ' ':
					Ypos += height
				}
			} else if event.Key() == tcell.KeyUp {
				Ypos -= 1
			} else if event.Key() == tcell.KeyDown {
				Ypos += 1
			} else if event.Key() == tcell.KeyPgUp {
				Ypos -= height
			} else if event.Key() == tcell.KeyPgDn {
				Ypos += height
			}
		}
	}
}

func main() {
	flag.Parse()
	args := flag.Args()

	if err := Less(args); err != nil {
		log.Fatal(err)
	}
}
