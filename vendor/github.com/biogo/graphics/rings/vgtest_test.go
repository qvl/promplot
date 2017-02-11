// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rings_test

import (
	"image"
	"image/color"

	"github.com/gonum/plot/vg"
)

const defaultDPI = 72

var _ vg.Canvas = (*canvas)(nil)

type canvas struct {
	dpi           float64
	base, actions []interface{}
}

func newCanvas(dpi float64, actions []interface{}) *canvas {
	return &canvas{
		dpi:     dpi,
		actions: actions,
		base:    actions,
	}
}

func (c *canvas) append(actions ...interface{}) {
	if actions == nil {
		c.actions = c.base
		return
	}
	c.actions = append(c.base, actions...)
}

type setWidth struct {
	w vg.Length
}

func (c *canvas) SetLineWidth(w vg.Length) {
	c.actions = append(c.actions, setWidth{w})
}

type setLineDash struct {
	dashes  []vg.Length
	offsets vg.Length
}

func (c *canvas) SetLineDash(dashes []vg.Length, offs vg.Length) {
	c.actions = append(c.actions, setLineDash{append([]vg.Length(nil), dashes...), offs})
}

type setColor struct {
	col color.Color
}

func (c *canvas) SetColor(col color.Color) {
	c.actions = append(c.actions, setColor{col})
}

type rotate struct {
	angle float64
}

func (c *canvas) Rotate(a float64) {
	c.actions = append(c.actions, rotate{a})
}

type translate struct{ x, y vg.Length }

func (c *canvas) Translate(pt vg.Point) {
	c.actions = append(c.actions, translate{pt.X, pt.Y})
}

type scale struct {
	x, y float64
}

func (c *canvas) Scale(x, y float64) {
	c.actions = append(c.actions, scale{x, y})
}

type push struct{}

func (c *canvas) Push() {
	c.actions = append(c.actions, push{})
}

type pop struct{}

func (c *canvas) Pop() {
	c.actions = append(c.actions, pop{})
}

type stroke struct {
	path vg.Path
}

func (c *canvas) Stroke(path vg.Path) {
	c.actions = append(c.actions, stroke{append(vg.Path(nil), path...)})
}

type fill struct {
	path vg.Path
}

func (c *canvas) Fill(path vg.Path) {
	c.actions = append(c.actions, fill{append(vg.Path(nil), path...)})
}

type fillString struct {
	font string
	size vg.Length
	x, y vg.Length
	str  string
}

func (c *canvas) FillString(font vg.Font, pt vg.Point, str string) {
	c.actions = append(c.actions, fillString{font.Name(), font.Size, pt.X, pt.Y, str})
}

type drawImage struct {
	rect vg.Rectangle
	image.Rectangle
}

func (c *canvas) DrawImage(rect vg.Rectangle, img image.Image) {
	c.actions = append(c.actions, drawImage{rect, img.Bounds()})
}
