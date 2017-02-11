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
)

// Scorer describes features that can provided scored values.
type Scorer interface {
	feat.Feature
	Scores() []float64
}

// ScoreRenderer is a type that produces a graphical representation of a score series
// for a Scores ring.
type ScoreRenderer interface {
	// Configure sets up the ScoreRenderer for set-wide values.
	// The min and max parameters may be ignored by an implementation.
	Configure(ca draw.Canvas, cen vg.Point, base ArcOfer, inner, outer vg.Length, min, max float64)

	// Render renders scores across the specified arc. Rendering may be
	// performed lazily.
	Render(Arc, Scorer)

	// Close finalises the rendering. For ScoreRenderers that do not
	// render lazily, this is a no-op.
	Close()
}

// Scores implements rendering of feat.Features as radial blocks.
type Scores struct {
	// Set holds a collection of features to render. Scores does not
	// make any check for Scorer overlap in Set.
	Set []Scorer

	// Base defines the targets of the rendered scores.
	Base ArcOfer

	// Renderer is the rendering implementation used to represent the
	// feature sets score data.
	Renderer ScoreRenderer

	// Min and Max hold the score range.
	Min, Max float64

	// Inner and Outer define the inner and outer radii of the blocks.
	Inner, Outer vg.Length

	// X and Y specify rendering location when Plot is called.
	X, Y float64
}

// NewScores returns a Scores based on the parameters, first checking that the provided features
// are able to be rendered. An error is returned if the features are not renderable.
func NewScores(fs []Scorer, base ArcOfer, inner, outer vg.Length, renderer ScoreRenderer) (*Scores, error) {
	min, max := math.Inf(1), math.Inf(-1)
	for _, f := range fs {
		if f.End() < f.Start() {
			return nil, errors.New("rings: inverted feature")
		}
		if loc := f.Location(); loc != nil {
			if f.Start() < loc.Start() || f.Start() > loc.End() {
				return nil, errors.New("rings: feature out of range")
			}
		}
		if _, err := base.ArcOf(nil, f); err != nil {
			return nil, err
		}
		for _, v := range f.Scores() {
			if math.IsNaN(v) {
				continue
			}
			min = math.Min(min, v)
			max = math.Max(max, v)
		}
	}
	if math.IsInf(max-min, 0) {
		return nil, errors.New("rings: score range is infinite")
	}
	return &Scores{
		Set:      fs,
		Base:     base,
		Renderer: renderer,
		Inner:    inner,
		Outer:    outer,
		Min:      min,
		Max:      max,
	}, nil
}

// DrawAt renders the feature of a Scores at cen in the specified drawing area,
// according to the Scores configuration.
func (r *Scores) DrawAt(ca draw.Canvas, cen vg.Point) {
	if len(r.Set) == 0 {
		return
	}

	r.Renderer.Configure(ca, cen, r.Base, r.Inner, r.Outer, r.Min, r.Max)
	for _, f := range r.Set {
		loc := f.Location()
		min := loc.Start()
		max := loc.End()

		if f.Start() < min || f.End() > max {
			continue
		}

		arc, err := r.Base.ArcOf(loc, f)
		if err != nil {
			panic(fmt.Sprint("rings: no arc for feature location:", err))
		}
		r.Renderer.Render(arc, f)
	}
	r.Renderer.Close()
}

// Plot calls DrawAt using the Scores' X and Y values as the drawing coordinates.
func (r *Scores) Plot(ca draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&ca)
	r.DrawAt(ca, vg.Point{trX(r.X), trY(r.Y)})
}

// GlyphBoxes returns a liberal glyphbox for the score rendering.
func (r *Scores) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	return []plot.GlyphBox{{
		X: plt.X.Norm(r.X),
		Y: plt.Y.Norm(r.Y),
		Rectangle: vg.Rectangle{
			Min: vg.Point{-r.Outer, -r.Outer},
			Max: vg.Point{r.Outer, r.Outer},
		},
	}}
}

// arcScore and arcScores are utility types for ordering Scores for
// ScoreRenderers that depend on sorted arcs.
type (
	arcScore struct {
		Arc
		Scorer
	}
	arcScores []arcScore
)

func (as arcScores) Len() int           { return len(as) }
func (as arcScores) Less(i, j int) bool { return as[i].Theta < as[j].Theta }
func (as arcScores) Swap(i, j int)      { as[i], as[j] = as[j], as[i] }

// Heat is a ScoreRenderer that represents feature scores as a color block.
type Heat struct {
	Palette   []color.Color
	Underflow color.Color
	Overflow  color.Color

	DrawArea draw.Canvas

	Center       vg.Point
	Inner, Outer vg.Length

	Min, Max float64
}

// Configure is called by Scores' DrawAt method. The min and max parameters are ignored if
// the Heat's Min and Max fields are both non-zero.
func (h *Heat) Configure(ca draw.Canvas, cen vg.Point, _ ArcOfer, inner, outer vg.Length, min, max float64) {
	h.DrawArea = ca
	h.Center = cen
	h.Inner = inner
	h.Outer = outer
	if h.Max == 0 && h.Min == 0 {
		h.Min = min
		h.Max = max
	}
}

