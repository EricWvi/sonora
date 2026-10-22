<template>
  <div class="empty-page">
    <n-h1>{{ route.meta.title }}</n-h1>
    <n-empty description="空空如也，怎么什么都没有啊" size="large" />
  </div>
</template>

<script setup lang="ts">
const route = useRoute();
</script>
