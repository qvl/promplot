// Package promplot provides tools for fetching, plotting and sharing Prometheus metrics data.
// This tools are used by the promplot binary but can also be independently use by other Go programs.
package promplot

// For possible values see:
// https://godoc.org/github.com/gonum/plot/vg/draw#NewFormattedCanvas
const imgFormat = "png"

// ImgExt is the format and extension of the created image file
const ImgExt = "." + imgFormat
