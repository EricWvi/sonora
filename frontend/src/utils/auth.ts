import { ApiError, loginUrl, source } from "@/api/source";
import type { Track } from "@/api/types";
import { useDataStore } from "@/stores";
import { debounce, type DebouncedFunc } from "lodash-es";

// 照搬 SPlayer utils/auth.ts 中的喜欢入口；其余写操作在阶段 5 接入，之前只提示。

/**
 * 喜欢或取消喜欢歌曲。服务端确认后才更新喜爱状态，失败时保持原样（路线图 5.1）
 * @param song 歌曲
 * @param like 是否喜欢
 */
export const toLikeSong: DebouncedFunc<
  (song: Track, like: boolean) => Promise<void>
> = debounce(
  async (song: Track, like: boolean): Promise<void> => {
    const dataStore = useDataStore();
    const { id } = song;
    try {
      if (like) await source.like(id);
      else await source.unlike(id);
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        location.assign(loginUrl);
        return;
      }
      window.$message.error(`${like ? "喜欢" : "取消"}音乐时发生错误`);
      console.error("❌ 更新喜欢歌曲时失败:", error);
      return;
    }
    // 请求期间可能有其他修改，按确认时的列表更新
    const likeList = dataStore.userLikeData.songs;
    const exists = likeList.includes(id);
    if (like && !exists) {
      dataStore.userLikeData.songs = [id, ...likeList];
      window.$message.success("已添加到我喜欢的音乐");
    } else if (!like && exists) {
      dataStore.userLikeData.songs = likeList.filter((s) => s !== id);
      window.$message.success("已取消喜欢");
    } else if (like && exists) {
      window.$message.info("我喜欢的音乐中已存在该歌曲");
    }
  },
  300,
  { leading: true, trailing: false },
);

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
