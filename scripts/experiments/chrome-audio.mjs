// Chrome audio experiment for phase 2 (work package 0.3): can the player
// play what the admin uploads, straight from the Go service, through the
// Web Audio chain SPlayer uses?
//
// It starts a throwaway Sonora dev instance with empty data, uploads the
// samples through the admin API, then in the browser checks per file:
// canPlayType, metadata, decodeAudioData, playback through
// MediaElementSource → 10 peaking filters → analyser → destination
// (SPlayer's AudioElementPlayer and AudioEffectManager), seeking and the
// Range requests behind it. Separately it checks the autoplay policy and
// setSinkId. Results go to .local/experiments/chrome-audio/<run>.json.
//
//   node scripts/experiments/chrome-audio.mjs [--channel chrome] [--sample FILE]...
//       [--damage FILE]
//
// --sample adds a file (use long ones: the test samples are a few seconds
// and load in one request). --damage adds a copy of FILE with 64 KiB in
// the middle zeroed. Without --channel the Playwright Chromium runs; it
// lacks proprietary codecs such as AAC, so M4A results need --channel
// chrome.

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
    damage: { type: "string" },
    port: { type: "string", default: "14572" },
    // SPlayer leaves the element's preload at Chrome's default, auto.
    preload: { type: "string", default: "auto" },
    mbps: { type: "string", default: "2" },
    run: {
      type: "string",
      default: new Date().toISOString().replace(/[:.]/g, "-"),
    },
  },
});
const origin = `http://127.0.0.1:${args.port}`;
const testdata = join(repo, "backend/internal/media/testdata");
const samples = [
  ...[
    "tagged.mp3",
    "vbr.mp3",
    "untagged.mp3",
    "tagged.flac",
    "tagged.m4a",
    "tagged.ogg",
    "tagged.opus",
  ].map((f) => join(testdata, f)),
  ...args.sample.map((f) => resolve(f)),
];

// --- throwaway Sonora instance ---------------------------------------------

const work = await mkdtemp(join(os.tmpdir(), "sonora-audio-"));
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

async function upload(files) {
  const login = await fetch(origin + "/admin/api/v1/login", {
    method: "POST",
    headers: { "Content-Type": "application/json", Origin: origin },
    body: JSON.stringify({ username: "admin", password: "admin" }),
  });
  const cookie = login.headers.get("set-cookie").split(";")[0];
  const tracks = [];
  for (const { path, name } of files) {
    const form = new FormData();
    form.append("file", new Blob([await readFile(path)]), name);
    const res = await fetch(origin + "/admin/api/v1/tracks", {
      method: "POST",
      headers: { Cookie: cookie, Origin: origin },
      body: form,
    });
    const body = await res.json();
    if (!res.ok) {
      tracks.push({
        name,
        rejected: `${res.status} ${body.code}: ${body.error}`,
      });
      continue;
    }
    const t = body.track;
    tracks.push({
      name,
      id: t.id,
      format: t.format,
      durationMs: t.durationMs,
      size: t.size,
    });
  }
  return tracks;
}

// damaged copies a file with 64 KiB in the middle zeroed. The headers stay
// intact, so it passes ingest like a file damaged after tagging would.
async function damaged(path) {
  const data = Buffer.from(await readFile(path));
  const at = Math.floor(data.length / 2);
  data.fill(0, at, at + 64 * 1024);
  const out = join(work, "damaged-" + basename(path));
  await writeFile(out, data);
  return { path: out, at: at / data.length };
}

// --- browser checks ------------------------------------------------------------

