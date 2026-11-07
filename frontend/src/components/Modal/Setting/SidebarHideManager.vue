<template>
  <div class="sidebar-hide-manager">
    <n-scrollbar style="max-height: 400px" trigger="none">
      <div class="list">
        <n-card
          v-for="item in sidebarItems"
          :key="item.key"
          :content-style="{
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
            padding: '16px',
          }"
          class="item"
        >
          <n-text class="name">{{ item.label }}</n-text>
          <n-switch
            :value="!settingStore.sidebarHide[item.key]"
            :round="false"
            @update:value="(val) => updateSetting(item.key, !val)"
          />
        </n-card>
      </div>
    </n-scrollbar>
  </div>
</template>

<script setup lang="ts">
// 照搬 SPlayer，只保留 Sonora 侧栏中可隐藏的最近播放与歌单。
import { NScrollbar } from "naive-ui";
import { useSettingStore } from "@/stores";

const settingStore = useSettingStore();

type SidebarHideKey = keyof typeof settingStore.sidebarHide;
type SidebarHideItem = { label: string; key: SidebarHideKey };

const sidebarItems: SidebarHideItem[] = [
  { label: "最近播放", key: "hideHistory" },
  { label: "歌单", key: "hideUserPlaylists" },
];

const updateSetting = (key: SidebarHideKey, val: boolean) => {
  settingStore.sidebarHide[key] = val;
};
</script>

<style scoped lang="scss">
.list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  .item {
    border-radius: 8px;
    .name {
      font-size: 16px;
      line-height: normal;
    }
    .n-switch {
      margin-left: auto;
    }
  }
}
</style>
