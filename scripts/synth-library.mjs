// Fills an empty Sonora data directory with a large synthetic library for
// the phase 6 stability and performance checks: 10,000 tracks, 2,000
// albums, 1,000 artists, and for the development listener 20 playlists of
// 500 tracks, 2,000 likes and a full play history. See
// docs/phase/phase-6-first-release.md.
//
// Uploading cannot build such a library from the real one, since ingest
// keeps one track per file content. This writes the database directly:
// each synthetic track borrows the format, duration, size and lyrics of a
// real track in turn and plays its file through a symlink of its own, so
// deleting a track through the admin API removes only the link. Album
// covers, the odd track cover and artist images reuse the real covers.
//
// The target must be a fresh data directory that a development build
// (-tags dev) has migrated, with one player API request made so the
// development listener exists (GET /api/v1/me), and the service stopped:
//
//   node scripts/synth-library.mjs --source backend/.data \
//     --data /tmp/x/data --media /tmp/x/media [--seed 6]
//
// The same source and seed always give the same library.

import { createHash } from "node:crypto";
import { cpSync, mkdirSync, realpathSync, symlinkSync } from "node:fs";
import { join, resolve } from "node:path";
import { DatabaseSync } from "node:sqlite";
import { parseArgs } from "node:util";

const { values: args } = parseArgs({
  options: {
    source: { type: "string", default: "backend/.data" },
    data: { type: "string" },
    media: { type: "string" },
    seed: { type: "string", default: "6" },
  },
});
if (!args.data || !args.media) {
  throw new Error("--data and --media are required");
}

const SCHEMA_VERSION = 10;
const TRACKS = 10_000;
const ALBUMS = 2_000;
const ARTISTS = 1_000;
const PLAYLISTS = 20;
const PLAYLIST_SIZE = 500;
const LIKES = 2_000;
const HISTORY = 500; // store.MaxHistory

// mulberry32: small, seedable and the same everywhere.
let state = Number(args.seed) >>> 0;
function random() {
  state = (state + 0x6d2b79f5) >>> 0;
  let t = state;
  t = Math.imul(t ^ (t >>> 15), t | 1);
  t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
}
const int = (lo, hi) => lo + Math.floor(random() * (hi - lo + 1));
const pick = (list) => list[Math.floor(random() * list.length)];
const chance = (p) => random() < p;
function sample(list, n) {
  const copy = [...list];
  for (let i = 0; i < n; i++) {
    const j = i + Math.floor(random() * (copy.length - i));
    [copy[i], copy[j]] = [copy[j], copy[i]];
  }
  return copy.slice(0, n);
}

// --- names ------------------------------------------------------------------

