import { defineStore } from "pinia";
import type { Track } from "@/api/types";

/**
 * 播放列表与喜爱的歌曲。字段名照搬 SPlayer stores/data.ts 中播放器用到的部分。
 * 队列保存在浏览器中，阶段 4.4 接入；喜爱来自 `likes/ids`。
 */
interface DataState {
  /** 播放列表 */
  playList: Track[];
  /** 原始播放列表（随机播放前的顺序） */
  originalPlayList: Track[];
  /** 用户喜欢的数据；Sonora 只有歌曲 */
  userLikeData: {
    songs: number[];
  };
}

export const useDataStore = defineStore("data", {
  state: (): DataState => ({
    playList: [],
    originalPlayList: [],
    userLikeData: { songs: [] },
  }),
  getters: {
    // 是否为喜欢的歌曲
    isLikeSong: (state) => (id: number) =>
      state.userLikeData.songs.includes(id),
  },
  actions: {
    /**
     * 更新播放列表。SPlayer 同时写入 IndexedDB，Sonora 在阶段 4.4 保存队列
     * @param data 播放列表
     */
    async setPlayList(data: Track[]) {
      this.playList = markRaw(data);
    },
    /**
     * 设置原始播放列表
     * @param data 原始播放列表
     */
    async setOriginalPlayList(data: Track[]) {
      this.originalPlayList = markRaw(data);
    },
    /** 清除原始播放列表 */
    async clearOriginalPlayList() {
      this.originalPlayList = [];
    },
  },
});
