// Package main starts the binary.
// Argument parsing, usage information and the actual execution can be found here.
// See package promplot for using piece directly from you own Go code.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
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
Currently only the Slack transport is implemented.


Flags:
`
	more = "\nFor more visit: https://qvl.io/promplot"
)

// Number of data points for the plot
const step = 100

func main() {
	var (
		silent      = flag.Bool("silent", false, "Suppress all output.")
		versionFlag = flag.Bool("version", false, "Print binary version.")
		promServer  = flag.String("url", "", "Required. URL of Prometheus server.")
		query       = flag.String("query", "", "Required. PQL query.")
		queryTime   = flags.UnixTime("time", time.Now(), "Required. Time for query (default is now). Format like the default format of the Unix date command.")
		duration    = flag.Duration("range", 0, "Required. Time to look back to. Format: 12h34m56s.")
		title       = flag.String("title", "Prometheus metrics", "Title of graph.")
		slackToken  = flag.String("slack", "", "Required. Slack API token (https://api.slack.com/docs/oauth-test-tokens).")
		channel     = flag.String("channel", "", "Required. Slack channel to post to.")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, more)
	}
	flag.Parse()

	if *versionFlag {
		fmt.Printf("ghbackup %s %s %s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// Required flag
	if *promServer == "" || *query == "" || *duration == 0 || *slackToken == "" || *channel == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Loggin helper
	log := func(format string, a ...interface{}) {
		if !*silent {
			fmt.Fprintf(os.Stderr, format, a...)
		}
	}

	// Fetch
	log("Querying Prometheus \"%s\".\n", *query)
	metrics, err := promplot.Metrics(*promServer, *query, queryTime(), *duration, step)
	fatal(err, "failed getting metrics")

	// Plot
	log("Creating plot \"%s\".\n", *title)
	file, err := promplot.Plot(metrics, *title)
	defer cleanup(file)
	fatal(err, "failed creating plot")

	// Upload
	log("Uploading to Slack channel \"%s\".\n", *channel)
	fatal(promplot.Slack(*slackToken, *channel, file, *title), "failed creating plot")

	log("Done.")
}

func cleanup(file string) {
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
