import type { Track } from "@/api/types";

/**
 * 队列中的一次出现。同一首歌可以多次入队，每次有自己的 `entryId`；删除、拖动、
 * 随机与恢复都按条目处理（路线图 5.1），比较时用 `entryId`，不比较对象：恢复后
 * 播放列表与原始播放列表中的同一条目是两个对象。SPlayer 的队列直接存歌曲，按歌曲 ID 去重。
 */
export type QueueEntry = Track & {
  /** 条目身份，在本浏览器中递增，不随歌曲变化 */
  entryId: number;
  /** FM 排的歌曲所属的专辑；用户插入的歌曲没有 */
  fmAlbumId?: number;
};

let lastEntryId = 0;

/**
 * 把歌曲包装成新的队列条目
 * @param track 歌曲；已经是条目时也会分配新的身份
 */
export function newEntry(track: Track & { fmAlbumId?: number }): QueueEntry {
  return { ...track, entryId: ++lastEntryId };
}

/** 恢复保存的队列后调用，之后分配的身份不与已有条目重复 */
export function reserveEntryIds(entries: readonly QueueEntry[]) {
  for (const entry of entries)
    lastEntryId = Math.max(lastEntryId, entry.entryId);
}

/**
 * 打乱数组，返回新数组（照搬 SPlayer utils/helper.ts 的 shuffleArray）
 * @param arr 原数组
 */
export function shuffleArray<T>(arr: readonly T[]): T[] {
  const copy = [...arr];
  for (let i = copy.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [copy[i], copy[j]] = [copy[j], copy[i]];
  }
  return copy;
}

/**
 * 随机顺序：指定的条目放在第一位，其余打乱。SPlayer 打乱整个队列，当前曲目
 * 可能落在中间，排在它前面的歌在不循环时就播不到了（2026-10-25 Eric 确定）。
 * @param list 队列
 * @param first 放在第一位的条目；省略时整个队列打乱
 */
export function shuffled(
  list: readonly QueueEntry[],
  first?: QueueEntry,
): QueueEntry[] {
  if (!first) return shuffleArray(list);
  return [
    first,
    ...shuffleArray(list.filter((e) => e.entryId !== first.entryId)),
  ];
}

/**
 * 在指定条目之后插入，返回新数组与插入的位置
 * @param list 队列
 * @param after 插在它之后；不在队列中时插到开头
 * @param entry 新条目
 */
export function insertAfter(
  list: readonly QueueEntry[],
  after: QueueEntry | undefined,
  entry: QueueEntry,
): { list: QueueEntry[]; index: number } {
  const index = after
    ? list.findIndex((e) => e.entryId === after.entryId) + 1
    : 0;
  const next = [...list];
  next.splice(index, 0, entry);
  return { list: next, index };
}

/**
 * 移除条目，返回新数组
 * @param list 队列
 * @param entry 要移除的条目
 */
export function without(
  list: readonly QueueEntry[],
  entry: QueueEntry,
): QueueEntry[] {
  return list.filter((e) => e.entryId !== entry.entryId);
}

/**
 * 移动条目后当前播放的位置（照搬 SPlayer PlayerController.moveSong 的计算）
 * @param playIndex 当前播放索引
 * @param fromIndex 移动前索引
 * @param toIndex 移动后索引
 */
export function indexAfterMove(
  playIndex: number,
  fromIndex: number,
  toIndex: number,
): number {
  if (playIndex === fromIndex) return toIndex;
  if (fromIndex < playIndex && toIndex >= playIndex) return playIndex - 1;
  if (fromIndex > playIndex && toIndex <= playIndex) return playIndex + 1;
  return playIndex;
}

/**
 * 移除条目后当前播放的位置。SPlayer 在移除最后一首时总是回到 0，即使当前播放的不是它，
 * 界面上的当前曲目因此跳到第一首而音频不变；Sonora 只在移除的正是当前曲目时才回到 0。
 * @param playIndex 当前播放索引
 * @param index 被移除的索引
 * @param length 移除前的队列长度
 */
export function indexAfterRemove(
  playIndex: number,
  index: number,
  length: number,
): number {
  if (index < playIndex) return playIndex - 1;
  if (index === playIndex && index === length - 1) return 0;
  return playIndex;
}
