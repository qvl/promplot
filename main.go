// Package main starts the binary.
// Argument parsing, usage information and the actual execution can be found here.
// See package promplot for using piece directly from you own Go code.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
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
One of -slack or -file must be set.


Flags:
`
	more = "\nFor more visit: https://qvl.io/promplot"
)

// Number of data points for the plot
const step = 100

func main() {
	var (
		silent      = flag.Bool("silent", false, "Optional. Suppress all output.")
		versionFlag = flag.Bool("version", false, "Print binary version.")
		promURL     = flag.String("url", "", "Required. URL of Prometheus server.")
		query       = flag.String("query", "", "Required. PQL query.")
		queryTime   = flags.UnixTime("time", time.Now(), "Time for query (default is now). Format like the default format of the Unix date command.")
		queryRange  = flags.Duration("range", 0, "Required. Time to look back to. Format: 5d12h34m56s")
		title       = flag.String("title", "Prometheus metrics", "Optional. Title of graph.")
		//
		format = flag.String("format", "png", "Optional. Image format. For possible values see: https://godoc.org/github.com/gonum/plot/vg/draw#NewFormattedCanvas")
	)

	var (
		file = flag.String("file", "", "File to save image to. Should have same extension as specified -format. Set -file to - to write to stdout.")
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
	var errs []string
	if *promURL == "" {
		errs = append(errs, "missing flag: -url")
	}
	if *query == "" {
		errs = append(errs, "missing flag: -query")
	}
	if *queryRange == 0 {
		errs = append(errs, "missing flag: -range")
	}
	if *file == "" && *slackToken == "" {
		errs = append(errs, "one of -file or -slack must be set")
	} else if *file != "" && *slackToken != "" {
		errs = append(errs, "only one of -file or -slack can be set")
	} else if *slackToken != "" && *channel == "" {
		errs = append(errs, "missing flag: -channel")
	}
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, strings.Join(errs, "\n")+"\n\nFor more info see %s -h\n", os.Args[0])
		os.Exit(1)
	}

	// Logging helper
	log := func(format string, a ...interface{}) {
		if !*silent {
			fmt.Fprintf(os.Stderr, format+"\n", a...)
		}
	}

	// Fetch from Prometheus
	log("Querying Prometheus %q", *query)
	metrics, err := promplot.Metrics(*promURL, *query, *queryTime, *queryRange, step)
	fatal(err, "failed to get metrics")

	// Plot
	log("Creating plot %q", *title)
	plot, err := promplot.Plot(metrics, *title, *format)
	fatal(err, "failed to create plot")

	// Write to file
	if *file != "" {
		var out *os.File
		if *file == "-" {
			log("Writing to stdout")
			out = os.Stdout
		} else {
			log("Writing to '%s'", *file)
			out, err = os.Create(*file)
			fatal(err, "failed to create file")
		}
		_, err = plot.WriteTo(out)
		fatal(err, "failed to copy to file")

		// Upload to Slack
	} else {
		log("Uploading to Slack channel %q", *channel)
		fatal(promplot.Slack(*slackToken, *channel, *title, plot), "failed to upload to Slack")
	}

	log("Done")
}

func fatal(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "msg: %v\n", err)
		os.Exit(1)
	}
}
