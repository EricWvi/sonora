package thumbhash

import (
	"encoding/base64"
	"image"
	"image/color"
	"testing"
)

func pattern(w, h int, alpha bool) []byte {
	px := make([]byte, w*h*4)
	for y := range h {
		for x := range w {
			j := (y*w + x) * 4
			for c := range 3 {
				px[j+c] = byte((x*7 + y*3 + c*50 + x*y*(c+1)) % 256)
			}
			px[j+3] = 255
			if alpha {
				px[j+3] = byte((x*5 + y*9) % 256)
			}
		}
	}
	return px
}

// The expected hashes come from rgbaToThumbHash in the thumbhash npm
// package 0.1.1, given the same pixels.
func TestEncodeMatchesReference(t *testing.T) {
	tests := []struct {
		w, h  int
		alpha bool
		want  string
	}{
		{64, 40, false, "4PcBDIJ2VFZVBjJZR2YUAhVAJA=="},
		{30, 50, true, "IPiBAwAYjRiSdvUgSCZQk2Z3VnhaJRA="},
		{100, 100, false, "4PcBBwB4oSRpmDY0AzRUhqhmSAY4JEMA"},
		{1, 1, true, "AAiCBQAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="},
		{7, 100, false, "3/cFEQRWZggSMoZTdTFgBxQ="},
	}
	for _, tt := range tests {
		got := base64.StdEncoding.EncodeToString(encodeRGBA(tt.w, tt.h, pattern(tt.w, tt.h, tt.alpha)))
		if got != tt.want {
			t.Errorf("%dx%d alpha=%v: %s, want %s", tt.w, tt.h, tt.alpha, got, tt.want)
		}
	}
}

func TestEncodeScalesLargeImages(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 400))
	for y := range 400 {
		for x := range 800 {
			img.Set(x, y, color.RGBA{uint8(x / 4), uint8(y / 2), 128, 255})
		}
	}
	h := Encode(img)
	raw, err := base64.StdEncoding.DecodeString(h)
	if err != nil || len(raw) < 5 || len(raw) > 25 {
		t.Fatalf("hash %q (%d bytes), %v", h, len(raw), err)
	}
	if raw[4]&0x80 == 0 {
		t.Error("landscape flag not set")
	}
	if Encode(image.NewRGBA(image.Rectangle{})) != "" {
		t.Error("empty image should give no hash")
	}
}
