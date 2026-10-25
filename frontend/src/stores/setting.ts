import { defineStore } from "pinia";
import type { ThemeColorType } from "@/types/color";

/**
 * 设置。字段名、含义与默认值照搬 SPlayer stores/setting.ts，只保留阶段 2 确定保留的设置项
 * （见 docs/phase/phase-2-player-baseline.md 的“设置”）；在线服务、下载、本地歌曲等不做的项不迁入。
 * 外观与歌词设置以后按用户保存在服务端，标为“设备”的项只保存在浏览器中。
 */
export interface SettingState {
  // ---- 外观 ----
  /** 明暗模式 */
  themeMode: "light" | "dark" | "auto";
  /** 主题类别 */
  themeColorType: ThemeColorType;
  /** 主题自定义颜色 */
  themeCustomColor: string;
  /** 全局着色 */
  themeGlobalColor: boolean;
  /** 主题变体 */
  themeVariant:
    | "primary"
    | "secondary"
    | "tertiary"
    | "neutral"
    | "neutralVariant"
    | "error";
  /** 主题跟随封面 */
  themeFollowCover: boolean;
  /** 字体设置样式 */
  fontSettingStyle: "single" | "multi" | "custom";
  /** 全局字体 */
  globalFont: "default" | string;
  /** 歌词区域字体 */
  LyricFont: "follow" | string;
  /** 日语歌词字体 */
  japaneseLyricFont: "follow" | string;
  /** 英语歌词字体 */
  englishLyricFont: "follow" | string;
  /** 韩语歌词字体 */
  koreanLyricFont: "follow" | string;
  /** 播放器类型 */
  playerType: "cover" | "record" | "fullscreen";
  /** 播放器封面/歌词占比 (0-100) */
  playerStyleRatio: number;
  /** 全屏封面渐变位置 (0-100) */
  playerFullscreenGradient: number;
  /** 背景类型 */
  playerBackgroundType: "none" | "animation" | "blur" | "color";
  /** 背景动画帧率 */
  playerBackgroundFps: number;
  /** 背景动画流动速度 */
  playerBackgroundFlowSpeed: number;
  /** 背景动画是否在歌曲暂停时暂停 */
  playerBackgroundPause: boolean;
  /** 背景动画是否响应低频音量 */
  playerBackgroundLowFreqVolume: boolean;
  /** 背景动画渲染比例 */
  playerBackgroundRenderScale: number;
  /** 播放器展开动画 */
  playerExpandAnimation: "up" | "flow";
  /** 播放器跟随封面主色 */
  playerFollowCoverColor: boolean;
  /** 动态封面 */
  dynamicCover: boolean;
  /** 是否显示音乐频谱 */
  showSpectrums: boolean;
  /** 播放器元素自动隐藏 */
  autoHidePlayerMeta: boolean;
  /** 显示倒计时 */
  countDownShow: boolean;
  /** 显示歌词条 */
  barLyricShow: boolean;
  /** 时间显示格式 */
  timeFormat: "current-total" | "remaining-total" | "current-remaining";
  /** 显示播放列表数量 */
  showPlaylistCount: boolean;
  /** 展示当前歌曲歌词状态信息 */
  showPlayMeta: boolean;
  /** 封面显示配置（SPlayer 的在线页面项不迁入） */
  hiddenCovers: {
    /** 首页 */
    home: boolean;
    /** 播放器 */
    player: boolean;
    /** 歌单详情/歌曲列表 */
    list: boolean;
    /** 艺术家详情 */
    artistDetail: boolean;
    /** 专辑 */
    album: boolean;
  };
  /** 隐藏全部封面 */
  hideAllCovers: boolean;
  /** 隐藏迷你播放器封面 */
  hideMiniPlayerCover: boolean;
  /** 全屏播放器界面元素显示配置（评论、下载、桌面歌词不迁入） */
  fullscreenPlayerElements: {
    like: boolean;
    addToPlaylist: boolean;
    moreSettings: boolean;
    copyLyric: boolean;
    lyricOffset: boolean;
    lyricSettings: boolean;
  };
  /** 显示歌曲专辑 */
  showSongAlbum: boolean;
  /** 显示歌曲时长 */
  showSongDuration: boolean;
  /** 显示歌曲操作 */
  showSongOperations: boolean;
  /** 显示歌曲歌手 */
  showSongArtist: boolean;
  /** 路由动画 */
  routeAnimation:
    | "none"
    | "fade"
    | "zoom"
    | "slide"
    | "up"
    | "flow"
    | "mask-left"
    | "mask-top";
  /** 菜单显示封面 */
  menuShowCover: boolean;
  /** 菜单展开项 */
  menuExpandedKeys: (string | number)[];
  /** 侧边栏隐藏（只保留 Sonora 侧栏中可隐藏的项） */
  sidebarHide: {
    /** 隐藏最近播放 */
    hideHistory: boolean;
    /** 隐藏创建的歌单 */
    hideUserPlaylists: boolean;
  };

