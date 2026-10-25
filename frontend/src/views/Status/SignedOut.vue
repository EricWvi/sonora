<!-- 退出登录后的状态页，沿用 Status 页面的样式。不直接跳转登录，否则 Authelia 会立即以原会话重新登录 -->
<template>
  <n-result
    class="signed-out"
    status="info"
    title="已退出登录"
    description="本浏览器的 Sonora 会话已结束"
  >
    <template #footer>
      <n-button type="primary" @click="login">重新登录</n-button>
    </template>
  </n-result>
</template>

<script setup lang="ts">
import { loginUrl } from "@/api/source";

const login = () => location.assign(loginUrl);
</script>

<style lang="scss" scoped>
.signed-out {
  height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
</style>
