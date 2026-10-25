// Captures Sonora's player in the states of the player baseline, from the
// dev server's `?fixture=baseline` scenario, for comparison with the
// reference captured by scripts/baseline/splayer-player.mjs --sonora.
// See docs/phase/phase-3-player-shell.md, 3.6.
//
//   node scripts/visual/player.mjs --base http://127.0.0.1:14571
//       [--run NAME] [--only REGEX] [--viewport 1440x1000]
//
// Each state starts in a fresh BrowserContext with the same viewport,
// locale and settings as the reference, opens the player state the
// fixture holds for the state's song and position, then takes the
// reference's steps and clip (player-states.mjs). Screenshots and a
// manifest go to .local/visual/<run>/<viewport>/; they show NetEase
// covers and stay in .local. Compare with
//
//   node scripts/baseline/diff.mjs \
//       .local/baseline/splayer/<run>-sonora/1440x1000 \
//       .local/visual/<run>/1440x1000 --out .local/visual/<run>/diff

import { mkdir, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { parseArgs } from "node:util";

import {
  clipFor,
  hideMain,
  setStore,
  settled,
  states,
} from "../baseline/player-states.mjs";

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
    base: { type: "string", default: "http://127.0.0.1:14559" },
  },
});
const [width, height] = args.viewport.split("x").map(Number);
const base = args.base;
const out = join(repo, ".local/visual", args.run, args.viewport);

// The settings the reference states start from (splayer-player.mjs),
// under Sonora's setting store.
const baseSettings = { themeMode: "light", autoHidePlayerMeta: false };

// The fixture's player state for a baseline state.
function playerOf(state) {
  if (!state.song) return null;
  if (state.seek === 0) return `${state.song}-start`;
  return `${state.song}-${state.playing ? "playing" : "paused"}`;
}

async function openState(browser, state) {
  const context = await browser.newContext({
    viewport: { width, height },
    deviceScaleFactor: 1,
    locale: "zh-CN",
    timezoneId: "Asia/Shanghai",
    colorScheme: "light",
  });
  const settings = { ...baseSettings, ...state.settings };
  await context.addInitScript(
    ([origin, settings]) => {
      if (location.origin !== origin || localStorage.getItem("sonora-setting"))
        return;
      localStorage.setItem("sonora-setting", JSON.stringify(settings));
    },
    [new URL(base).origin, settings],
  );
  try {
    return { context, ...(await prepareState(context, state)) };
  } catch (err) {
    await context.close();
    throw err;
  }
}

async function prepareState(context, state) {
  const page = await context.newPage();
  const errors = [];
  page.on("pageerror", (e) => errors.push(e.message));
  page.on("console", (m) => m.type() === "error" && errors.push(m.text()));
  const player = playerOf(state);
  await page.goto(
    `${base}/?fixture=baseline${player ? `&player=${player}` : ""}#/`,
    { waitUntil: "networkidle" },
  );
  if (player) await page.locator(".main-player.show").waitFor();
  if (state.hideMain) await hideMain(page);
  // Rest the pointer where it hovers nothing.
  await page.mouse.move(width / 2, 2);
  if (state.full) {
    await setStore(page, "status", { showFullPlayer: true });
    await page.locator(".full-player").waitFor({ timeout: 5000 });
    await page.waitForTimeout(1500); // expand animation and lyric scroll
  }
  await settled(page);
  // Vue transitions (play/pause icon, lyric line) finish before acting.
  await page.waitForTimeout(1000);
  if (state.settle) await page.waitForTimeout(state.settle);
  if (state.act) await state.act(page);
  return { page, errors };
}

const browser = await chromium.launch({ headless: true });
const manifest = {
  run: args.run,
  startedAt: new Date().toISOString(),
  environment: {
    browser: `Chromium ${browser.version()} (headless, Playwright)`,
    base,
    viewport: { width, height },
    deviceScaleFactor: 1,
    locale: "zh-CN",
  },
  baseSettings,
  states: [],
};
try {
  await mkdir(out, { recursive: true });
  const only = args.only ? new RegExp(args.only) : null;
  for (const state of states) {
    if (only && !only.test(state.id)) continue;
    const entry = { id: state.id, player: playerOf(state) };
    let ctx;
    try {
      ctx = await openState(browser, state);
      const file = `${state.id}.png`;
      await ctx.page.screenshot({
        path: join(out, file),
        clip: await clipFor(ctx.page, state),
        animations: state.animations === "allow" ? "allow" : "disabled",
        caret: "hide",
      });
      entry.file = file;
      if (ctx.errors.length) entry.pageErrors = ctx.errors;
      console.log(`ok   ${state.id}`);
    } catch (err) {
      entry.error = err.message.split("\n")[0];
      console.log(`FAIL ${state.id}: ${entry.error}`);
    } finally {
      await ctx?.context.close();
    }
    manifest.states.push(entry);
  }
} finally {
  manifest.finishedAt = new Date().toISOString();
  await writeFile(
    join(out, "manifest.json"),
    JSON.stringify(manifest, null, 2),
  );
  await browser.close();
}
console.log(out);
