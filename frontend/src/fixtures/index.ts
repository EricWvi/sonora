import { ApiError } from "@/api/error";
import type { DataSource } from "@/api/source";
import type {
  Album,
  Artist,
  ArtistRef,
  FmState,
  HistoryEntry,
  ListQuery,
  Lyrics,
  Page,
  Playlist,
  Track,
  User,
} from "@/api/types";
import type { PlayerState } from "@/viewmodels";

// 开发用 fixture，以 `?fixture=<场景>` 打开，`&player=<状态>` 指定播放器的初始状态。
// 数据与服务端开发模式的用户一致。`library` 的内容是虚构的；`baseline` 是阶段 2
// 播放器基线的歌曲，含网易内容，只在本机由 scripts/fixtures/baseline.mjs 生成到
// .local/fixtures/baseline/，不进入仓库。

const user: User = {
  name: "listener",
  email: "listener@sonora.test",
  canImport: false,
};

/** 播放器状态在 fixture 中以歌曲 ID 记录队列。 */
type PlayerFixture = {
  queue: number[];
  index: number;
  positionMs: number;
  playing: boolean;
  durationMs?: number;
};

type Library = {
  tracks: Track[];
  albums: Album[];
  artists: Artist[];
  playlists: { playlist: Playlist; trackIds: number[] }[];
  likes: number[];
  history: { trackId: number; playedAt: string }[];
  lyrics: Record<number, Lyrics>;
  coverUrls: Record<string, string>;
  players: Record<string, PlayerFixture>;
};

// --- library：虚构曲库 -------------------------------------------------------------

