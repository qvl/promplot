package promplot

import (
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/prometheus/common/model"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette/brewer"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// Only show important part of metric name
var labelText = regexp.MustCompile("\\{(.*)\\}")

// Plot creates a plot from metric data and saves it to a temporary file.
// It's the callers responsibility to remove the returned file when no longer needed.
func Plot(metrics model.Matrix, title, format string) (io.WriterTo, error) {
	p, err := plot.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create new plot: %v", err)
	}

	titleFont, err := vg.MakeFont("Helvetica-Bold", vg.Centimeter)
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}
	textFont, err := vg.MakeFont("Helvetica", 3*vg.Millimeter)
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}

	p.Title.Text = title
	p.Title.Font = titleFont
	p.Title.Padding = 2 * vg.Centimeter
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02\n15:04"}
	p.X.Tick.Label.Font = textFont
	p.Y.Tick.Label.Font = textFont
	p.Legend.Font = textFont
	p.Legend.Top = true
	p.Legend.YOffs = 15 * vg.Millimeter

	// Color palette for drawing lines
	paletteSize := 8
	palette, err := brewer.GetPalette(brewer.TypeAny, "Dark2", paletteSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get color palette: %v", err)
	}
	colors := palette.Colors()

	for s, sample := range metrics {
		data := make(plotter.XYs, len(sample.Values))
		for i, v := range sample.Values {
			data[i].X = float64(v.Timestamp.Unix())
			f, err := strconv.ParseFloat(v.Value.String(), 64)
			if err != nil {
				return nil, fmt.Errorf("sample value not float: %s", v.Value.String())
			}
			data[i].Y = f
		}

		l, err := plotter.NewLine(data)
		if err != nil {
			return nil, fmt.Errorf("failed to create line: %v", err)
		}
		l.LineStyle.Width = vg.Points(1)
		l.LineStyle.Color = colors[s%paletteSize]

		p.Add(l)
		if len(metrics) > 1 {
			m := labelText.FindStringSubmatch(sample.Metric.String())
			if m != nil {
				p.Legend.Add(m[1], l)
			}
		}
	}

	// Draw plot in canvas with margin
	margin := 6 * vg.Millimeter
	width := 24 * vg.Centimeter
	height := 20 * vg.Centimeter
	c, err := draw.NewFormattedCanvas(width, height, format)
	if err != nil {
		return nil, fmt.Errorf("failed to create canvas: %v", err)
	}
	p.Draw(draw.Crop(draw.New(c), margin, -margin, margin, -margin))

	return c, nil
}
