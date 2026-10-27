// Compares Sonora's player with the SPlayer reference: starts a throwaway
// Vite dev server, captures Sonora (player.mjs) in every viewport and
// state the reference holds, compares pixel by pixel (diff.mjs) and judges
// the result against the differences Eric accepted. Run by
// `task test:visual`. See docs/phase/phase-3-player-shell.md, 3.6.
//
//   node scripts/visual/compare.mjs [--reference DIR] [--viewport WxH]...
//       [--only REGEX] [--port 14571 | --base URL]
//
// Local only: the reference comes from the SPlayer instance on this
// machine (scripts/baseline/splayer-player.mjs --sonora) and the baseline
// fixture from .local (scripts/fixtures/baseline.mjs); neither is in Git.
// Playwright comes from ../palace. Screenshots, diffs and a summary go to
// .local/visual/test-visual/, which every run clears first. Exits 1 when a
// state differs beyond what was accepted, is missing or has page errors.

import { spawn } from "node:child_process";
import { existsSync } from "node:fs";
import { readdir, readFile, rm, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { parseArgs } from "node:util";

const repo = resolve(dirname(fileURLToPath(import.meta.url)), "../..");

const { values: args } = parseArgs({
  options: {
    reference: {
      type: "string",
      default: ".local/baseline/splayer/2026-10-25-sonora",
    },
    viewport: { type: "string", multiple: true },
    only: { type: "string" },
    port: { type: "string", default: "14571" },
    base: { type: "string" },
  },
});

const reference = resolve(repo, args.reference);
const out = join(repo, ".local/visual/test-visual");
const vite = join(repo, "frontend/node_modules/vite/bin/vite.js");
const playwright =
  process.env.PLAYWRIGHT_TEST ??
  resolve(repo, "../palace/node_modules/@playwright/test/index.mjs");

// Differences Eric accepted (phase 3, 3.5 and 3.2). In the full player
// and settings states, pixels that differ by at most 2 are left out first:
// cover colour extraction drifts by one on either side, and SPlayer's
// comment layer moves a pixel or two by as much. A state then passes when
// nothing is left or when one matching rule bounds the pixels left and
// their largest channel difference.
const drift = {
  states: /^(full|set)-/,
  delta: 2,
  reason: "封面取色 ±1，或评论层造成的 1–2 个像素",
};
const narrow = /^(1280x800|1025x800|1024x800|990x800)$/;
// The equalizer button replaces SPlayer's one-item "其他控制" menu; its area
// (recorded by player.mjs as masks) is left out of the comparison.
const maskReason = "均衡器按钮改用 Eq 图标（Eric 确认，2026-10-27）";
const accepted = [
  {
    states: /^bar-paused-start$/,
    pixels: 400,
    delta: 255,
    reason: "参考图的采集步骤使“00:00 /”低 1 像素",
  },
  {
    states: /^bar-volume-hover$/,
    viewports: /^1440x1000$/,
    pixels: 2,
    delta: 9,
    reason: "进度条末端的亚像素位置",
  },
  {
    states: /^(bar-paused|bar-volume-hover)$/,
    viewports: narrow,
    pixels: 9,
    delta: 9,
    reason: "窄窗口下进度条末端与播放按钮边缘",
  },
  {
    states: /^full-record-blur$/,
    pixels: 30000,
    delta: 70,
    reason: "SPlayer 的透明评论层改变合成分层（唱片样式）",
  },
  {
    states: /^(full-cover-blur|full-yrc)$/,
    viewports: narrow,
    pixels: 160,
    delta: 31,
    reason: "SPlayer 的透明评论层改变合成分层（窄窗口的控制栏图标）",
  },
  {
    states: /^set-yrc-off$/,
    pixels: 45000,
    delta: 255,
    reason: "数据差异：参考实例改用网易另给的 LRC",
  },
];

function fail(message) {
  console.error(message);
  process.exit(2);
}

function run(script, scriptArgs) {
  return new Promise((done, reject) => {
    const child = spawn(process.execPath, [script, ...scriptArgs], {
      cwd: repo,
      stdio: "inherit",
    });
    child.on("error", reject);
    child.on("exit", (code) =>
      code === 0 ? done() : reject(new Error(`${script} exited with ${code}`)),
    );
  });
}

async function startVite(port) {
  const base = `http://127.0.0.1:${port}`;
  const server = spawn(
    process.execPath,
    [vite, "--host", "127.0.0.1", "--port", port, "--strictPort"],
    { cwd: join(repo, "frontend"), stdio: ["ignore", "ignore", "inherit"] },
  );
  let exited = false;
  server.on("exit", () => (exited = true));
  const deadline = Date.now() + 30_000;
  while (Date.now() < deadline) {
    if (exited)
      fail(`Vite exited before listening on ${base}; is the port in use?`);
    try {
      if ((await fetch(base)).ok) return { base, server };
    } catch {
      // not listening yet
    }
    await new Promise((r) => setTimeout(r, 250));
  }
  server.kill();
  fail(`Vite did not answer on ${base} within 30 s`);
}

const pngs = async (dir) =>
  (await readdir(dir))
    .filter((f) => f.endsWith(".png"))
    .map((f) => f.slice(0, -4));

const readJSON = async (file) =>
  existsSync(file) ? JSON.parse(await readFile(file, "utf8")) : null;

// Same key order on both sides before comparing the renderer's inputs.
const canonical = (value) =>
  JSON.stringify(value, (_key, v) =>
    v && typeof v === "object" && !Array.isArray(v)
      ? Object.fromEntries(
          Object.entries(v).sort(([a], [b]) => (a < b ? -1 : 1)),
        )
      : v,
  );

function judge(viewport, id, diff, refEntry, entry) {
  if (!entry) return { result: "FAIL", detail: "Sonora 没有记录这一状态" };
  if (entry.error) return { result: "FAIL", detail: entry.error };
  if (entry.pageErrors?.length)
    return { result: "FAIL", detail: `页面报错：${entry.pageErrors[0]}` };
  // The flowing background renders by wall-clock time, so these states
  // compare the renderer's props and the cover hash, not pixels.
  if (refEntry?.background) {
    return canonical(refEntry.background) === canonical(entry.background)
      ? { result: "same", detail: "流动背景的渲染输入相同" }
      : {
          result: "FAIL",
          detail: `流动背景的渲染输入不同：${canonical(entry.background)}`,
        };
  }
  if (!diff || diff.missing) return { result: "FAIL", detail: "缺少截图" };
  if (diff.size)
    return { result: "FAIL", detail: `尺寸 ${diff.size.join(" / ")}` };
  if (!diff.count && !diff.masked) return { result: "same", detail: "" };
  if (!diff.count)
    return {
      result: "accepted",
      detail: `均衡器按钮 ${diff.masked} px（${maskReason}）`,
    };
  const detail = `${diff.count} px，最大差值 ${diff.maxDelta}，区域 ${diff.box.join(",")}`;
  let left = Object.entries(diff.deltas).map(([d, n]) => [Number(d), n]);
  const reasons = diff.masked ? [maskReason] : [];
  if (drift.states.test(id) && left.some(([d]) => d <= drift.delta)) {
    left = left.filter(([d]) => d > drift.delta);
    reasons.push(drift.reason);
  }
  const count = left.reduce((sum, [, n]) => sum + n, 0);
  const maxDelta = Math.max(0, ...left.map(([d]) => d));
  if (count) {
    const rule = accepted.find(
      (r) =>
        r.states.test(id) &&
        (!r.viewports || r.viewports.test(viewport)) &&
        count <= r.pixels &&
        maxDelta <= r.delta,
    );
    if (!rule) {
      const rest = reasons.length
        ? `；去掉差值不超过 ${drift.delta} 的像素后还有 ${count} px，最大差值 ${maxDelta}`
        : "";
      return { result: "FAIL", detail: detail + rest };
    }
    reasons.push(rule.reason);
  }
  return { result: "accepted", detail: `${detail}（${reasons.join("；")}）` };
}

// Prerequisites, with what to do when one is missing.
if (!existsSync(reference))
  fail(
    `No reference at ${reference}. Capture it from the SPlayer instance with ` +
      "scripts/baseline/splayer-player.mjs --sonora (see docs/splayer-session.md).",
  );
if (!existsSync(join(repo, ".local/fixtures/baseline")))
  fail("No baseline fixture; run node scripts/fixtures/baseline.mjs first.");
if (!existsSync(playwright))
  fail(`No Playwright at ${playwright}; set PLAYWRIGHT_TEST to its index.mjs.`);
if (!args.base && !existsSync(vite)) fail("No Vite; run task install:web.");

const viewports =
  args.viewport ??
  (await readdir(reference, { withFileTypes: true }))
    .filter((d) => d.isDirectory() && /^\d+x\d+$/.test(d.name))
    .map((d) => d.name);
const only = args.only ? new RegExp(args.only) : null;

await rm(out, { recursive: true, force: true });
const { base, server } = args.base
  ? { base: args.base, server: null }
  : await startVite(args.port);
const stop = () => server?.kill();
process.on("SIGINT", () => {
  stop();
  process.exit(130);
});

const results = [];
try {
  for (const viewport of viewports) {
    const refDir = join(reference, viewport);
    const ids = (await pngs(refDir)).filter((id) => !only || only.test(id));
    if (!ids.length) continue;
    const pattern = `^(${ids.join("|")})$`;
    console.log(`\n== ${viewport}: ${ids.length} states`);
    await run("scripts/visual/player.mjs", [
      "--base",
      base,
      "--run",
      "test-visual",
      "--viewport",
      viewport,
      "--only",
      pattern,
    ]);
    const diffDir = join(out, `diff-${viewport}`);
    const captured = await readJSON(join(out, viewport, "manifest.json"));
    const masksFile = join(out, `masks-${viewport}.json`);
    await writeFile(
      masksFile,
      JSON.stringify(
        Object.fromEntries(
          (captured?.states ?? []).map((s) => [`${s.id}.png`, s.masks ?? []]),
        ),
      ),
    );
    await run("scripts/baseline/diff.mjs", [
      refDir,
      join(out, viewport),
      "--out",
      diffDir,
      "--masks",
      masksFile,
    ]);
    const report = (await readJSON(join(diffDir, "report.json"))) ?? [];
    const refManifest = await readJSON(join(refDir, "manifest.json"));
    const manifest = await readJSON(join(out, viewport, "manifest.json"));
    for (const id of ids) {
      const diff = report.find((r) => r.file === `${id}.png`);
      const refEntry = refManifest?.states.find((s) => s.id === id);
      const entry = manifest?.states.find((s) => s.id === id);
      results.push({
        viewport,
        id,
        ...judge(viewport, id, diff, refEntry, entry),
      });
    }
  }
} finally {
  stop();
}

console.log("\n== 结论");
for (const r of results) {
  const mark = { same: "一致", accepted: "例外", FAIL: "不符" }[r.result];
  console.log(
    `${mark}  ${r.viewport}  ${r.id}${r.detail ? `  ${r.detail}` : ""}`,
  );
}
const failed = results.filter((r) => r.result === "FAIL");
const count = (result) => results.filter((r) => r.result === result).length;
console.log(
  `\n${results.length} 个状态：${count("same")} 一致，${count("accepted")} 属于已接受的例外，${failed.length} 不符。` +
    `\n截图与差异图在 ${out}`,
);
await writeFile(
  join(out, "summary.json"),
  JSON.stringify({ reference, base, results }, null, 2),
);
process.exit(failed.length || !results.length ? 1 : 0);
