package loudness

import "math"

// Channel weights of BS.1770 for the channel positions the decoders
// report: front and centre count once, surround 1.41 times, LFE not at all.
const (
	Front    = 1.0
	Surround = 1.41
	LFE      = 0.0
	// DualMono weights the one channel of a mono programme as the two
	// speakers it plays on, as EBU R128 s1 and ReplayGain 2 do: a mono
	// file plays at full level on both, 3 dB louder than one channel.
	DualMono = 2.0
)

// Meter measures interleaved PCM. Write any number of times, then call
// Result.
type Meter struct {
	weights []float64
	k       []kWeighting
	peaks   []truePeak

	// The 400 ms blocks step by 100 ms; sub holds the weighted energy of
	// the last four 100 ms sub-blocks, the newest at subs%4.
	step  int // frames per sub-block
	acc   float64
	accN  int
	sub   [4]float64
	subs  int
	hist  Histogram
	frame []float64
}

// NewMeter measures rate Hz audio with one channel per weight, in the
// order the samples interleave them.
func NewMeter(rate int, weights []float64) *Meter {
	m := &Meter{
		weights: weights,
		k:       make([]kWeighting, len(weights)),
		peaks:   make([]truePeak, len(weights)),
		step:    (rate + 5) / 10,
		frame:   make([]float64, len(weights)),
	}
	for c := range weights {
		m.k[c] = newKWeighting(float64(rate))
		m.peaks[c] = newTruePeak(rate)
	}
	return m
}

// WriteFloat32 takes interleaved samples in [-1, 1]; a trailing partial
// frame is ignored.
func (m *Meter) WriteFloat32(samples []float32) {
	nc := len(m.weights)
	for i := 0; i+nc <= len(samples); i += nc {
		for c := range nc {
			m.frame[c] = float64(samples[i+c])
		}
		m.writeFrame()
	}
}

// WriteFloat64 is WriteFloat32 for float64 samples.
func (m *Meter) WriteFloat64(samples []float64) {
	nc := len(m.weights)
	for i := 0; i+nc <= len(samples); i += nc {
		copy(m.frame, samples[i:i+nc])
		m.writeFrame()
	}
}

func (m *Meter) writeFrame() {
	var e float64
	for c, x := range m.frame {
		m.peaks[c].write(x)
		if w := m.weights[c]; w != 0 {
			y := m.k[c].filter(x)
			e += w * y * y
		}
	}
	m.acc += e
	if m.accN++; m.accN < m.step {
		return
	}
	m.sub[m.subs%4] = m.acc
	m.subs++
	m.acc, m.accN = 0, 0
	if m.subs >= 4 {
		m.hist.add((m.sub[0] + m.sub[1] + m.sub[2] + m.sub[3]) / float64(4*m.step))
	}
}

// Result is the measurement so far. Samples after the last whole 100 ms
// are not in a block yet, as the standard leaves a partial block out.
func (m *Meter) Result() Result {
	r := Result{Blocks: m.hist}
	for i := range m.peaks {
		r.TruePeak = max(r.TruePeak, m.peaks[i].peak)
	}
	return r
}

// --- K-weighting ----------------------------------------------------------

// biquad is a second-order IIR section, transposed direct form II.
type biquad struct {
	b0, b1, b2, a1, a2 float64
	z1, z2             float64
}

func (q *biquad) filter(x float64) float64 {
	y := q.b0*x + q.z1
	q.z1 = q.b1*x - q.a1*y + q.z2
	q.z2 = q.b2*x - q.a2*y
	return y
}

// kWeighting is the shelving pre-filter followed by the RLB high-pass.
type kWeighting struct{ shelf, highPass biquad }

