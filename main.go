// Package main starts the binary.
// Argument parsing, usage information and the actual execution can be found here.
// See package promplot for using piece directly from you own Go code.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"qvl.io/promplot/flags"
	"qvl.io/promplot/promplot"
)

// Can be set in build step using -ldflags
var version string

const (
	usage = `
Usage: %s [flags...]

Create and deliver plots from your Prometheus metrics.

Save plot to file or send it right to a slack channel.
At least one of -dir or -slack must be set.


Flags:
`
	more = "\nFor more visit: https://qvl.io/promplot"
)

// Number of data points for the plot
const step = 100

func main() {
	var (
		silent      = flag.Bool("silent", false, "Optional. Suppress all output.")
		versionFlag = flag.Bool("version", false, "Optional. Print binary version.")
		promServer  = flag.String("url", "", "Required. URL of Prometheus server.")
		query       = flag.String("query", "", "Required. PQL query.")
		queryTime   = flags.UnixTime("time", time.Now(), "Required. Time for query (default is now). Format like the default format of the Unix date command.")
		duration    = flag.Duration("range", 0, "Required. Time to look back to. Format: 12h34m56s.")
		title       = flag.String("title", "Prometheus metrics", "Optional. Title of graph.")
		name        = flag.String("name", "promplot-"+strconv.FormatInt(time.Now().Unix(), 10), "Optional. Image file name. '"+promplot.ImgExt+"' is appended, so don't include it here.")
	)

	var (
		dir = flag.String("dir", "", "Directory to save plot to. Set to save plot as local file.")
	)

	var (
		slackToken = flag.String("slack", "", "Slack API token (https://api.slack.com/docs/oauth-test-tokens). Set to post plot to Slack.")
		channel    = flag.String("channel", "", "Required when -slack is set. Slack channel to post to.")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, more)
	}
	flag.Parse()

	if *versionFlag {
		fmt.Printf("promplot %s %s %s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// Required flags
	if *promServer == "" || *query == "" || *duration == 0 || (*dir == "" && (*slackToken == "" || *channel == "")) {
		flag.Usage()
		os.Exit(1)
	}

	// Loggin helper
	log := func(format string, a ...interface{}) {
		if !*silent {
			fmt.Fprintf(os.Stderr, format+"\n", a...)
		}
	}

	// Fetch from Prometheus
	log("Querying Prometheus \"%s\"", *query)
	metrics, err := promplot.Metrics(*promServer, *query, queryTime(), *duration, step)
	fatal(err, "failed getting metrics")

	// Plot
	log("Creating plot \"%s\"", *title)
	file, err := promplot.Plot(metrics, *title)
	defer cleanup(file, *dir == "")
	fatal(err, "failed creating plot")

	// Save local file
	if *dir != "" {
		dest := filepath.Join(*dir, *name+promplot.ImgExt)
		log("Saving to '%s'", dest)
		if err = os.Rename(file, dest); err != nil {
			log("failed saving local file: ", err)
		} else {
			file = dest
		}
	}

	// Upload to Slack
	if *slackToken != "" {
		log("Uploading to Slack channel \"%s\"", *channel)
		fatal(promplot.Slack(*slackToken, *channel, file, *name, *title), "failed creating plot")
	}

	log("Done")
}

func cleanup(file string, dirty bool) {
	if !dirty {
		return
	}
	if file == "" {
		return
	}
	fatal(os.Remove(file), "failed deleting file")
}

func fatal(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "msg: %v\n", err)
		os.Exit(1)
	}
}
