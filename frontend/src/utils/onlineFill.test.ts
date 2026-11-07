import { describe, expect, it, vi } from "vitest";
import { artistFillMessage } from "./onlineFill";

vi.mock("@/api/source", () => ({
  source: {},
  ApiError: class extends Error {},
  loginUrl: "/auth/login",
}));
vi.mock("@/core/player/PlayerController", () => ({
  usePlayerController: () => ({}),
}));

describe("艺术家补全的提示", () => {
  it("说明补全了什么，以及网易云上也没有什么", () => {
    expect(
      artistFillMessage(["image", "description"], ["image", "description"]),
    ).toBe("已补全图片和简介");
    expect(artistFillMessage(["description"], ["description"])).toBe(
      "已补全简介",
    );
    expect(artistFillMessage(["description"], ["image", "description"])).toBe(
      "已补全简介，网易云上没有图片",
    );
    expect(artistFillMessage([], ["image"])).toBe(
      "网易云上也没有可以补全的资料",
    );
  });
});
