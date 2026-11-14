package media

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/oggvorbis"
	"github.com/mewkiz/flac"

	"sonora/internal/loudness"
)

// Loudness decodes the audio in r, a file of format f, and measures its
// loudness and true peak. MP3, FLAC and Ogg Vorbis are decoded; other
// formats, for which no pure Go decoder is good enough, report
// ErrUnsupported. A file that fails to decode reports ErrCorrupt.
//
// Unlike Probe this decodes every frame, which takes about a second for a
// four-minute song.
func Loudness(r io.ReadSeeker, f Format) (loudness.Result, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return loudness.Result{}, err
	}
	br := bufio.NewReaderSize(r, 64<<10)
	var res loudness.Result
	var err error
	switch f {
	case MP3:
		res, err = mp3Loudness(br)
	case FLAC:
		res, err = flacLoudness(br)
	case OGG:
		res, err = vorbisLoudness(br)
	default:
		return loudness.Result{}, ErrUnsupported
	}
	if err != nil && !errors.Is(err, ErrCorrupt) {
		err = fmt.Errorf("%w: %v", ErrCorrupt, err)
	}
	return res, err
}

func mp3Loudness(r io.Reader) (loudness.Result, error) {
	d, err := mp3.NewDecoder(r)
	if err != nil {
		return loudness.Result{}, err
	}
	// The decoder always yields 16-bit stereo, a mono file doubled onto
	// both channels, which is how dual mono is measured anyway.
	m := loudness.NewMeter(d.SampleRate(), []float64{loudness.Front, loudness.Front})
	buf := make([]byte, 32<<10)
	samples := make([]float32, len(buf)/2)
	pending := 0 // bytes of a partial sample carried to the next read
	for {
		n, err := d.Read(buf[pending:])
		n += pending
		whole := n &^ 1
		for i := 0; i < whole; i += 2 {
			samples[i/2] = float32(int16(binary.LittleEndian.Uint16(buf[i:]))) / 32768
		}
		m.WriteFloat32(samples[:whole/2])
		pending = copy(buf, buf[whole:n])
		if errors.Is(err, io.EOF) {
			return m.Result(), nil
		}
		if err != nil {
			return loudness.Result{}, err
		}
	}
}

func flacLoudness(r io.Reader) (loudness.Result, error) {
	s, err := flac.New(r)
	if err != nil {
		return loudness.Result{}, err
	}
	nc, bps := int(s.Info.NChannels), int(s.Info.BitsPerSample)
	m := loudness.NewMeter(int(s.Info.SampleRate), flacWeights(nc))
	scale := 1 / float64(int64(1)<<(bps-1))
	var samples []float64
	for {
		fr, err := s.ParseNext()
		if errors.Is(err, io.EOF) {
			return m.Result(), nil
		}
		if err != nil {
			return loudness.Result{}, err
		}
		if len(fr.Subframes) != nc {
			return loudness.Result{}, corrupt("FLAC 帧的声道数与 STREAMINFO 不符")
		}
		n := int(fr.BlockSize)
		samples = samples[:0]
		for i := range n {
			for c := range nc {
				samples = append(samples, float64(fr.Subframes[c].Samples[i])*scale)
			}
		}
		m.WriteFloat64(samples)
	}
}

func vorbisLoudness(r io.Reader) (loudness.Result, error) {
	d, err := oggvorbis.NewReader(r)
	if err != nil {
		return loudness.Result{}, err
	}
	nc := d.Channels()
	m := loudness.NewMeter(d.SampleRate(), vorbisWeights(nc))
	buf := make([]float32, 4096*nc)
	for {
		n, err := d.Read(buf)
		m.WriteFloat32(buf[:n])
		if errors.Is(err, io.EOF) {
			return m.Result(), nil
		}
		if err != nil {
			return loudness.Result{}, err
		}
	}
}

// flacWeights follows the FLAC channel order: L R C LFE, then the
// surrounds.
func flacWeights(n int) []float64 {
	const f, s, lfe = loudness.Front, loudness.Surround, loudness.LFE
	switch n {
	case 1:
		return []float64{loudness.DualMono}
	case 3:
		return []float64{f, f, f}
	case 4:
		return []float64{f, f, s, s}
	case 5:
		return []float64{f, f, f, s, s}
	case 6:
		return []float64{f, f, f, lfe, s, s}
	case 7:
		return []float64{f, f, f, lfe, s, s, s}
	case 8:
		return []float64{f, f, f, lfe, s, s, s, s}
	}
	return frontWeights(n)
}

// vorbisWeights follows the Vorbis channel order: L C R, then the
// surrounds, LFE last.
func vorbisWeights(n int) []float64 {
	const f, s, lfe = loudness.Front, loudness.Surround, loudness.LFE
	switch n {
	case 1:
		return []float64{loudness.DualMono}
	case 3:
		return []float64{f, f, f}
	case 4:
		return []float64{f, f, s, s}
	case 5:
		return []float64{f, f, f, s, s}
	case 6:
		return []float64{f, f, f, s, s, lfe}
	case 7:
		return []float64{f, f, f, s, s, s, lfe}
	case 8:
		return []float64{f, f, f, s, s, s, s, lfe}
	}
	return frontWeights(n)
}

func frontWeights(n int) []float64 {
	w := make([]float64, n)
	for i := range w {
		w[i] = loudness.Front
	}
	return w
}
