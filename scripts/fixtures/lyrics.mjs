// Runs in the browser, bundled by baseline.mjs: AMLL's TTML parser needs
// DOMParser. Converts a NetEase lyric into the format Sonora stores, parses
// it back the way Sonora parses stored lyrics and compares the lines.

import {
  parseTTML,
  parseYrc,
  stringifyTTML,
} from "@applemusic-like-lyrics/lyric";

import { parseLrc } from "../../reference/SPlayer/src/utils/lyric/parseLrc.ts";
import {
  alignLyricLines,
  alignLyrics,
  LrcFormat,
  parseSmartLrc,
} from "../../reference/SPlayer/src/utils/lyric/lyricParser.ts";
import { cleanTTMLTranslations } from "../../reference/SPlayer/src/utils/lyric/parseTTML.ts";

// fromNetEase builds the lines the way SPlayer's LyricManager does for a
// NetEase song (adoptLRC): the line or word lyric, then translation and
// romaji aligned to it.
function fromNetEase(raw) {
  const align = (lines, text, key) =>
    text ? alignLyrics(lines, parseLrc(text), key) : lines;
  if (raw.yrc?.lyric) {
    let lines = parseYrc(raw.yrc.lyric);
    lines = align(lines, raw.ytlrc?.lyric, "translatedLyric");
    return {
      word: true,
      lines: align(lines, raw.yromalrc?.lyric, "romanLyric"),
    };
  }
  let lines = parseLrc(raw.lrc.lyric);
  lines = align(lines, raw.tlyric?.lyric, "translatedLyric");
  return { word: false, lines: align(lines, raw.romalrc?.lyric, "romanLyric") };
}

const stamp = (ms) => {
  const m = Math.floor(ms / 60_000);
  const s = ((ms % 60_000) / 1000).toFixed(3).padStart(6, "0");
  return `[${String(m).padStart(2, "0")}:${s}]`;
};

// toLrc writes line lyrics as one LRC: translation and romaji follow the
// line under the same timestamp, and an empty line marks where a line
// ends when the next one starts later, since the LRC parser ends a line
// at the next timestamp.
function toLrc(lines) {
  const out = [];
  lines.forEach((line, i) => {
    if (line.words.length !== 1) throw new Error("not a line lyric");
    if (line.romanLyric && !line.translatedLyric)
      throw new Error(`romaji without translation at ${line.startTime}`);
    const t = stamp(line.startTime);
    out.push(t + line.words[0].word);
    if (line.translatedLyric) out.push(t + line.translatedLyric);
    if (line.romanLyric) out.push(t + line.romanLyric);
    const next = lines[i + 1]?.startTime ?? line.startTime + 10_000;
    if (line.endTime !== next) out.push(stamp(line.endTime));
  });
  return out.join("\n") + "\n";
}

// parseStored parses a stored lyric the way Sonora does: TTML with AMLL
// after SPlayer's translation cleanup; LRC with SPlayer's format
// detection, and line LRC aligned into translation and romaji.
function parseStored({ format, lyrics }) {
  if (format === "ttml")
    return {
      word: true,
      lines: parseTTML(cleanTTMLTranslations(lyrics)).lines,
    };
  const { format: lrc, lines } = parseSmartLrc(lyrics);
  if (lrc !== LrcFormat.Line) return { word: true, lines };
  return { word: false, lines: alignLyricLines(lines) };
}

// The fields of a line that reach the screen.
const shape = (lines) =>
  lines.map((l) => ({
    startTime: l.startTime,
    endTime: l.endTime,
    words: l.words.map((w) => ({
      startTime: w.startTime,
      endTime: w.endTime,
      word: w.word,
    })),
    translatedLyric: l.translatedLyric,
    romanLyric: l.romanLyric,
    isBG: l.isBG,
    isDuet: l.isDuet,
  }));

function difference(a, b) {
  a = shape(a);
  b = shape(b);
  if (a.length !== b.length) return `${a.length} lines against ${b.length}`;
  const i = a.findIndex(
    (line, i) => JSON.stringify(line) !== JSON.stringify(b[i]),
  );
  return i < 0
    ? null
    : `line ${i}: ${JSON.stringify(a[i])} against ${JSON.stringify(b[i])}`;
}

/**
 * convert returns the stored lyric for a NetEase lyric response and the
 * problems found: against the lines SPlayer builds from the NetEase files,
 * and against the lines the reference player held (`songLyric`: its
 * lrcData for line lyrics, yrcData for word lyrics).
 */
export function convert(raw, songLyric) {
  const source = fromNetEase(raw);
  const stored = source.word
    ? {
        format: "ttml",
        lyrics: stringifyTTML({ lines: source.lines, metadata: [] }),
      }
    : { format: "lrc", lyrics: toLrc(source.lines) };
  const parsed = parseStored(stored);
  return {
    stored,
    word: source.word,
    againstSource: [
      parsed.word === source.word ? null : "word timing lost",
      difference(parsed.lines, source.lines),
    ].filter(Boolean),
    againstPlayer: [
      difference(parsed.lines, songLyric[source.word ? "yrcData" : "lrcData"]),
    ].filter(Boolean),
  };
}
