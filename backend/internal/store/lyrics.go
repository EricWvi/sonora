package store

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
)

// MaxLyricsBytes caps lyrics text.
const MaxLyricsBytes = 512 << 10

// Lyrics formats. LRC covers plain, word-timed and enhanced LRC, which the
// player tells apart; TTML is the Apple Music style XML.
const (
	LRC  = "lrc"
	TTML = "ttml"
)

// lyricsFormat tells TTML from LRC by content, so pasted text and files
// with any extension work alike. Text that looks like XML must be a
// well-formed TTML document.
func lyricsFormat(text string) (string, error) {
	if !strings.HasPrefix(text, "<") {
		return LRC, nil
	}
	d := xml.NewDecoder(strings.NewReader(text))
	root := false
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", invalid("TTML 歌词格式错误：" + err.Error())
		}
		if el, ok := tok.(xml.StartElement); ok && !root {
			if el.Name.Local != "tt" {
				return "", invalid("XML 歌词只支持 TTML")
			}
			root = true
		}
	}
	if !root {
		return "", invalid("TTML 歌词格式错误：缺少 tt 元素")
	}
	return TTML, nil
}

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

// readLyrics decodes an uploaded lyrics file and finds its format.
func readLyrics(data []byte) (text, format string, err error) {
	if text, err = decodeLyrics(data); err != nil {
		return "", "", err
	}
	format, err = lyricsFormat(text)
	return text, format, err
}

// SetTrackLyrics replaces a track's lyrics with an uploaded LRC or TTML
// file or edited text. Empty input removes them.
func (s *Store) SetTrackLyrics(ctx context.Context, id int64, data []byte) (TrackDetail, error) {
	text, format, err := readLyrics(data)
	if err != nil {
		return TrackDetail{}, err
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE tracks SET lyrics = ?, lyrics_format = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		WHERE id = ?`, text, format, id)
	if err != nil {
		return TrackDetail{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return TrackDetail{}, ErrNotFound
	}
	return s.Track(ctx, id)
}
