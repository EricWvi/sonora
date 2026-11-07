import { ref } from "vue";
import type { FmState, Track } from "@/api/types";

/** 私人 FM 的列表来源（见 music store 的 playPlaylistId） */
export const FM_LIST = "fm";

/** 队列中保留的已播条目数：FM 没有尽头，队列不能一直变长 */
export const FM_HISTORY = 200;

/** 分组歌单变化时加一，卡片据此重新读取 */
export const fmRevision = ref(0);

/** 带所属专辑标记的歌曲：标记区分 FM 排的歌与用户插入的“下一首播放” */
export type FmTrack = Track & { fmAlbumId: number };

/** FM 状态中所有专辑的歌曲，按播放顺序 */
export const fmTracks = (state: FmState): FmTrack[] =>
  state.albums.flatMap((album) =>
    album.tracks.map((track) => ({ ...track, fmAlbumId: album.id })),
  );

/** 继续收听的那首歌，没有记录时是当前专辑的第一首 */
export const fmCurrentTrack = (state: FmState): FmTrack | undefined => {
  const tracks = fmTracks(state);
  return tracks.find((track) => track.id === state.trackId) ?? tracks[0];
};

/**
 * 要接到队列后面的歌曲：当前条目之后已有的专辑不再加入；当前专辑从继续收听的那首开始
 * @param state FM 状态
 * @param queuedAlbums 当前条目之后已在队列中的专辑
 */
export function fmTracksToAppend(
  state: FmState,
  queuedAlbums: ReadonlySet<number>,
): FmTrack[] {
  const head = state.albums[0]?.id;
  return fmTracks(state)
    .filter((track) => !queuedAlbums.has(track.fmAlbumId))
    .filter((track, _, tracks) => {
      if (track.fmAlbumId !== head || state.trackId == null) return true;
      const start = tracks.findIndex((t) => t.id === state.trackId);
      return start < 0 || tracks.indexOf(track) >= start;
    });
}
