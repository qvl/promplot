// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings

import (
	"errors"
	"fmt"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
)

// Blocks implements rendering of feat.Features representing 0 or 1 length features as radial lines.
type Spokes struct {
	// Set holds a collection of features to render.
	Set []feat.Feature

	// Base holds the elements that define the targets of the rendered spokes.
	Base ArcOfer

	// LineStyle determines the line style of each spoke. LineStyle is over-ridden
	// for each spoke if the feature describing the spoke is a LineStyler.
	LineStyle draw.LineStyle

	// Inner and Outer define the inner and outer radii of the spokes.
	Inner, Outer vg.Length

	// X and Y specify rendering location when Plot is called.
	X, Y float64
}

// NewSpokes returns a Spokes based on the parameters, first checking that the provided features
// are able to be rendered. An error is returned if the features are not renderable. The base of
// a Spokes ring cannot be an Arc or a Highlight.
func NewSpokes(fs []feat.Feature, base ArcOfer, inner, outer vg.Length) (*Spokes, error) {
	if inner > outer {
		return nil, errors.New("rings: inner radius greater than outer radius")
	}
	for _, f := range fs {
		if f.End() < f.Start() {
			return nil, errors.New("rings: inverted feature")
		}
		if f.End()-f.Start() > 1 {
			return nil, errors.New("rings: mark longer than one position")
		}
		if f.Start() < f.Location().Start() || f.Start() > f.Location().End() {
			return nil, errors.New("rings: mark out of range")
		}
		if _, err := base.ArcOf(nil, f); err != nil {
			return nil, err
		}
	}
	return &Spokes{
		Set:   fs,
		Inner: inner,
		Outer: outer,
		Base:  base,
	}, nil
}

// DrawAt renders the feature of a Spokes at cen in the specified drawing area,
// according to the Spokes configuration.
func (r *Spokes) DrawAt(ca draw.Canvas, cen vg.Point) {
	if len(r.Set) == 0 {
		return
	}

	var pa vg.Path
	for _, f := range r.Set {
		pa = pa[:0]

		loc := f.Location()
		min := loc.Start()
		max := loc.End()

		if f.Start() < min || f.Start() > max {
			continue
		}

		arc, err := r.Base.ArcOf(loc, f)
		if err != nil {
			panic(fmt.Sprintf("rings: no arc for feature location: %v\n%v", err, f))
		}

		pa.Move(cen.Add(Rectangular(arc.Theta, r.Inner)))
		pa.Line(cen.Add(Rectangular(arc.Theta, r.Outer)))

		var sty draw.LineStyle
		if ls, ok := f.(LineStyler); ok {
			sty = ls.LineStyle()
		} else {
			sty = r.LineStyle
		}
		if sty.Color != nil && sty.Width != 0 {
			ca.SetLineStyle(r.LineStyle)
			ca.Stroke(pa)
		}
	}
}

// XY returns the x and y coordinates of the Spokes.
func (r *Spokes) XY() (x, y float64) { return r.X, r.Y }

// Arc returns the base arc of the Spokes.
func (r *Spokes) Arc() Arc { return r.Base.Arc() }

// ArcOf returns the Arc location of the parameter. If the location is not found in
// the Spokes, an error is returned.
func (r *Spokes) ArcOf(loc, f feat.Feature) (Arc, error) { return r.Base.ArcOf(loc, f) }

// Plot calls DrawAt using the Spokes' X and Y values as the drawing coordinates.
func (r *Spokes) Plot(ca draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&ca)
	r.DrawAt(ca, vg.Point{trX(r.X), trY(r.Y)})
}

// GlyphBoxes returns a liberal glyphbox for the blocks rendering.
func (r *Spokes) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	return []plot.GlyphBox{{
		X: plt.X.Norm(r.X),
		Y: plt.Y.Norm(r.Y),
		Rectangle: vg.Rectangle{
			Min: vg.Point{-r.Outer, -r.Outer},
			Max: vg.Point{r.Outer, r.Outer},
		},
	}}
}
