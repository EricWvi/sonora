// Turns the player baseline's song data into a Sonora fixture: tracks,
// albums and artists in the shape of the player API, lyrics in the
// formats Sonora stores, the covers, and the queue and position of each
// baseline state. See docs/phase/phase-3-player-shell.md, 3.3.
//
// The baseline data is NetEase content and stays in .local: this reads
// .local/baseline/splayer/<run>/data/ and writes .local/fixtures/baseline/,
// which the dev server's `?fixture=baseline` scenario loads.
//
//   node scripts/fixtures/baseline.mjs [--run 2026-10-25]
//
// Sonora stores LRC and TTML only. 夜に駆ける (LRC with translation and
// romaji in separate NetEase files) becomes one LRC with the translation
// and romaji under the same timestamp; 紅蓮華 (YRC word timing) becomes
// TTML, the only stored format that keeps each word's start and end. The
// conversion goes through SPlayer's own parsers, and the result is parsed
// back the way Sonora parses stored lyrics and compared line by line with
// what SPlayer built from the NetEase files, and with what the reference
// player held (data/<song>-player.json from splayer-player.mjs
// --data-only).

import { createHash } from "node:crypto";
import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { parseArgs } from "node:util";

const repo = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const { values: args } = parseArgs({
  options: { run: { type: "string", default: "2026-10-25" } },
});
const data = join(repo, ".local/baseline/splayer", args.run, "data");
const out = join(repo, ".local/fixtures/baseline");

// The baseline songs and where each state stops; these match
// scripts/baseline/splayer-player.mjs.
const songs = {
  yoru: { id: 1500151583, seek: 54_000 },
  guren: { id: 1360592706, seek: 25_600 },
};

// --- lyrics -----------------------------------------------------------------------

// The conversion runs in Chromium (lyrics.mjs): AMLL's TTML parser needs
// DOMParser. esbuild comes with Vite; lodash-es and AMLL resolve from
// frontend/node_modules.
const frontend = join(repo, "frontend");
const esbuild = createRequire(
  createRequire(join(frontend, "package.json")).resolve("vite/package.json"),
)("esbuild");
const bundle = await esbuild.build({
  entryPoints: [join(repo, "scripts/fixtures/lyrics.mjs")],
  bundle: true,
  format: "iife",
  globalName: "lyrics",
  nodePaths: [join(frontend, "node_modules")],
  write: false,
  logLevel: "silent",
});
const { chromium } = await import(
  pathToFileURL(
    process.env.PLAYWRIGHT_TEST ??
      resolve(repo, "../palace/node_modules/@playwright/test/index.mjs"),
  ).href
);
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
await page.addScriptTag({ content: bundle.outputFiles[0].text });
const convert = (raw, songLyric) =>
  page.evaluate(([r, l]) => window.lyrics.convert(r, l), [raw, songLyric]);

// --- library --------------------------------------------------------------------

await rm(out, { recursive: true, force: true });
await mkdir(join(out, "covers"), { recursive: true });

const artists = new Map(); // NetEase id → Sonora artist
const artistRef = (ar) => {
  if (!artists.has(ar.id))
    artists.set(ar.id, {
      id: artists.size + 1,
      name: ar.name,
      image: null,
      trackCount: 0,
      albumCount: 0,
    });
  const { id, name } = artists.get(ar.id);
  return { id, name };
};
const artistOf = (ref) => [...artists.values()].find((a) => a.id === ref.id);

const albums = [];
const tracks = [];
const lyrics = {};
const players = {};
const checks = [];
const trackIds = new Map(); // NetEase id → Sonora id

for (const [key, song] of Object.entries(songs)) {
  const netease = JSON.parse(await readFile(join(data, `${key}.json`)));
  const player = JSON.parse(await readFile(join(data, `${key}-player.json`)));

  // Covers are named like the server names them: the first 16 bytes of
  // the SHA-256, hex. No ThumbHash encoder here, so the placeholder is
  // empty, as for an image the server cannot decode.
  const image = await readFile(join(data, `${key}-cover.jpg`));
  const name =
    createHash("sha256").update(image).digest("hex").slice(0, 32) + ".jpg";
  await writeFile(join(out, "covers", name), image);
  // NetEase's own thumbnails at the sizes the player asks for, so the
  // fixture shows the very pixels the reference showed.
  for (const size of [100, 300, 1024]) {
    const thumb = name.replace(/\.jpg$/, `-${size}.jpg`);
    await writeFile(
      join(out, "covers", thumb),
      await readFile(join(data, `${key}-cover-${size}.jpg`)),
    );
  }
  const cover = { name, thumbhash: "" };

  const al = netease.album.album;
  const albumArtists = al.artists.map(artistRef);
  // The year as the reference shows it, in Asia/Shanghai.
  const year = Number(
    new Date(al.publishTime).toLocaleString("en", {
      year: "numeric",
      timeZone: "Asia/Shanghai",
    }),
  );
  const album = {
    id: albums.length + 1,
    title: al.name,
    artists: albumArtists,
    year,
    cover,
    trackCount: 0,
    durationMs: 0,
  };
  albums.push(album);
  for (const ar of albumArtists) artistOf(ar).albumCount++;

  for (const s of netease.album.songs) {
    const track = {
      id: tracks.length + 1,
      title: s.name,
      artists: s.ar.map(artistRef),
      album: { id: album.id, title: album.title, artists: albumArtists },
      trackNo: s.no,
      discNo: Number(s.cd) || 1,
      year,
      durationMs: s.dt,
      // The reference streams FLAC for these (SQ, Hi-Res); the format is
      // not shown anywhere.
      format: "flac",
      cover,
      hasLyrics: s.id === song.id,
    };
    tracks.push(track);
    trackIds.set(s.id, track.id);
    album.trackCount++;
    album.durationMs += s.dt;
    for (const ar of track.artists) artistOf(ar).trackCount++;
  }

  // Lyrics: convert, parse back, compare.
  const result = await convert(netease.lyric, player.songLyric);
  lyrics[trackIds.get(song.id)] = result.stored;
  checks.push(
    [`${key}: ${result.stored.format}`, ...result.againstSource],
    [`${key}: against the reference player`, ...result.againstPlayer],
  );

  // The queue and position the baseline states start from.
  players[key] = {
    queue: player.playList.map((s) => trackIds.get(s.id)),
    index: player.status.playIndex,
    positionMs: song.seek,
    // What the audio element reported, which the player shows once the
    // audio loads; the track's own duration comes from NetEase's metadata.
    durationMs: player.status.duration,
  };
  if (player.playSong.id !== song.id || player.status.currentTime !== song.seek)
    throw new Error(`${key}: the reference player is on another song or time`);
}

await writeFile(
  join(out, "library.json"),
  JSON.stringify(
    { tracks, albums, artists: [...artists.values()], lyrics, players },
    null,
    2,
  ),
);

await browser.close();

let failed = false;
for (const [what, ...problems] of checks) {
  const found = problems.filter(Boolean);
  failed ||= found.length > 0;
  console.log(
    found.length ? `FAIL ${what}\n  ${found.join("\n  ")}` : `ok   ${what}`,
  );
}
console.log(out);
if (failed) process.exitCode = 1;
