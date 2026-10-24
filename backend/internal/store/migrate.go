package store

import (
	"database/sql"
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
//
// The users and sessions tables are left to the player's OIDC sign-in,
// which is copied from homepage in phase 4.
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
}

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	for i := version; i < len(migrations); i++ {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if err := migrations[i](tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
	}
	return nil
}
