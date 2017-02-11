// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bezier

import (
	"fmt"
	"math"
	"testing"

	"github.com/gonum/plot/vg"

	"gopkg.in/check.v1"
)

const epsilon = 1e-12

type approxChecker struct {
	*check.CheckerInfo
}

var approxEquals check.Checker = &approxChecker{
	&check.CheckerInfo{Name: "Approx", Params: []string{"obtained", "expected", "epsilon"}},
}

func (checker *approxChecker) Check(params []interface{}, names []string) (result bool, error string) {
	defer func() {
		if v := recover(); v != nil {
			result = false
			error = fmt.Sprint(v)
		}
	}()
	epsilon := params[2].(float64)
	p := [2]vg.Point{params[0].(vg.Point), params[1].(vg.Point)}
	return (math.Abs(float64(p[0].X-p[1].X)) <= epsilon || (math.IsNaN(float64(p[0].X)) && math.IsNaN(float64(p[1].X)))) &&
		(math.Abs(float64(p[0].Y-p[1].Y)) <= epsilon || (math.IsNaN(float64(p[0].Y)) && math.IsNaN(float64(p[1].Y)))), ""
}

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestNew(c *check.C) {
	for i, t := range []struct {
		ctrls []vg.Point
		curve Curve
	}{
		{
			ctrls: nil,
			curve: nil,
		},
		{
			ctrls: []vg.Point{{1, 2}, {3, 4}, {5, 6}, {7, 8}},
			curve: Curve{
				{Point: vg.Point{X: 1, Y: 2}, Control: vg.Point{X: 1, Y: 2}},
				{Point: vg.Point{X: 3, Y: 4}, Control: vg.Point{X: 9, Y: 12}},
				{Point: vg.Point{X: 5, Y: 6}, Control: vg.Point{X: 15, Y: 18}},
				{Point: vg.Point{X: 7, Y: 8}, Control: vg.Point{X: 7, Y: 8}},
			},
		},
		{
			ctrls: []vg.Point{{0, 0}, {0, 1}, {1, 1}, {1, 0}},
			curve: Curve{
				{Point: vg.Point{X: 0, Y: 0}, Control: vg.Point{X: 0, Y: 0}},
				{Point: vg.Point{X: 0, Y: 1}, Control: vg.Point{X: 0, Y: 3}},
				{Point: vg.Point{X: 1, Y: 1}, Control: vg.Point{X: 3, Y: 3}},
				{Point: vg.Point{X: 1, Y: 0}, Control: vg.Point{X: 1, Y: 0}},
			},
		},
		{
			ctrls: []vg.Point{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
			curve: Curve{
				{Point: vg.Point{X: 0, Y: 0}, Control: vg.Point{X: 0, Y: 0}},
				{Point: vg.Point{X: 0, Y: 1}, Control: vg.Point{X: 0, Y: 3}},
				{Point: vg.Point{X: 1, Y: 0}, Control: vg.Point{X: 3, Y: 0}},
				{Point: vg.Point{X: 1, Y: 1}, Control: vg.Point{X: 1, Y: 1}},
			},
		},
	} {
		bc := New(t.ctrls...)
		c.Check(bc, check.DeepEquals, t.curve, check.Commentf("Test %d: %+v", i, t.ctrls))
	}
}

func (s *S) TestPoint(c *check.C) {
	type tPoints []struct {
		t     float64
		point vg.Point
	}
	for i, t := range []struct {
		ctrls []vg.Point
		tPoints
	}{
		{
			ctrls: []vg.Point{{1, 2}, {3, 4}, {5, 6}, {7, 8}},
			tPoints: tPoints{
				{t: 0, point: vg.Point{1, 2}},
				{t: 0.1, point: vg.Point{1.6, 2.6}},
				{t: 0.2, point: vg.Point{2.2, 3.2}},
				{t: 0.3, point: vg.Point{2.8, 3.8}},
				{t: 0.4, point: vg.Point{3.4, 4.4}},
				{t: 0.5, point: vg.Point{4, 5}},
				{t: 0.6, point: vg.Point{4.6, 5.6}},
				{t: 0.7, point: vg.Point{5.2, 6.2}},
				{t: 0.8, point: vg.Point{5.8, 6.8}},
				{t: 0.9, point: vg.Point{6.4, 7.4}},
				{t: 1, point: vg.Point{7, 8}},
			},
		},
		{
			ctrls: []vg.Point{{0, 0}, {0, 1}, {1, 1}, {1, 0}},
			tPoints: tPoints{
				{t: 0, point: vg.Point{0, 0}},
				{t: 0.1, point: vg.Point{0.028, 0.27}},
				{t: 0.2, point: vg.Point{0.104, 0.48}},
				{t: 0.3, point: vg.Point{0.216, 0.63}},
				{t: 0.4, point: vg.Point{0.352, 0.72}},
				{t: 0.5, point: vg.Point{0.5, 0.75}},
				{t: 0.6, point: vg.Point{0.648, 0.72}},
				{t: 0.7, point: vg.Point{0.784, 0.63}},
				{t: 0.8, point: vg.Point{0.896, 0.48}},
				{t: 0.9, point: vg.Point{0.972, 0.27}},
				{t: 1, point: vg.Point{1, 0}},
			},
		},
		{
			ctrls: []vg.Point{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
			tPoints: tPoints{
				{t: 0, point: vg.Point{0, 0}},
				{t: 0.1, point: vg.Point{0.028, 0.244}},
				{t: 0.2, point: vg.Point{0.104, 0.392}},
				{t: 0.3, point: vg.Point{0.216, 0.468}},
				{t: 0.4, point: vg.Point{0.352, 0.496}},
				{t: 0.5, point: vg.Point{0.5, 0.5}},
				{t: 0.6, point: vg.Point{0.648, 0.504}},
				{t: 0.7, point: vg.Point{0.784, 0.532}},
				{t: 0.8, point: vg.Point{0.896, 0.608}},
				{t: 0.9, point: vg.Point{0.972, 0.756}},
				{t: 1, point: vg.Point{1, 1}},
			},
		},
	} {
		bc := New(t.ctrls...)
		for j, tPoint := range t.tPoints {
			c.Check(bc.Point(tPoint.t), approxEquals, t.tPoints[j].point, epsilon, check.Commentf("Test %d part %d: %+v", i, j, t.ctrls))
		}
	}
}

func (s *S) TestCurve(c *check.C) {
	for i, t := range []struct {
		ctrls  []vg.Point
		points []vg.Point
	}{
		{
			ctrls: []vg.Point{{1, 2}, {3, 4}, {5, 6}, {7, 8}},
			points: []vg.Point{
				{1, 2},
				{1.6, 2.6},
				{2.2, 3.2},
				{2.8, 3.8},
				{3.4, 4.4},
				{4, 5},
				{4.6, 5.6},
				{5.2, 6.2},
				{5.8, 6.8},
				{6.4, 7.4},
				{7, 8},
			},
		},
		{
			ctrls: []vg.Point{{0, 0}, {0, 1}, {1, 1}, {1, 0}},
			points: []vg.Point{
				{0, 0},
				{0.028, 0.27},
				{0.104, 0.48},
				{0.216, 0.63},
				{0.352, 0.72},
				{0.5, 0.75},
				{0.648, 0.72},
				{0.784, 0.63},
				{0.896, 0.48},
				{0.972, 0.27},
				{1, 0},
			},
		},
		{
			ctrls: []vg.Point{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
			points: []vg.Point{
				{0, 0},
				{0.028, 0.244},
				{0.104, 0.392},
				{0.216, 0.468},
				{0.352, 0.496},
				{0.5, 0.5},
				{0.648, 0.504},
				{0.784, 0.532},
				{0.896, 0.608},
				{0.972, 0.756},
				{1, 1},
			},
		},
		{
			ctrls:  []vg.Point{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
			points: []vg.Point{},
		},
		{
			ctrls: []vg.Point{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
			points: []vg.Point{
				{vg.Length(math.NaN()), vg.Length(math.NaN())},
			},
		}, {
			ctrls: []vg.Point{{0, 0}, {0, 1}, {1, 0}, {1, 1}},
			points: []vg.Point{
				{0, 0},
				{1, 1},
			},
		},
	} {
		bc := New(t.ctrls...).Curve(make([]vg.Point, len(t.points)))
		for j, p := range bc {
			c.Check(p, approxEquals, t.points[j], epsilon, check.Commentf("Test %d part %d: %+v", i, j, t.ctrls))
		}
	}
}
