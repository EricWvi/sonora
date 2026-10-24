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
//       [--viewport 1440x1000] [--base http://127.0.0.1:25884]
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
  },
});
const [width, height] = args.viewport.split("x").map(Number);
const base = args.base;
const out = join(repo, ".local/baseline/splayer", args.run);

// --- data ---------------------------------------------------------------

// The songs the states play. Each is reached by playing its album and
// switching to it, as a listener would.
const songs = {
  // THE BOOK, track 8: LRC with translation and romaji, no word timing.
  yoru: { album: 99079344, id: 1500151583, index: 7 },
  // 紅蓮華 single: word-by-word (YRC) with translation and romaji.
  guren: { album: 78702707, id: 1360592706, index: 0 },
};

// Seek points fall mid-line so a paused player shows a settled lyric.
const YORU_LINE = 54_000; // 52.408–56.848 どこか儚い空気を纏う君は
const GUREN_WORD = 25_600; // 心 at 25.230–25.950, half lit

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

// --- states -------------------------------------------------------------

const states = [
  // Bottom bar
  { id: "bar-paused-start", song: "yoru", seek: 0, clip: "bar" },
  { id: "bar-paused", song: "yoru", seek: YORU_LINE, clip: "bar" },
  {
    id: "bar-playing",
    song: "yoru",
    seek: YORU_LINE,
    playing: true,
    clip: "bar",
  },
  {
    id: "bar-paused-dark",
    song: "yoru",
    seek: YORU_LINE,
    settings: { themeMode: "dark" },
    clip: "bar",
  },
  {
    // Naive UI's slider shows the tooltip over its handle.
    id: "bar-progress-hover",
    song: "yoru",
    seek: YORU_LINE,
    clip: "zone",
    async act(page) {
      await page.locator(".main-player .n-slider-handle").hover();
      await page.waitForTimeout(600);
    },
  },
  {
    id: "bar-progress-drag",
    song: "yoru",
    seek: YORU_LINE,
    clip: "zone",
    async act(page) {
      const handle = await page
        .locator(".main-player .n-slider-handle")
        .boundingBox();
      const rail = await page.locator(".main-player .n-slider").boundingBox();
      await page.mouse.move(
        handle.x + handle.width / 2,
        handle.y + handle.height / 2,
      );
      await page.mouse.down();
      await page.mouse.move(
        rail.x + rail.width * 0.4,
        handle.y + handle.height / 2,
        { steps: 10 },
      );
      await page.waitForTimeout(600);
      // Released after the screenshot when the context closes.
    },
  },
  {
    id: "bar-volume-hover",
    song: "yoru",
    seek: YORU_LINE,
    clip: "zone",
    async act(page) {
      // PlayerRightMenu: other controls, volume, queue.
      await page.locator(".main-player .right-menu .menu-icon").nth(1).hover();
      await page.waitForTimeout(800);
    },
  },

  // Queue drawer
  {
    id: "queue-open",
    song: "yoru",
    seek: YORU_LINE,
    clip: "drawer",
    async act(page) {
      await setStore(page, "status", { playListShow: true });
      await page.locator(".n-drawer .song-node").first().waitFor();
      await page.waitForTimeout(800);
    },
  },
  {
    id: "queue-hover",
    song: "yoru",
    seek: YORU_LINE,
    clip: "drawer",
    async act(page) {
      await setStore(page, "status", { playListShow: true });
      const item = page.locator(".n-drawer .song-node").nth(2);
      await item.waitFor();
      await page.waitForTimeout(800);
      await item.hover();
      await page.waitForTimeout(400);
    },
  },
  {
    id: "queue-empty",
    song: null,
    clip: "drawer",
    async act(page) {
      await setStore(page, "status", { playListShow: true });
      await page.locator(".n-drawer").waitFor();
      await page.waitForTimeout(800);
    },
  },

  // Full-screen player: styles and backgrounds
  ...[
    ["cover", "blur"],
    ["cover", "color"],
    ["cover", "none"],
    ["record", "blur"],
    ["fullscreen", "blur"],
  ].map(([type, bg]) => ({
    id: `full-${type}-${bg}`,
    song: "yoru",
    seek: YORU_LINE,
    settings: { playerType: type, playerBackgroundType: bg },
    full: true,
  })),
  {
    id: "full-cover-blur-dark",
    song: "yoru",
    seek: YORU_LINE,
    settings: { themeMode: "dark" },
    full: true,
  },
  {
    id: "full-cover-blur-playing",
    song: "yoru",
    seek: YORU_LINE,
    playing: true,
    full: true,
  },
  {
    id: "full-cover-blur-hidden",
    song: "yoru",
    seek: YORU_LINE,
    settings: { autoHidePlayerMeta: true },
    full: true,
    async act(page) {
      // Leave the player idle until the controls hide (3 s in SPlayer).
      await page.mouse.move(width / 2, height / 2);
      await page.waitForTimeout(4000);
    },
  },
  ...[0, 2000].map((t) => ({
    id: `full-cover-animation-t${t}`,
    song: "yoru",
    seek: YORU_LINE,
    settings: { playerBackgroundType: "animation" },
    full: true,
    // The flowing background is a canvas; frames are compared by time.
    animations: "allow",
    async act(page) {
      await page.waitForTimeout(1500 + t);
    },
  })),

  // Full-screen player: lyrics
  {
    id: "full-lrc-plain",
    song: "yoru",
    seek: YORU_LINE,
    settings: { showTran: false, showRoma: false },
    full: true,
  },
  {
    id: "full-lrc-amll",
    song: "yoru",
    seek: YORU_LINE,
    settings: { useAMLyrics: true },
    full: true,
    // AMLL animates lines with JS springs, which CSS animation control
    // does not stop; give them time to settle.
    settle: 3000,
  },
  { id: "full-yrc", song: "guren", seek: GUREN_WORD, full: true },
  {
    id: "full-yrc-amll",
    song: "guren",
    seek: GUREN_WORD,
    settings: { useAMLyrics: true },
    full: true,
    // AMLL animates lines with JS springs, which CSS animation control
    // does not stop; give them time to settle.
    settle: 3000,
  },
];

