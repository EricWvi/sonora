import { defineStore } from "pinia";
import type { Track } from "@/api/types";
import type { SongLyric } from "@/types/lyric";
import { coverSizes, coverUrl } from "@/utils/cover";

/**
 * 当前歌曲与歌词。字段名照搬 SPlayer stores/music.ts，歌曲是播放端 API 的 `Track`。
 * 不迁入私人 FM、每日推荐与桌面歌词的同步。
 */
interface MusicState {
  playSong: Track;
  /**
   * 当前播放列表的来源，如 `album-3`、`playlist-5`、`likes`；空字符串表示其他列表。
   * SPlayer 用网易的数字 ID，Sonora 的专辑与歌单 ID 会重复，所以带上类型
   */
  playPlaylistId: string;
  songLyric: SongLyric;
}

// 默认音乐数据
const defaultMusicData: Track = {
  id: 0,
  title: "未播放歌曲",
  artists: [],
  album: null,
  trackNo: 0,
  discNo: 0,
  year: 0,
  durationMs: 0,
  format: "",
  cover: null,
  hasLyrics: false,
};

const songCoverUrl = (song: Track, size: keyof typeof coverSizes) =>
  coverUrl(song.cover, size);

export const useMusicStore = defineStore("music", {
  state: (): MusicState => ({
    // 当前播放歌曲
    playSong: { ...defaultMusicData },
    // 当前播放列表 ID
    playPlaylistId: "",
    // 当前歌曲歌词
    songLyric: {
      lrcData: [], // 普通歌词
      yrcData: [], // 逐字歌词
    },
  }),
  getters: {
    // 是否具有歌词
    isHasLrc(state): boolean {
      return state.songLyric.lrcData.length > 0;
    },
    // 是否具有逐字歌词
    isHasYrc(state): boolean {
      return state.songLyric.yrcData.length > 0;
    },
    // 是否有播放器
    isHasPlayer(state): boolean {
      return state.playSong?.id !== 0;
    },
    /** 歌曲封面 */
    songCover(state): string {
      return songCoverUrl(state.playSong, "s");
    },
  },
  actions: {
    /** 重置音乐数据 */
    resetMusicData() {
      this.playSong = { ...defaultMusicData };
      this.playPlaylistId = "";
      this.setSongLyric({ lrcData: [], yrcData: [] }, true);
    },
    /**
     * 设置/更新歌曲歌词数据
     * @param updates 部分或完整歌词数据
     * @param replace 是否覆盖（true：用提供的数据覆盖并为缺省字段置空；false：合并更新）
     */
    setSongLyric(updates: Partial<SongLyric>, replace: boolean = false) {
      if (replace) {
        this.songLyric = {
          lrcData: updates.lrcData ?? [],
          yrcData: updates.yrcData ?? [],
        };
      } else {
        this.songLyric = {
          lrcData: updates.lrcData ?? this.songLyric.lrcData,
          yrcData: updates.yrcData ?? this.songLyric.yrcData,
        };
      }
    },
    // 获取歌曲封面；没有封面时为 SPlayer 的默认封面
    getSongCover(size: keyof typeof coverSizes = "s") {
      return songCoverUrl(this.playSong, size);
    },
  },
});
