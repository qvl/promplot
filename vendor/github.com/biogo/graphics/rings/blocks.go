// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
)

// Blocks implements rendering of feat.Features as radial blocks.
type Blocks struct {
	// Set holds a collection of features to render.
	Set []feat.Feature

	// Base defines the targets of the rendered blocks.
	Base ArcOfer

	// Color determines the fill color of each block. If Color is not nil each block is rendered
	// filled with the specified color, otherwise no fill is performed. This behaviour is
	// over-ridden if the feature describing the block is a FillColorer.
	Color color.Color

	// LineStyle determines the line style of each block. LineStyle behaviour
	// is over-ridden if the feature describing a block is a LineStyler.
	LineStyle draw.LineStyle

	// Inner and Outer define the inner and outer radii of the blocks.
	Inner, Outer vg.Length

	// X and Y specify rendering location when Plot is called.
	X, Y float64
}

// NewBlocks returns a Blocks based on the parameters, first checking that the provided features
// are able to be rendered. An error is returned if the features are not renderable.
func NewBlocks(fs []feat.Feature, base ArcOfer, inner, outer vg.Length) (*Blocks, error) {
	if inner > outer {
		return nil, errors.New("rings: inner radius greater than outer radius")
	}
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
	return &Blocks{
		Set:   fs,
		Inner: inner,
		Outer: outer,
		Base:  base,
	}, nil
}

// NewGappedBlocks is a convenience wrapper of NewBlocks that guarantees to provide a valid ArcOfer based
// of the provided Arcer. If the provided Arcer is an ArcOfer it is tested for validity and a new ArcOfer is
// created only if needed.
func NewGappedBlocks(fs []feat.Feature, base Arcer, inner, outer vg.Length, gap float64) (*Blocks, error) {
	if inner > outer {
		return nil, errors.New("rings: inner radius greater than outer radius")
	}
	var b ArcOfer
	switch base := base.(type) {
	case ArcOfer:
		b = base
		for _, f := range fs {
			if _, err := base.ArcOf(f, nil); err != nil {
				b = NewGappedArcs(base, fs, gap)
				break
			}
		}
	default:
		b = NewGappedArcs(base, fs, gap)
	}
	return NewBlocks(fs, b, inner, outer)
}

// DrawAt renders the feature of a Blocks at cen in the specified drawing area,
// according to the Blocks configuration.
func (r *Blocks) DrawAt(ca draw.Canvas, cen vg.Point) {
	if len(r.Set) == 0 {
		return
	}

	var pa vg.Path
	for _, f := range r.Set {
		pa = pa[:0]

		arc, err := r.Base.ArcOf(f.Location(), f)
		if err != nil {
			panic(fmt.Sprintf("rings: no arc for feature location: %v", err))
		}

		pa.Move(cen.Add(Rectangular(arc.Theta, r.Inner)))
		pa.Arc(cen, r.Inner, float64(arc.Theta), float64(arc.Phi))
		if arc.Phi == Clockwise*Complete || arc.Phi == CounterClockwise*Complete {
			if c, ok := f.(feat.Conformationer); ok && c.Conformation() == feat.Circular {
				pa.Move(cen.Add(Rectangular(arc.Theta+arc.Phi, r.Outer)))
			}
		}
		pa.Arc(cen, r.Outer, float64(arc.Theta+arc.Phi), float64(-arc.Phi))
		pa.Close()

		if c, ok := f.(FillColorer); ok {
			ca.SetColor(c.FillColor())
			ca.Fill(pa)
		} else if r.Color != nil {
			ca.SetColor(r.Color)
			ca.Fill(pa)
		}

		var sty draw.LineStyle
		if ls, ok := f.(LineStyler); ok {
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

// XY returns the x and y coordinates of the Blocks.
func (r *Blocks) XY() (x, y float64) { return r.X, r.Y }

// Arc returns the base arc of the Blocks.
func (r *Blocks) Arc() Arc { return r.Base.Arc() }

// ArcOf returns the Arc location of the parameter. If the location is not found in
// the Blocks, an error is returned.
func (r *Blocks) ArcOf(loc, f feat.Feature) (Arc, error) { return r.Base.ArcOf(loc, f) }

type featureOrienter interface {
	feat.Feature
	feat.Orienter
}

// globalOrientation returns the orientation of a feature depending on it parent features' orientations.
func globalOrientation(f featureOrienter) feat.Orientation {
	if fo, ok := f.Location().(featureOrienter); ok {
		return globalOrientation(fo) * f.Orientation()
	}
	return f.Orientation()
}

// Plot calls DrawAt using the Blocks' X and Y values as the drawing coordinates.
func (r *Blocks) Plot(ca draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&ca)
	r.DrawAt(ca, vg.Point{trX(r.X), trY(r.Y)})
}

// GlyphBoxes returns a liberal glyphbox for the blocks rendering.
func (r *Blocks) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	return []plot.GlyphBox{{
		X: plt.X.Norm(r.X),
		Y: plt.Y.Norm(r.Y),
		Rectangle: vg.Rectangle{
			Min: vg.Point{-r.Outer, -r.Outer},
			Max: vg.Point{r.Outer, r.Outer},
		},
	}}
}
