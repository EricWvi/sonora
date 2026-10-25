import { source } from "@/api/source";
import type { Cover } from "@/api/types";

/** SPlayer 的封面尺寸 `coverSize` 对应的服务端缩略图；`xl` 与 `cover` 使用原图。 */
export const coverSizes = {
  s: 100,
  m: 300,
  l: 1024,
  xl: undefined,
  cover: undefined,
} as const;

/** SPlayer 的默认图片，按对象类型区分。 */
export const defaultCovers = {
  song: "/images/song.jpg",
  album: "/images/album.jpg",
  artist: "/images/artist.jpg",
} as const;

/**
 * 封面地址；没有封面时为 SPlayer 的默认图片
 * @param cover 封面
 * @param size SPlayer 的封面尺寸
 * @param fallback 默认图片
 */
export const coverUrl = (
  cover: Cover | null | undefined,
  size: keyof typeof coverSizes = "m",
  fallback: string = defaultCovers.song,
): string => (cover ? source.coverUrl(cover, coverSizes[size]) : fallback);
