// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"sort"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
	"github.com/biogo/graphics/bezier"
)

// Sail implements rendering of feat.Feature associations as sails. A sail is conceptually
// a hyper edge connecting a number of features.
type Sail struct {
	// Set holds a collection of connected features to render.
	// If the features are feat.Orienters this is taken into account according to Twist.
	Set []feat.Feature

	// Base holds the element that defines the end targets of the rendered sail.
	Base ArcOfer
	// Radius indicates the distance of the sail end points from the center of the plot.
	Radius vg.Length

	// Twist indicates how feature orientation should be rendered.
	//
	// None indicates no explicit twist; sail ends are draw so that the start positions
	// each feature and the end positions of each feature are connected by Bézier curves.
	// Thus if features are numbered in order of their appearance along an arc:
	//
	//  f₀.Start -arc-> f₀.End -Bézier-> f₁.End -arc-> f₁.Start -Bézier-> ... f₀.Start
	//
	// Flat indicates sails should be rendered so that sail ends do not twist; paths are
	// drawn in angle sort order with each feature's end points joined by arcs.
	//
	// Individual allows a feature to define the twist of its sail end depending on its
	// orientation:
	//
	//  -1 - as if the Twist flag were set, ignoring all other flags except Reverse.
	//   0 - according to the states of all other Twist flags.
	//  +1 - as if the Flat flag were set, ignoring all other flags except Reverse.
	//
	// Twisted indicates sails should be rendered so that sail ends twist; the overall
	// progression is in angle sort order with Bézier paths drawn in angle sort order
	// and each feature's end points joined by arcs in reverse angle sort order.
	//
	// Reverse inverts all twist behaviour.
	//
	// If Twist has both Flat and Twisted flags set, DrawAt and Plot will panic.
	Twist Twist

	// Bezier describes the Bézier configuration for sail rendering.
	Bezier *Bezier

	// Color determines the fill color of each sail. If Color is not nil each sail is
	// rendered filled with the specified color, otherwise no fill is performed.
	Color color.Color

	// LineStyle determines the line style of each sail. LineStyle behaviour is over-ridden
	// for end point arcs if the feature describing an end point is a LineStyler.
	LineStyle draw.LineStyle

	// X and Y specify rendering location when Plot is called.
	X, Y float64
}

// NewSail returns a Sail based on the parameters, first checking that the provided features
// are able to be rendered. An error is returned if the features are not renderable. The base of
// a Sail ring cannot be an Arc or a Highlight.
func NewSail(fs []feat.Feature, base ArcOfer, r vg.Length) (*Sail, error) {
	for _, f := range fs {
		if f.End() < f.Start() {
			return nil, errors.New("rings: inverted feature")
		}
		if _, err := base.ArcOf(nil, f); err != nil {
			return nil, err
		}
	}
	return &Sail{
		Set:    fs,
		Base:   base,
		Radius: r,
	}, nil
}

// angleFeat and angleFeats are helper types required to determine render order and twist.
type (
	angleFeat struct {
		angles [2]Angle
		feat.Feature
	}
	angleFeats []angleFeat
)

func (af angleFeats) Len() int { return len(af) }
func (af angleFeats) Less(i, j int) bool {
	return af[i].angles[0]+af[i].angles[1] < af[j].angles[0]+af[j].angles[1]
}
func (af angleFeats) Swap(i, j int) { af[i], af[j] = af[j], af[i] }

