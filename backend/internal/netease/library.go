package netease

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// songDetailChunk is how many songs one /song/detail call takes.
	songDetailChunk = 500
	// likedPlaylist is the specialType of a user's "liked songs" playlist.
	likedPlaylist = 5
	// sublistPage is the page size for an account's collected albums and artists.
	sublistPage = 100
	// maxSublistPages bounds that walk, in case hasMore never turns false.
	maxSublistPages = 100
)

// The JSON of the types below matches the player API's own tracks, playlists,
// albums and artists, so the player's lists take either; Source tells them
// apart, and Cover.URL stands in for a local cover's content hash.

// Ref names an artist.
type Ref struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Cover is a NetEase image. Name and Thumbhash stay empty: they belong to
// local covers.
type Cover struct {
	Name      string `json:"name"`
	Thumbhash string `json:"thumbhash"`
	URL       string `json:"url"`
}

// AlbumRef is the album a song belongs to.
type AlbumRef struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Artists []Ref  `json:"artists"`
}

// Song is a NetEase song in the player's track shape.
type Song struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Artists    []Ref     `json:"artists"`
	Album      *AlbumRef `json:"album"`
	TrackNo    int       `json:"trackNo"`
	DiscNo     int       `json:"discNo"`
	Year       int       `json:"year"`
	DurationMs int64     `json:"durationMs"`
	Format     string    `json:"format"`
	Cover      *Cover    `json:"cover"`
	HasLyrics  bool      `json:"hasLyrics"`
	Source     string    `json:"source"`
	// Playable is false for a song NetEase offers nobody, such as one
	// whose rights ended. A song is playable to some accounts only by a
	// trial; that surfaces when its address is requested.
	Playable bool `json:"playable"`
}

