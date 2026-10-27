<template>
  <div class="cover-manager">
    <n-scrollbar style="max-height: 400px">
      <div class="list">
        <n-card
          v-for="item in coverItems"
          :key="item.key"
          :content-style="{
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
            padding: '16px',
          }"
          class="item"
        >
          <n-text class="name">{{ item.name }}</n-text>
          <n-switch
            :value="!settingStore.hiddenCovers[item.key]"
            :round="false"
            @update:value="
              (val) => (settingStore.hiddenCovers[item.key] = !val)
            "
          />
        </n-card>
      </div>
    </n-scrollbar>
    <div class="footer">
      <n-button @click="toggleAll">
        {{ isAllHidden ? "显示全部" : "隐藏全部" }}
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
// 照搬 SPlayer，条目换成 Sonora 中有封面的界面。
import { computed } from "vue";
import { useSettingStore } from "@/stores";
import type { SettingState } from "@/stores/setting";
import { NScrollbar, NButton } from "naive-ui";

const settingStore = useSettingStore();

type CoverItem = {
  key: keyof SettingState["hiddenCovers"];
  name: string;
};

const coverItems: CoverItem[] = [
  { key: "home", name: "首页" },
  { key: "playlist", name: "歌单" },
  { key: "artist", name: "艺术家" },
  { key: "player", name: "播放器" },
  { key: "list", name: "歌单详情/歌曲列表" },
  { key: "artistDetail", name: "艺术家详情" },
  { key: "album", name: "专辑" },
];

const isAllHidden = computed(() => {
  return coverItems.every((item) => settingStore.hiddenCovers[item.key]);
});

const toggleAll = () => {
  const target = !isAllHidden.value;
  coverItems.forEach((item) => {
    settingStore.hiddenCovers[item.key] = target;
  });
};
</script>

<style scoped lang="scss">
.cover-manager {
  display: flex;
  flex-direction: column;
  .list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding-right: 12px;
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
  .footer {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
}
</style>