// checkTrack plays the file through the element and SPlayer's Web Audio
// chain, then seeks. The page has been clicked, so the AudioContext may
// start.
async function checkTrack(page, track, contentType) {
  return page.evaluate(
    async ({ id, contentType, expectedMs, preload }) => {
      const url = `/api/v1/tracks/${id}/audio`;
      const r = { canPlayType: new Audio().canPlayType(contentType) };
      const timeout = (ms, what) =>
        new Promise((_, no) =>
          setTimeout(() => no(new Error(`timeout: ${what}`)), ms),
        );
      const once = (el, ok, bad = "error") =>
        new Promise((yes, no) => {
          el.addEventListener(ok, () => yes(), { once: true });
          el.addEventListener(
            bad,
            () =>
              no(
                new Error(
                  `${bad}: ${el.error?.code} ${el.error?.message ?? ""}`,
                ),
              ),
            {
              once: true,
            },
          );
        });

      // Element playback through SPlayer's chain.
      const audio = new Audio();
      audio.crossOrigin = "anonymous";
      audio.preload = preload;
      const loadStart = performance.now();
      audio.src = url;
      const ctx = new AudioContext();
      const source = ctx.createMediaElementSource(audio);
      let node = source;
      for (const f of [31, 62, 125, 250, 500, 1000, 2000, 4000, 8000, 16000]) {
        const filter = ctx.createBiquadFilter();
        filter.type = "peaking";
        filter.frequency.value = f;
        filter.Q.value = 1;
        filter.gain.value = 3;
        node.connect(filter);
        node = filter;
      }
      const analyser = ctx.createAnalyser();
      node.connect(analyser);
      analyser.connect(ctx.destination);
      const level = () => {
        const d = new Float32Array(analyser.fftSize);
        analyser.getFloatTimeDomainData(d);
        return Math.sqrt(d.reduce((s, v) => s + v * v, 0) / d.length);
      };
      try {
        await Promise.race([
          once(audio, "loadedmetadata"),
          timeout(60_000, "metadata"),
        ]);
        r.metadata = {
          ok: true,
          afterMs: Math.round(performance.now() - loadStart),
          durationMs: Math.round(audio.duration * 1000),
          expectedMs,
        };
        await ctx.resume();
        await audio.play();
        await new Promise((s) => setTimeout(s, 1200));
        r.play = {
          ok: true,
          contextState: ctx.state,
          rms: +level().toFixed(5),
          currentTime: +audio.currentTime.toFixed(2),
        };

        // Seek to 60 % and check playback carries on from there.
        const target = audio.duration * 0.6;
        const t0 = performance.now();
        audio.currentTime = target;
        await Promise.race([once(audio, "seeked"), timeout(10_000, "seeked")]);
        const seekMs = Math.round(performance.now() - t0);
        await new Promise((s) => setTimeout(s, 1000));
        r.seek = {
          ok: audio.currentTime > target,
          targetS: +target.toFixed(2),
          after1sS: +audio.currentTime.toFixed(2),
          seekMs,
          rms: +level().toFixed(5),
        };
      } catch (e) {
        r.error = e.message;
      }
      r.mediaError = audio.error
        ? { code: audio.error.code, message: audio.error.message }
        : null;
      audio.pause();
      await ctx.close();
      return r;
    },
    {
      id: track.id,
      contentType,
      expectedMs: track.durationMs,
      preload: args.preload,
    },
  );
}

// checkDecode decodes the whole file with Web Audio, as a fallback
// player or a waveform view would.
async function checkDecode(page, track) {
  return page.evaluate(async (id) => {
    const url = `/api/v1/tracks/${id}/audio`;
    const r = {};
    // Web Audio decode of the whole file.
    try {
      const bytes = await (await fetch(url)).arrayBuffer();
      const ctx = new OfflineAudioContext(2, 44100, 44100);
      const buf = await ctx.decodeAudioData(bytes);
      r.decode = {
        ok: true,
        durationMs: Math.round(buf.duration * 1000),
        sampleRate: buf.sampleRate,
        channels: buf.numberOfChannels,
      };
    } catch (e) {
      r.decode = { ok: false, error: `${e.name}: ${e.message}` };
    }

    return r.decode;
  }, track.id);
}