function fictionalLibrary(): Library {
  const artists: Artist[] = [
    "晨雾乐队",
    "Aurora Lane",
    "林间",
    "The Long Name Orchestra of Somewhere Very Far Away",
    "周末",
  ].map((name, i) => ({
    id: i + 1,
    name,
    image: null,
    // 第一位艺术家带简介，用来查看艺术家页的简介行与弹窗
    description: i ? "" : "开发数据的歌手简介。\n第二行用来确认弹窗保留换行。",
    trackCount: 0,
    albumCount: 0,
  }));
  const ref = (id: number): ArtistRef => ({ id, name: artists[id - 1].name });

  const albums: Album[] = [];
  const tracks: Track[] = [];
  // 时长按序号生成，保证每次相同。
  const duration = (n: number) => 150_000 + ((n * 37_123) % 150_000);
  const addTrack = (
    title: string,
    artistIds: number[],
    album: Album | null,
    trackNo: number,
    discNo = 1,
  ) => {
    const track: Track = {
      id: tracks.length + 1,
      title,
      artists: artistIds.map(ref),
      album: album && {
        id: album.id,
        title: album.title,
        artists: album.artists,
      },
      trackNo,
      discNo,
      year: album?.year ?? 2020,
      durationMs: duration(tracks.length + 1),
      format: "mp3",
      cover: album?.cover ?? null,
      hasLyrics: false,
    };
    tracks.push(track);
    if (album) {
      album.trackCount++;
      album.durationMs += track.durationMs;
    }
    for (const id of new Set([
      ...artistIds,
      ...(album?.artists.map((a) => a.id) ?? []),
    ]))
      artists[id - 1].trackCount++;
  };
  const addAlbum = (
    title: string,
    artistIds: number[],
    year: number,
    discs: string[][],
    trackArtists?: number[][],
  ) => {
    const album: Album = {
      id: albums.length + 1,
      title,
      artists: artistIds.map(ref),
      year,
      cover: null,
      // 第一张专辑带简介，用来查看专辑页的简介行与弹窗
      description: albums.length
        ? ""
        : "开发数据的专辑简介。\n第二行用来确认弹窗保留换行。",
      trackCount: 0,
      durationMs: 0,
    };
    albums.push(album);
    for (const id of artistIds) artists[id - 1].albumCount++;
    let n = 0;
    discs.forEach((titles, disc) =>
      titles.forEach((title, i) =>
        addTrack(
          title,
          trackArtists?.[n++] ?? artistIds,
          album,
          i + 1,
          disc + 1,
        ),
      ),
    );
  };

  addAlbum("海边的车站", [1], 2024, [
    [
      "清晨的第一班车",
      "海风",
      "站台",
      "旧明信片",
      "潮汐表",
      "灯塔",
      "回程",
      "夜航",
    ],
  ]);
  addAlbum("Northbound", [2], 2023, [
    ["Departure", "Pine Road", "Snowline", "Lantern", "Quiet Hours", "Polaris"],
    ["Afterglow", "Northern Sky", "Homeward", "Coda"],
  ]);
  addAlbum("雨季", [3], 2022, [
    ["雨季", "窗", "青苔", "伞下", "六月", "积水", "晴"],
  ]);
  addAlbum(
    "合作单曲集",
    [1, 3],
    2025,
    [["山与海", "同一条街", "Meet Me Halfway", "尾声"]],
    [
      [1, 3],
      [3, 1],
      [1, 2, 3],
      [1, 3],
    ],
  );
  addAlbum("一张标题很长很长很长很长很长很长很长很长的专辑", [4], 2021, [
    [
      "Overture for a Very Long Afternoon Spent Waiting by the Window",
      "Intermezzo",
      "Finale",
    ],
  ]);
  addTrack("没有专辑的歌", [5], null, 0);
  addTrack("Demo 2026-10-01", [5], null, 0);

  const lyrics: Record<number, Lyrics> = {
    1: {
      format: "lrc",
      lyrics: [
        "[00:01.000]第一班车还没有来",
        "[00:01.000]The first bus has not come yet",
        "[00:06.500]站台上只有风",
        "[00:06.500]Only the wind on the platform",
        "[00:12.000]我数着远处的灯",
        "[00:12.000]I count the lights far away",
        "[00:18.000]",
      ].join("\n"),
    },
    9: {
      format: "lrc",
      lyrics: [
        "[00:02.000]Leaving before the sun",
        "[00:08.000]Down the pine road",
        "[00:14.000]",
      ].join("\n"),
    },
  };
  for (const id of Object.keys(lyrics)) tracks[Number(id) - 1].hasLyrics = true;

  const playlistFixture = (
    id: number,
    name: string,
    description: string,
    trackIds: number[],
    day: number,
  ) => {
    const list = trackIds.map((t) => tracks[t - 1]);
    return {
      playlist: {
        id,
        name,
        description,
        cover: list.find((t) => t.cover)?.cover ?? null,
        trackCount: list.length,
        durationMs: list.reduce((sum, t) => sum + t.durationMs, 0),
        createdAt: `2026-10-${day}T12:00:00Z`,
        updatedAt: `2026-10-${day + 2}T08:30:00Z`,
      },
      trackIds,
    };
  };

  return {
    tracks,
    albums,
    artists,
    playlists: [
      playlistFixture(4, "通勤路上", "早晚各一小时", [1, 2, 3, 9, 10], 20),
      playlistFixture(3, "夜跑", "", [12, 13, 28], 19),
      playlistFixture(2, "写代码时听", "没有人声或人声很少的歌", [], 18),
      playlistFixture(1, "一个名字很长很长很长很长很长很长的歌单", "", [5], 17),
    ],
    likes: [12, 3, 20, 7],
    history: [5, 1, 19, 28, 9].map((trackId, i) => ({
      trackId,
      playedAt: `2026-10-24T${String(21 - i).padStart(2, "0")}:15:00Z`,
    })),
    lyrics,
    coverUrls: {},
    players: {
      paused: {
        queue: [1, 2, 3, 4, 5, 6, 7, 8],
        index: 0,
        positionMs: 7_000,
        playing: false,
      },
      playing: {
        queue: [1, 2, 3, 4, 5, 6, 7, 8],
        index: 0,
        positionMs: 7_000,
        playing: true,
      },
    },
  };
}

