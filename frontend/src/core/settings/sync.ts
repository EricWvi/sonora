import { watch } from "vue";
import { isPlainObject } from "lodash-es";
import { source, type DataSource } from "@/api/source";
import { connection } from "@/api/connection";
import { useSettingStore, useShortcutStore } from "@/stores";
import { defaultSettingState, type SettingState } from "@/stores/setting";
import { defaultShortcutState, type ShortcutList } from "@/stores/shortcut";

/**
 * 设置与服务端同步（阶段 5 的决定，见 docs/phase/phase-5-player-pages.md）。
 *
 * - setting store 与快捷键中除设备相关项外全部上传；均衡器与音量在 status store，不上传。
 * - 打开页面时先用浏览器中的缓存（pinia 持久化）渲染，读取服务端后以服务端为准，缺少或类型
 *   不符的字段用默认值；服务端为 {} 时上传浏览器中的设置。读取失败时只用本地，本地的改动
 *   等下次读取成功后上传。
 * - 最后一次改动后 1 秒整体上传，关闭页面时用 keepalive 补发。后写者胜；页面重新可见且没有
 *   待上传的改动时重新读取。
 */

/** 留在浏览器中的设备相关设置：输出设备、同步补偿与侧栏展开状态 */
export const DEVICE_KEYS = [
  "playDevice",
  "audioDelayCompensation",
  "menuExpandedKeys",
] as const;

type DeviceKey = (typeof DEVICE_KEYS)[number];

/** 上传到服务端的设置 */
export type SyncedSettings = {
  setting: Omit<SettingState, DeviceKey>;
  /** 快捷键的组合键，名称不上传 */
  shortcuts: Record<keyof ShortcutList, string>;
};

/** 从 store 的状态中取出要上传的部分 */
export function pickSynced(
  setting: SettingState,
  shortcutList: ShortcutList,
): SyncedSettings {
  const synced = { ...setting } as Partial<SettingState>;
  for (const key of DEVICE_KEYS) delete synced[key];
  const shortcuts = {} as Record<keyof ShortcutList, string>;
  for (const key of Object.keys(shortcutList) as (keyof ShortcutList)[])
    shortcuts[key] = shortcutList[key].shortcut;
  // 经 JSON 复制：store 中的对象是响应式代理，上传的也是 JSON
  return {
    setting: JSON.parse(JSON.stringify(synced)) as SyncedSettings["setting"],
    shortcuts,
  };
}

/** 默认的上传部分 */
export const defaultSynced = () =>
  pickSynced(defaultSettingState(), defaultShortcutState().shortcutList);

/**
 * 按默认值的形状取值：类型不同时用默认值；对象逐个字段处理，未知字段丢弃；
 * 数组要求每一项与默认值的第一项形状相同（默认为空数组时只要求是数组）。
 */
function conform(value: unknown, fallback: unknown): unknown {
  if (Array.isArray(fallback)) {
    if (!Array.isArray(value)) return fallback;
    if (!fallback.length) return value;
    const item: unknown = fallback[0];
    const valid = value.every((v) =>
      isPlainObject(item)
        ? isPlainObject(v) && sameShape(v as object, item as object)
        : typeof v === typeof item,
    );
    return valid ? value : fallback;
  }
  if (isPlainObject(fallback)) {
    if (!isPlainObject(value)) return fallback;
    const source = value as Record<string, unknown>;
    const result: Record<string, unknown> = {};
    for (const [key, def] of Object.entries(fallback as object))
      result[key] = key in source ? conform(source[key], def) : def;
    return result;
  }
  return typeof value === typeof fallback ? value : fallback;
}

/** 对象的每个已知字段类型都与默认值相同，且没有缺少的字段 */
function sameShape(value: object, fallback: object): boolean {
  const source = value as Record<string, unknown>;
  return Object.entries(fallback).every(
    ([key, def]) => key in source && typeof source[key] === typeof def,
  );
}

/** 校验服务端保存的设置，缺少或类型不符的字段用默认值 */
export function resolveRemote(remote: unknown): SyncedSettings {
  const fallback = defaultSynced();
  const data = isPlainObject(remote) ? (remote as Record<string, unknown>) : {};
  return {
    setting: conform(
      data.setting,
      fallback.setting,
    ) as SyncedSettings["setting"],
    shortcuts: conform(
      data.shortcuts,
      fallback.shortcuts,
    ) as SyncedSettings["shortcuts"],
  };
}

export type SettingsSyncDeps = {
  source: Pick<DataSource, "settings" | "saveSettings">;
  /** 读取当前要上传的设置 */
  read(): SyncedSettings;
  /** 应用服务端的设置 */
  apply(settings: SyncedSettings): void;
  /** 改动后等待多久上传（毫秒） */
  delay?: number;
};

