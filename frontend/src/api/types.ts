// 播放端 API 的数据形态，见 docs/api.md。字段与服务端响应一致，不含文件路径，也不含网易字段
// （本地专辑与艺术家关联的在线 ID 叫 `onlineId`）。

export type User = {
  name: string;
  email: string;
  /** 是否可以把在线歌曲导入曲库（指定 Authelia 组的成员） */
  canImport: boolean;
};

/** 封面与艺术家图片：内容哈希文件名与 ThumbHash 占位。 */
export type Cover = {
  name: string;
  thumbhash: string;
  /** 在线内容的图片地址；有值时 `name` 与 `thumbhash` 为空 */
  url?: string;
};

/** 内容来自网易云；本地内容没有这个字段。 */
export type Source = "online";

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
  source?: Source;
  /** 在线歌曲是否可以播放；本地歌曲没有这个字段 */
  playable?: boolean;
  /** 本地歌曲关联的在线 ID，没有时为 null；在线歌曲没有这个字段 */
  onlineId?: number | null;
  /**
   * 音量均衡的增益（dB），服务端按歌曲响度算好，见 docs/api.md 的“音量均衡”；
   * 尚未测量或无法测量时为 null，在线歌曲没有这个字段
   */
  gainDb?: number | null;
};

/** 从关联的在线内容补全了哪些字段。 */
export type FillField = "image" | "description" | "lyrics";

/** 漫游 FM 排好的一张专辑，歌曲按播放顺序。 */
export type FmAlbum = {
  id: number;
  tracks: Track[];
};

/** 漫游 FM：当前专辑、接下来的专辑、继续收听的歌曲与分组歌单（按优先顺序）。 */
export type FmState = {
  albums: FmAlbum[];
  trackId: number | null;
  groupIds: number[];
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
  source?: Source;
  /** 本地专辑关联的在线专辑（导入或管理后台填写），没有时为 null；在线专辑没有这个字段 */
  onlineId?: number | null;
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
  source?: Source;
  /** 本地艺术家关联的在线艺术家，没有时为 null；在线艺术家没有这个字段 */
  onlineId?: number | null;
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
  /** 隐私歌单，只有在线歌单有 */
  private?: boolean;
  source?: Source;
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
  /** 不含没有标记为单曲的专辑里的歌曲，没有专辑的歌曲保留。 */
  onlySingles?: boolean;
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
  /** 网易云网页上歌曲、专辑、艺术家与歌单的地址模板，`{id}` 换成 ID；服务器没有配置网易云时没有 */
  links?: Partial<Record<OnlineLinkType, string>>;
}

/** 有网页地址的在线内容。 */
export type OnlineLinkType = "song" | "album" | "artist" | "playlist";

/** 二维码登录的进度，`code` 为 800 过期、801 等待扫码、802 待确认、803 成功。 */
export interface NeteaseQRState {
  code: number;
  nickname: string;
  avatarUrl: string;
}

/**
 * 本地与在线歌曲混合的列表。`onlineFailed` 表示在线部分没有读到（登录失效或接口服务不可用），
 * 列表只有本地歌曲。
 */
export type MixedPage<T> = Page<T> & { onlineFailed?: boolean };

/** 在线列表的种类，对应同一个详情页。 */
export type OnlineListType = "playlist" | "album" | "artist";

/** 在线歌单、专辑或艺术家的详情，都以歌单的形态返回。 */
export type OnlineList = {
  playlist: Playlist;
  tracks: Track[];
};

/** 导入在线专辑前的预览，见 docs/api.md。 */
export type AlbumImportPreview = {
  title: string;
  total: number;
  /** 可以播放、曲库中还没有的歌曲 */
  toImport: number;
  /** 已导入过的歌曲；其中 `moving` 首还在单曲导入时的专辑里，会移入这张专辑 */
  inLibrary: number;
  moving: number;
  unplayable: number;
  /** 歌曲并入的曲库专辑；为 null 时新建 */
  target: Album | null;
  /** 曲库中其他同名专辑的数量，导入后不会自动合并 */
  sameTitle: number;
};

/** 专辑导入任务。`state` 为 failed 时提前中止，`error` 说明原因。 */
export type ImportJob = {
  id: number;
  albumId: number;
  title: string;
  state: "queued" | "running" | "done" | "failed";
  total: number;
  done: number;
  created: number;
  moved: number;
  existing: number;
  unplayable: number;
  failed: { title: string; reason: string }[];
  error: string;
};
