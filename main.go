// Package main starts the binary.
// Argument parsing, usage information and the actual execution can be found here.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/palette/brewer"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/nlopes/slack"
	"github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/common/model"
	"qvl.io/promplot/flags"
)

// Can be set in build step using -ldflags
var version string

const (
	usage = `
Usage: %s [flags...]

Create and deliver plots from your Prometheus metrics.
Currently only the slack transport is implemented.


Flags:
`
	more = "\nFor more visit: https://qvl.io/promplot"
)

const (
	steps  = 100
	imgExt = ".png"
)

func main() {
	var (
		silent      = flag.Bool("silent", false, "Surpress all output.")
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

	if !*silent {
		fmt.Fprintf(os.Stderr, "Querying Prometheus \"%s\".\n", *query)
	}
	metrics, err := getMetrics(*promServer, *query, queryTime(), *duration)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed getting metrics:", err)
		os.Exit(1)
	}

	if !*silent {
		fmt.Fprintf(os.Stderr, "Creating plot \"%s\".\n", *title)
	}
	file, err := createPlot(metrics, *title)
	defer func() {
		if file == "" {
			return
		}
		if err := os.Remove(file); err != nil {
			fmt.Fprintln(os.Stderr, "failed deleting file:", err)
			os.Exit(1)
		}
	}()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed creating plot:", err)
		os.Exit(1)
	}

	if !*silent {
		fmt.Fprintf(os.Stderr, "Uploading to Slack channel \"%s\".\n", *channel)
	}
	if err = sendSlack(*slackToken, *channel, file, *title); err != nil {
		fmt.Fprintln(os.Stderr, "failed uploading to Slack:", err)
		os.Exit(1)
	}

	if !*silent {
		fmt.Fprintln(os.Stderr, "Done.")
	}
}

func getMetrics(server, query string, queryTime time.Time, duration time.Duration) (model.Matrix, error) {
	client, err := prometheus.New(prometheus.Config{Address: server})
	if err != nil {
		return nil, fmt.Errorf("failed creating Prometheus client: %v", err)
	}

	api := prometheus.NewQueryAPI(client)
	value, err := api.QueryRange(context.Background(), query, prometheus.Range{
		Start: queryTime.Add(-duration),
		End:   queryTime,
		Step:  duration / steps,
	})
	if err != nil {
		return nil, fmt.Errorf("failed querying Prometheus: %v", err)
	}

	metrics, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unsupported result format: %s", value.Type().String())
	}

	return metrics, nil
}

func createPlot(metrics model.Matrix, title string) (string, error) {
	p, err := plot.New()
	if err != nil {
		return "", fmt.Errorf("failed creating new plot: %v", err)
	}

	p.Title.Text = title
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Value"
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02\n15:04"}

	palette, err := brewer.GetPalette(brewer.TypeAny, "Spectral", max(len(metrics), 3))
	if err != nil {
		return "", fmt.Errorf("cannot get color palette: %v", err)
	}
	colors := palette.Colors()

	for s, sample := range metrics {
		data := make(plotter.XYs, len(sample.Values))
		for i, v := range sample.Values {
			data[i].X = float64(v.Timestamp.Unix())
			f, err := strconv.ParseFloat(v.Value.String(), 64)
			if err != nil {
				return "", fmt.Errorf("sample value not float: %s", v.Value.String())
			}
			data[i].Y = f
		}

		l, err := plotter.NewLine(data)
		if err != nil {
			return "", fmt.Errorf("failed creating line: %v", err)
		}
		l.LineStyle.Width = vg.Points(1)
		l.LineStyle.Color = colors[s]

		p.Add(l)
		p.Legend.Add(sample.Metric.String(), l)
		p.Legend.Top = true
	}

	file := filepath.Join(os.TempDir(), "promplot-"+strconv.FormatInt(time.Now().Unix(), 10)+imgExt)

	if err := p.Save(10*vg.Inch, 10*vg.Inch, file); err != nil {
		return "", fmt.Errorf("failed saving plot: %v", err)
	}

	return file, nil
}

func sendSlack(token, channel, file, title string) error {
	api := slack.New(token)
	params := slack.FileUploadParameters{
		Title:    title,
		Filetype: "image/png",
		Filename: title + ".png",
		File:     file,
		Channels: []string{channel},
	}
	_, err := api.UploadFile(params)
	return err
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
