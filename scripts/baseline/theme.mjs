// Compares the theme SPlayer and Sonora generate from the same settings
// and covers: the CSS variables on <body>, the cover theme in the status
// store and the theme and overrides handed to Naive UI's config provider.
// See docs/phase/phase-3-player-shell.md (3.2).
//
//   node scripts/baseline/theme.mjs --target splayer [--run NAME]
//   node scripts/baseline/theme.mjs --target sonora [--run NAME]
//       [--sonora http://127.0.0.1:14559]
//   node scripts/baseline/theme.mjs --compare [--run NAME]
//
// Each case starts in a fresh BrowserContext with its settings in
// localStorage. Cover cases serve a fixed image from the page's own origin
// and run each project's own cover pipeline on it: SPlayer through its
// player controller's afterPlaySetup, Sonora by importing utils/color.ts
// from its Vite dev server. The covers are the baseline songs' covers from
// .local/baseline/splayer/2026-10-25/data plus two generated images.
// Results go to .local/baseline/splayer/<run>/theme/.

import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { isDeepStrictEqual, parseArgs } from "node:util";

import { prepareSPlayerSession } from "../splayer-session.mjs";

const repo = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const { values: args } = parseArgs({
  options: {
    target: { type: "string" },
    compare: { type: "boolean", default: false },
    run: { type: "string", default: "2026-10-25" },
    splayer: { type: "string", default: "http://127.0.0.1:25884" },
    sonora: { type: "string", default: "http://127.0.0.1:14559" },
    concurrency: { type: "string", default: "4" },
  },
});
const out = join(repo, ".local/baseline/splayer", args.run, "theme");
const coverData = join(repo, ".local/baseline/splayer/2026-10-25/data");

// --- cases ----------------------------------------------------------------

const themeColors = JSON.parse(
  await readFile(
    join(repo, "frontend/src/assets/data/themeColor.json"),
    "utf8",
  ),
);
const presets = Object.keys(themeColors).filter(
  (k) => k !== "custom" && k !== "solid",
);
const modes = ["light", "dark"];
const variants = [
  "primary",
  "secondary",
  "tertiary",
  "neutral",
  "neutralVariant",
  "error",
];
const covers = ["yoru", "guren", "gray", "gradient"];

const cases = [];
for (const themeMode of modes)
  for (const themeGlobalColor of [false, true]) {
    for (const themeColorType of [...presets, "solid"])
      cases.push({
        id: `${themeColorType}-${themeMode}-${themeGlobalColor ? "global" : "local"}`,
        settings: { themeMode, themeGlobalColor, themeColorType },
      });
    cases.push({
      id: `custom-${themeMode}-${themeGlobalColor ? "global" : "local"}`,
      settings: {
        themeMode,
        themeGlobalColor,
        themeColorType: "custom",
        themeCustomColor: "#3366ff",
      },
    });
    for (const cover of covers)
      cases.push({
        id: `cover-${cover}-${themeMode}-${themeGlobalColor ? "global" : "local"}`,
        settings: { themeMode, themeGlobalColor, themeFollowCover: true },
        cover,
      });
  }
for (const themeMode of modes)
  for (const themeVariant of variants)
    cases.push({
      id: `variant-${themeVariant}-${themeMode}`,
      settings: { themeMode, themeVariant, themeGlobalColor: true },
    });
cases.push({
  id: "cover-yoru-light-local-nofollow",
  settings: {
    themeMode: "light",
    themeFollowCover: true,
    playerFollowCoverColor: false,
  },
  cover: "yoru",
});
cases.push({
  id: "font-light",
  settings: { themeMode: "light", globalFont: "Noto Serif CJK SC" },
});

// --- page helpers -------------------------------------------------------------

const PINIA = `document.querySelector("#app").__vue_app__.config.globalProperties.$pinia`;

