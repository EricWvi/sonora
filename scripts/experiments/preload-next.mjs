// Preload experiment for phase 5 ("下一首歌曲预载"): does fetching the start
// of the next track ahead of time make the player's element start it sooner?
//
// It starts a throwaway Sonora dev instance with empty data, uploads three
// long MP3s (different files, so no case can reuse another's data), then
// switches one audio element (crossOrigin "anonymous", like
// AudioElementPlayer) to each track under a throttled network and measures
// the time from setting src to loadedmetadata and canplay:
//
//   cold     no preload
//   fetch    fetch(url, Range: bytes=0-3MiB) read to the end first
//   element  a hidden Audio element (same crossOrigin, preload --preload,
//            auto by default) on the URL first, kept alive until the switch
//
// Network requests for the track (Range, status, cache) are recorded too.
// Results go to .local/experiments/preload-next/<run>.json.
//
//   node scripts/experiments/preload-next.mjs --sample A.mp3 --sample B.mp3
//       --sample C.mp3 [--mbps 8] [--wait 8] [--preload auto|metadata]
//       [--channel chrome]

import { execFileSync, spawn } from "node:child_process";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import { basename, dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { parseArgs } from "node:util";

const repo = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const { chromium } = await import(
  pathToFileURL(
    process.env.PLAYWRIGHT_TEST ??
      resolve(repo, "../palace/node_modules/@playwright/test/index.mjs"),
  ).href
);

const { values: args } = parseArgs({
  options: {
    channel: { type: "string" },
    sample: { type: "string", multiple: true, default: [] },
    port: { type: "string", default: "14573" },
    mbps: { type: "string", default: "8" },
    // Seconds the preload gets before the switch (the current track playing).
    wait: { type: "string", default: "8" },
    // preload attribute of the hidden element in the "element" case.
    preload: { type: "string", default: "auto" },
    run: {
      type: "string",
      default: new Date().toISOString().replace(/[:.]/g, "-"),
    },
  },
});
if (args.sample.length !== 3) {
  console.error("need three --sample files (long MP3s)");
  process.exit(2);
}
const origin = `http://127.0.0.1:${args.port}`;
const cases = ["cold", "fetch", "element"];

// --- throwaway Sonora instance ---------------------------------------------

const work = await mkdtemp(join(os.tmpdir(), "sonora-preload-"));
let server;
async function startServer() {
  const bin = join(work, "sonora");
  execFileSync("go", ["build", "-tags", "dev", "-o", bin, "./cmd/sonora"], {
    cwd: join(repo, "backend"),
  });
  const config = join(work, "config.yaml");
  await writeFile(
    config,
    `listen: "127.0.0.1:${args.port}"\ndata_dir: "${join(work, "data")}"\nmedia_dir: "${join(work, "media")}"\ndev: true\n`,
  );
  server = spawn(bin, ["-config", config], {
    stdio: ["ignore", "ignore", "pipe"],
  });
  for (let i = 0; i < 50; i++) {
    try {
      if ((await fetch(origin + "/api/version")).ok) return;
    } catch {}
    await new Promise((r) => setTimeout(r, 200));
  }
  throw new Error("Sonora did not start");
}

async function upload(paths) {
  const login = await fetch(origin + "/admin/api/v1/login", {
    method: "POST",
    headers: { "Content-Type": "application/json", Origin: origin },
    body: JSON.stringify({ username: "admin", password: "admin" }),
  });
  const cookie = login.headers.get("set-cookie").split(";")[0];
  const tracks = [];
  for (const path of paths) {
    const form = new FormData();
    form.append("file", new Blob([await readFile(path)]), basename(path));
    const res = await fetch(origin + "/admin/api/v1/tracks", {
      method: "POST",
      headers: { Cookie: cookie, Origin: origin },
      body: form,
    });
    const body = await res.json();
    if (!res.ok) throw new Error(`upload ${path}: ${res.status} ${body.error}`);
    tracks.push({ id: body.track.id, size: body.track.size });
  }
  return tracks;
}

// --- browser ------------------------------------------------------------------

let browser;
try {
  await startServer();
  const tracks = await upload(args.sample);
  browser = await chromium.launch({ channel: args.channel });
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.goto(origin + "/api/version");
  const cdp = await context.newCDPSession(page);
  await cdp.send("Network.enable");
  await cdp.send("Network.emulateNetworkConditions", {
    offline: false,
    latency: 20,
    downloadThroughput: ((Number(args.mbps) * 1_000_000) / 8) | 0,
    uploadThroughput: ((Number(args.mbps) * 1_000_000) / 8) | 0,
  });
  const requests = [];
  const byId = new Map();
  cdp.on("Network.requestWillBeSent", (e) => {
    if (!e.request.url.includes("/audio")) return;
    const r = {
      t: e.timestamp,
      url: e.request.url.replace(origin, ""),
      range: e.request.headers.Range ?? e.request.headers.range ?? null,
      bytes: 0,
    };
    byId.set(e.requestId, r);
    requests.push(r);
  });
  cdp.on("Network.responseReceived", (e) => {
    const r = byId.get(e.requestId);
    if (!r) return;
    r.status = e.response.status;
    r.fromDiskCache = e.response.fromDiskCache;
    r.contentRange = e.response.headers["Content-Range"] ?? null;
  });
  cdp.on("Network.dataReceived", (e) => {
    const r = byId.get(e.requestId);
    if (r) r.bytes += e.encodedDataLength || e.dataLength;
  });
  // The element is created once and reused, like AudioElementPlayer.
  await page.evaluate(() => {
    const el = new Audio();
    el.crossOrigin = "anonymous";
    window.main = el;
    window.kept = [];
  });

  const results = [];
  for (const [i, name] of cases.entries()) {
    const url = `/api/v1/tracks/${tracks[i].id}/audio`;
    const before = requests.length;
    const result = await page.evaluate(
      async ({ name, url, wait, preload }) => {
        const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
        const preloadStart = performance.now();
        let preloaded = null;
        if (name === "fetch") {
          const res = await fetch(url, {
            headers: { Range: `bytes=0-${3 * 1024 * 1024 - 1}` },
          });
          preloaded = (await res.arrayBuffer()).byteLength;
        } else if (name === "element") {
          const el = new Audio();
          el.crossOrigin = "anonymous";
          el.preload = preload;
          el.src = url;
          window.kept.push(el);
        }
        const spent = performance.now() - preloadStart;
        if (name !== "cold") await sleep(Math.max(0, wait * 1000 - spent));
        const el = window.main;
        const start = performance.now();
        const at = {};
        const done = new Promise((yes) => {
          el.addEventListener(
            "loadedmetadata",
            () => (at.loadedmetadata = performance.now() - start),
            { once: true },
          );
          el.addEventListener(
            "canplay",
            () => {
              at.canplay = performance.now() - start;
              yes();
            },
            { once: true },
          );
          el.addEventListener(
            "error",
            () => {
              at.error = el.error?.code;
              yes();
            },
            { once: true },
          );
        });
        el.src = url;
        await Promise.race([done, sleep(60_000)]);
        el.pause();
        return {
          name,
          preloadMs: Math.round(spent),
          preloadedBytes: preloaded,
          ...Object.fromEntries(
            Object.entries(at).map(([k, v]) => [k, Math.round(v)]),
          ),
        };
      },
      { name, url, wait: Number(args.wait), preload: args.preload },
    );
    result.size = tracks[i].size;
    result.requests = requests.slice(before);
    results.push(result);
    console.log(
      `${name}: loadedmetadata ${result.loadedmetadata} ms, canplay ${result.canplay} ms` +
        (result.error ? `, error ${result.error}` : ""),
    );
  }

  const out = join(repo, ".local/experiments/preload-next");
  await mkdir(out, { recursive: true });
  const report = {
    run: args.run,
    browser: browser.version(),
    channel: args.channel ?? "chromium",
    throttle: `${args.mbps} Mbit/s, 20 ms latency`,
    waitSeconds: Number(args.wait),
    preload: args.preload,
    results,
  };
  await writeFile(
    join(out, `${args.run}.json`),
    JSON.stringify(report, null, 2),
  );
  console.log(`results: ${join(out, `${args.run}.json`)}`);
} finally {
  await browser?.close();
  server?.kill();
  await rm(work, { recursive: true, force: true });
}
