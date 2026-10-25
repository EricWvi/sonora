// 照搬 SPlayer utils/format.ts 中播放器用到的函数。

/**
 * 移除文本中的括号内容（支持中英文括号）
 * @param text 原始文本
 * @returns 处理后的文本
 */
export const removeBrackets = (text: string | undefined): string => {
  if (!text) return "";
  return text.replace(/[（(][^）)]*[）)]/g, "").trim();
};

/**
 * 检测歌词语言
 * @param lyric 歌词内容
 * @returns 语言代码（"ja" | "zh-CN" | "en"）
 */
export const getLyricLanguage = (
  lyric: string,
): "ja" | "ko" | "zh-CN" | "en" => {
  if (!lyric || typeof lyric !== "string") return "en";
  // 判断日语 根据平假名和片假名
  if (/[\u3040-\u309F\u30A0-\u30FF]/.test(lyric)) return "ja";
  // 判断韩语 根据韩文音节
  if (/[\uAC00-\uD7AF]/.test(lyric)) return "ko";
  // 判断简体中文 根据中日韩统一表意文字基本区
  if (/[\u4E00-\u9FFF]/.test(lyric)) return "zh-CN";
  // 默认英语
  return "en";
};
