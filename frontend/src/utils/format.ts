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