// --- helpers --------------------------------------------------------------

// pinia reaches every SPlayer store; the reference build keeps
// __vue_app__ on #app.
const PINIA = `document.querySelector("#app").__vue_app__.config.globalProperties.$pinia`;
const CONTROLLER = `window.__SPLAYER_PLAYER_CONTROLLER__`;

async function setStore(page, store, patch) {
  await page.evaluate(
    ([pinia, store, patch]) => eval(pinia)._s.get(store).$patch(patch),
    [PINIA, store, patch],
  );
}

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

// settled waits for web fonts and the images in view; lazy images below
// the fold never load.
async function settled(page) {
  await page.evaluate(async () => {
    await document.fonts.ready;
    const inView = (i) => {
      const r = i.getBoundingClientRect();
      return (
        r.width > 0 &&
        r.bottom > 0 &&
        r.top < innerHeight &&
        r.right > 0 &&
        r.left < innerWidth
      );
    };
    const pending = [...document.images].filter(
      (i) => !i.complete && inView(i),
    );
    await Promise.race([
      Promise.all(
        pending.map(
          (i) =>
            new Promise(
              (r) =>
                i.addEventListener("load", r) || i.addEventListener("error", r),
            ),
        ),
      ),
      new Promise((r) => setTimeout(r, 10_000)),
    ]);
  });
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
    if (!state.playing) {
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

// Clips keep the player and leave out the page behind it, which shows
// NetEase content that changes between runs (hot searches, recommendations).
//   bar     the bottom bar
//   zone    the bottom bar and the 180 px above it, where the progress
//           tooltip and the volume popover open
//   drawer  the queue drawer
//   (none)  the whole viewport, for the full-screen player
async function clipFor(page, state) {
  const round = (b) => ({
    x: Math.round(b.x),
    y: Math.round(b.y),
    width: Math.round(b.width),
    height: Math.round(b.height),
  });
  switch (state.clip) {
    case "bar": {
      const box = await page.locator(".main-player").boundingBox();
      return round({ x: 0, y: box.y, width, height: box.height });
    }
    case "zone":
      return { x: 0, y: height - 260, width, height: 260 };
    case "drawer":
      return round(await page.locator(".n-drawer").boundingBox());
  }
  return undefined;
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
    }
  }
  await context.close();
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
  baseSettings,
  states: [],
};
try {
  await mkdir(join(out, args.viewport), { recursive: true });
  await snapshot(browser);
  const only = args.only ? new RegExp(args.only) : null;
  let fontsRecorded = false;
  for (const state of states) {
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
  await writeFile(
    join(out, args.viewport, "manifest.json"),
    JSON.stringify(manifest, null, 2),
  );
  await browser.close();
}
console.log(out);
