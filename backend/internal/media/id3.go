package media

import (
	"encoding/binary"
	"io"
	"strings"
	"unicode/utf16"
)

// id3TextValues reads the given text frames of the ID3v2.3/2.4 tag at the
// start of r, keeping each frame's NUL-separated values apart. The tag
// library joins them with nothing, so "周杰伦\x00费玉清" would read as
// "周杰伦费玉清". A malformed tag yields what was read before the damage;
// the tag library has already decided whether the file is readable.
func id3TextValues(r io.ReadSeeker, ids ...string) map[string][]string {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil
	}
	head := make([]byte, 10)
	if _, err := io.ReadFull(r, head); err != nil || !hasPrefix(head, "ID3") {
		return nil
	}
	version, flags := head[3], head[5]
	if version != 3 && version != 4 {
		return nil
	}
	// ReadAll grows with the data actually present, not the declared size.
	body, err := io.ReadAll(io.LimitReader(r, int64(syncsafe(head[6:10]))))
	if err != nil {
		return nil
	}
	if version == 3 && flags&0x80 != 0 {
		body = undoUnsync(body)
	}
	if flags&0x40 != 0 && len(body) >= 4 {
		n := int(syncsafe(body[:4])) // 2.4 counts the size field itself
		if version == 3 {
			n = int(binary.BigEndian.Uint32(body[:4])) + 4
		}
		if n < 0 || n > len(body) {
			return nil
		}
		body = body[n:]
	}

	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	values := make(map[string][]string)
	for len(body) >= 10 && body[0] != 0 {
		id := string(body[:4])
		size := int(binary.BigEndian.Uint32(body[4:8]))
		if version == 4 {
			size = int(syncsafe(body[4:8]))
		}
		format := body[9]
		if size < 0 || size > len(body)-10 {
			break
		}
		data := body[10 : 10+size]
		body = body[10+size:]
		if !want[id] {
			continue
		}
		if version == 4 {
			if format&0x0C != 0 { // compressed or encrypted
				continue
			}
			if format&0x01 != 0 { // data length indicator
				if len(data) < 4 {
					continue
				}
				data = data[4:]
			}
			if format&0x02 != 0 {
				data = undoUnsync(data)
			}
		} else {
			if format&0xC0 != 0 { // compressed or encrypted
				continue
			}
			if format&0x20 != 0 { // grouping identity
				if len(data) < 1 {
					continue
				}
				data = data[1:]
			}
		}
		values[id] = id3Strings(data)
	}
	return values
}

// id3Strings decodes the NUL-separated strings of a text frame, dropping
// empty ones.
func id3Strings(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	enc, text := data[0], data[1:]
	var parts [][]byte
	switch enc {
	case 0, 3:
		parts = splitTerminated(text, 1)
	case 1, 2:
		parts = splitTerminated(text, 2)
	default:
		return nil
	}
	var out []string
	bigEndian := enc == 2
	for _, p := range parts {
		var s string
		switch enc {
		case 0:
			r := make([]rune, len(p))
			for i, b := range p {
				r[i] = rune(b)
			}
			s = string(r)
		case 1:
			// Each string carries its own BOM; one without inherits the order.
			if len(p) >= 2 && p[0] == 0xFF && p[1] == 0xFE {
				bigEndian, p = false, p[2:]
			} else if len(p) >= 2 && p[0] == 0xFE && p[1] == 0xFF {
				bigEndian, p = true, p[2:]
			}
			s = decodeUTF16(p, bigEndian)
		case 2:
			s = decodeUTF16(p, true)
		case 3:
			s = strings.ToValidUTF8(string(p), "�")
		}
		if s = clean(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// splitTerminated splits b at NUL characters of the given width.
func splitTerminated(b []byte, width int) [][]byte {
	var parts [][]byte
	start := 0
	for i := 0; i+width <= len(b); i += width {
		if b[i] == 0 && (width == 1 || b[i+1] == 0) {
			parts = append(parts, b[start:i])
			start = i + width
		}
	}
	return append(parts, b[start:])
}

func decodeUTF16(b []byte, bigEndian bool) string {
	u := make([]uint16, len(b)/2)
	for i := range u {
		if bigEndian {
			u[i] = binary.BigEndian.Uint16(b[2*i:])
		} else {
			u[i] = binary.LittleEndian.Uint16(b[2*i:])
		}
	}
	return string(utf16.Decode(u))
}

// undoUnsync reverses ID3 unsynchronisation: 0xFF 0x00 becomes 0xFF.
func undoUnsync(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		out = append(out, b[i])
		if b[i] == 0xFF && i+1 < len(b) && b[i+1] == 0 {
			i++
		}
	}
	return out
}

func syncsafe(b []byte) uint32 {
	return uint32(b[0]&0x7F)<<21 | uint32(b[1]&0x7F)<<14 | uint32(b[2]&0x7F)<<7 | uint32(b[3]&0x7F)
}
