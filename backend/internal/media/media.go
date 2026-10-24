// Package media identifies uploaded audio files and reads their duration,
// tags, embedded cover and embedded lyrics.
//
// The format is sniffed from the file content, never from the file name.
// Only formats desktop Chrome can play are accepted. A file whose container
// structure cannot be read is reported as corrupt; the audio frames
// themselves are not decoded.
package media

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

var (
	// ErrUnsupported reports a file that is not audio in an accepted format.
	ErrUnsupported = errors.New("unsupported format")
	// ErrCorrupt reports an accepted format whose structure is broken.
	ErrCorrupt = errors.New("corrupt file")
)

// Format is an accepted audio format. The value doubles as the file
// extension used in storage.
type Format string

const (
	MP3  Format = "mp3"
	FLAC Format = "flac"
	M4A  Format = "m4a" // AAC in an MP4 container
	OGG  Format = "ogg" // Vorbis in Ogg
	OPUS Format = "opus"
)

// ContentType is the MIME type used when the file is served.
func (f Format) ContentType() string {
	switch f {
	case MP3:
		return "audio/mpeg"
	case FLAC:
		return "audio/flac"
	case M4A:
		return "audio/mp4"
	case OGG:
		return "audio/ogg"
	case OPUS:
		return "audio/ogg; codecs=opus"
	}
	return "application/octet-stream"
}

// Info is what an audio file says about itself.
type Info struct {
	Format   Format
	Duration time.Duration
	Tags     Tags
}

// Tags are the metadata fields Sonora uses. Missing text fields are empty
// and missing numbers are zero.
type Tags struct {
	Title string
	// Artists and AlbumArtists keep the separate values of a multi-value
	// field in tag order. A single value is never split on separators,
	// since names like "AC/DC" contain them.
	Artists      []string
	Album        string
	AlbumArtists []string
	Track        int
	Disc         int
	Year         int
	Lyrics       string
	Picture      *Picture
}

// multiSeparator joins the values of a multi-value title or album, which
// unlike artists have nowhere to keep them apart.
const multiSeparator = ", "

// Picture is an embedded cover image.
type Picture struct {
	MIME string
	Data []byte
}

// Probe identifies the audio file in r, which is size bytes long.
func Probe(r io.ReadSeeker, size int64) (Info, error) {
	head := make([]byte, 12)
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return Info{}, err
	}
	n, err := io.ReadFull(r, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return Info{}, err
	}
	head = head[:n]

	var info Info
	switch {
	case hasPrefix(head, "fLaC"):
		info.Format = FLAC
		info.Duration, err = flacDuration(r, size)
	case hasPrefix(head, "OggS"):
		info.Format, info.Duration, err = oggDuration(r, size)
	case len(head) >= 8 && string(head[4:8]) == "ftyp":
		info.Format = M4A
		info.Duration, err = mp4Duration(r, size)
	case hasPrefix(head, "ID3") || isMP3Sync(head):
		info.Format = MP3
		info.Duration, err = mp3Duration(r, size)
	default:
		return Info{}, ErrUnsupported
	}
	if err != nil {
		return Info{}, err
	}
	if info.Duration <= 0 {
		return Info{}, fmt.Errorf("%w: 无法确定时长", ErrCorrupt)
	}
	if info.Tags, err = readTags(r); err != nil {
		return Info{}, err
	}
	return info, nil
}

func readTags(r io.ReadSeeker) (Tags, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return Tags{}, err
	}
	m, err := tag.ReadFrom(r)
	if errors.Is(err, tag.ErrNoTagsFound) {
		return Tags{}, nil
	}
	if err != nil {
		return Tags{}, fmt.Errorf("%w: 标签无法读取", ErrCorrupt)
	}
	t := Tags{
		Title:  clean(m.Title()),
		Album:  clean(m.Album()),
		Year:   m.Year(),
		Lyrics: clean(m.Lyrics()),
	}
	if a := clean(m.Artist()); a != "" {
		t.Artists = []string{a}
	}
	if a := clean(m.AlbumArtist()); a != "" {
		t.AlbumArtists = []string{a}
	}
	// The tag library flattens multi-value fields; reread those it would
	// get wrong.
	var multi map[string][]string
	var keys [4]string // title, artist, album, album artist
	switch m.Format() {
	case tag.ID3v2_3, tag.ID3v2_4:
		keys = [4]string{"TIT2", "TPE1", "TALB", "TPE2"}
		multi = id3TextValues(r, keys[:]...)
	case tag.VORBIS:
		keys = [4]string{"title", "artist", "album", "albumartist"}
		multi = vorbisValues(r)
	}
	if v := multi[keys[0]]; len(v) > 1 {
		t.Title = strings.Join(v, multiSeparator)
	}
	if v := multi[keys[1]]; len(v) > 1 {
		t.Artists = v
	}
	if v := multi[keys[2]]; len(v) > 1 {
		t.Album = strings.Join(v, multiSeparator)
	}
	if v := multi[keys[3]]; len(v) > 1 {
		t.AlbumArtists = v
	}
	t.Track, _ = m.Track()
	t.Disc, _ = m.Disc()
	raw := m.Raw()
	if t.Track == 0 {
		t.Track = leadingNumber(raw["tracknumber"])
	}
	if t.Disc == 0 {
		t.Disc = leadingNumber(raw["discnumber"])
	}
	if t.Lyrics == "" {
		t.Lyrics = rawLyrics(raw)
	}
	if p := m.Picture(); p != nil && len(p.Data) > 0 {
		t.Picture = &Picture{MIME: p.MIMEType, Data: p.Data}
	}
	if t.Year < 0 || t.Year > 9999 {
		t.Year = 0
	}
	return t, nil
}

// rawLyrics finds lyrics the tag library does not map: the UNSYNCEDLYRICS
// Vorbis comment, or an ID3 TXXX frame named LYRICS, UNSYNCEDLYRICS or
// USLT as written by foobar2000 and ffmpeg.
func rawLyrics(raw map[string]any) string {
	if s, ok := raw["unsyncedlyrics"].(string); ok && clean(s) != "" {
		return clean(s)
	}
	if c, ok := raw["TXXX"].(*tag.Comm); ok {
		switch strings.ToUpper(c.Description) {
		case "LYRICS", "UNSYNCEDLYRICS", "USLT":
			return clean(c.Text)
		}
	}
	return ""
}

// leadingNumber parses "3" or "3/10" as 3.
func leadingNumber(v any) int {
	s, _ := v.(string)
	s, _, _ = strings.Cut(strings.TrimSpace(s), "/")
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// clean trims whitespace and the NUL terminators some taggers leave.
func clean(s string) string {
	return strings.Trim(s, " \t\r\n\x00")
}

func hasPrefix(b []byte, prefix string) bool {
	return len(b) >= len(prefix) && string(b[:len(prefix)]) == prefix
}

func corrupt(msg string) error { return fmt.Errorf("%w: %s", ErrCorrupt, msg) }

// readAt reads exactly len(buf) bytes at off, mapping a short read to a
// corrupt-file error.
func readAt(r io.ReadSeeker, off int64, buf []byte) error {
	if _, err := r.Seek(off, io.SeekStart); err != nil {
		return err
	}
	if _, err := io.ReadFull(r, buf); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return corrupt("文件不完整")
		}
		return err
	}
	return nil
}

func seconds(samples uint64, rate uint32) time.Duration {
	if rate == 0 {
		return 0
	}
	return time.Duration(float64(samples) / float64(rate) * float64(time.Second))
}
