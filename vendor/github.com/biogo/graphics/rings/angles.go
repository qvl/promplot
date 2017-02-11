// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings

import (
	"errors"
	"math"

	"github.com/biogo/biogo/feat"
	"github.com/gonum/plot/vg"
)

// Arcer is a type that describes an arc of circle.
type Arcer interface {
	Arc() Arc
}

// ArcOfer is an Arcer that contains a collection of features mapped to its span.
type ArcOfer interface {
	Arcer

	// ArcOf must return a non-nil error if the feat.Feature is not found by
	// the receiver or the query is nil, unless the receiver is an Arc. When
	// the receiver is an Arc the error returned is always nil.
	ArcOf(loc, f feat.Feature) (Arc, error)
}

// Point represents a 2-D point.
type Point struct {
	X, Y float64
}

// Normalize returns the angle corresponding to theta in the range [0, 2*math.Pi).
func Normalize(theta Angle) Angle { return Angle(math.Mod(float64(theta)+2*math.Pi, 2*math.Pi)) }

// Rectangular returns the rectangular coordinates for the location defined by theta and r
// in polar coordinates.
func Rectangular(theta Angle, r vg.Length) vg.Point {
	if r == 0 {
		return vg.Point{0, 0}
	}
	return vg.Point{X: vg.Length(math.Cos(float64(theta)) * float64(r)), Y: vg.Length(math.Sin(float64(theta)) * float64(r))}
}

// Polar returns the polar coordinates of a point.
func Polar(p vg.Point) (theta Angle, r vg.Length) {
	if (p == vg.Point{0, 0}) {
		return 0, 0
	}
	return Normalize(Angle(math.Atan2(float64(p.Y), float64(p.X)))), vg.Length(math.Hypot(float64(p.X), float64(p.Y)))
}

// Angle represents an angle in radians. Angles increase in the counter clockwise direction.
type Angle float64

var (
	angleNaN = Angle(math.NaN())
	arcNaN   = Arc{angleNaN, angleNaN}
)

const (
	Clockwise        Angle = -1
	CounterClockwise Angle = 1

	Complete Angle = Angle(2 * math.Pi)
)

// Arc represents an arc of a circle.
type Arc struct {
	Theta Angle // Initial angle of an arc in radians.
	Phi   Angle // The sweep of the arc in radians.
}

// Arc returns a copy of the Arc.
func (a Arc) Arc() Arc { return a }

// Contains returns a boolean indicating whether the parameter falls within the
// arc described by the receiver.
func (a Arc) Contains(alpha Angle) bool {
	if a.Phi >= 0 {
		a.Phi += a.Theta
	} else {
		a.Theta, a.Phi = a.Theta+a.Phi, a.Theta
	}
	alpha = Normalize(alpha)

	return alpha >= Normalize(a.Theta) && alpha <= Normalize(a.Phi)
}

// Arcs is the base ArcOfer implementation provided by the rings package.
type Arcs struct {
	Base Arc                  // Base represents the complete span of the Arcs.
	Arcs map[feat.Feature]Arc // Arcs provides a lookup for features within the span.
}

// NewGappedArcs returns an Arcs that maps the provided features to the base arc with
// a fractional gap between each feature.
func NewGappedArcs(base Arcer, fs []feat.Feature, gap float64) Arcs {
	arcs := make(map[feat.Feature]Arc, len(fs))

	var total float64
	for _, f := range fs {
		total += float64(f.Len())
	}

	arc := base.Arc()
	scale := arc.Phi * Angle((1-gap*float64(len(fs)))/total)
	g := Angle(gap) * arc.Phi

	theta := arc.Theta + g/2
	for _, f := range fs {
		if fo, ok := f.(featureOrienter); ok && globalOrientation(fo) == feat.Reverse {
			phi := Angle(f.Len()) * scale
			arcs[f] = Arc{Theta: Normalize(theta + phi), Phi: -phi}
		} else {
			arcs[f] = Arc{Theta: Normalize(theta), Phi: Angle(f.Len()) * scale}
		}
		theta += Angle(f.Len())*scale + g
	}

	return Arcs{Base: arc, Arcs: arcs}
}

// Arc returns the base arc of the Arcs.
func (a Arcs) Arc() Arc { return a.Base }

// ArcOf returns the arc of a feature in the context of the provided location.
//
// The behaviour of ArcOf depends on the the nil status of loc and f:
//
//  - if both loc and f are non-nil, f must have a sub-feature relationship with loc,
//    and the returned arc will be the arc of f.
//  - if either of loc or f are nil, then the arc of the non-nil parameter will be
//    returned.
//  - if both loc and f are nil, and no nil feature is found in the Arcs, the base arc
//    will be returned.
//
// If no matching feature is found a non-nil error is returned.
func (a Arcs) ArcOf(loc, f feat.Feature) (Arc, error) {
	var q feat.Feature
	switch {
	case loc != nil && f != nil:
		if !contains(loc, f) {
			return arcNaN, errors.New("rings: location is not parent of feature")
		}
		if f.Start() < loc.Start() || f.Start() > loc.End() {
			return arcNaN, errors.New("rings: feature out of range")
		}
		if fa, ok := a.containingArcOf(loc); ok {
			min, max := loc.Start(), loc.End()

			scale := fa.Phi / Angle(max-min)
			start, end := Angle(f.Start()-min)*scale, Angle(f.End()-min)*scale

			return Arc{start + fa.Theta, end - start}, nil
		}
		return arcNaN, errors.New("rings: location not found")
	case f != nil:
		q = f
	case loc != nil:
		q = loc
	}
	if fa, ok := a.containingArcOf(q); ok {
		return fa, nil
	} else if loc == nil && f == nil {
		return a.Base, nil
	}
	return arcNaN, errors.New("rings: location not found")
}

func contains(loc, f feat.Feature) bool {
	if loc == f {
		return true
	}
	for q := f; q != nil; {
		q = q.Location()
		if q == loc {
			return true
		}
	}
	return false
}

func (a Arcs) containingArcOf(f feat.Feature) (Arc, bool) {
	for q := f; q != nil; q = q.Location() {
		arc, ok := a.Arcs[q]
		if ok {
			return arc, ok
		}
	}
	if arc, ok := a.Arcs[nil]; ok {
		return arc, ok
	}
	return arcNaN, false
}

// Lens is an ArcOfer that can arbitrarily scale the arc of a feature.
type Lens struct {
	ArcOfer

	// Optics is a contextual arc scaling function.
	Optics func(loc, f feat.Feature, base, arc Arc) (Arc, error)
}

// ArcOf returns a scaled arc determined by a call to the Optics function using the
// arc returned by the embedded ArcOfer's ArcOf method.
func (l Lens) ArcOf(loc, f feat.Feature) (Arc, error) {
	arc, err := l.ArcOfer.ArcOf(loc, f)
	if err != nil {
		return arc, err
	}
	return l.Optics(loc, f, l.ArcOfer.Arc(), arc)
}
