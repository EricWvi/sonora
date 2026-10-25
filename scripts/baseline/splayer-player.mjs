// Captures the player baseline from the SPlayer reference instance: the
// bottom bar, the queue drawer and the full-screen player in a fixed set
// of states. See docs/phase/phase-2-player-baseline.md.
//
// Every state starts in a fresh BrowserContext with the reference session
// (scripts/splayer-session.mjs), plays the same NetEase songs, seeks to a
// fixed time and sets the store state directly. Screenshots, the data
// snapshot and a manifest per viewport go to .local/baseline/splayer/<run>/;
// they show account details and never leave .local.
//
//   node scripts/baseline/splayer-player.mjs [--run NAME] [--only REGEX]
//       [--viewport 1440x1000] [--base http://127.0.0.1:25884] [--data-only]
//
// --sonora captures the reference the way Sonora is to look: SPlayer's own
// settings hide what Eric decided Sonora does not show (the quality tag
// and quality switch, the tag row under the title, the comment and
// download buttons), and the song alias is cleared. Playing states stop
// at their seek point and are marked playing in the store, and the page
// behind the bar is hidden where a clip takes it in. Sonora's screenshots
// are compared with this set; the plain run stays the record of SPlayer.
//
// --data-only saves only the data snapshot and what the player holds for
// each song (data/<song>-player.json), without screenshots or a manifest.
//
// Playwright comes from ../palace; set PLAYWRIGHT_TEST to another
// @playwright/test index.mjs to use a different install.

import { execFileSync } from "node:child_process";
import { mkdir, writeFile } from "node:fs/promises";
import os from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { parseArgs } from "node:util";

import { prepareSPlayerSession } from "../splayer-session.mjs";
import {
  clipFor,
  GUREN_WORD,
  hideMain,
  PINIA,
  setStore,
  settled,
  songs,
  states,
  YORU_LINE,
} from "./player-states.mjs";

const repo = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const { chromium } = await import(
  pathToFileURL(
    process.env.PLAYWRIGHT_TEST ??
      resolve(repo, "../palace/node_modules/@playwright/test/index.mjs"),
  ).href
);

const { values: args } = parseArgs({
  options: {
    run: {
      type: "string",
      default: new Date().toISOString().replace(/[:.]/g, "-"),
    },
    only: { type: "string" },
    viewport: { type: "string", default: "1440x1000" },
    base: { type: "string", default: "http://127.0.0.1:25884" },
    "data-only": { type: "boolean", default: false },
    sonora: { type: "boolean", default: false },
  },
});
const [width, height] = args.viewport.split("x").map(Number);
const base = args.base;
const out = join(repo, ".local/baseline/splayer", args.run);

// Settings every state starts from. The schema version and agreement
// version keep SPlayer from migrating them or opening the agreement.
const baseSettings = {
  schemaVersion: 12,
  userAgreementVersion: "v2.0",
  themeMode: "light",
  // Shown states are captured with the controls pinned; auto-hiding is
  // its own state.
  autoHidePlayerMeta: false,
};

// Settings that hide what Sonora removes (--sonora).
const sonoraSettings = {
  showPlayerQuality: false,
  showPlayMeta: false,
  fullscreenPlayerElements: { comments: false, download: false },
};
if (args.sonora) Object.assign(baseSettings, sonoraSettings);

// --- helpers --------------------------------------------------------------

const CONTROLLER = `window.__SPLAYER_PLAYER_CONTROLLER__`;

async function readState(page) {
  return page.evaluate((pinia) => {
    const p = eval(pinia);
    const status = p._s.get("status");
    const music = p._s.get("music");
    return {
      id: music.playSong.id,
      playing: status.playStatus,
      loading: status.playLoading,
      currentTime: status.currentTime,
      duration: status.duration,
      lrc: music.songLyric.lrcData.length,
      yrc: music.songLyric.yrcData.length,
    };
  }, PINIA);
}

async function waitUntil(page, what, fn, timeout = 30_000) {
  const start = Date.now();
  for (;;) {
    const s = await readState(page);
    if (fn(s)) return s;
    if (Date.now() - start > timeout)
      throw new Error(`timed out waiting for ${what}: ${JSON.stringify(s)}`);
    await page.waitForTimeout(200);
  }
}

async function audioStill(page) {
  let last = null;
  for (let i = 0; i < 50; i++) {
    const t = await page.evaluate((c) => eval(c).getSeek(), CONTROLLER);
    if (t === last) return;
    last = t;
    await page.waitForTimeout(300);
  }
  throw new Error("audio keeps playing after pause");
}

async function openState(browser, state) {
  const context = await newContext(browser, state);
  try {
    return { context, ...(await prepareState(context, state)) };
  } catch (err) {
    await context.close();
    throw err;
  }
}