export class SettingsSync {
  /** 已与服务端一致的设置（JSON），读取成功前为 null */
  private synced: string | null = null;
  /** 浏览器中最近一次与服务端一致、或启动时的设置，用来判断读取前有没有本地改动 */
  private localBase: string;
  private timer: ReturnType<typeof setTimeout> | null = null;
  private loading: Promise<void> | null = null;
  /** 上传按先后依次进行，避免较早的请求后到 */
  private queue: Promise<void> = Promise.resolve();
  private uploading = 0;

  constructor(private deps: SettingsSyncDeps) {
    this.localBase = this.snapshot();
  }

  private snapshot() {
    return JSON.stringify(this.deps.read());
  }

  /** 是否已读取过服务端的设置 */
  get loaded() {
    return this.synced !== null;
  }

  /** 是否有尚未上传的改动 */
  get dirty() {
    const current = this.snapshot();
    return this.loaded ? current !== this.synced : current !== this.localBase;
  }

  /** 读取服务端的设置。有待上传的改动时上传本地的设置，否则以服务端为准。 */
  load(): Promise<void> {
    this.loading ??= this.doLoad().finally(() => (this.loading = null));
    return this.loading;
  }

  private async doLoad() {
    let remote: Record<string, unknown>;
    try {
      remote = await this.deps.source.settings();
    } catch (error) {
      console.warn("读取服务端设置失败，使用浏览器中的设置", error);
      return;
    }
    // 服务端从未保存过，或本地有尚未上传的改动：上传本地的设置
    if (!Object.keys(remote).length || this.dirty || this.timer) {
      await this.flush();
      return;
    }
    this.deps.apply(resolveRemote(remote));
    this.synced = this.localBase = this.snapshot();
  }

  /** 设置改动时调用：读取过服务端后，最后一次改动 1 秒后上传 */
  changed() {
    if (!this.loaded) return;
    if (this.timer) clearTimeout(this.timer);
    this.timer = null;
    if (!this.dirty) return;
    this.timer = setTimeout(() => void this.flush(), this.deps.delay ?? 1000);
  }

  /** 立即上传（没有改动时不上传） */
  flush(options?: { keepalive?: boolean }): Promise<void> {
    if (this.timer) clearTimeout(this.timer);
    this.timer = null;
    this.queue = this.queue.then(() => this.upload(options?.keepalive));
    return this.queue;
  }

  private async upload(keepalive?: boolean) {
    const data = this.deps.read();
    const json = JSON.stringify(data);
    if (json === this.synced) return;
    this.uploading++;
    try {
      await this.deps.source.saveSettings(data, { keepalive });
      this.synced = this.localBase = json;
    } catch (error) {
      // 保留改动，下次修改或页面重新可见时再上传
      console.warn("上传设置失败", error);
    } finally {
      this.uploading--;
    }
  }

  /** 页面重新可见：有待上传的改动时上传，否则重新读取 */
  visible() {
    if (this.loading || this.uploading || this.timer) return;
    if (this.loaded && this.dirty) void this.flush();
    else void this.load();
  }

  /** 关闭页面：有待上传的改动时用 keepalive 补发 */
  pagehide() {
    if (this.loaded && (this.timer || this.dirty))
      void this.flush({ keepalive: true });
  }
}

let instance: SettingsSync | null = null;

/** 启动时调用：读取服务端设置，之后跟随 store 的改动上传 */
export function initSettingsSync() {
  if (instance) return instance;
  const settingStore = useSettingStore();
  const shortcutStore = useShortcutStore();
  const sync = new SettingsSync({
    source,
    read: () => pickSynced(settingStore.$state, shortcutStore.shortcutList),
    apply: ({ setting, shortcuts }) => {
      settingStore.$patch((state) => Object.assign(state, setting));
      shortcutStore.$patch((state) => {
        for (const key of Object.keys(shortcuts) as (keyof ShortcutList)[])
          state.shortcutList[key].shortcut = shortcuts[key];
      });
    },
  });
  instance = sync;
  watch(
    () =>
      JSON.stringify(
        pickSynced(settingStore.$state, shortcutStore.shortcutList),
      ),
    () => sync.changed(),
  );
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") sync.visible();
  });
  window.addEventListener("pagehide", () => sync.pagehide());
  // 服务恢复后：上传期间积累的改动，或重新读取
  connection.onReconnect(() => sync.visible());
  void sync.load();
  return sync;
}

/** 已启动的同步，未启动时为 null */
export const useSettingsSync = () => instance;
