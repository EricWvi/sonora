import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { defaultSettingState } from "@/stores/setting";
import { defaultShortcutState } from "@/stores/shortcut";
import {
  DEVICE_KEYS,
  SettingsSync,
  pickSynced,
  resolveRemote,
  type SyncedSettings,
} from "./sync";

vi.mock("@/api/source", () => ({ source: {} }));

/** 浏览器中的设置：一份 setting store 与快捷键 */
function localState() {
  return {
    setting: defaultSettingState(),
    shortcuts: defaultShortcutState().shortcutList,
  };
}

/** 可控的服务端：记录上传，读取与上传可以失败 */
function fakeServer(initial: Record<string, unknown> = {}) {
  const server = {
    data: structuredClone(initial),
    saved: [] as { data: Record<string, unknown>; keepalive?: boolean }[],
    failRead: false,
    failWrite: false,
    settings: vi.fn(async () => {
      if (server.failRead) throw new Error("无法连接服务器");
      return structuredClone(server.data);
    }),
    saveSettings: vi.fn(
      async (
        data: Record<string, unknown>,
        options?: { keepalive?: boolean },
      ) => {
        if (server.failWrite) throw new Error("无法连接服务器");
        server.data = structuredClone(data);
        server.saved.push({ data: server.data, keepalive: options?.keepalive });
      },
    ),
  };
  return server;
}

function setup(
  server: ReturnType<typeof fakeServer>,
  // 启动前浏览器中已有的设置
  prepare?: (local: ReturnType<typeof localState>) => void,
) {
  const local = localState();
  prepare?.(local);
  const sync = new SettingsSync({
    source: server,
    read: () => pickSynced(local.setting, local.shortcuts),
    apply: ({ setting, shortcuts }: SyncedSettings) => {
      Object.assign(local.setting, setting);
      for (const key of Object.keys(shortcuts) as (keyof typeof shortcuts)[])
        local.shortcuts[key].shortcut = shortcuts[key];
    },
  });
  return { local, sync };
}

