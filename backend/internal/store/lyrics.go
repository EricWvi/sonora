package store

import (
	"bytes"
	"context"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
)

// MaxLyricsBytes caps lyrics text.
const MaxLyricsBytes = 512 << 10

// decodeLyrics turns an uploaded lyrics file into text. LRC files from
// Chinese sources are often GB18030 rather than UTF-8, and Windows tools
// write UTF-16 with a byte order mark.
func decodeLyrics(data []byte) (string, error) {
	if len(data) > MaxLyricsBytes {
		return "", invalid("歌词不能超过 512 KB")
	}
	var text string
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}), bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		out, err := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Bytes(data)
		if err != nil {
			return "", invalid("歌词编码无法识别")
		}
		text = string(out)
	case utf8.Valid(data):
		text = string(bytes.TrimPrefix(data, []byte("\xEF\xBB\xBF")))
	default:
		out, err := simplifiedchinese.GB18030.NewDecoder().Bytes(data)
		if err != nil || !utf8.Valid(out) {
			return "", invalid("歌词编码无法识别，请使用 UTF-8")
		}
		text = string(out)
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.TrimSpace(text), nil
}

// SetTrackLyrics replaces a track's lyrics with an uploaded file or edited
// text. Empty input removes them.
func (s *Store) SetTrackLyrics(ctx context.Context, id int64, data []byte) (TrackDetail, error) {
	text, err := decodeLyrics(data)
	if err != nil {
		return TrackDetail{}, err
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE tracks SET lyrics = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ?`, text, id)
	if err != nil {
		return TrackDetail{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return TrackDetail{}, ErrNotFound
	}
	return s.Track(ctx, id)
}
