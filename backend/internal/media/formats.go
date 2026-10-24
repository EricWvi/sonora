package media

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"
)

// --- MP3 -------------------------------------------------------------------

// mp3Frame is a parsed MPEG audio frame header.
type mp3Frame struct {
	version    int // 1, 2 or 25 (MPEG 2.5)
	layer      int
	bitrate    int // bits per second
	sampleRate int
	mono       bool
	padding    int
}

var mp3Bitrates = map[[2]int][16]int{
	{1, 1}: {0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448},
	{1, 2}: {0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384},
	{1, 3}: {0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320},
	{2, 1}: {0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256},
	{2, 2}: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},
	{2, 3}: {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},
}

var mp3SampleRates = map[int][3]int{
	1:  {44100, 48000, 32000},
	2:  {22050, 24000, 16000},
	25: {11025, 12000, 8000},
}

func isMP3Sync(b []byte) bool {
	_, ok := parseMP3Frame(b)
	return ok
}

func parseMP3Frame(b []byte) (mp3Frame, bool) {
	if len(b) < 4 || b[0] != 0xFF || b[1]&0xE0 != 0xE0 {
		return mp3Frame{}, false
	}
	var f mp3Frame
	switch (b[1] >> 3) & 3 {
	case 0:
		f.version = 25
	case 2:
		f.version = 2
	case 3:
		f.version = 1
	default:
		return f, false
	}
	switch (b[1] >> 1) & 3 {
	case 1:
		f.layer = 3
	case 2:
		f.layer = 2
	case 3:
		f.layer = 1
	default:
		return f, false
	}
	bi, si := int(b[2]>>4), int((b[2]>>2)&3)
	if bi == 0 || bi == 15 || si == 3 {
		return f, false
	}
	tableVersion := f.version
	if tableVersion == 25 {
		tableVersion = 2
	}
	f.bitrate = mp3Bitrates[[2]int{tableVersion, f.layer}][bi] * 1000
	f.sampleRate = mp3SampleRates[f.version][si]
	f.padding = int((b[2] >> 1) & 1)
	f.mono = b[3]>>6 == 3
	return f, true
}

func (f mp3Frame) samples() int {
	switch {
	case f.layer == 1:
		return 384
	case f.layer == 3 && f.version != 1:
		return 576
	}
	return 1152
}

func (f mp3Frame) length() int {
	if f.layer == 1 {
		return (12*f.bitrate/f.sampleRate + f.padding) * 4
	}
	return f.samples()/8*f.bitrate/f.sampleRate + f.padding
}

// xingOffset is where a Xing/Info header starts inside the first frame:
// after the 4-byte header and the side information.
func (f mp3Frame) xingOffset() int {
	switch {
	case f.version == 1 && !f.mono:
		return 4 + 32
	case f.version != 1 && f.mono:
		return 4 + 9
	}
	return 4 + 17
}

func mp3Duration(r io.ReadSeeker, size int64) (time.Duration, error) {
	start := int64(0)
	head := make([]byte, 10)
	if err := readAt(r, 0, head); err != nil {
		return 0, err
	}
	if hasPrefix(head, "ID3") {
		tagSize := int64(head[6]&0x7F)<<21 | int64(head[7]&0x7F)<<14 | int64(head[8]&0x7F)<<7 | int64(head[9]&0x7F)
		start = 10 + tagSize
		if head[5]&0x10 != 0 {
			start += 10 // footer
		}
	}
	if start >= size {
		return 0, corrupt("没有音频数据")
	}

	// Look for two consecutive valid frame headers so stray 0xFF bytes in
	// padding are not mistaken for audio.
	window := make([]byte, min(int64(64<<10), size-start))
	if err := readAt(r, start, window); err != nil {
		return 0, err
	}
	pos, frame := -1, mp3Frame{}
	for i := 0; i+4 <= len(window); i++ {
		f, ok := parseMP3Frame(window[i:])
		if !ok {
			continue
		}
		next := i + f.length()
		if next+4 <= len(window) {
			if _, ok := parseMP3Frame(window[next:]); !ok {
				continue
			}
		} else if int64(next) < size-start {
			continue
		}
		pos, frame = i, f
		break
	}
	if pos < 0 {
		return 0, corrupt("找不到 MP3 音频帧")
	}

	first := window[pos:]
	if d, ok := mp3XingDuration(frame, first); ok {
		return d, nil
	}
	// VBRI sits at a fixed offset after the 4-byte header and 32 bytes.
	if len(first) >= 36+18 && string(first[36:40]) == "VBRI" {
		frames := binary.BigEndian.Uint32(first[36+14:])
		return seconds(uint64(frames)*uint64(frame.samples()), uint32(frame.sampleRate)), nil
	}

	// Constant bitrate: the audio bytes divided by the bitrate.
	audio := size - start - int64(pos)
	tail := make([]byte, 3)
	if size >= 128 {
		if err := readAt(r, size-128, tail); err != nil {
			return 0, err
		}
		if string(tail) == "TAG" {
			audio -= 128
		}
	}
	return time.Duration(float64(audio*8) / float64(frame.bitrate) * float64(time.Second)), nil
}

