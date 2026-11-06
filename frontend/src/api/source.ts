import { ApiError } from "./error";
import { connection } from "./connection";
import { onlineImage } from "@/utils/onlineImage";
import type {
  Album,
  AlbumQuery,
  Artist,
  Cover,
  HistoryEntry,
  ListQuery,
  Lyrics,
  MixedPage,
  NeteaseQRState,
  NeteaseStatus,
  OnlineList,
  OnlineListType,
  Page,
  Playlist,
  PlaylistInput,
  Track,
  TrackQuery,
  User,
} from "./types";
import type { PlayerState } from "@/viewmodels";

export { ApiError };

export type CoverSize = 96 | 100 | 300 | 1024;

/** 播放端读取数据的唯一入口。页面与 store 只通过它访问服务端，开发时可整体换成 fixture。 */
export interface DataSource {
  me(): Promise<User>;
  /** 网易云账号的状态。Cookie 只保存在服务端，不经过播放端。 */
  neteaseStatus(): Promise<NeteaseStatus>;
  /** 开始二维码登录，`url` 是二维码的内容。 */
  neteaseQRStart(): Promise<{ key: string; url: string }>;
  /** 查询二维码登录的进度；成功时服务端已保存会话。 */
  neteaseQRCheck(key: string): Promise<NeteaseQRState>;
  neteaseCookieLogin(cookie: string): Promise<NeteaseStatus>;
  /** 只退出网易云，不影响 Sonora 的登录。 */
  neteaseLogout(): Promise<void>;
  /** 我喜欢的音乐：本地与在线按喜欢时间混排，同一首歌只有一条，优先本地。 */
  mixedLikes(
    query?: Pick<ListQuery, "offset" | "limit">,
  ): Promise<MixedPage<Track>>;
  /** 最近播放：本地与在线按时间混排。 */
  mixedHistory(
    query?: Pick<ListQuery, "offset" | "limit">,
  ): Promise<MixedPage<HistoryEntry>>;
  /** 记录一首在线歌曲的播放，服务端保存歌曲的快照。 */
  recordOnlinePlay(song: Track): Promise<void>;
  /** 自己创建的与收藏的在线歌单。 */
  onlinePlaylists(): Promise<{ created: Playlist[]; collected: Playlist[] }>;
  onlineAlbums(): Promise<Page<Album>>;
  onlineArtists(): Promise<Page<Artist>>;
  onlineDailySongs(): Promise<Page<Track>>;
  onlineList(type: OnlineListType, id: number): Promise<OnlineList>;
  onlineArtistAlbums(
    id: number,
    query?: Pick<ListQuery, "offset" | "limit">,
  ): Promise<Page<Album>>;
  createOnlinePlaylist(input: {
    name: string;
    private: boolean;
  }): Promise<Playlist>;
  updateOnlinePlaylist(id: number, input: PlaylistInput): Promise<void>;
  deleteOnlinePlaylist(id: number): Promise<void>;
  addToOnlinePlaylist(id: number, songIds: number[]): Promise<void>;
  /** 在线歌曲的播放地址，带时效，每次播放时获取。 */
  onlineSongUrl(id: number): Promise<string>;
  onlineLyrics(id: number): Promise<Lyrics>;
  /** 由服务端下载并入库，耗时取决于网络。已导入过时 `status` 为 `duplicate`。 */
  importOnlineSong(
    id: number,
  ): Promise<{ status: "created" | "duplicate"; track: Track }>;
  onlineLikeIds(): Promise<number[]>;
  onlineLike(id: number): Promise<void>;
  onlineUnlike(id: number): Promise<void>;
  tracks(query?: TrackQuery): Promise<Page<Track>>;
  track(id: number): Promise<Track>;
  lyrics(trackId: number): Promise<Lyrics>;
  albums(query?: AlbumQuery): Promise<Page<Album>>;
  album(id: number): Promise<Album>;
  artists(
    query?: Pick<ListQuery, "q" | "offset" | "limit">,
  ): Promise<Page<Artist>>;
  artist(id: number): Promise<Artist>;
  likes(query?: ListQuery): Promise<Page<Track>>;
  /** 全部喜爱歌曲的 ID，用于在任意列表中标记喜爱状态。 */
  likeIds(): Promise<number[]>;
  history(
    query?: Pick<ListQuery, "offset" | "limit">,
  ): Promise<Page<HistoryEntry>>;
  /** 自己的歌单，新建的在前；`q` 匹配歌单名。 */
  playlists(query?: Pick<ListQuery, "q">): Promise<Page<Playlist>>;
  playlist(id: number): Promise<Playlist>;
  playlistTracks(id: number, query?: ListQuery): Promise<Page<Track>>;
  /** 喜爱歌曲，幂等。 */
  like(trackId: number): Promise<void>;
  /** 取消喜爱，幂等。 */
  unlike(trackId: number): Promise<void>;
  /** 记录一次播放；什么时候算一次播放由播放控制器决定。 */
  recordPlay(trackId: number): Promise<void>;
  /** 清空最近播放。 */
  clearHistory(): Promise<void>;
  createPlaylist(input: PlaylistInput): Promise<Playlist>;
  updatePlaylist(id: number, input: PlaylistInput): Promise<Playlist>;
  deletePlaylist(id: number): Promise<void>;
  /** 按给定顺序追加到歌单末尾，已在歌单中的跳过；`added` 是实际加入的数量。 */
  addToPlaylist(
    id: number,
    trackIds: number[],
  ): Promise<{ added: number; playlist: Playlist }>;
  removeFromPlaylist(id: number, trackId: number): Promise<Playlist>;
  /**
   * 保存歌单的新顺序，`trackIds` 为歌单的全部歌曲。与服务端不一致时为 409
   * `conflict`，说明本地的副本已过期。
   */
  setPlaylistOrder(id: number, trackIds: number[]): Promise<Playlist>;
  /**
   * 音频地址的 HTTP 状态，0 表示服务不可达。音频元素出错时只给出 MediaError，
   * 播放控制器用它区分未登录、歌曲已删除与网络问题。
   */
  audioStatus(trackId: number): Promise<number>;
  /** 当前用户保存的设置（外观、歌词等，设备相关的不上传），从未保存时为 `{}`。 */
  settings(): Promise<Record<string, unknown>>;
  /**
   * 整体替换保存的设置。`keepalive` 用于关闭页面时补发，页面卸载后请求仍会发出。
   */
  saveSettings(
    data: Record<string, unknown>,
    options?: { keepalive?: boolean },
  ): Promise<void>;
  /** 结束本浏览器的会话，不退出 Authelia。 */
  logout(): Promise<void>;
  /** 不指定尺寸时为原图；尺寸为长边像素，见 docs/api.md 的 `covers/{name}`。 */
  coverUrl(cover: Cover, size?: CoverSize): string;
  audioUrl(trackId: number): string;
}

