import { describe, expect, it, vi } from "vitest";
import type { ShortcutList } from "@/stores/shortcut";
import { defaultShortcutState } from "@/stores/shortcut";
import { matchesShortcut, matchedShortcuts } from "./useShortcut";

// composable 经 PlayerController 间接导入数据源；测试不访问服务端。
vi.mock("@/api/source", () => ({
  source: { coverUrl: () => "", audioUrl: (id: number) => `/audio/${id}` },
  fixturePlayer: null,
  usingFixture: true,
  loginUrl: "/auth/login",
}));

/** 构造 keydown 事件的字段 */
const keyEvent = (code: string, modifiers: Partial<KeyboardEvent> = {}) =>
  ({
    code,
    ctrlKey: false,
    metaKey: false,
    shiftKey: false,
    altKey: false,
    ...modifiers,
  }) as KeyboardEvent;

const list = defaultShortcutState().shortcutList;

describe("快捷键匹配", () => {
  it("命中默认的组合键", () => {
    expect(
      matchesShortcut(keyEvent("Space", { ctrlKey: true }), "CmdOrCtrl+Space"),
    ).toBe(true);
    expect(matchesShortcut(keyEvent("KeyP"), "KeyP")).toBe(true);
    expect(matchesShortcut(keyEvent("ArrowRight"), "ArrowRight")).toBe(true);
  });

  it("修饰键状态不吻合时不命中", () => {
    // 裸空格不等于 CmdOrCtrl+Space
    expect(matchesShortcut(keyEvent("Space"), "CmdOrCtrl+Space")).toBe(false);
    // 多按了修饰键
    expect(
      matchesShortcut(
        keyEvent("Space", { ctrlKey: true, shiftKey: true }),
        "CmdOrCtrl+Space",
      ),
    ).toBe(false);
    expect(matchesShortcut(keyEvent("KeyP", { altKey: true }), "KeyP")).toBe(
      false,
    );
  });

  it("主键不同时不命中", () => {
    expect(matchesShortcut(keyEvent("KeyL"), "KeyP")).toBe(false);
  });

  it("列出事件命中的快捷键", () => {
    // 默认配置：CmdOrCtrl+Space 播放暂停
    expect(
      matchedShortcuts(keyEvent("Space", { ctrlKey: true }), list),
    ).toEqual(["playOrPause"]);
    // 未注册的按键
    expect(matchedShortcuts(keyEvent("KeyA"), list)).toEqual([]);
    // 裸空格（未注册）不应命中任何快捷键
    expect(matchedShortcuts(keyEvent("Space"), list)).toEqual([]);
  });

  it("播放暂停绑定为裸空格时命中", () => {
    const bound: ShortcutList = {
      ...list,
      playOrPause: { name: "播放 / 暂停", shortcut: "Space" },
    };
    expect(matchedShortcuts(keyEvent("Space"), bound)).toEqual(["playOrPause"]);
  });
});
