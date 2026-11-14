package loudness

import (
	"math"
	"testing"
)

// segment is a stretch of a sine wave at a level in dBFS (peak amplitude,
// as EBU Tech 3341 states its test signals).
type segment struct {
	dbfs    float64
	seconds float64
}

// sine renders the segments as interleaved samples, the same sine on every
// channel, keeping the phase continuous across segments.
func sine(rate, channels int, freq float64, segs ...segment) []float64 {
	var out []float64
	n := 0
	for _, s := range segs {
		amp := math.Pow(10, s.dbfs/20)
		if math.IsInf(s.dbfs, -1) {
			amp = 0
		}
		for range int(s.seconds * float64(rate)) {
			v := amp * math.Sin(2*math.Pi*freq*float64(n)/float64(rate))
			for range channels {
				out = append(out, v)
			}
			n++
		}
	}
	return out
}

func measure(rate int, weights []float64, samples []float64) Result {
	m := NewMeter(rate, weights)
	m.WriteFloat64(samples)
	return m.Result()
}

var stereo = []float64{Front, Front}

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %.3f, want %.3f ± %.2f", name, got, want, tol)
	}
}

func integrated(t *testing.T, r Result) float64 {
	t.Helper()
	l, ok := r.Integrated()
	if !ok {
		t.Fatal("no loudness measured")
	}
	return l
}

// EBU Tech 3341 cases 1 and 2: a stereo 1 kHz sine reads its own level.
func TestSineReadsItsLevel(t *testing.T) {
	for _, rate := range []int{44100, 48000, 32000, 22050} {
		for _, level := range []float64{-23, -33} {
			r := measure(rate, stereo, sine(rate, 2, 1000, segment{level, 20}))
			near(t, "integrated", integrated(t, r), level, 0.1)
		}
	}
}

// EBU Tech 3341 cases 3 and 4: quiet passages around the programme fall
// below the relative gate and do not lower it.
func TestRelativeGateDropsQuietPassages(t *testing.T) {
	for _, quiet := range []float64{-36, -72} {
		r := measure(48000, stereo, sine(48000, 2, 1000, segment{quiet, 10}, segment{-23, 60}, segment{quiet, 10}))
		near(t, "integrated", integrated(t, r), -23, 0.1)
	}
}

// EBU Tech 3341 case 5: passages within 10 LU of the mean all count.
func TestRelativeGateKeepsNearbyPassages(t *testing.T) {
	r := measure(48000, stereo, sine(48000, 2, 1000, segment{-26, 20}, segment{-20, 20.1}, segment{-26, 20}))
	near(t, "integrated", integrated(t, r), -23, 0.1)
}

// Silence and passages below -70 LUFS leave nothing to measure.
func TestAbsoluteGate(t *testing.T) {
	r := measure(48000, stereo, sine(48000, 2, 1000, segment{math.Inf(-1), 5}, segment{-80, 5}))
	if l, ok := r.Integrated(); ok {
		t.Errorf("integrated = %.2f, want none", l)
	}
	if r.Blocks.Len() != 0 {
		t.Errorf("blocks = %d, want 0", r.Blocks.Len())
	}
	// Too short for one 400 ms block.
	r = measure(48000, stereo, sine(48000, 2, 1000, segment{-23, 0.35}))
	if _, ok := r.Integrated(); ok {
		t.Error("a 350 ms programme has a loudness")
	}
}

// A mono programme counts as dual mono, so it reads as the same sine on
// two channels would.
func TestDualMono(t *testing.T) {
	r := measure(44100, []float64{DualMono}, sine(44100, 1, 1000, segment{-23, 10}))
	near(t, "integrated", integrated(t, r), -23, 0.1)
}

// The LFE channel does not count and surrounds weigh 1.41 (+1.5 dB).
func TestChannelWeights(t *testing.T) {
	r := measure(48000, []float64{Front, LFE}, sine(48000, 2, 1000, segment{-20, 10}))
	near(t, "front + LFE", integrated(t, r), -23.01, 0.1)
	r = measure(48000, []float64{Surround}, sine(48000, 1, 1000, segment{-20, 10}))
	near(t, "surround", integrated(t, r), -23.01+10*math.Log10(1.41), 0.1)
}

