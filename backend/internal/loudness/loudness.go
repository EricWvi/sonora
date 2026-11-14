// Package loudness measures programme loudness and true peak as ITU-R
// BS.1770-5 and EBU R128 define them.
//
// A Meter takes PCM samples and keeps the energy of every 400 ms gating
// block (75 % overlap) in a Histogram. Integrated loudness gates the
// blocks at -70 LUFS, then 10 LU below the mean of what remains.
package loudness

import "math"

const (
	absoluteGate = -70.0 // LUFS
	relativeGate = -10.0 // LU below the absolute-gated mean
	// offset is the -0.691 dB of BS.1770 that makes a 1 kHz sine read its
	// own level after K-weighting.
	offset = -0.691
)

// Result is the measurement of one programme.
type Result struct {
	Blocks Histogram
	// TruePeak is the highest absolute sample value after 4x
	// oversampling, linear, 1 being full scale. It is at least the sample
	// peak.
	TruePeak float64
}

// Integrated is the gated loudness in LUFS; ok is false when no block is
// louder than -70 LUFS, as for silence or a programme shorter than 400 ms.
func (r Result) Integrated() (lufs float64, ok bool) { return r.Blocks.Integrated() }

// TruePeakDB is the true peak in dBTP, -Inf for silence.
func (r Result) TruePeakDB() float64 { return 20 * math.Log10(r.TruePeak) }

// The histogram bins block loudness in 0.1 LU steps from the absolute gate
// up; louder blocks than the top bin, which digital audio cannot reach,
// share it. Each bin sums the energies of its blocks, so gated means are
// exact except for the one bin a relative threshold falls in.
const (
	binWidth = 0.1
	binCount = 800 // -70 to +10 LUFS
)

// Histogram holds the gating blocks of a programme, blocks at or below the
// absolute gate left out. The zero value is empty.
type Histogram struct {
	count  [binCount]uint32
	energy [binCount]float64
}

// add records a block of mean-square energy e (the channel-weighted sum).
func (h *Histogram) add(e float64) {
	l := offset + 10*math.Log10(e)
	if !(l > absoluteGate) { // also drops NaN
		return
	}
	i := min(int((l-absoluteGate)/binWidth), binCount-1)
	h.count[i]++
	h.energy[i] += e
}

// Len is the number of blocks above the absolute gate.
func (h *Histogram) Len() int {
	n := 0
	for _, c := range h.count {
		n += int(c)
	}
	return n
}

// Integrated is the gated loudness of the blocks in LUFS; ok is false when
// there are none.
func (h *Histogram) Integrated() (lufs float64, ok bool) {
	var n, sum float64
	for i, c := range h.count {
		n += float64(c)
		sum += h.energy[i]
	}
	if n == 0 {
		return 0, false
	}
	threshold := offset + 10*math.Log10(sum/n) + relativeGate
	n, sum = 0, 0
	for i, c := range h.count {
		// A bin counts when the mean loudness of its blocks passes the
		// threshold; only the bin the threshold falls in can be off.
		if c == 0 || offset+10*math.Log10(h.energy[i]/float64(c)) <= threshold {
			continue
		}
		n += float64(c)
		sum += h.energy[i]
	}
	if n == 0 {
		return 0, false
	}
	return offset + 10*math.Log10(sum/n), true
}