// --- baseline：阶段 2 播放器基线的歌曲 ------------------------------------------------

// 路径在仓库根目录的 .local 下；文件不存在时两个 glob 都为空。
const baselineData = import.meta.glob<BaselineFile>(
  "../../../.local/fixtures/baseline/library.json",
  { import: "default" },
);
const baselineCovers = import.meta.glob<string>(
  "../../../.local/fixtures/baseline/covers/*",
  { query: "?url", import: "default", eager: true },
);

/** scripts/fixtures/baseline.mjs 的输出。 */
type BaselineFile = {
  tracks: Track[];
  albums: Album[];
  artists: Artist[];
  lyrics: Record<number, Lyrics>;
  /** 基线状态开始时的队列与位置：yoru 为《夜に駆ける》，guren 为《紅蓮華》。 */
  players: Record<"yoru" | "guren", Omit<PlayerFixture, "playing">>;
};

async function baselineLibrary(): Promise<Library> {
  const load = Object.values(baselineData)[0];
  if (!load) {
    throw new Error(
      "没有 baseline 数据：先运行 node scripts/fixtures/baseline.mjs（需要 .local 中的阶段 2 基线）",
    );
  }
  const data = await load();
  const coverUrls: Record<string, string> = {};
  for (const [path, url] of Object.entries(baselineCovers)) {
    coverUrls[path.slice(path.lastIndexOf("/") + 1)] = url;
  }
  const { yoru, guren } = data.players;
  return {
    tracks: data.tracks,
    albums: data.albums,
    artists: data.artists,
    playlists: [],
    // 参考账号没有喜欢这两首歌，基线中的喜欢按钮是空心的。
    likes: [],
    history: [],
    lyrics: data.lyrics,
    coverUrls,
    players: {
      // 对应基线的 bar-paused-start
      "yoru-start": { ...yoru, positionMs: 0, playing: false },
      "yoru-paused": { ...yoru, playing: false },
      "yoru-playing": { ...yoru, playing: true },
      "guren-paused": { ...guren, playing: false },
      "guren-playing": { ...guren, playing: true },
    },
  };
}

// --- 查询 -----------------------------------------------------------------------------

// 与服务端一致：名称按拼音、数字按数值比较。
const collator = new Intl.Collator("zh-CN", { numeric: true });

const matches = (q: string | undefined, ...texts: string[]) =>
  !q || texts.some((text) => text.toLowerCase().includes(q.toLowerCase()));

/**
 * 搜索时按相关度排序，与服务端相同：`tiers` 从重要到次要分组，组内完全相同在前，
 * 其次是开头相同，再次是包含；同一档保持列表原来的顺序。
 */
function byRelevance<T>(
  list: T[],
  q: string | undefined,
  tiers: (item: T) => string[][],
): T[] {
  if (!q) return list;
  const key = q.toLowerCase();
  const rank = (item: T) => {
    const groups = tiers(item);
    for (const [i, texts] of groups.entries()) {
      const lower = texts.map((text) => text.toLowerCase());
      if (lower.some((text) => text === key)) return i * 3;
      if (lower.some((text) => text.startsWith(key))) return i * 3 + 1;
      if (lower.some((text) => text.includes(key))) return i * 3 + 2;
    }
    return groups.length * 3;
  };
  return list
    .map((item, index) => ({ item, index, rank: rank(item) }))
    .sort((a, b) => a.rank - b.rank || a.index - b.index)
    .map(({ item }) => item);
}

/** 按 `sort` 与 `order` 排序；`default` 保持列表自身的顺序，相等的项也保持列表顺序。 */
function sortTracks(list: Track[], query: ListQuery = {}): Track[] {
  const { sort = "default", order = "asc" } = query;
  if (sort === "default") return list;
  const key: Record<
    Exclude<typeof sort, "default">,
    (a: Track, b: Track) => number
  > = {
    title: (a, b) => collator.compare(a.title, b.title),
    artist: (a, b) =>
      collator.compare(a.artists[0]?.name ?? "", b.artists[0]?.name ?? ""),
    album: (a, b) =>
      collator.compare(a.album?.title ?? "", b.album?.title ?? ""),
    duration: (a, b) => a.durationMs - b.durationMs,
    // fixture 中歌曲 ID 的顺序即入库顺序。
    createTime: (a, b) => a.id - b.id,
  };
  const sign = order === "desc" ? -1 : 1;
  return [...list].sort((a, b) => sign * key[sort](a, b));
}

