<template>
  <n-config-provider
    :locale="zhCN"
    :date-locale="dateZhCN"
    :theme="appearance.dark ? darkTheme : null"
    :theme-overrides="themeOverrides"
    abstract
    inline-theme-disabled
    preflight-style-disabled
  >
    <n-global-style />
    <n-loading-bar-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <n-message-provider :max="1" placement="bottom">
            <n-modal-provider>
              <slot />
            </n-modal-provider>
          </n-message-provider>
        </n-notification-provider>
      </n-dialog-provider>
    </n-loading-bar-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { darkTheme, zhCN, dateZhCN, type GlobalThemeOverrides } from "naive-ui";
import {
  argbFromHex,
  themeFromSourceColor,
  Hct,
  redFromArgb,
  greenFromArgb,
  blueFromArgb,
} from "@material/material-color-utilities";
import { useAppearanceStore } from "@/stores/appearance";

const appearance = useAppearanceStore();

// 按 SPlayer 默认 secondary 色板和非全局着色规则生成主题。
const themeOverrides = computed<GlobalThemeOverrides>(() => {
  const palette = themeFromSourceColor(argbFromHex("#fe7971")).palettes
    .secondary;
  const color = Hct.from(
    palette.hue,
    palette.chroma,
    appearance.dark ? 90 : 10,
  ).toInt();
  const primary = `${redFromArgb(color)}, ${greenFromArgb(color)}, ${blueFromArgb(color)}`;
  const background = appearance.dark ? "16, 16, 20" : "246, 246, 246";
  const surface = appearance.dark ? "24, 24, 28" : "255, 255, 255";
  for (const [key, value] of Object.entries({
    primary,
    background,
    "surface-container": surface,
  })) {
    document.body.style.setProperty(`--${key}`, value);
    document.body.style.setProperty(`--${key}-hex`, `rgb(${value})`);
  }
  return {
    common: {
      fontFamily:
        'v-sans, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"',
      primaryColor: `rgb(${primary})`,
      primaryColorHover: `rgba(${primary}, 0.78)`,
      primaryColorPressed: `rgba(${primary}, 0.26)`,
      primaryColorSuppl: `rgba(${primary}, 0.12)`,
    },
    Icon: { color: `rgb(${primary})` },
    Tooltip: { color: `rgb(${surface})`, textColor: `rgb(${primary})` },
    Tabs: {
      colorSegment: `rgb(${surface})`,
      tabColorSegment: `rgba(${primary}, 0.12)`,
    },
  };
});
</script>
