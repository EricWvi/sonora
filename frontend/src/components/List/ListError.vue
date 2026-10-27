<!-- 列表加载失败。样式沿用 SPlayer 流媒体页面的空状态（n-empty 加主色按钮）。
内容已被删除（missing）时重试没有意义，按钮换成 404 页的“返回上一级”。 -->
<template>
  <n-empty :description="error" size="large" class="list-error">
    <template #extra>
      <n-button
        v-if="missing"
        type="primary"
        strong
        secondary
        @click="router.go(-1)"
      >
        返回上一级
      </n-button>
      <n-button v-else type="primary" strong secondary @click="emit('retry')">
        <template #icon>
          <SvgIcon name="Refresh" />
        </template>
        重试
      </n-button>
    </template>
  </n-empty>
</template>

<script setup lang="ts">
defineProps<{ error: string; missing?: boolean }>();
const emit = defineEmits<{ retry: [] }>();
const router = useRouter();
</script>

<style lang="scss" scoped>
.list-error {
  margin-top: 60px;
  // 详情页给空状态留出头部的位置（main.scss 的列表页通用样式）；出错时不显示头部
  padding-top: 0 !important;
}
</style>
