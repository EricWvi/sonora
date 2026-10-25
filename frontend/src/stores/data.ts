import { defineStore } from "pinia";
import type { QueueEntry } from "@/core/player/queue";

/**
 * 播放列表与喜爱的歌曲。字段名照搬 SPlayer stores/data.ts 中播放器用到的部分。
 * 播放列表只由播放控制器修改，并由它保存到浏览器中（core/player/storage.ts）；
 * 喜爱来自 `likes/ids`。
 */
interface DataState {
  /** 播放列表；同一首歌可以出现多次，每次是一个条目 */
  playList: QueueEntry[];
  /** 原始播放列表（随机播放前的顺序）；不在随机模式时为空 */
  originalPlayList: QueueEntry[];
  /** 搜索历史 */
  searchHistory: string[];
  /** 用户喜欢的数据；Sonora 只有歌曲 */
  userLikeData: {
    songs: number[];
  };
}

export const useDataStore = defineStore("data", {
  state: (): DataState => ({
    playList: [],
    originalPlayList: [],
    searchHistory: [],
    userLikeData: { songs: [] },
  }),
  getters: {
    // 是否为喜欢的歌曲
    isLikeSong: (state) => (id: number) =>
      state.userLikeData.songs.includes(id),
  },
  actions: {
    /**
     * 更新播放列表
     * @param data 播放列表
     */
    async setPlayList(data: QueueEntry[]) {
      this.playList = markRaw(data);
    },
    /**
     * 设置原始播放列表
     * @param data 原始播放列表
     */
    async setOriginalPlayList(data: QueueEntry[]) {
      this.originalPlayList = markRaw(data);
    },
  },
  // 持久化：搜索历史保存在浏览器中（SPlayer 的 data-store 同样保存它）
  persist: {
    key: "sonora-data",
    storage: localStorage,
    pick: ["searchHistory"],
  },
});