describe("设置同步", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("上传的设置不含设备相关项，快捷键只有组合键", () => {
    const { setting, shortcuts } = pickSynced(
      defaultSettingState(),
      defaultShortcutState().shortcutList,
    );
    for (const key of DEVICE_KEYS) expect(setting).not.toHaveProperty(key);
    expect(setting).toHaveProperty("themeMode", "auto");
    expect(shortcuts.closePlayer).toBe("Escape");
  });

  it("服务端为 {} 时上传浏览器中的设置", async () => {
    const server = fakeServer();
    const { local, sync } = setup(server);
    local.setting.themeMode = "dark";
    await sync.load();
    expect(server.saved).toHaveLength(1);
    expect(server.data).toMatchObject({ setting: { themeMode: "dark" } });
  });

  it("服务端有设置时以服务端为准，设备相关项不变", async () => {
    const remote = pickSynced(
      { ...defaultSettingState(), themeMode: "dark", lyricFontSize: 30 },
      defaultShortcutState().shortcutList,
    );
    remote.shortcuts.openPlayer = "KeyO";
    const server = fakeServer(remote);
    const { local, sync } = setup(server, (local) => {
      local.setting.themeMode = "light";
      local.setting.playDevice = "speaker";
      local.setting.audioDelayCompensation = 120;
    });
    await sync.load();
    expect(local.setting.themeMode).toBe("dark");
    expect(local.setting.lyricFontSize).toBe(30);
    expect(local.shortcuts.openPlayer.shortcut).toBe("KeyO");
    expect(local.setting.playDevice).toBe("speaker");
    expect(local.setting.audioDelayCompensation).toBe(120);
    expect(server.saved).toHaveLength(0);
  });

  it("服务端缺少或类型不符的字段用默认值，未知字段丢弃", () => {
    const resolved = resolveRemote({
      setting: {
        themeMode: "dark",
        lyricFontSize: "30",
        hiddenCovers: { home: true, unknown: true },
        homePageSections: [{ key: "album" }],
        excludeLyricsUserKeywords: ["作词"],
        removed: 1,
      },
      shortcuts: { openPlayer: "KeyO", playOrPause: 1 },
    });
    const defaults = defaultSettingState();
    expect(resolved.setting.themeMode).toBe("dark");
    expect(resolved.setting.lyricFontSize).toBe(defaults.lyricFontSize);
    expect(resolved.setting.hiddenCovers).toEqual({
      ...defaults.hiddenCovers,
      home: true,
    });
    expect(resolved.setting.homePageSections).toEqual(
      defaults.homePageSections,
    );
    expect(resolved.setting.excludeLyricsUserKeywords).toEqual(["作词"]);
    expect(resolved.setting.showTran).toBe(defaults.showTran);
    expect(resolved.setting).not.toHaveProperty("removed");
    expect(resolved.shortcuts.openPlayer).toBe("KeyO");
    expect(resolved.shortcuts.playOrPause).toBe("CmdOrCtrl+Space");
  });

  it("改动 1 秒后整体上传，连续改动只上传一次", async () => {
    const server = fakeServer(
      pickSynced(defaultSettingState(), defaultShortcutState().shortcutList),
    );
    const { local, sync } = setup(server);
    await sync.load();
    local.setting.lyricFontSize = 40;
    sync.changed();
    await vi.advanceTimersByTimeAsync(600);
    local.setting.lyricFontSize = 42;
    sync.changed();
    await vi.advanceTimersByTimeAsync(600);
    expect(server.saved).toHaveLength(0);
    await vi.advanceTimersByTimeAsync(400);
    expect(server.saved).toHaveLength(1);
    expect(server.data).toMatchObject({ setting: { lyricFontSize: 42 } });
  });

  it("只改设备相关项时不上传", async () => {
    const server = fakeServer(
      pickSynced(defaultSettingState(), defaultShortcutState().shortcutList),
    );
    const { local, sync } = setup(server);
    await sync.load();
    local.setting.audioDelayCompensation = 50;
    sync.changed();
    await vi.advanceTimersByTimeAsync(2000);
    expect(server.saved).toHaveLength(0);
  });

  it("读取失败时只用本地，本地改动在下次读取成功后上传", async () => {
    const server = fakeServer(
      pickSynced(
        { ...defaultSettingState(), themeMode: "dark" },
        defaultShortcutState().shortcutList,
      ),
    );
    server.failRead = true;
    const { local, sync } = setup(server);
    await sync.load();
    expect(sync.loaded).toBe(false);
    local.setting.themeMode = "light";
    sync.changed();
    await vi.advanceTimersByTimeAsync(2000);
    expect(server.saved).toHaveLength(0);
    server.failRead = false;
    sync.visible();
    await vi.advanceTimersByTimeAsync(0);
    expect(local.setting.themeMode).toBe("light");
    expect(server.data).toMatchObject({ setting: { themeMode: "light" } });
  });

  it("页面重新可见且没有待上传的改动时重新读取", async () => {
    const server = fakeServer(
      pickSynced(defaultSettingState(), defaultShortcutState().shortcutList),
    );
    const { local, sync } = setup(server);
    await sync.load();
    // 别的设备改了设置
    server.data = pickSynced(
      { ...defaultSettingState(), playerType: "record" },
      defaultShortcutState().shortcutList,
    );
    sync.visible();
    await vi.advanceTimersByTimeAsync(0);
    expect(local.setting.playerType).toBe("record");
    // 有待上传的改动时不读取，等上传
    local.setting.playerType = "fullscreen";
    sync.changed();
    server.data = pickSynced(
      { ...defaultSettingState(), playerType: "cover" },
      defaultShortcutState().shortcutList,
    );
    sync.visible();
    await vi.advanceTimersByTimeAsync(1000);
    expect(local.setting.playerType).toBe("fullscreen");
    expect(server.data).toMatchObject({
      setting: { playerType: "fullscreen" },
    });
  });

  it("关闭页面时用 keepalive 补发待上传的改动", async () => {
    const server = fakeServer(
      pickSynced(defaultSettingState(), defaultShortcutState().shortcutList),
    );
    const { local, sync } = setup(server);
    await sync.load();
    local.setting.showTran = false;
    sync.changed();
    sync.pagehide();
    await vi.advanceTimersByTimeAsync(0);
    expect(server.saved).toEqual([
      expect.objectContaining({ keepalive: true }),
    ]);
    await vi.advanceTimersByTimeAsync(2000);
    expect(server.saved).toHaveLength(1);
  });

  it("上传失败时保留改动，页面重新可见时重传", async () => {
    const server = fakeServer(
      pickSynced(defaultSettingState(), defaultShortcutState().shortcutList),
    );
    const { local, sync } = setup(server);
    await sync.load();
    server.failWrite = true;
    local.setting.lyricsBlur = true;
    sync.changed();
    await vi.advanceTimersByTimeAsync(1000);
    expect(sync.dirty).toBe(true);
    server.failWrite = false;
    sync.visible();
    await vi.advanceTimersByTimeAsync(0);
    expect(server.data).toMatchObject({ setting: { lyricsBlur: true } });
    expect(sync.dirty).toBe(false);
  });
});
