import type { ArtistRef, SortField, SortOrder, Track } from "@/api/types";

type TrackIdentity = Pick<Track, "id" | "source">;

/** 是否为在线歌曲。 */
export const isOnline = (track: { source?: string }): boolean =>
  track.source === "online";

/** 在线歌曲是否可以播放。本地歌曲没有这个字段，总是可以。 */
export const isPlayable = (track: { playable?: boolean }): boolean =>
  track.playable !== false;

/** 是否显示“导入到曲库”：可以播放的在线歌曲，且用户有导入权限。 */
export const canImportSong = (
  track: { source?: string; playable?: boolean },
  user: { canImport: boolean } | null,
): boolean => !!user?.canImport && isOnline(track) && isPlayable(track);

/**
 * 列表的来源标记（见 music store 的 playPlaylistId）。在线列表的 ID 可能与本地的相同，
 * 所以带前缀，与在线详情页一致。
 */
export const listKey = (
  type: "playlist" | "album" | "artist",
  item: { id: number; source?: string },
): string =>
  isOnline(item) ? `online-${type}-${item.id}` : `${type}-${item.id}`;

/** 歌单或专辑的详情页。 */
export const listRoute = (
  type: "playlist" | "album",
  item: { id: number; source?: string },
) =>
  isOnline(item)
    ? { name: "online-list", query: { type, id: item.id } }
    : { name: type, query: { id: item.id } };

/**
 * 歌曲的身份。本地与网易云的 ID 都是整数，可能相同，比较歌曲时不能只比较 `id`。
 */
export const trackKey = (track: TrackIdentity): string =>
  isOnline(track) ? `online:${track.id}` : String(track.id);

export const sameTrack = (a: TrackIdentity, b: TrackIdentity): boolean =>
  a.id === b.id && isOnline(a) === isOnline(b);

const collator = new Intl.Collator("zh-CN", { numeric: true });
const firstArtist = (artists: ArtistRef[]) => artists[0]?.name ?? "";

/**
 * 在浏览器中排序混合列表。在线歌曲不在数据库中，服务端的排序不包含它们；规则与服务端相同
 * （对应 SPlayer 的 `Intl.Collator`）。在线歌曲没有入库时间，`createTime` 保持列表顺序。
 */
export function sortTracks(
  list: readonly Track[],
  field: SortField,
  order: SortOrder | "default",
): Track[] {
  if (field === "default" || order === "default" || field === "createTime")
    return [...list];
  const sign = order === "desc" ? -1 : 1;
  const compare: Record<string, (a: Track, b: Track) => number> = {
    title: (a, b) => collator.compare(a.title, b.title),
    artist: (a, b) =>
      collator.compare(firstArtist(a.artists), firstArtist(b.artists)),
    album: (a, b) =>
      collator.compare(a.album?.title ?? "", b.album?.title ?? ""),
    duration: (a, b) => a.durationMs - b.durationMs,
  };
  return [...list].sort((a, b) => sign * compare[field](a, b));
}

/** 歌曲的艺术家的详情页。 */
export const artistRoute = (song: Track, id: number) =>
  isOnline(song)
    ? { name: "online-list", query: { type: "artist", id } }
    : { name: "artist", query: { id } };

/** 歌曲所属专辑的详情页。 */
export const albumRoute = (song: Track) =>
  isOnline(song)
    ? { name: "online-list", query: { type: "album", id: song.album?.id } }
    : { name: "album", query: { id: song.album?.id } };
