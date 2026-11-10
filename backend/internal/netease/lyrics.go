package netease

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// The lyric formats Compose returns, as the player API names them.
const (
	lyricLRC  = "lrc"
	lyricTTML = "ttml"
)

// alignTolerance is how far apart in time a translation may be from its line,
// in milliseconds. SPlayer's alignLyrics uses the same.
const alignTolerance = 300

// LyricSet is what /lyric/new returns for a song.
type LyricSet struct {
	LRC         string
	Translation string // lines timed like LRC
	Roman       string // lines timed like LRC
	YRC         string
	YTranslated string // lines timed like YRC
	YRoman      string // lines timed like YRC
}

// Compose makes one document of the set: TTML when there are word timings,
// since that is the format that keeps them with a translation and romanization
// in one file, otherwise LRC. A translation or romanization NetEase does not
// have stays out; nothing is fetched from elsewhere to fill it in.
func (l LyricSet) Compose() (text, format string) {
	if lines := parseYRC(l.YRC); len(lines) > 0 {
		// Songs with word timings normally have their own timed translation;
		// fall back on the line-timed one, which the tolerance usually matches.
		return renderTTML(lines,
			align(lines, parseLRC(firstNonEmpty(l.YTranslated, l.Translation))),
			align(lines, parseLRC(firstNonEmpty(l.YRoman, l.Roman)))), lyricTTML
	}
	if strings.TrimSpace(l.LRC) == "" {
		return "", ""
	}
	if strings.TrimSpace(l.Translation) == "" && strings.TrimSpace(l.Roman) == "" {
		return l.LRC, lyricLRC
	}
	return mergeLRC(parseLRC(l.LRC), parseLRC(l.Translation), parseLRC(l.Roman)), lyricLRC
}

