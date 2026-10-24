// Package store persists admin accounts and the music library.
//
// Metadata lives in SQLite inside data_dir; cover images live in
// <data_dir>/covers, named by content hash. Uploaded audio lives in the
// media volume under a layout only this package knows: file paths never
// leave it, callers address media by stable IDs.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite" // CGO-free SQLite driver
)

var (
	// ErrNotFound reports a missing record.
	ErrNotFound = errors.New("not found")
	// ErrInvalid reports rejected input; the wrapped message is user-facing.
	ErrInvalid = errors.New("invalid input")
	// ErrConflict reports an operation that the current state forbids; the
	// wrapped message is user-facing.
	ErrConflict = errors.New("conflict")
)

// Store is safe for concurrent use.
type Store struct {
	db       *sql.DB
	coverDir string
	mediaDir string
	// incoming holds uploads being received. It sits inside the media
	// volume so accepting a file is a rename, not a copy.
	incoming string
	// files serialises every step that writes or removes files together
	// with the rows that reference them, so garbage collection never
	// removes a file a concurrent request is about to reference.
	files sync.Mutex
	// decoding bounds concurrent thumbnail decodes, which hold whole
	// images in memory.
	decoding chan struct{}
}

// Open opens (and migrates) the database inside dataDir and prepares the
// media volume.
func Open(dataDir, mediaDir string) (*Store, error) {
	coverDir := filepath.Join(dataDir, "covers")
	incoming := filepath.Join(mediaDir, ".incoming")
	for _, dir := range []string{coverDir, filepath.Join(coverDir, "thumbs"), mediaDir, incoming} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create data dir: %w", err)
		}
	}
	// Uploads interrupted by a restart are never resumed.
	if leftovers, err := filepath.Glob(filepath.Join(incoming, "*")); err == nil {
		for _, f := range leftovers {
			os.Remove(f)
		}
	}
	dsn := "file:" + filepath.Join(dataDir, "sonora.db") +
		"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	// A single connection serialises writes and keeps transactions simple.
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	s := &Store{db: db, coverDir: coverDir, mediaDir: mediaDir, incoming: incoming, decoding: make(chan struct{}, 2)}
	if err := s.backfillPlaceholders(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("cover placeholders: %w", err)
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) inTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func invalid(msg string) error  { return fmt.Errorf("%w: %s", ErrInvalid, msg) }
func conflict(msg string) error { return fmt.Errorf("%w: %s", ErrConflict, msg) }
