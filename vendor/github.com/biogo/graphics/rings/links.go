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
	"github.com/biogo/graphics/bezier"
)

// Links implements rendering of feat.Feature associations as Bézier curves.
type Links struct {
	// Set holds a collection of feature pairs to render.
	Set []Pair

	// Ends holds the elements that define the end targets of the rendered ribbons.
	Ends [2]ArcOfer
	// Radii indicates the distance of the ribbon end points from the center of the plot.
	Radii [2]vg.Length

	// Bezier describes the Bézier configuration for link rendering.
	Bezier *Bezier

	// LineStyle determines the line style of each link Bézier curve. LineStyle behaviour
	// is over-ridden if the Pair describing features is a LineStyler.
	LineStyle draw.LineStyle

	// X and Y specify rendering location when Plot is called.
	X, Y float64
}

// NewLinks returns a Links based on the parameters, first checking that the provided features
// are able to be rendered. An error is returned if the features are not renderable. The ends of
// a Links ring cannot be an Arc or a Highlight.
func NewLinks(fp []Pair, ends [2]ArcOfer, r [2]vg.Length) (*Links, error) {
	for _, p := range fp {
		for i, f := range p.Features() {
			if f.End() < f.Start() {
				return nil, errors.New("rings: inverted feature")
			}
			if _, err := ends[i].ArcOf(nil, f); err != nil {
				return nil, err
			}
		}
	}
	return &Links{
		Set:   fp,
		Ends:  ends,
		Radii: r,
	}, nil
}

// DrawAt renders the feature pairs of a Links at cen in the specified drawing area,
// according to the Links configuration.
func (r *Links) DrawAt(ca draw.Canvas, cen vg.Point) {
	if len(r.Set) == 0 {
		return
	}

	// Check if we have a Bézier and we want more than one segment in the curve.
	bez := r.Bezier != nil && r.Bezier.Segments > 1

	var pa vg.Path
loop:
	for _, fp := range r.Set {
		p := fp.Features()
		loc := [2]feat.Feature{p[0].Location(), p[1].Location()}
		var min, max [2]int
		for j, l := range loc {
			min[j] = l.Start()
			max[j] = l.End()
		}

		var angles [2]Angle
		for j, f := range p {
			if f.Start() < min[j] || f.Start() > max[j] {
				continue loop
			}

			arc, err := r.Ends[j].ArcOf(f.Location(), f)
			if err != nil {
				panic(fmt.Sprint("rings: no arc for feature location:", err))
			}
			angles[j] = Normalize(arc.Theta)
		}

		pa = pa[:0]
		pa.Move(cen.Add(Rectangular(angles[0], r.Radii[0])))
		// Bézier from angles[0]@radius[0] to angles[1]@radius[1] through
		// r.Bezier if it is not nil and we wanted more than 1 segment;
		// otherwise straight lines.
		if bez {
			b := bezier.New(
				r.Bezier.ControlPoints(angles, r.Radii)...,
			)
			for i := 1; i <= r.Bezier.Segments; i++ {
				pa.Line(cen.Add(b.Point(float64(i) / float64(r.Bezier.Segments))))
			}
		} else {
			pa.Line(cen.Add(Rectangular(angles[1], r.Radii[1])))
		}

		var sty draw.LineStyle
		if ls, ok := fp.(LineStyler); ok {
			sty = ls.LineStyle()
		} else {
			sty = r.LineStyle
		}
		if sty.Color != nil && sty.Width != 0 {
			ca.SetLineStyle(sty)
			ca.Stroke(pa)
		}
	}
}

// Plot calls DrawAt using the Links' X and Y values as the drawing coordinates.
func (r *Links) Plot(ca draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&ca)
	r.DrawAt(ca, vg.Point{trX(r.X), trY(r.Y)})
}

// GlyphBoxes returns a liberal glyphbox for the links rendering.
func (r *Links) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	if len(r.Set) == 0 {
		return nil
	}

	rad := float64(r.Radii[0])
	if float64(r.Radii[1]) > rad {
		rad = float64(r.Radii[1])
	}

	// If draw a Bézier we need to see if the radius is increased,
	// so we mock the drawing, just keeping a record of the furthest
	// distance from the origin. This may change to be more conservative.
	if r.Bezier != nil && r.Bezier.Segments > 1 {
	loop:
		for _, fp := range r.Set {
			p := fp.Features()
			loc := [2]feat.Feature{p[0].Location(), p[1].Location()}
			var min, max [2]int
			for j, l := range loc {
				min[j] = l.Start()
				max[j] = l.End()
			}

			var angles [2]Angle
			for j, f := range p {
				if f.Start() < min[j] || f.End() > max[j] {
					continue loop
				}

				arc, err := r.Ends[j].ArcOf(f.Location(), f)
				if err != nil {
					panic(fmt.Sprint("rings: no arc for feature location:", err))
				}
				angles[j] = Normalize(arc.Theta)
			}

			b := bezier.New(
				r.Bezier.ControlPoints(angles, r.Radii)...,
			)
			for k := 0; k <= r.Bezier.Segments; k++ {
				e := b.Point(float64(k) / float64(r.Bezier.Segments))
				if d := math.Hypot(float64(e.X), float64(e.Y)); d > rad {
					rad = d
				}
			}
		}
	}

	return []plot.GlyphBox{{
		X: plt.X.Norm(r.X),
		Y: plt.Y.Norm(r.Y),
		Rectangle: vg.Rectangle{
			Min: vg.Point{vg.Length(-rad), vg.Length(-rad)},
			Max: vg.Point{vg.Length(rad), vg.Length(rad)},
		},
	}}
}
