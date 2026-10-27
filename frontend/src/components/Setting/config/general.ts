// 照搬 SPlayer components/Setting/config/general.ts，条目按阶段 2 的设置表与阶段 5 的决定：
// 保留搜索设置（搜索历史、搜索建议、搜索框行为、隐藏括号与别名）与重置；不做系统行为、热搜榜、
// 评论排除、分享链接格式、备份与恢复。搜索建议来自曲库，在网页中也显示（SPlayer 只在开启
// 在线服务时显示），说明随之修改。
// “重置所有设置”会上传默认值，影响所有设备；“清除全部数据”只清浏览器中的数据并重置设置，
// 不动服务端的歌单、喜爱与历史，也不退出登录（SPlayer 会退出登录）。
import {
  useDataStore,
  useSettingStore,
  useShortcutStore,
  useStatusStore,
} from "@/stores";
import { usePlayerController } from "@/core/player/PlayerController";
import { useSettingsSync } from "@/core/settings/sync";
import type { SettingConfig } from "@/types/settings";

export const useGeneralSettings = (): SettingConfig => {
  const dataStore = useDataStore();
  const settingStore = useSettingStore();
  const shortcutStore = useShortcutStore();
  const statusStore = useStatusStore();
  const player = usePlayerController();

  // --- Reset Logic (from other.ts) ---
  const resetSetting = () => {
    window.$dialog.warning({
      title: "警告",
      content: "此操作将重置所有设置，是否继续?",
      positiveText: "确定",
      negativeText: "取消",
      onPositiveClick: () => {
        settingStore.$reset();
        window.$message.success("设置重置完成");
      },
    });
  };

  const clearAllData = () => {
    window.$dialog.warning({
      title: "高危操作",
      content: "此操作将重置所有设置并清除浏览器中的全部数据，是否继续?",
      positiveText: "确定",
      negativeText: "取消",
      onPositiveClick: async () => {
        // 先把默认设置上传，否则刷新后会读回服务端的设置
        settingStore.$reset();
        shortcutStore.$reset();
        await useSettingsSync()?.flush();
        // 播放队列与位置（IndexedDB）
        await player.cleanPlayList();
        // 搜索历史、均衡器与音量等
        dataStore.$reset();
        statusStore.$reset();
        for (const key of [
          "sonora-setting",
          "sonora-shortcut",
          "sonora-status",
          "sonora-data",
        ])
          window.localStorage.removeItem(key);
        window.sessionStorage.clear();
        window.$message.loading("数据清除完成，软件即将热重载", {
          duration: 3000,
          onAfterLeave: () => window.location.reload(),
        });
      },
    });
  };

  return {
    groups: [
      {
        title: "搜索设置",
        items: [
          {
            key: "showSearchHistory",
            label: "显示搜索历史",
            description: "是否在搜索框的默认显示内容中显示当前搜索历史",
            type: "switch",
            value: computed({
              get: () => settingStore.showSearchHistory,
              set: (v) => (settingStore.showSearchHistory = v),
            }),
          },
          {
            key: "enableSearchKeyword",
            label: "搜索关键词建议",
            type: "switch",
            description:
              "输入时在搜索框下方显示曲库中匹配的歌曲、专辑、艺术家与歌单",
            value: computed({
              get: () => settingStore.enableSearchKeyword,
              set: (v) => (settingStore.enableSearchKeyword = v),
            }),
          },
          {
            key: "searchInputBehavior",
            label: "搜索框行为",
            type: "select",
            description: "自定义搜索框的行为模式",
            options: [
              { label: "保留搜索词", value: "normal" },
              { label: "失焦后清空", value: "clear" },
              { label: "同步搜索词", value: "sync" },
            ],
            value: computed({
              get: () => settingStore.searchInputBehavior,
              set: (v) => (settingStore.searchInputBehavior = v),
            }),
          },
          {
            key: "hideBracketedContent",
            label: "隐藏括号与别名",
            type: "switch",
            description: "隐藏歌曲名与专辑名中的括号内容和别名",
            value: computed({
              get: () => settingStore.hideBracketedContent,
              set: (v) => (settingStore.hideBracketedContent = v),
            }),
          },
        ],
      },
      {
        title: "重置",
        items: [
          {
            key: "resetSetting",
            label: "重置所有设置",
            type: "button",
            description: "重置所有设置，恢复软件默认值",
            buttonLabel: "重置设置",
            action: resetSetting,
            componentProps: { type: "warning" },
          },
          {
            key: "clearAllData",
            label: "清除全部数据",
            type: "button",
            description: "重置所有设置，清除浏览器中的全部数据",
            buttonLabel: "清除全部",
            action: clearAllData,
            componentProps: { type: "error" },
          },
        ],
      },
    ],
  };
};
