package store

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

// schemaAt creates sonora.db in dataDir with only the first n migrations.
func schemaAt(t *testing.T, dataDir string, n int) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(dataDir, "sonora.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatal(err)
	}
	for i := range n {
		if err := applyMigration(db, i); err != nil {
			t.Fatalf("migration %d: %v", i+1, err)
		}
	}
}

func userVersion(t *testing.T, dataDir string) int {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(dataDir, "sonora.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var v int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestMigrateFromEveryVersion(t *testing.T) {
	for from := 1; from < len(migrations); from++ {
		dataDir := t.TempDir()
		schemaAt(t, dataDir, from)
		s, err := Open(dataDir, t.TempDir())
		if err != nil {
			t.Fatalf("from v%d: %v", from, err)
		}
		s.Close()
		if v := userVersion(t, dataDir); v != len(migrations) {
			t.Fatalf("from v%d: version = %d", from, v)
		}
	}
}

func TestFailedMigrationRollsBack(t *testing.T) {
	dataDir := t.TempDir()
	schemaAt(t, dataDir, len(migrations))
	saved := migrations
	t.Cleanup(func() { migrations = saved })
	migrations = append(saved[:len(saved):len(saved)],
		execSQL(`CREATE TABLE half_done (a); INSERT INTO missing VALUES (1)`))

	if _, err := Open(dataDir, t.TempDir()); err == nil {
		t.Fatal("broken migration succeeded")
	}
	if v := userVersion(t, dataDir); v != len(saved) {
		t.Fatalf("version = %d", v)
	}
	db, err := sql.Open("sqlite", "file:"+filepath.Join(dataDir, "sonora.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name = 'half_done'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("half_done tables = %d, %v", n, err)
	}
}

func TestOpenRefusesNewerSchema(t *testing.T) {
	dataDir := t.TempDir()
	schemaAt(t, dataDir, len(migrations))
	db, err := sql.Open("sqlite", "file:"+filepath.Join(dataDir, "sonora.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`PRAGMA user_version = 99`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	_, err = Open(dataDir, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "newer than this release") {
		t.Fatalf("err = %v", err)
	}
	if v := userVersion(t, dataDir); v != 99 {
		t.Fatalf("version = %d", v)
	}
}

func TestMigrationMarksOneTrackAlbumsSingle(t *testing.T) {
	dataDir := t.TempDir()
	schemaAt(t, dataDir, 6)
	db, err := sql.Open("sqlite", "file:"+filepath.Join(dataDir, "sonora.db"))
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`INSERT INTO albums (id, title) VALUES (1, 'one'), (2, 'two')`,
		`INSERT INTO tracks (title, album_id, duration_ms, format, size, sha256, file, original_name) VALUES
			('a', 1, 1, 'mp3', 1, 'h1', 'f1', 'a.mp3'),
			('b', 2, 1, 'mp3', 1, 'h2', 'f2', 'b.mp3'),
			('c', 2, 1, 'mp3', 1, 'h3', 'f3', 'c.mp3')`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	db.Close()
	s, err := Open(dataDir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for id, want := range map[int64]bool{1: true, 2: false} {
		a, err := s.Album(t.Context(), id)
		if err != nil || a.Single != want {
			t.Errorf("album %d single = %v, %v; want %v", id, a.Single, err, want)
		}
		// migration 8 leaves existing albums without a description
		if a.Description != "" {
			t.Errorf("album %d description = %q", id, a.Description)
		}
	}
}
