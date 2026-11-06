package media

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bogem/id3v2/v2"
)

// Meta is what WriteTags stores in a file.
type Meta struct {
	Title   string
	Artists []string
	Album   string
	// AlbumArtists is written to TPE2.
	AlbumArtists []string
	Year         int
	// TrackNo and DiscNo are left out when 0.
	TrackNo int
	DiscNo  int
	Picture []byte
}

// id3Separator keeps the values of an ID3v2.4 text frame apart, which is what
// id3TextValues reads back: names like "AC/DC" cannot be split on anything else.
const id3Separator = "\x00"

// WriteTags replaces the ID3 tag of the MP3 file at path. Nothing the file
// carried before is kept.
func WriteTags(path string, m Meta) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: false})
	if err != nil {
		return fmt.Errorf("open tag: %w", err)
	}
	defer tag.Close()
	tag.SetVersion(4)
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	tag.SetTitle(m.Title)
	tag.SetArtist(strings.Join(m.Artists, id3Separator))
	tag.SetAlbum(m.Album)
	if len(m.AlbumArtists) > 0 {
		tag.AddTextFrame("TPE2", id3v2.EncodingUTF8, strings.Join(m.AlbumArtists, id3Separator))
	}
	if m.Year > 0 {
		tag.SetYear(strconv.Itoa(m.Year))
	}
	if m.TrackNo > 0 {
		tag.AddTextFrame("TRCK", id3v2.EncodingUTF8, strconv.Itoa(m.TrackNo))
	}
	if m.DiscNo > 0 {
		tag.AddTextFrame("TPOS", id3v2.EncodingUTF8, strconv.Itoa(m.DiscNo))
	}
	if len(m.Picture) > 0 {
		tag.AddAttachedPicture(id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    http.DetectContentType(m.Picture),
			PictureType: id3v2.PTFrontCover,
			Picture:     m.Picture,
		})
	}
	if err := tag.Save(); err != nil {
		return fmt.Errorf("save tag: %w", err)
	}
	return nil
}
