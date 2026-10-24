<script setup lang="ts">
import { computed } from "vue";
import { RouterView, useRoute, useRouter } from "vue-router";
import {
  NButton,
  NLayout,
  NLayoutContent,
  NLayoutHeader,
  NLayoutSider,
  NMenu,
  NText,
  useMessage,
  type MenuOption,
} from "naive-ui";
import { request, type ApiError } from "../api";
import { session, signedOut } from "../session";

const route = useRoute();
const router = useRouter();
const message = useMessage();

const menu: MenuOption[] = [
  { label: "概况", key: "/" },
  { label: "上传", key: "/upload" },
  { label: "歌曲", key: "/tracks" },
  { label: "专辑", key: "/albums" },
  { label: "艺术家", key: "/artists" },
];
const active = computed(() => route.path);
const title = computed(() => route.meta.title as string | undefined);

async function logout() {
  try {
    await request("POST", "/logout");
  } catch (err) {
    message.error((err as ApiError).message);
    return;
  }
  signedOut();
  router.replace("/login");
}
</script>

<template>
  <n-layout has-sider style="height: 100vh">
    <n-layout-sider bordered :width="200">
      <div class="brand">Sonora</div>
      <n-menu
        :value="active"
        :options="menu"
        @update:value="(key: string) => router.push(key)"
      />
    </n-layout-sider>
    <n-layout>
      <n-layout-header bordered class="header">
        <n-text strong style="font-size: 16px">{{ title }}</n-text>
        <div class="user">
          <n-text depth="3">{{ session.admin?.username }}</n-text>
          <n-button size="small" quaternary @click="logout">退出</n-button>
        </div>
      </n-layout-header>
      <n-layout-content
        content-style="padding: 24px"
        :native-scrollbar="false"
        style="height: calc(100vh - 57px)"
      >
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<style scoped>
.brand {
  font-size: 18px;
  font-weight: 600;
  padding: 18px 24px 12px;
}
.header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}
.user {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
