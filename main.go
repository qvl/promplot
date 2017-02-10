// Package main starts the binary.
// Argument parsing, usage information and the actual execution can be found here.
// See package promplot for using piece directly from you own Go code.
package main

import (
	"flag"
	"fmt"
	"io"
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

Save plot to file or send it right to a slack channel.
At least one of -slack or -file must be set.


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
		duration    = flags.Duration("range", 0, "Required. Time to look back to. Format: 5d12h34m56s")
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
	if *promServer == "" || *query == "" || *duration == 0 || (*file == "" && (*slackToken == "" || *channel == "")) {
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
	metrics, err := promplot.Metrics(*promServer, *query, *queryTime, *duration, step)
	fatal(err, "failed getting metrics")

	// Plot
	log("Creating plot \"%s\"", *title)
	tmp, err := promplot.Plot(metrics, *title, *format)
	defer cleanup(tmp)
	fatal(err, "failed creating plot")

	// Write to file
	if *file != "" {
		f, err := os.Open(tmp)
		fatal(err, "failed opening tmp file")
		defer func() {
			if err := f.Close(); err != nil {
				panic(fmt.Errorf("failed closing file: %v", err))
			}
		}()
		var out *os.File
		if *file == "-" {
			out = os.Stdout
			log("Writing to stdout")
		} else {
			out, err = os.Create(*file)
			fatal(err, "failed creating file")
			log("Writing to '%s'", *file)
		}
		_, err = io.Copy(out, f)
		fatal(err, "failed copying file")
	}

	// Upload to Slack
	if *slackToken != "" {
		log("Uploading to Slack channel \"%s\"", *channel)
		fatal(promplot.Slack(*slackToken, *channel, tmp, *title), "failed creating plot")
	}

	log("Done")
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
