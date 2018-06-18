// Copyright 2018 The oksvg Authors. All rights reserved.
//
// created: 5/12/2018 by S.R.Wiley
package rasterx

import (
	"image/color"
	"math"
	"sort"
)

const (
	ObjectBoundingBox GradientUnits = iota
	UserSpaceOnUse
)

const (
	PadSpread SpreadMethod = iota
	ReflectSpread
	RepeatSpread
)

const epsilonF = 1e-5

type (
	SpreadMethod  byte
	GradientUnits byte

	GradStop struct {
		StopColor color.Color
		Offset    float64
		Opacity   float64
	}

	Gradient struct {
		Points   [5]float64
		Stops    []GradStop
		Bounds   struct{ X, Y, W, H float64 }
		Matrix   Matrix2D
		Spread   SpreadMethod
		Units    GradientUnits
		IsRadial bool
	}
)

func ApplyOpacity(c color.Color, opacity float64) color.NRGBA {
	r, g, b, _ := c.RGBA()
	return color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(opacity * 0xFF)}
}

// tColor takes the paramaterized value along the gradient's stops and
// returns a color depending on the spreadMethod value of the gradient and
// the gradient's slice of stop values.
func (g *Gradient) tColor(t, opacity float64) color.Color {
	d := len(g.Stops)
	// These cases can be taken care of early on
	if t >= 1.0 && g.Spread == PadSpread {
		s := g.Stops[d-1]
		return ApplyOpacity(s.StopColor, s.Opacity*opacity)
	}
	if t <= 0.0 && g.Spread == PadSpread {
		return ApplyOpacity(g.Stops[0].StopColor, g.Stops[0].Opacity*opacity)
	}

	var modRange float64 = 1.0
	if g.Spread == ReflectSpread {
		modRange = 2.0
	}
	mod := math.Mod(t, modRange)
	if mod < 0 {
		mod += modRange
	}

	place := 0 // Advance to place where mod is greater than the indicated stop
	for place != len(g.Stops) && mod > g.Stops[place].Offset {
		place++
	}
	switch g.Spread {
	case RepeatSpread:
		var s1, s2 GradStop
		switch place {
		case 0, d:
			s1, s2 = g.Stops[d-1], g.Stops[0]
		default:
			s1, s2 = g.Stops[place-1], g.Stops[place]
		}
		return g.blendStops(mod, opacity, s1, s2, false)
	case ReflectSpread:
		switch place {
		case 0:
			return ApplyOpacity(g.Stops[0].StopColor, g.Stops[0].Opacity*opacity)
		case d:
			// Advance to place where mod-1 is greater than the stop indicated by place in reverse of the stop slice.
			// Since this is the reflect spead mode, the mod interval is two, allowing the stop list to be
			// iterated in reverse before repeating the sequence.
			for place != d*2 && mod-1 > (1-g.Stops[d*2-place-1].Offset) {
				place++
			}
			switch place {
			case d:
				s := g.Stops[d-1]
				return ApplyOpacity(s.StopColor, s.Opacity*opacity)
			case d * 2:
				return ApplyOpacity(g.Stops[0].StopColor, g.Stops[0].Opacity*opacity)
			default:
				return g.blendStops(mod-1, opacity,
					g.Stops[d*2-place], g.Stops[d*2-place-1], true)
			}
		default:
			return g.blendStops(mod, opacity,
				g.Stops[place-1], g.Stops[place], false)
		}
	default: // PadSpread
		switch place {
		case 0:
			return ApplyOpacity(g.Stops[0].StopColor, g.Stops[0].Opacity*opacity)
		case len(g.Stops):
			s := g.Stops[len(g.Stops)-1]
			return ApplyOpacity(s.StopColor, s.Opacity*opacity)
		default:
			return g.blendStops(mod, opacity, g.Stops[place-1], g.Stops[place], false)
		}
	}
}

func (g *Gradient) blendStops(t, opacity float64, s1, s2 GradStop, flip bool) color.Color {
	s1off := s1.Offset
	if s1.Offset > s2.Offset && !flip { // happens in repeat spread mode
		s1off -= 1
		if t > 1 {
			t -= 1
		}
	}
	if s2.Offset == s1off {
		return ApplyOpacity(s2.StopColor, s2.Opacity)
	}
	if flip {
		t = 1 - t
	}
	tp := (t - s1off) / (s2.Offset - s1off)
	r1, g1, b1, _ := s1.StopColor.RGBA()
	r2, g2, b2, _ := s2.StopColor.RGBA()

	return ApplyOpacity(color.RGBA{
		uint8((float64(r1)*(1-tp) + float64(r2)*tp) / 256),
		uint8((float64(g1)*(1-tp) + float64(g2)*tp) / 256),
		uint8((float64(b1)*(1-tp) + float64(b2)*tp) / 256),
		0xFF}, (s1.Opacity*(1-tp)+s2.Opacity*tp)*opacity)
}

