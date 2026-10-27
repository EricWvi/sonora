// 照搬 SPlayer components/Setting/config/keyboard.ts。SPlayer 网页版在页面内响应全部快捷键，
// 设置里只能改“页面内快捷键”三项；Sonora 去掉全局快捷键（开关与“全局”输入框）和桌面歌词，
// 原“全局快捷键更改”一组改为“播放控制”，在页面内生效并可修改（阶段 5 的决定）。
import { useShortcutStore } from "@/stores";
import type { ShortcutList } from "@/stores/shortcut";
import type { SettingItem, SettingConfig } from "@/types/settings";
import { markRaw } from "vue";
import ShortcutRecorder from "../components/ShortcutRecorder.vue";

export const useKeyboardSettings = (): SettingConfig => {
  const shortcutStore = useShortcutStore();

  const createShortcutItems = (filterKeys: string[]): SettingItem[] => {
    return Object.entries(shortcutStore.shortcutList)
      .filter(([key]) => filterKeys.includes(key))
      .map(([key, item]) => ({
        key,
        label: item.name,
        type: "custom",
        component: markRaw(ShortcutRecorder),
        componentProps: { shortcutKey: key as keyof ShortcutList },
      }));
  };

  // 页面快捷键的 Key
  const pageShortcutKeys = ["openPlayer", "openPlayList", "closePlayer"];
  // 播放控制快捷键的 Key（SPlayer 的全局快捷键）
  const playShortcutKeys = Object.keys(shortcutStore.shortcutList).filter(
    (key) => !pageShortcutKeys.includes(key),
  );

  return {
    groups: [
      {
        title: "播放控制",
        items: createShortcutItems(playShortcutKeys),
      },
      {
        title: "恢复默认",
        items: [
          {
            key: "resetShortcut",
            label: "恢复默认快捷键",
            type: "button",
            buttonLabel: "恢复默认",
            action: () => {
              window.$dialog.warning({
                title: "重置快捷键",
                content: "确定重置当前快捷键配置？",
                positiveText: "重置",
                negativeText: "取消",
                onPositiveClick: () => {
                  shortcutStore.$reset();
                  window.$message.success("快捷键重置成功");
                },
              });
            },
          },
        ],
      },
      {
        title: "页面内快捷键",
        items: createShortcutItems(pageShortcutKeys),
      },
    ],
  };
};
