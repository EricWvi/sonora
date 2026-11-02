import { describe, expect, it } from "vitest";
import { gridWindow } from "./useVirtualGrid";

// 行高 100、间距 20：第 i 行从 120·i 开始
describe("gridWindow", () => {
  it("renders the rows in view plus the buffer", () => {
    const w = gridWindow([], 100, 20, 50, 1000, 1500, 2);
    // 视口 1000–1500 覆盖第 8（960–1060）到第 12 行（1440–1540）
    expect([w.start, w.end]).toEqual([6, 14]);
    expect(w.padTop).toBe(6 * 120);
    // 总高 50·120−20；第 14 行底部在 14·120+100
    expect(w.padBottom).toBe(50 * 120 - 20 - (14 * 120 + 100));
  });

  it("keeps the total height when rows are measured", () => {
    const heights = [200, 200, undefined, 150];
    const w = gridWindow(heights, 100, 20, 6, 0, 10, 0);
    expect([w.start, w.end]).toEqual([0, 0]);
    const total = 200 + 200 + 100 + 150 + 100 + 100 + 5 * 20;
    expect(w.padTop + 200 + w.padBottom).toBe(total);
  });

  it("stays at the last row when scrolled past the end", () => {
    const w = gridWindow([], 100, 20, 3, 5000, 5500, 1);
    expect([w.start, w.end]).toEqual([1, 2]);
    expect(w.padBottom).toBe(0);
  });

  it("is empty without rows", () => {
    expect(gridWindow([], 100, 20, 0, 0, 500, 2)).toEqual({
      start: 0,
      end: -1,
      padTop: 0,
      padBottom: 0,
    });
  });
});
