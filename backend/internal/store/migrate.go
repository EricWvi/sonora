package store

import (
	"database/sql"
	"errors"
	"fmt"
)

// A migration runs inside its own transaction.
type migration func(tx *sql.Tx) error

func execSQL(query string) migration {
	return func(tx *sql.Tx) error {
		_, err := tx.Exec(query)
		return err
	}
}

// migrations are applied in order; PRAGMA user_version records progress.
// Never edit a released migration, append a new one instead.

var migrations = []migration{
	execSQL(`
	CREATE TABLE admins (
		id            INTEGER PRIMARY KEY,
		username      TEXT    NOT NULL UNIQUE COLLATE NOCASE,
		password_hash TEXT    NOT NULL,
		created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);

	CREATE TABLE admin_sessions (
		token_hash   TEXT    PRIMARY KEY,
		admin_id     INTEGER NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
		created_at   INTEGER NOT NULL,
		last_seen_at INTEGER NOT NULL,
		expires_at   INTEGER NOT NULL
	);
	CREATE INDEX admin_sessions_admin ON admin_sessions (admin_id);

	-- AUTOINCREMENT keeps IDs of deleted media from being reused, so a
	-- stale reference can never point at different music.
	CREATE TABLE artists (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT    NOT NULL UNIQUE COLLATE NOCASE,
		created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);

	-- artist_id is the album artist; NULL when no tag names one.
	CREATE TABLE albums (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		title      TEXT    NOT NULL,
		artist_id  INTEGER REFERENCES artists (id),
		year       INTEGER NOT NULL DEFAULT 0,
		cover      TEXT,
		created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);
	CREATE UNIQUE INDEX albums_title_artist ON albums (title COLLATE NOCASE, IFNULL(artist_id, 0));
	CREATE INDEX albums_artist ON albums (artist_id);

	-- file is relative to the media volume and never leaves the store.
	-- cover is only set when the track's art differs from its album's.
	CREATE TABLE tracks (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		title         TEXT    NOT NULL,
		artist_id     INTEGER REFERENCES artists (id),
		album_id      INTEGER REFERENCES albums (id),
		track_no      INTEGER NOT NULL DEFAULT 0,
		disc_no       INTEGER NOT NULL DEFAULT 0,
		year          INTEGER NOT NULL DEFAULT 0,
		duration_ms   INTEGER NOT NULL,
		format        TEXT    NOT NULL,
		size          INTEGER NOT NULL,
		sha256        TEXT    NOT NULL UNIQUE,
		file          TEXT    NOT NULL,
		original_name TEXT    NOT NULL,
		cover         TEXT,
		lyrics        TEXT    NOT NULL DEFAULT '',
		created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);
	CREATE INDEX tracks_artist ON tracks (artist_id);
	CREATE INDEX tracks_album ON tracks (album_id, disc_no, track_no);
	`),

	// Tracks and albums can have several artists. An album is identified by
	// its title and the set of its artists, kept in artist_key as sorted
	// artist IDs joined by commas ('' for none) so it can be indexed.
	// sqlite_sequence is carried over so deleted IDs stay retired.
	execSQL(`
	CREATE TABLE new_albums (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		title      TEXT    NOT NULL,
		artist_key TEXT    NOT NULL DEFAULT '',
		year       INTEGER NOT NULL DEFAULT 0,
		cover      TEXT,
		created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);
	INSERT INTO new_albums (id, title, artist_key, year, cover, created_at)
		SELECT id, title, IFNULL(artist_id, ''), year, cover, created_at FROM albums;

	CREATE TABLE new_tracks (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		title         TEXT    NOT NULL,
		album_id      INTEGER REFERENCES albums (id),
		track_no      INTEGER NOT NULL DEFAULT 0,
		disc_no       INTEGER NOT NULL DEFAULT 0,
		year          INTEGER NOT NULL DEFAULT 0,
		duration_ms   INTEGER NOT NULL,
		format        TEXT    NOT NULL,
		size          INTEGER NOT NULL,
		sha256        TEXT    NOT NULL UNIQUE,
		file          TEXT    NOT NULL,
		original_name TEXT    NOT NULL,
		cover         TEXT,
		lyrics        TEXT    NOT NULL DEFAULT '',
		created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);
	INSERT INTO new_tracks (id, title, album_id, track_no, disc_no, year, duration_ms, format, size,
			sha256, file, original_name, cover, lyrics, created_at, updated_at)
		SELECT id, title, album_id, track_no, disc_no, year, duration_ms, format, size,
			sha256, file, original_name, cover, lyrics, created_at, updated_at FROM tracks;

	-- position orders the artists as the tags or the admin listed them.
	CREATE TABLE track_artists (
		track_id  INTEGER NOT NULL REFERENCES tracks (id) ON DELETE CASCADE,
		artist_id INTEGER NOT NULL REFERENCES artists (id),
		position  INTEGER NOT NULL,
		PRIMARY KEY (track_id, artist_id)
	) WITHOUT ROWID;
	INSERT INTO track_artists SELECT id, artist_id, 0 FROM tracks WHERE artist_id IS NOT NULL;

	CREATE TABLE album_artists (
		album_id  INTEGER NOT NULL REFERENCES albums (id) ON DELETE CASCADE,
		artist_id INTEGER NOT NULL REFERENCES artists (id),
		position  INTEGER NOT NULL,
		PRIMARY KEY (album_id, artist_id)
	) WITHOUT ROWID;
	INSERT INTO album_artists SELECT id, artist_id, 0 FROM albums WHERE artist_id IS NOT NULL;

	DELETE FROM sqlite_sequence WHERE name IN ('new_albums', 'new_tracks');
	INSERT INTO sqlite_sequence (name, seq)
		SELECT 'new_' || name, seq FROM sqlite_sequence WHERE name IN ('albums', 'tracks');

	DROP TABLE tracks;
	DROP TABLE albums;
	ALTER TABLE new_albums RENAME TO albums;
	ALTER TABLE new_tracks RENAME TO tracks;

	CREATE UNIQUE INDEX albums_title_artists ON albums (title COLLATE NOCASE, artist_key);
	CREATE INDEX tracks_album ON tracks (album_id, disc_no, track_no);
	CREATE INDEX track_artists_artist ON track_artists (artist_id);
	CREATE INDEX album_artists_artist ON album_artists (artist_id);
	`),

	// thumbhash is the cover's blurred placeholder, '' when the image
	// could not be decoded. Rows for covers stored before this migration
	// are filled in when the store opens.
	execSQL(`
	CREATE TABLE covers (
		name      TEXT PRIMARY KEY,
		thumbhash TEXT NOT NULL
	) WITHOUT ROWID;
	`),

	// Listeners of the player, signed in through OIDC. Copied from
	// homepage: sessions never expire, a row lives until the user signs
	// out.
	execSQL(`
	CREATE TABLE users (
		id         INTEGER PRIMARY KEY,
		issuer     TEXT NOT NULL,
		subject    TEXT NOT NULL,
		name       TEXT NOT NULL DEFAULT '',
		email      TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		UNIQUE (issuer, subject)
	);

	CREATE TABLE sessions (
		token_hash   TEXT    PRIMARY KEY,
		user_id      INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
		created_at   INTEGER NOT NULL,
		last_seen_at INTEGER NOT NULL
	);
	CREATE INDEX sessions_user ON sessions (user_id);
	`),

	// What each listener keeps: liked tracks, playlists and play history.
	// Rows go with their user or track. id orders likes and history,
	// newest last, since several can share a second.
	execSQL(`
	CREATE TABLE likes (
		id         INTEGER PRIMARY KEY,
		user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
		track_id   INTEGER NOT NULL REFERENCES tracks (id) ON DELETE CASCADE,
		created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		UNIQUE (user_id, track_id)
	);
	CREATE INDEX likes_track ON likes (track_id);

	-- One row per track: playing it again moves it to the top.
	CREATE TABLE play_history (
		id        INTEGER PRIMARY KEY,
		user_id   INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
		track_id  INTEGER NOT NULL REFERENCES tracks (id) ON DELETE CASCADE,
		played_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		UNIQUE (user_id, track_id)
	);
	CREATE INDEX play_history_track ON play_history (track_id);

	CREATE TABLE playlists (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id     INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
		name        TEXT    NOT NULL,
		description TEXT    NOT NULL DEFAULT '',
		created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	);
	CREATE INDEX playlists_user ON playlists (user_id);

	-- A track appears at most once in a playlist; position orders it.
	CREATE TABLE playlist_tracks (
		playlist_id INTEGER NOT NULL REFERENCES playlists (id) ON DELETE CASCADE,
		track_id    INTEGER NOT NULL REFERENCES tracks (id) ON DELETE CASCADE,
		position    INTEGER NOT NULL,
		PRIMARY KEY (playlist_id, track_id)
	) WITHOUT ROWID;
	CREATE INDEX playlist_tracks_order ON playlist_tracks (playlist_id, position);
	CREATE INDEX playlist_tracks_track ON playlist_tracks (track_id);
	`),
}

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version >= len(migrations) {
		return nil
	}
	// Rebuilding a referenced table needs foreign keys off, which only
	// takes effect outside a transaction; each migration checks them
	// before committing instead. The store holds a single connection.
	if _, err := db.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		return err
	}
	defer db.Exec(`PRAGMA foreign_keys = ON`)
	for i := version; i < len(migrations); i++ {
		if err := applyMigration(db, i); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
	}
	return nil
}

func applyMigration(db *sql.DB, i int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := migrations[i](tx); err != nil {
		return err
	}
	rows, err := tx.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		return err
	}
	broken := rows.Next()
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if broken {
		return errors.New("foreign key violation")
	}
	if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
		return err
	}
	return tx.Commit()
}