const zhSurnames = [
  ..."王李张刘陈杨黄赵吴周徐孙马朱胡郭何林罗高郑梁谢宋唐许韩冯邓曹彭曾萧田董潘袁蔡蒋余于杜叶程魏苏吕丁任沈姚卢姜崔钟谭陆汪范金石廖贾夏韦付方白邹孟熊秦邱江尹薛闫段雷侯龙史陶黎贺顾毛郝龚邵万钱严覃武戴莫孔向汤",
];
const zhGiven = [
  ..."子文俊浩宇轩然欣怡思雨涵梓晨辰一诺可馨悦嘉瑞安泽阳佳琪明亮雪婷海峰志伟晓燕春秋冬夏星月云天心语初若知南北青山清风小乐",
];
const zhBands = ["乐队", "乐团", "组合", "合唱团", "与朋友们", "三重奏"];
const zhWords = [
  ..."夜光海风雨花雪月星河山城梦歌春秋夏冬心路远岸灯火声影时间旅人故乡黄昏清晨天空少年白日飞鸟森林港口站台列车信笺回忆",
];
const zhPhrases = [
  "的",
  "与",
  "之",
  "在",
  "和",
  "向",
  "里的",
  "以外的",
  "尽头的",
  "深处的",
];
const enFirst = [
  "Alex",
  "Bella",
  "Chris",
  "Dana",
  "Eli",
  "Fiona",
  "Grace",
  "Henry",
  "Ivy",
  "Jack",
  "Kate",
  "Leo",
  "Mia",
  "Noah",
  "Olivia",
  "Paul",
  "Quinn",
  "Ruby",
  "Sam",
  "Tara",
  "Uma",
  "Victor",
  "Wendy",
  "Xavier",
  "Yara",
  "Zoe",
];
const enLast = [
  "Adams",
  "Brooks",
  "Carter",
  "Diaz",
  "Evans",
  "Foster",
  "Gray",
  "Hayes",
  "Irwin",
  "Jones",
  "King",
  "Lee",
  "Morgan",
  "Nash",
  "Owens",
  "Parker",
  "Reed",
  "Stone",
  "Turner",
  "Vale",
  "Walsh",
  "Young",
];
const enWords = [
  "Blue",
  "Midnight",
  "Silver",
  "Echo",
  "River",
  "Paper",
  "Golden",
  "Neon",
  "Quiet",
  "Summer",
  "Winter",
  "Wild",
  "Electric",
  "Glass",
  "Ocean",
  "City",
  "Velvet",
  "Lonely",
  "Northern",
  "Crimson",
  "Morning",
  "Shadow",
  "Highway",
  "Dream",
  "Heart",
  "Fire",
  "Rain",
  "Light",
  "Moon",
  "Garden",
];
const jaNames = [
  "佐藤",
  "鈴木",
  "高橋",
  "田中",
  "渡辺",
  "伊藤",
  "山本",
  "中村",
  "小林",
  "加藤",
  "吉田",
  "山田",
  "松本",
  "井上",
  "木村",
];
const jaGiven = [
  "ひかり",
  "さくら",
  "あおい",
  "ゆい",
  "はると",
  "そら",
  "みなと",
  "りん",
  "かえで",
  "ことね",
  "ユウ",
  "ミク",
  "レン",
  "アキ",
  "ナギ",
];
const jaWords = [
  "夜明け",
  "さよなら",
  "ひだまり",
  "青い春",
  "花火",
  "流れ星",
  "君の声",
  "未来へ",
  "ミライ",
  "トキメキ",
  "ハートビート",
  "群青",
  "残響",
  "灯り",
];

function artistName() {
  const r = random();
  if (r < 0.5)
    return (
      pick(zhSurnames) + pick(zhGiven) + (chance(0.6) ? pick(zhGiven) : "")
    );
  if (r < 0.6) return pick(zhWords) + pick(zhWords) + pick(zhBands);
  if (r < 0.75) return `${pick(enFirst)} ${pick(enLast)}`;
  if (r < 0.85) return `The ${pick(enWords)} ${pick(enWords)}s`;
  if (r < 0.9) return `${int(1, 99)} ${pick(enWords)}`;
  return pick(jaNames) + pick(jaGiven);
}

function title(lang) {
  if (lang === "zh") {
    const w = () => pick(zhWords) + pick(zhWords);
    return chance(0.5) ? w() : w() + pick(zhPhrases) + w();
  }
  if (lang === "en") {
    const n = int(1, 3);
    return Array.from({ length: n }, () => pick(enWords)).join(" ");
  }
  return pick(jaWords) + (chance(0.4) ? pick(jaWords) : "");
}

const langOf = (name) =>
  /[぀-ヿ]/.test(name) || jaNames.some((n) => name.startsWith(n))
    ? "ja"
    : /[A-Za-z0-9]/.test(name[0])
      ? "en"
      : "zh";

function unique(make, taken) {
  for (;;) {
    const name = make();
    const key = name.toLowerCase();
    if (!taken.has(key)) {
      taken.add(key);
      return name;
    }
  }
}

// --- timestamps --------------------------------------------------------------

const END = Date.parse("2026-11-01T00:00:00Z");
const START = END - 2 * 365 * 86_400_000;
const stamp = (ms) => new Date(ms).toISOString().replace(/\.\d{3}Z$/, "Z");

// --- source ------------------------------------------------------------------

const source = resolve(args.source);
const src = new DatabaseSync(join(source, "data/sonora.db"), {
  readOnly: true,
});
const realTracks = src
  .prepare(
    `SELECT file, format, size, duration_ms, lyrics, lyrics_format FROM tracks ORDER BY id`,
  )
  .all();
const covers = src
  .prepare(
    `SELECT name, thumbhash FROM covers WHERE thumbhash <> '' ORDER BY name`,
  )
  .all();
src.close();
if (realTracks.length === 0 || covers.length === 0) {
  throw new Error(`no tracks or covers in ${source}`);
}
const sourceMedia = realpathSync(join(source, "media"));

// --- target ------------------------------------------------------------------