// read collects what the theme produced. The overrides are read from the
// outermost Naive UI ConfigProvider's props, which production builds keep.
const read = (pinia) => {
  const style = document.body.style;
  const vars = {};
  for (let i = 0; i < style.length; i++) {
    const name = style[i];
    if (name.startsWith("--")) vars[name] = style.getPropertyValue(name);
  }
  const walk = (vnode) => {
    if (!vnode) return null;
    if (vnode.component) {
      const c = vnode.component;
      if (c.type?.name === "ConfigProvider") return c.props;
      const found = walk(c.subTree);
      if (found) return found;
    }
    if (Array.isArray(vnode.children))
      for (const child of vnode.children) {
        const found = walk(child);
        if (found) return found;
      }
    return null;
  };
  // The mount container keeps the root vnode; production builds clear
  // app._instance.
  const config = walk(document.querySelector("#app")._vnode);
  const status = eval(pinia)._s.get("status");
  return JSON.parse(
    JSON.stringify({
      vars,
      songCoverTheme: status.songCoverTheme,
      theme: config?.theme?.name ?? null,
      themeOverrides: config?.themeOverrides ?? null,
    }),
  );
};

const coverReady = (pinia) =>
  !!eval(pinia)._s.get("status").songCoverTheme?.main;

// prepareCovers writes the two generated covers next to the results: a
// flat grey (SPlayer's monotonous branch) and a two-colour gradient.
async function prepareCovers(browser) {
  await mkdir(join(out, "covers"), { recursive: true });
  const page = await browser.newPage();
  const make = (kind) =>
    page.evaluate((kind) => {
      const canvas = document.createElement("canvas");
      canvas.width = canvas.height = 300;
      const ctx = canvas.getContext("2d");
      if (kind === "gray") {
        ctx.fillStyle = "#808080";
      } else {
        const g = ctx.createLinearGradient(0, 0, 300, 300);
        g.addColorStop(0, "#1e6fd9");
        g.addColorStop(1, "#f2a03d");
        ctx.fillStyle = g;
      }
      ctx.fillRect(0, 0, 300, 300);
      return canvas.toDataURL("image/png").split(",")[1];
    }, kind);
  for (const kind of ["gray", "gradient"])
    await writeFile(
      join(out, "covers", `${kind}.png`),
      Buffer.from(await make(kind), "base64"),
    );
  await page.close();
}

async function coverFile(name) {
  const file = ["yoru", "guren"].includes(name)
    ? join(coverData, `${name}-cover.jpg`)
    : join(out, "covers", `${name}.png`);
  return {
    body: await readFile(file),
    contentType: file.endsWith(".png") ? "image/png" : "image/jpeg",
  };
}

const targets = {
  splayer: {
    base: args.splayer,
    storageKey: "setting-store",
    // userAgreementVersion skips the first-run agreement; the prefetch
    // would ask NetEase for the next song.
    baseSettings: {
      schemaVersion: 12,
      userAgreementVersion: "v2.0",
      useNextPrefetch: false,
    },
    prepare: (context, base) => prepareSPlayerSession(context, base),
    path: "/#/",
    runCover: (page, url) =>
      page.evaluate(
        async ([pinia, url]) => {
          const music = eval(pinia)._s.get("music");
          music.playSong = {
            ...music.playSong,
            id: 1,
            type: "song",
            path: undefined,
            cover: url,
            coverSize: { s: url, m: url, l: url, xl: url },
          };
          await window.__SPLAYER_PLAYER_CONTROLLER__.afterPlaySetup(
            music.playSong,
          );
        },
        [PINIA, url],
      ),
  },
  sonora: {
    base: args.sonora,
    storageKey: "sonora-setting",
    baseSettings: {},
    prepare: async () => {},
    path: "/?fixture=library#/",
    runCover: (page, url) =>
      page.evaluate(async (url) => {
        const { getCoverColor } = await import("/src/utils/color.ts");
        await getCoverColor(url);
      }, url),
  },
};

