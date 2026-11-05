// 播放端 API 的数据形态，见 docs/api.md。字段与服务端响应一致，不含文件路径，也不含网易字段。

export type User = {
  name: string;
  email: string;
};

/** 封面与艺术家图片：内容哈希文件名与 ThumbHash 占位。 */
export type Cover = {
  name: string;
  thumbhash: string;
};

export type Page<T> = {
  items: T[];
  total: number;
};

/** 歌曲、专辑中的艺术家，按署名顺序。 */
export type ArtistRef = {
  id: number;
  name: string;
};

/** 歌曲所属的专辑。 */
export type AlbumRef = {
  id: number;
  title: string;
  artists: ArtistRef[];
};

export type Track = {
  id: number;
  title: string;
  artists: ArtistRef[];
  album: AlbumRef | null;
  trackNo: number;
  discNo: number;
  year: number;
  durationMs: number;
  format: string;
  /** 歌曲实际显示的封面：单曲封面优先，否则为专辑封面。 */
  cover: Cover | null;
  hasLyrics: boolean;
};

export type Album = {
  id: number;
  title: string;
  artists: ArtistRef[];
  year: number;
  cover: Cover | null;
  /** 管理后台填写的简介，没有时为空字符串。 */
  description: string;
  trackCount: number;
  durationMs: number;
};

export type Artist = {
  id: number;
  name: string;
  /** 管理后台上传的艺术家图片。 */
  image: Cover | null;
  /** 管理后台填写的简介，没有时为空字符串。 */
  description: string;
  trackCount: number;
  albumCount: number;
};

export type Playlist = {
  id: number;
  name: string;
  description: string;
  cover: Cover | null;
  trackCount: number;
  durationMs: number;
  createdAt: string;
  updatedAt: string;
};

/** 新建与修改歌单的内容：名称 1–100 个字符，简介最多 1000 个字符。 */
export type PlaylistInput = {
  name: string;
  description: string;
};

export type HistoryEntry = {
  track: Track;
  playedAt: string;
};

/** 歌词原文与格式；没有歌词时都是空字符串。解析由播放端完成。 */
export type Lyrics = {
  format: "lrc" | "ttml" | "";
  lyrics: string;
};

/** 歌曲列表的排序字段，与 SPlayer 的 `sortFieldOptions` 同名。 */
export type SortField =
  "default" | "title" | "artist" | "album" | "duration" | "createTime";

export type SortOrder = "asc" | "desc";

/** 列表的分页、搜索与排序参数，省略时使用服务端默认值。 */
export type ListQuery = {
  q?: string;
  offset?: number;
  limit?: number;
  sort?: SortField;
  order?: SortOrder;
};

export type TrackQuery = ListQuery & {
  albumId?: number;
  artistId?: number;
};

export type AlbumQuery = Omit<ListQuery, "sort"> & {
  /** 专辑列表只接受按标题（`default`）与入库时间排序。 */
  sort?: "default" | "createTime";
  artistId?: number;
  /** 不含标记为单曲的专辑（专辑页与首页用，Sonora 新增）。 */
  hideSingles?: boolean;
};

/** 登录的网易云账号，头像地址由网易云的图片服务提供。 */
export interface NeteaseProfile {
  userId: number;
  nickname: string;
  avatarUrl: string;
  /** 0 表示不是会员 */
  vipType: number;
}

/**
 * 当前用户的网易云账号状态：`disabled` 服务器没有配置接口服务；`signedOut` 没有登录；
 * `signedIn` 已登录；`expired` 登录已失效；`unavailable` 接口服务暂时不可用。
 */
export interface NeteaseStatus {
  state: "disabled" | "signedOut" | "signedIn" | "expired" | "unavailable";
  profile: NeteaseProfile | null;
}

/** 二维码登录的进度，`code` 为 800 过期、801 等待扫码、802 待确认、803 成功。 */
export interface NeteaseQRState {
  code: number;
  nickname: string;
  avatarUrl: string;
}
