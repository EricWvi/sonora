<!-- 列表页的标题与操作栏：歌曲、专辑、艺术家、最近播放。
结构与样式照搬 SPlayer 本地音乐与流媒体页面（views/Local/layout.vue）的标题、状态与菜单。 -->
<template>
  <div class="list-page-header">
    <div class="title">
      <n-text class="keyword">{{ title }}</n-text>
      <n-flex class="status">
        <slot name="status" />
      </n-flex>
    </div>
    <n-flex class="menu" justify="space-between">
      <n-flex class="left" align="flex-end">
        <slot name="left" />
      </n-flex>
      <n-flex class="right" justify="end">
        <slot name="right" />
      </n-flex>
    </n-flex>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string }>();
</script>

<style lang="scss" scoped>
.list-page-header {
  .title {
    display: flex;
    align-items: flex-end;
    line-height: normal;
    margin-top: 12px;
    margin-bottom: 20px;
    height: 40px;
    .keyword {
      font-size: 30px;
      font-weight: bold;
      margin-right: 12px;
      line-height: normal;
    }
    .status {
      font-size: 15px;
      font-weight: normal;
      line-height: 30px;
      :deep(.item) {
        display: flex;
        align-items: center;
        opacity: 0.9;
        .n-icon {
          margin-right: 4px;
        }
      }
    }
  }
  .menu {
    width: 100%;
    margin-bottom: 20px;
    height: 40px;
    :deep(.n-button) {
      height: 40px;
      transition: all 0.3s var(--n-bezier);
    }
    :deep(.more) {
      width: 40px;
    }
    :deep(.search) {
      height: 40px;
      width: 130px;
      display: flex;
      align-items: center;
      border-radius: 25px;
      transition: all 0.3s var(--n-bezier);
      &.n-input--focus {
        width: 200px;
      }
    }
    :deep(.n-tabs) {
      width: 320px;
      --n-tab-border-radius: 25px !important;
      .n-tabs-rail {
        outline: 1px solid var(--n-tab-color-segment);
      }
    }
  }
}
</style>