// twist returns alters the sail twist depending on the relative orientation
// of the provided feature and the Twist flags of the receiver.
func (r *Sail) twist(af []angleFeat) {
	for i, f := range af {
		var orient feat.Orientation
		switch {
		case r.Twist&(Flat|Twisted) == Flat|Twisted:
			panic("rings: cannot specify flat and twisted")
		case r.Twist == None:
			// fs[0].Start() -> fs[0].End() -> fs[1].Start() -> fs[1].End() {... -> fs[0].Start()}
			af[i].angles[0], af[i].angles[1] = af[i].angles[1], af[i].angles[0]
		case r.Twist&Individual != 0:
			if o, ok := f.Feature.(feat.Orienter); ok {
				switch orient = o.Orientation(); orient {
				case feat.Reverse, feat.NotOriented:
					// We do nothing in this case, since we already have the correct order:
					// fs[0].Start() -> fs[0].End() -> fs[1].Start() -> fs[1].End() {... -> fs[0].Start()}
					// If we have asked for flat or twisted, let that case handle the twist.
				case feat.Forward:
					// p[0].Start() -> p[0].End() -> p[1].End() -> p[1].Start() {... -> p[0].Start()}
					af[i].angles[0], af[i].angles[1] = af[i].angles[1], af[i].angles[0]
				default:
					panic("rings: illegal orientation")
				}
			} else {
				// Individual is equivalent to None if relative orientation is not available:
				// fs[0].Start() -> fs[0].End() -> fs[1].End() -> fs[1].Start() {... -> fs[0].Start()}
				af[i].angles[0], af[i].angles[1] = af[i].angles[1], af[i].angles[0]
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
					if af[i].angles[0] > af[i].angles[1] {
						// Points are not relatively flat, so swap.
						af[i].angles[0], af[i].angles[1] = af[i].angles[1], af[i].angles[0]
					}
				} else {
					if af[i].angles[0] < af[i].angles[1] {
						// Points are not relatively twisted, so swap.
						af[i].angles[0], af[i].angles[1] = af[i].angles[1], af[i].angles[0]
					}
				}
			}
		}
		if r.Twist&Reverse != 0 {
			// Swap the order of the second pair of points to reverse the order.
			af[i].angles[0], af[i].angles[1] = af[i].angles[1], af[i].angles[0]
		}
	}
}

// DrawAt renders the features of a Sail at cen in the specified drawing area,
// according to the Sail configuration.
// DrawAt will panic if the feature pairs being linked both satisfy feat.Orienter and the
// product of orientations is not in feat.{Forward,NotOriented,Reverse}.
func (r *Sail) DrawAt(ca draw.Canvas, cen vg.Point) {
	if len(r.Set) == 0 {
		return
	}

	// Check if we have a Bézier and we want more than one segment in the curve.
	bez := r.Bezier != nil && r.Bezier.Segments > 1

	// Make an angle sorted slice of features.
	af := make(angleFeats, len(r.Set))
	var i, j int
	for i, j = 0, 0; i < len(r.Set); i, j = i+1, j+1 {
		f := r.Set[i]
		var min, max int
		loc := f.Location()
		if loc != nil {
			min = loc.Start()
			max = loc.End()
		}
		if f.Start() < min || f.End() > max {
			j--
			continue
		}

		af[j].Feature = f
		arc, err := r.Base.ArcOf(loc, f)
		if err != nil {
			panic(fmt.Sprint("rings: no arc for feature location:", err))
		}
		af[j].angles[0] = Normalize(arc.Theta)
		af[j].angles[1] = Normalize(arc.Theta + arc.Phi)
	}
	af = af[:j]
	sort.Sort(af)
	r.twist(af)

	var pa vg.Path
	pa.Move(cen.Add(Rectangular(af[0].angles[0], r.Radius)))
	arcs := make([]int, len(af))
	for i, f := range af {
		// Arc from f.angles[0] to f.angles[1] with radius r.Radius around cen.
		arcs[i] = len(pa) // Remember where the arcs are.
		start := f.angles[0]
		end := f.angles[1]
		pa.Arc(cen, r.Radius, float64(start), float64(end-start))

		// Bézier from f.angles[1]@radius to (circular successor of f).angles[0]@radius
		// through r.Bezier if it is not nil and we wanted more than 1 segment;
		// otherwise straight lines.
		next := af[(i+1)%len(af)].angles[0]
		if bez {
			b := bezier.New(
				r.Bezier.ControlPoints(
					[2]Angle{end, next},
					[2]vg.Length{r.Radius, r.Radius},
				)...,
			)
			for i := 1; i <= r.Bezier.Segments; i++ {
				pa.Line(cen.Add(b.Point(float64(i) / float64(r.Bezier.Segments))))
			}
		} else {
			pa.Line(cen.Add(Rectangular(next, r.Radius)))
		}
	}

	if r.Color != nil {
		ca.SetColor(r.Color)
		ca.Fill(pa)
	}

	if r.LineStyle.Color != nil && r.LineStyle.Width != 0 {
		// Change Arc vg.PathComps to Move vg.PathComps where necessary.
		for i, f := range af {
			if _, ok := f.Feature.(LineStyler); ok {
				// The feature wants to define its own line style, so don't draw arc.
				end := f.angles[1]
				pa[arcs[i]] = vg.PathComp{
					Type: vg.MoveComp,
					Pos:  cen.Add(Rectangular(end, r.Radius)),
				}
			}
		}

		if r.LineStyle.Color != nil && r.LineStyle.Width != 0 {
			ca.SetLineStyle(r.LineStyle)
			ca.Stroke(pa)
		}
	}

	for _, f := range af {
		// Draw feature ends according to the feature's linestyle if it has one.
		if ls, ok := f.Feature.(LineStyler); ok {
			pa = pa[:0]
			//Arc from f.angles[0] to f.angles[1] with radius r.Radius around cen.
			start := f.angles[0]
			end := f.angles[1]
			pa.Move(cen.Add(Rectangular(start, r.Radius)))
			pa.Arc(cen, r.Radius, float64(start), float64(end-start))
			ca.SetLineStyle(ls.LineStyle())
			ca.Stroke(pa)
		}
	}
}

