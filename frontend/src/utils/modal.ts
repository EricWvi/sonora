import type { ArtistRef, Track } from "@/api/types";
import type { SettingType } from "@/types/settings";
import { NScrollbar } from "naive-ui";
import router from "@/router";
import { isFunction } from "lodash-es";

// 照搬 SPlayer utils/modal.ts 中播放器、页面与设置用到的入口。

// 单例弹窗管理：跟踪已打开的弹窗类型
const openedModals = new Set<string>();

/**
 * 检查弹窗是否已打开，若已打开则显示提示
 * @param modalKey 弹窗唯一标识
 * @param warningMessage 已打开时的提示信息
 * @returns 是否已打开
 */
const isModalOpen = (modalKey: string, warningMessage?: string): boolean => {
  if (openedModals.has(modalKey)) {
    if (warningMessage) window.$message.warning(warningMessage);
    return true;
  }
  return false;
};

/**
 * 标记弹窗为打开状态
 */
const setModalOpen = (modalKey: string): void => {
  openedModals.add(modalKey);
};

/**
 * 标记弹窗为关闭状态
 */
const setModalClosed = (modalKey: string): void => {
  openedModals.delete(modalKey);
};

/** 打开歌单界面配置弹窗 */
export const openPlaylistPageManager = async () => {
  const { default: PlaylistPageManager } =
    await import("@/components/Modal/Setting/PlaylistPageManager.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "500px" },
    title: "歌单界面配置",
    content: () => {
      return h(PlaylistPageManager);
    },
  });
};

/** 打开全屏播放器配置弹窗 */
export const openFullscreenPlayerManager = async () => {
  const { default: FullscreenPlayerManager } =
    await import("@/components/Modal/Setting/FullscreenPlayerManager.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "500px" },
    title: "全屏播放器配置",
    content: () => {
      return h(FullscreenPlayerManager);
    },
  });
};

/** 打开右键菜单配置弹窗 */
export const openContextMenuManager = async () => {
  const { default: ContextMenuManager } =
    await import("@/components/Modal/Setting/ContextMenuManager.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "500px" },
    title: "右键菜单配置",
    content: () => {
      return h(ContextMenuManager);
    },
  });
};

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
 * @param type 设置分类
 * @param scrollTo 滚动到的设置项
 */
export const openSetting = async (
  type: SettingType = "general",
  scrollTo?: string,
) => {
  if (isModalOpen("setting", "设置页面已打开")) return;
  setModalOpen("setting");
  const { default: MainSetting } =
    await import("@/components/Setting/MainSetting.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    maskClosable: false,
    closeOnEsc: false,
    bordered: false,
    class: "main-setting",
    content: () => {
      return h(MainSetting, { type, scrollTo });
    },
    onAfterLeave: () => {
      setModalClosed("setting");
    },
  });
};

/** 打开侧边栏显示管理弹窗 */
export const openSidebarHideManager = async () => {
  const { default: SidebarHideManager } =
    await import("@/components/Modal/Setting/SidebarHideManager.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "500px" },
    title: "侧边栏显示管理",
    content: () => {
      return h(SidebarHideManager);
    },
  });
};

/** 打开封面显示配置弹窗 */
export const openCoverManager = async () => {
  const { default: CoverManager } =
    await import("@/components/Modal/Setting/CoverManager.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "500px" },
    title: "封面显示配置",
    content: () => {
      return h(CoverManager);
    },
  });
};

/** 打开首页栏目配置弹窗 */
export const openHomePageSectionManager = async () => {
  const { default: HomePageSectionManager } =
    await import("@/components/Modal/Setting/HomePageSectionManager.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "500px" },
    title: "首页栏目配置",
    content: () => {
      return h(HomePageSectionManager);
    },
  });
};

/** 打开字体设置弹窗 */
export const openFontManager = async () => {
  const { default: FontManager } =
    await import("@/components/Modal/Setting/FontManager.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "700px" },
    title: "字体设置",
    content: () => {
      return h(FontManager);
    },
  });
};

/** 打开主题配置弹窗 */
export const openThemeConfig = async () => {
  if (isModalOpen("themeConfig", "主题配置已打开")) return;
  setModalOpen("themeConfig");
  const { default: ThemeConfig } =
    await import("@/components/Modal/ThemeConfig.vue");
  window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    showMask: false,
    draggable: true,
    style: { width: "500px" },
    title: "主题配置",
    size: "small",
    content: () => {
      return h(ThemeConfig);
    },
    onAfterLeave: () => {
      setModalClosed("themeConfig");
    },
  });
};

/** 打开歌词排除弹窗 */
export const openExcludeLyric = async () => {
  const { default: ExcludeLyrics } =
    await import("@/components/Modal/Setting/ExcludeLyrics.vue");
  const modal = window.$modal.create({
    preset: "card",
    transformOrigin: "center",
    autoFocus: false,
    style: { width: "600px" },
    title: "歌词排除",
    content: () => {
      return h(ExcludeLyrics, {
        onClose: () => modal.destroy(),
      });
    },
  });
};

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
