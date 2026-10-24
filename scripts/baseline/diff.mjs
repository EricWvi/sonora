// Compares the screenshots of two baseline runs pixel by pixel and writes
// a diff image for every pair that differs. Used to measure the noise of
// repeated captures and, later, to compare Sonora with SPlayer.
//
//   node scripts/baseline/diff.mjs <dir A> <dir B> [--out DIR] [--threshold N]
//
// A pixel differs when any channel differs by more than threshold (0–255,
// default 0). The report lists, per file, the differing pixels and the
// bounding box they fall in. Decoding and comparing run in Chromium's
// canvas, so no image library is needed.

import { mkdir, readdir, readFile, writeFile } from "node:fs/promises";
import { basename, join, resolve } from "node:path";
import { dirname } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { parseArgs } from "node:util";

const repo = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const { chromium } = await import(
  pathToFileURL(
    process.env.PLAYWRIGHT_TEST ??
      resolve(repo, "../palace/node_modules/@playwright/test/index.mjs"),
  ).href
);

const { values: opts, positionals } = parseArgs({
  allowPositionals: true,
  options: {
    out: { type: "string" },
    threshold: { type: "string", default: "0" },
  },
});
if (positionals.length !== 2) {
  console.error("usage: diff.mjs <dir A> <dir B> [--out DIR] [--threshold N]");
  process.exit(2);
}
const [dirA, dirB] = positionals.map((p) => resolve(p));
const threshold = Number(opts.threshold);
const outDir = opts.out && resolve(opts.out);
if (outDir) await mkdir(outDir, { recursive: true });

const files = (await readdir(dirA)).filter((f) => f.endsWith(".png")).sort();
const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
const report = [];
try {
  for (const file of files) {
    let b;
    try {
      b = await readFile(join(dirB, file));
    } catch {
      report.push({ file, missing: true });
      continue;
    }
    const a = await readFile(join(dirA, file));
    const result = await page.evaluate(
      async ([a, b, threshold]) => {
        const load = async (b64) => {
          const img = new Image();
          img.src = "data:image/png;base64," + b64;
          await img.decode();
          return img;
        };
        const [ia, ib] = await Promise.all([load(a), load(b)]);
        if (ia.width !== ib.width || ia.height !== ib.height) {
          return {
            size: [`${ia.width}x${ia.height}`, `${ib.width}x${ib.height}`],
          };
        }
        const w = ia.width;
        const h = ia.height;
        const pixels = (img) => {
          const c = new OffscreenCanvas(w, h);
          const ctx = c.getContext("2d");
          ctx.drawImage(img, 0, 0);
          return ctx.getImageData(0, 0, w, h).data;
        };
        const pa = pixels(ia);
        const pb = pixels(ib);
        const diff = new OffscreenCanvas(w, h);
        const dctx = diff.getContext("2d");
        const out = dctx.createImageData(w, h);
        let count = 0;
        let maxDelta = 0;
        let box = null;
        for (let i = 0; i < pa.length; i += 4) {
          const d = Math.max(
            Math.abs(pa[i] - pb[i]),
            Math.abs(pa[i + 1] - pb[i + 1]),
            Math.abs(pa[i + 2] - pb[i + 2]),
            Math.abs(pa[i + 3] - pb[i + 3]),
          );
          // Faded copy of A, with differing pixels in red.
          const grey = (pa[i] + pa[i + 1] + pa[i + 2]) / 3;
          out.data[i] = out.data[i + 1] = out.data[i + 2] = 200 + grey / 5;
          out.data[i + 3] = 255;
          if (d > threshold) {
            count++;
            maxDelta = Math.max(maxDelta, d);
            const p = i / 4;
            const x = p % w;
            const y = Math.floor(p / w);
            box = box
              ? [
                  Math.min(box[0], x),
                  Math.min(box[1], y),
                  Math.max(box[2], x),
                  Math.max(box[3], y),
                ]
              : [x, y, x, y];
            out.data[i] = 255;
            out.data[i + 1] = out.data[i + 2] = 0;
          }
        }
        let png = null;
        if (count) {
          dctx.putImageData(out, 0, 0);
          const blob = await diff.convertToBlob({ type: "image/png" });
          const bytes = new Uint8Array(await blob.arrayBuffer());
          let s = "";
          for (let i = 0; i < bytes.length; i += 0x8000)
            s += String.fromCharCode(...bytes.subarray(i, i + 0x8000));
          png = btoa(s);
        }
        return { pixels: w * h, count, maxDelta, box, png };
      },
      [a.toString("base64"), b.toString("base64"), threshold],
    );
    const { png, ...entry } = result;
    if (png && outDir) {
      await writeFile(
        join(outDir, basename(file, ".png") + "-diff.png"),
        Buffer.from(png, "base64"),
      );
    }
    report.push({ file, ...entry });
  }
} finally {
  await browser.close();
}

for (const r of report) {
  if (r.missing) console.log(`${r.file}: missing in B`);
  else if (r.size) console.log(`${r.file}: size ${r.size[0]} vs ${r.size[1]}`);
  else if (!r.count) console.log(`${r.file}: identical`);
  else {
    const pct = ((100 * r.count) / r.pixels).toFixed(3);
    const [x0, y0, x1, y1] = r.box;
    console.log(
      `${r.file}: ${r.count} px (${pct}%), max Δ ${r.maxDelta}, box ${x0},${y0}–${x1},${y1}`,
    );
  }
}
if (outDir)
  await writeFile(join(outDir, "report.json"), JSON.stringify(report, null, 2));
