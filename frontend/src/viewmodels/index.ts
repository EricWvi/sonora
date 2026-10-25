import type { Album, Artist, Cover, Playlist, Track } from "@/api/types";

// 页面与播放器组件展示用的数据。字段沿用播放端 API 的形态（docs/api.md）：
// `artists` 是数组，`cover` 是 `{name, thumbhash}`，不含文件路径与网易字段。

/**
 * 歌曲列表的一行：序号、标题（封面与艺术家）、专辑、喜欢、时长。
 * 序号是行在列表中的位置，由列表决定。
 */
export type SongRow = Track & {
  liked: boolean;
};

export function songRows(
  tracks: readonly Track[],
  likedIds: ReadonlySet<number>,
): SongRow[] {
  return tracks.map((track) => ({ ...track, liked: likedIds.has(track.id) }));
}

/** 专辑、艺术家、歌单网格中的一张封面卡片。 */
export type CoverCard =
  | { type: "album"; album: Album }
  | { type: "artist"; artist: Artist }
  | { type: "playlist"; playlist: Playlist };

/**
 * 详情页头部（沿用 SPlayer ListDetail.vue 的样式），内容按阶段 2 的“详情页头部”表：
 *
 * - 专辑：标题、艺术家、年份、歌曲数、总时长；
 * - 歌单：名称、简介、歌曲数、创建与更新时间；
 * - 艺术家：名称、歌曲数与专辑数；
 * - 我喜欢的音乐：歌曲数。
 */
export type DetailHeader =
  | { type: "album"; album: Album }
  | { type: "playlist"; playlist: Playlist }
  | { type: "artist"; artist: Artist }
  | { type: "likes"; trackCount: number };

/** 卡片与详情头部显示的图片；为 null 时由组件换成 SPlayer 的默认图片。 */
export function coverOf(item: CoverCard | DetailHeader): Cover | null {
  switch (item.type) {
    case "album":
      return item.album.cover;
    case "artist":
      return item.artist.image;
    case "playlist":
      return item.playlist.cover;
    case "likes":
      return null;
  }
}

/**
 * 播放器状态：队列、当前曲目、播放位置与是否在播放。
 * 队列与位置保存在浏览器中；开发时由 fixture 指定（`?player=`）。
 */
export type PlayerState = {
  queue: Track[];
  /** 当前曲目在队列中的位置；队列为空时为 -1。 */
  index: number;
  playing: boolean;
  positionMs: number;
};

/** 当前曲目；队列为空时为 null。 */
export function currentTrack(state: PlayerState): Track | null {
  return state.queue[state.index] ?? null;
}
