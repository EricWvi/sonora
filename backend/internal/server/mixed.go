package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"sonora/internal/admin"
	"sonora/internal/netease"
	"sonora/internal/store"
)

// maxSnapshotBytes caps a history snapshot of a NetEase song.
const maxSnapshotBytes = 16 << 10

// mixedEntry is one row of a list that merges local and NetEase songs. A row
// has either a local track or a NetEase ID; song is filled in once the page
// is known, since looking songs up is what costs.
type mixedEntry struct {
	at        time.Time
	local     *playerTrack
	neteaseID int64
	song      *netease.Song
	playedAt  string
}

// mixedPage is a merged list. OnlineFailed says the NetEase half is missing
// because the session ended or the API service is down, not because there
// is nothing there, so the player can say so.
type mixedPage struct {
	Items        []any `json:"items"`
	Total        int   `json:"total"`
	OnlineFailed bool  `json:"onlineFailed"`
}

type mixedHistoryEntry struct {
	Track    any    `json:"track"`
	PlayedAt string `json:"playedAt"`
}

// parseStoreTime reads a timestamp the store wrote. A bad one sorts last.
func parseStoreTime(raw string) time.Time {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return t
}

// sortMixed puts the newest first. Local entries come first in the input and
// the sort is stable, so they win a tie.
func sortMixed(entries []mixedEntry) {
	slices.SortStableFunc(entries, func(a, b mixedEntry) int { return b.at.Compare(a.at) })
}

// mergeSameSong keeps one row per song, the local one, when the same song is
// there both as a local track and as a NetEase one. That row takes the later
// of the two times.
func mergeSameSong(local, online []mixedEntry, imported map[int64]int64) []mixedEntry {
	byTrack := make(map[int64]int, len(local))
	for i, e := range local {
		byTrack[e.local.ID] = i
	}
	merged := slices.Clone(local)
	for _, e := range online {
		if i, ok := byTrack[imported[e.neteaseID]]; ok {
			if e.at.After(merged[i].at) {
				merged[i].at = e.at
			}
			continue
		}
		merged = append(merged, e)
	}
	sortMixed(merged)
	return merged
}

// fillSongs looks up the NetEase songs on the page and drops any NetEase no
// longer has.
func (s *server) fillSongs(ctx context.Context, cookie string, page []mixedEntry) ([]mixedEntry, error) {
	var ids []int64
	for _, e := range page {
		if e.local == nil && e.song == nil {
			ids = append(ids, e.neteaseID)
		}
	}
	if len(ids) == 0 {
		return page, nil
	}
	songs, err := s.opts.Netease.Songs(ctx, cookie, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]*netease.Song, len(songs))
	for i := range songs {
		byID[songs[i].ID] = &songs[i]
	}
	out := make([]mixedEntry, 0, len(page))
	for _, e := range page {
		if e.local == nil && e.song == nil {
			if e.song = byID[e.neteaseID]; e.song == nil {
				continue
			}
		}
		out = append(out, e)
	}
	return out, nil
}

// optionalSession returns the session behind a merged list, if the listener
// has a usable one. Merged lists fall back to the local half, so a missing
// session is no error.
func (s *server) optionalSession(ctx context.Context, user int64) (store.NeteaseSession, bool, error) {
	if s.opts.Netease == nil {
		return store.NeteaseSession{}, false, nil
	}
	session, err := s.store.NeteaseSession(ctx, user)
	if errors.Is(err, store.ErrNotFound) {
		return session, false, nil
	}
	return session, err == nil, err
}

// onlineFailure reports whether err is one a merged list survives.
func onlineFailure(err error) bool {
	if errors.Is(err, netease.ErrUnavailable) || errors.Is(err, netease.ErrExpired) {
		slog.Warn("netease merged list", "err", err)
		return true
	}
	return false
}

func pageOf[T any](items []T, w store.Window) []T {
	offset, limit := max(w.Offset, 0), w.Limit
	if limit <= 0 {
		limit = 50
	}
	if offset >= len(items) {
		return nil
	}
	return items[offset:min(offset+min(limit, 500), len(items))]
}

