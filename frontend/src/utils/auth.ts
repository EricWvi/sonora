import type { Track } from "@/api/types";

// 照搬 SPlayer utils/auth.ts 中的喜欢入口。写入服务端在阶段 4.4 接入，之前只提示。

/**
 * 喜欢或取消喜欢歌曲
 * @param _song 歌曲
 * @param _like 是否喜欢
 */
export const toLikeSong = async (_song: Track, _like: boolean) => {
  window.$message.info("喜欢将在后续阶段接入");
};

/**
 * 从歌单中删除歌曲。写入服务端在阶段 5 接入，之前只提示
 * @param _playlistId 歌单 ID
 * @param _ids 歌曲 ID
 */
export const deleteSongs = async (_playlistId: number, _ids: number[]) => {
  window.$message.info("从歌单中删除将在后续阶段接入");
};

/** 清空最近播放。写入服务端在阶段 5 接入，之前只提示 */
export const clearHistory = async (): Promise<boolean> => {
  window.$message.info("清空最近播放将在后续阶段接入");
  return false;
};

/**
 * 删除歌单。确认与写入在阶段 5 接入，之前只提示
 * @param _id 歌单 ID
 */
export const deletePlaylist = async (_id: number) => {
  window.$message.info("删除歌单将在后续阶段接入");
};

/**
 * 保存歌单顺序。写入服务端在阶段 5 接入（冲突时刷新），之前只提示
 * @param _id 歌单 ID
 * @param _trackIds 歌单的全部歌曲的新顺序
 */
export const savePlaylistOrder = async (_id: number, _trackIds: number[]) => {
  window.$message.info("保存歌单顺序将在后续阶段接入");
};
