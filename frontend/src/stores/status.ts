import { defineStore } from "pinia";
import type { ColorScheme, RGB } from "@/types/color";

/**
 * 运行状态。字段名与默认值照搬 SPlayer stores/status.ts，按阶段逐步迁入：
 * 目前只有侧栏与主题需要的字段，播放器的状态在阶段 3.5 加入。
 */
export interface StatusState {
  /** 菜单折叠状态 */
  menuCollapsed: boolean;
  /** 封面主题 */
  songCoverTheme: {
    /** 封面主题颜色 */
    main?: RGB;
    /** 封面主题颜色（亮色） */
    light?: ColorScheme;
    /** 封面主题颜色（暗色） */
    dark?: ColorScheme;
  };
  /** 主内容高度 */
  mainContentHeight: number;
}

export const useStatusStore = defineStore("status", {
  state: (): StatusState => ({
    menuCollapsed: false,
    songCoverTheme: {},
    mainContentHeight: 0,
  }),
  getters: {
    /** 播放器主色 */
    mainColor(state) {
      const mainColor = state.songCoverTheme?.main;
      if (!mainColor) return "239, 239, 239";
      return `${mainColor.r}, ${mainColor.g}, ${mainColor.b}`;
    },
  },
  persist: {
    key: "sonora-status",
    storage: localStorage,
    pick: ["menuCollapsed", "songCoverTheme"],
  },
});
