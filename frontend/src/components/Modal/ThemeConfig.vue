<template>
  <n-scrollbar
    style="max-height: 70vh"
    :content-style="{
      display: 'flex',
      flexDirection: 'column',
      gap: '12px',
    }"
  >
    <div class="config-section">
      <n-card class="set-item">
        <div class="label">
          <n-text class="name">全局着色</n-text>
          <n-text class="tip" :depth="3">将主题色应用至所有元素</n-text>
        </div>
        <n-switch
          v-model:value="settingStore.themeGlobalColor"
          class="set"
          :round="false"
          @update:value="themeGlobalColorChange"
        />
      </n-card>
      <n-card class="set-item">
        <div class="label">
          <n-text class="name">动态取色</n-text>
          <n-text class="tip" :depth="3">主题色跟随歌曲封面</n-text>
        </div>
        <n-switch
          v-model:value="settingStore.themeFollowCover"
          :disabled="isEmpty(statusStore.songCoverTheme)"
          class="set"
          :round="false"
        />
      </n-card>
      <n-card class="set-item">
        <div class="label">
          <n-text class="name">主题变体</n-text>
          <n-text class="tip" :depth="3"
            >调整颜色生成算法风格，请勿随意修改</n-text
          >
        </div>
        <n-select
          v-model:value="settingStore.themeVariant"
          :options="variantOptions"
          class="set"
          size="small"
          style="width: 140px"
          @update:value="themeGlobalColorChange(true)"
        />
      </n-card>
    </div>
    <!-- 颜色模式：选择主题色 -->
    <div
      class="color-section"
      :class="{ disabled: settingStore.themeFollowCover }"
    >
      <n-text class="section-title" :depth="2">选择主题色</n-text>
      <div class="color-grid">
        <div
          v-for="(colorData, key) in themeColors"
          :key="key"
          class="color-item"
          :class="{
            active:
              settingStore.themeColorType === key &&
              !settingStore.themeFollowCover,
          }"
          :style="{
            '--color':
              key === 'custom'
                ? settingStore.themeCustomColor || '#888'
                : colorData.color,
          }"
          :title="colorData.name"
          @click="
            !settingStore.themeFollowCover &&
            key !== 'custom' &&
            selectColor(key as ThemeColorType)
          "
        >
          <!-- 普通颜色 -->
          <template v-if="key !== 'custom'">
            <div class="color-circle">
              <Transition name="fade">
                <SvgIcon
                  v-if="
                    settingStore.themeColorType === key &&
                    !settingStore.themeFollowCover
                  "
                  name="Check"
                  :size="20"
                />
              </Transition>
            </div>
            <n-text class="color-name" :depth="2">{{ colorData.name }}</n-text>
          </template>
          <!-- 自定义颜色 -->
          <template v-else>
            <div class="color-circle custom-trigger">
              <SvgIcon
                v-if="
                  settingStore.themeColorType === 'custom' &&
                  !settingStore.themeFollowCover
                "
                name="Check"
                :size="16"
              />
              <n-color-picker
                v-model:value="settingStore.themeCustomColor"
                :show-alpha="false"
                :modes="['hex']"
                placement="top"
                :disabled="settingStore.themeFollowCover"
                class="color-picker-overlay"
                @update:show="(show: boolean) => show && selectColor('custom')"
              />
            </div>
            <n-text class="color-name" :depth="2">{{ colorData.name }}</n-text>
          </template>
        </div>
      </div>
    </div>
  </n-scrollbar>
</template>

<script setup lang="ts">
// 照搬 SPlayer components/Modal/ThemeConfig.vue，去掉自定义背景（图片或视频）。
import { useMusicStore, useSettingStore, useStatusStore } from "@/stores";
import { isEmpty } from "lodash-es";
import themeColor from "@/assets/data/themeColor.json";
import { getCoverColor } from "@/utils/color";
import type { ThemeColorType } from "@/types/color";

const musicStore = useMusicStore();
const settingStore = useSettingStore();
const statusStore = useStatusStore();

// 主题颜色变体选项
const variantOptions = [
  { label: "主色", value: "primary" },
  { label: "次色", value: "secondary" },
  { label: "第三色", value: "tertiary" },
  { label: "中性色", value: "neutral" },
  { label: "中性变体", value: "neutralVariant" },
  { label: "错误色", value: "error" },
];

// 主题颜色数据
const themeColors = themeColor as Record<
  string,
  { label: string; name: string; color: string }
>;

// 选择颜色
const selectColor = (key: ThemeColorType) => {
  settingStore.themeColorType = key;
};

// 全局着色更改
const themeGlobalColorChange = (val: boolean) => {
  if (val) getCoverColor(musicStore.songCover);
};
</script>

<style lang="scss" scoped>
.config-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  .set-item {
    border-radius: 8px;
    :deep(.n-card__content) {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 8px 12px;
    }
    .label {
      display: flex;
      flex-direction: column;
      .name {
        font-size: 14px;
      }
      .tip {
        font-size: 12px;
        margin-top: 2px;
      }
    }
  }
}
.color-section {
  transition: opacity 0.3s ease;
  &.disabled {
    opacity: 0.4;
    pointer-events: none;
  }
  .section-title {
    display: block;
    font-size: 13px;
    margin-bottom: 12px;
  }
  .color-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(60px, 1fr));
    gap: 12px;
    .color-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 6px;
      padding: 8px;
      border-radius: 8px;
      cursor: pointer;
      transition:
        background-color 0.3s,
        box-shadow 0.3s;
      &:hover {
        background-color: rgba(var(--primary), 0.08);
      }
      &.active {
        background-color: rgba(var(--primary), 0.12);
        .color-circle {
          box-shadow: 0 0 0 3px rgba(var(--primary), 0.3);
        }
      }
      .color-circle {
        position: relative;
        width: 36px;
        height: 36px;
        border-radius: 50%;
        background-color: var(--color);
        display: flex;
        align-items: center;
        justify-content: center;
        transition: background-color 0.3s;
        color: #fff;
        &.custom-trigger {
          cursor: pointer;
          overflow: hidden;
          .color-picker-overlay {
            position: absolute;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            opacity: 0;
            cursor: pointer;
          }
        }
      }
      .color-name {
        font-size: 12px;
        text-align: center;
      }
    }
  }
}
</style>
