// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings

import (
	"image/color"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
)

// Highlight implements rendering a colored arc.
type Highlight struct {
	// Base describes the arc through which the highlight should be drawn.
	Base Arc

	// Color determines the fill color of the highlight.
	Color color.Color

	// LineStyle determines the line style of the highlight.
	LineStyle draw.LineStyle

	// Inner and Outer define the inner and outer radii of the blocks.
	Inner, Outer vg.Length

	// X and Y specify rendering location when Plot is called.
	X, Y float64
}

// NewHighlight returns a Highlight based on the parameters, first checking that the provided features
// are able to be rendered. An error is returned if the features are not renderable.
func NewHighlight(col color.Color, base Arc, inner, outer vg.Length) *Highlight {
	return &Highlight{
		Color: col,
		Base:  base,
		Inner: inner,
		Outer: outer,
	}
}

// DrawAt renders the feature of a Highlight at cen in the specified drawing area,
// according to the Highlight configuration.
func (r *Highlight) DrawAt(ca draw.Canvas, cen vg.Point) {
	if r.Color == nil && (r.LineStyle.Color == nil || r.LineStyle.Width == 0) {
		return
	}

	var pa vg.Path

	pa.Move(cen.Add(Rectangular(r.Base.Theta, r.Inner)))
	pa.Arc(cen, r.Inner, float64(r.Base.Theta), float64(r.Base.Phi))
	if r.Base.Phi == Clockwise*Complete || r.Base.Phi == CounterClockwise*Complete {
		pa.Move(cen.Add(Rectangular(r.Base.Theta+r.Base.Phi, r.Outer)))
	}
	pa.Arc(cen, r.Outer, float64(r.Base.Theta+r.Base.Phi), float64(-r.Base.Phi))
	pa.Close()

	if r.Color != nil {
		ca.SetColor(r.Color)
		ca.Fill(pa)
	}
	if r.LineStyle.Color != nil && r.LineStyle.Width != 0 {
		ca.SetLineStyle(r.LineStyle)
		ca.Stroke(pa)
	}
}

// XY returns the x and y coordinates of the Highlight.
func (r *Highlight) XY() (x, y float64) { return r.X, r.Y }

// Arc returns the arc of the Highlight.
func (r *Highlight) Arc() Arc { return r.Base }

// Plot calls DrawAt using the Highlight's X and Y values as the drawing coordinates.
func (r *Highlight) Plot(ca draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&ca)
	r.DrawAt(ca, vg.Point{trX(r.X), trY(r.Y)})
}

// GlyphBoxes returns a liberal glyphbox for the highlight rendering.
func (r *Highlight) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	return []plot.GlyphBox{{
		X: plt.X.Norm(r.X),
		Y: plt.Y.Norm(r.Y),
		Rectangle: vg.Rectangle{
			Min: vg.Point{-r.Outer, -r.Outer},
			Max: vg.Point{r.Outer, r.Outer},
		},
	}}
}
