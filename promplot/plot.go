package promplot

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/palette/brewer"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/prometheus/common/model"
)

// Only show important part of metric name
var labelText = regexp.MustCompile("\\{(.*)\\}")

// Plot creates a plot from metric data and saves it to a temporary file.
// It's the callers responsibility to remove the returned file when no longer needed.
func Plot(metrics model.Matrix, title string) (string, error) {
	p, err := plot.New()
	if err != nil {
		return "", fmt.Errorf("failed creating new plot: %v", err)
	}

	titleFont, err := vg.MakeFont("Helvetica-Bold", vg.Centimeter)
	if err != nil {
		return "", fmt.Errorf("failed creating font: %v", err)
	}
	textFont, err := vg.MakeFont("Helvetica", 3*vg.Millimeter)
	if err != nil {
		return "", fmt.Errorf("failed creating font: %v", err)
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
	palette, err := brewer.GetPalette(brewer.TypeAny, "Dark2", max(len(metrics), 3))
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
		p.Legend.Add(labelText.FindStringSubmatch(sample.Metric.String())[1], l)
	}

	file := filepath.Join(os.TempDir(), "promplot-"+strconv.FormatInt(time.Now().Unix(), 10)+ImgExt)

	if err := p.Save(24*vg.Centimeter, 20*vg.Centimeter, file); err != nil {
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