const db = new DatabaseSync(join(resolve(args.data), "sonora.db"));
if (db.prepare(`PRAGMA user_version`).get().user_version !== SCHEMA_VERSION) {
  throw new Error(
    `target schema is not version ${SCHEMA_VERSION}; start the service on it once`,
  );
}
if (db.prepare(`SELECT COUNT(*) n FROM tracks`).get().n > 0) {
  throw new Error("target library is not empty");
}
const user = db
  .prepare(`SELECT id FROM users WHERE issuer = 'dev' AND subject = 'listener'`)
  .get();
if (!user) {
  throw new Error("development listener missing; request /api/v1/me once");
}

cpSync(join(source, "data/covers"), join(resolve(args.data), "covers"), {
  recursive: true,
});

db.exec("BEGIN");
const insertCover = db.prepare(
  `INSERT OR IGNORE INTO covers (name, thumbhash) VALUES (?, ?)`,
);
for (const c of covers) insertCover.run(c.name, c.thumbhash);

// Artists, with a few names that start with another artist's name to
// exercise the search ranking.
const artistKeys = new Set();
const artistNames = [];
for (let i = 0; i < 5; i++) {
  const base = unique(artistName, artistKeys);
  artistNames.push(base);
  artistNames.push(unique(() => `${base} Band`, artistKeys));
  artistNames.push(unique(() => `${base}乐队`, artistKeys));
}
while (artistNames.length < ARTISTS)
  artistNames.push(unique(artistName, artistKeys));
const insertArtist = db.prepare(
  `INSERT INTO artists (name, image, created_at) VALUES (?, ?, ?) RETURNING id`,
);
const artists = sample(artistNames, artistNames.length).map((name) => ({
  name,
  id: insertArtist.get(
    name,
    chance(0.3) ? pick(covers).name : null,
    stamp(START),
  ).id,
}));

// Albums. Everyone has one; the rest go to a few artists far more often
// than to others (Zipf), so some artists have dozens.
const weights = artists.map((_, i) => 1 / (i + 1) ** 0.9);
const total = weights.reduce((a, b) => a + b, 0);
function popularArtist() {
  let r = random() * total;
  for (let i = 0; i < artists.length; i++) {
    r -= weights[i];
    if (r <= 0) return artists[i];
  }
  return artists.at(-1);
}
const owners = [...sample(artists, artists.length)];
while (owners.length < ALBUMS) owners.push(popularArtist());

// Tracks per album: singles, EPs and full albums, averaging five, then
// nudged until they add up exactly.
const sizes = owners.map(() => {
  const r = random();
  return r < 0.4 ? 1 : r < 0.65 ? int(3, 5) : int(8, 12);
});
for (let sum = sizes.reduce((a, b) => a + b, 0); sum !== TRACKS;) {
  const i = int(0, ALBUMS - 1);
  if (sum < TRACKS && sizes[i] < 20) (sizes[i]++, sum++);
  else if (sum > TRACKS && sizes[i] > 1) (sizes[i]--, sum--);
}

const albumKeys = new Set();
const insertAlbum = db.prepare(
  `INSERT INTO albums (title, artist_key, year, cover, created_at) VALUES (?, ?, ?, ?, ?) RETURNING id`,
);
const insertAlbumArtist = db.prepare(
  `INSERT INTO album_artists (album_id, artist_id, position) VALUES (?, ?, ?)`,
);
const insertTrack = db.prepare(`
  INSERT INTO tracks (title, album_id, track_no, disc_no, year, duration_ms, format, size, sha256, file,
    original_name, cover, lyrics, lyrics_format, created_at, updated_at)
  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`);
const insertTrackArtist = db.prepare(
  `INSERT INTO track_artists (track_id, artist_id, position) VALUES (?, ?, ?)`,
);

