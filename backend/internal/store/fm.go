package store

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"math/rand/v2"
	"slices"
)

// The private radio: an endless list of whole albums for one listener.
//
// A cycle is one big album followed by a segment of small ones. Within a
// pool the albums heard the fewest times come first, ties at random, so
// over time every album is heard about as often. Small albums only take
// part once one of the listener's group playlists holds them, and a segment
// keeps each group's albums together, the groups in a random order.
//
// Nothing here depends on rounds or on the size of the library when a plan
// was drawn: each draw looks at the current counts, so albums coming and
// going even out on their own.

const (
	// fmBigAlbumTracks is the track count from which an album is big.
	fmBigAlbumTracks = 10
	// fmSoloSegment is the segment size when no big album sets the pace.
	fmSoloSegment = 20
	// fmLookahead is how many albums the plan keeps lined up, the current
	// one included, so the player can load the next one early.
	fmLookahead = 2
	// FMMaxAhead caps how many albums one read may ask for, the current one
	// included. The Android home screen shows five.
	FMMaxAhead = 10
	// fmMaxDraws bounds the draws one call makes when albums keep turning
	// out to have nothing to add.
	fmMaxDraws = 4
)

// FMAlbum is a planned album with its tracks in playing order.
type FMAlbum struct {
	ID     int64
	Tracks []Track
}

// FMState is the listener's radio: the current album, the next ones, the
// track to resume the current album at (nil from its start) and the
// playlists used as groups, in priority order.
type FMState struct {
	Albums   []FMAlbum
	TrackID  *int64
	GroupIDs []int64
}

// fmCandidate is an album that can be drawn. Group is 0 for a big album,
// otherwise the playlist the small album belongs to. Blocked albums count
// towards the averages but are not drawn: they are in the plan already or
// were just skipped.
type fmCandidate struct {
	id      int64
	plays   int
	group   int64
	blocked bool
}

// fmSegmentSize is how many small albums follow a big one: as many as bring
// the small albums' average count up to the big albums', which on a steady
// library comes to small albums ÷ big albums. A sudden change in the library
// is made up over several segments instead of in one huge or tiny segment.
func fmSegmentSize(bigPlays, bigs, smallPlays, smalls int) int {
	if smalls == 0 {
		return 0
	}
	if bigs == 0 {
		return fmSoloSegment
	}
	ratio := float64(smalls) / float64(bigs)
	want := math.Ceil(float64(bigPlays)/float64(bigs)*float64(smalls) - float64(smallPlays))
	low := math.Max(1, math.Floor(ratio*0.75))
	high := math.Ceil(ratio * 1.25)
	return int(math.Max(low, math.Min(high, want)))
}

// fmDraw takes the n least heard albums, ties at random.
func fmDraw(from []fmCandidate, n int, rnd *rand.Rand) []int64 {
	shuffled := slices.Clone(from)
	rnd.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	slices.SortStableFunc(shuffled, func(a, b fmCandidate) int { return a.plays - b.plays })
	ids := make([]int64, 0, n)
	for _, c := range shuffled[:min(n, len(shuffled))] {
		ids = append(ids, c.id)
	}
	return ids
}

// fmPlanCycle draws the next cycle in playing order.
func fmPlanCycle(candidates []fmCandidate, rnd *rand.Rand) []int64 {
	var bigs []fmCandidate
	var groups []int64
	members := map[int64][]fmCandidate{}
	playSum := map[int64]int{}
	bigPlays, smalls, smallPlays := 0, 0, 0
	for _, c := range candidates {
		if c.group == 0 {
			bigPlays += c.plays
			bigs = append(bigs, c)
			continue
		}
		if !slices.Contains(groups, c.group) {
			groups = append(groups, c.group)
		}
		members[c.group] = append(members[c.group], c)
		playSum[c.group] += c.plays
		smalls++
		smallPlays += c.plays
	}

	var cycle []int64
	big := fmDraw(slices.DeleteFunc(slices.Clone(bigs), func(c fmCandidate) bool { return c.blocked }), 1, rnd)
	cycle = append(cycle, big...)
	size := fmSegmentSize(bigPlays+len(big), len(bigs), smallPlays, smalls)

	// Each place goes to the group furthest behind on average, which comes
	// to a share in proportion to the groups' sizes and corrects itself.
	available := map[int64][]fmCandidate{}
	for _, g := range groups {
		available[g] = slices.DeleteFunc(slices.Clone(members[g]), func(c fmCandidate) bool { return c.blocked })
	}
	share := map[int64]int{}
	for range size {
		rnd.Shuffle(len(groups), func(i, j int) { groups[i], groups[j] = groups[j], groups[i] })
		var pick int64
		best := math.Inf(1)
		for _, g := range groups {
			if share[g] >= len(available[g]) {
				continue
			}
			if behind := float64(playSum[g]+share[g]) / float64(len(members[g])); behind < best {
				pick, best = g, behind
			}
		}
		if pick == 0 {
			break
		}
		share[pick]++
	}
	rnd.Shuffle(len(groups), func(i, j int) { groups[i], groups[j] = groups[j], groups[i] })
	for _, g := range groups {
		cycle = append(cycle, fmDraw(available[g], share[g], rnd)...)
	}
	return cycle
}

