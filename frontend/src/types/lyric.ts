import type { LyricLine } from "@applemusic-like-lyrics/lyric";

/**
 * 歌词数据类型（照搬 SPlayer types/lyric.d.ts）
 */
export interface SongLyric {
  lrcData: LyricLine[];
  yrcData: LyricLine[];
}
