// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rings implements a number of graphical representations of genomic features
// and feature associations using the idioms developed in the Circos distribution.
//
// The rings package borrows significantly from the ideas of Circos and shares some implementation
// details in order to run as a work-a-like. Circos is available from http://circos.ca/.
package rings

import (
	"image/color"

	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
)

// Twist is a flag type used to specify Ribbon and Sail twist behaviour. Specific interpretation
// of Twist flags is documented in the relevant types.
type Twist uint

const (
	None       Twist = 0         // None indicates no explicit twist.
	Flat       Twist = 1 << iota // Render feature connections without twist.
	Individual                   // Allow a feature or feature pair to define its ribbon twist.
	Twisted                      // Render feature connections with twist.
	Reverse                      // Reverse inverts all twist behaviour.
)

// ColorFunc allows dynamic assignment of color to objects based on passed parameters.
type ColorFunc func(interface{}) color.Color

// LineStyleFunc allows dynamic assignment of line styles to objects based on passed parameters.
type LineStyleFunc func(interface{}) draw.LineStyle

// Pair represents a pair of associated features.
type Pair interface {
	Features() [2]feat.Feature
}

// TextStyler is a type that can define its text style. For the purposes of the rings package
// the lines of a LineStyler that returns a nil Color or a TextStyle with Font.Size of 0 are not rendered.
type TextStyler interface {
	TextStyle() draw.TextStyle
}

// LineStyler is a type that can define its drawing line style. For the purposes of the rings package
// the lines of a LineStyler that returns a nil Color or a LineStyle with width 0 are not rendered.
type LineStyler interface {
	LineStyle() draw.LineStyle
}

// FillColorer is a type that can define its fill color. For the purposes of the rings package
// a FillColoer that returns a nil Color is not rendered filled.
type FillColorer interface {
	FillColor() color.Color
}

// XYer is a type that returns its x and y coordinates.
type XYer interface {
	XY() (x, y float64)
}
