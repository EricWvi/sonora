package store

import (
	"context"
	"errors"
	"testing"
)

const sampleTTML = `<?xml version="1.0" encoding="UTF-8"?>
<tt xmlns="http://www.w3.org/ns/ttml" xmlns:ttm="http://www.w3.org/ns/ttml#metadata">
<body><div><p begin="00:01.000" end="00:03.000"><span begin="00:01.000" end="00:02.000">刮风</span><span begin="00:02.000" end="00:03.000">这天</span></p></div></body>
</tt>`

func TestLyricsFormat(t *testing.T) {
	for _, c := range []struct {
		text, format string
	}{
		{"", LRC},
		{"[00:01.00]晴天", LRC},
		{"[00:01.00]<00:01.00>刮<00:01.50>风", LRC}, // enhanced LRC
		{"只有文字", LRC},
		{sampleTTML, TTML},
		{`<tt xmlns="http://www.w3.org/ns/ttml"/>`, TTML},
	} {
		if got, err := lyricsFormat(c.text); err != nil || got != c.format {
			t.Errorf("lyricsFormat(%.20q) = %q, %v; want %q", c.text, got, err, c.format)
		}
	}
	for _, text := range []string{
		`<tt><body><p>未闭合</body></tt>`,
		`<html><body/></html>`,
		`<?xml version="1.0"?>`,
	} {
		if _, err := lyricsFormat(text); !errors.Is(err, ErrInvalid) {
			t.Errorf("lyricsFormat(%q) = %v, want ErrInvalid", text, err)
		}
	}
}

func TestLyricsKeepTheirFormat(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	tr := ingest(t, s, sample(t, "tagged.flac"), "a.flac", nil).Track
	if tr.LyricsFormat != LRC {
		t.Fatalf("embedded lyrics format = %q", tr.LyricsFormat)
	}
	d, err := s.SetTrackLyrics(ctx, tr.ID, []byte("\xEF\xBB\xBF"+sampleTTML))
	if err != nil || d.LyricsFormat != TTML || d.Lyrics != sampleTTML {
		t.Fatalf("ttml: %q %v", d.LyricsFormat, err)
	}
	if _, err := s.SetTrackLyrics(ctx, tr.ID, []byte("<tt><p>")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("broken ttml: %v", err)
	}
	if d, _ = s.Track(ctx, tr.ID); d.LyricsFormat != TTML {
		t.Fatalf("rejected upload changed lyrics: %q", d.LyricsFormat)
	}
	if d, err = s.SetTrackLyrics(ctx, tr.ID, nil); err != nil || d.LyricsFormat != "" || d.HasLyrics {
		t.Fatalf("cleared: %+v %v", d.Track, err)
	}

	// A sidecar file decides the format at upload.
	res := ingest(t, s, sample(t, "tagged.opus"), "b.opus", []byte(sampleTTML))
	if res.Track.LyricsFormat != TTML {
		t.Fatalf("sidecar ttml format = %q", res.Track.LyricsFormat)
	}
	page, _ := s.Tracks(ctx, TrackQuery{})
	for _, it := range page.Items {
		if it.ID == res.Track.ID && it.LyricsFormat != TTML {
			t.Fatalf("list format = %q", it.LyricsFormat)
		}
	}
}

func TestYRCUploadBecomesTTML(t *testing.T) {
	yrc := "{\"t\":0,\"c\":[{\"tx\":\"作词: 周杰伦\"}]}\n[1000,2000](1000,1000,0)刮风(2000,1000,0)这天"
	text, format, err := readLyrics([]byte(yrc))
	if err != nil || format != TTML {
		t.Fatalf("readLyrics(yrc) = %q, %v; want TTML", format, err)
	}
	if got, err := lyricsFormat(text); err != nil || got != TTML {
		t.Errorf("converted YRC is not valid TTML: %q, %v", got, err)
	}
	if _, format, _ := readLyrics([]byte("[00:01.00]晴天")); format != LRC {
		t.Errorf("LRC read as %q", format)
	}
}
