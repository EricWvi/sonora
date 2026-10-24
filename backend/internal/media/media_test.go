package media

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

// The samples in testdata were encoded by ffmpeg from a 3-second 440 Hz
// tone; want durations come from ffprobe.
func probeFile(t *testing.T, name string) (Info, error) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return Probe(bytes.NewReader(data), int64(len(data)))
}

const wantLyrics = "[00:00.00]故事的小黄花\n[00:01.50]从出生那年就飘着"

func TestProbeTaggedSamples(t *testing.T) {
	tests := []struct {
		file     string
		format   Format
		duration time.Duration
		cover    bool
		lyrics   bool
	}{
		{"tagged.mp3", MP3, 3 * time.Second, true, true},
		{"tagged.flac", FLAC, 3 * time.Second, true, true},
		{"tagged.m4a", M4A, 3 * time.Second, true, true},
		{"tagged.ogg", OGG, 3 * time.Second, false, false},
		{"tagged.opus", OPUS, 3006500 * time.Microsecond, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			info, err := probeFile(t, tt.file)
			if err != nil {
				t.Fatal(err)
			}
			if info.Format != tt.format {
				t.Errorf("format = %q, want %q", info.Format, tt.format)
			}
			if diff := (info.Duration - tt.duration).Abs(); diff > 60*time.Millisecond {
				t.Errorf("duration = %v, want %v", info.Duration, tt.duration)
			}
			tags := info.Tags
			if tags.Title != "晴天" || !slices.Equal(tags.Artists, []string{"周杰伦"}) || tags.Album != "叶惠美" {
				t.Errorf("tags = %+v", tags)
			}
			if tags.Track != 3 || tags.Year != 2003 {
				t.Errorf("track = %d, year = %d", tags.Track, tags.Year)
			}
			if got := tags.Picture != nil; got != tt.cover {
				t.Errorf("has cover = %v, want %v", got, tt.cover)
			}
			if tt.cover && !bytes.HasPrefix(tags.Picture.Data, []byte{0xFF, 0xD8}) {
				t.Error("cover is not the embedded JPEG")
			}
			if tt.lyrics && strings.ReplaceAll(tags.Lyrics, "\r\n", "\n") != wantLyrics {
				t.Errorf("lyrics = %q", tags.Lyrics)
			}
		})
	}
}

func TestProbeMP3Variants(t *testing.T) {
	vbr, err := probeFile(t, "vbr.mp3")
	if err != nil {
		t.Fatal(err)
	}
	if diff := (vbr.Duration - 3*time.Second).Abs(); diff > 60*time.Millisecond {
		t.Errorf("VBR duration = %v", vbr.Duration)
	}
	if vbr.Tags.Title != "VBR" {
		t.Errorf("VBR title = %q", vbr.Tags.Title)
	}

	// No ID3 tag and no Xing header: duration from the constant bitrate.
	plain, err := probeFile(t, "untagged.mp3")
	if err != nil {
		t.Fatal(err)
	}
	if diff := (plain.Duration - 3030167*time.Microsecond).Abs(); diff > 30*time.Millisecond {
		t.Errorf("CBR duration = %v", plain.Duration)
	}
	if !reflect.DeepEqual(plain.Tags, Tags{}) {
		t.Errorf("untagged file has tags %+v", plain.Tags)
	}
}

