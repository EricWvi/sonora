import { ApiError, loginUrl, source } from "@/api/source";
import type { Album, Artist, FillField, Track } from "@/api/types";
import { usePlayerController } from "@/core/player/PlayerController";

// 从关联的网易云内容补全本地的简介、图片与歌词（Sonora 新增，Eric，2026-11-07；SPlayer 没有本地
// 与在线的关联）。入口只对有导入权限的用户、在内容为空时显示，服务端也只写空的字段。

/** 正在补全的内容，重复点击不再发请求 */
const running = new Set<string>();

/**
 * 发请求并在进行中显示提示
 * @param key 内容的标识
 * @param request 请求
 * @returns 响应；失败或已在进行中时为 null
 */
async function fill<T>(
  key: string,
  request: () => Promise<T>,
): Promise<T | null> {
  if (running.has(key)) return null;
  running.add(key);
  const loading = window.$message.loading("正在从网易云获取…", { duration: 0 });
  try {
    return await request();
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      location.assign(loginUrl);
      return null;
    }
    window.$message.error(
      error instanceof ApiError ? error.message : "补全失败，请重试",
    );
    console.error("❌ 补全失败", error);
    return null;
  } finally {
    running.delete(key);
    loading.destroy();
  }
}

/**
 * 补全专辑简介
 * @param id 专辑
 * @returns 更新后的专辑；没有变化时为 null
 */
export const fillAlbum = async (id: number): Promise<Album | null> => {
  const result = await fill(`album-${id}`, () => source.fillAlbum(id));
  if (!result) return null;
  if (!result.filled.length) {
    window.$message.info("网易云上也没有这张专辑的简介");
    return null;
  }
  window.$message.success("已补全专辑简介");
  return result.album;
};

const FIELD_NAMES: Record<FillField, string> = {
  image: "图片",
  description: "简介",
  lyrics: "歌词",
};

/** 艺术家补全结果的提示：有一项网易云上也没有时一并说明 */
export const artistFillMessage = (
  filled: FillField[],
  missing: FillField[],
): string => {
  const names = (fields: FillField[]) =>
    fields.map((field) => FIELD_NAMES[field]).join("和");
  const lacking = missing.filter((field) => !filled.includes(field));
  if (!filled.length) return "网易云上也没有可以补全的资料";
  return lacking.length
    ? `已补全${names(filled)}，网易云上没有${names(lacking)}`
    : `已补全${names(filled)}`;
};

/**
 * 补全艺术家的图片与简介
 * @param artist 艺术家
 * @returns 更新后的艺术家；没有变化时为 null
 */
export const fillArtist = async (artist: Artist): Promise<Artist | null> => {
  const result = await fill(`artist-${artist.id}`, () =>
    source.fillArtist(artist.id),
  );
  if (!result) return null;
  const missing: FillField[] = [];
  if (!artist.image) missing.push("image");
  if (!artist.description) missing.push("description");
  const message = artistFillMessage(result.filled, missing);
  if (!result.filled.length) {
    window.$message.info(message);
    return null;
  }
  window.$message.success(message);
  return result.artist;
};

/**
 * 补全歌词。歌曲正在播放时立即显示
 * @param song 歌曲
 */
export const fillLyrics = async (song: Track) => {
  const result = await fill(`track-${song.id}`, () =>
    source.fillTrack(song.id),
  );
  if (!result) return;
  if (!result.filled.length) {
    window.$message.info("网易云上也没有这首歌的歌词");
    return;
  }
  // 列表里的歌曲对象决定下次打开菜单时是否还显示这一项
  song.hasLyrics = true;
  usePlayerController().lyricsAdded(song);
  window.$message.success("已补全歌词");
};
