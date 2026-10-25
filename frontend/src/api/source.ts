import { ApiError } from "./error";
import type { Cover, Page, Playlist, User } from "./types";

export { ApiError };

/** 播放端读取数据的唯一入口。页面与 store 只通过它访问服务端，开发时可整体换成 fixture。 */
export interface DataSource {
  me(): Promise<User>;
  playlists(): Promise<Page<Playlist>>;
  /** 结束本浏览器的会话，不退出 Authelia。 */
  logout(): Promise<void>;
  coverUrl(cover: Cover, size?: 96 | 300): string;
}

async function request<T>(method: string, path: string): Promise<T> {
  let res: Response;
  try {
    res = await fetch(path, { method, cache: "no-store" });
  } catch {
    throw new ApiError(0, "无法连接服务器");
  }
  const data: unknown = await res.json().catch(() => null);
  if (!res.ok) {
    const body = data as { error?: string; code?: string } | null;
    throw new ApiError(
      res.status,
      body?.error ?? `请求失败（${res.status}）`,
      body?.code,
    );
  }
  return data as T;
}

const httpSource: DataSource = {
  me: () => request("GET", "/api/v1/me"),
  playlists: () => request("GET", "/api/v1/playlists?limit=500"),
  async logout() {
    let res: Response;
    try {
      res = await fetch("/auth/logout", { method: "POST", cache: "no-store" });
    } catch {
      throw new ApiError(0, "无法连接服务器，未能退出登录");
    }
    if (!res.ok) throw new ApiError(res.status, "退出登录失败");
  },
  coverUrl: (cover, size) =>
    `/api/v1/covers/${encodeURIComponent(cover.name)}${size ? `?size=${size}` : ""}`,
};

/**
 * 开发服务器上以 `?fixture=<场景>` 打开时使用 fixture，不需要 Go 服务；
 * 生产构建不包含 fixture。
 */
async function createSource(): Promise<DataSource> {
  if (import.meta.env.DEV) {
    const scenario = new URLSearchParams(location.search).get("fixture");
    if (scenario) {
      const { createFixtureSource } = await import("@/fixtures");
      return createFixtureSource(scenario);
    }
  }
  return httpSource;
}

export const source = await createSource();

export const loginUrl = "/auth/login";
