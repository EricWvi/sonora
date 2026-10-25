// The player states of the baseline, shared by the SPlayer capture
// (splayer-player.mjs) and the Sonora capture (scripts/visual/player.mjs)
// so both take the same steps and clip the same areas.

import { readFile } from "node:fs/promises";

// --- data ---------------------------------------------------------------

// The songs the states play. Each is reached by playing its album and
// switching to it, as a listener would.
export const songs = {
  // THE BOOK, track 8: LRC with translation and romaji, no word timing.
  yoru: { album: 99079344, id: 1500151583, index: 7 },
  // 紅蓮華 single: word-by-word (YRC) with translation and romaji.
  guren: { album: 78702707, id: 1360592706, index: 0 },
};

// Seek points fall mid-line so a paused player shows a settled lyric.
export const YORU_LINE = 54_000; // 52.408–56.848 どこか儚い空気を纏う君は
export const GUREN_WORD = 25_600; // 心 at 25.230–25.950, half lit

// --- states -------------------------------------------------------------

export const states = [
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
    // The page behind differs between SPlayer and Sonora (hideMain).
    hideMain: true,
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
    // The page behind differs between SPlayer and Sonora (hideMain).
    hideMain: true,
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
    // The page behind differs between SPlayer and Sonora (hideMain).
    hideMain: true,
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
    // Held mid-drag: the drop line, the drag label and the dragged item.
    id: "queue-drag",
    song: "yoru",
    seek: YORU_LINE,
    clip: "drawer",
    async act(page) {
      await setStore(page, "status", { playListShow: true });
      const nodes = page.locator(".n-drawer .song-node");
      await nodes.nth(5).waitFor();
      await page.waitForTimeout(800);
      const handle = await nodes.nth(1).locator(".drag-handle").boundingBox();
      // The lower half of the fifth item: the line falls between 5 and 6.
      const target = await nodes.nth(4).boundingBox();
      await page.mouse.move(
        handle.x + handle.width / 2,
        handle.y + handle.height / 2,
      );
      await page.mouse.down();
      await page.mouse.move(
        handle.x + handle.width / 2,
        target.y + target.height * 0.75,
        { steps: 10 },
      );
      await page.waitForTimeout(600);
      // Released after the screenshot when the context closes.
    },
  },
  {
    id: "queue-empty",
    song: null,
    clip: "drawer",
    // Without a song there is no bar, and the page behind shows through
    // the drawer's rounded corner (hideMain).
    hideMain: true,
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
      const { width, height } = page.viewportSize();
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

// hideMain hides the page behind the player for states whose clip takes
// in part of it, so only the player and its popovers are compared: #main,
// and the song list's floating buttons, which SPlayer's album page
// teleports to <body>.
export async function hideMain(page) {
  await page.addStyleTag({
    content: "#main, body > .list-menu { visibility: hidden !important; }",
  });
}

// pinia reaches every SPlayer store; the reference build keeps
// __vue_app__ on #app.
export const PINIA = `document.querySelector("#app").__vue_app__.config.globalProperties.$pinia`;

export async function setStore(page, store, patch) {
  await page.evaluate(
    ([pinia, store, patch]) => eval(pinia)._s.get(store).$patch(patch),
    [PINIA, store, patch],
  );
}

// settled waits for web fonts and the images in view; lazy images below
// the fold never load.
export async function settled(page) {
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

// Clips keep the player and leave out the page behind it, which shows
// NetEase content that changes between runs (hot searches, recommendations).
//   bar     the bottom bar
//   zone    the bottom bar and the 180 px above it, where the progress
//           tooltip and the volume popover open
//   drawer  the queue drawer
//   (none)  the whole viewport, for the full-screen player
export async function clipFor(page, state) {
  const { width, height } = page.viewportSize();
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

// A run with --only captures some states again; the manifest keeps the
// entries of the other states from the earlier manifest, in table order.
export async function mergeManifest(file, manifest) {
  let earlier;
  try {
    earlier = JSON.parse(await readFile(file, "utf8"));
  } catch {
    return manifest;
  }
  const now = new Set(manifest.states.map((s) => s.id));
  const kept = earlier.states.filter((s) => !now.has(s.id));
  const order = new Map(states.map((s, i) => [s.id, i]));
  manifest.states = [...kept, ...manifest.states].sort(
    (a, b) => (order.get(a.id) ?? 1e9) - (order.get(b.id) ?? 1e9),
  );
  return manifest;
}
