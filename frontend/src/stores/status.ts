import { defineStore } from "pinia";
import type { SortField } from "@/api/types";
import type { ColorScheme, RGB } from "@/types/color";
import type { RepeatModeType, ShuffleModeType } from "@/types/play-mode";

/**
 * 运行状态。字段名与默认值照搬 SPlayer stores/status.ts，只迁入 Sonora 保留的功能：
 * 侧栏、主题与播放器。不迁入的有倍速、定时关闭、AB 循环、音质与音源、私人 FM、
 * 桌面与任务栏歌词、评论、更新、自定义背景。
 */
export interface StatusState {
  /** 菜单折叠状态 */
  menuCollapsed: boolean;
  /** 播放控制条 */
  showPlayBar: boolean;
  /** 全屏播放器 */
  showFullPlayer: boolean;
  /** 播放器功能显示 */
  playerMetaShow: boolean;
  /** 播放列表状态 */
  playListShow: boolean;
  /** 播放状态 */
  playStatus: boolean;
  /** 播放加载状态 */
  playLoading: boolean;
  /** 播放音量 */
  playVolume: number;
  /** 静音前音量 */
  playVolumeMute: number;
  /** 循环模式 */
  repeatMode: RepeatModeType;
  /** 随机模式 */
  shuffleMode: ShuffleModeType;
  /** 封面主题 */
  songCoverTheme: {
    /** 封面主题颜色 */
    main?: RGB;
    /** 封面主题颜色（亮色） */
    light?: ColorScheme;
    /** 封面主题颜色（暗色） */
    dark?: ColorScheme;
  };
  /** 纯净歌词模式 */
  pureLyricMode: boolean;
  /** 当前是否正使用 TTML 歌词 */
  usingTTMLLyric: boolean;
  /** 当前播放索引 */
  playIndex: number;
  /** 歌词播放索引 */
  lyricIndex: number;
  /** 歌词加载状态 */
  lyricLoading: boolean;
  /** 当前播放时间 */
  currentTime: number;
  /** 歌曲总时长 */
  duration: number;
  /** 实时播放进度 */
  progress: number;
  /** 每首歌曲的进度偏移（按歌曲 id 记忆） */
  currentTimeOffsetMap: Record<number, number>;
  /** 主内容高度 */
  mainContentHeight: number;
  /** 搜索框焦点 */
  searchFocus: boolean;
  /** 搜索框输入值 */
  searchInputValue: string;
  /** 列表排序字段（Sonora 只有服务端支持的字段） */
  listSortField: SortField;
  /** 列表排序顺序 */
  listSortOrder: "default" | "asc" | "desc";
  /** 均衡器是否开启 */
  eqEnabled: boolean;
  /** 均衡器 10 段增益（dB） */
  eqBands: number[];
  /** 均衡器当前预设 key */
  eqPreset: string;
}

