#  :chart_with_upwards_trend: promplot

[![GoDoc](https://godoc.org/qvl.io/promplot?status.svg)](https://godoc.org/qvl.io/promplot)
[![Build Status](https://travis-ci.org/qvl/promplot.svg?branch=master)](https://travis-ci.org/qvl/promplot)
[![Go Report Card](https://goreportcard.com/badge/qvl.io/promplot)](https://goreportcard.com/report/qvl.io/promplot)


`promplot` is an opinionated tool to create plots from your [Prometheus](https://prometheus.io/) metrics and automatically sends them to you.

Currently the only implemented transport is [Slack](https://slack.com/).
But feel free to [add a new one](#development)!


    Usage: promplot [flags...]

    Create and deliver plots from your Prometheus metrics.

    Save plot to file or send it right to a slack channel.
    At least one of -dir or -slack must be set.


    Flags:
      -channel string
          Required when -slack is set. Slack channel to post to.
      -dir string
          Directory to save plot to. Set to save plot as local file.
      -name string
          Optional. Image file name. '.png' is appended, so don't include it here. (default "promplot-<unixtime>")
      -query string
          Required. PQL query.
      -range duration
          Required. Time to look back to. Format: 12h34m56s.
      -silent
          Optional. Suppress all output.
      -slack string
          Slack API token (https://api.slack.com/docs/oauth-test-tokens). Set to post plot to Slack.
      -time value
          Required. Time for query (default is now). Format like the default format of the Unix date command.
      -title string
          Optional. Title of graph. (default "Prometheus metrics")
      -url string
          Required. URL of Prometheus server.
      -version
          Optional. Print binary version.


## Example

It's simple to create a shell script for multiple plots:

```sh
common="-url $promurl -channel stats -slack $slacktoken"

promplot $common \
  -title "Free memory in MB" \
  -query "node_memory_MemFree /1024 /1024" \
  -range "24h"

promplot $common \
  -title "Free disk space in GB" \
  -query "node_filesystem_free /1024 /1024 /1024" \
  -range "24h"

promplot $common \
  -title "Open file descriptors" \
  -query "process_open_fds" \
  -range "24h"
```

And with a scheduler like [sleepto](https://qvl.io/sleepto) you can easily automate this script to run every day or once a week.


## Install

- With [Go](https://golang.org/):
```
go get qvl.io/promplot
```

- With [Homebrew](http://brew.sh/):
```
brew install qvl/tap/promplot
```

- Download from https://github.com/qvl/promplot/releases



## Development

Make sure to use `gofmt` and create a [Pull Request](https://github.com/qvl/promplot/pulls).

When changing external dependencies please use [godep](https://github.com/tools/godep/) to vendor them.


### Releasing

Push a new Git tag and [GoReleaser](https://github.com/goreleaser/releaser) will automatically create a release.


## License

[MIT](./license)
