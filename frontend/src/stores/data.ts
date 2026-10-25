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
    /**
     * 将歌曲插入到当前播放位置之后，并移除队列中重复的这首歌
     * @param song 歌曲
     * @param index 当前播放索引
     * @returns 插入后的歌曲索引
     */
    async setNextPlaySong(song: Track, index: number): Promise<number> {
      // 若为空,则直接添加
      if (this.playList.length === 0) {
        this.playList = [song];
        return 0;
      }
      // 避免直接修改 state
      const newList = [...this.playList];
      // 在当前播放位置之后插入歌曲
      const indexAdd = index + 1;
      newList.splice(indexAdd, 0, song);
      // 移除重复的歌曲（如果存在）
      const finalList = newList.filter(
        (item, idx) => idx === indexAdd || item.id !== song.id,
      );
      this.playList = markRaw(finalList);
      // 返回刚刚插入的歌曲索引
      return finalList.indexOf(song);
    },
    /** 清除原始播放列表 */
    async clearOriginalPlayList() {
      this.originalPlayList = [];
    },
  },
  // 持久化：搜索历史保存在浏览器中（SPlayer 的 data-store 同样保存它）
  persist: {
    key: "sonora-data",
    storage: localStorage,
    pick: ["searchHistory"],
  },
});
