package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"sonora/internal/admin"
	"sonora/internal/auth"
	"sonora/internal/netease"
	"sonora/internal/store"
)

// Album imports (Eric, 2026-11-06) run in the background, since an album can
// be hundreds of megabytes: one album at a time, one song at a time, each song
// entering the library on its own. A single import can slip in between two
// songs. The queue lives in memory: after a restart the songs already in stay,
// and importing the album again adds the rest.

// Import job states.
const (
	importQueued  = "queued"
	importRunning = "running"
	importDone    = "done"
	// importFailed means the import stopped early, such as when the NetEase
	// session expired; Error says why.
	importFailed = "failed"
)

// albumImportGap is the least time between two songs an album import fetches
// from NetEase (Eric, 2026-11-07). It counts from the last fetch, so songs
// that fetch nothing, being in the library already or not playable, wait at
// most once after one that did. Tests shorten it.
var albumImportGap = 5 * time.Second

// keepFinishedImports is how many finished jobs the server remembers, so that
// a listener who reloads the page still learns how theirs went.
const keepFinishedImports = 50

// importFailure is a song an album import could not add.
type importFailure struct {
	Title  string `json:"title"`
	Reason string `json:"reason"`
}

// importJob is an album import. Created, Moved, Existing, Unplayable and
// Failed add up to Done.
type importJob struct {
	ID      int64  `json:"id"`
	AlbumID int64  `json:"albumId"`
	Title   string `json:"title"`
	State   string `json:"state"`
	Total   int    `json:"total"`
	Done    int    `json:"done"`
	Created int    `json:"created"`
	// Moved are songs imported on their own before, moved into the album.
	Moved int `json:"moved"`
	// Existing are songs already in the library, left where they are.
	Existing   int             `json:"existing"`
	Unplayable int             `json:"unplayable"`
	Failed     []importFailure `json:"failed"`
	Error      string          `json:"error"`

	user  int64
	album netease.ImportAlbum
}

// importQueue holds the album imports of all listeners.
type importQueue struct {
	mu      sync.Mutex
	jobs    []*importJob
	lastID  int64
	running bool
	// lastFetch is when the worker last fetched a song from NetEase. Only
	// the worker uses it, and workers run one after another.
	lastFetch time.Time
}

// add queues an import of album for user, or returns the one of theirs
// already waiting or running; start tells whether a worker must be started.
func (q *importQueue) add(user int64, album netease.ImportAlbum) (job importJob, start bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, j := range q.jobs {
		if j.user == user && j.AlbumID == album.ID && (j.State == importQueued || j.State == importRunning) {
			return j.snapshot(), false
		}
	}
	q.lastID++
	j := &importJob{
		ID: q.lastID, AlbumID: album.ID, Title: album.Title, State: importQueued,
		Total: len(album.Tracks), Failed: []importFailure{}, user: user, album: album,
	}
	q.jobs = append(q.jobs, j)
	start = !q.running
	q.running = true
	return j.snapshot(), start
}

// take marks the next queued job running, or returns nil and lets the worker
// stop when there is none.
func (q *importQueue) take() *importJob {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, j := range q.jobs {
		if j.State == importQueued {
			j.State = importRunning
			return j
		}
	}
	q.running = false
	return nil
}

// update changes a job under the lock.
func (q *importQueue) update(j *importJob, change func(*importJob)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	change(j)
	if j.State == importDone || j.State == importFailed {
		q.forgetOld()
	}
}

func (q *importQueue) forgetOld() {
	finished := 0
	for i := len(q.jobs) - 1; i >= 0; i-- {
		if s := q.jobs[i].State; s != importDone && s != importFailed {
			continue
		}
		if finished++; finished > keepFinishedImports {
			q.jobs = slices.Delete(q.jobs, i, i+1)
		}
	}
}

// of returns the user's jobs, newest first.
func (q *importQueue) of(user int64) []importJob {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := []importJob{}
	for i := len(q.jobs) - 1; i >= 0; i-- {
		if q.jobs[i].user == user {
			out = append(out, q.jobs[i].snapshot())
		}
	}
	return out
}

func (j *importJob) snapshot() importJob {
	c := *j
	c.Failed = slices.Clone(j.Failed)
	return c
}

