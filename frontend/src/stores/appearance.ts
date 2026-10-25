import { defineStore } from "pinia";

// 字段名与默认值沿用 SPlayer 的 setting/status store，阶段 3.2 迁入完整设置。
export const useAppearanceStore = defineStore("sonora-appearance", {
  state: () => ({
    dark: false,
    menuCollapsed: false,
    /** 侧栏歌单显示封面 */
    menuShowCover: true,
    /** 侧栏展开的分组 */
    menuExpandedKeys: [] as string[],
  }),
  persist: true,
});