func queryIDs(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func fmGroupIDs(ctx context.Context, tx *sql.Tx, userID int64) ([]int64, error) {
	return queryIDs(ctx, tx, `SELECT playlist_id FROM fm_groups WHERE user_id = ? ORDER BY position`, userID)
}

// fmAlbumGroups assigns each album the group holding most of its tracks; a
// tie goes to the group listed first.
func fmAlbumGroups(ctx context.Context, tx *sql.Tx, userID int64, groups []int64) (map[int64]int64, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT t.album_id, pt.playlist_id, COUNT(*)
		FROM fm_groups g
		JOIN playlist_tracks pt ON pt.playlist_id = g.playlist_id
		JOIN tracks t ON t.id = pt.track_id
		WHERE g.user_id = ? AND t.album_id IS NOT NULL
		GROUP BY t.album_id, pt.playlist_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rank := func(group int64) int { return slices.Index(groups, group) }
	assigned := map[int64]int64{}
	held := map[int64]int{}
	for rows.Next() {
		var album, group int64
		var count int
		if err := rows.Scan(&album, &group, &count); err != nil {
			return nil, err
		}
		if count > held[album] || (count == held[album] && rank(group) < rank(assigned[album])) {
			assigned[album], held[album] = group, count
		}
	}
	return assigned, rows.Err()
}

// fmCandidates loads the albums that can be drawn, giving those seen for the
// first time their row. A new album starts at the lowest count of its pool:
// from zero it would keep coming up until it caught up with the others. The
// row has to exist, so that "never heard yet" and "just joined" stay apart.
func fmCandidates(ctx context.Context, tx *sql.Tx, userID int64) ([]fmCandidate, error) {
	groups, err := fmGroupIDs(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	assigned, err := fmAlbumGroups(ctx, tx, userID, groups)
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT a.album_id, a.tracks, f.plays, IFNULL(f.skipped, 0),
			EXISTS (SELECT 1 FROM fm_plan p WHERE p.user_id = ? AND p.album_id = a.album_id)
		FROM (SELECT album_id, COUNT(*) AS tracks FROM tracks WHERE album_id IS NOT NULL GROUP BY album_id) a
		LEFT JOIN fm_albums f ON f.album_id = a.album_id AND f.user_id = ?
		ORDER BY a.album_id`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var candidates []fmCandidate
	var unseen []int // indexes into candidates
	for rows.Next() {
		var c fmCandidate
		var tracks int
		var plays sql.NullInt64
		var skipped, planned bool
		if err := rows.Scan(&c.id, &tracks, &plays, &skipped, &planned); err != nil {
			return nil, err
		}
		if tracks < fmBigAlbumTracks {
			if c.group = assigned[c.id]; c.group == 0 {
				continue // a small album outside the groups waits to be sorted
			}
		}
		c.blocked = skipped || planned
		c.plays = int(plays.Int64)
		if !plays.Valid {
			unseen = append(unseen, len(candidates))
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if len(unseen) == 0 {
		return candidates, nil
	}

	// The lowest count among the known albums of the same pool; failing
	// that, any album the listener has a count for, or 0 at the very start.
	var anyLowest sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MIN(plays) FROM fm_albums WHERE user_id = ?`, userID).Scan(&anyLowest); err != nil {
		return nil, err
	}
	isBig := func(c fmCandidate) bool { return c.group == 0 }
	lowest := map[bool]int{}
	known := map[bool]bool{}
	isUnseen := make(map[int]bool, len(unseen))
	for _, i := range unseen {
		isUnseen[i] = true
	}
	for i, c := range candidates {
		if isUnseen[i] {
			continue
		}
		if pool := isBig(c); !known[pool] || c.plays < lowest[pool] {
			lowest[pool], known[pool] = c.plays, true
		}
	}
	for _, i := range unseen {
		c := &candidates[i]
		c.plays = int(anyLowest.Int64)
		if pool := isBig(*c); known[pool] {
			c.plays = lowest[pool]
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO fm_albums (user_id, album_id, plays) VALUES (?, ?, ?)`,
			userID, c.id, c.plays); err != nil {
			return nil, err
		}
	}
	return candidates, nil
}

// fmFillPlan drops albums left without tracks and draws until the plan
// holds ahead albums, or nothing more can be drawn.
func fmFillPlan(ctx context.Context, tx *sql.Tx, userID int64, ahead int) error {
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM fm_plan WHERE user_id = ? AND album_id NOT IN
			(SELECT album_id FROM tracks WHERE album_id IS NOT NULL)`, userID); err != nil {
		return err
	}
	rnd := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	// Without groups a cycle is a single big album, so asking further ahead
	// takes as many more draws.
	for range fmMaxDraws + ahead {
		var planned, last int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*), IFNULL(MAX(position), 0) FROM fm_plan WHERE user_id = ?`,
			userID).Scan(&planned, &last); err != nil {
			return err
		}
		if planned >= ahead {
			return nil
		}
		candidates, err := fmCandidates(ctx, tx, userID)
		if err != nil {
			return err
		}
		cycle := fmPlanCycle(candidates, rnd)
		// A skip only sits out the draw right after it.
		if _, err := tx.ExecContext(ctx, `UPDATE fm_albums SET skipped = 0 WHERE user_id = ? AND skipped = 1`, userID); err != nil {
			return err
		}
		if len(cycle) == 0 {
			if planned > 0 || !slices.ContainsFunc(candidates, func(c fmCandidate) bool { return c.blocked }) {
				return nil
			}
			continue // everything was just skipped: draw again, now they are free
		}
		for i, album := range cycle {
			if _, err := tx.ExecContext(ctx, `INSERT INTO fm_plan (user_id, position, album_id) VALUES (?, ?, ?)`,
				userID, last+i+1, album); err != nil {
				return err
			}
		}
	}
	return nil
}

func fmReadState(ctx context.Context, tx *sql.Tx, userID int64, ahead int) (FMState, error) {
	state := FMState{Albums: []FMAlbum{}}
	var err error
	if state.GroupIDs, err = fmGroupIDs(ctx, tx, userID); err != nil {
		return state, err
	}
	albums, err := queryIDs(ctx, tx, `SELECT album_id FROM fm_plan WHERE user_id = ? ORDER BY position LIMIT ?`, userID, ahead)
	if err != nil {
		return state, err
	}
	for _, id := range albums {
		tracks, err := fmAlbumTracks(ctx, tx, id)
		if err != nil {
			return state, err
		}
		state.Albums = append(state.Albums, FMAlbum{ID: id, Tracks: tracks})
	}
	if len(albums) == 0 {
		return state, nil
	}
	// A track moved to another album since no longer says where to resume.
	var track sql.NullInt64
	err = tx.QueryRowContext(ctx, `
		SELECT p.track_id FROM fm_plan p JOIN tracks t ON t.id = p.track_id AND t.album_id = p.album_id
		WHERE p.user_id = ? AND p.album_id = ? ORDER BY p.position LIMIT 1`, userID, albums[0]).Scan(&track)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return state, err
	}
	if track.Valid {
		state.TrackID = &track.Int64
	}
	return state, nil
}

func fmAlbumTracks(ctx context.Context, tx *sql.Tx, albumID int64) ([]Track, error) {
	rows, err := tx.QueryContext(ctx, `SELECT `+trackColumns+trackFrom+
		` WHERE t.album_id = ? ORDER BY t.disc_no, t.track_no, t.id`, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tracks := []Track{}
	for rows.Next() {
		t, err := scanTrack(rows)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

func (s *Store) fmUpdate(ctx context.Context, userID int64, change func(*sql.Tx) error) (FMState, error) {
	return s.fmUpdateAhead(ctx, userID, fmLookahead, change)
}

func (s *Store) fmUpdateAhead(ctx context.Context, userID int64, ahead int, change func(*sql.Tx) error) (FMState, error) {
	var state FMState
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := change(tx); err != nil {
			return err
		}
		if err := fmFillPlan(ctx, tx, userID, ahead); err != nil {
			return err
		}
		var err error
		state, err = fmReadState(ctx, tx, userID, ahead)
		return err
	})
	return state, err
}

// FM returns the listener's radio, lining up albums if needed.
func (s *Store) FM(ctx context.Context, userID int64) (FMState, error) {
	return s.FMAhead(ctx, userID, fmLookahead)
}

// FMAhead is FM with ahead albums, the current one included, instead of the
// player's two: the Android home screen shows the albums coming up. They are
// lined up for real, so they are what the radio goes on to play. ahead is
// clamped to 1..FMMaxAhead.
func (s *Store) FMAhead(ctx context.Context, userID int64, ahead int) (FMState, error) {
	ahead = max(1, min(ahead, FMMaxAhead))
	return s.fmUpdateAhead(ctx, userID, ahead, func(*sql.Tx) error { return nil })
}

// fmLeave takes the current album off the plan. It reports false for any
// other album: another tab or device got there first, or the request is
// repeated. A small library can line the same album up again right away, so
// looking further down the plan would count it twice.
func fmLeave(ctx context.Context, tx *sql.Tx, userID, albumID int64) (bool, error) {
	res, err := tx.ExecContext(ctx, `
		DELETE FROM fm_plan WHERE user_id = ? AND album_id = ? AND position =
			(SELECT MIN(position) FROM fm_plan WHERE user_id = ?)`, userID, albumID, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// fmPassOver treats the albums planned before position as skipped.
func fmPassOver(ctx context.Context, tx *sql.Tx, userID int64, position int) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE fm_albums SET skipped = 1 WHERE user_id = ? AND album_id IN
			(SELECT album_id FROM fm_plan WHERE user_id = ? AND position < ?)`, userID, userID, position); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM fm_plan WHERE user_id = ? AND position < ?`, userID, position)
	return err
}

// FMFinish counts an album heard to its end and moves on. Repeating the
// request changes nothing.
func (s *Store) FMFinish(ctx context.Context, userID, albumID int64) (FMState, error) {
	return s.fmUpdate(ctx, userID, func(tx *sql.Tx) error {
		left, err := fmLeave(ctx, tx, userID, albumID)
		if err != nil || !left {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			UPDATE fm_albums SET plays = plays + 1, skipped = 0 WHERE user_id = ? AND album_id = ?`, userID, albumID)
		return err
	})
}