// checkDamaged plays across the zeroed region and reports what happens.
async function checkDamaged(page, track, fraction) {
  return page.evaluate(
    async ({ id, fraction }) => {
      const audio = new Audio(`/api/v1/tracks/${id}/audio`);
      const events = [];
      for (const e of ["stalled", "waiting", "error", "ended", "seeked"]) {
        audio.addEventListener(e, () =>
          events.push(`${e}@${audio.currentTime.toFixed(2)}`),
        );
      }
      await new Promise((r) =>
        audio.addEventListener("loadedmetadata", r, { once: true }),
      );
      const start = Math.max(0, audio.duration * fraction - 3);
      audio.currentTime = start;
      await audio.play().catch((e) => events.push("play: " + e.name));
      await new Promise((r) => setTimeout(r, 7000));
      const r = {
        startS: +start.toFixed(2),
        damageAtS: +(audio.duration * fraction).toFixed(2),
        after7sS: +audio.currentTime.toFixed(2),
        paused: audio.paused,
        error: audio.error
          ? `${audio.error.code} ${audio.error.message}`
          : null,
        events,
      };
      audio.pause();
      return r;
    },
    { id: track.id, fraction },
  );
}

// checkPolicy loads a same-origin page whose own script tries to start
// audio before and after a click. page.evaluate counts as a user gesture,
// so the attempts must come from the page itself.
const policyPage = `<!doctype html><title>policy</title><script>
  const ctx = new AudioContext();
  window.__ctx = ctx;
  const attempt = async () => {
    const audio = new Audio("/api/v1/tracks/1/audio");
    try {
      await audio.play();
      audio.pause();
      return "allowed";
    } catch (e) {
      return e.name;
    }
  };
  (async () => {
    window.__before = {
      contextState: ctx.state,
      elementPlay: await attempt(),
      userActivation: navigator.userActivation.hasBeenActive,
    };
  })();
  addEventListener("click", async () => {
    const play = attempt();
    await ctx.resume();
    window.__after = {
      contextState: ctx.state,
      elementPlay: await play,
      userActivation: navigator.userActivation.hasBeenActive,
    };
  });
</script>`;

async function checkPolicy(browser) {
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.route(origin + "/__policy", (route) =>
    route.fulfill({ contentType: "text/html", body: policyPage }),
  );
  await page.goto(origin + "/__policy");
  await page.waitForFunction(() => window.__before);
  const before = await page.evaluate(() => window.__before);
  await page.mouse.click(10, 10);
  await page.waitForFunction(() => window.__after);
  const after = await page.evaluate(() => window.__after);
  const sink = await page.evaluate(async () => {
    const r = {
      elementSetSinkId:
        typeof HTMLMediaElement.prototype.setSinkId === "function",
      contextSetSinkId: typeof AudioContext.prototype.setSinkId === "function",
    };
    const devices = await navigator.mediaDevices.enumerateDevices();
    r.audioOutputs = devices
      .filter((d) => d.kind === "audiooutput")
      .map((d) => ({ id: d.deviceId ? "set" : "", label: d.label }));
    try {
      await window.__ctx.setSinkId("");
      r.contextDefault = "ok";
    } catch (e) {
      r.contextDefault = `${e.name}: ${e.message}`;
    }
    try {
      await window.__ctx.setSinkId({ type: "none" });
      r.contextNone = "ok";
    } catch (e) {
      r.contextNone = `${e.name}: ${e.message}`;
    }
    try {
      await window.__ctx.setSinkId("no-such-device");
      r.contextUnknown = "ok";
    } catch (e) {
      r.contextUnknown = `${e.name}: ${e.message}`;
    }
    return r;
  });
  await context.close();
  return { beforeGesture: before, afterGesture: after, setSinkId: sink };
}

// --- run ------------------------------------------------------------------------

