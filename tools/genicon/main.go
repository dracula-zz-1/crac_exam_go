package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const S = 2048 // supersample
const N = 1024

func clamp(v, a, b float64) float64 {
	if v < a {
		return a
	}
	if v > b {
		return b
	}
	return v
}
func smooth(d float64) float64 { return clamp(0.5-d, 0, 1) }
func mix(a, b color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(a.A)*t),
	}
}
func over(dst color.NRGBA, src color.NRGBA, a float64) color.NRGBA {
	sa := float64(src.A) / 255 * a
	da := float64(dst.A) / 255
	oa := sa + da*(1-sa)
	if oa == 0 {
		return color.NRGBA{}
	}
	return color.NRGBA{
		R: uint8(clamp((float64(src.R)*sa+float64(dst.R)*da*(1-sa))/oa, 0, 255)),
		G: uint8(clamp((float64(src.G)*sa+float64(dst.G)*da*(1-sa))/oa, 0, 255)),
		B: uint8(clamp((float64(src.B)*sa+float64(dst.B)*da*(1-sa))/oa, 0, 255)),
		A: uint8(oa * 255),
	}
}

// sdf for rounded box centered 0,0 half-size h, radius r
func sdfBox(x, y, h, r float64) float64 {
	qx := math.Abs(x) - h + r
	qy := math.Abs(y) - h + r
	out := math.Hypot(math.Max(qx, 0), math.Max(qy, 0)) + math.Min(math.Max(qx, qy), 0) - r
	return out
}
func ring(x, y, rad, thick float64) float64 {
	d := math.Abs(math.Hypot(x, y) - rad)
	return d - thick/2
}

func main() {
	img := image.NewNRGBA(image.Rect(0, 0, S, S))
	bg := color.NRGBA{0x11, 0x1a, 0x22, 0xff}
	amber := color.NRGBA{0xe8, 0xa3, 0x3d, 0xff}
	teal := color.NRGBA{0x2a, 0x9d, 0x8f, 0xff}
	for py := 0; py < S; py++ {
		for px := 0; px < S; px++ {
			x := float64(px)/S*2 - 1
			y := float64(py)/S*2 - 1
			c := color.NRGBA{0, 0, 0, 0}
			// rounded square
			d := sdfBox(x, y, 0.94, 0.30)
			c = over(c, bg, smooth(d*(S/2)))
			// subtle vertical gradient overlay
			g := color.NRGBA{0x1b, 0x28, 0x33, 0xff}
			c = over(c, g, 0.35*(1-(y+1)/2))
			// signal rings
			for i, rad := range []float64{0.20, 0.34, 0.48} {
				a := 0.95 - float64(i)*0.22
				c = over(c, amber, smooth(ring(x, y, rad, 0.052)*(S/2))*a)
			}
			// center dot
			c = over(c, amber, smooth((math.Hypot(x, y)-0.105)*(S/2)))
			// waveform below
			if y > 0.46 && y < 0.62 {
				wave := 0.52 + 0.035*math.Sin(x*11)
				if math.Abs(y-wave) < 0.012 {
					c = over(c, teal, 0.85*smooth((math.Abs(y-wave)-0.012)*(S/2)*-1))
				}
			}
			img.SetNRGBA(px, py, c)
		}
	}
	// box downsample to 1024
	out := image.NewNRGBA(image.Rect(0, 0, N, N))
	factor := S / N
	for y := 0; y < N; y++ {
		for x := 0; x < N; x++ {
			var r, g, b, a, n float64
			for dy := 0; dy < factor; dy++ {
				for dx := 0; dx < factor; dx++ {
					c := img.NRGBAAt(x*factor+dx, y*factor+dy)
					af := float64(c.A) / 255
					r += float64(c.R) * af
					g += float64(c.G) * af
					b += float64(c.B) * af
					a += af
					n++
				}
			}
			if a == 0 {
				continue
			}
			out.SetNRGBA(x, y, color.NRGBA{uint8(r / a), uint8(g / a), uint8(b / a), uint8(a / n * 255)})
		}
	}
	f, err := os.Create(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, out); err != nil {
		panic(err)
	}
}
