// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings

import (
	"errors"
	"fmt"
	"image/color"
	"math"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
	"github.com/biogo/graphics/bezier"
)

// Ribbons implements rendering of feat.Feature associations as ribbons.
type Ribbons struct {
	// Set holds a collection of feature pairs to render.
	// If the features are both feat.Orienters this is taken into account according to Twist.
	Set []Pair

	// Ends holds the elements that define the end targets of the rendered ribbons.
	Ends [2]ArcOfer
	// Radii indicates the distance of the ribbon end points from the center of the plot.
	Radii [2]vg.Length

	// Twist indicates how feature orientation should be rendered.
	//
	// None indicates no explicit twist; ribbons are draw so that the start positions
	// each feature and the end positions of each feature are connected by Bézier curves.
	//
	//  f₀.Start -arc-> f₀.End -Bézier-> f₁.End -arc-> f₁.Start -Bézier-> f₀.Start
	//
	// Flat indicates ribbons should be rendered so that ribbons do not twist; paths are
	// drawn in angle sort order with each feature's end points joined by arcs.
	//
	// Individual allows a feature pair to define its ribbon twist; feature pairs where
	// both features satisfy feat.Orienter are rendered according to the product of their
	// orientations:
	//
	//  +1 - as if the Twist flag were set, ignoring all other flags except Reverse.
	//   0 - according to the states of all other Twist flags.
	//  -1 - as if the Flat flag were set, ignoring all other flags except Reverse.
	//
	// Twisted indicates ribbons should be rendered so that ribbons twist; paths of the
	// first feature are drawn in angle sort order and paths of the second are drawn in
	// reverse angle sort order, with each feature's end points joined by arcs.
	//
	// Reverse inverts all twist behaviour.
	//
	// If Twist has both Flat and Twisted flags set, DrawAt and Plot will panic.
	Twist Twist

	// Bezier describes the Bézier configuration for ribbon rendering.
	Bezier *Bezier

	// Color determines the fill color of each ribbon. If Color is not nil each ribbon is
	// rendered filled with the specified color, otherwise no fill is performed. This
	// behaviour is over-ridden if the feature describing the block is a FillColorer.
	Color color.Color

	// LineStyle determines the line style of each ribbon. LineStyle behaviour is over-ridden
	// for end point arcs if the feature describing an end point is a LineStyler and for
	// Bézier curves if the Pair is a LineStyler.
	LineStyle draw.LineStyle

	// X and Y specify rendering location when Plot is called.
	X, Y float64
}

// NewRibbons returns a Ribbons based on the parameters, first checking that the provided features
// are able to be rendered. An error is returned if the features are not renderable. The ends of
// a Ribbons ring cannot be an Arc or a Highlight.
func NewRibbons(fp []Pair, ends [2]ArcOfer, r [2]vg.Length) (*Ribbons, error) {
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
	return &Ribbons{
		Set:   fp,
		Ends:  ends,
		Radii: r,
	}, nil
}

// twist returns alters the ribbon twist depending on the relative orientation
// of the provided features and the Twist flags of the receiver.
func (r *Ribbons) twist(angles *[4]Angle, fp Pair) {
	p := fp.Features()
	var orient feat.Orientation
	switch {
	case r.Twist&(Flat|Twisted) == Flat|Twisted:
		panic("rings: cannot specify flat and twisted")
	case r.Twist == None:
		// p[0].Start() -> p[0].End() -> p[1].End() -> p[1].Start() {-> p[0].Start()}
		angles[2], angles[3] = angles[3], angles[2]
	case r.Twist&Individual != 0:
		var (
			o  [2]feat.Orienter
			ok [2]bool
		)
		o[0], ok[0] = p[0].(feat.Orienter)
		o[1], ok[1] = p[1].(feat.Orienter)
		if ok[0] && ok[1] {
			switch orient = o[0].Orientation() * o[1].Orientation(); orient {
			case feat.Forward:
				// p[0].Start() -> p[0].End() -> p[1].End() -> p[1].Start() {-> p[0].Start()}
				angles[2], angles[3] = angles[3], angles[2]
			case feat.Reverse, feat.NotOriented:
				// We do nothing in this case, since we already have the correct order:
				// p[0].Start() -> p[0].End() -> p[1].Start() -> p[1].End() {-> p[0].Start()}
				// If we have asked for flat or twisted, let that case handle the twist.
			default:
				panic("rings: illegal orientation")
			}
		} else {
			// Individual is equivalent to None if relative orientation is not available:
			// p[0].Start() -> p[0].End() -> p[1].End() -> p[1].Start() {-> p[0].Start()}
			angles[2], angles[3] = angles[3], angles[2]
		}
		if r.Twist&(Flat|Twisted) == 0 {
			break
		}
		fallthrough
	case r.Twist&(Flat|Twisted) != 0:
		if orient == feat.NotOriented {
			// Test relative positions on the arc of the start and end points
			// for each case of flat or twisted.
			if r.Twist&Flat != 0 {
				if (angles[0] > angles[1]) == (angles[2] < angles[3]) {
					// Points are not relatively flat, so swap.
					angles[2], angles[3] = angles[3], angles[2]
				}
			} else {
				if (angles[0] > angles[1]) != (angles[2] < angles[3]) {
					// Points are not relatively twisted, so swap.
					angles[2], angles[3] = angles[3], angles[2]
				}
			}
		}
	}
	if r.Twist&Reverse != 0 {
		// Swap the order of the second pair of points to reverse the order.
		angles[2], angles[3] = angles[3], angles[2]
	}
}