// newKWeighting derives both stages for any sample rate from the analog
// prototypes, as libebur128 does; at 48 kHz they match the coefficients
// BS.1770 lists.
func newKWeighting(rate float64) kWeighting {
	var k kWeighting
	{
		const f0, gain, q = 1681.974450955533, 3.999843853973347, 0.7071752369554196
		K := math.Tan(math.Pi * f0 / rate)
		vh := math.Pow(10, gain/20)
		vb := math.Pow(vh, 0.4996667741545416)
		a0 := 1 + K/q + K*K
		k.shelf = biquad{
			b0: (vh + vb*K/q + K*K) / a0,
			b1: 2 * (K*K - vh) / a0,
			b2: (vh - vb*K/q + K*K) / a0,
			a1: 2 * (K*K - 1) / a0,
			a2: (1 - K/q + K*K) / a0,
		}
	}
	{
		const f0, q = 38.13547087602444, 0.5003270373238773
		K := math.Tan(math.Pi * f0 / rate)
		a0 := 1 + K/q + K*K
		k.highPass = biquad{
			b0: 1, b1: -2, b2: 1,
			a1: 2 * (K*K - 1) / a0,
			a2: (1 - K/q + K*K) / a0,
		}
	}
	return k
}

func (k *kWeighting) filter(x float64) float64 { return k.highPass.filter(k.shelf.filter(x)) }

// --- True peak ------------------------------------------------------------

// tapsPerPhase is the length of each polyphase branch of the oversampling
// filter; 12 matches the 48-tap filter of BS.1770 Annex 2 at 4x.
const tapsPerPhase = 12

// truePeak oversamples one channel to at least 176.4 kHz and keeps the
// highest absolute value, the sample peak included.
type truePeak struct {
	phases [][tapsPerPhase]float64
	// bound is the most any phase can amplify: no output exceeds bound
	// times the largest input in the history, which lets the filter be
	// skipped while the history is too quiet to raise the peak.
	bound float64
	hist  [2 * tapsPerPhase]float64 // the history twice over, so a window is contiguous
	pos   int
	peak  float64
}

func newTruePeak(rate int) truePeak {
	factor := 4
	switch {
	case rate >= 176400:
		factor = 1
	case rate >= 88200:
		factor = 2
	}
	t := truePeak{phases: oversamplingFilter(factor)}
	for _, ph := range t.phases {
		var s float64
		for _, h := range ph {
			s += math.Abs(h)
		}
		t.bound = max(t.bound, s)
	}
	return t
}

// oversamplingFilter splits a windowed-sinc low-pass at the original
// Nyquist frequency into factor phases, each scaled to unity gain at DC.
// Phase p, applied to the newest input first, yields the output p/factor
// of a sample after the group delay.
func oversamplingFilter(factor int) [][tapsPerPhase]float64 {
	n := factor * tapsPerPhase
	centre := float64(n-1) / 2
	phases := make([][tapsPerPhase]float64, factor)
	for p := range factor {
		var sum float64
		for k := range tapsPerPhase {
			i := k*factor + p
			x := (float64(i) - centre) / float64(factor)
			h := 1.0
			if x != 0 {
				h = math.Sin(math.Pi*x) / (math.Pi * x)
			}
			// Blackman window over the whole prototype.
			w := 0.42 - 0.5*math.Cos(2*math.Pi*(float64(i)+0.5)/float64(n)) +
				0.08*math.Cos(4*math.Pi*(float64(i)+0.5)/float64(n))
			phases[p][k] = h * w
			sum += h * w
		}
		for k := range tapsPerPhase {
			phases[p][k] /= sum
		}
	}
	return phases
}

func (t *truePeak) write(x float64) {
	ax := math.Abs(x)
	t.peak = max(t.peak, ax)
	if len(t.phases) == 1 {
		return
	}
	// hist[pos:pos+tapsPerPhase] runs from the newest sample to the oldest.
	t.pos = (t.pos + tapsPerPhase - 1) % tapsPerPhase
	t.hist[t.pos] = x
	t.hist[t.pos+tapsPerPhase] = x
	win := t.hist[t.pos : t.pos+tapsPerPhase]
	var loudest float64
	for _, v := range win {
		loudest = max(loudest, math.Abs(v))
	}
	if loudest*t.bound <= t.peak {
		return
	}
	for p := range t.phases {
		ph := &t.phases[p]
		var y float64
		for k, v := range win {
			y += ph[k] * v
		}
		t.peak = max(t.peak, math.Abs(y))
	}
}
