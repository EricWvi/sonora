import { parseTTML, type LyricLine } from "@applemusic-like-lyrics/lyric";
import { isOnline } from "@/utils/track";
import { source } from "@/api/source";
import type { Lyrics, Track } from "@/api/types";
import {
  keywords as defaultKeywords,
  regexes as defaultRegexes,
} from "@/assets/data/exclude";
import { useMusicStore, useSettingStore, useStatusStore } from "@/stores";
import type { SongLyric } from "@/types/lyric";
import { applyBracketReplacement } from "@/utils/lyric/lyricFormat";
import {
  alignLyricLines,
  isWordLevelFormat,
  parseSmartLrc,
} from "@/utils/lyric/lyricParser";
import { stripLyricMetadata } from "@/utils/lyric/lyricStripper";
import { cleanTTMLTranslations } from "@/utils/lyric/parseTTML";

/**
 * 歌词管理。处理流程照搬 SPlayer core/player/LyricManager.ts：Sonora 的歌词只有服务端保存的
 * 一份，格式为 LRC 或 TTML，按 SPlayer 处理本地歌词的方式解析（TTML 作为逐字歌词；LRC 识别
 * 普通、逐字与增强型，普通 LRC 对齐翻译与音译），之后的排除、括号替换与规范化与原项目相同。
 * 不迁入网易、QQ 音乐与在线 TTML 歌词源、缓存、繁简转换、脏词还原与桌面歌词。
 */
class LyricManager {
  /** 当前歌词请求，切歌后丢弃旧请求的结果 */
  private lyricReqSeq = 0;

  /** 重置当前歌曲的歌词数据 */
  public resetSongLyric() {
    const musicStore = useMusicStore();
    const statusStore = useStatusStore();
    musicStore.setSongLyric({ lrcData: [], yrcData: [] }, true);
    statusStore.usingTTMLLyric = false;
    statusStore.lyricIndex = -1;
    statusStore.lyricLoading = false;
  }

  /**
   * 解析服务端保存的歌词
   * @param lyrics 歌词原文与格式
   * @returns 歌词数据与是否为 TTML
   */
  public parseLyrics({ format, lyrics }: Lyrics): {
    data: SongLyric;
    usingTTMLLyric: boolean;
  } {
    if (!lyrics)
      return { data: { lrcData: [], yrcData: [] }, usingTTMLLyric: false };
    // TTML 直接返回
    if (format === "ttml") {
      const sorted = cleanTTMLTranslations(lyrics);
      const lines = parseTTML(sorted)?.lines || [];
      return {
        data: { lrcData: [], yrcData: lines },
        usingTTMLLyric: lines.length > 0,
      };
    }
    const { format: lrcFormat, lines } = parseSmartLrc(lyrics);
    // 如果是逐字格式，直接作为 yrcData
    if (isWordLevelFormat(lrcFormat)) {
      return { data: { lrcData: [], yrcData: lines }, usingTTMLLyric: false };
    }
    // 普通格式
    return {
      data: { lrcData: alignLyricLines(lines), yrcData: [] },
      usingTTMLLyric: false,
    };
  }

  /**
   * 处理歌词排除
   * @param lyricData 歌词数据
   * @param song 目标歌曲
   * @param usingTTMLLyric 是否使用 TTML 歌词
   * @returns 处理后的歌词数据
   */
  private handleLyricExclude(
    lyricData: SongLyric,
    song: Track,
    usingTTMLLyric: boolean,
  ): SongLyric {
    const settingStore = useSettingStore();
    const {
      enableExcludeLyrics,
      excludeLyricsUserKeywords,
      excludeLyricsUserRegexes,
    } = settingStore;

    if (!enableExcludeLyrics) return lyricData;

    // 合并默认规则和用户自定义规则
    const mergedKeywords = [
      ...new Set([...defaultKeywords, ...(excludeLyricsUserKeywords ?? [])]),
    ];
    const mergedRegexes = [
      ...new Set([...defaultRegexes, ...(excludeLyricsUserRegexes ?? [])]),
    ];

    const options = {
      keywords: mergedKeywords,
      regexPatterns: mergedRegexes,
      matchMetadata: {
        title: song.id !== 0 ? song.title : undefined,
        artists: song.artists.map((artist) => artist.name).filter(Boolean),
      },
    };

    const lrcData = stripLyricMetadata(lyricData.lrcData || [], options);
    let yrcData = lyricData.yrcData || [];
    if (!usingTTMLLyric || settingStore.enableExcludeLyricsTTML) {
      yrcData = stripLyricMetadata(yrcData, options);
    }
    return { lrcData, yrcData };
  }

  /**
   * 规范化歌词行时间
   * @param lines 歌词行
   */
  private normalizeLyricLines(lines: LyricLine[]) {
    lines.forEach((line) => {
      // 修复 startTime / endTime 为 0 或 invalid 的情况
      if ((!line.startTime || line.startTime <= 0) && line.words?.length) {
        line.startTime = line.words[0].startTime;
      }
      if ((!line.endTime || line.endTime <= 0) && line.words?.length) {
        line.endTime = line.words[line.words.length - 1].endTime;
      }
    });
  }

  /**
   * 设置最终歌词
   * @param lyricData 歌词数据
   */
  private setFinalLyric(lyricData: SongLyric) {
    const musicStore = useMusicStore();
    const statusStore = useStatusStore();
    // 应用括号替换
    lyricData = applyBracketReplacement(lyricData);
    // 规范化时间
    this.normalizeLyricLines(lyricData.yrcData);
    this.normalizeLyricLines(lyricData.lrcData);
    // 如果只有逐字歌词
    if (lyricData.lrcData.length === 0 && lyricData.yrcData.length > 0) {
      // 构成普通歌词
      lyricData.lrcData = lyricData.yrcData.map((line) => ({
        ...line,
        words: [
          {
            word: line.words?.map((w) => w.word)?.join("") || "",
            startTime: line.startTime || 0,
            endTime: line.endTime || 0,
            romanWord:
              line.words?.map((w) => w.romanWord)?.join("") || undefined,
          },
        ],
      }));
    }
    // 设置歌词
    musicStore.setSongLyric(lyricData, true);
    // 结束加载状态
    statusStore.lyricLoading = false;
  }

  /**
   * 处理歌词
   * @param song 歌曲对象
   */
  public async handleLyric(song: Track) {
    const statusStore = useStatusStore();
    const req = ++this.lyricReqSeq;
    try {
      const lyrics: Lyrics = song.hasLyrics
        ? await (isOnline(song)
            ? source.onlineLyrics(song.id)
            : source.lyrics(song.id))
        : { format: "", lyrics: "" };
      if (this.lyricReqSeq !== req) return;
      const { data, usingTTMLLyric } = this.parseLyrics(lyrics);
      statusStore.usingTTMLLyric = usingTTMLLyric;
      this.setFinalLyric(this.handleLyricExclude(data, song, usingTTMLLyric));
    } catch (error) {
      if (this.lyricReqSeq !== req) return;
      console.error("❌ 处理歌词失败:", error);
      this.resetSongLyric();
    }
  }
}

let instance: LyricManager | null = null;

export const useLyricManager = (): LyricManager => {
  if (!instance) instance = new LyricManager();
  return instance;
};