export const useStatusStore = defineStore("status", {
  state: (): StatusState => ({
    menuCollapsed: false,
    showPlayBar: true,
    playStatus: false,
    playLoading: true,
    playListShow: false,
    showFullPlayer: false,
    playerMetaShow: true,
    currentTime: 0,
    duration: 0,
    progress: 0,
    currentTimeOffsetMap: {},
    songCoverTheme: {},
    pureLyricMode: false,
    usingTTMLLyric: false,
    playIndex: -1,
    lyricIndex: -1,
    lyricLoading: false,
    playVolume: 0.7,
    playVolumeMute: 0,
    repeatMode: "off",
    shuffleMode: "off",
    mainContentHeight: 0,
    searchFocus: false,
    searchInputValue: "",
    listSortField: "default",
    listSortOrder: "default",
    eqEnabled: false,
    eqBands: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    eqPreset: "acoustic",
  }),
  getters: {
    // 播放音量图标
    playVolumeIcon(state) {
      const volume = state.playVolume;
      return volume === 0
        ? "VolumeOff"
        : volume < 0.4
          ? "VolumeMute"
          : volume < 0.7
            ? "VolumeDown"
            : "VolumeUp";
    },
    /** 播放模式图标 */
    shuffleIcon() {
      return "Shuffle";
    },
    /** 循环模式图标 */
    repeatIcon(state) {
      if (state.repeatMode === "one") {
        return "RepeatSong";
      }
      return "Repeat";
    },
    /** 音量百分比 */
    playVolumePercent(state) {
      return Math.round(state.playVolume * 100);
    },
    /** 播放器主色 */
    mainColor(state) {
      const mainColor = state.songCoverTheme?.main;
      if (!mainColor) return "239, 239, 239";
      return `${mainColor.r}, ${mainColor.g}, ${mainColor.b}`;
    },
  },
  actions: {
    /**
     * 获取指定歌曲的偏移
     * @param songId 歌曲 id
     * @returns 歌曲的偏移量（单位：毫秒，正值为歌词提前），默认为 0
     */
    getSongOffset(songId?: number): number {
      if (!songId) return 0;
      const offsetTime = this.currentTimeOffsetMap?.[songId] ?? 0;
      return Math.floor(offsetTime * 1000);
    },
    /**
     * 设置指定歌曲的偏移
     * @param songId 歌曲 id
     * @param offset 偏移量（单位：毫秒，正值为歌词提前）
     */
    setSongOffset(songId?: number, offset: number = 0) {
      if (!songId) return;
      if (!this.currentTimeOffsetMap) this.currentTimeOffsetMap = {};
      // 将毫秒转换为秒存储（保留2位小数）
      const offsetSeconds = offset / 1000;
      const fixed = Number(offsetSeconds.toFixed(2));
      if (fixed === 0) {
        // 为 0 时移除记录，避免占用空间
        delete this.currentTimeOffsetMap[songId];
      } else {
        this.currentTimeOffsetMap[songId] = fixed;
      }
    },
    /**
     * 调整指定歌曲的偏移（增量）
     * @param songId 歌曲 id
     * @param delta 偏移增量（单位：毫秒，默认 500ms，正值为歌词提前）
     */
    incSongOffset(songId?: number, delta: number = 500) {
      if (!songId) return;
      const current = this.getSongOffset(songId);
      const next = current + delta;
      if (next === 0) {
        delete this.currentTimeOffsetMap[songId];
      } else {
        this.setSongOffset(songId, next);
      }
    },
    /** 重置指定歌曲的偏移为 0 */
    resetSongOffset(songId?: number) {
      if (!songId) return;
      // 直接删除该歌曲记录
      if (this.currentTimeOffsetMap && songId in this.currentTimeOffsetMap) {
        delete this.currentTimeOffsetMap[songId];
      }
    },
    /**
     * 切换循环模式
     * 顺序: List -> One -> Off -> List
     */
    toggleRepeat() {
      if (this.repeatMode === "list") {
        this.repeatMode = "one";
      } else if (this.repeatMode === "one") {
        this.repeatMode = "off";
      } else {
        this.repeatMode = "list";
      }
    },
    /**
     * 切换随机模式
     * 顺序: Off -> On -> Off
     */
    toggleShuffle() {
      if (this.shuffleMode === "off") {
        this.shuffleMode = "on";
      } else {
        this.shuffleMode = "off";
      }
    },
    /**
     * 设置 EQ 开关
     * @param enabled 是否开启
     */
    setEqEnabled(enabled: boolean) {
      this.eqEnabled = enabled;
    },
    /**
     * 设置 EQ 10 段增益（dB）
     * @param bands 长度 10 的 dB 数组
     */
    setEqBands(bands: number[]) {
      if (Array.isArray(bands) && bands.length === 10) {
        this.eqBands = [...bands];
      }
    },
    /**
     * 设置 EQ 预设名
     */
    setEqPreset(preset: string) {
      this.eqPreset = preset;
    },
    /** 重置播放状态 */
    resetPlayStatus() {
      this.$patch({
        currentTime: 0,
        duration: 0,
        progress: 0,
        lyricIndex: -1,
        playStatus: false,
        playLoading: false,
        playListShow: false,
        showFullPlayer: false,
        playIndex: -1,
        repeatMode: "off",
        shuffleMode: "off",
      });
    },
  },
  // 持久化。播放队列与位置（`currentTime`、`playIndex` 等）在阶段 4.4 接入播放时一并保存。
  persist: {
    key: "sonora-status",
    storage: localStorage,
    pick: [
      "menuCollapsed",
      "currentTimeOffsetMap",
      "pureLyricMode",
      "playVolume",
      "playVolumeMute",
      "repeatMode",
      "shuffleMode",
      "songCoverTheme",
      "listSortField",
      "listSortOrder",
      "eqEnabled",
      "eqBands",
      "eqPreset",
    ],
  },
});
