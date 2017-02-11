// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings

import (
	"errors"
	"fmt"
	"math"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
)

// Scale represents the circular axis of ring.
type Scale struct {
	// Set holds a collection of features to render scales for.
	Set []feat.Feature

	// Base defines the targets of the rendered blocks.
	Base ArcOfer

	// Radius define the radius of the axis.
	Radius vg.Length

	// LineStyle is the style of the axis line.
	LineStyle draw.LineStyle

	// Tick describes the scale's tick configuration.
	Tick TickConfig

	// Grid describes the scales grid configuration.
	Grid ScaleGrid

	X, Y float64
}

type ScaleGrid struct {
	// Inner and Outer specify the extend of radial grid lines.
	Inner, Outer vg.Length

	// LineStyle is the style of the axis line.
	LineStyle draw.LineStyle
}

// NewScale returns a Scale based on the parameters, first checking that the provided feature
// scales are able to be rendered. An error is returned if the scales are not renderable.
func NewScale(fs []feat.Feature, base ArcOfer, r vg.Length) (*Scale, error) {
	for _, f := range fs {
		if f.End() < f.Start() {
			return nil, errors.New("rings: inverted feature")
		}
		if loc := f.Location(); loc != nil {
			if f.Start() < loc.Start() || f.Start() > loc.End() {
				return nil, errors.New("rings: feature out of range")
			}
		}
		if _, err := base.ArcOf(f, nil); err != nil {
			return nil, err
		}
	}
	s := &Scale{
		Set:    fs,
		Base:   base,
		Radius: r,
	}
	s.Tick.Marker = plot.DefaultTicks{}

	return s, nil
}

// DrawAt renders the scales at cen in the specified drawing area, according to the
// Scale configuration.
func (r *Scale) DrawAt(ca draw.Canvas, cen vg.Point) {
	if len(r.Set) == 0 {
		return
	}

	var pa vg.Path
	for _, f := range r.Set {
		pa = pa[:0]

		min := f.Start()
		max := f.End()

		// TODO(kortschak) Remove this senseless waste of lines.
		if f.Start() < min || f.End() > max {
			continue
		}

		arc, err := r.Base.ArcOf(f, nil)
		if err != nil {
			panic(fmt.Sprint("rings: no arc for feature location:", err))
		}
		scale := arc.Phi / Angle(max-min)

		// These loops are split to reduce the amount of style changing between elements.
		marks := r.Tick.Marker.Ticks(float64(f.Start()), float64(f.End()))

		if r.Grid.Inner != r.Grid.Outer && r.Grid.LineStyle.Color != nil && r.Grid.LineStyle.Width != 0 {
			ca.SetLineStyle(r.Grid.LineStyle)
			for _, mark := range marks {
				iv := int(mark.Value)
				if iv < f.Start() || iv > f.End() {
					continue
				}
				pa = pa[:0]

				angle := Angle(iv-min)*scale + arc.Theta

				pa.Move(cen.Add(Rectangular(angle, r.Grid.Inner)))
				pa.Line(cen.Add(Rectangular(angle, r.Grid.Outer)))

				ca.Stroke(pa)
			}
		}

		if r.LineStyle.Color != nil && r.LineStyle.Width != 0 {
			start := arc.Theta
			end := Angle(f.End()-min)*scale + arc.Theta
			pa = pa[:0]
			pa.Move(cen.Add(Rectangular(start, r.Radius)))
			pa.Arc(cen, r.Radius, float64(start), float64(end-start))

			ca.SetLineStyle(r.LineStyle)
			ca.Stroke(pa)
		}

		if r.Tick.LineStyle.Color != nil && r.Tick.LineStyle.Width != 0 && r.Tick.Length != 0 {
			ca.SetLineStyle(r.LineStyle)
			for _, mark := range marks {
				iv := int(mark.Value)
				if iv < f.Start() || iv > f.End() {
					continue
				}
				pa = pa[:0]

				angle := Angle(iv-min)*scale + arc.Theta

				var length vg.Length
				if mark.IsMinor() {
					length = r.Tick.Length / 2
				} else {
					length = r.Tick.Length
				}
				pa.Move(cen.Add(Rectangular(angle, r.Radius)))
				pa.Line(cen.Add(Rectangular(angle, r.Radius+length)))

				ca.Stroke(pa)
			}
		}

		if r.Tick.Label.Color != nil {
			for _, mark := range marks {
				iv := int(mark.Value)
				if iv < f.Start() || iv > f.End() || mark.IsMinor() {
					continue
				}

				angle := Angle(iv-min)*scale + arc.Theta
				pt := cen.Add(Rectangular(angle, r.Radius+r.Tick.Length+r.Tick.Label.Font.Extents().Height))
				var (
					rot            Angle
					xalign, yalign float64
				)
				if r.Tick.Placement == nil {
					rot, xalign, yalign = DefaultPlacement(angle)
				} else {
					rot, xalign, yalign = r.Tick.Placement(angle)
				}
				r.Tick.Label.XAlign = draw.XAlignment(xalign)
				r.Tick.Label.YAlign = draw.YAlignment(yalign)
				r.Tick.Label.Rotation = float64(rot)

				ca.FillText(r.Tick.Label, pt, mark.Label)
			}
		}
	}
}

// Plot calls DrawAt using the Scale's X and Y values as the drawing coordinates.
func (r *Scale) Plot(ca draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&ca)
	r.DrawAt(ca, vg.Point{trX(r.X), trY(r.Y)})
}

// GlyphBoxes returns a liberal glyphbox for the label rendering.
func (r *Scale) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	grid := math.Max(float64(r.Grid.Inner), float64(r.Grid.Outer))
	radius := math.Max(float64(r.Radius+r.Tick.Length), grid)
	radius = math.Max(radius, float64(r.Tick.Label.Font.Extents().Height*2))
	return []plot.GlyphBox{{
		X: plt.X.Norm(r.X),
		Y: plt.Y.Norm(r.Y),
		Rectangle: vg.Rectangle{
			Min: vg.Point{-vg.Length(radius), -vg.Length(radius)},
			Max: vg.Point{vg.Length(radius), vg.Length(radius)},
		},
	}}
}
