<!-- Cookie 登录：照搬 SPlayer LoginCookie.vue 的布局。与 SPlayer 不同：Cookie 交给 Go 服务校验后
保存在服务端，无效的 Cookie 当场提示；不做桌面端的“自动获取”。 -->
<template>
  <div class="login-cookie">
    <n-alert :bordered="false" title="如何获取 Cookie">
      <template #icon>
        <SvgIcon name="Help" />
      </template>
      在已登录的官方网页端，从浏览器开发者工具复制请求头中的
      Cookie，粘贴到下面。Cookie 只保存在 Sonora 服务端。
    </n-alert>
    <n-input
      v-model:value="cookie"
      :autosize="{ minRows: 3, maxRows: 6 }"
      type="textarea"
      placeholder="请输入 Cookie"
    />
    <n-flex class="menu">
      <n-button type="primary" :loading="loading" @click="login">登录</n-button>
    </n-flex>
  </div>
</template>

<script setup lang="ts">
import SvgIcon from "@/components/Global/SvgIcon.vue";
import { ApiError, source } from "@/api/source";
import { useUserStore } from "@/stores/user";

const emit = defineEmits<{ close: []; success: [] }>();

const userStore = useUserStore();
const cookie = ref("");
const loading = ref(false);

async function login(): Promise<void> {
  if (!cookie.value.trim()) {
    window.$message.warning("请输入 Cookie");
    return;
  }
  loading.value = true;
  try {
    userStore.netease = await source.neteaseCookieLogin(cookie.value);
    window.$message.success("登录成功");
    emit("success");
  } catch (error) {
    window.$message.error(
      error instanceof ApiError ? error.message : "登录失败，请重试",
    );
  } finally {
    loading.value = false;
  }
}
</script>

<style lang="scss" scoped>
.login-cookie {
  .n-input,
  .n-button {
    width: 100%;
    margin-top: 20px;
  }
  .menu {
    margin-top: 20px;
    .n-button {
      width: auto;
      flex: 1;
      margin: 0;
    }
  }
}
</style>