// mp3XingDuration reads the frame count, and the LAME encoder delay and
// padding when present, from a Xing or Info header.
func mp3XingDuration(f mp3Frame, frame []byte) (time.Duration, bool) {
	off := f.xingOffset()
	if len(frame) < off+8 {
		return 0, false
	}
	id := string(frame[off : off+4])
	if id != "Xing" && id != "Info" {
		return 0, false
	}
	flags := binary.BigEndian.Uint32(frame[off+4:])
	if flags&1 == 0 || len(frame) < off+12 {
		return 0, false
	}
	frames := uint64(binary.BigEndian.Uint32(frame[off+8:]))
	samples := frames * uint64(f.samples())

	lame := off + 8 + 4
	if flags&2 != 0 {
		lame += 4
	}
	if flags&4 != 0 {
		lame += 100
	}
	if flags&8 != 0 {
		lame += 4
	}
	if len(frame) >= lame+24 && (bytes.HasPrefix(frame[lame:], []byte("LAME")) || bytes.HasPrefix(frame[lame:], []byte("Lavc"))) {
		p := frame[lame+21:]
		delay := uint64(p[0])<<4 | uint64(p[1])>>4
		pad := uint64(p[1]&0x0F)<<8 | uint64(p[2])
		if delay+pad < samples {
			samples -= delay + pad
		}
	}
	return seconds(samples, uint32(f.sampleRate)), true
}

// --- FLAC ------------------------------------------------------------------

func flacDuration(r io.ReadSeeker, size int64) (time.Duration, error) {
	off := int64(4)
	var total uint64
	var rate uint32
	for {
		header := make([]byte, 4)
		if err := readAt(r, off, header); err != nil {
			return 0, err
		}
		last := header[0]&0x80 != 0
		kind := header[0] & 0x7F
		length := int64(header[1])<<16 | int64(header[2])<<8 | int64(header[3])
		if off+4+length > size {
			return 0, corrupt("FLAC 元数据不完整")
		}
		if kind == 0 {
			if length < 34 {
				return 0, corrupt("FLAC STREAMINFO 无效")
			}
			info := make([]byte, 34)
			if err := readAt(r, off+4, info); err != nil {
				return 0, err
			}
			// Bytes 10-17: 20 bits sample rate, 3 channels, 5 bits per
			// sample, 36 bits total samples.
			v := binary.BigEndian.Uint64(info[10:18])
			rate = uint32(v >> 44)
			total = v & (1<<36 - 1)
		}
		off += 4 + length
		if last {
			break
		}
	}
	if rate == 0 {
		return 0, corrupt("FLAC 缺少 STREAMINFO")
	}
	sync := make([]byte, 2)
	if err := readAt(r, off, sync); err != nil {
		return 0, err
	}
	if sync[0] != 0xFF || sync[1]&0xFE != 0xF8 {
		return 0, corrupt("找不到 FLAC 音频帧")
	}
	return seconds(total, rate), nil
}

// --- MP4 -------------------------------------------------------------------

type mp4Box struct {
	kind       string
	start, end int64 // payload range
}

// mp4Children lists the boxes in [start, end).
func mp4Children(r io.ReadSeeker, start, end int64) ([]mp4Box, error) {
	var boxes []mp4Box
	for off := start; off < end; {
		if end-off < 8 {
			return nil, corrupt("MP4 结构不完整")
		}
		h := make([]byte, 8)
		if err := readAt(r, off, h); err != nil {
			return nil, err
		}
		size := int64(binary.BigEndian.Uint32(h))
		header := int64(8)
		switch size {
		case 0:
			size = end - off
		case 1:
			ext := make([]byte, 8)
			if err := readAt(r, off+8, ext); err != nil {
				return nil, err
			}
			size = int64(binary.BigEndian.Uint64(ext))
			header = 16
		}
		if size < header || off+size > end {
			return nil, corrupt("MP4 结构不完整")
		}
		boxes = append(boxes, mp4Box{kind: string(h[4:8]), start: off + header, end: off + size})
		off += size
	}
	return boxes, nil
}

func findBox(boxes []mp4Box, kind string) (mp4Box, bool) {
	for _, b := range boxes {
		if b.kind == kind {
			return b, true
		}
	}
	return mp4Box{}, false
}

func mp4Duration(r io.ReadSeeker, size int64) (time.Duration, error) {
	top, err := mp4Children(r, 0, size)
	if err != nil {
		return 0, err
	}
	moov, ok := findBox(top, "moov")
	if !ok {
		return 0, corrupt("MP4 缺少 moov")
	}
	if _, ok := findBox(top, "mdat"); !ok {
		return 0, corrupt("MP4 缺少音频数据")
	}
	traks, err := mp4Children(r, moov.start, moov.end)
	if err != nil {
		return 0, err
	}
	for _, trak := range traks {
		if trak.kind != "trak" {
			continue
		}
		d, codec, err := mp4AudioTrack(r, trak)
		if err != nil {
			return 0, err
		}
		if codec == "" {
			continue // not a sound track, e.g. cover art
		}
		if codec != "mp4a" {
			return 0, ErrUnsupported // ALAC and others Chrome cannot play
		}
		return d, nil
	}
	return 0, ErrUnsupported
}

