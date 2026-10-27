import { defineStore } from "pinia";

/**
 * 快捷键。照搬 SPlayer stores/shortcut.ts 的按键与默认值，去掉桌面歌词与全局快捷键：
 * 与 SPlayer 网页版相同，全部快捷键只在页面内响应（见 composables/useShortcut.ts）。
 */
export type ShortcutType = {
  name: string;
  shortcut: string;
};

export interface ShortcutList {
  playOrPause: ShortcutType;
  playPrev: ShortcutType;
  playNext: ShortcutType;
  seekForward: ShortcutType;
  seekBackward: ShortcutType;
  volumeUp: ShortcutType;
  volumeDown: ShortcutType;
  openPlayer: ShortcutType;
  openPlayList: ShortcutType;
  closePlayer: ShortcutType;
}

interface ShortcutStore {
  shortcutList: ShortcutList;
}

/** 默认快捷键，同步时用来补齐与校验服务端的设置 */
export const defaultShortcutState = (): ShortcutStore => ({
  // 全部快捷键
  shortcutList: {
    // 播放或暂停
    playOrPause: {
      name: "播放 / 暂停",
      shortcut: "CmdOrCtrl+Space",
    },
    // 上一曲 / 下一曲
    playPrev: {
      name: "上一曲",
      shortcut: "CmdOrCtrl+ArrowLeft",
    },
    playNext: {
      name: "下一曲",
      shortcut: "CmdOrCtrl+ArrowRight",
    },
    // 快进 / 快退
    seekForward: {
      name: "快进 5 秒",
      shortcut: "ArrowRight",
    },
    seekBackward: {
      name: "快退 5 秒",
      shortcut: "ArrowLeft",
    },
    // 音量加减
    volumeUp: {
      name: "音量加",
      shortcut: "CmdOrCtrl+ArrowUp",
    },
    volumeDown: {
      name: "音量减",
      shortcut: "CmdOrCtrl+ArrowDown",
    },
    // 打开播放界面
    openPlayer: {
      name: "打开播放界面",
      shortcut: "KeyP",
    },
    // 打开播放列表
    openPlayList: {
      name: "打开播放列表",
      shortcut: "KeyL",
    },
    // 关闭播放界面
    closePlayer: {
      name: "关闭播放界面",
      shortcut: "Escape",
    },
  },
});

export const useShortcutStore = defineStore("shortcut", {
  state: defaultShortcutState,
  // 持久化
  persist: {
    key: "sonora-shortcut",
    storage: localStorage,
  },
});