// K-weighting lifts high frequencies by about 4 dB and cuts the lowest.
func TestKWeightingResponse(t *testing.T) {
	r := measure(48000, stereo, sine(48000, 2, 10000, segment{-23, 10}))
	// The shelf lifts 10 kHz by 4.04 dB; -0.691 dB of the 1 kHz gain is
	// taken back by the offset.
	near(t, "10 kHz", integrated(t, r), -23+4.04-0.69, 0.1)
	r = measure(48000, stereo, sine(48000, 2, 20, segment{-23, 10}))
	if l := integrated(t, r); l > -23-10 {
		t.Errorf("20 Hz reads %.2f, want well below -33", l)
	}
}

// At 48 kHz the derived filters are those BS.1770 lists.
func TestKWeightingCoefficients(t *testing.T) {
	k := newKWeighting(48000)
	const tol = 1e-8
	for _, c := range []struct {
		name      string
		got, want float64
	}{
		{"shelf b0", k.shelf.b0, 1.53512485958697},
		{"shelf b1", k.shelf.b1, -2.69169618940638},
		{"shelf b2", k.shelf.b2, 1.19839281085285},
		{"shelf a1", k.shelf.a1, -1.69065929318241},
		{"shelf a2", k.shelf.a2, 0.73248077421585},
		{"high-pass a1", k.highPass.a1, -1.99004745483398},
		{"high-pass a2", k.highPass.a2, 0.99007225036621},
	} {
		if math.Abs(c.got-c.want) > tol {
			t.Errorf("%s = %.14f, want %.14f", c.name, c.got, c.want)
		}
	}
}

// EBU Tech 3341 true-peak cases put a sine at a quarter of the sample rate
// 45° out of phase with the samples, so every sample misses the crest by
// 3 dB. BS.1770 allows +0.2/-0.4 dB at 4x oversampling.
func TestTruePeakBetweenSamples(t *testing.T) {
	for _, rate := range []int{44100, 48000} {
		const amp = 0.5
		var samples []float64
		for n := range rate {
			v := amp * math.Sin(math.Pi/2*float64(n)+math.Pi/4)
			samples = append(samples, v, v)
		}
		r := measure(rate, stereo, samples)
		near(t, "true peak dBTP", r.TruePeakDB(), 20*math.Log10(amp), 0.2)
	}
}

// Other frequencies and phases stay within the tolerance, and a true peak
// is never below the sample peak.
func TestTruePeakTolerance(t *testing.T) {
	const rate, amp = 48000, 0.25
	for _, freq := range []float64{100, 997, 5000, 11025, 15000} {
		for _, phase := range []float64{0, 0.3, 0.77} {
			var samples []float64
			var samplePeak float64
			for n := range rate / 2 {
				// Fade in over 10 ms: a sine switched on at full level is
				// not band-limited, and the interpolation rings at the step.
				v := amp * math.Sin(2*math.Pi*freq*float64(n)/rate+phase) * fadeIn(n, rate/100)
				samplePeak = max(samplePeak, math.Abs(v))
				samples = append(samples, v)
			}
			r := measure(rate, []float64{Front}, samples)
			got := r.TruePeakDB() - 20*math.Log10(amp)
			if got > 0.2 || got < -0.4 {
				t.Errorf("%g Hz phase %g: true peak off by %.3f dB", freq, phase, got)
			}
			if r.TruePeak < samplePeak {
				t.Errorf("%g Hz: true peak %.4f below sample peak %.4f", freq, r.TruePeak, samplePeak)
			}
		}
	}
}

func fadeIn(n, length int) float64 {
	if n >= length {
		return 1
	}
	return 0.5 - 0.5*math.Cos(math.Pi*float64(n)/float64(length))
}

func TestSilencePeak(t *testing.T) {
	r := measure(48000, stereo, make([]float64, 96000))
	if r.TruePeak != 0 || !math.IsInf(r.TruePeakDB(), -1) {
		t.Errorf("silence peaks at %v", r.TruePeak)
	}
}