const api = "/api/v1/";

/** 把查询参数拼到路径后，省略未设置的项。 */
function withQuery(path: string, query: object = {}): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined && value !== "") params.set(key, String(value));
  }
  const search = params.toString();
  return search ? `${path}?${search}` : path;
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  keepalive?: boolean,
): Promise<T> {
  let res: Response;
  try {
    res = await fetch(path, {
      method,
      cache: "no-store",
      keepalive,
      ...(body === undefined
        ? {}
        : {
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
          }),
    });
  } catch {
    connection.unreachable();
    throw new ApiError(0, "无法连接服务器");
  }
  const data: unknown = await res.json().catch(() => null);
  const failure = data as { error?: string; code?: string } | null;
  if (proxyFailure(res.status, !!failure?.error)) {
    connection.unreachable();
    throw new ApiError(0, "无法连接服务器");
  }
  connection.reached();
  if (!res.ok) {
    if (res.status === 401) connection.unauthorized();
    throw new ApiError(
      res.status,
      failure?.error ?? `请求失败（${res.status}）`,
      failure?.code,
    );
  }
  return data as T;
}

/**
 * Go 服务的错误响应都带 `{error}` 正文；没有正文的 5xx 来自服务前面的代理
 * （Vite 或反向代理），说明服务没有响应，按不可达处理。
 */
const proxyFailure = (status: number, hasErrorBody: boolean) =>
  status >= 500 && !hasErrorBody;

const get = <T>(path: string, query?: object) =>
  request<T>("GET", withQuery(api + path, query));

// 写请求成功时为 204，没有正文。
const send = async (method: string, path: string, body?: unknown) => {
  await request<null>(method, api + path, body);
};

// 返回正文的写请求。
const write = <T>(method: string, path: string, body?: unknown) =>
  request<T>(method, api + path, body);

