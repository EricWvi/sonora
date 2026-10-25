import type { ArtistRef, Track } from "@/api/types";
import { NScrollbar } from "naive-ui";
import router from "@/router";
import { isFunction } from "lodash-es";

// 照搬 SPlayer utils/modal.ts 中播放器用到的入口。设置弹窗在阶段 5 接入，之前只提示。

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
  const { default: JumpArtist } =
    await import("@/components/Modal/JumpArtist.vue");
  const modal = window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "600px" },
    title: "跳转到歌手",
    content: () => {
      return h(JumpArtist, { artist: data, onClose: () => modal.destroy() });
    },
  });
};

// 添加到歌单
export const openPlaylistAdd = async (data: Track[]) => {
  if (!data.length) return window.$message.warning("请正确选择歌曲");
  const { default: PlaylistAdd } =
    await import("@/components/Modal/PlaylistAdd.vue");
  const modal = window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "600px" },
    title: "添加到歌单",
    content: () => {
      return h(PlaylistAdd, { data, onClose: () => modal.destroy() });
    },
  });
};

// 新建歌单
export const openCreatePlaylist = async () => {
  const { default: CreatePlaylist } =
    await import("@/components/Modal/CreatePlaylist.vue");
  const modal = window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "600px" },
    title: "新建歌单",
    content: () => {
      return h(CreatePlaylist, { onClose: () => modal.destroy() });
    },
  });
};

// 均衡器
export const openEqualizer = async () => {
  const { default: Equalizer } =
    await import("@/components/Modal/Equalizer.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "620px" },
    title: "均衡器",
    content: () => {
      return h(Equalizer);
    },
  });
};

// 复制歌词
export const openCopyLyrics = async () => {
  const { default: CopyLyrics } =
    await import("@/components/Modal/CopyLyrics.vue");
  const modal = window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "500px" },
    title: "复制歌词",
    content: () => {
      return h(CopyLyrics, {
        onClose: () => modal.destroy(),
      });
    },
  });
};

/**
 * 打开设置
 * @param _type 设置分类
 */
export const openSetting = async (_type?: string) => pending("设置");

/**
 * 打开简介
 * @param content 简介内容
 * @param title 标题
 */
export const openDescModal = (content: string, title: string = "歌单简介") => {
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "600px" },
    title,
    content: () => {
      return h(
        NScrollbar,
        { style: { maxHeight: "400px" } },
        {
          default: () =>
            h(
              "div",
              { style: { whiteSpace: "pre-wrap" } },
              { default: () => content },
            ),
        },
      );
    },
  });
};

/**
 * 编辑歌单
 * @param id 歌单 ID
 * @param data 歌单信息
 * @param func 回调函数
 */
export const openUpdatePlaylist = async (
  id: number,
  data: { name: string; description?: string },
  func: () => Promise<void>,
) => {
  const { default: UpdatePlaylist } =
    await import("@/components/Modal/UpdatePlaylist.vue");
  const modal = window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "600px" },
    title: "编辑歌单",
    content: () => {
      return h(UpdatePlaylist, {
        id,
        data,
        onSuccess: () => {
          modal.destroy();
          // 触发回调
          if (isFunction(func)) func();
        },
      });
    },
  });
};
