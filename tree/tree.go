package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	NoColor bool     `short:"n" long:"no-color" description:"Dont output tree with color default: [auto]"`
	Mode    bool     `short:"m" long:"mode" description:"show the files mode"`
	Size    bool     `short:"s" long:"size" description:"show the files size"`
	Depth   int      `short:"l" long:"level" description:"Depth, N number of levels to traverse"`
	Ignore  []string `short:"i" long:"ignore" description:"list of directories to ignore"`
	Verbose bool     `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

const (
	esc       = "\x1b["
	clear     = esc + "0m"
	red       = "31m"
	green     = "32m"
	yellow    = "33m"
	blue      = "34m"
	magenta   = "35m"
	cyan      = "36m"
	white     = "37m"
	normal    = "0;"
	italic    = "3;"
	underline = "4;"
	reverse   = "7;"
)

func Color(s, color, style string) string {
	return fmt.Sprintf("%s%s%s%s%s", esc, style, color, s, clear)
}

func getFileType(file string, fi os.FileInfo) string {
	mode := fi.Mode()
	ext := filepath.Ext(file)
	var style string

	switch {
	case contains([]string{".bat", ".btm", ".cmd", ".com", ".exe"}, ext):
		style = Color(" "+fi.Name(), white, normal)
	case contains([]string{".dll", ".so", ".o"}, ext):
		style = Color(" "+fi.Name(), white, normal)
	case contains([]string{".arj", ".bz2", ".deb", ".gz", ".lzh", ".rpm", ".br", ".7z",
		".tar", ".taz", ".tb2", ".tbz2", ".tbz", ".tgz", ".tz", ".tz2", ".z", ".zip", ".zoo", ".xz"}, ext):
		style = Color(" "+fi.Name(), red, normal)
	case contains([]string{".asf", ".avi", ".bmp", ".flac", ".gif", ".jpg",
		"jpeg", ".m2a", ".m2v", ".mov", ".mp3", ".mpeg", ".mpg", ".ogg", ".ppm",
		".rm", ".tga", ".tif", ".wav", ".wmv", ".opus",
		".xbm", ".xpm"}, ext):
		style = Color(" "+fi.Name(), cyan, normal)
	case contains([]string{".png", ".webp", ".webm", ".jpg", ".tiff", ".gif", ".jpeg", ".svg"}, ext):
		style = Color(" "+fi.Name(), blue, normal)
	case contains([]string{".iso", ".img"}, ext):
		style = Color(" "+fi.Name(), red, normal)
	case contains([]string{".bin", ".out"}, ext):
		style = Color(" "+fi.Name(), red, normal)
	case contains([]string{".go", ".mod", ".sum"}, ext):
		style = Color(" "+fi.Name(), cyan, normal)
	case contains([]string{".py"}, ext):
		style = Color(" "+fi.Name(), yellow, normal)
	case contains([]string{".c", ".h", ".cpp", ".cc", ".c++", ".cxx"}, ext):
		style = Color(" "+fi.Name(), yellow, normal)
	case contains([]string{".sh"}, ext):
		style = Color(" "+fi.Name(), green, normal)
	case contains([]string{".txt", ".text"}, ext):
		style = Color(" "+fi.Name(), white, normal)
	case contains([]string{".json", ".yml", ".yaml", ".toml"}, ext):
		style = Color(" "+fi.Name(), green, italic)
	case contains([]string{".conf", ".ini", ".cfg", ".nix"}, ext):
		style = Color(" "+fi.Name(), cyan, italic)
	case contains([]string{".lock"}, ext):
		style = Color("  "+fi.Name(), yellow, underline)
	case contains([]string{".md"}, ext):
		style = Color("  "+fi.Name(), green, underline)
	case contains([]string{".js", ".ts"}, ext):
		style = Color("  "+fi.Name(), yellow, normal)
	case contains([]string{".rs"}, ext):
		style = Color(" "+fi.Name(), yellow, normal)
	case contains([]string{"Cargo.toml"}, fi.Name()):
		style = Color(" "+fi.Name(), yellow, underline)
	case "LICENSE" == file:
		style = Color(" "+fi.Name(), yellow, underline)
	case "Makefile" == file:
		style = Color(" "+fi.Name(), yellow, underline)
	case "justfile" == file:
		style = Color(" "+fi.Name(), yellow, underline)
	case "Cargo.toml" == file:
		style = Color(" "+fi.Name(), yellow, underline)
	case mode&os.ModeNamedPipe != 0:
		style = Color(" "+fi.Name(), yellow, normal)
	case mode&os.ModeSocket != 0:
		style = Color(" "+fi.Name(), green, italic)
	case mode&os.ModeDevice != 0 || mode&os.ModeCharDevice != 0:
		style = Color(" "+fi.Name(), yellow, normal)
	case mode&os.ModeSymlink != 0:
		if _, err := filepath.EvalSymlinks(file); err != nil {
			style = Color(" "+fi.Name(), red, normal)
		} else {
			style = Color(" "+fi.Name(), green, normal)
		}
	case mode&73 != 0:
		style = Color(" "+fi.Name(), green, normal)
	default:
		return file
	}
	return style
}

func contains(slice []string, str string) bool {
	for _, val := range slice {
		if val == strings.ToLower(str) {
			return true
		}
	}
	return false
}

func getsize(size int64) string {
	switch {
	// bytes
	case size < 1000:
		return fmt.Sprintf("%d bytes", size)
	// kb
	case size < (1000 * 1000):
		return fmt.Sprintf("%.1fkb", float64(size)/1000.0)
	// mb
	case size < (1000 * 1000 * 1000):
		return fmt.Sprintf("%.1fmb", float64(size)/(1000*1000))
	// gb
	case size < (1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.1fgb", float64(size)/(1000*1000*1000))
	}
	// default
	return ""
}

func fileInfo(fi fs.FileInfo) string {
	var info string
	if fi.IsDir() && !opts.NoColor {
		info = Color(" "+fi.Name(), blue, italic)
	} else if !opts.NoColor {
		info = getFileType(fi.Name(), fi)
	} else {
		info = fi.Name()
	}

	if opts.Size && !fi.IsDir() {

		value := fmt.Sprintf(" (%v)", getsize(fi.Size()))

		if !opts.NoColor {
			info += Color(value, cyan, italic)
		} else {
			info += value
		}
	}

	if opts.Mode {
		value := fmt.Sprintf(" [%o]", (fi.Mode()<<21)>>21)
		if !opts.NoColor {
			info += Color(value, magenta, italic)
		} else {
			info += value
		}
	}
	return info
}

func printTree(path string, depth int, padprefix string) error {
	dirent, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

OUTER:
	for index, fi := range dirent {
		if fi.IsDir() {
			for _, ignore := range opts.Ignore {
				if fi.Name() == ignore {
					continue OUTER
				}
			}
		}

		prefix := padprefix + "├──"

		if index == len(dirent)-1 {
			prefix = padprefix + "└──"
		}

		pad := padprefix + "│  "
		if index == len(dirent)-1 {
			pad = padprefix + "   "
		}

		item := fmt.Sprintf("%s %s", prefix, fileInfo(fi))
		fmt.Println(item)

		if fi.IsDir() && opts.Depth > 1 && depth != opts.Depth-1 {
			printTree(filepath.Join(path, fi.Name()), depth+1, pad)
		}
	}
	return nil
}

func Tree(args []string) error {
	for _, dir := range args {
		var path string
		if filepath.IsAbs(dir) {
			path = dir
		} else {
			workdir, _ := os.Getwd()
			path = filepath.Join(workdir, dir)
		}

		printTree(path, 0, "")
	}
	return nil
}

func init() {
	opts.Depth = 5
	opts.Ignore = []string{".git"}
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Tree(args); err != nil {
		log.Fatal(err)
	}
}