const trackIDs = [];
let n = 0;
for (let a = 0; a < ALBUMS; a++) {
  const albumArtists = [owners[a]];
  if (chance(0.05)) {
    const other = popularArtist();
    if (other !== owners[a]) albumArtists.push(other);
  }
  const lang = langOf(owners[a].name);
  const key = albumArtists
    .map((x) => x.id)
    .sort((x, y) => x - y)
    .join(",");
  const albumTitle = unique(() => title(lang), {
    has: (t) => albumKeys.has(`${t}\n${key}`),
    add: (t) => albumKeys.add(`${t}\n${key}`),
  });
  const year = int(1970, 2026);
  const cover = chance(0.95) ? pick(covers).name : null;
  const added = START + Math.floor(random() * (END - START));
  const albumID = insertAlbum.get(
    albumTitle,
    key,
    year,
    cover,
    stamp(added),
  ).id;
  albumArtists.forEach((x, i) => insertAlbumArtist.run(albumID, x.id, i));

  const size = sizes[a];
  const discs = size >= 8 && chance(0.08) ? 2 : 1;
  const perDisc = Math.ceil(size / discs);
  for (let t = 0; t < size; t++) {
    const real = realTracks[n % realTracks.length];
    const sha = createHash("sha256")
      .update(`sonora-synth:${args.seed}:${n}`)
      .digest("hex");
    const ext = real.file.slice(real.file.lastIndexOf("."));
    const file = `${sha.slice(0, 2)}/${sha}${ext}`;
    mkdirSync(join(args.media, sha.slice(0, 2)), { recursive: true });
    symlinkSync(join(sourceMedia, real.file), join(args.media, file));

    const trackTitle = size === 1 ? albumTitle : title(lang);
    let trackCover = null;
    if (cover && chance(0.03)) {
      do trackCover = pick(covers).name;
      while (trackCover === cover);
    }
    const at = stamp(added + t * 1000);
    const id = insertTrack.get(
      trackTitle,
      albumID,
      discs === 1 ? t + 1 : (t % perDisc) + 1,
      discs === 1 ? 0 : Math.floor(t / perDisc) + 1,
      year,
      real.duration_ms,
      real.format,
      real.size,
      sha,
      file,
      trackTitle + ext,
      trackCover,
      real.lyrics,
      real.lyrics_format,
      at,
      at,
    ).id;
    const trackArtists = [...albumArtists];
    if (chance(0.15)) {
      const feat = pick(artists);
      if (!trackArtists.includes(feat)) trackArtists.push(feat);
    }
    trackArtists.forEach((x, i) => insertTrackArtist.run(id, x.id, i));
    trackIDs.push(id);
    n++;
  }
}

// The development listener's likes, history and playlists.
const insertLike = db.prepare(
  `INSERT INTO likes (user_id, track_id, created_at) VALUES (?, ?, ?)`,
);
sample(trackIDs, LIKES).forEach((id, i) =>
  insertLike.run(
    user.id,
    id,
    stamp(START + Math.floor(((END - START) * i) / LIKES)),
  ),
);
const insertPlay = db.prepare(
  `INSERT INTO play_history (user_id, track_id, played_at) VALUES (?, ?, ?)`,
);
sample(trackIDs, HISTORY).forEach((id, i) =>
  insertPlay.run(user.id, id, stamp(END - (HISTORY - i) * 600_000)),
);

const playlistNames = [
  "通勤路上",
  "深夜电台",
  "周末早晨",
  "跑步",
  "专注工作",
  "雨天",
  "旅行",
  "老歌",
  "Chill",
  "Indie Picks",
  "Road Trip",
  "Workout Mix",
  "Jazz & Coffee",
  "睡前",
  "夏日海边",
  "冬天的歌",
  "J-Pop",
  "合唱",
  "Live 现场",
  "收藏夹",
];
const insertPlaylist = db.prepare(
  `INSERT INTO playlists (user_id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?) RETURNING id`,
);
const insertEntry = db.prepare(
  `INSERT INTO playlist_tracks (playlist_id, track_id, position) VALUES (?, ?, ?)`,
);
for (let p = 0; p < PLAYLISTS; p++) {
  const created = START + Math.floor(random() * (END - START) * 0.8);
  const id = insertPlaylist.get(
    user.id,
    playlistNames[p],
    chance(0.5) ? `${playlistNames[p]}时听的歌` : "",
    stamp(created),
    stamp(created + Math.floor(random() * (END - created))),
  ).id;
  sample(trackIDs, PLAYLIST_SIZE).forEach((t, i) => insertEntry.run(id, t, i));
}
db.exec("COMMIT");

const count = (table) => db.prepare(`SELECT COUNT(*) n FROM ${table}`).get().n;
console.log(
  [
    "artists",
    "albums",
    "tracks",
    "playlists",
    "playlist_tracks",
    "likes",
    "play_history",
    "covers",
  ]
    .map((t) => `${t} ${count(t)}`)
    .join(", "),
);
db.close();