func (s *server) listMixedLikes(w http.ResponseWriter, r *http.Request) {
	ctx, user := r.Context(), userID(r)
	liked, err := s.store.LikedEntries(ctx, user)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	local := make([]mixedEntry, len(liked))
	for i, e := range liked {
		t := toPlayerTrack(e.Track)
		local[i] = mixedEntry{at: parseStoreTime(e.LikedAt), local: &t}
	}
	session, signedIn, err := s.optionalSession(ctx, user)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	entries, failed := local, false
	if signedIn {
		online, err := s.opts.Netease.Likes(ctx, session.Cookie, session.AccountID)
		switch {
		case err == nil:
			imported, err := s.store.ImportedTracks(ctx)
			if err != nil {
				writeStoreError(w, r, err)
				return
			}
			rows := make([]mixedEntry, len(online))
			for i, l := range online {
				rows[i] = mixedEntry{at: l.At, neteaseID: l.ID}
			}
			entries = mergeSameSong(local, rows, imported)
		case onlineFailure(err):
			failed = true
		default:
			writeNeteaseError(w, r, err)
			return
		}
	}
	page := pageOf(entries, window(r))
	if signedIn && !failed {
		if page, err = s.fillSongs(ctx, session.Cookie, page); err != nil {
			if !onlineFailure(err) {
				writeNeteaseError(w, r, err)
				return
			}
			failed = true
		}
	}
	items := make([]any, 0, len(page))
	for _, e := range page {
		switch {
		case e.local != nil:
			items = append(items, e.local)
		case e.song != nil:
			items = append(items, e.song)
		}
	}
	respond(w, r, mixedPage{Items: items, Total: len(entries), OnlineFailed: failed}, nil)
}

func (s *server) listMixedHistory(w http.ResponseWriter, r *http.Request) {
	ctx, user := r.Context(), userID(r)
	played, err := s.store.History(ctx, user, store.Window{Limit: store.MaxHistory})
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	local := make([]mixedEntry, len(played.Items))
	for i, e := range played.Items {
		t := toPlayerTrack(e.Track)
		local[i] = mixedEntry{at: parseStoreTime(e.PlayedAt), local: &t, playedAt: e.PlayedAt}
	}
	entries := local
	// Without a NetEase login the list is local only. The rows stay, so
	// signing back in brings them back.
	_, signedIn, err := s.optionalSession(ctx, user)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	if signedIn {
		plays, err := s.store.OnlineHistory(ctx, user)
		if err != nil {
			writeStoreError(w, r, err)
			return
		}
		imported, err := s.store.ImportedTracks(ctx)
		if err != nil {
			writeStoreError(w, r, err)
			return
		}
		rows := make([]mixedEntry, 0, len(plays))
		for _, p := range plays {
			song, err := netease.DecodeSong(p.Song)
			if err != nil {
				slog.Error("online history snapshot", "netease_id", p.NeteaseID, "err", err)
				continue
			}
			rows = append(rows, mixedEntry{
				at: parseStoreTime(p.PlayedAt), neteaseID: p.NeteaseID, song: &song, playedAt: p.PlayedAt,
			})
		}
		entries = mergeSameSong(local, rows, imported)
	}
	page := pageOf(entries, window(r))
	items := make([]any, len(page))
	for i, e := range page {
		var track any = e.song
		if e.local != nil {
			track = e.local
		}
		items[i] = mixedHistoryEntry{Track: track, PlayedAt: e.playedAt}
	}
	respond(w, r, mixedPage{Items: items, Total: len(entries)}, nil)
}

// recordOnlinePlay takes {"song": {...}}, the song as the player has it, and
// keeps it as the history snapshot.
func (s *server) recordOnlinePlay(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Song netease.Song `json:"song"`
	}
	if !admin.Decode(w, r, &in) {
		return
	}
	if in.Song.ID <= 0 || in.Song.Title == "" {
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "无效的歌曲")
		return
	}
	in.Song.Source = "online"
	raw, err := netease.EncodeSong(in.Song)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	if len(raw) > maxSnapshotBytes {
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "歌曲信息过长")
		return
	}
	respondEmpty(w, r, s.store.RecordOnlinePlay(r.Context(), userID(r), in.Song.ID, raw))
}

func (s *server) removeOnlinePlay(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(w, r, "id")
	if !ok {
		return
	}
	respondEmpty(w, r, s.store.RemoveFromOnlineHistory(r.Context(), userID(r), id))
}

func (s *server) listNeteasePlaylists(w http.ResponseWriter, r *http.Request) {
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	created, collected, err := s.opts.Netease.Playlists(r.Context(), session.Cookie, session.AccountID)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respond(w, r, map[string]any{"created": created, "collected": collected}, nil)
}

// onlineID reads a NetEase ID from the path.
func onlineID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		admin.WriteError(w, http.StatusNotFound, "not_found", "内容不存在或已被删除")
		return 0, false
	}
	return id, true
}

