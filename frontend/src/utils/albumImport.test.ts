import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ImportJob } from "@/api/types";

const jobs: ImportJob[] = [];
vi.mock("@/api/source", () => ({
  ApiError: class extends Error {},
  loginUrl: "/auth/login",
  source: { importJobs: async () => structuredClone(jobs) },
}));

// The player's message provider shows one message at a time (max 1): a new
// one replaces the old without telling it.
let shown: { content: string } | null = null;
const message = (content: string) => {
  const m = { content, destroy: () => void (shown === m && (shown = null)) };
  shown = m;
  return m;
};
const api = {
  loading: message,
  info: message,
  success: message,
  error: message,
  destroyAll: () => void (shown = null),
};

const job = (id: number, patch: Partial<ImportJob> = {}): ImportJob => ({
  id,
  albumId: id,
  title: `A${id}`,
  state: "running",
  total: 10,
  done: 0,
  created: 0,
  moved: 0,
  existing: 0,
  unplayable: 0,
  failed: [],
  error: "",
  ...patch,
});

describe("album import progress", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal("window", { $message: api, $dialog: {} });
    shown = null;
    jobs.length = 0;
  });
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.resetModules();
  });

  it("shares one message, comes back after another one, and reports each result", async () => {
    const { resumeAlbumImports } = await import("./albumImport");
    jobs.push(job(1, { done: 1 }), job(2, { state: "queued" }));
    await resumeAlbumImports();
    expect(shown?.content).toBe("正在导入《A1》 1/10，另有 1 张等待");

    window.$message.info("别的提示");
    expect(shown?.content).toBe("别的提示");
    jobs[0].done = 2;
    await vi.advanceTimersByTimeAsync(1600);
    expect(shown?.content).toBe("别的提示");
    await vi.advanceTimersByTimeAsync(3000);
    expect(shown?.content).toBe("正在导入《A1》 2/10，另有 1 张等待");

    jobs[0] = job(1, { state: "done", done: 10, created: 10 });
    jobs[1] = job(2, { done: 3 });
    await vi.advanceTimersByTimeAsync(1600);
    expect(shown?.content).toBe("《A1》导入完成：新增 10 首");
    await vi.advanceTimersByTimeAsync(3200);
    expect(shown?.content).toBe("正在导入《A2》 3/10");

    jobs[1] = job(2, { state: "done", done: 10, created: 9, existing: 1 });
    await vi.advanceTimersByTimeAsync(1600);
    expect(shown?.content).toBe("《A2》导入完成：新增 9 首，已在曲库 1 首");
  });
});
