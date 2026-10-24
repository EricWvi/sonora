// Package thumbhash encodes images as ThumbHash, a ~25-byte summary the
// client decodes into a blurred placeholder while the real image loads.
//
// The encoder is ported from the reference implementation by Evan Wallace
// (https://github.com/evanw/thumbhash, MIT), which the admin console and
// player use to decode.
package thumbhash

import (
	"encoding/base64"
	"image"
	"math"

	"golang.org/x/image/draw"
)

// maxSide is the largest input the encoder accepts; more pixels only cost
// time.
const maxSide = 100

// Encode scales img to fit 100×100 and returns its ThumbHash in standard
// base64.
func Encode(img image.Image) string {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return ""
	}
	if w > maxSide || h > maxSide {
		scale := float64(maxSide) / float64(max(w, h))
		w = max(1, int(math.Round(float64(w)*scale)))
		h = max(1, int(math.Round(float64(h)*scale)))
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, b, draw.Src, nil)
	return base64.StdEncoding.EncodeToString(encodeRGBA(w, h, dst.Pix))
}

// round matches JavaScript's Math.round for the non-negative values the
// encoder rounds.
func round(x float64) int { return int(math.Floor(x + 0.5)) }

// encodeRGBA is rgbaToThumbHash from the reference implementation: rgba
// holds w×h non-premultiplied pixels, w and h at most 100.
func encodeRGBA(w, h int, rgba []byte) []byte {
	n := w * h

	// Determine the average color.
	var avgR, avgG, avgB, avgA float64
	for i, j := 0, 0; i < n; i, j = i+1, j+4 {
		alpha := float64(rgba[j+3]) / 255
		avgR += alpha / 255 * float64(rgba[j])
		avgG += alpha / 255 * float64(rgba[j+1])
		avgB += alpha / 255 * float64(rgba[j+2])
		avgA += alpha
	}
	if avgA > 0 {
		avgR /= avgA
		avgG /= avgA
		avgB /= avgA
	}

	hasAlpha := avgA < float64(n)
	lLimit := 7 // fewer luminance bits if there's alpha
	if hasAlpha {
		lLimit = 5
	}
	lx := max(1, round(float64(lLimit*w)/float64(max(w, h))))
	ly := max(1, round(float64(lLimit*h)/float64(max(w, h))))

	// Convert to LPQA (luminance, yellow-blue, red-green, alpha),
	// composited atop the average color.
	l := make([]float64, n)
	p := make([]float64, n)
	q := make([]float64, n)
	a := make([]float64, n)
	for i, j := 0, 0; i < n; i, j = i+1, j+4 {
		alpha := float64(rgba[j+3]) / 255
		r := avgR*(1-alpha) + alpha/255*float64(rgba[j])
		g := avgG*(1-alpha) + alpha/255*float64(rgba[j+1])
		b := avgB*(1-alpha) + alpha/255*float64(rgba[j+2])
		l[i] = (r + g + b) / 3
		p[i] = (r+g)/2 - b
		q[i] = r - g
		a[i] = alpha
	}

	// Encode with the DCT into DC (constant) and normalised AC (varying)
	// terms.
	encodeChannel := func(channel []float64, nx, ny int) (dc float64, ac []float64, scale float64) {
		fx := make([]float64, w)
		for cy := 0; cy < ny; cy++ {
			for cx := 0; cx*ny < nx*(ny-cy); cx++ {
				for x := range w {
					fx[x] = math.Cos(math.Pi / float64(w) * float64(cx) * (float64(x) + 0.5))
				}
				f := 0.0
				for y := range h {
					fy := math.Cos(math.Pi / float64(h) * float64(cy) * (float64(y) + 0.5))
					for x := range w {
						f += channel[x+y*w] * fx[x] * fy
					}
				}
				f /= float64(n)
				if cx > 0 || cy > 0 {
					ac = append(ac, f)
					scale = max(scale, math.Abs(f))
				} else {
					dc = f
				}
			}
		}
		if scale > 0 {
			for i := range ac {
				ac[i] = 0.5 + 0.5/scale*ac[i]
			}
		}
		return dc, ac, scale
	}
	lDC, lAC, lScale := encodeChannel(l, max(3, lx), max(3, ly))
	pDC, pAC, pScale := encodeChannel(p, 3, 3)
	qDC, qAC, qScale := encodeChannel(q, 3, 3)
	var aDC, aScale float64
	var aAC []float64
	if hasAlpha {
		aDC, aAC, aScale = encodeChannel(a, 5, 5)
	}

	// Write the constants.
	isLandscape := w > h
	header24 := round(63*lDC) | round(31.5+31.5*pDC)<<6 | round(31.5+31.5*qDC)<<12 | round(31*lScale)<<18
	if hasAlpha {
		header24 |= 1 << 23
	}
	header16 := lx
	if isLandscape {
		header16 = ly
	}
	header16 |= round(63*pScale)<<3 | round(63*qScale)<<9
	if isLandscape {
		header16 |= 1 << 15
	}
	hash := []byte{byte(header24), byte(header24 >> 8), byte(header24 >> 16), byte(header16), byte(header16 >> 8)}
	acStart := 5
	if hasAlpha {
		acStart = 6
		hash = append(hash, byte(round(15*aDC)|round(15*aScale)<<4))
	}

	// Write the varying factors, two per byte.
	channels := [][]float64{lAC, pAC, qAC}
	if hasAlpha {
		channels = append(channels, aAC)
	}
	acIndex := 0
	for _, ac := range channels {
		for _, f := range ac {
			i := acStart + acIndex>>1
			for len(hash) <= i {
				hash = append(hash, 0)
			}
			hash[i] |= byte(round(15*f) << ((acIndex & 1) << 2))
			acIndex++
		}
	}
	return hash
}
