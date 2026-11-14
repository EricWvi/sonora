// 照搬 SPlayer components/Setting/config/play.ts，条目按阶段 2 的设置表与阶段 5 的决定：
// - 去掉自动混音、音质、Fuck AI/DJ/*** Mode、播放引擎、Web Audio 延迟策略、播放试听与音乐解锁。
// - 音量平衡与 SPlayer 不同（Eric，2026-11-14）：默认开启，放在“音频设置”；增益由服务端按响度算好，
//   不需要歌曲标签，没有“平衡模式”子项，说明随之修改。
// - 自动播放与音频输出设备在网页中也显示（SPlayer 只在 Electron 中）。输出设备只走 Web Audio 的分支；
//   Chrome 要有麦克风权限才列出设备，第一次展开下拉框且设备名为空时请求一次，拿到后立即停用。
// - “下一首歌曲预载”在 Sonora 中预加载下一首的音频，说明随之修改。
// - “音频与歌词同步补偿”始终显示、始终生效（SPlayer 只在延迟策略为 playback 时），不做自动补偿。
import { useSettingStore } from "@/stores";
import { usePlayerController } from "@/core/player/PlayerController";
import { renderOption } from "@/utils/helper";
import { openFmGroupManager } from "@/utils/modal";
import type { SettingConfig } from "@/types/settings";
import type { SelectOption } from "naive-ui";
import { uniqBy } from "lodash-es";

import { computed, ref } from "vue";

