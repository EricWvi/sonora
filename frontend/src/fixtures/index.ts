import { ApiError } from "@/api/error";
import type { DataSource } from "@/api/source";
import type { Page, Playlist, User } from "@/api/types";

// 开发用 fixture，以 `?fixture=<场景>` 打开。数据与服务端开发模式的用户一致，内容是虚构的。

const user: User = { name: "listener", email: "listener@sonora.test" };

const playlists: Playlist[] = [
  ["通勤路上", "早晚各一小时"],
  ["夜跑", ""],
  ["写代码时听", "没有人声或人声很少的歌"],
  ["一个名字很长很长很长很长很长很长的歌单", ""],
].map(([name, description], index) => ({
  id: 4 - index,
  name,
  description,
  cover: null,
  trackCount: 0,
  durationMs: 0,
  createdAt: `2026-10-${20 + index}T12:00:00Z`,
  updatedAt: `2026-10-${20 + index}T12:00:00Z`,
}));

const page = <T>(items: T[]): Page<T> => ({ items, total: items.length });

type Data = { playlists: Playlist[] };

const scenarios: Record<string, Data | "loading" | "unreachable"> = {
  /** 有数据 */
  library: { playlists },
  /** 空曲库，没有歌单 */
  empty: { playlists: [] },
  /** 请求一直没有返回 */
  loading: "loading",
  /** 服务不可达 */
  unreachable: "unreachable",
};

export function createFixtureSource(name: string): DataSource {
  const scenario = scenarios[name];
  if (!scenario) {
    throw new Error(
      `未知的 fixture 场景 "${name}"，可用：${Object.keys(scenarios).join("、")}`,
    );
  }
  const respond = <T>(value: (data: Data) => T): Promise<T> => {
    if (scenario === "loading") return new Promise(() => {});
    if (scenario === "unreachable")
      return Promise.reject(new ApiError(0, "无法连接服务器"));
    return Promise.resolve(structuredClone(value(scenario)));
  };
  return {
    me: () => respond(() => user),
    playlists: () => respond((data) => page(data.playlists)),
    logout: () => Promise.resolve(),
    coverUrl: (cover) => `/fixtures/covers/${cover.name}`,
  };
}
