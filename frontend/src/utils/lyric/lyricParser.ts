// 照搬 SPlayer utils/lyric/lyricParser.ts，去掉 QRC 解析与 lyricLinesToTTML（Sonora 不支持 QRC；
// 后者不写 itunes:key，AMLL 1.0.1 读不回来）。
import type { LyricLine, LyricWord } from "@applemusic-like-lyrics/lyric";
import { cloneDeep } from "lodash-es";
import { parseLrc } from "./parseLrc";

/**
 * LRC 格式类型
 */
export enum LrcFormat {
  /** 普通逐行 LRC */
  Line = "line",
  /** 逐字 LRC：[00:28.850]曲[00:32.455]：[00:36.060]钱 */
  WordByWord = "word-by-word",
  /** 增强型 LRC (ESLyric)：[01:37.305]<01:37.624>怕<01:37.943>你 */
  Enhanced = "enhanced",
}

// 预编译正则表达式
const META_TAG_REGEX = /^\[[a-z]+:/i;
const TIME_TAG_REGEX = /\[(\d{2}):(\d{2})\.(\d{1,})\]/g;
const ENHANCED_TIME_TAG_REGEX = /<(\d{2}):(\d{2})\.(\d{1,})>/;
// 移除全局带状态的正则，改为在函数内使用 matchAll 或重新构建
const LINE_TIME_REGEX = /^\[(\d{2}):(\d{2})\.(\d{1,})\]/;
const DEFAULT_WORD_DURATION = 1000;
const ALIGN_TOLERANCE_MS = 300;

/**
 * 解析时间戳为毫秒
 * 使用字符串补齐处理，避免浮点数计算误差
 */
const parseTimeToMs = (min: string, sec: string, ms: string): number => {
  const minutes = parseInt(min, 10);
  const seconds = parseInt(sec, 10);
  // 补齐到 3 位 (例如 "5" -> "500", "05" -> "050", "1234" -> "123")
  const msNormalized = ms.padEnd(3, "0").slice(0, 3);
  const milliseconds = parseInt(msNormalized, 10);
  return minutes * 60 * 1000 + seconds * 1000 + milliseconds;
};

/**
 * 创建 LyricWord 对象
 */
const createWord = (
  word: string,
  startTime: number,
  endTime: number = startTime,
): LyricWord => ({
  word,
  startTime,
  endTime,
});

/**
 * 创建 LyricLine 对象
 */
const createLine = (
  words: LyricWord[],
  startTime: number,
  endTime: number = 0,
): LyricLine => ({
  words,
  startTime,
  endTime,
  translatedLyric: "",
  romanLyric: "",
  isBG: false,
  isDuet: false,
});

/**
 * 检测 LRC 格式类型
 */
export const detectLrcFormat = (content: string): LrcFormat => {
  const lines = content.split(/\r?\n/);
  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (!line || META_TAG_REGEX.test(line)) continue;
    // 检查增强型LRC
    if (ENHANCED_TIME_TAG_REGEX.test(line)) {
      return LrcFormat.Enhanced;
    }
    // 检查逐字LRC
    const matches = line.match(TIME_TAG_REGEX);
    if (matches && matches.length > 1) {
      return LrcFormat.WordByWord;
    }
  }
  return LrcFormat.Line;
};

/**
 * 解析逐字 LRC 格式
 * 优化：在解析过程中直接计算 endTime，避免二次遍历
 */
export const parseWordByWordLrc = (content: string): LyricLine[] => {
  const result: LyricLine[] = [];
  let prevLine: LyricLine | null = null;
  const WORD_BY_WORD_PATTERN = /\[(\d{2}):(\d{2})\.(\d{1,})\]([^[\]]*)/g;

  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || META_TAG_REGEX.test(line)) continue;

    const words: LyricWord[] = [];
    let lineStartTime = Infinity;

    let prevWord: LyricWord | null = null;

    const matches = line.matchAll(WORD_BY_WORD_PATTERN);

    for (const match of matches) {
      const startTime = parseTimeToMs(match[1], match[2], match[3]);
      const wordText = match[4];

      if (!wordText && words.length === 0) continue;

      lineStartTime = Math.min(lineStartTime, startTime);

      // 设置上一个字的结束时间
      if (prevWord) {
        prevWord.endTime = startTime;
      }

      if (wordText) {
        const newWord = createWord(wordText, startTime);
        words.push(newWord);
        prevWord = newWord;
      }
    }

    // 处理行内最后一个字
    if (prevWord) {
      prevWord.endTime = prevWord.startTime + DEFAULT_WORD_DURATION;
    }

    if (words.length > 0) {
      const lineObj = createLine(
        words,
        lineStartTime === Infinity ? 0 : lineStartTime,
      );
      // 设置行结束时间为最后一个字的结束时间
      lineObj.endTime = words[words.length - 1].endTime;

      // 修正上一行的结束时间 (Single Pass)
      if (prevLine) {
        const prevLastWord = prevLine.words[prevLine.words.length - 1];
        // 只有当当前行开始时间晚于上一行最后一个字的开始时间时，才进行截断
        if (lineObj.startTime > prevLastWord.startTime) {
          prevLastWord.endTime = Math.min(
            prevLastWord.endTime,
            lineObj.startTime,
          );
          prevLine.endTime = prevLastWord.endTime;
        }
      }

      result.push(lineObj);
      prevLine = lineObj;
    }
  }

  return result;
};