const httpSource: DataSource = {
  me: () => get("me"),
  neteaseStatus: () => get("netease/status"),
  neteaseQRStart: () => write("POST", "netease/login/qr"),
  neteaseQRCheck: (key) => get(`netease/login/qr/${encodeURIComponent(key)}`),
  neteaseCookieLogin: (cookie) =>
    write("POST", "netease/login/cookie", { cookie }),
  neteaseLogout: () => send("POST", "netease/logout"),
  mixedLikes: (query) => get("likes/mixed", query),
  mixedHistory: (query) => get("history/mixed", query),
  recordOnlinePlay: (song) => send("POST", "history/online", { song }),
  onlinePlaylists: () => get("netease/playlists"),
  onlineAlbums: () => get("netease/albums"),
  onlineArtists: () => get("netease/artists"),
  onlineDailySongs: () => get("netease/recommend/songs"),
  onlineList: (type, id) => get(`netease/${type}s/${id}`),
  onlineArtistAlbums: (id, query) => get(`netease/artists/${id}/albums`, query),
  createOnlinePlaylist: (input) => write("POST", "netease/playlists", input),
  updateOnlinePlaylist: (id, input) =>
    send("PUT", `netease/playlists/${id}`, input),
  deleteOnlinePlaylist: (id) => send("DELETE", `netease/playlists/${id}`),
  addToOnlinePlaylist: (id, ids) =>
    send("POST", `netease/playlists/${id}/tracks`, { ids }),
  onlineSongUrl: async (id) =>
    (await get<{ url: string }>(`netease/songs/${id}/url`)).url,
  onlineLyrics: (id) => get(`netease/songs/${id}/lyrics`),
  importOnlineSong: (id) => write("POST", `netease/songs/${id}/import`),
  onlineLikeIds: async () =>
    (await get<{ ids: number[] }>("netease/likes/ids")).ids,
  onlineLike: (id) => send("PUT", `netease/likes/${id}`),
  onlineUnlike: (id) => send("DELETE", `netease/likes/${id}`),
  tracks: (query) => get("tracks", query),
  track: (id) => get(`tracks/${id}`),
  lyrics: (trackId) => get(`tracks/${trackId}/lyrics`),
  albums: (query) => get("albums", query),
  album: (id) => get(`albums/${id}`),
  artists: (query) => get("artists", query),
  artist: (id) => get(`artists/${id}`),
  likes: (query) => get("likes", query),
  likeIds: async () => (await get<{ ids: number[] }>("likes/ids")).ids,
  history: (query) => get("history", query),
  playlists: (query) => get("playlists", { ...query, limit: 500 }),
  playlist: (id) => get(`playlists/${id}`),
  playlistTracks: (id, query) => get(`playlists/${id}/tracks`, query),
  like: (trackId) => send("PUT", `likes/${trackId}`),
  unlike: (trackId) => send("DELETE", `likes/${trackId}`),
  recordPlay: (trackId) => send("POST", "history", { trackId }),
  clearHistory: () => send("DELETE", "history"),
  createPlaylist: (input) => write("POST", "playlists", input),
  updatePlaylist: (id, input) => write("PUT", `playlists/${id}`, input),
  deletePlaylist: (id) => send("DELETE", `playlists/${id}`),
  addToPlaylist: (id, trackIds) =>
    write("POST", `playlists/${id}/tracks`, { trackIds }),
  removeFromPlaylist: (id, trackId) =>
    write("DELETE", `playlists/${id}/tracks/${trackId}`),
  setPlaylistOrder: (id, trackIds) =>
    write("PUT", `playlists/${id}/order`, { trackIds }),
  settings: () => get("settings"),
  saveSettings: async (data, options) => {
    await request<null>("PUT", api + "settings", data, options?.keepalive);
  },
  async audioStatus(trackId) {
    try {
      const res = await fetch(httpSource.audioUrl(trackId), {
        headers: { Range: "bytes=0-0" },
        cache: "no-store",
      });
      await res.body?.cancel();
      const json = !!res.headers
        .get("Content-Type")
        ?.includes("application/json");
      if (proxyFailure(res.status, json)) {
        connection.unreachable();
        return 0;
      }
      connection.reached();
      return res.status;
    } catch {
      connection.unreachable();
      return 0;
    }
  },
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
    cover.url
      ? onlineImage(cover.url, size ?? 1024)
      : withQuery(`${api}covers/${encodeURIComponent(cover.name)}`, { size }),
  audioUrl: (trackId) => `${api}tracks/${trackId}/audio`,
};

/**
 * 开发服务器上以 `?fixture=<场景>` 打开时使用 fixture，不需要 Go 服务；
 * `&player=<状态>` 再指定播放器的初始状态。生产构建不包含 fixture。
 */
async function createSource(): Promise<{
  source: DataSource;
  player: PlayerState | null;
}> {
  if (import.meta.env.DEV) {
    const params = new URLSearchParams(location.search);
    const scenario = params.get("fixture");
    if (scenario) {
      const { createFixture } = await import("@/fixtures");
      return createFixture(scenario, params.get("player"));
    }
  }
  return { source: httpSource, player: null };
}

const created = await createSource();

export const source = created.source;

/** 是否在使用 fixture；fixture 没有音频，播放器换成不出声的引擎，也不保存队列。 */
export const usingFixture = source !== httpSource;

/**
 * fixture 指定的播放器初始状态，代替浏览器中保存的队列与播放位置；
 * 不使用 fixture 时为 null。
 */
export const fixturePlayer = created.player;

export const loginUrl = "/auth/login";
