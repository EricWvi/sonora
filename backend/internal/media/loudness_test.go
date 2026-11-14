package media

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// The want values are what ffmpeg's ebur128 filter (peak=true) reads for
// the samples, plus 3.01 dB of loudness: the samples are mono, which
// ffmpeg measures as one channel and Sonora as dual mono.
func TestLoudnessOfSamples(t *testing.T) {
	for _, c := range []struct {
		file           string
		format         Format
		lufs, truePeak float64
	}{
		{"tagged.mp3", MP3, -22.2 + 3.01, -18.5},
		{"untagged.mp3", MP3, -22.2 + 3.01, -18.5},
		{"vbr.mp3", MP3, -21.7 + 3.01, -17.8},
		{"tagged.flac", FLAC, -21.8 + 3.01, -18.1},
		{"tagged.ogg", OGG, -21.7 + 3.01, -17.8},
	} {
		data, err := os.ReadFile(filepath.Join("testdata", c.file))
		if err != nil {
			t.Fatal(err)
		}
		r, err := Loudness(bytes.NewReader(data), c.format)
		if err != nil {
			t.Errorf("%s: %v", c.file, err)
			continue
		}
		l, ok := r.Integrated()
		if !ok || l < c.lufs-0.15 || l > c.lufs+0.15 {
			t.Errorf("%s: loudness %.2f (%v), want %.2f", c.file, l, ok, c.lufs)
		}
		// ffmpeg rounds to 0.1 dB.
		if tp := r.TruePeakDB(); tp < c.truePeak-0.15 || tp > c.truePeak+0.15 {
			t.Errorf("%s: true peak %.2f, want %.2f", c.file, tp, c.truePeak)
		}
	}
}

func TestLoudnessUnsupportedAndBroken(t *testing.T) {
	for _, f := range []Format{M4A, OPUS} {
		if _, err := Loudness(bytes.NewReader(nil), f); !errors.Is(err, ErrUnsupported) {
			t.Errorf("%s: err = %v, want ErrUnsupported", f, err)
		}
	}
	if _, err := Loudness(bytes.NewReader([]byte("not audio at all")), MP3); !errors.Is(err, ErrCorrupt) {
		t.Errorf("garbage MP3: err = %v, want ErrCorrupt", err)
	}
}
