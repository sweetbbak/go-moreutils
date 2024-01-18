package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Output    string `short:"o" long:"output" description:"output file"`
	Url       string `short:"u" long:"url" description:"url to retrieve"`
	UserAgent string `short:"U" long:"user-agent" description:"user agent to use"`
	Timeout   string `short:"t" long:"timeout" description:"N seconds timeout for request to fail"`
	NoClobber bool   `short:"n" long:"no-overwrite" description:"do not overwrite output file if it already exists"`
	Verbose   bool   `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func getUrlIndex(args []string) int {
	for i, item := range args {
		if strings.Contains(item, "http") || strings.Contains(item, "www") {
			return i
		}
	}
	return 0
}

func defaultOutput(url string) string {
	if url == "" || strings.HasSuffix(url, "/") {
		return "index.html"
	}
	return path.Base(url)
}

func request(cl *http.Client, url string, output string) error {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", opts.UserAgent)

	Debug("Retrieving url: %v\n", url)

	resp, err := cl.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Request failed: %v", resp.StatusCode)
	}

	var creation int
	if opts.NoClobber {
		creation = os.O_EXCL | os.O_CREATE | os.O_RDWR
	} else {
		creation = os.O_CREATE | os.O_RDWR
	}

	f, err := os.OpenFile(output, creation, 0o644)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	return err
}

func Wget(args []string) error {
	if opts.Url == "" {
		i := getUrlIndex(args)
		opts.Url = args[i]
	}

	url, err := url.Parse(opts.Url)
	if err != nil {
		return err
	}

	if opts.Output == "" {
		opts.Output = defaultOutput(url.Path)
	}

	t := opts.Timeout
	var d time.Duration
	if strings.HasSuffix(t, "s") {
		d, err = time.ParseDuration(t)
		if err != nil {
			return err
		}
	} else {
		d, err = time.ParseDuration(t + "s")
		if err != nil {
			return err
		}
	}

	client := tls_client(d)
	if err := request(client, opts.Url, opts.Output); err != nil {
		return err
	}

	return nil
}

func init() {
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	opts.UserAgent = userAgent
	opts.Timeout = "10s"
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Wget(args); err != nil {
		log.Fatal(err)
	}
}
