<!-- 列表加载失败。样式沿用 SPlayer 流媒体页面的空状态（n-empty 加主色按钮）。 -->
<template>
  <n-empty :description="error" size="large" class="list-error">
    <template #extra>
      <n-button type="primary" strong secondary @click="emit('retry')">
        <template #icon>
          <SvgIcon name="Refresh" />
        </template>
        重试
      </n-button>
    </template>
  </n-empty>
</template>

<script setup lang="ts">
defineProps<{ error: string }>();
const emit = defineEmits<{ retry: [] }>();
</script>

<style lang="scss" scoped>
.list-error {
  margin-top: 60px;
  // 详情页给空状态留出头部的位置（main.scss 的列表页通用样式）；出错时不显示头部
  padding-top: 0 !important;
}
</style>
