// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"image/color"
	"math"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
	"github.com/biogo/biogo/feat/genome"
	mouse "github.com/biogo/biogo/feat/genome/mouse/mm10"
	"github.com/biogo/graphics/rings"
)

func main() {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	sty := plotter.DefaultLineStyle
	sty.Width /= 2

	chr := make([]feat.Feature, len(mouse.Chromosomes))
	for i, c := range mouse.Chromosomes {
		chr[i] = c
	}
	mm, err := rings.NewGappedBlocks(chr, rings.Arc{rings.Complete / 4 * rings.CounterClockwise, rings.Complete * rings.Clockwise}, 100, 110, 0.005)
	if err != nil {
		panic(err)
	}
	mm.LineStyle = sty
	p.Add(mm)

	bands := make([]feat.Feature, len(mouse.Bands))
	cens := make([]feat.Feature, 0, len(mouse.Chromosomes))
	for i, b := range mouse.Bands {
		bands[i] = colorBand{b}
		s := b.Start()
		// This condition depends on p -> q sort order in the $karyotype.Bands variable. All standard genome package follow this.
		if b.Band[0] == 'q' && (s == 0 || mouse.Bands[i-1].Band[0] == 'p') {
			cens = append(cens, colorBand{&genome.Band{Band: "cen", Desc: "Band", StartPos: s, EndPos: s, Giemsa: "acen", Chr: b.Location()}})
		}
	}
	b, err := rings.NewBlocks(bands, mm, 100, 110)
	if err != nil {
		panic(err)
	}
	p.Add(b)
	c, err := rings.NewBlocks(cens, mm, 100, 110)
	if err != nil {
		panic(err)
	}
	p.Add(c)

	font, err := vg.MakeFont("Helvetica", 7)
	if err != nil {
		panic(err)
	}
	lb, err := rings.NewLabels(mm, 117, rings.NameLabels(mm.Set)...)
	if err != nil {
		panic(err)
	}
	lb.TextStyle = draw.TextStyle{Color: color.Gray16{0}, Font: font}
	p.Add(lb)

	p.HideAxes()

	if err := p.Save(300, 300, "mouse.svg"); err != nil {
		panic(err)
	}
}

type colorBand struct {
	*genome.Band
}

func (b colorBand) FillColor() color.Color {
	switch b.Giemsa {
	case "acen":
		return color.RGBA{R: 0xff, A: 0xff}
	case "gneg":
		return color.Gray{0xff}
	case "gpos25":
		return color.Gray{3 * math.MaxUint8 / 4}
	case "gpos33":
		return color.Gray{2 * math.MaxUint8 / 3}
	case "gpos50":
		return color.Gray{math.MaxUint8 / 2}
	case "gpos66":
		return color.Gray{math.MaxUint8 / 3}
	case "gpos75":
		return color.Gray{math.MaxUint8 / 4}
	case "gpos100":
		return color.Gray{0x0}
	default:
		panic("unexpected giemsa value")
	}
}

func (b colorBand) LineStyle() draw.LineStyle {
	switch b.Giemsa {
	case "acen":
		return draw.LineStyle{Color: color.RGBA{R: 0xff, A: 0xff}, Width: 1}
	case "gneg", "gpos25", "gpos33", "gpos50", "gpos66", "gpos75", "gpos100":
		return draw.LineStyle{}
	default:
		panic("unexpected giemsa value")
	}
}
