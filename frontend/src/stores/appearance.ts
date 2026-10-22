import { defineStore } from "pinia";

export const useAppearanceStore = defineStore("sonora-appearance", {
  state: () => ({ dark: false, menuCollapsed: false }),
  persist: true,
});
