<!-- 搜索框的默认内容。照搬 SPlayer components/Search/SearchDefault.vue，只保留搜索历史，
不迁入热搜榜（阶段 2 的“设置 → 常规”）。 -->
<template>
  <Transition name="fadeDown" mode="out-in">
    <n-card v-if="isShow" class="search-default" content-style="padding: 0">
      <n-scrollbar class="scrollbar">
        <!-- 搜索历史 -->
        <div v-if="isShowSearchHistory" class="history">
          <div class="title">
            <SvgIcon name="History" />
            <n-text class="name">搜索历史 </n-text>
            <SvgIcon
              class="delete"
              name="Delete"
              @click.stop="deleteSearchHistory"
            />
          </div>
          <n-flex class="history-list">
            <n-tag
              v-for="(item, index) in dataStore.searchHistory"
              :key="index"
              :bordered="false"
              round
              @click="emit('toSearch', item)"
            >
              {{ item.length > 10 ? item.slice(0, 10) + "..." : item }}
            </n-tag>
          </n-flex>
        </div>
      </n-scrollbar>
    </n-card>
  </Transition>
</template>

<script setup lang="ts">
import { useSettingStore, useStatusStore, useDataStore } from "@/stores";

const emit = defineEmits<{
  toSearch: [keyword: string];
}>();

const dataStore = useDataStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();

// 是否展示 SearchDefault
const isShow = computed(() => {
  return (
    !statusStore.searchInputValue &&
    statusStore.searchFocus &&
    isShowSearchHistory.value
  );
});

// 是否展示搜索历史
const isShowSearchHistory = computed(() => {
  return settingStore.showSearchHistory && dataStore.searchHistory.length > 0;
});

// 删除搜索历史
const deleteSearchHistory = () => {
  window.$dialog.warning({
    title: "删除搜索历史",
    content: "确认删除全部的搜索历史？这将无法恢复！",
    positiveText: "确认",
    negativeText: "取消",
    onPositiveClick: () => {
      dataStore.searchHistory = [];
    },
  });
};
</script>

<style lang="scss" scoped>
.search-default {
  position: absolute;
  left: 0;
  top: 50px;
  width: 300px;
  border-radius: 8px;
  z-index: 101;
  &.fadeDown-enter-to {
    transition-delay: 0.25s;
  }
  :deep(.scrollbar) {
    max-height: calc(100vh - 160px);
    .n-scrollbar-content {
      padding: 10px;
    }
  }
  @media (max-width: 768px) {
    width: 100%;
  }
  .title {
    display: flex;
    flex-direction: row;
    align-items: center;
    margin-bottom: 12px;
    .n-icon {
      font-size: 18px;
      margin-right: 4px;
    }
  }
  .history {
    margin-bottom: 20px;
    .delete {
      margin-left: auto;
      opacity: 0.6;
      transition: opacity 0.3s;
      cursor: pointer;
      &:hover {
        opacity: 1;
      }
    }
  }
  .hot-list {
    .hot-item {
      display: flex;
      flex-direction: row;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 8px;
      border-radius: 8px;
      padding: 6px;
      transition: background-color 0.3s;
      cursor: pointer;
      .num {
        width: 30px;
        height: 30px;
        min-width: 30px;
        font-weight: bold;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 16px;
        margin-right: 8px;
      }
      .data {
        flex: 1;
        width: 100%;
        padding-right: 8px;
        .name {
          font-size: 16px;
          display: flex;
          flex-direction: row;
          align-items: center;
          .text {
            margin-right: 8px;
          }
        }
      }
      .n-tag {
        pointer-events: none;
      }
      .hot {
        display: flex;
        flex-direction: row;
        align-items: center;
        font-size: 12px;
        opacity: 0.6;
        .n-icon {
          font-size: 14px;
        }
      }
      &:last-child {
        margin-bottom: 0;
      }
      &:hover {
        background-color: rgba(var(--primary), 0.12);
      }
    }
  }
}
</style>