  // ---- 歌词 ----
  /** 歌词字体大小模式 */
  lyricFontSizeMode: "fixed" | "adaptive";
  /** 歌词字体大小 */
  lyricFontSize: number;
  /** 歌词翻译字体大小 */
  lyricTranFontSize: number;
  /** 歌词音译字体大小 */
  lyricRomaFontSize: number;
  /** 歌词字重设置 */
  lyricFontWeight: number;
  /** 显示逐字歌词 */
  showWordLyrics: boolean;
  /** 显示歌词翻译 */
  showTran: boolean;
  /** 显示歌词音译 */
  showRoma: boolean;
  /** 调换翻译与音译位置 */
  swapTranRoma: boolean;
  /** 显示逐字音译 */
  showWordsRoma: boolean;
  /** 歌词动画 */
  lyricTransition: "slide" | "fade";
  /** 歌词位置 */
  lyricsPosition: "flex-start" | "center" | "flex-end";
  /** 歌词滚动位置偏移量 */
  lyricsScrollOffset: number;
  /** 歌词水平位置偏移量 */
  lyricHorizontalOffset: number;
  /** 歌词默认靠右（对唱互换） */
  lyricAlignRight: boolean;
  /** 歌词模糊 */
  lyricsBlur: boolean;
  /** 歌词混合模式 */
  lyricsBlendMode: "screen" | "plus-lighter";
  /** 隐藏歌词括号内容和别名 */
  hideBracketedContent: boolean;
  /** 替换歌词括号内容 */
  replaceLyricBrackets: boolean;
  /** 歌词括号替换预设 */
  bracketReplacementPreset:
    "dash" | "angleBrackets" | "cornerBrackets" | "custom";
  /** 自定义歌词括号替换内容 */
  customBracketReplacement: string;
  /** 是否启用排除歌词 */
  enableExcludeLyrics: boolean;
  /** 「排除歌词」是否适用于 TTML */
  enableExcludeLyricsTTML: boolean;
  /** 用户自定义的排除歌词关键字 */
  excludeLyricsUserKeywords: string[];
  /** 用户自定义的排除歌词正则表达式 */
  excludeLyricsUserRegexes: string[];
  /** 是否使用 AMLL 歌词 */
  useAMLyrics: boolean;
  /** 是否使用 AMLL 歌词弹簧效果 */
  useAMSpring: boolean;
  /** 隐藏已播放歌词 */
  hidePassedLines: boolean;
  /** 文字动画的渐变宽度 */
  wordFadeWidth: number;
  /** 歌词时延调节步长（毫秒） */
  lyricOffsetStep: number;

  // ---- 播放 ----
  /** 自动播放 */
  autoPlay: boolean;
  /** 预载下一首 */
  useNextPrefetch: boolean;
  /** 记忆最后进度 */
  memoryLastSeek: boolean;
  /** 是否禁止休眠 */
  preventSleep: boolean;
  /** 显示进度条悬浮信息 */
  progressTooltipShow: boolean;
  /** 进度条悬浮时显示歌词 */
  progressLyricShow: boolean;
  /** 进度调节吸附最近歌词 */
  progressAdjustLyric: boolean;
  /** 渐入渐出 */
  songVolumeFade: boolean;
  /** 渐入渐出时间 */
  songVolumeFadeTime: number;
  /** 播放设备（设备） */
  playDevice: "default" | string;
  /** 音频延迟手动补偿（毫秒，设备） */
  audioDelayCompensation: number;
  /** 是否启用 ReplayGain (音量平衡) */
  enableReplayGain: boolean;
  /** ReplayGain 模式: 轨道增益 (track) 或 专辑增益 (album) */
  replayGainMode: "track" | "album";

  // ---- 常规 ----
  /** 显示搜索历史 */
  showSearchHistory: boolean;
  /** 启用搜索关键词获取（搜索建议） */
  enableSearchKeyword: boolean;
  /** 搜索框行为 */
  searchInputBehavior: "normal" | "clear" | "sync";
  /** 显示主页问好 */
  showHomeGreeting: boolean;
  /** 是否使用 keep-alive */
  useKeepAlive: boolean;
}