// songList answers with a NetEase list and its songs, whichever kind it is:
// the player shows all of them on one page.
func (s *server) songList(w http.ResponseWriter, r *http.Request,
	load func(ctx context.Context, cookie string, id int64) (netease.Playlist, []netease.Song, error)) {
	id, ok := onlineID(w, r)
	if !ok {
		return
	}
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	list, songs, err := load(r.Context(), session.Cookie, id)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respond(w, r, map[string]any{"playlist": list, "tracks": songs}, nil)
}

func (s *server) getNeteasePlaylist(w http.ResponseWriter, r *http.Request) {
	s.songList(w, r, s.opts.Netease.PlaylistSongs)
}

func (s *server) getNeteaseAlbum(w http.ResponseWriter, r *http.Request) {
	s.songList(w, r, s.opts.Netease.AlbumSongs)
}

func (s *server) getNeteaseArtist(w http.ResponseWriter, r *http.Request) {
	s.songList(w, r, s.opts.Netease.ArtistSongs)
}

func (s *server) listNeteaseArtistAlbums(w http.ResponseWriter, r *http.Request) {
	id, ok := onlineID(w, r)
	if !ok {
		return
	}
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	page := window(r)
	if page.Limit <= 0 {
		page.Limit = artistAlbumPage
	}
	page.Offset = max(page.Offset, 0)
	albums, total, err := s.opts.Netease.ArtistAlbums(r.Context(), session.Cookie, id, page.Offset, page.Limit)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respond(w, r, map[string]any{"items": albums, "total": total}, nil)
}

// artistAlbumPage is the page size when the request names none.
const artistAlbumPage = 50

// maxPlaylistBody caps a NetEase playlist's name and description.
const maxPlaylistBody = 16 << 10

// readNeteasePlaylist reads a playlist's fields from the request; the name
// must not be empty.
func readNeteasePlaylist(w http.ResponseWriter, r *http.Request) (name, description string, private, ok bool) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPlaylistBody)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "请求格式错误")
		return "", "", false, false
	}
	name = strings.TrimSpace(body.Name)
	if name == "" {
		writeNeteaseError(w, r, fmt.Errorf("%w: 歌单名称不能为空", store.ErrInvalid))
		return "", "", false, false
	}
	return name, strings.TrimSpace(body.Description), body.Private, true
}

func (s *server) createNeteasePlaylist(w http.ResponseWriter, r *http.Request) {
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	name, _, private, ok := readNeteasePlaylist(w, r)
	if !ok {
		return
	}
	playlist, err := s.opts.Netease.CreatePlaylist(r.Context(), session.Cookie, name, private)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respond(w, r, playlist, nil)
}

// ownNeteasePlaylist finds a playlist the account created, or answers the
// request. Writes go through it so that a stray ID cannot reach "liked songs"
// or a collected playlist: on NetEase those are playlists too.
func (s *server) ownNeteasePlaylist(w http.ResponseWriter, r *http.Request) (store.NeteaseSession, netease.Playlist, bool) {
	id, ok := onlineID(w, r)
	if !ok {
		return store.NeteaseSession{}, netease.Playlist{}, false
	}
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return session, netease.Playlist{}, false
	}
	created, _, err := s.opts.Netease.Playlists(r.Context(), session.Cookie, session.AccountID)
	if err != nil {
		writeNeteaseError(w, r, err)
		return session, netease.Playlist{}, false
	}
	for _, p := range created {
		if p.ID == id {
			return session, p, true
		}
	}
	writeNeteaseError(w, r, netease.ErrNotFound)
	return session, netease.Playlist{}, false
}

func (s *server) updateNeteasePlaylist(w http.ResponseWriter, r *http.Request) {
	session, current, ok := s.ownNeteasePlaylist(w, r)
	if !ok {
		return
	}
	name, description, _, ok := readNeteasePlaylist(w, r)
	if !ok {
		return
	}
	// Only what changed: these are two requests, and either can be refused.
	if name != current.Name {
		if err := s.opts.Netease.RenamePlaylist(r.Context(), session.Cookie, current.ID, name); err != nil {
			writeNeteaseError(w, r, err)
			return
		}
	}
	if description != current.Description {
		if err := s.opts.Netease.DescribePlaylist(r.Context(), session.Cookie, current.ID, description); err != nil {
			writeNeteaseError(w, r, err)
			return
		}
	}
	respondEmpty(w, r, nil)
}