// albumImportOf is what the store keeps of a NetEase album.
func albumImportOf(a netease.ImportAlbum) (store.AlbumImport, map[string]int64) {
	in := store.AlbumImport{
		NeteaseID: a.ID, Title: a.Title, Year: a.Year, Description: a.Description, Single: len(a.Tracks) == 1,
	}
	artists := a.Artists
	if len(artists) == 0 && len(a.Tracks) > 0 {
		artists = a.Tracks[0].Song.Artists
	}
	ids := map[string]int64{}
	for _, r := range artists {
		in.Artists = append(in.Artists, r.Name)
		ids[strings.TrimSpace(r.Name)] = r.ID
	}
	return in, ids
}

// albumImportPreview is what the player shows before an album import starts.
type albumImportPreview struct {
	Title string `json:"title"`
	Total int    `json:"total"`
	// ToImport are the playable songs not in the library yet.
	ToImport int `json:"toImport"`
	// InLibrary are songs already imported; Moving of them will be moved
	// into the album.
	InLibrary  int `json:"inLibrary"`
	Moving     int `json:"moving"`
	Unplayable int `json:"unplayable"`
	// Target is the library album the songs join, nil when the import
	// creates one.
	Target *playerAlbum `json:"target"`
	// SameTitle counts other albums with this title, which stay apart.
	SameTitle int `json:"sameTitle"`
}

func (s *server) importableAlbum(w http.ResponseWriter, r *http.Request) (netease.ImportAlbum, bool) {
	if !s.canImport(auth.User(r.Context())) {
		admin.WriteError(w, http.StatusForbidden, "import_forbidden", "没有导入歌曲的权限")
		return netease.ImportAlbum{}, false
	}
	id, ok := onlineID(w, r)
	if !ok {
		return netease.ImportAlbum{}, false
	}
	session, ok := s.neteaseSession(w, r)
	if !ok {
		return netease.ImportAlbum{}, false
	}
	album, err := s.opts.Netease.AlbumForImport(r.Context(), session.Cookie, id)
	if err != nil {
		writeNeteaseError(w, r, err)
		return album, false
	}
	return album, true
}

