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

/**
 * 检测所有输入行的共同最小缩进，将其从每一行中删除，如果第一行和最后一行是空白行，也将其删除
 * @param string 字符串
 * @param lineSplit 分割时的换行符
 * @param lineJoin 连接时的换行符
 * @returns 去除缩进后的字符串
 */
export const trimIndentString = (
  string: string,
  lineSplit: string = "\n",
  lineJoin: string = lineSplit,
): string => {
  if (!string) return "";
  const lines = string.split(lineSplit);
  // 删除第一行和最后一行的空白行
  const relevantLines = lines.filter(
    (line, index) =>
      (index !== 0 && index !== lines.length - 1) || line.trim() !== "",
  );
  // 移除每行的最小缩进
  const minIndent = relevantLines
    .filter((line) => line.trim() !== "")
    .map((line) => line.match(/^\s*/)?.[0].length ?? 0)
    .reduce((min, indent) => Math.min(min, indent), Infinity);
  const trimmedLines = relevantLines.map((line) => line.slice(minIndent));
  return trimmedLines.join(lineJoin);
};

/**
 * 设置中多行描述的模板标签功能
 * 删除最小公共缩进并将换行符转换为 HTML <br> 标记
 *
 * @see trimIndentString
 */
export const descMultiline = (
  strings: TemplateStringsArray,
  ...values: unknown[]
): string => {
  const fullString = String.raw(strings, ...values);
  return trimIndentString(fullString, "\n", "<br>");
};