// mp4AudioTrack returns the duration and sample entry codec of a sound
// track, or an empty codec for other track types.
func mp4AudioTrack(r io.ReadSeeker, trak mp4Box) (time.Duration, string, error) {
	children, err := mp4Children(r, trak.start, trak.end)
	if err != nil {
		return 0, "", err
	}
	mdia, ok := findBox(children, "mdia")
	if !ok {
		return 0, "", nil
	}
	if children, err = mp4Children(r, mdia.start, mdia.end); err != nil {
		return 0, "", err
	}
	hdlr, ok := findBox(children, "hdlr")
	if !ok || hdlr.end-hdlr.start < 12 {
		return 0, "", nil
	}
	handler := make([]byte, 4)
	if err := readAt(r, hdlr.start+8, handler); err != nil {
		return 0, "", err
	}
	if string(handler) != "soun" {
		return 0, "", nil
	}
	mdhd, ok := findBox(children, "mdhd")
	if !ok {
		return 0, "", corrupt("MP4 缺少 mdhd")
	}
	hd := make([]byte, min(mdhd.end-mdhd.start, 32))
	if err := readAt(r, mdhd.start, hd); err != nil {
		return 0, "", err
	}
	var scale uint32
	var duration uint64
	switch {
	case hd[0] == 1 && len(hd) >= 32:
		scale = binary.BigEndian.Uint32(hd[20:])
		duration = binary.BigEndian.Uint64(hd[24:])
	case hd[0] == 0 && len(hd) >= 20:
		scale = binary.BigEndian.Uint32(hd[12:])
		duration = uint64(binary.BigEndian.Uint32(hd[16:]))
	default:
		return 0, "", corrupt("MP4 mdhd 无效")
	}

	codec := "unknown"
	minf, ok := findBox(children, "minf")
	if ok {
		if children, err = mp4Children(r, minf.start, minf.end); err != nil {
			return 0, "", err
		}
		if stbl, ok := findBox(children, "stbl"); ok {
			if children, err = mp4Children(r, stbl.start, stbl.end); err != nil {
				return 0, "", err
			}
			if stsd, ok := findBox(children, "stsd"); ok && stsd.end-stsd.start >= 16 {
				entry := make([]byte, 4)
				if err := readAt(r, stsd.start+12, entry); err != nil {
					return 0, "", err
				}
				codec = string(entry)
			}
		}
	}
	return seconds(duration, scale), codec, nil
}

// --- Ogg -------------------------------------------------------------------

// oggPage is the part of an Ogg page header that matters here.
type oggPage struct {
	granule uint64
	serial  uint32
	payload int64 // offset of the first packet
}

func readOggPage(r io.ReadSeeker, off int64) (oggPage, error) {
	h := make([]byte, 27)
	if err := readAt(r, off, h); err != nil {
		return oggPage{}, err
	}
	if string(h[:4]) != "OggS" {
		return oggPage{}, corrupt("Ogg 页无效")
	}
	segs := make([]byte, h[26])
	if err := readAt(r, off+27, segs); err != nil {
		return oggPage{}, err
	}
	return oggPage{
		granule: binary.LittleEndian.Uint64(h[6:]),
		serial:  binary.LittleEndian.Uint32(h[14:]),
		payload: off + 27 + int64(len(segs)),
	}, nil
}

func oggDuration(r io.ReadSeeker, size int64) (Format, time.Duration, error) {
	first, err := readOggPage(r, 0)
	if err != nil {
		return "", 0, err
	}
	id := make([]byte, 19)
	if err := readAt(r, first.payload, id); err != nil {
		return "", 0, err
	}
	var format Format
	var rate uint32
	var preSkip uint64
	switch {
	case string(id[:7]) == "\x01vorbis":
		format, rate = OGG, binary.LittleEndian.Uint32(id[12:])
	case string(id[:8]) == "OpusHead":
		format, rate = OPUS, 48000
		preSkip = uint64(binary.LittleEndian.Uint16(id[10:]))
	default:
		return "", 0, ErrUnsupported // Ogg FLAC, Speex, Theora...
	}

	// The last page of the stream carries the total granule position.
	tailLen := min(size, 64<<10)
	tail := make([]byte, tailLen)
	if err := readAt(r, size-tailLen, tail); err != nil {
		return "", 0, err
	}
	for i := bytes.LastIndex(tail, []byte("OggS")); i >= 0; i = bytes.LastIndex(tail[:i], []byte("OggS")) {
		page, err := readOggPage(r, size-tailLen+int64(i))
		if err != nil || page.serial != first.serial || page.granule == ^uint64(0) {
			continue
		}
		if page.granule <= preSkip {
			break
		}
		return format, seconds(page.granule-preSkip, rate), nil
	}
	return "", 0, corrupt("找不到 Ogg 结束页")
}
