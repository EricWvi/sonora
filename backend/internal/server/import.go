package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"sonora/internal/admin"
	"sonora/internal/auth"
	"sonora/internal/media"
	"sonora/internal/netease"
	"sonora/internal/store"
)

const (
	// importTimeout bounds one import, most of which is the download.
	importTimeout = 5 * time.Minute
	// maxImportAudio and maxImportCover match the limits of the admin upload.
	maxImportAudio = 1 << 30
	maxImportCover = 10 << 20
)

// canImport tells whether the listener may import songs: a member of the
// configured group, as of their last login.
func (s *server) canImport(user store.User) bool {
	return s.opts.Netease != nil && s.opts.ImportGroup != "" && slices.Contains(user.Groups, s.opts.ImportGroup)
}

// importNeteaseSong copies a song into the library: the file comes straight
// from NetEase to here, never through the browser. The song's tags are written
// to the file first, so that the ordinary ingest reads the same values the file
// carries from then on.
func (s *server) importNeteaseSong(w http.ResponseWriter, r *http.Request) {
	if !s.canImport(auth.User(r.Context())) {
		admin.WriteError(w, http.StatusForbidden, "import_forbidden", "没有导入歌曲的权限")
		return
	}
	id, ok := onlineID(w, r)
	if !ok {
		return
	}
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), importTimeout)
	defer cancel()

	// One at a time, so that a second click on the same song finds the first
	// one's track instead of downloading it again.
	s.importing.Lock()
	defer s.importing.Unlock()

	existing, err := s.store.ImportedTrack(ctx, id)
	if err == nil {
		respond(w, r, importResult(store.Duplicate, existing), nil)
		return
	}
	if !errors.Is(err, store.ErrNotFound) {
		writeStoreError(w, r, err)
		return
	}
	songs, err := s.opts.Netease.Songs(ctx, session.Cookie, []int64{id})
	if err == nil && len(songs) == 0 {
		err = netease.ErrNotFound
	}
	if err != nil {
		writeImportError(w, r, err)
		return
	}
	detail, err := s.downloadSong(ctx, session.Cookie, songImport{song: songs[0]})
	if err != nil {
		writeImportError(w, r, err)
		return
	}
	respond(w, r, importResult(detail.Status, detail.Track), nil)
}

func importResult(status store.IngestStatus, track store.TrackDetail) map[string]any {
	return map[string]any{"status": status, "track": toPlayerTrack(track.Track)}
}

// errNoFile and errNotMP3 are songs that cannot be imported for this account;
// errDownload is a file that could not be fetched from NetEase's CDN, which
// says nothing about the API service.
var (
	errNoFile   = errors.New("no full version available")
	errNotMP3   = errors.New("not an MP3")
	errDownload = errors.New("download failed")
)

func writeImportError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errNoFile):
		admin.WriteError(w, http.StatusNotFound, "netease_no_url", "这首歌暂时无法导入：账号没有完整版本")
	case errors.Is(err, errNotMP3):
		admin.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_format", "目前只能导入 MP3")
	default:
		writeNeteaseError(w, r, err)
	}
}

// songImport is one song to download into the library.
type songImport struct {
	song netease.Song
	// album is set by an album import; nil imports a single.
	album *store.AlbumImport
	// albumArtists are the NetEase IDs of the album's artists by name.
	albumArtists    map[string]int64
	trackNo, discNo int
	// covers keeps the covers an album import downloaded, by address: the
	// songs of an album mostly share one.
	covers map[string][]byte
}

func (s *server) downloadSong(ctx context.Context, cookie string, in songImport) (store.IngestResult, error) {
	client := s.opts.Netease
	song := in.song
	file, err := client.SongFile(ctx, cookie, song.ID)
	if err != nil {
		return store.IngestResult{}, err
	}
	if file.URL == "" {
		return store.IngestResult{}, errNoFile
	}
	if file.Type != "mp3" {
		return store.IngestResult{}, errNotMP3
	}
	lyrics, err := client.Lyrics(ctx, cookie, song.ID)
	if err != nil {
		return store.IngestResult{}, err
	}
	meta := media.Meta{Title: song.Title, Year: song.Year, TrackNo: in.trackNo, DiscNo: in.discNo}
	ids := store.NeteaseIDs{Song: song.ID, Artists: map[string]int64{}, FromAlbum: in.album}
	for _, a := range song.Artists {
		meta.Artists = append(meta.Artists, a.Name)
		// The library trims the names it reads back from the tags.
		ids.Artists[strings.TrimSpace(a.Name)] = a.ID
	}
	if in.album != nil {
		meta.Album = in.album.Title
		meta.AlbumArtists = in.album.Artists
		ids.Album = in.album.NeteaseID
		for name, id := range in.albumArtists {
			if _, ok := ids.Artists[name]; !ok {
				ids.Artists[name] = id
			}
		}
	} else {
		// The album artists of a single are the song's artists rather than
		// NetEase's album artists (Eric, 2026-11-05).
		meta.AlbumArtists = meta.Artists
		if song.Album != nil {
			meta.Album = song.Album.Title
			ids.Album = song.Album.ID
		}
	}
	if song.Cover != nil {
		picture, ok := in.covers[song.Cover.URL]
		if !ok {
			var cover bytes.Buffer
			if err := client.Download(ctx, song.Cover.URL, &cover, maxImportCover); err != nil {
				return store.IngestResult{}, fmt.Errorf("%w: cover: %w", errDownload, err)
			}
			picture = cover.Bytes()
			if in.covers != nil {
				in.covers[song.Cover.URL] = picture
			}
		}
		meta.Picture = picture
	}

	tmp, err := s.store.IncomingFile()
	if err != nil {
		return store.IngestResult{}, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if err := client.Download(ctx, file.URL, tmp, maxImportAudio); err != nil {
		var rejected *netease.RejectedError
		if errors.As(err, &rejected) {
			return store.IngestResult{}, err
		}
		return store.IngestResult{}, fmt.Errorf("%w: %w", errDownload, err)
	}
	if err := tmp.Close(); err != nil {
		return store.IngestResult{}, err
	}
	if err := media.WriteTags(tmp.Name(), meta); err != nil {
		return store.IngestResult{}, err
	}
	tagged, err := os.Open(tmp.Name())
	if err != nil {
		return store.IngestResult{}, err
	}
	defer tagged.Close()
	upload, err := s.store.Receive(tagged, song.Title+".mp3")
	if err != nil {
		return store.IngestResult{}, err
	}
	text, _ := lyrics.Compose()
	return s.store.IngestImported(ctx, upload, []byte(text), ids)
}
