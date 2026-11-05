import { ApiError, loginUrl, source } from "@/api/source";
import { isOnline } from "@/utils/track";
import type { Track } from "@/api/types";
import { useDataStore, useUserStore } from "@/stores";
import { debounce, isFunction, type DebouncedFunc } from "lodash-es";

// 照搬 SPlayer utils/auth.ts 中的喜欢与歌单写操作，写入由 Go 服务确认后才更新界面。

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
    const online = isOnline(song);
    const list = online ? "onlineSongs" : "songs";
    try {
      if (online)
        await (like ? source.onlineLike(id) : source.onlineUnlike(id));
      else if (like) await source.like(id);
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
    const likeList = dataStore.userLikeData[list];
    const exists = likeList.includes(id);
    if (like && !exists) {
      dataStore.userLikeData[list] = [id, ...likeList];
      window.$message.success("已添加到我喜欢的音乐");
    } else if (!like && exists) {
      dataStore.userLikeData[list] = likeList.filter((s) => s !== id);
      window.$message.success("已取消喜欢");
    } else if (like && exists) {
      window.$message.info("我喜欢的音乐中已存在该歌曲");
    }
  },
  300,
  { leading: true, trailing: false },
);

/**
 * 写入失败时的提示。会话失效时回到登录页；服务端或网络的说明直接展示，
 * 其他错误用 `fallback`
 */
const reportWriteError = (error: unknown, fallback: string) => {
  if (error instanceof ApiError && error.status === 401) {
    location.assign(loginUrl);
    return;
  }
  window.$message.error(error instanceof ApiError ? error.message : fallback);
  console.error("❌", fallback, error);
};

/**
 * 从歌单中删除歌曲（照搬 SPlayer 的确认与提示），成功后调用 `callback`
 * @param pid 歌单 ID
 * @param ids 歌曲 ID
 */
export const deleteSongs = async (
  pid: number,
  ids: number[],
  options: { callback?: () => void; songName?: string } = {},
) => {
  const { callback, songName } = options;
  window.$dialog.warning({
    title: "删除歌曲",
    content:
      ids?.length > 1
        ? "确定删除这些选中的歌曲吗？"
        : songName
          ? `确定删除歌曲 ${songName} 吗？`
          : "确定删除这个歌曲吗？",
    positiveText: "删除",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        for (const id of ids) await source.removeFromPlaylist(pid, id);
      } catch (error) {
        reportWriteError(error, "删除歌曲失败，请重试");
        return;
      }
      if (isFunction(callback)) callback();
      window.$message.success("删除成功");
      // 歌单的封面与数量可能变化
      await useUserStore().loadPlaylists();
    },
  });
};

/**
 * 清空最近播放。确认对话框在页面中
 * @returns 是否已清空
 */
export const clearHistory = async (): Promise<boolean> => {
  try {
    await source.clearHistory();
    return true;
  } catch (error) {
    reportWriteError(error, "清空最近播放失败，请重试");
    return false;
  }
};

/**
 * 删除歌单（照搬 SPlayer 的确认与提示），成功后调用 `callback`
 * @param id 歌单 ID
 */
export const deletePlaylist = async (id: number, callback?: () => void) => {
  window.$dialog.warning({
    title: "删除歌单",
    content: "确认删除这个歌单？该操作无法撤销！",
    positiveText: "删除",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await source.deletePlaylist(id);
      } catch (error) {
        reportWriteError(error, "删除歌单失败，请重试");
        return;
      }
      window.$message.success("歌单删除成功");
      await useUserStore().loadPlaylists();
      if (isFunction(callback)) callback();
    },
  });
};

// 顺序按拖动的先后依次保存，避免较早的请求后到、覆盖较新的顺序。
let orderSaving: Promise<unknown> = Promise.resolve();

/**
 * 保存歌单顺序
 * @param id 歌单 ID
 * @param trackIds 歌单的全部歌曲的新顺序
 * @returns 是否已保存；失败时页面应重新读取歌单
 */
export const savePlaylistOrder = (
  id: number,
  trackIds: number[],
): Promise<boolean> => {
  const saving = orderSaving.then(async () => {
    try {
      await source.setPlaylistOrder(id, trackIds);
      return true;
    } catch (error) {
      if (error instanceof ApiError && error.status === 409) {
        window.$message.warning("歌单已在别处修改，已重新载入");
      } else {
        reportWriteError(error, "保存排序失败，请重试");
      }
      return false;
    }
  });
  orderSaving = saving;
  return saving;
};
