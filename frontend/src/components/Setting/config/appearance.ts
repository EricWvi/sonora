// 照搬 SPlayer components/Setting/config/appearance.ts，条目按阶段 2 的设置表与阶段 5 的补充：
// 去掉无边框窗口、自定义代码、动态封面、评论、音质与特权等标签；“展示播放状态信息”控制的
// 歌名下整行标签已不显示，一并去掉。说明文字中 Sonora 没有的功能（自定义图片、下载、评论、
// 歌单广场等）换成 Sonora 的对应项。音乐频谱在网页中也显示（SPlayer 只在 Electron 中）。
import { useSettingStore } from "@/stores";
import {
  openFontManager,
  openThemeConfig,
  openSidebarHideManager,
  openHomePageSectionManager,
  openPlaylistPageManager,
  openFullscreenPlayerManager,
  openCoverManager,
  openContextMenuManager,
} from "@/utils/modal";
import type { SettingConfig } from "@/types/settings";
import { computed } from "vue";

export const useAppearanceSettings = (): SettingConfig => {
  const settingStore = useSettingStore();

  return {
    groups: [
      {
        title: "主题与风格",
        items: [
          {
            key: "themeMode",
            label: "主题模式",
            type: "select",
            description: "调整全局主题明暗模式",
            options: [
              { label: "跟随系统", value: "auto" },
              { label: "浅色模式", value: "light" },
              { label: "深色模式", value: "dark" },
            ],
            value: computed({
              get: () => settingStore.themeMode,
              set: (v) => (settingStore.themeMode = v),
            }),
          },
          {
            key: "themeConfig",
            label: "主题配置",
            type: "button",
            description: "更改主题色",
            buttonLabel: "配置",
            action: openThemeConfig,
          },
          {
            key: "fontConfig",
            label: "全局字体",
            type: "button",
            description: "统一配置全局及歌词区域的字体",
            buttonLabel: "配置",
            action: openFontManager,
          },
        ],
      },
      {
        title: "界面布局",
        items: [
          {
            key: "sidebarHide",
            label: "侧边栏管理",
            type: "button",
            description: "配置需要在侧边栏显示的菜单项",
            buttonLabel: "配置",
            action: openSidebarHideManager,
          },
          {
            key: "homePageSection",
            label: "首页栏目",
            type: "button",
            description: "调整首页各栏目的显示顺序或隐藏不需要的栏目",
            buttonLabel: "配置",
            action: openHomePageSectionManager,
          },
          {
            key: "playlistPageElements",
            label: "歌单界面",
            type: "button",
            description: "自定义歌单界面的拥有者、时间、描述显示",
            buttonLabel: "配置",
            action: openPlaylistPageManager,
          },
          {
            key: "fullscreenPlayer",
            label: "全屏播放器",
            type: "button",
            description:
              "自定义全屏播放器的显示元素（喜欢、添加到歌单、歌词设置等）",
            buttonLabel: "配置",
            action: openFullscreenPlayerManager,
          },
          {
            key: "contextMenu",
            label: "右键菜单",
            type: "button",
            description: "自定义歌曲右键菜单的显示选项",
            buttonLabel: "配置",
            action: openContextMenuManager,
          },
          {
            key: "menuShowCover",
            label: "侧边栏显示歌单封面",
            type: "switch",
            description: "是否在侧边栏显示歌单的封面（如有）",
            value: computed({
              get: () => settingStore.menuShowCover,
              set: (v) => (settingStore.menuShowCover = v),
            }),
          },
          {
            key: "showPlaylistCount",
            label: "显示播放列表数量",
            type: "switch",
            description: "在右下角的播放列表按钮处显示播放列表的歌曲数量",
            value: computed({
              get: () => settingStore.showPlaylistCount,
              set: (v) => (settingStore.showPlaylistCount = v),
            }),
          },
          {
            key: "routeAnimation",
            label: "页面切换动画",
            type: "select",
            description: "选择页面切换时的动画效果",
            options: [
              { label: "无动画", value: "none" },
              { label: "淡入淡出", value: "fade" },
              { label: "缩放", value: "zoom" },
              { label: "滑动", value: "slide" },
              { label: "上浮", value: "up" },
              { label: "流体", value: "flow" },
              { label: "左右遮罩", value: "mask-left" },
              { label: "上下遮罩", value: "mask-top" },
            ],
            value: computed({
              get: () => settingStore.routeAnimation,
              set: (v) => (settingStore.routeAnimation = v),
            }),
          },
        ],
      },
      {
        title: "播放器外观",
        items: [
          {
            key: "playerType",
            label: "播放器样式",
            type: "select",
            description: "播放器主体样式",
            options: [
              { label: "封面模式", value: "cover" },
              { label: "唱片模式", value: "record" },
              { label: "全屏封面", value: "fullscreen" },
            ],
            value: computed({
              get: () => settingStore.playerType,
              set: (v) => (settingStore.playerType = v),
            }),
            condition: () => true,
            children: computed(() => {
              const type = settingStore.playerType;
              if (type === "cover" || type === "record") {
                return [
                  {
                    key: "playerStyleRatio",
                    label: "封面 / 歌词占比",
                    type: "slider",
                    description: "调整全屏播放器的封面与歌词的宽度比例",
                    min: 30,
                    max: 70,
                    step: 1,
                    marks: { 50: "默认" },
                    formatTooltip: (v) => `${v}%`,
                    value: computed({
                      get: () => settingStore.playerStyleRatio,
                      set: (v) => (settingStore.playerStyleRatio = v),
                    }),
                  },
                ];
              }
              if (type === "fullscreen") {
                return [
                  {
                    key: "playerFullscreenGradient",
                    label: "封面过渡位置",
                    type: "slider",
                    description: "调整全屏封面右侧的渐变过渡位置",
                    min: 0,
                    max: 100,
                    step: 1,
                    marks: { 15: "默认" },
                    formatTooltip: (v) => `${v}%`,
                    value: computed({
                      get: () => settingStore.playerFullscreenGradient,
                      set: (v) => (settingStore.playerFullscreenGradient = v),
                    }),
                  },
                ];
              }
              return [];
            }),
          },
          {
            key: "playerBackgroundType",
            label: "播放器背景样式",
            type: "select",
            description: "切换播放器背景类型",
            options: [
              { label: "流体效果", value: "animation" },
              { label: "封面模糊", value: "blur" },
              { label: "封面主色", value: "color" },
            ],
            value: computed({
              get: () => settingStore.playerBackgroundType,
              set: (v) => (settingStore.playerBackgroundType = v),
            }),
            condition: () => settingStore.playerBackgroundType === "animation",
            children: [
              {
                key: "playerBackgroundFps",
                label: "背景动画帧率",
                type: "input-number",
                description: "单位 fps，最小 24，最大 240",
                min: 24,
                max: 256,
                show: () => settingStore.playerBackgroundType === "animation",
                value: computed({
                  get: () => settingStore.playerBackgroundFps,
                  set: (v) => (settingStore.playerBackgroundFps = v),
                }),
              },
              {
                key: "playerBackgroundFlowSpeed",
                label: "背景动画流动速度",
                type: "input-number",
                description: "单位 倍数，最小 0.1，最大 10",
                min: 0.1,
                max: 10,
                show: () => settingStore.playerBackgroundType === "animation",
                value: computed({
                  get: () => settingStore.playerBackgroundFlowSpeed,
                  set: (v) => (settingStore.playerBackgroundFlowSpeed = v),
                }),
              },
              {
                key: "playerBackgroundRenderScale",
                label: "背景渲染缩放比例",
                type: "input-number",
                description:
                  "设置当前渲染缩放比例，默认 0.5。适当提高此值（如 1.0 或 1.5）可以减少分界线锯齿，让效果更好，但也会增加显卡压力",
                min: 0.1,
                max: 3,
                show: () => settingStore.playerBackgroundType === "animation",
                value: computed({
                  get: () => settingStore.playerBackgroundRenderScale,
                  set: (v) => (settingStore.playerBackgroundRenderScale = v),
                }),
              },
              {
                key: "playerBackgroundPause",
                label: "背景动画暂停时暂停",
                type: "switch",
                description: "在暂停时是否也暂停背景动画",
                show: () => settingStore.playerBackgroundType === "animation",
                value: computed({
                  get: () => settingStore.playerBackgroundPause,
                  set: (v) => (settingStore.playerBackgroundPause = v),
                }),
              },
              {
                key: "playerBackgroundLowFreqVolume",
                label: "背景跳动效果",
                type: "switch",
                description: "使流体背景根据音乐低频节拍产生脉动效果",
                show: () => settingStore.playerBackgroundType === "animation",
                value: computed({
                  get: () => settingStore.playerBackgroundLowFreqVolume,
                  set: (v) => (settingStore.playerBackgroundLowFreqVolume = v),
                }),
              },
            ],
          },
          {
            key: "playerExpandAnimation",
            label: "播放器展开动画",
            type: "select",
            description: "选择播放器展开时的动画效果",
            options: [
              { label: "上浮", value: "up" },
              { label: "流体", value: "flow" },
            ],
            value: computed({
              get: () => settingStore.playerExpandAnimation,
              set: (v) => (settingStore.playerExpandAnimation = v),
            }),
          },
          {
            key: "playerFollowCoverColor",
            label: "播放器主色跟随封面",
            type: "switch",
            description: "播放器主颜色是否跟随封面主色，下一曲生效",
            value: computed({
              get: () => settingStore.playerFollowCoverColor,
              set: (v) => (settingStore.playerFollowCoverColor = v),
            }),
          },
          {
            key: "showSpectrums",
            label: "音乐频谱",
            type: "switch",
            description: "开启音乐频谱会影响性能或增加内存占用，如遇问题请关闭",
            value: computed({
              get: () => settingStore.showSpectrums,
              set: (v) => (settingStore.showSpectrums = v),
            }),
          },
        ],
      },
      {
        title: "界面元素显示",
        items: [
          {
            key: "coverManager",
            label: "封面显示管理",
            type: "button",
            description: "配置各界面封面是否显示（如首页、专辑、播放器等）",
            buttonLabel: "配置",
            action: openCoverManager,
          },
          {
            key: "autoHidePlayerMeta",
            label: "播放器元素自动隐藏",
            type: "switch",
            description: "鼠标静止一段时间或者离开播放器时自动隐藏控制元素",
            value: computed({
              get: () => settingStore.autoHidePlayerMeta,
              set: (v) => (settingStore.autoHidePlayerMeta = v),
            }),
          },
          {
            key: "barLyricShow",
            label: "底栏显示歌词",
            type: "switch",
            description: "在播放时将歌手信息更改为歌词",
            value: computed({
              get: () => settingStore.barLyricShow,
              set: (v) => (settingStore.barLyricShow = v),
            }),
          },
          {
            key: "countDownShow",
            label: "显示前奏倒计时",
            type: "switch",
            description: "部分歌曲前奏可能存在显示错误",
            value: computed({
              get: () => settingStore.countDownShow,
              set: (v) => (settingStore.countDownShow = v),
            }),
          },
          {
            key: "timeFormat",
            label: "时间显示格式",
            type: "select",
            description:
              "底栏右侧和播放页面底部的时间如何显示（单击时间可以快速切换）",
            options: [
              { label: "播放时间 / 总时长", value: "current-total" },
              { label: "剩余时间 / 总时长", value: "remaining-total" },
              { label: "播放时间 / 剩余时间", value: "current-remaining" },
            ],
            value: computed({
              get: () => settingStore.timeFormat,
              set: (v) => (settingStore.timeFormat = v),
            }),
          },
        ],
      },
      {
        title: "歌曲列表显示",
        items: [
          {
            key: "showSongAlbum",
            label: "显示专辑",
            type: "switch",
            description: "在歌曲列表中显示专辑列",
            value: computed({
              get: () => settingStore.showSongAlbum,
              set: (v) => (settingStore.showSongAlbum = v),
            }),
          },
          {
            key: "showSongArtist",
            label: "显示歌手",
            type: "switch",
            description: "在歌曲列表中显示歌手信息",
            value: computed({
              get: () => settingStore.showSongArtist,
              set: (v) => (settingStore.showSongArtist = v),
            }),
          },
          {
            key: "showSongDuration",
            label: "显示时长",
            type: "switch",
            description: "在歌曲列表中显示时长列",
            value: computed({
              get: () => settingStore.showSongDuration,
              set: (v) => (settingStore.showSongDuration = v),
            }),
          },
          {
            key: "showSongOperations",
            label: "显示操作",
            type: "switch",
            description: "在歌曲列表中显示操作列（收藏等）",
            value: computed({
              get: () => settingStore.showSongOperations,
              set: (v) => (settingStore.showSongOperations = v),
            }),
          },
          {
            key: "hideBracketedContent",
            label: "隐藏括号内容",
            type: "switch",
            description: "隐藏括号内的内容，如 (Live)、(伴奏) 等",
            value: computed({
              get: () => settingStore.hideBracketedContent,
              set: (v) => (settingStore.hideBracketedContent = v),
            }),
          },
        ],
      },
    ],
  };
};