export const usePlaySettings = (): SettingConfig => {
  const settingStore = useSettingStore();
  const player = usePlayerController();

  const outputDevices = ref<SelectOption[]>([]);
  // 是否已请求过麦克风权限
  let permissionRequested = false;

  // 获取全部输出设备
  const getOutputDevices = async () => {
    try {
      const allDevices = await navigator.mediaDevices.enumerateDevices();
      const devices = uniqBy(
        allDevices.filter(
          (device) => device.kind === "audiooutput" && device.deviceId,
        ),
        "groupId",
      );
      const outputData = devices.filter(
        (device) => device.kind === "audiooutput",
      );
      outputDevices.value = outputData.length
        ? outputData.map((device) => ({
            label: device.label,
            value: device.deviceId,
          }))
        : // 没有权限时 Chrome 不给出设备，只能使用默认设备
          [{ label: "默认", value: "default" }];

      // 验证已保存的设备是否在当前设备列表中
      if (
        outputData.length &&
        settingStore.playDevice &&
        !outputData.some((d) => d.deviceId === settingStore.playDevice)
      ) {
        settingStore.playDevice = "default";
      }
    } catch (e) {
      console.error("获取 WebAudio 设备失败", e);
    }
  };

  // 第一次展开下拉框时，如果设备名为空，请求麦克风权限后重新列出
  const requestDevicePermission = async (show: boolean) => {
    if (!show || permissionRequested) return;
    if (outputDevices.value.some((d) => d.value !== "default" && d.label))
      return;
    permissionRequested = true;
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      stream.getTracks().forEach((track) => track.stop());
    } catch (e) {
      console.warn("未获得麦克风权限，只能使用默认设备", e);
      return;
    }
    await getOutputDevices();
  };

  // 切换输出设备
  const playDeviceChange = async (deviceId: string) => {
    // 找到对应的 label 用于显示
    const option = outputDevices.value.find((d) => d.value === deviceId);
    const label = option?.label || deviceId;

    try {
      await player.toggleOutputDevice(deviceId);
      settingStore.playDevice = deviceId;
      window.$message.success(`已切换输出设备为 ${label}`);
    } catch (e) {
      window.$message.error(`切换输出设备失败: ${e}`);
    }
  };

  const onActivate = () => {
    getOutputDevices();
  };

  return {
    onActivate,
    groups: [
      {
        title: "播放控制",
        items: [
          {
            key: "autoPlay",
            label: "自动播放",
            type: "switch",
            description: "启动软件时是否自动播放",
            value: computed({
              get: () => settingStore.autoPlay,
              set: (v) => (settingStore.autoPlay = v),
            }),
          },
          {
            key: "fmGroups",
            label: "漫游 FM 分组歌单",
            type: "button",
            description:
              "不足 10 首的专辑按这些歌单分组，不在其中的不进入漫游 FM；10 首及以上的专辑不受影响",
            buttonLabel: "配置",
            action: openFmGroupManager,
          },
          {
            key: "useNextPrefetch",
            label: "下一首歌曲预载",
            type: "switch",
            description: "提前加载下一首歌曲的开头部分，提升切换速度",
            value: computed({
              get: () => settingStore.useNextPrefetch,
              set: (v) => (settingStore.useNextPrefetch = v),
            }),
          },
          {
            key: "memoryLastSeek",
            label: "记忆上次播放位置",
            type: "switch",
            description: "程序启动时恢复上次播放位置",
            value: computed({
              get: () => settingStore.memoryLastSeek,
              set: (v) => (settingStore.memoryLastSeek = v),
            }),
          },
          {
            key: "preventSleep",
            label: "阻止系统息屏",
            type: "switch",
            description: "是否在播放界面阻止系统息屏",
            value: computed({
              get: () => settingStore.preventSleep,
              set: (v) => (settingStore.preventSleep = v),
            }),
          },
          {
            key: "progressTooltipShow",
            label: "显示进度条悬浮信息",
            type: "switch",
            value: computed({
              get: () => settingStore.progressTooltipShow,
              set: (v) => (settingStore.progressTooltipShow = v),
            }),
            children: [
              {
                key: "progressLyricShow",
                label: "进度条悬浮时显示歌词",
                type: "switch",
                value: computed({
                  get: () => settingStore.progressLyricShow,
                  set: (v) => (settingStore.progressLyricShow = v),
                }),
              },
            ],
          },
          {
            key: "progressAdjustLyric",
            label: "进度调节吸附最近歌词",
            type: "switch",
            description: "进度调节时从当前时间最近一句歌词开始播放",
            value: computed({
              get: () => settingStore.progressAdjustLyric,
              set: (v) => (settingStore.progressAdjustLyric = v),
            }),
          },
          {
            key: "songVolumeFade",
            label: "音乐渐入渐出",
            type: "switch",
            value: computed({
              get: () => settingStore.songVolumeFade,
              set: (v) => (settingStore.songVolumeFade = v),
            }),
            children: [
              {
                key: "songVolumeFadeTime",
                label: "渐入渐出时长",
                type: "input-number",
                description: "单位 ms，最小 200，最大 2000",
                min: 200,
                max: 2000,
                suffix: "ms",
                value: computed({
                  get: () => settingStore.songVolumeFadeTime,
                  set: (v) => (settingStore.songVolumeFadeTime = v),
                }),
              },
            ],
          },
        ],
      },
      {
        title: "音频设置",
        items: [
          {
            key: "audioDelayCompensation",
            label: "音频与歌词同步补偿",
            type: "input-number",
            description:
              "手动补偿音频与歌词进度延迟。<br>正值歌词变快，负值歌词进度变慢。",
            tags: [{ text: "Beta", type: "warning" }],
            min: -1000,
            max: 1000,
            step: 10,
            suffix: "ms",
            value: computed({
              get: () => settingStore.audioDelayCompensation,
              set: (v) => (settingStore.audioDelayCompensation = v ?? 0),
            }),
            defaultValue: 0,
          },
          {
            key: "enableReplayGain",
            label: "音量平衡",
            type: "switch",
            description:
              "按服务端测得的响度调整音量，让不同歌曲听起来差不多响（在线歌曲不调整）",
            value: computed({
              get: () => settingStore.enableReplayGain,
              set: (v) => (settingStore.enableReplayGain = v),
            }),
          },
          {
            key: "playDevice",
            label: "音频输出设备",
            type: "select",
            description: "新增或移除音频设备后请重新打开设置",
            options: outputDevices,
            componentProps: {
              renderOption,
              "onUpdate:show": requestDevicePermission,
            },
            value: computed({
              get: () => settingStore.playDevice,
              set: (v) => playDeviceChange(v),
            }),
          },
        ],
      },
    ],
  };
};