export const useSettingStore = defineStore("setting", {
  state: (): SettingState => ({
    themeMode: "auto",
    themeColorType: "default",
    themeCustomColor: "#fe7971",
    themeFollowCover: false,
    themeGlobalColor: false,
    themeVariant: "secondary",
    fontSettingStyle: "single",
    globalFont: "default",
    LyricFont: "follow",
    japaneseLyricFont: "follow",
    englishLyricFont: "follow",
    koreanLyricFont: "follow",
    playerType: "cover",
    playerStyleRatio: 50,
    playerFullscreenGradient: 15,
    playerBackgroundType: "blur",
    playerBackgroundFps: 30,
    playerBackgroundFlowSpeed: 4,
    playerBackgroundPause: false,
    playerBackgroundLowFreqVolume: false,
    playerBackgroundRenderScale: 0.5,
    playerExpandAnimation: "up",
    playerFollowCoverColor: true,
    dynamicCover: false,
    showSpectrums: false,
    autoHidePlayerMeta: true,
    countDownShow: true,
    barLyricShow: true,
    timeFormat: "current-total",
    showPlaylistCount: true,
    showPlayMeta: true,
    hiddenCovers: {
      home: false,
      player: false,
      list: false,
      artistDetail: false,
      album: false,
    },
    hideAllCovers: false,
    hideMiniPlayerCover: false,
    fullscreenPlayerElements: {
      like: true,
      addToPlaylist: true,
      moreSettings: true,
      copyLyric: true,
      lyricOffset: true,
      lyricSettings: true,
    },
    showSongAlbum: true,
    showSongDuration: true,
    showSongOperations: true,
    showSongArtist: true,
    routeAnimation: "slide",
    menuShowCover: true,
    menuExpandedKeys: [],
    sidebarHide: {
      hideHistory: false,
      hideUserPlaylists: false,
    },
    lyricFontSizeMode: "adaptive",
    lyricFontSize: 46,
    lyricTranFontSize: 22,
    lyricRomaFontSize: 18,
    lyricFontWeight: 700,
    showWordLyrics: true,
    showTran: true,
    showRoma: true,
    swapTranRoma: false,
    showWordsRoma: true,
    lyricTransition: "slide",
    lyricsPosition: "flex-start",
    lyricsScrollOffset: 0.25,
    lyricHorizontalOffset: 10,
    lyricAlignRight: false,
    lyricsBlur: false,
    lyricsBlendMode: "screen",
    hideBracketedContent: false,
    replaceLyricBrackets: false,
    bracketReplacementPreset: "dash",
    customBracketReplacement: "-",
    enableExcludeLyrics: true,
    enableExcludeLyricsTTML: false,
    excludeLyricsUserKeywords: [],
    excludeLyricsUserRegexes: [],
    useAMLyrics: false,
    useAMSpring: false,
    hidePassedLines: false,
    wordFadeWidth: 0.5,
    lyricOffsetStep: 500,
    autoPlay: false,
    useNextPrefetch: true,
    memoryLastSeek: true,
    preventSleep: false,
    progressTooltipShow: true,
    progressLyricShow: true,
    progressAdjustLyric: false,
    songVolumeFade: true,
    songVolumeFadeTime: 300,
    playDevice: "default",
    audioDelayCompensation: 0,
    enableReplayGain: false,
    replayGainMode: "track",
    showSearchHistory: true,
    enableSearchKeyword: true,
    searchInputBehavior: "normal",
    showHomeGreeting: true,
    useKeepAlive: true,
  }),
  getters: {
    /** 获取淡入淡出时间 */
    getFadeTime(state): number {
      return state.songVolumeFade ? state.songVolumeFadeTime : 0;
    },
  },
  actions: {
    // 更换明暗模式
    setThemeMode(mode?: "auto" | "light" | "dark") {
      if (mode === undefined) {
        if (this.themeMode === "auto") {
          this.themeMode = "light";
        } else if (this.themeMode === "light") {
          this.themeMode = "dark";
        } else {
          this.themeMode = "auto";
        }
      } else {
        this.themeMode = mode;
      }
      window.$message.info(
        `已切换至
        ${
          this.themeMode === "auto"
            ? "跟随系统"
            : this.themeMode === "light"
              ? "浅色模式"
              : "深色模式"
        }`,
        {
          showIcon: false,
        },
      );
    },
  },
  persist: {
    key: "sonora-setting",
    storage: localStorage,
  },
});
