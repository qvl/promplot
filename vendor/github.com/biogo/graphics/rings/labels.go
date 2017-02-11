// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings

import (
	"fmt"
	"math"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
)

// Labeler is a type that can be used to label a block in a ring.
type Labeler interface {
	Label() string
}

// Label is a string that satisfies the Labeler interface. A Label may be used to
// label an Arc or a Highlight.
type Label string

// Label returns the string used to label a feature.
func (l Label) Label() string { return string(l) }

type locater interface {
	Labeler
	location() feat.Feature
}

// featLabel is a wrapper for feat.Feature to allow it to be used as a Labeler. The
// location method is required to allow the distinction between a feat.Feature that
// provides its own Label method.
type featLabel struct{ feat.Feature }

func (l featLabel) Label() string          { return l.Feature.Name() }
func (l featLabel) location() feat.Feature { return l.Feature }

// NameLabels returns a Labeler slice built from the provided slice of features. The
// labels returned are generated from the features' Name() values.
func NameLabels(fs []feat.Feature) []Labeler {
	l := make([]Labeler, len(fs))
	for i, f := range fs {
		if fl, ok := f.(locater); ok {
			l[i] = fl
		} else {
			l[i] = featLabel{f}
		}
	}
	return l
}

// Labels implements rendering of radial labels.
type Labels struct {
	// Labels contains the set of labels. Labelers that are feat.Features and are found
	// in the Base ArcOfer label the identified block with the string returned by
	// their Name method.
	Labels []Labeler

	// Base describes the ring holding the features to be labeled.
	Base ArcOfer

	// TextStyle determines the text style of each label. TextStyle behaviour
	// is over-ridden if the Label describing a block is a TextStyler.
	TextStyle draw.TextStyle

	// Radius define the inner radius of the labels.
	Radius vg.Length

	// Placement determines the text rotation and alignment. If Placement is
	// nil, DefaultPlacement is used.
	Placement TextPlacement

	// X and Y specify rendering location when Plot is called.
	X, Y float64
}

// NewLabels returns a Labels based on the parameters, first checking that the provided set of labels
// are able to be rendered; an Arc or Highlight may only take a single label, otherwise the labels
// must be a feat.Feature that can be found in the base ring. An error is returned if the labels are
// not renderable. If base is an XYer, the returned base XY values are used to populate the Labels' X
// and Y fields.
func NewLabels(base Arcer, r vg.Length, ls ...Labeler) (*Labels, error) {
	var b ArcOfer
	switch base := base.(type) {
	case ArcOfer:
		for _, l := range ls {
			var err error
			switch l := l.(type) {
			case locater:
				_, err = base.ArcOf(l.location(), nil)
			case feat.Feature:
				_, err = base.ArcOf(l, nil)
			default:
				_, err = base.ArcOf(nil, nil)
			}
			if err != nil {
				return nil, err
			}
		}
		b = base
	default:
		if len(ls) > 1 {
			return nil, fmt.Errorf("rings: cannot label a type %T with more than one feature", base)
		}
		arc := base.Arc()
		b = Arcs{Base: arc, Arcs: map[feat.Feature]Arc{feat.Feature(nil): arc}}
	}
	var x, y float64
	if xy, ok := base.(XYer); ok {
		x, y = xy.XY()
	}
	return &Labels{
		Labels: ls,
		Base:   b,
		Radius: r,
		X:      x,
		Y:      y,
	}, nil
}

// DrawAt renders the text of a Labels at cen in the specified drawing area,
// according to the Labels configuration.
func (r *Labels) DrawAt(ca draw.Canvas, cen vg.Point) {
	for _, l := range r.Labels {
		var sty draw.TextStyle
		if ts, ok := l.(TextStyler); ok {
			sty = ts.TextStyle()
		} else {
			sty = r.TextStyle
		}
		if sty.Color == nil || sty.Font.Size == 0 {
			continue
		}

		var (
			arc Arc
			err error
		)
		switch l := l.(type) {
		case locater:
			arc, err = r.Base.ArcOf(l.location().Location(), l.location())
		case feat.Feature:
			arc, err = r.Base.ArcOf(l.Location(), l)
		default:
			arc, err = r.Base.ArcOf(nil, nil)
		}
		if err != nil {
			panic(fmt.Sprint("rings: no arc for feature location:", err))
		}

		angle := arc.Theta + arc.Phi/2
		pt := cen.Add(Rectangular(angle, r.Radius))
		var (
			rot            Angle
			xalign, yalign float64
		)
		if r.Placement == nil {
			rot, xalign, yalign = DefaultPlacement(angle)
		} else {
			rot, xalign, yalign = r.Placement(angle)
		}
		sty.XAlign = draw.XAlignment(xalign)
		sty.YAlign = draw.YAlignment(yalign)
		sty.Rotation = float64(rot)
		ca.FillText(sty, pt, l.Label())
	}
}

// Plot calls DrawAt using the Labels' X and Y values as the drawing coordinates.
func (r *Labels) Plot(ca draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&ca)
	r.DrawAt(ca, vg.Point{trX(r.X), trY(r.Y)})
}

// GlyphBoxes returns a liberal glyphbox for the label rendering.
func (r *Labels) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	return []plot.GlyphBox{{
		X: plt.X.Norm(r.X),
		Y: plt.Y.Norm(r.Y),
		Rectangle: vg.Rectangle{
			Min: vg.Point{-r.Radius, -r.Radius},
			Max: vg.Point{r.Radius, r.Radius},
		},
	}}
}

// TextPlacement is used to determine text rotation and alignment by a Labels ring.
type TextPlacement func(Angle) (rot Angle, xadjust, yadjust float64)

var (
	DefaultPlacement TextPlacement = tangential
	Horizontal       TextPlacement = horizontal
	Radial           TextPlacement = radial
	Tangential       TextPlacement = tangential
)

func horizontal(a Angle) (rot Angle, xalign, yalign float64) {
	return 0, math.Cos(float64(a))/2 - 0.5, math.Sin(float64(a))/2 - 0.5
}

func radial(a Angle) (rot Angle, xalign, yalign float64) {
	return a, -0.5, -0.5
}

func tangential(a Angle) (rot Angle, xalign, yalign float64) {
	return a - math.Pi/2, -0.5, -0.5
}