// FMSkip moves on from an album without counting it and keeps it out of the
// next draw. Next time it comes up it plays from its start.
func (s *Store) FMSkip(ctx context.Context, userID, albumID int64) (FMState, error) {
	return s.fmUpdate(ctx, userID, func(tx *sql.Tx) error {
		left, err := fmLeave(ctx, tx, userID, albumID)
		if err != nil || !left {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE fm_albums SET skipped = 1 WHERE user_id = ? AND album_id = ?`, userID, albumID)
		return err
	})
}

// FMSetPosition records the track being played, for resuming after the
// listener played something else. A track of a later planned album means
// the listener jumped ahead; a track of no planned album changes nothing.
func (s *Store) FMSetPosition(ctx context.Context, userID, trackID int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		var position int
		err := tx.QueryRowContext(ctx, `
			SELECT p.position FROM fm_plan p JOIN tracks t ON t.album_id = p.album_id
			WHERE p.user_id = ? AND t.id = ? ORDER BY p.position LIMIT 1`, userID, trackID).Scan(&position)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := fmPassOver(ctx, tx, userID, position); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE fm_plan SET track_id = ? WHERE user_id = ? AND position = ?`, trackID, userID, position)
		return err
	})
}

// FMSetGroups replaces the playlists used as groups, in priority order. The
// current album stays; the albums lined up after it are drawn again.
func (s *Store) FMSetGroups(ctx context.Context, userID int64, playlistIDs []int64) (FMState, error) {
	return s.fmUpdate(ctx, userID, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM fm_groups WHERE user_id = ?`, userID); err != nil {
			return err
		}
		seen := map[int64]bool{}
		for _, id := range playlistIDs {
			if seen[id] {
				continue
			}
			seen[id] = true
			if _, err := s.playlist(ctx, tx, userID, id); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO fm_groups (user_id, playlist_id, position) VALUES (?, ?, ?)`,
				userID, id, len(seen)); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx, `
			DELETE FROM fm_plan WHERE user_id = ? AND position >
				(SELECT MIN(position) FROM fm_plan WHERE user_id = ?)`, userID, userID)
		return err
	})
}