function page<T>(items: T[], query: ListQuery = {}): Page<T> {
  const { offset = 0, limit = 50 } = query;
  return { items: items.slice(offset, offset + limit), total: items.length };
}

const notFound = () => new ApiError(404, "内容不存在", "not_found");

function librarySource(lib: Library): DataSource {
  // 设置只存在内存中，从未保存时为 {}（与服务端相同）
  let settings: Record<string, unknown> = {};
  const track = (id: number) => {
    const found = lib.tracks.find((t) => t.id === id);
    if (!found) throw notFound();
    return found;
  };
  const ownPlaylist = (id: number) => {
    const found = lib.playlists.find((p) => p.playlist.id === id);
    if (!found) throw notFound();
    return found;
  };
  // 歌单内容变化后，按服务端的规则重算封面、数量与时长。
  const touch = (found: Library["playlists"][number]) => {
    const list = found.trackIds.map(track);
    Object.assign(found.playlist, {
      cover: list.find((t) => t.cover)?.cover ?? null,
      trackCount: list.length,
      durationMs: list.reduce((sum, t) => sum + t.durationMs, 0),
      updatedAt: new Date().toISOString(),
    });
    return found.playlist;
  };
  // 歌曲默认新入库在前。
  const newestFirst = () => [...lib.tracks].reverse();
  // 只算艺术家演唱的歌曲，不含其专辑中的其他歌曲（与服务端相同）
  const byArtist = (t: Track, id: number) => t.artists.some((a) => a.id === id);

  // 返回副本，页面修改结果不影响 fixture。
  // 漫游 FM：按专辑依次轮转，真实的排法在服务端
  const fmAlbums = () =>
    lib.albums
      .map((album) => ({
        id: album.id,
        tracks: lib.tracks.filter((t) => t.album?.id === album.id),
      }))
      .filter((album) => album.tracks.length);
  let fmHead = 0;
  let fmTrackId: number | null = null;
  let fmGroupIds: number[] = [];
  const fmState = (): FmState => {
    const albums = fmAlbums();
    return {
      albums: albums.length
        ? [albums[fmHead % albums.length], albums[(fmHead + 1) % albums.length]]
        : [],
      trackId: fmTrackId,
      groupIds: fmGroupIds,
    };
  };
  const fmMoveOn = (albumId: number) => {
    if (fmState().albums[0]?.id !== albumId) return;
    fmHead++;
    fmTrackId = null;
  };

  const reply = async <T>(value: () => T): Promise<T> =>
    structuredClone(value());

  return {
    me: () => reply(() => user),
    // fixture 没有网易云接口服务，与服务器未配置时相同。
    neteaseStatus: () => reply(() => ({ state: "disabled", profile: null })),
    neteaseQRStart: () => Promise.reject(new ApiError(503, "没有配置")),
    neteaseQRCheck: () => Promise.reject(new ApiError(503, "没有配置")),
    neteaseCookieLogin: () => Promise.reject(new ApiError(503, "没有配置")),
    neteaseLogout: () => reply(() => undefined),
    // 没有在线部分时，混合列表就是本地列表。
    mixedLikes: (query) => reply(() => page(lib.likes.map(track), query)),
    mixedHistory: (query) =>
      reply(() =>
        page<HistoryEntry>(
          lib.history.map((e) => ({
            track: track(e.trackId),
            playedAt: e.playedAt,
          })),
          query,
        ),
      ),
    recordOnlinePlay: () => reply(() => undefined),
    onlinePlaylists: () => reply(() => ({ created: [], collected: [] })),
    onlineAlbums: () => reply(() => ({ items: [], total: 0 })),
    onlineArtists: () => reply(() => ({ items: [], total: 0 })),
    onlineDailySongs: () => reply(() => ({ items: [], total: 0 })),
    onlineList: () => Promise.reject(new ApiError(503, "没有配置")),
    onlineArtistAlbums: () => reply(() => ({ items: [], total: 0 })),
    createOnlinePlaylist: () => Promise.reject(new ApiError(503, "没有配置")),
    updateOnlinePlaylist: () => Promise.reject(new ApiError(503, "没有配置")),
    deleteOnlinePlaylist: () => Promise.reject(new ApiError(503, "没有配置")),
    addToOnlinePlaylist: () => Promise.reject(new ApiError(503, "没有配置")),
    removeFromOnlinePlaylist: () =>
      Promise.reject(new ApiError(503, "没有配置")),
    publishOnlinePlaylist: () => Promise.reject(new ApiError(503, "没有配置")),
    onlineSongUrl: () => Promise.reject(new ApiError(503, "没有配置")),
    onlineLyrics: () => reply(() => ({ format: "" as const, lyrics: "" })),
    importOnlineSong: () => Promise.reject(new ApiError(503, "没有配置")),
    previewAlbumImport: () => Promise.reject(new ApiError(503, "没有配置")),
    importOnlineAlbum: () => Promise.reject(new ApiError(503, "没有配置")),
    importJobs: () => reply(() => []),
    onlineLikeIds: () => reply(() => []),
    setOnlineSubscription: () => Promise.reject(new ApiError(503, "没有配置")),
    onlineLike: () => reply(() => undefined),
    onlineUnlike: () => reply(() => undefined),
    tracks: (query = {}) =>
      reply(() => {
        let list = newestFirst();
        if (query.albumId) {
          list = lib.tracks
            .filter((t) => t.album?.id === query.albumId)
            .sort((a, b) => a.discNo - b.discNo || a.trackNo - b.trackNo);
        }
        if (query.artistId)
          list = list.filter((t) => byArtist(t, query.artistId!));
        list = list.filter((t) =>
          matches(
            query.q,
            t.title,
            t.album?.title ?? "",
            ...t.artists.map((a) => a.name),
          ),
        );
        if (!query.sort || query.sort === "default")
          list = byRelevance(list, query.q, (t) => [
            [t.title],
            [t.album?.title ?? "", ...t.artists.map((a) => a.name)],
          ]);
        return page(sortTracks(list, query), query);
      }),
    track: (id) => reply(() => track(id)),
    lyrics: (trackId) =>
      reply(() => {
        track(trackId);
        return lib.lyrics[trackId] ?? { format: "", lyrics: "" };
      }),
    albums: (query = {}) =>
      reply(() => {
        let list = [...lib.albums].sort((a, b) =>
          collator.compare(a.title, b.title),
        );
        if (query.sort === "createTime") {
          const sign = query.order === "desc" ? -1 : 1;
          list.sort((a, b) => sign * (a.id - b.id));
        }
        if (query.artistId)
          list = list.filter((a) =>
            a.artists.some((r) => r.id === query.artistId),
          );
        list = list.filter((a) =>
          matches(query.q, a.title, ...a.artists.map((r) => r.name)),
        );
        if (query.sort !== "createTime")
          list = byRelevance(list, query.q, (a) => [
            [a.title],
            a.artists.map((r) => r.name),
          ]);
        return page(list, query);
      }),
    album: (id) =>
      reply(() => {
        const found = lib.albums.find((a) => a.id === id);
        if (!found) throw notFound();
        return found;
      }),
    artists: (query = {}) =>
      reply(() =>
        page(
          byRelevance(
            [...lib.artists]
              .sort((a, b) => collator.compare(a.name, b.name))
              .filter((a) => matches(query.q, a.name)),
            query.q,
            (a) => [[a.name]],
          ),
          query,
        ),
      ),
    artist: (id) =>
      reply(() => {
        const found = lib.artists.find((a) => a.id === id);
        if (!found) throw notFound();
        return found;
      }),
    // fixture 的用户没有导入权限，界面上没有这些入口
    fillAlbum: (id) =>
      reply(() => {
        const album = lib.albums.find((a) => a.id === id);
        if (!album) throw notFound();
        return { album, filled: [] };
      }),
    fillArtist: (id) =>
      reply(() => {
        const artist = lib.artists.find((a) => a.id === id);
        if (!artist) throw notFound();
        return { artist, filled: [] };
      }),
    fillTrack: (id) => reply(() => ({ track: track(id), filled: [] })),
    fm: () => reply(fmState),
    fmFinish: (albumId) =>
      reply(() => {
        fmMoveOn(albumId);
        return fmState();
      }),
    fmSkip: (albumId) =>
      reply(() => {
        fmMoveOn(albumId);
        return fmState();
      }),
    fmPosition: (trackId) =>
      reply(() => {
        fmTrackId = trackId;
      }),
    fmSetGroups: (playlistIds) =>
      reply(() => {
        fmGroupIds = playlistIds;
        return fmState();
      }),
    likes: (query) =>
      reply(() => page(sortTracks(lib.likes.map(track), query), query)),
    likeIds: () => reply(() => lib.likes),
    history: (query) =>
      reply(() =>
        page<HistoryEntry>(
          lib.history.map((e) => ({
            track: track(e.trackId),
            playedAt: e.playedAt,
          })),
          query,
        ),
      ),
    playlists: (query = {}) =>
      reply(() =>
        page(
          byRelevance(
            lib.playlists
              .map((p) => p.playlist)
              .filter((p) => matches(query.q, p.name)),
            query.q,
            (p) => [[p.name]],
          ),
        ),
      ),
    playlist: (id) => reply(() => ownPlaylist(id).playlist),
    playlistTracks: (id, query) =>
      reply(() =>
        page(sortTracks(ownPlaylist(id).trackIds.map(track), query), query),
      ),
    // 写操作只改内存中的数据，刷新页面后恢复原样。
    like: (trackId) =>
      reply(() => {
        track(trackId);
        if (!lib.likes.includes(trackId)) lib.likes.unshift(trackId);
      }),
    unlike: (trackId) =>
      reply(() => {
        lib.likes = lib.likes.filter((id) => id !== trackId);
      }),
    recordPlay: (trackId) =>
      reply(() => {
        track(trackId);
        lib.history = [
          { trackId, playedAt: new Date().toISOString() },
          ...lib.history.filter((e) => e.trackId !== trackId),
        ];
      }),
    clearHistory: () =>
      reply(() => {
        lib.history = [];
      }),
    createPlaylist: (input) =>
      reply(() => {
        const now = new Date().toISOString();
        const playlist: Playlist = {
          id: Math.max(0, ...lib.playlists.map((p) => p.playlist.id)) + 1,
          name: input.name,
          description: input.description,
          cover: null,
          trackCount: 0,
          durationMs: 0,
          createdAt: now,
          updatedAt: now,
        };
        lib.playlists.unshift({ playlist, trackIds: [] });
        return playlist;
      }),
    updatePlaylist: (id, input) =>
      reply(() => {
        const found = ownPlaylist(id);
        Object.assign(found.playlist, input);
        return touch(found);
      }),
    deletePlaylist: (id) =>
      reply(() => {
        ownPlaylist(id);
        lib.playlists = lib.playlists.filter((p) => p.playlist.id !== id);
      }),
    addToPlaylist: (id, trackIds) =>
      reply(() => {
        const found = ownPlaylist(id);
        trackIds.forEach(track);
        const fresh = [...new Set(trackIds)].filter(
          (t) => !found.trackIds.includes(t),
        );
        found.trackIds.unshift(...fresh);
        return { added: fresh.length, playlist: touch(found) };
      }),
    removeFromPlaylist: (id, trackId) =>
      reply(() => {
        const found = ownPlaylist(id);
        found.trackIds = found.trackIds.filter((t) => t !== trackId);
        return touch(found);
      }),
    setPlaylistOrder: (id, trackIds) =>
      reply(() => {
        const found = ownPlaylist(id);
        const same =
          trackIds.length === found.trackIds.length &&
          [...trackIds].sort().join() === [...found.trackIds].sort().join();
        if (!same)
          throw new ApiError(409, "歌单已在别处修改，请刷新后重试", "conflict");
        found.trackIds = [...trackIds];
        return touch(found);
      }),
    audioStatus: () => reply(() => 200),
    settings: () => reply(() => settings),
    saveSettings: (data) =>
      reply(() => {
        settings = structuredClone(data);
      }),
    logout: () => Promise.resolve(),
    // baseline 带有网易按 100、300、1024 缩放的图，没有对应尺寸时用原图。
    coverUrl: (cover, size) =>
      (size && lib.coverUrls[cover.name.replace(/(\.\w+)$/, `-${size}$1`)]) ||
      (lib.coverUrls[cover.name] ?? ""),
    // fixture 没有音频，播放器用不出声的引擎，地址只用来区分歌曲。
    audioUrl: (trackId) => `fixture:audio/${trackId}`,
  };
}