// DrawAt renders the feature pairs of a Ribbons at cen in the specified drawing area,
// according to the Ribbons configuration.
// DrawAt will panic if the feature pairs being linked both satisfy feat.Orienter and the
// product of orientations is not in feat.{Forward,NotOriented,Reverse}.
func (r *Ribbons) DrawAt(ca draw.Canvas, cen vg.Point) {
	if len(r.Set) == 0 {
		return
	}

	// Check if we have a Bézier and we want more than one segment in the curve.
	bez := r.Bezier != nil && r.Bezier.Segments > 1

	var pa vg.Path
loop:
	for _, fp := range r.Set {
		p := fp.Features()
		var min, max [2]int
		for j, loc := range [2]feat.Feature{p[0].Location(), p[1].Location()} {
			min[j] = loc.Start()
			max[j] = loc.End()
		}

		var angles [4]Angle
		// At the end of this loop we have:
		// p[0].Start() -> p[0].End() -> p[1].Start() -> p[1].End() {-> p[0].Start()}
		for j, f := range p {
			if f.Start() < min[j] || f.End() > max[j] {
				continue loop
			}

			arc, err := r.Ends[j].ArcOf(f.Location(), f)
			if err != nil {
				panic(fmt.Sprint("rings: no arc for feature location:", err))
			}

			angles[j*2] = Normalize(arc.Theta)
			angles[j*2+1] = Normalize(arc.Theta + arc.Phi)
		}
		r.twist(&angles, fp)

		pa = pa[:0]
		pa.Move(cen.Add(Rectangular(angles[0], r.Radii[0])))
		var arcs [2]int
		for j, rad := range r.Radii {
			// Arc from angles[j*2] to angles[j*2+1] with radius rad around cen.
			arcs[j] = len(pa) // Remember where the arcs are.
			start := angles[j*2]
			end := angles[j*2+1]
			pa.Arc(cen, rad, float64(start), float64(end-start))

			// Bézier from angles[j*2+1]@radius[j] to angles[(j*2+2)%4]@radius[1-j]
			// through r.Bezier if it is not nil and we wanted more than 1 segment;
			// otherwise straight lines.
			next := angles[(j*2+2)%4]
			if bez {
				b := bezier.New(
					r.Bezier.ControlPoints(
						[2]Angle{end, next},
						[2]vg.Length{rad, r.Radii[1-j]},
					)...,
				)
				for i := 1; i <= r.Bezier.Segments; i++ {
					pa.Line(cen.Add(b.Point(float64(i) / float64(r.Bezier.Segments))))
				}
			} else {
				pa.Line(cen.Add(Rectangular(next, r.Radii[1-j])))
			}
		}

		var col color.Color
		if c, ok := fp.(FillColorer); ok {
			col = c.FillColor()
		} else {
			col = r.Color
		}
		if col != nil {
			ca.SetColor(col)
			ca.Fill(pa)
		}

		if ls, ok := fp.(LineStyler); ok || (r.LineStyle.Color != nil && r.LineStyle.Width != 0) {
			// Change Arc vg.PathComps to Move vg.PathComps where necessary.
			for j, rad := range r.Radii {
				if _, ok := p[j].(LineStyler); ok {
					// The feature wants to define its own line style, so don't draw arc.
					end := angles[j*2+1]
					pa[arcs[j]] = vg.PathComp{
						Type: vg.MoveComp,
						Pos:  cen.Add(Rectangular(end, rad)),
					}
				}
			}

			var sty draw.LineStyle
			if ok {
				sty = ls.LineStyle()
			} else {
				sty = r.LineStyle
			}
			if sty.Color != nil && sty.Width != 0 {
				ca.SetLineStyle(sty)
				ca.Stroke(pa)
			}
		}

		// Draw feature ends according to the feature's linestyle if it has one.
		for j, rad := range r.Radii {
			if f, ok := p[j].(LineStyler); ok {
				pa = pa[:0]
				//Arc from angles[j*2] to angles[j*2+1] with radius rad around cen.
				start := angles[j*2]
				end := angles[j*2+1]
				pa.Move(cen.Add(Rectangular(start, rad)))
				pa.Arc(cen, rad, float64(start), float64(end-start))
				ca.SetLineStyle(f.LineStyle())
				ca.Stroke(pa)
			}
		}
	}
}

// Plot calls DrawAt using the Ribbons' X and Y values as the drawing coordinates.
func (r *Ribbons) Plot(ca draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&ca)
	r.DrawAt(ca, vg.Point{trX(r.X), trY(r.Y)})
}

// GlyphBoxes returns a liberal glyphbox for the ribbons rendering.
func (r *Ribbons) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
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
			var min, max [2]int
			for j, loc := range [2]feat.Feature{p[0].Location(), p[1].Location()} {
				if loc != nil {
					min[j] = loc.Start()
					max[j] = loc.End()
				}
			}

			var angles [4]Angle
			for j, f := range p {
				if f.Start() < min[j] || f.End() > max[j] {
					continue loop
				}

				arc, err := r.Ends[j].ArcOf(f.Location(), f)
				if err != nil {
					panic(fmt.Sprint("rings: no arc for feature location:", err))
				}
				angles[j*2] = Normalize(arc.Theta)
				angles[j*2+1] = Normalize(arc.Theta + arc.Phi)
			}
			r.twist(&angles, fp)

			for j := range r.Radii {
				end := angles[j*2+1]
				next := angles[(j*2+2)%4]
				b := bezier.New(
					r.Bezier.ControlPoints(
						[2]Angle{end, next},
						[2]vg.Length{r.Radii[j], r.Radii[1-j]},
					)...,
				)
				for k := 0; k <= r.Bezier.Segments; k++ {
					e := b.Point(float64(k) / float64(r.Bezier.Segments))
					if d := math.Hypot(float64(e.X), float64(e.Y)); d > rad {
						rad = d
					}
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
