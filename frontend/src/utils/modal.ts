import type { ArtistRef, Track } from "@/api/types";
import router from "@/router";

// 照搬 SPlayer utils/modal.ts 中播放器用到的入口。弹窗（选择艺术家、添加到歌单、
// 复制更多信息、均衡器、复制歌词、设置）在阶段 5 接入，之前只提示。

const pending = (name: string) =>
  window.$message.info(`${name}将在后续阶段接入`);

/**
 * 跳转到歌手：一两位歌手时直接跳转，更多时选择（与 SPlayer 相同）
 * @param data 歌手列表
 * @param id 点击的歌手
 */
export const openJumpArtist = async (data: ArtistRef[], id?: number) => {
  if (data.length <= 2 && id) {
    router.push({ name: "artist", query: { id } });
    return;
  }
  pending("选择歌手");
};

// 添加到歌单
export const openPlaylistAdd = async (data: Track[]) => {
  if (!data.length) return window.$message.warning("请正确选择歌曲");
  pending("添加到歌单");
};

// 复制更多信息
export const openCopySongInfo = async (_songId: number) =>
  pending("复制更多信息");

// 均衡器
export const openEqualizer = async () => pending("均衡器");

// 复制歌词
export const openCopyLyrics = async () => pending("复制歌词");

/**
 * 打开设置
 * @param _type 设置分类
 */
export const openSetting = async (_type?: string) => pending("设置");
