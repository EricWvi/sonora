// 照搬 SPlayer utils/time.ts 中播放器用到的函数。
import dayjs from "dayjs";
import duration from "dayjs/plugin/duration";

dayjs.extend(duration);

/** 毫秒转为时间字符串 (mm:ss 或 H:mm:ss) */
export const msToTime = (milliseconds: number): string => {
  const format = milliseconds < 3600000 ? "mm:ss" : "H:mm:ss";
  return dayjs.duration(milliseconds, "milliseconds").format(format);
};

/**
 * 计算进度条百分比
 * @param currentTime 当前时间
 * @param duration 总时长
 * @returns 进度百分比，精确到 0.01，范围 0-100
 */
export const calculateProgress = (
  currentTime: number,
  duration: number,
): number => {
  if (duration === 0) return 0;
  const progress = (currentTime / duration) * 100;
  return Math.min(Math.round(progress * 100) / 100, 100);
};