const result = { run: args.run, startedAt: new Date().toISOString() };
let browser;
try {
  await startServer();
  const files = samples.map((path) => ({ path, name: basename(path) }));
  let damage;
  if (args.damage) {
    damage = await damaged(resolve(args.damage));
    files.push({ path: damage.path, name: basename(damage.path) });
  }
  files.push({
    path: join(testdata, "truncated.flac"),
    name: "truncated.flac",
  });
  const tracks = await upload(files);

  browser = await chromium.launch({
    headless: true,
    ...(args.channel ? { channel: args.channel } : {}),
  });
  result.environment = {
    os: `${os.type()} ${os.release()} ${os.arch()}`,
    browser: `${args.channel ?? "chromium"} ${browser.version()} (headless, Playwright)`,
    preload: args.preload,
    throttle: `${args.mbps} Mbit/s, 20 ms latency for element playback`,
  };
  // Chrome's default for a site without media engagement: audio needs a
  // user activation. Headless Chromium otherwise allows it.
  const strict = await chromium.launch({
    headless: true,
    args: ["--autoplay-policy=document-user-activation-required"],
    ...(args.channel ? { channel: args.channel } : {}),
  });
  try {
    result.policy = {
      default: await checkPolicy(browser),
      userActivationRequired: await checkPolicy(strict),
    };
  } finally {
    await strict.close();
  }

  const context = await browser.newContext();
  const page = await context.newPage();
  const requests = [];
  page.on("response", (res) => {
    const m = res.url().match(/\/tracks\/(\d+)\/audio/);
    if (m) {
      requests.push({
        id: Number(m[1]),
        range: res.request().headers().range ?? null,
        status: res.status(),
        type: res.headers()["content-type"],
      });
    }
  });
  await page.goto(origin + "/api/version");
  await page.mouse.click(10, 10); // user activation for AudioContext and play()
  const cdp = await context.newCDPSession(page);
  await cdp.send("Network.enable");
  await cdp.send("Network.setCacheDisabled", { cacheDisabled: true });
  const throttled = {
    offline: false,
    latency: 20,
    downloadThroughput: ((Number(args.mbps) * 1_000_000) / 8) | 0,
    uploadThroughput: -1,
  };
  const unthrottled = {
    offline: false,
    latency: 0,
    downloadThroughput: -1,
    uploadThroughput: -1,
  };

  const contentTypes = {
    mp3: "audio/mpeg",
    flac: "audio/flac",
    m4a: "audio/mp4",
    ogg: "audio/ogg",
    opus: "audio/ogg; codecs=opus",
  };
  result.tracks = [];
  for (const t of tracks) {
    if (t.rejected) {
      result.tracks.push(t);
      console.log(`${t.name}: rejected at upload (${t.rejected})`);
      continue;
    }
    const isDamaged = damage && t.name === basename(damage.path);
    // The element streams at --mbps (2 Mbit/s), so a long file is still loading
    // when it seeks and the seek has to fetch a new range.
    await cdp.send("Network.emulateNetworkConditions", throttled);
    const checks = isDamaged
      ? { damaged: await checkDamaged(page, t, damage.at) }
      : await checkTrack(page, t, contentTypes[t.format]);
    await cdp.send("Network.emulateNetworkConditions", unthrottled);
    if (!isDamaged) checks.decode = await checkDecode(page, t);
    const entry = {
      ...t,
      ...checks,
      requests: requests.filter((r) => r.id === t.id),
    };
    result.tracks.push(entry);
    console.log(`${t.name}: ${JSON.stringify(checks)}`);
    console.log(
      `  requests: ${entry.requests.map((r) => `${r.range ?? "-"}→${r.status}`).join(", ")}`,
    );
  }
  await context.close();
} finally {
  await browser?.close();
  server?.kill();
  await rm(work, { recursive: true, force: true });
  result.finishedAt = new Date().toISOString();
  const outDir = join(repo, ".local/experiments/chrome-audio");
  await mkdir(outDir, { recursive: true });
  const out = join(outDir, `${args.run}.json`);
  await writeFile(out, JSON.stringify(result, null, 2));
  console.log(out);
}
