<script setup lang="ts">
import { useRouter } from "vue-router";
import {
  NButton,
  NConfigProvider,
  NDialogProvider,
  NElement,
  NGlobalStyle,
  NMessageProvider,
  NResult,
  NSpin,
  darkTheme,
  dateZhCN,
  useOsTheme,
  zhCN,
} from "naive-ui";
import { handleUnauthenticated } from "./api";
import { boot, ensureStatus, session, signedOut } from "./session";

const router = useRouter();
const osTheme = useOsTheme();

async function retry() {
  await ensureStatus();
  // Re-run the guard now that the server state is known.
  if (session.loaded) await router.replace(router.currentRoute.value.fullPath);
}

handleUnauthenticated(() => {
  if (!session.admin) return;
  signedOut();
  const here = router.currentRoute.value.fullPath;
  router.replace({ path: "/login", query: { next: here, expired: "1" } });
});
</script>

<template>
  <n-config-provider
    :theme="osTheme === 'dark' ? darkTheme : null"
    :locale="zhCN"
    :date-locale="dateZhCN"
  >
    <n-global-style />
    <n-element class="root">
      <n-message-provider>
        <n-dialog-provider>
          <div v-if="!session.loaded && boot.error" class="center">
            <n-result
              status="error"
              title="无法连接 Sonora 服务"
              :description="boot.error.message"
            >
              <template #footer>
                <n-button @click="retry">重试</n-button>
              </template>
            </n-result>
          </div>
          <div v-else-if="!session.loaded" class="center">
            <n-spin />
          </div>
          <router-view v-else />
        </n-dialog-provider>
      </n-message-provider>
    </n-element>
  </n-config-provider>
</template>

<style>
body {
  margin: 0;
}
.root a {
  color: var(--primary-color);
  text-decoration: none;
  cursor: pointer;
}
.root a:hover {
  color: var(--primary-color-hover);
}
.center {
  display: grid;
  place-items: center;
  min-height: 100vh;
}
</style>