async function newContext(browser, state) {
  const context = await browser.newContext({
    viewport: { width, height },
    deviceScaleFactor: 1,
    locale: "zh-CN",
    timezoneId: "Asia/Shanghai",
    colorScheme: "light",
  });
  await prepareSPlayerSession(context, base);
  const settings = { ...baseSettings, ...state.settings };
  await context.addInitScript(
    ([origin, settings]) => {
      if (location.origin !== origin || localStorage.getItem("setting-store"))
        return;
      localStorage.setItem("setting-store", JSON.stringify(settings));
    },
    [new URL(base).origin, settings],
  );
  return context;
}

// prepareState plays the state's song, seeks, pauses and applies the
// state's own steps.
async function prepareState(context, state) {
  const page = await context.newPage();
  const errors = [];
  page.on("pageerror", (e) => errors.push(e.message));

  if (!state.song) {
    await page.goto(base + "/#/", { waitUntil: "networkidle" });
  } else {
    const song = songs[state.song];
    await page.goto(`${base}/#/album?id=${song.album}`, {
      waitUntil: "networkidle",
    });
    await page.getByRole("button", { name: "播放" }).first().click();
    await waitUntil(page, "playback", (s) => s.id !== 0 && s.playing);
    if (song.index > 0) {
      await page.evaluate(
        ([c, i]) => eval(c).togglePlayIndex(i, true),
        [CONTROLLER, song.index],
      );
    }
    await waitUntil(
      page,
      "the song",
      (s) => s.id === song.id && s.playing && s.duration > 0 && s.lrc > 0,
    );
    await page.evaluate(
      ([c, t]) => eval(c).setSeek(t),
      [CONTROLLER, state.seek],
    );
    // --sonora pauses playing states too and only marks them playing
    // below: the UI reads the store, and a running clock would leave the
    // screenshot at whatever time it was taken.
    if (!state.playing || args.sonora) {
      await page.evaluate((c) => eval(c).pause(), CONTROLLER);
      await waitUntil(page, "pause", (s) => !s.playing);
      // pause() fades out and stops the audio from a timer, and a seek
      // before then cancels it. Wait for the audio to stand still, then
      // seek so word highlights stop at the same point every time.
      await audioStill(page);
      await page.evaluate(
        ([c, t]) => eval(c).setSeek(t),
        [CONTROLLER, state.seek],
      );
      await audioStill(page);
    }
  }
  // Sonora has no song alias (--sonora).
  if (args.sonora && state.song)
    await setStore(page, "music", { playSong: { alia: "" } });
  if (args.sonora && state.playing)
    await setStore(page, "status", { playStatus: true });
  if (args.sonora && state.hideMain) await hideMain(page);
  // Rest the pointer where it hovers nothing.
  await page.mouse.move(width / 2, 2);
  if (state.full) {
    await setStore(page, "status", { showFullPlayer: true });
    await page.locator(".full-player").waitFor();
    await page.waitForTimeout(1500); // expand animation and lyric scroll
  }
  await settled(page);
  // Toasts such as 已开始播放 leave after a few seconds.
  await page
    .locator(".n-message")
    .first()
    .waitFor({ state: "detached", timeout: 10_000 })
    .catch(() => {});
  // Vue transitions (play/pause icon, lyric line) finish before acting.
  await page.waitForTimeout(1000);
  if (state.settle) await page.waitForTimeout(state.settle);
  if (state.act) await state.act(page);
  return { page, errors };
}

// fontsOf reports the platform fonts Chrome used to render each element,
// so Sonora can be compared on the same fonts.
async function fontsOf(page, selectors) {
  const cdp = await page.context().newCDPSession(page);
  await cdp.send("DOM.enable");
  await cdp.send("CSS.enable");
  const { root } = await cdp.send("DOM.getDocument", { depth: -1 });
  const fonts = {};
  for (const sel of selectors) {
    const { nodeId } = await cdp.send("DOM.querySelector", {
      nodeId: root.nodeId,
      selector: sel,
    });
    if (!nodeId) continue;
    const { fonts: used } = await cdp.send("CSS.getPlatformFontsForNode", {
      nodeId,
    });
    fonts[sel] = used.map(
      (f) => `${f.familyName} (${f.postScriptName}) ×${f.glyphCount}`,
    );
  }
  await cdp.detach();
  return fonts;
}