// Plot calls DrawAt using the Sail's X and Y values as the drawing coordinates.
func (r *Sail) Plot(ca draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&ca)
	r.DrawAt(ca, vg.Point{trX(r.X), trY(r.Y)})
}

// GlyphBoxes returns a liberal glyphbox for the ribbons rendering.
func (r *Sail) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	if len(r.Set) == 0 {
		return nil
	}

	rad := float64(r.Radius)

	// If draw a Bézier we need to see if the radius is increased,
	// so we mock the drawing, just keeping a record of the furthest
	// distance from the origin. This may change to be more conservative.
	if r.Bezier != nil && r.Bezier.Segments > 1 {
		// Make an angle sorted slice of features.
		af := make(angleFeats, len(r.Set))
		var i, j int
		for i, j = 0, 0; i < len(r.Set); i, j = i+1, j+1 {
			f := r.Set[i]
			loc := f.Location()
			min := loc.Start()
			max := loc.End()
			if f.Start() < min || f.End() > max {
				j--
				continue
			}

			af[j].Feature = f
			arc, err := r.Base.ArcOf(loc, f)
			if err != nil {
				panic(fmt.Sprint("rings: no arc for feature location:", err))
			}
			af[j].angles[0] = Normalize(arc.Theta)
			af[j].angles[1] = Normalize(arc.Theta + arc.Phi)
		}
		af = af[:j]
		sort.Sort(af)
		r.twist(af)

		for i, f := range af {
			// Bézier from f.angles[1]@radius to (circular successor of f).angles[0]@radius
			// through r.Bezier if it is not nil and we wanted more than 1 segment;
			// otherwise straight lines.
			end := f.angles[1]
			next := af[(i+1)%len(af)].angles[0]
			b := bezier.New(
				r.Bezier.ControlPoints(
					[2]Angle{end, next},
					[2]vg.Length{r.Radius, r.Radius},
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

	return []plot.GlyphBox{{
		X: plt.X.Norm(r.X),
		Y: plt.Y.Norm(r.Y),
		Rectangle: vg.Rectangle{
			Min: vg.Point{vg.Length(-rad), vg.Length(-rad)},
			Max: vg.Point{vg.Length(rad), vg.Length(rad)},
		},
	}}
}
