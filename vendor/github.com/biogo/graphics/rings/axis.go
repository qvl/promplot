// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings

import (
	"fmt"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
)

// Axis represents the radial axis of ring, usually a Scores.
type Axis struct {
	// Angle specifies the angular location of the axis.
	Angle Angle

	// Label describes the axis label configuration.
	Label AxisLabel

	// LineStyle is the style of the axis line.
	LineStyle draw.LineStyle

	// Tick describes the scale's tick configuration.
	Tick TickConfig

	// Grid is the style of the grid lines.
	Grid draw.LineStyle
}

// AxisLabel describes an axis label format and text.
type AxisLabel struct {
	// Text is the axis label string.
	Text string

	// TextStyle is the style of the axis label text.
	draw.TextStyle

	// Placement determines the text rotation and alignment.
	// If Placement is nil, DefaultPlacement is used.
	Placement TextPlacement
}

// TickConfig describes an axis tick configuration.
type TickConfig struct {
	// Label is the TextStyle on the tick labels.
	Label draw.TextStyle

	// LineStyle is the LineStyle of the tick lines.
	LineStyle draw.LineStyle

	// Placement determines the text rotation and alignment.
	// If Placement is nil, DefaultPlacement is used.
	Placement TextPlacement

	// Length is the length of a major tick mark.
	// Minor tick marks are half of the length of major
	// tick marks.
	Length vg.Length

	// Marker returns the tick marks. Any tick marks
	// returned by the Marker function that are not in
	// range of the axis are not drawn.
	Marker plot.Ticker
}

// drawAt renders the axis at cen in the specified drawing area, according to the
// Axis configuration.
func (r *Axis) drawAt(ca draw.Canvas, cen vg.Point, fs []Scorer, base ArcOfer, inner, outer vg.Length, min, max float64) {
	locMap := make(map[feat.Feature]struct{})

	var (
		pa vg.Path

		marks []plot.Tick

		scale = (outer - inner) / vg.Length(max-min)
	)
	for _, f := range fs {
		locMap[f.Location()] = struct{}{}
	}
	if r.Grid.Color != nil && r.Grid.Width != 0 {
		for loc := range locMap {
			arc, err := base.ArcOf(loc, nil)
			if err != nil {
				panic(fmt.Sprint("rings: no arc for feature location:", err))
			}

			ca.SetLineStyle(r.Grid)
			marks = r.Tick.Marker.Ticks(min, max)
			for _, mark := range marks {
				if mark.Value < min || mark.Value > max {
					continue
				}
				pa = pa[:0]

				radius := vg.Length(mark.Value-min)*scale + inner

				pa.Move(cen.Add(Rectangular(arc.Theta, radius)))
				pa.Arc(cen, radius, float64(arc.Theta), float64(arc.Phi))

				ca.Stroke(pa)
			}
		}
	}

	if r.LineStyle.Color != nil && r.LineStyle.Width != 0 {
		pa = pa[:0]

		pa.Move(cen.Add(Rectangular(r.Angle, inner)))
		pa.Line(cen.Add(Rectangular(r.Angle, outer)))

		ca.SetLineStyle(r.LineStyle)
		ca.Stroke(pa)
	}

	if r.Tick.LineStyle.Color != nil && r.Tick.LineStyle.Width != 0 && r.Tick.Length != 0 {
		ca.SetLineStyle(r.Tick.LineStyle)
		if marks == nil {
			marks = r.Tick.Marker.Ticks(min, max)
		}
		for _, mark := range marks {
			if mark.Value < min || mark.Value > max {
				continue
			}
			pa = pa[:0]

			radius := vg.Length(mark.Value-min)*scale + inner

			var length vg.Length
			if mark.IsMinor() {
				length = r.Tick.Length / 2
			} else {
				length = r.Tick.Length
			}
			off := Rectangular(r.Angle+Complete/4, length)
			e := Rectangular(r.Angle, radius)
			pa.Move(cen.Add(e))
			pa.Line(cen.Add(e.Add(off)))

			ca.Stroke(pa)

			if mark.IsMinor() || r.Tick.Label.Color == nil {
				continue
			}

			pt := cen.Add(Rectangular(r.Angle, radius).Add(vg.Point{off.X * 2, off.Y * 2}))
			var (
				rot            Angle
				xalign, yalign float64
			)
			if r.Tick.Placement == nil {
				rot, xalign, yalign = DefaultPlacement(r.Angle)
			} else {
				rot, xalign, yalign = r.Tick.Placement(r.Angle)
			}
			r.Tick.Label.XAlign = draw.XAlignment(xalign)
			r.Tick.Label.YAlign = draw.YAlignment(yalign)
			r.Tick.Label.Rotation = float64(rot)
			ca.FillText(r.Tick.Label, pt, mark.Label)
		}
	}

	if r.Label.Text != "" && r.Label.Color != nil {
		pt := cen.Add(Rectangular(r.Angle, (inner+outer)/2))
		var (
			rot            Angle
			xalign, yalign float64
		)
		if r.Label.Placement == nil {
			rot, xalign, yalign = DefaultPlacement(r.Angle)
		} else {
			rot, xalign, yalign = r.Label.Placement(r.Angle)
		}
		r.Label.TextStyle.XAlign = draw.XAlignment(xalign)
		r.Label.TextStyle.YAlign = draw.YAlignment(yalign)
		r.Label.TextStyle.Rotation = float64(rot)
		ca.FillText(r.Label.TextStyle, pt, r.Label.Text)
	}
}