func (s *server) deleteNeteasePlaylist(w http.ResponseWriter, r *http.Request) {
	session, current, ok := s.ownNeteasePlaylist(w, r)
	if !ok {
		return
	}
	if err := s.opts.Netease.DeletePlaylist(r.Context(), session.Cookie, current.ID); err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respondEmpty(w, r, nil)
}

func (s *server) addToNeteasePlaylist(w http.ResponseWriter, r *http.Request) {
	session, current, ok := s.ownNeteasePlaylist(w, r)
	if !ok {
		return
	}
	var body struct {
		IDs []int64 `json:"ids"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPlaylistBody)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "请求格式错误")
		return
	}
	if len(body.IDs) == 0 {
		writeNeteaseError(w, r, fmt.Errorf("%w: 没有选择歌曲", store.ErrInvalid))
		return
	}
	if err := s.opts.Netease.AddToPlaylist(r.Context(), session.Cookie, current.ID, body.IDs); err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respondEmpty(w, r, nil)
}

func (s *server) listNeteaseAlbums(w http.ResponseWriter, r *http.Request) {
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	albums, err := s.opts.Netease.CollectedAlbums(r.Context(), session.Cookie)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respond(w, r, map[string]any{"items": albums, "total": len(albums)}, nil)
}

func (s *server) listNeteaseDailySongs(w http.ResponseWriter, r *http.Request) {
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	songs, err := s.opts.Netease.DailySongs(r.Context(), session.Cookie)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respond(w, r, map[string]any{"items": songs, "total": len(songs)}, nil)
}

func (s *server) listNeteaseArtists(w http.ResponseWriter, r *http.Request) {
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	artists, err := s.opts.Netease.CollectedArtists(r.Context(), session.Cookie)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respond(w, r, map[string]any{"items": artists, "total": len(artists)}, nil)
}

func (s *server) getNeteaseSongURL(w http.ResponseWriter, r *http.Request) {
	id, ok := onlineID(w, r)
	if !ok {
		return
	}
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	u, err := s.opts.Netease.SongURL(r.Context(), session.Cookie, id)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	if u == "" {
		admin.WriteError(w, http.StatusNotFound, "netease_no_url", "这首歌暂时无法播放")
		return
	}
	respond(w, r, map[string]string{"url": u}, nil)
}

func (s *server) getNeteaseLyrics(w http.ResponseWriter, r *http.Request) {
	id, ok := onlineID(w, r)
	if !ok {
		return
	}
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	lyrics, err := s.opts.Netease.Lyrics(r.Context(), session.Cookie, id)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	text, format := lyrics.Compose()
	respond(w, r, map[string]string{"format": format, "lyrics": text}, nil)
}

func (s *server) listNeteaseLikeIDs(w http.ResponseWriter, r *http.Request) {
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	ids, err := s.opts.Netease.LikedIDs(r.Context(), session.Cookie, session.AccountID)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respond(w, r, map[string][]int64{"ids": ids}, nil)
}

func (s *server) setNeteaseLike(w http.ResponseWriter, r *http.Request) {
	id, ok := onlineID(w, r)
	if !ok {
		return
	}
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	err := s.opts.Netease.SetLike(r.Context(), session.Cookie, id, r.Method == http.MethodPut)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respondEmpty(w, r, nil)
}

// subscribeNeteaseAlbum collects (PUT) or uncollects (DELETE) an album on
// NetEase, like SPlayer's album page.
func (s *server) subscribeNeteaseAlbum(w http.ResponseWriter, r *http.Request) {
	s.setNeteaseSubscription(w, r, s.opts.Netease.SubscribeAlbum)
}

// subscribeNeteasePlaylist collects (PUT) or uncollects (DELETE) someone
// else's playlist on NetEase, like SPlayer's playlist page.
func (s *server) subscribeNeteasePlaylist(w http.ResponseWriter, r *http.Request) {
	s.setNeteaseSubscription(w, r, s.opts.Netease.SubscribePlaylist)
}

// followNeteaseArtist follows (PUT) or unfollows (DELETE) an artist on
// NetEase, like SPlayer's artist page.
func (s *server) followNeteaseArtist(w http.ResponseWriter, r *http.Request) {
	s.setNeteaseSubscription(w, r, s.opts.Netease.FollowArtist)
}

func (s *server) setNeteaseSubscription(w http.ResponseWriter, r *http.Request,
	set func(ctx context.Context, cookie string, id int64, on bool) error) {
	id, ok := onlineID(w, r)
	if !ok {
		return
	}
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return
	}
	if err := set(r.Context(), session.Cookie, id, r.Method == http.MethodPut); err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respondEmpty(w, r, nil)
}
