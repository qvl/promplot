package promplot

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/palette/brewer"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/prometheus/common/model"
)

// Plot creates a plot from metric data and saves it to a temporary file.
// It's the callers responsibility to remove the returned file when no longer needed.
func Plot(metrics model.Matrix, title string) (string, error) {
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