// --- 场景 -----------------------------------------------------------------------------

function emptyLibrary(): Library {
  return {
    tracks: [],
    albums: [],
    artists: [],
    playlists: [],
    likes: [],
    history: [],
    lyrics: {},
    coverUrls: {},
    players: {},
  };
}

/** 每个请求都以同样的方式结束：一直不返回，或服务不可达。 */
function failingSource(fail: () => Promise<never>): DataSource {
  const source = librarySource(emptyLibrary());
  for (const key of Object.keys(source) as (keyof DataSource)[]) {
    if (key === "logout" || key === "coverUrl" || key === "audioUrl") continue;
    (source[key] as () => Promise<never>) = fail;
  }
  return source;
}

const scenarios: Record<string, () => Promise<Library | DataSource>> = {
  /** 有数据（虚构） */
  library: async () => fictionalLibrary(),
  /** 阶段 2 播放器基线的歌曲：《THE BOOK》与《紅蓮華》 */
  baseline: baselineLibrary,
  /** 空曲库，没有歌单 */
  empty: async () => emptyLibrary(),
  /** 请求一直没有返回 */
  loading: async () => failingSource(() => new Promise(() => {})),
  /** 服务不可达 */
  unreachable: async () =>
    failingSource(() => Promise.reject(new ApiError(0, "无法连接服务器"))),
};

export async function createFixture(
  name: string,
  playerName: string | null,
): Promise<{ source: DataSource; player: PlayerState | null }> {
  const scenario = scenarios[name];
  if (!scenario) {
    throw new Error(
      `未知的 fixture 场景 "${name}"，可用：${Object.keys(scenarios).join("、")}`,
    );
  }
  const data = await scenario();
  if (!("players" in data)) {
    if (playerName) throw new Error(`fixture 场景 "${name}" 没有播放器状态`);
    return { source: data, player: null };
  }
  if (!playerName) return { source: librarySource(data), player: null };
  const player = data.players[playerName];
  if (!player) {
    const names = Object.keys(data.players);
    throw new Error(
      `fixture 场景 "${name}" 没有播放器状态 "${playerName}"` +
        (names.length ? `，可用：${names.join("、")}` : ""),
    );
  }
  return {
    source: librarySource(data),
    player: {
      queue: player.queue.map((id) => {
        const track = data.tracks.find((t) => t.id === id);
        if (!track) throw new Error(`播放器队列中的歌曲 ${id} 不存在`);
        return structuredClone(track);
      }),
      index: player.index,
      playing: player.playing,
      positionMs: player.positionMs,
      durationMs: player.durationMs,
    },
  };
}
