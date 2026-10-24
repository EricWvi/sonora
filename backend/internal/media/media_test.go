package media

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
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
			if tags.Title != "晴天" || tags.Artist != "周杰伦" || tags.Album != "叶惠美" {
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
	if plain.Tags != (Tags{}) {
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