/**
 * 解析增强型 LRC 格式 (ESLyric)
 */
export const parseEnhancedLrc = (content: string): LyricLine[] => {
  const result: LyricLine[] = [];
  let prevLine: LyricLine | null = null;
  const ENHANCED_WORD_PATTERN = /<(\d{2}):(\d{2})\.(\d{1,})>([^<]*)/g;

  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || META_TAG_REGEX.test(line)) continue;

    const lineTimeMatch = LINE_TIME_REGEX.exec(line);
    if (!lineTimeMatch) continue;

    const lineStartTime = parseTimeToMs(
      lineTimeMatch[1],
      lineTimeMatch[2],
      lineTimeMatch[3],
    );
    const contentAfterTime = line.slice(lineTimeMatch[0].length);

    const words: LyricWord[] = [];

    // 检查是否有增强型标记
    if (ENHANCED_TIME_TAG_REGEX.test(contentAfterTime)) {
      let prevWord: LyricWord | null = null;

      const matches = contentAfterTime.matchAll(ENHANCED_WORD_PATTERN);

      for (const match of matches) {
        const startTime = parseTimeToMs(match[1], match[2], match[3]);
        const wordText = match[4];

        if (prevWord) {
          prevWord.endTime = startTime;
        }

        if (wordText) {
          const newWord = createWord(wordText, startTime);
          words.push(newWord);
          prevWord = newWord;
        }
      }

      if (prevWord) {
        prevWord.endTime = prevWord.startTime + DEFAULT_WORD_DURATION; // 默认兜底
      }
    } else {
      // 无增强型标记，作为整行处理
      const text = contentAfterTime.trim();
      if (text) {
        words.push(
          createWord(
            text,
            lineStartTime,
            lineStartTime + DEFAULT_WORD_DURATION,
          ),
        ); // 默认持续1s
      }
    }

    if (words.length > 0) {
      const lineObj = createLine(words, lineStartTime);
      lineObj.endTime = words[words.length - 1].endTime;

      // 修正上一行的结束时间 (Single Pass)
      if (prevLine) {
        const prevLastWord = prevLine.words[prevLine.words.length - 1];
        if (lineObj.startTime > prevLastWord.startTime) {
          prevLastWord.endTime = Math.min(
            prevLastWord.endTime,
            lineObj.startTime,
          );
          prevLine.endTime = prevLastWord.endTime;
        }
      }

      result.push(lineObj);
      prevLine = lineObj;
    }
  }

  return result;
};

/**
 * 智能解析 LRC 歌词
 */
export const parseSmartLrc = (
  content: string,
): { format: LrcFormat; lines: LyricLine[] } => {
  const format = detectLrcFormat(content);

  let lines: LyricLine[];
  switch (format) {
    case LrcFormat.WordByWord:
      lines = parseWordByWordLrc(content);
      break;
    case LrcFormat.Enhanced:
      lines = parseEnhancedLrc(content);
      break;
    default:
      lines = parseLrc(content) || [];
  }

  console.log(`[LyricParser] 检测到歌词格式: ${format}, 共 ${lines.length} 行`);
  return { format, lines };
};

/**
 * 判断解析结果是否为逐字格式
 */
export const isWordLevelFormat = (format: LrcFormat): boolean =>
  format === LrcFormat.WordByWord || format === LrcFormat.Enhanced;

/**
 * 歌词内容对齐
 * 使用双指针算法实现 O(N) 复杂度
 * @param lyrics 歌词数据 (Readonly)
 * @param otherLyrics 其他歌词数据
 * @param key 对齐类型
 * @returns 对齐后的歌词数据 (新副本)
 */