func TestProbeRejectsBrokenAndForeignFiles(t *testing.T) {
	if _, err := probeFile(t, "truncated.flac"); !errors.Is(err, ErrCorrupt) {
		t.Errorf("truncated FLAC: err = %v, want ErrCorrupt", err)
	}
	if _, err := probeFile(t, "cover.jpg"); !errors.Is(err, ErrUnsupported) {
		t.Errorf("JPEG: err = %v, want ErrUnsupported", err)
	}
	for name, data := range map[string][]byte{
		"empty":     nil,
		"text":      []byte("hello, this is not audio"),
		"id3 only":  append([]byte("ID3\x04\x00\x00\x00\x00\x00\x10"), make([]byte, 16)...),
		"flac head": []byte("fLaC\x80\x00\x00\x22"),
		"ogg head":  []byte("OggS\x00\x02"),
	} {
		_, err := Probe(bytes.NewReader(data), int64(len(data)))
		if !errors.Is(err, ErrCorrupt) && !errors.Is(err, ErrUnsupported) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestProbeRejectsTruncatedLossyFiles(t *testing.T) {
	for _, name := range []string{"tagged.m4a", "tagged.ogg"} {
		data, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		cut := data[:len(data)/2]
		if name == "tagged.ogg" {
			// Keep only the first page so no end-of-stream page remains.
			cut = data[:200]
		}
		if _, err := Probe(bytes.NewReader(cut), int64(len(cut))); !errors.Is(err, ErrCorrupt) {
			t.Errorf("%s cut: err = %v, want ErrCorrupt", name, err)
		}
	}
}

// Taggers such as Mp3tag write lyrics as an ID3v2.3 USLT frame, which
// ffmpeg cannot produce; prepend one to the untagged sample.
func TestProbeMP3USLTLyrics(t *testing.T) {
	audio, err := os.ReadFile(filepath.Join("testdata", "untagged.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	// Encoding 3 (UTF-8), language, empty description, text.
	body := append([]byte("\x03chi\x00"), wantLyrics...)
	frame := append([]byte("USLT"), byte(len(body)>>24), byte(len(body)>>16), byte(len(body)>>8), byte(len(body)), 0, 0)
	frame = append(frame, body...)
	n := len(frame)
	header := []byte{'I', 'D', '3', 3, 0, 0, byte(n >> 21 & 0x7F), byte(n >> 14 & 0x7F), byte(n >> 7 & 0x7F), byte(n & 0x7F)}
	data := append(append(header, frame...), audio...)

	info, err := Probe(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if info.Tags.Lyrics != wantLyrics {
		t.Errorf("lyrics = %q", info.Tags.Lyrics)
	}
	if diff := (info.Duration - 3030167*time.Microsecond).Abs(); diff > 30*time.Millisecond {
		t.Errorf("duration with ID3 = %v", info.Duration)
	}
}

// id3Tag builds an ID3v2 tag of the given version from frames whose
// bodies already start with the text encoding byte.
func id3Tag(version byte, frames ...[2]string) []byte {
	size := func(n int) []byte {
		if version == 4 {
			return []byte{byte(n >> 21 & 0x7F), byte(n >> 14 & 0x7F), byte(n >> 7 & 0x7F), byte(n & 0x7F)}
		}
		return []byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
	}
	var body []byte
	for _, f := range frames {
		body = append(append(append(body, f[0]...), size(len(f[1]))...), 0, 0)
		body = append(body, f[1]...)
	}
	n := len(body)
	header := []byte{'I', 'D', '3', version, 0, 0, byte(n >> 21 & 0x7F), byte(n >> 14 & 0x7F), byte(n >> 7 & 0x7F), byte(n & 0x7F)}
	return append(header, body...)
}

func TestProbeMP3MultiValueText(t *testing.T) {
	audio, err := os.ReadFile(filepath.Join("testdata", "untagged.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	// UTF-16 with a BOM per value, as ID3v2.4 requires.
	utf16LE := "\x01\xff\xfeA\x00\x00\x00\xff\xfeB\x00\x00\x00"
	tests := []struct {
		name         string
		tag          []byte
		artists      []string
		album        string
		albumArtists []string
	}{
		{"v2.4 utf-8", id3Tag(4,
			[2]string{"TIT2", "\x03晴天"},
			[2]string{"TPE1", "\x03周杰伦\x00费玉清\x00"},
			[2]string{"TALB", "\x03叶惠美"},
			[2]string{"TPE2", "\x03周杰伦"},
		), []string{"周杰伦", "费玉清"}, "叶惠美", []string{"周杰伦"}},
		{"v2.4 utf-16 and latin-1", id3Tag(4,
			[2]string{"TIT2", "\x03晴天"},
			[2]string{"TPE1", utf16LE},
			[2]string{"TPE2", "\x00Simon\x00Garfunkel"},
		), []string{"A", "B"}, "", []string{"Simon", "Garfunkel"}},
		// A single value is never split; a multi-value album has no list
		// to go into, so its values are joined.
		{"v2.3 single artist, multi-value album", id3Tag(3,
			[2]string{"TIT2", "\x03晴天"},
			[2]string{"TPE1", "\x03AC/DC"},
			[2]string{"TALB", "\x03Vol. 1\x00Vol. 2"},
		), []string{"AC/DC"}, "Vol. 1, Vol. 2", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := append(tt.tag, audio...)
			info, err := Probe(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				t.Fatal(err)
			}
			tags := info.Tags
			if tags.Title != "晴天" || !slices.Equal(tags.Artists, tt.artists) || tags.Album != tt.album ||
				!slices.Equal(tags.AlbumArtists, tt.albumArtists) {
				t.Errorf("tags = %+v", tags)
			}
		})
	}
}

// vorbisComment builds a Vorbis comment from KEY=value fields.
func vorbisComment(fields ...string) []byte {
	le := func(n int) []byte { return binary.LittleEndian.AppendUint32(nil, uint32(n)) }
	b := append(le(6), "sonora"...)
	b = append(b, le(len(fields))...)
	for _, f := range fields {
		b = append(append(b, le(len(f))...), f...)
	}
	return b
}

// oggPageBytes builds one Ogg page holding segments whose lacing values
// are given, with the checksum the tag library verifies.
func oggPageBytes(kind byte, granule uint64, seq uint32, lacing []byte, payload []byte) []byte {
	p := append([]byte("OggS"), 0, kind)
	p = binary.LittleEndian.AppendUint64(p, granule)
	p = binary.LittleEndian.AppendUint32(p, 1) // serial
	p = binary.LittleEndian.AppendUint32(p, seq)
	p = append(p, 0, 0, 0, 0, byte(len(lacing)))
	p = append(append(p, lacing...), payload...)
	var crc uint32
	for _, b := range p {
		crc ^= uint32(b) << 24
		for range 8 {
			if crc&0x80000000 != 0 {
				crc = crc<<1 ^ 0x04c11db7
			} else {
				crc <<= 1
			}
		}
	}
	binary.LittleEndian.PutUint32(p[22:], crc)
	return p
}

// oggFile builds a three-second Ogg stream whose comment packet spans two
// pages.
func oggFile(id, commentPrefix string, rate uint64, comment []byte) []byte {
	packet := append([]byte(commentPrefix), comment...)
	if len(packet) <= 255 || len(packet)-255 >= 255 {
		panic("comment packet must span exactly one full segment")
	}
	f := oggPageBytes(0x02, 0, 0, []byte{byte(len(id))}, []byte(id))
	f = append(f, oggPageBytes(0, ^uint64(0), 1, []byte{255}, packet[:255])...)
	f = append(f, oggPageBytes(0x01, 0, 2, []byte{byte(len(packet) - 255)}, packet[255:])...)
	return append(f, oggPageBytes(0x04, 3*rate, 3, []byte{1}, []byte{0})...)
}

func TestProbeVorbisMultiValueFields(t *testing.T) {
	comment := vorbisComment(
		"TITLE=晴天", "ARTIST=周杰伦", "artist=费玉清", "ALBUM=叶惠美", "ALBUMARTIST=周杰伦",
		"COMMENT="+strings.Repeat("x", 300),
	)
	streamInfo := make([]byte, 34)
	binary.BigEndian.PutUint64(streamInfo[10:], 44100<<44|3*44100)
	flac := append([]byte("fLaC\x00\x00\x00\x22"), streamInfo...)
	flac = append(flac, 0x84, byte(len(comment)>>16), byte(len(comment)>>8), byte(len(comment)))
	flac = append(append(flac, comment...), 0xFF, 0xF8)

	vorbisID := "\x01vorbis\x00\x00\x00\x00\x01\x44\xac\x00\x00" + strings.Repeat("\x00", 15)
	opusID := "OpusHead\x01\x01\x00\x00\x80\xbb\x00\x00\x00\x00\x00"
	tests := []struct {
		name string
		data []byte
	}{
		{"flac", flac},
		{"ogg", oggFile(vorbisID, "\x03vorbis", 44100, comment)},
		{"opus", oggFile(opusID, "OpusTags", 48000, comment)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := Probe(bytes.NewReader(tt.data), int64(len(tt.data)))
			if err != nil {
				t.Fatal(err)
			}
			tags := info.Tags
			if tags.Title != "晴天" || !slices.Equal(tags.Artists, []string{"周杰伦", "费玉清"}) || tags.Album != "叶惠美" ||
				!slices.Equal(tags.AlbumArtists, []string{"周杰伦"}) {
				t.Errorf("tags = %+v", tags)
			}
		})
	}
}