// YRCToTTML turns an uploaded YRC file into TTML, the format the player reads
// word timings from. ok is false when raw has no YRC lines.
func YRCToTTML(raw string) (ttml string, ok bool) {
	lines := parseYRC(raw)
	if len(lines) == 0 {
		return "", false
	}
	return renderTTML(lines, align(lines, nil), align(lines, nil)), true
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

type yrcWord struct {
	start, duration int
	text            string
}

type yrcLine struct {
	start, duration int
	words           []yrcWord
	bg              *yrcLine // background vocal sung over the line
}

var (
	yrcLineHead = regexp.MustCompile(`^\[(\d+),(\d+)\]`)
	// Only the timing is matched: the words themselves may hold parentheses.
	yrcWordHead = regexp.MustCompile(`\((\d+),(\d+),\d+\)`)
	qrcWordTail = regexp.MustCompile(`\((\d+),(\d+)\)`)
	lrcStamp    = regexp.MustCompile(`\[(\d+):(\d+)(?:[.:](\d+))?\]`)
)

// parseYRC reads NetEase's word-timed lyric, one line of
// "[start,duration](start,duration,0)word(start,duration,0)word". Credit lines
// in JSON have no timing and are left out.
func parseYRC(raw string) []yrcLine {
	var lines []yrcLine
	for _, row := range strings.Split(raw, "\n") {
		row = strings.TrimRight(row, "\r")
		head := yrcLineHead.FindStringSubmatch(row)
		if head == nil {
			continue
		}
		body := row[len(head[0]):]
		line := yrcLine{start: atoi(head[1]), duration: atoi(head[2])}
		marks := yrcWordHead.FindAllStringSubmatchIndex(body, -1)
		for i, m := range marks {
			end := len(body)
			if i+1 < len(marks) {
				end = marks[i+1][0]
			}
			line.words = append(line.words, yrcWord{
				start:    atoi(body[m[2]:m[3]]),
				duration: atoi(body[m[4]:m[5]]),
				text:     body[m[1]:end],
			})
		}
		if len(line.words) > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}

// parseQRC reads QQ Music's word-timed lyric, one line of
// "[start,duration]word(start,duration)word(start,duration)", the timing
// after each word rather than before it as in YRC. Tags such as [ti:] have no
// timing and are left out. A line wrapped in parentheses is a background
// vocal, as AMLL's parseQrc reads it; it goes with the line before it.
func parseQRC(raw string) []yrcLine {
	var lines []yrcLine
	for _, row := range strings.Split(raw, "\n") {
		row = strings.TrimSpace(row)
		head := yrcLineHead.FindStringSubmatch(row)
		if head == nil {
			continue
		}
		body := strings.TrimSpace(row[len(head[0]):])
		line := yrcLine{start: atoi(head[1]), duration: atoi(head[2])}
		from := 0
		for _, m := range qrcWordTail.FindAllStringSubmatchIndex(body, -1) {
			line.words = append(line.words, yrcWord{
				start:    atoi(body[m[2]:m[3]]),
				duration: atoi(body[m[4]:m[5]]),
				text:     body[from:m[0]],
			})
			from = m[1]
		}
		if len(line.words) == 0 {
			continue
		}
		first, last := &line.words[0], &line.words[len(line.words)-1]
		if (strings.HasPrefix(first.text, "(") || strings.HasPrefix(first.text, "（")) &&
			(strings.HasSuffix(last.text, ")") || strings.HasSuffix(last.text, "）")) &&
			len(lines) > 0 && lines[len(lines)-1].bg == nil {
			first.text = strings.TrimPrefix(strings.TrimPrefix(first.text, "("), "（")
			last.text = strings.TrimSuffix(strings.TrimSuffix(last.text, ")"), "）")
			lines[len(lines)-1].bg = &line
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

// QRCToTTML turns an uploaded QRC file into TTML. ok is false when raw has no
// QRC lines.
func QRCToTTML(raw string) (ttml string, ok bool) {
	lines := parseQRC(raw)
	if len(lines) == 0 {
		return "", false
	}
	return renderTTML(lines, align(lines, nil), align(lines, nil)), true
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

type lrcLine struct {
	start int
	text  string
}

// parseLRC reads timed lines, a line with several stamps becoming one per
// stamp. Tags such as [by:name] and JSON credits have no timing and are left
// out. Lines stay in file order.
func parseLRC(raw string) []lrcLine {
	var lines []lrcLine
	for _, row := range strings.Split(raw, "\n") {
		row = strings.TrimRight(row, "\r")
		rest := row
		var stamps []int
		for {
			m := lrcStamp.FindStringSubmatchIndex(rest)
			if m == nil || m[0] != 0 {
				break
			}
			ms := 0
			if m[6] >= 0 {
				frac := rest[m[6]:m[7]]
				ms = atoi((frac + "00")[:3])
			}
			stamps = append(stamps, atoi(rest[m[2]:m[3]])*60000+atoi(rest[m[4]:m[5]])*1000+ms)
			rest = rest[m[1]:]
		}
		for _, start := range stamps {
			lines = append(lines, lrcLine{start: start, text: strings.TrimSpace(rest)})
		}
	}
	return lines
}

// align finds for each line the text of other that starts at about the same
// time, "" without one. NetEase writes "//" for a line it has no translation
// of. Both lists are walked in order, as SPlayer's alignLyrics does.
func align(lines []yrcLine, other []lrcLine) []string {
	out := make([]string, len(lines))
	for i, j := 0, 0; i < len(lines) && j < len(other); {
		diff := lines[i].start - other[j].start
		switch {
		case diff >= -alignTolerance && diff <= alignTolerance:
			if t := other[j].text; t != "//" {
				out[i] = t
			}
			i++
			j++
		case diff < 0:
			i++
		default:
			j++
		}
	}
	return out
}

// mergeLRC writes translation and romanization under each line with the same
// stamp: the player's parser takes the second line of a stamp as the
// translation and the third as the romanization. A song with a romanization
// but no translation therefore shows it in the translation's place.
func mergeLRC(lines, translation, roman []lrcLine) string {
	asYRC := make([]yrcLine, len(lines))
	for i, l := range lines {
		asYRC[i] = yrcLine{start: l.start}
	}
	trans := align(asYRC, translation)
	rom := align(asYRC, roman)
	var b strings.Builder
	for i, l := range lines {
		stamp := fmt.Sprintf("[%02d:%02d.%03d]", l.start/60000, l.start/1000%60, l.start%1000)
		b.WriteString(stamp + l.text + "\n")
		if l.text == "" {
			continue
		}
		if trans[i] != "" {
			b.WriteString(stamp + trans[i] + "\n")
		}
		if rom[i] != "" {
			b.WriteString(stamp + rom[i] + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func ttmlTime(ms int) string {
	return fmt.Sprintf("%02d:%02d:%02d.%03d", ms/3600000, ms/60000%60, ms/1000%60, ms%1000)
}

func escapeXML(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

// renderTTML writes the lines in the form the player's AMLL parser reads. The
// spaces between words go between the spans, which is where that parser and
// Apple-style TTML keep them.
func renderTTML(lines []yrcLine, translation, roman []string) string {
	var b strings.Builder
	b.WriteString(`<tt xmlns="http://www.w3.org/ns/ttml" xmlns:ttm="http://www.w3.org/ns/ttml#metadata" xmlns:itunes="http://music.apple.com/lyric-ttml-internal" xml:lang="und">`)
	last := lines[len(lines)-1]
	fmt.Fprintf(&b, `<body dur="%s"><div begin="%s" end="%s">`,
		ttmlTime(last.start+last.duration), ttmlTime(lines[0].start), ttmlTime(last.start+last.duration))
	for i, line := range lines {
		fmt.Fprintf(&b, `<p begin="%s" end="%s" itunes:key="L%d" ttm:agent="v1">`,
			ttmlTime(line.start), ttmlTime(line.start+line.duration), i+1)
		writeWords(&b, line.words)
		if bg := line.bg; bg != nil {
			fmt.Fprintf(&b, `<span ttm:role="x-bg" begin="%s" end="%s">`,
				ttmlTime(bg.start), ttmlTime(bg.start+bg.duration))
			writeWords(&b, bg.words)
			b.WriteString(`</span>`)
		}
		if translation[i] != "" {
			fmt.Fprintf(&b, `<span ttm:role="x-translation" xml:lang="zh-CN">%s</span>`, escapeXML(translation[i]))
		}
		if roman[i] != "" {
			fmt.Fprintf(&b, `<span ttm:role="x-roman">%s</span>`, escapeXML(roman[i]))
		}
		b.WriteString(`</p>`)
	}
	b.WriteString(`</div></body></tt>`)
	return b.String()
}

func writeWords(b *strings.Builder, words []yrcWord) {
	for _, w := range words {
		word := strings.TrimSpace(w.text)
		if word == "" {
			b.WriteString(" ")
			continue
		}
		if strings.HasPrefix(w.text, " ") {
			b.WriteString(" ")
		}
		fmt.Fprintf(b, `<span begin="%s" end="%s">%s</span>`,
			ttmlTime(w.start), ttmlTime(w.start+w.duration), escapeXML(word))
		if strings.HasSuffix(w.text, " ") {
			b.WriteString(" ")
		}
	}
}
