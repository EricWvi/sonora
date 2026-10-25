// 照搬 SPlayer utils/time.ts 中播放器与页面用到的函数。
import dayjs from "dayjs";
import duration from "dayjs/plugin/duration";

dayjs.extend(duration);

/** 毫秒转为时间字符串 (mm:ss 或 H:mm:ss) */
export const msToTime = (milliseconds: number): string => {
  const format = milliseconds < 3600000 ? "mm:ss" : "H:mm:ss";
  return dayjs.duration(milliseconds, "milliseconds").format(format);
};

/**
 * 毫秒转为秒
 * @param milliseconds 毫秒
 * @param decimalPlaces 保留的小数位数
 */
export const msToS = (
  milliseconds: number,
  decimalPlaces: number = 2,
): number => {
  return Number((milliseconds / 1000).toFixed(decimalPlaces));
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

/**
 * 格式化时间戳
 * @param timestamp 时间戳（毫秒）或 ISO 时间字符串
 * @param format 时间格式，默认 "YYYY-MM-DD"
 * @param keepSameYear 是否保留同年年份，默认为 false（即同年省略年份）
 */
export const formatTimestamp = (
  timestamp: number | string | undefined,
  format: string = "YYYY-MM-DD",
  keepSameYear: boolean = false,
): string => {
  if (!timestamp) return "";
  const date = dayjs(timestamp);
  const shouldOmitYear = !keepSameYear && date.year() === dayjs().year();
  return date.format(shouldOmitYear ? format.replace("YYYY-", "") : format);
};

/** 根据当前时间返回问候语 */
export const getGreeting = (): string => {
  const hour = dayjs().hour();
  const greetings: [number, string][] = [
    [6, "凌晨好"],
    [9, "早上好"],
    [12, "上午好"],
    [14, "中午好"],
    [17, "下午好"],
    [19, "傍晚好"],
    [22, "晚上好"],
    [24, "夜深了"],
  ];
  return greetings.find(([limit]) => hour < limit)?.[1] ?? "夜深了";
};
