// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"fmt"
	"image/color"
	"math/rand"
	"os"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"

	"github.com/biogo/biogo/feat"
	"github.com/biogo/graphics/rings"
)

const name = "example_rings"

var extension string

func init() {
	flag.StringVar(&extension, "format", "svg", "specifies the output format of the example: eps, jpg, jpeg, pdf, png, svg, and tiff.")
	flag.Parse()
	for _, s := range []string{"eps", "jpg", "jpeg", "pdf", "png", "svg", "tiff"} {
		if extension == s {
			return
		}
	}
	flag.Usage()
	os.Exit(1)
}

func floatPtr(f float64) *float64 { return &f }

func main() {
	rand.Seed(int64(0))

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	sty := plotter.DefaultLineStyle
	sty.Width /= 5

	h := rings.NewHighlight(color.NRGBA{R: 243, G: 243, B: 21, A: 255}, rings.Arc{0, rings.Complete / 2 * rings.Clockwise}, 30, 120)
	h.LineStyle = sty
	p.Add(h)

	g := byte(0)
	for i := vg.Length(0); i < 3; i++ {
		bs, err := rings.NewGappedBlocks(randomFeatures(rand.Intn(10), 1000, 1000000, false, sty), rings.Arc{0, rings.Complete * rings.Clockwise}, 50+i*8, 55+i*8, 0.005)
		if err != nil {
			panic(err)
		}
		bs.Color = color.RGBA{R: 196, G: g, B: 128, A: 255}
		g += 60
		p.Add(bs)
	}

	bs, err := rings.NewGappedBlocks(randomFeatures(3, 100000, 1000000, false, sty), rings.Arc{0, rings.Complete * rings.Clockwise}, 80, 100, 0.01)
	if err != nil {
		panic(err)
	}
	bs.Set[0].(*fs).orient = feat.Forward
	bs.Set[1].(*fs).orient = feat.Forward
	bs.Set[2].(*fs).orient = feat.Forward
	bs.LineStyle = sty
	bs.Color = color.RGBA{R: 196, G: g + 24, B: 128, A: 255}
	g += 60
	p.Add(bs)

	font, err := vg.MakeFont("Helvetica", 10)
	if err != nil {
		panic(err)
	}
	lb, err := rings.NewLabels(bs, 110, rings.NameLabels(bs.Set)...)
	if err != nil {
		panic(err)
	}
	lb.TextStyle = draw.TextStyle{Color: color.Gray16{0}, Font: font}
	p.Add(lb)

	m := randomFeatures(400, bs.Set[1].Start(), bs.Set[1].End(), true, sty)
	for _, mf := range m {
		mf.(*fs).location = bs.Set[1]
	}
	ms, err := rings.NewSpokes(m, bs, 73, 78)
	if err != nil {
		panic(err)
	}
	ms.LineStyle = sty
	p.Add(ms)

	redSty := plotter.DefaultLineStyle
	redSty.Width *= 2
	redSty.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	blueSty := plotter.DefaultLineStyle
	blueSty.Width *= 2
	blueSty.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	bf := []rings.Pair{
		fp{
			feats: [2]*fs{
				{
					start: bs.Set[1].Start(), end: bs.Set[1].Start() + bs.Set[1].Len()/4,
					orient: feat.Reverse, location: bs.Set[1],
					style: redSty,
				},
				{
					start: bs.Set[2].Start() + 7*bs.Set[2].Len()/8, end: bs.Set[2].End(),
					orient: feat.Forward, location: bs.Set[2],
					style: blueSty,
				},
			},
			sty: sty,
		},
	}
	rs, err := rings.NewRibbons(bf, [2]rings.ArcOfer{bs, bs}, [2]vg.Length{47, 47})
	if err != nil {
		panic(err)
	}
	rs.Bezier = &rings.Bezier{Segments: 20}
	rs.Bezier.Radius.Length = 47 / 2
	rs.Twist = rings.Individual | rings.Flat
	rs.LineStyle = sty
	rs.Color = bs.Color
	p.Add(rs)

	sf := []feat.Feature{
		&fs{
			start: bs.Set[0].Start() + 2*bs.Set[0].Len()/5, end: bs.Set[0].End() - 2*bs.Set[0].Len()/5,
			orient: feat.NotOriented, location: bs.Set[0],
			style: redSty,
		},
		&fs{
			start: bs.Set[1].Start() + 2*bs.Set[1].Len()/5, end: bs.Set[1].End() - 2*bs.Set[1].Len()/5,
			orient: feat.NotOriented, location: bs.Set[1],
			style: redSty,
		},
		&fs{
			start: bs.Set[2].Start() + 2*bs.Set[2].Len()/5, end: bs.Set[2].End() - 2*bs.Set[2].Len()/5,
			orient: feat.Reverse, location: bs.Set[2],
			style: blueSty,
		},
	}
	s, err := rings.NewSail(sf, bs, 47)
	if err != nil {
		panic(err)
	}
	s.Bezier = &rings.Bezier{Segments: 20}
	s.Twist = rings.Individual | rings.Flat
	s.LineStyle = sty
	s.Color = color.NRGBA{R: 196, G: g, B: 128, A: 127}
	p.Add(s)

	mp := make([]rings.Pair, 20)
	for i := range mp {
		mp[i] = fp{feats: [2]*fs{m[i].(*fs), m[len(m)/2+i].(*fs)}, sty: sty}
	}
	ls, err := rings.NewLinks(mp, [2]rings.ArcOfer{bs, bs}, [2]vg.Length{47, 47})
	if err != nil {
		panic(err)
	}
	ls.Bezier = &rings.Bezier{Segments: 20,
		Radius: rings.LengthDist{Length: 2 * 47 / 3, Min: floatPtr(0.95), Max: floatPtr(1.05)},
		Crest:  &rings.FactorDist{Factor: 2, Min: floatPtr(0.7), Max: floatPtr(1.4)},
	}
	ls.LineStyle = sty
	p.Add(ls)

	p.Add(plotter.NewGlyphBoxes())

	p.HideAxes()

	if p.Save(300, 300, fmt.Sprintf("%s.%s", name, extension)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type fs struct {
	start, end int
	name       string
	location   feat.Feature
	orient     feat.Orientation
	style      draw.LineStyle
}

func (f *fs) Start() int                    { return f.start }
func (f *fs) End() int                      { return f.end }
func (f *fs) Len() int                      { return f.end - f.start }
func (f *fs) Name() string                  { return f.name }
func (f *fs) Description() string           { return "bogus" }
func (f *fs) Location() feat.Feature        { return f.location }
func (f *fs) Orientation() feat.Orientation { return f.orient }
func (f *fs) LineStyle() draw.LineStyle     { return f.style }

type fp struct {
	feats [2]*fs
	sty   draw.LineStyle
}

func (p fp) Features() [2]feat.Feature { return [2]feat.Feature{p.feats[0], p.feats[1]} }
func (p fp) LineStyle() draw.LineStyle {
	var col color.RGBA
	for _, f := range p.feats {
		r, g, b, a := f.style.Color.RGBA()
		col.R += byte(r / 2)
		col.G += byte(g / 2)
		col.B += byte(b / 2)
		col.A += byte(a / 2)
	}
	p.sty.Color = col
	return p.sty
}

func randomFeatures(n, min, max int, single bool, sty draw.LineStyle) []feat.Feature {
	data := make([]feat.Feature, n)
	for i := range data {
		start := rand.Intn(max-min) + min
		var end int
		if !single {
			end = rand.Intn(max - start)
		}
		data[i] = &fs{
			start: start,
			end:   start + end,
			name:  fmt.Sprintf("feature%v", i),
			style: sty,
		}
	}
	return data
}