// Playlist is a NetEase playlist in the player's playlist shape.
type Playlist struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cover       *Cover `json:"cover"`
	TrackCount  int    `json:"trackCount"`
	DurationMs  int64  `json:"durationMs"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Source      string `json:"source"`
}

// Album is an album the account collected.
type Album struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Artists    []Ref  `json:"artists"`
	Cover      *Cover `json:"cover"`
	Year       int    `json:"year"`
	TrackCount int    `json:"trackCount"`
	DurationMs int64  `json:"durationMs"`
	Source     string `json:"source"`
}

// Artist is an artist the account collected.
type Artist struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Image      *Cover `json:"image"`
	TrackCount int    `json:"trackCount"`
	AlbumCount int    `json:"albumCount"`
	Source     string `json:"source"`
}

// Liked is one entry of the "liked songs" playlist.
type Liked struct {
	ID int64
	At time.Time
}

const sourceOnline = "online"

func cover(u string) *Cover {
	if u == "" {
		return nil
	}
	return &Cover{URL: strings.Replace(u, "http://", "https://", 1)}
}

func millis(ms int64) string {
	if ms <= 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format(time.RFC3339)
}

func refs(in []Ref) []Ref {
	if in == nil {
		return []Ref{}
	}
	return in
}

type apiSong struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Artists     []Ref  `json:"ar"`
	Dt          int64  `json:"dt"`
	PublishTime int64  `json:"publishTime"`
	Album       struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		PicURL string `json:"picUrl"`
	} `json:"al"`
	// Privilege comes inside the song from /album and /artists; /song/detail
	// returns privileges separately.
	Privilege *struct {
		St int `json:"st"`
	} `json:"privilege"`
}

// blockedStatus says from a privilege's st whether NetEase offers the song to nobody.
func blockedStatus(st int) bool { return st < 0 }

// listed reads a song that carries its own privilege.
func (a apiSong) listed() Song {
	return a.song(a.Privilege == nil || !blockedStatus(a.Privilege.St))
}

func (a apiSong) song(playable bool) Song {
	out := Song{
		ID: a.ID, Title: a.Name, Artists: refs(a.Artists), DurationMs: a.Dt,
		Format: "mp3", Cover: cover(a.Album.PicURL), HasLyrics: true,
		Source: sourceOnline, Playable: playable,
	}
	if a.Album.ID != 0 {
		out.Album = &AlbumRef{ID: a.Album.ID, Title: a.Album.Name, Artists: []Ref{}}
	}
	if a.PublishTime > 0 {
		out.Year = time.UnixMilli(a.PublishTime).UTC().Year()
	}
	return out
}

// Songs returns the songs in the order of ids, leaving out any NetEase no
// longer knows.
func (c *Client) Songs(ctx context.Context, cookie string, ids []int64) ([]Song, error) {
	found := make(map[int64]Song, len(ids))
	for start := 0; start < len(ids); start += songDetailChunk {
		chunk := ids[start:min(start+songDetailChunk, len(ids))]
		parts := make([]string, len(chunk))
		for i, id := range chunk {
			parts[i] = strconv.FormatInt(id, 10)
		}
		var out struct {
			Songs      []apiSong `json:"songs"`
			Privileges []struct {
				ID int64 `json:"id"`
				St int   `json:"st"`
			} `json:"privileges"`
		}
		if err := c.call(ctx, "/song/detail", cookie, url.Values{"ids": {strings.Join(parts, ",")}}, &out); err != nil {
			return nil, err
		}
		blocked := map[int64]bool{}
		for _, p := range out.Privileges {
			blocked[p.ID] = blockedStatus(p.St)
		}
		for _, s := range out.Songs {
			found[s.ID] = s.song(!blocked[s.ID])
		}
	}
	songs := make([]Song, 0, len(ids))
	for _, id := range ids {
		if s, ok := found[id]; ok {
			songs = append(songs, s)
		}
	}
	return songs, nil
}

type apiPlaylist struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CoverImgURL string  `json:"coverImgUrl"`
	TrackCount  int     `json:"trackCount"`
	Subscribed  bool    `json:"subscribed"`
	SpecialType int     `json:"specialType"`
	CreateTime  int64   `json:"createTime"`
	UpdateTime  int64   `json:"updateTime"`
	TrackIDs    []struct {
		ID int64 `json:"id"`
		At int64 `json:"at"`
	} `json:"trackIds"`
}

func (a apiPlaylist) playlist() Playlist {
	out := Playlist{
		ID: a.ID, Name: a.Name, Cover: cover(a.CoverImgURL), TrackCount: a.TrackCount,
		CreatedAt: millis(a.CreateTime), UpdatedAt: millis(a.UpdateTime), Source: sourceOnline,
	}
	if a.Description != nil {
		out.Description = *a.Description
	}
	return out
}

func (c *Client) userPlaylists(ctx context.Context, cookie string, account int64) ([]apiPlaylist, error) {
	var out struct {
		Playlist []apiPlaylist `json:"playlist"`
	}
	params := url.Values{"uid": {strconv.FormatInt(account, 10)}, "limit": {"1000"}}
	if err := c.call(ctx, "/user/playlist", cookie, params, &out); err != nil {
		return nil, err
	}
	return out.Playlist, nil
}

// Playlists returns the account's own playlists and the ones it collected.
// "Liked songs" is in neither: the sidebar's liked entry covers it.
func (c *Client) Playlists(ctx context.Context, cookie string, account int64) (created, collected []Playlist, err error) {
	all, err := c.userPlaylists(ctx, cookie, account)
	if err != nil {
		return nil, nil, err
	}
	created, collected = []Playlist{}, []Playlist{}
	for _, p := range all {
		switch {
		case p.SpecialType == likedPlaylist:
		case p.Subscribed:
			collected = append(collected, p.playlist())
		default:
			created = append(created, p.playlist())
		}
	}
	return created, collected, nil
}

func (c *Client) playlistDetail(ctx context.Context, cookie string, id int64) (apiPlaylist, error) {
	var out struct {
		Playlist *apiPlaylist `json:"playlist"`
	}
	params := url.Values{"id": {strconv.FormatInt(id, 10)}}
	if err := c.call(ctx, "/playlist/detail", cookie, params, &out); err != nil {
		return apiPlaylist{}, err
	}
	if out.Playlist == nil {
		return apiPlaylist{}, ErrNotFound
	}
	return *out.Playlist, nil
}

// PlaylistSongs returns a playlist with all of its songs.
func (c *Client) PlaylistSongs(ctx context.Context, cookie string, id int64) (Playlist, []Song, error) {
	p, err := c.playlistDetail(ctx, cookie, id)
	if err != nil {
		return Playlist{}, nil, err
	}
	ids := make([]int64, len(p.TrackIDs))
	for i, t := range p.TrackIDs {
		ids[i] = t.ID
	}
	songs, err := c.Songs(ctx, cookie, ids)
	return p.playlist(), songs, err
}

// Likes returns the account's liked songs, most recently liked first. Unlike
// /likelist, the playlist says when each one was liked.
func (c *Client) Likes(ctx context.Context, cookie string, account int64) ([]Liked, error) {
	all, err := c.userPlaylists(ctx, cookie, account)
	if err != nil {
		return nil, err
	}
	for _, p := range all {
		if p.SpecialType != likedPlaylist {
			continue
		}
		detail, err := c.playlistDetail(ctx, cookie, p.ID)
		if err != nil {
			return nil, err
		}
		liked := make([]Liked, len(detail.TrackIDs))
		for i, t := range detail.TrackIDs {
			liked[i] = Liked{ID: t.ID, At: time.UnixMilli(t.At)}
		}
		return liked, nil
	}
	return []Liked{}, nil
}

// sublist walks a paged list of what the account collected.
func sublist[T any](ctx context.Context, c *Client, path, cookie string) ([]T, error) {
	all := []T{}
	for page := range maxSublistPages {
		var out struct {
			Data    []T  `json:"data"`
			HasMore bool `json:"hasMore"`
		}
		params := url.Values{
			"limit":  {strconv.Itoa(sublistPage)},
			"offset": {strconv.Itoa(page * sublistPage)},
		}
		if err := c.call(ctx, path, cookie, params, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Data...)
		if !out.HasMore {
			break
		}
	}
	return all, nil
}

// CollectedAlbums returns the albums the account collected.
func (c *Client) CollectedAlbums(ctx context.Context, cookie string) ([]Album, error) {
	data, err := sublist[struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		PicURL  string `json:"picUrl"`
		Size    int    `json:"size"`
		Artists []Ref  `json:"artists"`
	}](ctx, c, "/album/sublist", cookie)
	if err != nil {
		return nil, err
	}
	albums := make([]Album, len(data))
	for i, a := range data {
		albums[i] = Album{
			ID: a.ID, Title: a.Name, Artists: refs(a.Artists), Cover: cover(a.PicURL),
			TrackCount: a.Size, Source: sourceOnline,
		}
	}
	return albums, nil
}

// CollectedArtists returns the artists the account collected.
func (c *Client) CollectedArtists(ctx context.Context, cookie string) ([]Artist, error) {
	data, err := sublist[struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		PicURL    string `json:"picUrl"`
		AlbumSize int    `json:"albumSize"`
	}](ctx, c, "/artist/sublist", cookie)
	if err != nil {
		return nil, err
	}
	artists := make([]Artist, len(data))
	for i, a := range data {
		artists[i] = Artist{
			ID: a.ID, Name: a.Name, Image: cover(a.PicURL), AlbumCount: a.AlbumSize, Source: sourceOnline,
		}
	}
	return artists, nil
}

// EncodeSong and DecodeSong keep a song as a history snapshot, which survives
// the song leaving NetEase.
func EncodeSong(s Song) (string, error) {
	raw, err := json.Marshal(s)
	return string(raw), err
}

// DecodeSong reads a snapshot written by EncodeSong.
func DecodeSong(raw string) (Song, error) {
	var s Song
	err := json.Unmarshal([]byte(raw), &s)
	return s, err
}

// playbackLevel is the quality asked for, SPlayer's default.
const playbackLevel = "exhigh"

// SongFile is the file behind a song at the playback quality.
type SongFile struct {
	// URL is signed and short-lived, so it is fetched per use. "" when the
	// account gets no full version: SPlayer does not play trials unless asked
	// to, and neither does Sonora.
	URL string
	// Type is the file type NetEase reports, such as "mp3".
	Type string
}

// SongFile asks where the song's file is.
func (c *Client) SongFile(ctx context.Context, cookie string, id int64) (SongFile, error) {
	var out struct {
		Data []struct {
			URL           string          `json:"url"`
			Type          string          `json:"type"`
			FreeTrialInfo json.RawMessage `json:"freeTrialInfo"`
		} `json:"data"`
	}
	params := url.Values{"id": {strconv.FormatInt(id, 10)}, "level": {playbackLevel}}
	if err := c.call(ctx, "/song/url/v1", cookie, params, &out); err != nil {
		return SongFile{}, err
	}
	if len(out.Data) == 0 || out.Data[0].URL == "" {
		return SongFile{}, nil
	}
	if t := string(out.Data[0].FreeTrialInfo); t != "" && t != "null" {
		return SongFile{}, nil
	}
	return SongFile{URL: playableURL(out.Data[0].URL), Type: out.Data[0].Type}, nil
}

// SongURL returns where to play the song, "" when it cannot be played.
func (c *Client) SongURL(ctx context.Context, cookie string, id int64) (string, error) {
	f, err := c.SongFile(ctx, cookie, id)
	return f.URL, err
}

// Download writes the file at u to w, at most limit bytes. It goes to NetEase's
// CDN rather than the API service and sends no session: the address is signed.
func (c *Client) Download(ctx context.Context, u string, w io.Writer, limit int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("%w: download: bad address", ErrUnavailable)
	}
	resp, err := c.Downloads.Do(req)
	if err != nil {
		return fmt.Errorf("%w: download: request failed", ErrUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: download: status %d", ErrUnavailable, resp.StatusCode)
	}
	n, err := io.Copy(w, io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return fmt.Errorf("%w: download interrupted", ErrUnavailable)
	}
	if n > limit {
		return &RejectedError{Message: "文件超过大小限制"}
	}
	return nil
}

// playableURL makes an address the page can use, as SPlayer does on the web:
// the page is https, and m804 and m704 do not serve browsers.
func playableURL(u string) string {
	u = strings.Replace(u, "http://", "https://", 1)
	u = strings.ReplaceAll(u, "m804.music.126.net", "m801.music.126.net")
	return strings.ReplaceAll(u, "m704.music.126.net", "m701.music.126.net")
}

// Lyrics returns everything NetEase has for the song's lyric.
func (c *Client) Lyrics(ctx context.Context, cookie string, id int64) (LyricSet, error) {
	type part struct {
		Lyric string `json:"lyric"`
	}
	var out struct {
		Lrc      part `json:"lrc"`
		Tlyric   part `json:"tlyric"`
		Romalrc  part `json:"romalrc"`
		Yrc      part `json:"yrc"`
		Ytlrc    part `json:"ytlrc"`
		Yromalrc part `json:"yromalrc"`
	}
	params := url.Values{"id": {strconv.FormatInt(id, 10)}}
	if err := c.call(ctx, "/lyric/new", cookie, params, &out); err != nil {
		return LyricSet{}, err
	}
	return LyricSet{
		LRC: out.Lrc.Lyric, Translation: out.Tlyric.Lyric, Roman: out.Romalrc.Lyric,
		YRC: out.Yrc.Lyric, YTranslated: out.Ytlrc.Lyric, YRoman: out.Yromalrc.Lyric,
	}, nil
}

// LikedIDs returns the IDs of the account's liked songs.
func (c *Client) LikedIDs(ctx context.Context, cookie string, account int64) ([]int64, error) {
	var out struct {
		IDs []int64 `json:"ids"`
	}
	params := url.Values{"uid": {strconv.FormatInt(account, 10)}}
	if err := c.call(ctx, "/likelist", cookie, params, &out); err != nil {
		return nil, err
	}
	if out.IDs == nil {
		return []int64{}, nil
	}
	return out.IDs, nil
}

// SetLike likes or unlikes the song on NetEase.
func (c *Client) SetLike(ctx context.Context, cookie string, id int64, like bool) error {
	var out struct {
		Code int `json:"code"`
	}
	params := url.Values{"id": {strconv.FormatInt(id, 10)}, "like": {strconv.FormatBool(like)}}
	if err := c.call(ctx, "/like", cookie, params, &out); err != nil {
		return err
	}
	if out.Code != http.StatusOK {
		return fmt.Errorf("%w: /like: code %d", ErrUnavailable, out.Code)
	}
	return nil
}

// AlbumSongs returns an album in the playlist shape, with its songs.
func (c *Client) AlbumSongs(ctx context.Context, cookie string, id int64) (Playlist, []Song, error) {
	var out struct {
		Album *struct {
			Name        string  `json:"name"`
			PicURL      string  `json:"picUrl"`
			Description *string `json:"description"`
		} `json:"album"`
		Songs []apiSong `json:"songs"`
	}
	params := url.Values{"id": {strconv.FormatInt(id, 10)}}
	if err := c.call(ctx, "/album", cookie, params, &out); err != nil {
		return Playlist{}, nil, err
	}
	if out.Album == nil {
		return Playlist{}, nil, ErrNotFound
	}
	list := Playlist{
		ID: id, Name: out.Album.Name, Cover: cover(out.Album.PicURL), TrackCount: len(out.Songs), Source: sourceOnline,
	}
	if out.Album.Description != nil {
		list.Description = *out.Album.Description
	}
	songs := make([]Song, len(out.Songs))
	for i, s := range out.Songs {
		songs[i] = s.listed()
		if songs[i].Cover == nil {
			songs[i].Cover = list.Cover
		}
	}
	return list, songs, nil
}

// ImportAlbum is a NetEase album as an album import needs it.
type ImportAlbum struct {
	ID    int64
	Title string
	// Artists are NetEase's album artists, such as "群星" for a compilation.
	Artists     []Ref
	Year        int
	Description string
	Tracks      []AlbumTrack
}

// AlbumTrack is a song of an album with its place on it. Song.TrackNo and
// Song.DiscNo stay 0, as everywhere the player shows online songs.
type AlbumTrack struct {
	Song    Song
	TrackNo int
	DiscNo  int
}

// leadingNumber reads a NetEase track or disc number, which comes as a number
// or as a string such as "01" or "02 [Target Deluxe Edition]"; 0 when it has
// none.
type leadingNumber int

func (n *leadingNumber) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	v, _ := strconv.Atoi(s[:min(end, 6)])
	*n = leadingNumber(v)
	return nil
}

// AlbumForImport returns the album with what importing it writes: album
// artists, year, description and each song's track and disc number.
func (c *Client) AlbumForImport(ctx context.Context, cookie string, id int64) (ImportAlbum, error) {
	var out struct {
		Album *struct {
			Name        string  `json:"name"`
			PicURL      string  `json:"picUrl"`
			Description *string `json:"description"`
			PublishTime int64   `json:"publishTime"`
			Artists     []Ref   `json:"artists"`
			Artist      *Ref    `json:"artist"`
		} `json:"album"`
		Songs []struct {
			apiSong
			No leadingNumber `json:"no"`
			Cd leadingNumber `json:"cd"`
		} `json:"songs"`
	}
	params := url.Values{"id": {strconv.FormatInt(id, 10)}}
	if err := c.call(ctx, "/album", cookie, params, &out); err != nil {
		return ImportAlbum{}, err
	}
	if out.Album == nil {
		return ImportAlbum{}, ErrNotFound
	}
	a := ImportAlbum{ID: id, Title: out.Album.Name, Artists: refs(out.Album.Artists)}
	if len(a.Artists) == 0 && out.Album.Artist != nil && out.Album.Artist.Name != "" {
		a.Artists = []Ref{*out.Album.Artist}
	}
	if out.Album.Description != nil {
		a.Description = *out.Album.Description
	}
	if out.Album.PublishTime > 0 {
		a.Year = time.UnixMilli(out.Album.PublishTime).UTC().Year()
	}
	albumCover := cover(out.Album.PicURL)
	for _, s := range out.Songs {
		song := s.listed()
		if song.Cover == nil {
			song.Cover = albumCover
		}
		// The songs of /album carry no publishTime of their own.
		if song.Year == 0 {
			song.Year = a.Year
		}
		a.Tracks = append(a.Tracks, AlbumTrack{Song: song, TrackNo: int(s.No), DiscNo: int(s.Cd)})
	}
	return a, nil
}

// ArtistSongs returns an artist in the playlist shape, with their top songs.
func (c *Client) ArtistSongs(ctx context.Context, cookie string, id int64) (Playlist, []Song, error) {
	var out struct {
		Artist *struct {
			Name      string `json:"name"`
			PicURL    string `json:"picUrl"`
			BriefDesc string `json:"briefDesc"`
		} `json:"artist"`
		HotSongs []apiSong `json:"hotSongs"`
	}
	params := url.Values{"id": {strconv.FormatInt(id, 10)}}
	if err := c.call(ctx, "/artists", cookie, params, &out); err != nil {
		return Playlist{}, nil, err
	}
	if out.Artist == nil {
		return Playlist{}, nil, ErrNotFound
	}
	songs := make([]Song, len(out.HotSongs))
	for i, s := range out.HotSongs {
		songs[i] = s.listed()
	}
	return Playlist{
		ID: id, Name: out.Artist.Name, Description: out.Artist.BriefDesc, Cover: cover(out.Artist.PicURL),
		TrackCount: len(songs), Source: sourceOnline,
	}, songs, nil
}

// ArtistAlbums returns a page of the artist's albums and how many there are.
// DailySongs returns the account's daily recommendations; NetEase renews them
// at 6:00.
func (c *Client) DailySongs(ctx context.Context, cookie string) ([]Song, error) {
	var out struct {
		Data struct {
			DailySongs []apiSong `json:"dailySongs"`
		} `json:"data"`
	}
	if err := c.call(ctx, "/recommend/songs", cookie, url.Values{}, &out); err != nil {
		return nil, err
	}
	songs := make([]Song, 0, len(out.Data.DailySongs))
	for _, s := range out.Data.DailySongs {
		songs = append(songs, s.listed())
	}
	return songs, nil
}

func (c *Client) ArtistAlbums(ctx context.Context, cookie string, id int64, offset, limit int) ([]Album, int, error) {
	var out struct {
		Artist *struct {
			AlbumSize int `json:"albumSize"`
		} `json:"artist"`
		HotAlbums []struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			PicURL      string `json:"picUrl"`
			Size        int    `json:"size"`
			PublishTime int64  `json:"publishTime"`
			Artists     []Ref  `json:"artists"`
		} `json:"hotAlbums"`
		More bool `json:"more"`
	}
	params := url.Values{
		"id":     {strconv.FormatInt(id, 10)},
		"limit":  {strconv.Itoa(limit)},
		"offset": {strconv.Itoa(offset)},
	}
	if err := c.call(ctx, "/artist/album", cookie, params, &out); err != nil {
		return nil, 0, err
	}
	if out.Artist == nil {
		return nil, 0, ErrNotFound
	}
	albums := make([]Album, len(out.HotAlbums))
	for i, a := range out.HotAlbums {
		albums[i] = Album{
			ID: a.ID, Title: a.Name, Artists: refs(a.Artists), Cover: cover(a.PicURL),
			TrackCount: a.Size, Source: sourceOnline,
		}
		if a.PublishTime > 0 {
			albums[i].Year = time.UnixMilli(a.PublishTime).UTC().Year()
		}
	}
	// albumSize can count albums the list does not return; without this the
	// player would keep asking for a next page that never comes.
	total := out.Artist.AlbumSize
	if !out.More {
		total = offset + len(albums)
	}
	return albums, total, nil
}

// writeResult is what the write endpoints answer.
type writeResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Msg     string `json:"msg"`
}

func (w writeResult) check() error {
	switch w.Code {
	case http.StatusOK:
		return nil
	case needLogin:
		return ErrExpired
	}
	return &RejectedError{Message: cmp.Or(w.Message, w.Msg)}
}

func (c *Client) write(ctx context.Context, path, cookie string, params url.Values) error {
	var out writeResult
	if err := c.call(ctx, path, cookie, params, &out); err != nil {
		return err
	}
	return out.check()
}

// CreatePlaylist creates a playlist on the account. NetEase takes only a name
// and a privacy here; a description has to be set afterwards.
func (c *Client) CreatePlaylist(ctx context.Context, cookie, name string, private bool) (Playlist, error) {
	privacy := "0"
	if private {
		privacy = "10"
	}
	var out struct {
		writeResult
		Playlist *apiPlaylist `json:"playlist"`
	}
	params := url.Values{"name": {name}, "privacy": {privacy}}
	if err := c.call(ctx, "/playlist/create", cookie, params, &out); err != nil {
		return Playlist{}, err
	}
	if err := out.check(); err != nil {
		return Playlist{}, err
	}
	if out.Playlist == nil {
		return Playlist{}, fmt.Errorf("%w: /playlist/create: no playlist", ErrUnavailable)
	}
	return out.Playlist.playlist(), nil
}

// RenamePlaylist and DescribePlaylist are separate because the service's
// /playlist/update clears the tags unless they are sent again, and Sonora does
// not show tags. It also pastes the values into JSON unescaped.
func (c *Client) RenamePlaylist(ctx context.Context, cookie string, id int64, name string) error {
	return c.write(ctx, "/playlist/name/update", cookie,
		url.Values{"id": {strconv.FormatInt(id, 10)}, "name": {name}})
}

// DescribePlaylist sets the playlist's description.
func (c *Client) DescribePlaylist(ctx context.Context, cookie string, id int64, description string) error {
	return c.write(ctx, "/playlist/desc/update", cookie,
		url.Values{"id": {strconv.FormatInt(id, 10)}, "desc": {description}})
}

// DeletePlaylist deletes the playlist from the account.
func (c *Client) DeletePlaylist(ctx context.Context, cookie string, id int64) error {
	return c.write(ctx, "/playlist/delete", cookie, url.Values{"id": {strconv.FormatInt(id, 10)}})
}

// AddToPlaylist adds songs to the playlist. The service wraps this endpoint's
// answer in a "body" when it succeeds and returns it bare when NetEase refuses.
func (c *Client) AddToPlaylist(ctx context.Context, cookie string, id int64, songs []int64) error {
	parts := make([]string, len(songs))
	for i, song := range songs {
		parts[i] = strconv.FormatInt(song, 10)
	}
	var out struct {
		writeResult
		Body *writeResult `json:"body"`
	}
	params := url.Values{
		"op": {"add"}, "pid": {strconv.FormatInt(id, 10)}, "tracks": {strings.Join(parts, ",")},
	}
	if err := c.call(ctx, "/playlist/tracks", cookie, params, &out); err != nil {
		return err
	}
	if out.Body != nil {
		return out.Body.check()
	}
	return out.check()
}

// SubscribeAlbum collects or uncollects the album on the account.
func (c *Client) SubscribeAlbum(ctx context.Context, cookie string, id int64, on bool) error {
	return c.write(ctx, "/album/sub", cookie, url.Values{"id": {strconv.FormatInt(id, 10)}, "t": {subscribeFlag(on)}})
}

// FollowArtist follows or unfollows the artist on the account.
func (c *Client) FollowArtist(ctx context.Context, cookie string, id int64, on bool) error {
	return c.write(ctx, "/artist/sub", cookie, url.Values{"id": {strconv.FormatInt(id, 10)}, "t": {subscribeFlag(on)}})
}

// SubscribePlaylist collects or uncollects someone else's playlist on the
// account.
func (c *Client) SubscribePlaylist(ctx context.Context, cookie string, id int64, on bool) error {
	return c.write(ctx, "/playlist/subscribe", cookie, url.Values{"id": {strconv.FormatInt(id, 10)}, "t": {subscribeFlag(on)}})
}

// subscribeFlag is the t of /album/sub, /artist/sub and /playlist/subscribe: 1 subscribes, any
// other value unsubscribes, as SPlayer sends it.
func subscribeFlag(on bool) string {
	if on {
		return "1"
	}
	return "0"
}

// Links are where NetEase shows a song, album, artist or playlist on the web,
// with "{id}" for its ID. The player has no NetEase addresses of its own, so
// it builds the share links and source pages from these. SPlayer's default
// "web" format; its "mobile" format is not offered.
var Links = map[string]string{
	"song":     "https://music.163.com/#/song?id={id}",
	"album":    "https://music.163.com/#/album?id={id}",
	"artist":   "https://music.163.com/#/artist?id={id}",
	"playlist": "https://music.163.com/#/playlist?id={id}",
}
