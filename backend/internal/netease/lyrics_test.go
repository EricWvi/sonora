package netease

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestComposePlainLRCStaysAsItIs(t *testing.T) {
	raw := "{\"t\":0,\"c\":[]}\n[00:01.00]line"
	if text, format := (LyricSet{LRC: raw}).Compose(); text != raw || format != "lrc" {
		t.Fatalf("%q %q", text, format)
	}
	if text, format := (LyricSet{}).Compose(); text != "" || format != "" {
		t.Fatalf("empty: %q %q", text, format)
	}
}

func TestComposeMergesTranslationIntoLRC(t *testing.T) {
	text, format := LyricSet{
		LRC:         "[by:x]\n[00:01.00]one\n[00:02.50]\n[00:03.50]two\n[00:05.00]three",
		Translation: "[by:y]\n[00:01.10]一\n[00:03.50]//\n[00:05.00]三",
		Roman:       "[00:05.00]san",
	}.Compose()
	want := "[00:01.000]one\n[00:01.000]一\n[00:02.500]\n[00:03.500]two\n[00:05.000]three\n[00:05.000]三\n[00:05.000]san"
	if format != "lrc" || text != want {
		t.Fatalf("%q\nwant %q", text, want)
	}
}

func TestComposeTurnsWordTimingsIntoTTML(t *testing.T) {
	text, format := LyricSet{
		LRC: "[00:01.00]ignored",
		YRC: "{\"t\":0,\"c\":[]}\n" +
			"[1000,2000](1000,500,0)Hi (1500,500,0)(you) (2000,1000,0)<&>\n" +
			"[70000,1000](70000,1000,0)再见",
		YTranslated: "[00:01.000]你好\n[01:10.000]//",
		YRoman:      "[00:01.100]roman",
		Translation: "[00:01.00]not this one",
	}.Compose()
	if format != "ttml" {
		t.Fatalf("format %q", format)
	}
	for _, want := range []string{
		`<p begin="00:00:01.000" end="00:00:03.000" itunes:key="L1" ttm:agent="v1">`,
		`<span begin="00:00:01.000" end="00:00:01.500">Hi</span> <span begin="00:00:01.500" end="00:00:02.000">(you)</span> `,
		`&lt;&amp;&gt;</span>`,
		`<span ttm:role="x-translation" xml:lang="zh-CN">你好</span>`,
		`<span ttm:role="x-roman">roman</span>`,
		`<p begin="00:01:10.000" end="00:01:11.000" itunes:key="L2" ttm:agent="v1"><span begin="00:01:10.000" end="00:01:11.000">再见</span></p>`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %s in\n%s", want, text)
		}
	}
	d := xml.NewDecoder(strings.NewReader(text))
	for {
		if _, err := d.Token(); err != nil {
			if err.Error() != "EOF" {
				t.Fatalf("not well formed: %v", err)
			}
			break
		}
	}
}

func TestComposeFallsBackOnLineTimedTranslation(t *testing.T) {
	text, _ := LyricSet{
		YRC:         "[1000,1000](1000,1000,0)one",
		Translation: "[00:01.20]一",
	}.Compose()
	if !strings.Contains(text, ">一</span>") {
		t.Fatal(text)
	}
}

func TestQRCToTTML(t *testing.T) {
	text, ok := QRCToTTML("[ti:17岁]\n[ar:刘德华]\n[offset:0]\n" +
		"[266,1551]17(266,311)岁(577,148) - (726,148)刘(874,287)\r\n" +
		"[5545,10751]Billy (6081,312)Chan(6393,9903)\n" +
		"[6000,2000](ooh (6000,1000)yeah)(7000,1000)\n" +
		"[19000,1000]好(19000,1000)\n" +
		"[20000,1000]（啊）(20000,1000)")
	if !ok {
		t.Fatal("no QRC lines")
	}
	for _, want := range []string{
		`<p begin="00:00:00.266" end="00:00:01.817" itunes:key="L1" ttm:agent="v1"><span begin="00:00:00.266" end="00:00:00.577">17</span><span begin="00:00:00.577" end="00:00:00.725">岁</span> <span begin="00:00:00.726" end="00:00:00.874">-</span> <span begin="00:00:00.874" end="00:00:01.161">刘</span></p>`,
		`<span begin="00:00:06.081" end="00:00:06.393">Billy</span> <span begin="00:00:06.393" end="00:00:16.296">Chan</span>`,
		`<span ttm:role="x-bg" begin="00:00:06.000" end="00:00:08.000"><span begin="00:00:06.000" end="00:00:07.000">ooh</span> <span begin="00:00:07.000" end="00:00:08.000">yeah</span></span></p>`,
		`>好</span><span ttm:role="x-bg" begin="00:00:20.000" end="00:00:21.000"><span begin="00:00:20.000" end="00:00:21.000">啊</span></span></p>`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %s in\n%s", want, text)
		}
	}
	if strings.Contains(text, "ti:") || strings.Contains(text, `key="L4"`) {
		t.Errorf("tags or extra lines in\n%s", text)
	}
	if _, ok := QRCToTTML("[1000,2000](1000,500,0)YRC 不是 QRC"); ok {
		t.Error("YRC read as QRC")
	}
}