func (g *Gradient) GetColorFunction(opacity float64) interface{} {
	switch len(g.Stops) {
	case 0:
		return ApplyOpacity(color.RGBA{255, 0, 255, 255}, opacity) // default error color for gradient w/o stops.
	case 1:
		return ApplyOpacity(g.Stops[0].StopColor, opacity) // Illegal, I think, should really should not happen.
	}

	// sort by offset in ascending order
	sort.Slice(g.Stops, func(i, j int) bool {
		return g.Stops[i].Offset < g.Stops[j].Offset
	})

	w, h := float64(g.Bounds.W), float64(g.Bounds.H)
	oriX, oriY := float64(g.Bounds.X), float64(g.Bounds.Y)
	gradT := Identity.Translate(oriX, oriY).Scale(w, h).
		Mult(g.Matrix).Scale(1/w, 1/h).Translate(-oriX, -oriY).Invert()

	if g.IsRadial {

		cx := g.Bounds.X + g.Bounds.W*g.Points[0]
		cy := g.Bounds.Y + g.Bounds.H*g.Points[1]
		rx := g.Bounds.W * g.Points[4]
		ry := g.Bounds.H * g.Points[4]

		if g.Points[0] == g.Points[2] && g.Points[1] == g.Points[3] {
			// When the focus and center are the same things are much simpler;
			// t is just distance from center
			// scaled by the bounds aspect ratio times r
			return ColorFunc(func(xi, yi int) color.Color {
				x, y := gradT.Transform(float64(xi)+0.5, float64(yi)+0.5)
				dx := float64(x) - cx
				dy := float64(y) - cy
				return g.tColor(math.Sqrt(dx*dx/(rx*rx)+dy*dy/(ry*ry)), opacity)
			})
		} else {
			fx := g.Bounds.X + g.Bounds.W*g.Points[2]
			fy := g.Bounds.Y + g.Bounds.H*g.Points[3]

			//Scale
			fx /= rx
			fy /= ry
			cx /= rx
			cy /= ry

			dfx := fx - cx
			dfy := fy - cy

			if dfx*dfx+dfy*dfy > 1 { // Focus outside of circle; use intersection
				// point of line from center to focus and circle as per SVG specs.
				nfx, nfy, intersects := RayCircleIntersectionF(fx, fy, cx, cy, cx, cy, 1.0-epsilonF)
				fx, fy = nfx, nfy
				if intersects == false {
					return color.RGBA{255, 255, 0, 255} // should not happen
				}
			}
			return ColorFunc(func(xi, yi int) color.Color {
				x, y := gradT.Transform(float64(xi)+0.5, float64(yi)+0.5)

				ex := float64(x) / rx
				ey := float64(y) / ry

				t1x, t1y, intersects := RayCircleIntersectionF(ex, ey, fx, fy, cx, cy, 1.0)
				if intersects == false { //In this case, use the last stop color
					s := g.Stops[len(g.Stops)-1]
					return ApplyOpacity(s.StopColor, s.Opacity*opacity)
				}
				tdx, tdy := t1x-fx, t1y-fy
				dx, dy := ex-fx, ey-fy
				if tdx*tdx+tdy*tdy < epsilonF {
					s := g.Stops[len(g.Stops)-1]
					return ApplyOpacity(s.StopColor, s.Opacity*opacity)
				}
				return g.tColor(math.Sqrt(dx*dx+dy*dy)/math.Sqrt(tdx*tdx+tdy*tdy), opacity)
			})
		}
	} else {

		p1x := g.Bounds.X + g.Bounds.W*g.Points[0]
		p1y := g.Bounds.Y + g.Bounds.H*g.Points[1]
		p2x := g.Bounds.X + g.Bounds.W*g.Points[2]
		p2y := g.Bounds.Y + g.Bounds.H*g.Points[3]

		dx := p2x - p1x
		dy := p2y - p1y
		d := (dx*dx + dy*dy) // self inner prod
		return ColorFunc(func(xi, yi int) color.Color {
			x, y := gradT.Transform(float64(xi)+0.5, float64(yi)+0.5)
			dfx := x - p1x
			dfy := y - p1y
			return g.tColor((dx*dfx+dy*dfy)/d, opacity)
		})
	}
}