// snapshot saves what Sonora needs to replay the same data: song details,
// lyrics and cover images.
async function snapshot(browser) {
  const context = await browser.newContext();
  await prepareSPlayerSession(context, base);
  const page = await context.newPage();
  await page.goto(base + "/#/", { waitUntil: "networkidle" });
  const dir = join(out, "data");
  await mkdir(dir, { recursive: true });
  for (const [key, song] of Object.entries(songs)) {
    const data = await page.evaluate(async ({ album, id }) => {
      const get = async (path) => (await fetch(path)).json();
      return {
        album: await get(`/api/netease/album?id=${album}`),
        lyric: await get(`/api/netease/lyric/new?id=${id}`),
      };
    }, song);
    await writeFile(join(dir, `${key}.json`), JSON.stringify(data, null, 2));
    const picUrl = data.album.album?.picUrl;
    if (picUrl) {
      const res = await page.request.get(picUrl);
      await writeFile(join(dir, `${key}-cover.jpg`), await res.body());
      // The sizes SPlayer asks for (coverSize s, m, l), as NetEase scales
      // them; Sonora's fixture serves these for ?size=.
      for (const size of [100, 300, 1024]) {
        const res = await page.request.get(`${picUrl}?param=${size}y${size}`);
        await writeFile(
          join(dir, `${key}-cover-${size}.jpg`),
          await res.body(),
        );
      }
    }
  }
  await context.close();
  for (const key of Object.keys(songs)) await snapshotPlayer(browser, key);
}

// snapshotPlayer saves what the player holds once the song is paused at
// its seek point: the song, the queue, the lyrics after SPlayer's whole
// pipeline (exclusion, bracket replacement, normalising) and the lyric
// mode. Sonora's fixture is checked against these.
async function snapshotPlayer(browser, key) {
  const seek = { yoru: YORU_LINE, guren: GUREN_WORD }[key];
  const { context, page } = await openState(browser, { song: key, seek });
  try {
    const player = await page.evaluate((pinia) => {
      const p = eval(pinia);
      const status = p._s.get("status");
      const music = p._s.get("music");
      const data = p._s.get("data");
      return JSON.parse(
        JSON.stringify({
          playSong: music.playSong,
          playList: data.playList,
          songLyric: music.songLyric,
          status: {
            playIndex: status.playIndex,
            currentTime: status.currentTime,
            duration: status.duration,
            usingTTMLLyric: status.usingTTMLLyric,
            usingQRCLyric: status.usingQRCLyric,
          },
        }),
      );
    }, PINIA);
    await writeFile(
      join(out, "data", `${key}-player.json`),
      JSON.stringify(player, null, 2),
    );
  } finally {
    await context.close();
  }
}

// --- run ----------------------------------------------------------------------

const browser = await chromium.launch({
  headless: true,
  args: ["--autoplay-policy=no-user-gesture-required"],
});
const manifest = {
  run: args.run,
  startedAt: new Date().toISOString(),
  environment: {
    os: `${os.type()} ${os.release()} ${os.arch()}`,
    browser: `Chromium ${browser.version()} (headless, Playwright)`,
    splayerCommit: execFileSync("git", [
      "-C",
      join(repo, "reference/SPlayer"),
      "rev-parse",
      "HEAD",
    ])
      .toString()
      .trim(),
    base,
    viewport: { width, height },
    deviceScaleFactor: 1,
    locale: "zh-CN",
  },
  songs,
  sonora: args.sonora,
  baseSettings,
  states: [],
};
try {
  if (!args["data-only"])
    await mkdir(join(out, args.viewport), { recursive: true });
  await snapshot(browser);
  const only = args.only ? new RegExp(args.only) : null;
  let fontsRecorded = false;
  for (const state of args["data-only"] ? [] : states) {
    if (only && !only.test(state.id)) continue;
    const entry = {
      id: state.id,
      song: state.song,
      seek: state.seek ?? null,
      playing: !!state.playing,
      full: !!state.full,
      settings: { ...baseSettings, ...state.settings },
    };
    let ctx;
    try {
      try {
        ctx = await openState(browser, state);
      } catch (err) {
        // The reference instance talks to NetEase; one retry absorbs a
        // slow response.
        console.log(`retry ${state.id}: ${err.message.split("\n")[0]}`);
        ctx = await openState(browser, state);
      }
      const file = join(args.viewport, `${state.id}.png`);
      await ctx.page.screenshot({
        path: join(out, file),
        clip: await clipFor(ctx.page, state),
        animations: state.animations === "allow" ? "allow" : "disabled",
        caret: "hide",
      });
      entry.file = file;
      entry.state = await readState(ctx.page);
      if (!fontsRecorded && state.full) {
        manifest.environment.fonts = await fontsOf(ctx.page, [
          ".main-player .name",
          ".full-player .player-data .name",
          ".full-player .lyric-content .content .text, .full-player .lrc-line",
        ]);
        fontsRecorded = true;
      }
      if (ctx.errors.length) entry.pageErrors = ctx.errors;
      console.log(`ok   ${state.id}`);
    } catch (err) {
      entry.error = err.message;
      console.log(`FAIL ${state.id}: ${err.message}`);
    } finally {
      await ctx?.context.close();
    }
    manifest.states.push(entry);
  }
} finally {
  manifest.finishedAt = new Date().toISOString();
  if (!args["data-only"])
    await writeFile(
      join(out, args.viewport, "manifest.json"),
      JSON.stringify(manifest, null, 2),
    );
  await browser.close();
}
console.log(out);
