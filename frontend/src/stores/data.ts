import { defineStore } from "pinia";
import type { Track } from "@/api/types";

/**
 * 播放列表与喜爱的歌曲。字段名照搬 SPlayer stores/data.ts 中播放器用到的部分。
 * 队列保存在浏览器中，阶段 4.4 接入；喜爱来自 `likes/ids`。
 */
interface DataState {
  /** 播放列表 */
  playList: Track[];
  /** 用户喜欢的数据；Sonora 只有歌曲 */
  userLikeData: {
    songs: number[];
  };
}

export const useDataStore = defineStore("data", {
  state: (): DataState => ({
    playList: [],
    userLikeData: { songs: [] },
  }),
  getters: {
    // 是否为喜欢的歌曲
    isLikeSong: (state) => (id: number) =>
      state.userLikeData.songs.includes(id),
  },
});