// Render renders the values in scores across the specified arc from inner to outer.
// Rendering is performed eagerly.
func (h *Heat) Render(arc Arc, scorer Scorer) {
	scores := scorer.Scores()

	ps := float64(len(h.Palette)-1) / (h.Max - h.Min)

	// Define block progression inner to outer.
	d := (h.Outer - h.Inner) / vg.Length(len(scores))
	rad := h.Inner

	var pa vg.Path
	for _, v := range scores {
		pa = pa[:0]

		pa.Move(h.Center.Add(Rectangular(arc.Theta, rad)))
		pa.Arc(h.Center, rad, float64(arc.Theta), float64(arc.Phi))
		rad += d
		pa.Arc(h.Center, rad, float64(arc.Theta+arc.Phi), float64(-arc.Phi))
		pa.Close()

		var c color.Color
		switch {
		case math.IsNaN(v), math.IsInf(v, 0):
		case v < h.Min:
			c = h.Underflow
		case v > h.Max:
			c = h.Overflow
		default:
			c = h.Palette[int((v-h.Min)*ps+0.5)]
		}
		if c != nil {
			h.DrawArea.SetColor(c)
			h.DrawArea.Fill(pa)
		}
	}
}

// Close is a no-op.
func (h *Heat) Close() {}

// Trace is a ScoreRenderer that represents feature scores as a trace line.
type Trace struct {
	// LineStyles determines the lines style for each trace.
	LineStyles []draw.LineStyle

	// Join specifies whether adjacent features should be joined with radial lines.
	// It is overridden by the returned value of JoinTrace if the Scorer is a TraceJoiner.
	Join bool

	Base ArcOfer

	DrawArea draw.Canvas

	Center       vg.Point
	Inner, Outer vg.Length

	Min, Max float64

	// Axis represents a radial axis configuration
	Axis *Axis

	values arcScores
}

// Configure is called by Scores' DrawAt method. The min and max parameters are ignored if
// the Trace's Min and Max fields are both non-zero.
func (t *Trace) Configure(ca draw.Canvas, cen vg.Point, base ArcOfer, inner, outer vg.Length, min, max float64) {
	t.values = t.values[:0]
	t.DrawArea = ca
	t.Center = cen
	t.Base = base
	t.Inner = inner
	t.Outer = outer
	if t.Max == 0 && t.Min == 0 {
		t.Min = min
		t.Max = max
	}
}

// TraceJoiner is a type that can specify whether the traces for its scores should
// be joined when adjacent.
type TraceJoiner interface {
	// JoinTrace returns whether the ith score value should be part of a joined trace.
	JoinTrace(i int) bool
}

// Render add the scores at the specified arc for lazy rendering.
func (t *Trace) Render(arc Arc, scorer Scorer) {
	t.values = append(t.values, arcScore{arc, scorer})
}

// Close renders the added scores and axis.
func (t *Trace) Close() {
	if t.Axis != nil {
		set := make([]Scorer, len(t.values))
		for i, s := range t.values {
			set[i] = s.Scorer
		}
		t.Axis.drawAt(t.DrawArea, t.Center, set, t.Base, t.Inner, t.Outer, t.Min, t.Max)
	}

	sort.Sort(t.values)

	rs := float64(t.Outer-t.Inner) / (t.Max - t.Min)

	var pa vg.Path
	for i, arc := range t.values {
		for j, as := range arc.Scores() {
			if math.IsNaN(as) {
				continue
			}
			pa = pa[:0]

			if arc.Phi < 0 {
				arc.Theta, arc.Phi = arc.Theta+arc.Phi, -arc.Phi
			}

			var join, joined bool
			if tj, ok := arc.Scorer.(TraceJoiner); ok {
				join = tj.JoinTrace(j)
			} else {
				join = t.Join
			}
			if join && i != 0 && adjacent(t.values[i-1].Scorer, arc.Scorer) {
				prev := t.values[i-1].Scores()[j]
				if !math.IsNaN(prev) && ((t.Min <= as && as <= t.Max) || (t.Min <= prev && prev <= t.Max)) {
					joined = true

					prev = math.Min(math.Max(prev, t.Min), t.Max)
					as := math.Min(math.Max(as, t.Min), t.Max)

					pa.Move(t.Center.Add(Rectangular(arc.Theta, vg.Length((prev-t.Min)*rs)+t.Inner)))
					pa.Line(t.Center.Add(Rectangular(arc.Theta, vg.Length((as-t.Min)*rs)+t.Inner)))
				}
			}

			if t.Min <= as && as <= t.Max {
				rad := vg.Length((as-t.Min)*rs) + t.Inner
				if !joined {
					pa.Move(t.Center.Add(Rectangular(arc.Theta, rad)))
				}
				pa.Arc(t.Center, rad, float64(arc.Theta), float64(arc.Phi))
			}

			sty := t.LineStyles[j]
			if sty.Color != nil && sty.Width != 0 {
				t.DrawArea.SetLineStyle(sty)
				t.DrawArea.Stroke(pa)
			}
		}
	}
}

func adjacent(a, b feat.Feature) bool {
	return a.Location() == b.Location() && a.Start() == b.End() || b.Start() == a.End()
}