async function runCase(browser, target, c) {
  const context = await browser.newContext({
    viewport: { width: 1440, height: 1000 },
    deviceScaleFactor: 1,
    locale: "zh-CN",
    timezoneId: "Asia/Shanghai",
    colorScheme: "light",
  });
  try {
    await target.prepare(context, target.base);
    const origin = new URL(target.base).origin;
    await context.addInitScript(
      ([origin, key, settings]) => {
        if (location.origin === origin)
          localStorage.setItem(key, JSON.stringify(settings));
      },
      [origin, target.storageKey, { ...target.baseSettings, ...c.settings }],
    );
    await context.route(`${origin}/__sonora-theme/**`, async (route) => {
      const name = new URL(route.request().url()).pathname.split("/").pop();
      await route.fulfill(await coverFile(name));
    });
    const page = await context.newPage();
    const errors = [];
    page.on("pageerror", (e) => errors.push(e.message));
    await page.goto(target.base + target.path);
    await page.waitForFunction(
      (pinia) => {
        try {
          return !!eval(pinia);
        } catch {
          return false;
        }
      },
      PINIA,
      { timeout: 30_000 },
    );
    await page.waitForTimeout(500);
    if (c.cover) {
      // material-color-utilities' Wsmeans quantizer seeds missing clusters
      // with Math.random, so the same cover gives colours one unit apart
      // between runs. Both sides draw from the same seeded sequence.
      await page.evaluate(() => {
        let seed = 0x2f6b9c1d;
        Math.random = () => {
          seed = (seed + 0x6d2b79f5) | 0;
          let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
          t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
          return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
        };
      });
      await target.runCover(page, `${origin}/__sonora-theme/${c.cover}`);
      await page.waitForFunction(coverReady, PINIA, { timeout: 15_000 });
    }
    // Provider applies the theme in a watcher after the store changes.
    await page.waitForTimeout(500);
    return { id: c.id, ...(await page.evaluate(read, PINIA)), errors };
  } finally {
    await context.close();
  }
}

async function capture(name) {
  const target = targets[name];
  if (!target) throw new Error(`unknown target ${name}`);
  const { chromium } = await import(
    pathToFileURL(
      process.env.PLAYWRIGHT_TEST ??
        resolve(repo, "../palace/node_modules/@playwright/test/index.mjs"),
    ).href
  );
  const browser = await chromium.launch();
  await prepareCovers(browser);
  const results = [];
  const queue = [...cases];
  const worker = async () => {
    for (let c; (c = queue.shift());) {
      results.push(await runCase(browser, target, c));
      process.stdout.write(".");
    }
  };
  await Promise.all(Array.from({ length: Number(args.concurrency) }, worker));
  process.stdout.write("\n");
  await browser.close();
  results.sort((a, b) => a.id.localeCompare(b.id));
  await writeFile(
    join(out, `${name}.json`),
    JSON.stringify(
      {
        target: name,
        base: target.base,
        at: new Date().toISOString(),
        results,
      },
      null,
      2,
    ),
  );
  const failed = results.filter((r) => r.errors.length);
  console.log(
    `${name}: ${results.length} cases, ${failed.length} with page errors`,
  );
}

async function compare() {
  const load = async (name) =>
    JSON.parse(await readFile(join(out, `${name}.json`), "utf8")).results;
  const [a, b] = await Promise.all([load("splayer"), load("sonora")]);
  const byId = new Map(b.map((r) => [r.id, r]));
  let differ = 0;
  for (const ref of a) {
    const got = byId.get(ref.id);
    const parts = ["vars", "songCoverTheme", "theme", "themeOverrides"].filter(
      (k) => !got || !isDeepStrictEqual(ref[k], got[k]),
    );
    if (parts.length) {
      differ++;
      console.log(`${ref.id}: ${parts.join(", ")} differ`);
    }
  }
  console.log(
    `${a.length} cases, ${a.length - differ} identical, ${differ} different`,
  );
  process.exitCode = differ ? 1 : 0;
}

if (args.compare) await compare();
else await capture(args.target);