func (s *server) previewAlbumImport(w http.ResponseWriter, r *http.Request) {
	album, ok := s.importableAlbum(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	ids := make([]int64, len(album.Tracks))
	for i, t := range album.Tracks {
		ids[i] = t.Song.ID
	}
	places, err := s.store.ImportedPlacements(ctx, ids)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	out := albumImportPreview{Title: album.Title, Total: len(album.Tracks)}
	for _, t := range album.Tracks {
		p, in := places[t.Song.ID]
		switch {
		case in:
			out.InLibrary++
			if p.Alone {
				out.Moving++
			}
		case !t.Song.Playable:
			out.Unplayable++
		default:
			out.ToImport++
		}
	}
	target, err := s.store.ImportTarget(ctx, album.ID)
	switch {
	case err == nil:
		a := toPlayerAlbum(target)
		out.Target = &a
	case !errors.Is(err, store.ErrNotFound):
		writeStoreError(w, r, err)
		return
	}
	if out.SameTitle, err = s.store.SameTitleAlbums(ctx, album.Title, album.ID); err != nil {
		writeStoreError(w, r, err)
		return
	}
	respond(w, r, out, nil)
}

func (s *server) startAlbumImport(w http.ResponseWriter, r *http.Request) {
	album, ok := s.importableAlbum(w, r)
	if !ok {
		return
	}
	job, start := s.imports.add(userID(r), album)
	if start {
		go s.runImports()
	}
	w.Header().Set("Cache-Control", "no-store")
	admin.WriteJSON(w, http.StatusAccepted, job)
}

func (s *server) listImports(w http.ResponseWriter, r *http.Request) {
	if !s.canImport(auth.User(r.Context())) {
		admin.WriteError(w, http.StatusForbidden, "import_forbidden", "没有导入歌曲的权限")
		return
	}
	respond(w, r, map[string]any{"items": s.imports.of(userID(r))}, nil)
}

func (s *server) runImports() {
	for job := s.imports.take(); job != nil; job = s.imports.take() {
		s.runAlbumImport(job)
	}
}

// errSignedOut is the listener signing out of NetEase during an import.
var errSignedOut = errors.New("signed out of netease")

// Song outcomes of an album import.
const (
	songCreated    = "created"
	songMoved      = "moved"
	songExisting   = "existing"
	songUnplayable = "unplayable"
)

func (s *server) runAlbumImport(job *importJob) {
	album, artists := albumImportOf(job.album)
	covers := map[string][]byte{}
	for _, t := range job.album.Tracks {
		outcome, err := s.importAlbumSong(job.user, songImport{
			song: t.Song, album: &album, albumArtists: artists, trackNo: t.TrackNo, discNo: t.DiscNo, covers: covers,
		})
		if err != nil && abortsImport(err) {
			slog.Warn("album import stopped", "album", job.AlbumID, "err", err)
			s.imports.update(job, func(j *importJob) {
				j.State, j.Error = importFailed, abortReason(err)
			})
			return
		}
		if err != nil {
			slog.Warn("album import song", "album", job.AlbumID, "song", t.Song.ID, "err", err)
		}
		s.imports.update(job, func(j *importJob) {
			j.Done++
			switch {
			case err != nil:
				j.Failed = append(j.Failed, importFailure{Title: t.Song.Title, Reason: failureReason(err)})
			case outcome == songCreated:
				j.Created++
			case outcome == songMoved:
				j.Moved++
			case outcome == songExisting:
				j.Existing++
			case outcome == songUnplayable:
				j.Unplayable++
			}
		})
	}
	s.imports.update(job, func(j *importJob) { j.State = importDone })
}

// importAlbumSong adds one song of an album import. The session is read for
// every song, so signing out of NetEase stops the import.
func (s *server) importAlbumSong(user int64, in songImport) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), importTimeout)
	defer cancel()
	session, err := s.store.NeteaseSession(ctx, user)
	if errors.Is(err, store.ErrNotFound) {
		return "", errSignedOut
	}
	if err != nil {
		return "", err
	}
	// Wait outside the lock, so that a single import is not held up.
	time.Sleep(time.Until(s.imports.lastFetch.Add(albumImportGap)))
	s.importing.Lock()
	defer s.importing.Unlock()
	places, err := s.store.ImportedPlacements(ctx, []int64{in.song.ID})
	if err != nil {
		return "", err
	}
	if p, ok := places[in.song.ID]; ok {
		if !p.Alone {
			return songExisting, nil
		}
		// Moving creates the album, and with it album artists that may be new.
		added := s.newArtists(ctx, in.albumArtistRefs())
		moved, err := s.store.MoveImported(ctx, p.TrackID, *in.album, in.albumArtists, in.trackNo, in.discNo)
		if err != nil || !moved {
			return songExisting, err
		}
		if len(added) > 0 {
			s.fillNewArtists(ctx, session.Cookie, added)
			s.imports.lastFetch = time.Now()
		}
		return songMoved, nil
	}
	if !in.song.Playable {
		return songUnplayable, nil
	}
	added := s.newArtists(ctx, append(slices.Clone(in.song.Artists), in.albumArtistRefs()...))
	res, err := s.downloadSong(ctx, session.Cookie, in)
	if err == nil && res.Status == store.Created {
		s.fillNewArtists(ctx, session.Cookie, added)
	}
	s.imports.lastFetch = time.Now()
	if err != nil {
		return "", err
	}
	if res.Status == store.Duplicate {
		return songExisting, nil
	}
	return songCreated, nil
}

// abortsImport tells the failures that would fail every song after this one,
// NetEase limiting the account among them.
func abortsImport(err error) bool {
	return errors.Is(err, errSignedOut) || errors.Is(err, netease.ErrExpired) || errors.Is(err, netease.ErrThrottled) ||
		(errors.Is(err, netease.ErrUnavailable) && !errors.Is(err, errDownload))
}

func abortReason(err error) string {
	switch {
	case errors.Is(err, errSignedOut):
		return "已退出网易云"
	case errors.Is(err, netease.ErrExpired):
		return "网易云登录已失效，请重新登录"
	case errors.Is(err, netease.ErrThrottled):
		// Each song again would be refused as well, and press NetEase more.
		return "网易云请求过于频繁，请稍后再导入"
	}
	return "网易云接口服务不可用，请稍后重试"
}

func failureReason(err error) string {
	var rejected *netease.RejectedError
	switch {
	case errors.Is(err, errNoFile):
		return "账号没有完整版本"
	case errors.Is(err, errNotMP3):
		return "不是 MP3"
	case errors.As(err, &rejected):
		return rejected.Message
	case errors.Is(err, errDownload):
		return "下载失败"
	}
	return "入库失败"
}
