package server

import (
	"net/http"
	"strings"
	"testing"

	"sonora/internal/store"
)

func TestLibraryChanges(t *testing.T) {
	ts := newTest(t)
	ts.setup()
	tr := decodeBody[store.IngestResult](t, ts.upload("a.flac", sample(t, "tagged.flac"), "")).Track
	adminCookie := ts.cookie
	ts.cookie = ""
	ts.signIn("alice", "Alice")
	listenerCookie := ts.cookie

	rec := ts.do("GET", playerAPI+"library/changes", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("changes: %d %s", rec.Code, rec.Body)
	}
	for _, field := range []string{"originalName", "size", "createdAt", "neteaseId"} {
		if strings.Contains(rec.Body.String(), `"`+field) {
			t.Fatalf("changes expose %s: %s", field, rec.Body)
		}
	}
	c := decodeBody[libraryChanges](t, rec)
	if c.More || c.Reset || len(c.Tracks) != 1 || len(c.Albums) != 1 || len(c.Artists) != 1 || c.Cursor == 0 {
		t.Fatalf("changes = %+v", c)
	}
	got := c.Tracks[0]
	if got.ID != tr.ID || got.SortKey == "" || got.Cover == nil || got.Album.ID != tr.Album.ID ||
		c.Albums[0].SortKey == "" || c.Albums[0].Single || c.Artists[0].SortKey == "" {
		t.Fatalf("items = %+v / %+v / %+v", got, c.Albums[0], c.Artists[0])
	}

	// Deleting the track deletes its album and artist too.
	ts.cookie = adminCookie
	if rec := ts.do("DELETE", api+"tracks/"+itoa(tr.ID), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body)
	}
	ts.cookie = listenerCookie
	rec = ts.do("GET", playerAPI+"library/changes?since="+itoa(c.Cursor), "")
	after := decodeBody[libraryChanges](t, rec)
	if len(after.Tracks) != 0 || len(after.Deleted.Tracks) != 1 || len(after.Deleted.Albums) != 1 ||
		len(after.Deleted.Artists) != 1 || after.Cursor <= c.Cursor {
		t.Fatalf("after delete = %s", rec.Body)
	}

	// Ahead of the log: start over.
	rec = ts.do("GET", playerAPI+"library/changes?since="+itoa(after.Cursor+100), "")
	if r := decodeBody[libraryChanges](t, rec); !r.Reset {
		t.Fatalf("ahead = %s", rec.Body)
	}
	for _, bad := range []string{"-1", "x"} {
		if rec := ts.do("GET", playerAPI+"library/changes?since="+bad, ""); rec.Code != http.StatusBadRequest {
			t.Fatalf("since=%s: %d", bad, rec.Code)
		}
	}
}