export const alignLyrics = (
  lyrics: Readonly<LyricLine[]>,
  otherLyrics: Readonly<LyricLine[]>,
  key: "translatedLyric" | "romanLyric",
): LyricLine[] => {
  if (!lyrics.length || !otherLyrics.length)
    return cloneDeep(lyrics) as LyricLine[];

  const result = cloneDeep(lyrics) as LyricLine[];

  let i = 0;
  let j = 0;

  while (i < result.length && j < otherLyrics.length) {
    const line = result[i];
    const other = otherLyrics[j];
    const diff = line.startTime - other.startTime;

    if (Math.abs(diff) <= ALIGN_TOLERANCE_MS) {
      // 匹配成功
      line[key] = other.words.map((word) => word.word).join("");
      i++;
      j++;
    } else if (diff < 0) {
      // 当前歌词时间较早，移动当前指针
      i++;
    } else {
      // 目标歌词时间较早，移动目标指针
      j++;
    }
  }
  return result;
};

/**
 * 对齐歌词的翻译和音译
 * 根据开始时间将同一时间的多行歌词分为一组，第一行作为主句，第二行作为翻译，第三行作为音译
 * @param lyrics 未设置翻译和音译的歌词数据 (Readonly)
 * @param endTime 对齐时如何处理附加行的结束时间（忽略、匹配、设为最大值）
 * @param maxTimeDiff 允许匹配的最大时间差（单位：毫秒），超过该时间差的行将不会被视为同一行
 * @param skipSort 跳过排序步骤以只对齐相邻行
 * @returns 对齐后的歌词数据 (新副本)
 */
export const alignLyricLines = (
  lyrics: Readonly<LyricLine[]>,
  {
    endTime = "set",
    maxTimeDiff = 0, // 默认严格匹配
    skipSort = false,
  }: Partial<{
    endTime: "ignore" | "match" | "set";
    maxTimeDiff: number;
    skipSort: boolean;
  }> = {},
): LyricLine[] => {
  if (!lyrics.length) return [];
  // 获取开始时间
  const toStartTime = (line: LyricLine) =>
    Number(line?.startTime ?? line?.words?.[0]?.startTime ?? 0);
  // 获取结束时间
  const toEndTime = (line: LyricLine) =>
    Number(
      line?.endTime ?? line?.words?.[line?.words?.length - 1]?.endTime ?? 0,
    );
  // 取内容
  const toText = (line: LyricLine) =>
    String(line?.words?.map((w) => w.word).join("") || "").trim();
  // 是否匹配
  const isTimeMatch = (
    baseLine: LyricLine | undefined,
    addLine: LyricLine | undefined,
  ) => {
    if (!baseLine || !addLine) return false;
    const timeDiff = Math.abs(toStartTime(baseLine) - toStartTime(addLine));
    if (!Number.isFinite(timeDiff)) return false;
    if (timeDiff > maxTimeDiff) return false;
    if (endTime === "match") {
      const endTimeDiff = Math.abs(toEndTime(baseLine) - toEndTime(addLine));
      if (!Number.isFinite(endTimeDiff)) return false;
      if (endTimeDiff > maxTimeDiff) return false;
    }
    return true;
  };
  // 按开始时间分组
  const sorted = skipSort
    ? lyrics
    : [...lyrics].sort((a, b) => toStartTime(a) - toStartTime(b));
  const groups: LyricLine[][] = [];
  for (const line of sorted) {
    const last = groups[groups.length - 1]?.[0];
    if (isTimeMatch(last, line)) groups[groups.length - 1].push(line);
    else groups.push([line]);
  }
  // 合并附加行
  const mergeAddLine = (
    baseLine: LyricLine,
    addLine: LyricLine | undefined,
    key: "translatedLyric" | "romanLyric",
  ) => {
    if (baseLine[key] || !addLine) return;
    const addText = toText(addLine);
    if (!addText) return;
    baseLine[key] = addText;
    // 如果需要设置主行的结束时间，则将主行的结束时间设置为主行和附加行结束时间的较大值
    if (endTime !== "set") return;
    const oldEndTime = toEndTime(baseLine);
    const addEndTime = toEndTime(addLine);
    if (!Number.isFinite(addEndTime) || addEndTime <= oldEndTime) return;
    baseLine.endTime = addEndTime;
    // 考虑句中最后一个字的结束时间
    if (baseLine.words?.length) {
      const lastWord = baseLine.words[baseLine.words.length - 1];
      const lastWordEndTime = lastWord.endTime;
      if (lastWordEndTime === oldEndTime) {
        lastWord.endTime = addEndTime;
      }
    }
  };
  // 组装：第 1 行主句；第 2 行翻译；第 3 行音译；其余行舍去
  const aligned = groups.map((group) => {
    // 使用 cloneDeep 保证数据不可变性
    const base = cloneDeep(group[0]) as LyricLine;
    const tran = group[1];
    const roma = group[2];
    mergeAddLine(base, tran, "translatedLyric");
    mergeAddLine(base, roma, "romanLyric");
    return base;
  });
  return aligned;
};
