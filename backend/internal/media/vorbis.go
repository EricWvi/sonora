package media

import (
	"encoding/binary"
	"io"
	"strings"
)

// maxCommentBytes bounds the Vorbis comment read; embedded pictures make
// it the largest metadata a file carries.
const maxCommentBytes = 64 << 20

// vorbisValues reads the Vorbis comment of a FLAC, Ogg Vorbis or Opus file
// in r and returns every value of each field, keyed by lowercase name. The
// tag library keeps only the last of repeated fields, so ARTIST=A and
// ARTIST=B would read as "B". A malformed comment yields nil; the tag
// library has already decided whether the file is readable.
func vorbisValues(r io.ReadSeeker) map[string][]string {
	head := make([]byte, 4)
	if err := readAt(r, 0, head); err != nil {
		return nil
	}
	var comment []byte
	switch string(head) {
	case "fLaC":
		comment = flacComment(r)
	case "OggS":
		packet := oggPacket(r, 1) // identification, then comment
		switch {
		case hasPrefix(packet, "\x03vorbis"):
			comment = packet[7:]
		case hasPrefix(packet, "OpusTags"):
			comment = packet[8:]
		}
	}
	if comment == nil {
		return nil
	}
	next := func() ([]byte, bool) {
		if len(comment) < 4 {
			return nil, false
		}
		n := binary.LittleEndian.Uint32(comment)
		if uint64(n) > uint64(len(comment)-4) {
			return nil, false
		}
		s := comment[4 : 4+n]
		comment = comment[4+n:]
		return s, true
	}
	if _, ok := next(); !ok { // vendor string
		return nil
	}
	if len(comment) < 4 {
		return nil
	}
	count := binary.LittleEndian.Uint32(comment)
	comment = comment[4:]
	values := make(map[string][]string)
	for range count {
		field, ok := next()
		if !ok {
			return nil
		}
		k, v, ok := strings.Cut(string(field), "=")
		if !ok {
			continue
		}
		if v = clean(strings.ToValidUTF8(v, "�")); v != "" {
			k = strings.ToLower(k)
			values[k] = append(values[k], v)
		}
	}
	return values
}

// flacComment returns the body of the VORBIS_COMMENT metadata block.
func flacComment(r io.ReadSeeker) []byte {
	off := int64(4)
	for {
		header := make([]byte, 4)
		if err := readAt(r, off, header); err != nil {
			return nil
		}
		length := int64(header[1])<<16 | int64(header[2])<<8 | int64(header[3])
		if header[0]&0x7F == 4 {
			body := make([]byte, length)
			if err := readAt(r, off+4, body); err != nil {
				return nil
			}
			return body
		}
		if header[0]&0x80 != 0 {
			return nil
		}
		off += 4 + length
	}
}

// oggPacket reassembles packet n (from 0) of the first logical stream,
// following it across pages.
func oggPacket(r io.ReadSeeker, n int) []byte {
	var packet []byte
	var serial uint32
	off := int64(0)
	for first := true; ; first = false {
		h := make([]byte, 27)
		if err := readAt(r, off, h); err != nil || string(h[:4]) != "OggS" {
			return nil
		}
		segs := make([]byte, h[26])
		if err := readAt(r, off+27, segs); err != nil {
			return nil
		}
		off += 27 + int64(len(segs))
		if first {
			serial = binary.LittleEndian.Uint32(h[14:])
		}
		if binary.LittleEndian.Uint32(h[14:]) != serial {
			for _, s := range segs {
				off += int64(s)
			}
			continue
		}
		for _, s := range segs {
			if n == 0 {
				if len(packet)+int(s) > maxCommentBytes {
					return nil
				}
				buf := make([]byte, s)
				if err := readAt(r, off, buf); err != nil {
					return nil
				}
				packet = append(packet, buf...)
			}
			off += int64(s)
			if s < 255 { // a lacing value below 255 ends the packet
				if n == 0 {
					return packet
				}
				n--
			}
		}
	}
}
