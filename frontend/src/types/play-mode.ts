// 照搬 SPlayer types/shared/play-mode.ts；Sonora 不做心动模式，随机模式只有开与关。

/**
 * 循环模式
 *
 * off: 不循环 | list: 列表循环 | one: 单曲循环
 */
export type RepeatModeType = "off" | "list" | "one";

/**
 * 随机模式
 *
 * off: 关闭 | on: 随机播放
 */
export type ShuffleModeType = "off" | "on";
