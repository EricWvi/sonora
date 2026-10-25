<!-- 顶栏用户入口：按 SPlayer User.vue 改编，只保留用户名与退出登录 -->
<template>
  <n-popover
    :show="menuShow"
    style="padding: 12px; max-width: 240px"
    trigger="manual"
    @clickoutside="menuShow = false"
  >
    <template #trigger>
      <div
        class="user"
        :style="{ pointerEvents: menuShow ? 'none' : 'auto' }"
        @click="menuShow = !!userStore.user && !menuShow"
      >
        <div class="avatar">
          <n-avatar round>
            <SvgIcon name="Person" :depth="3" :size="26" />
          </n-avatar>
        </div>
        <n-flex v-if="isDesktop" :wrap="false" class="user-data" size="small">
          <n-text class="name text-hidden">
            {{ userStore.user?.name ?? "" }}
          </n-text>
          <SvgIcon
            :class="['down', { open: menuShow }]"
            name="DropDown"
            :depth="3"
          />
        </n-flex>
      </div>
    </template>
    <div class="user-menu" @click="menuShow = false">
      <n-flex class="user-info" align="center" justify="center" vertical>
        <n-text class="nickname text-hidden">{{ userStore.user?.name }}</n-text>
        <n-text
          v-if="userStore.user?.email"
          class="email text-hidden"
          :depth="3"
        >
          {{ userStore.user.email }}
        </n-text>
      </n-flex>
      <n-divider />
      <n-button
        :focusable="false"
        class="logout"
        strong
        secondary
        round
        @click="confirmLogout"
      >
        <template #icon>
          <SvgIcon name="Power" />
        </template>
        退出登录
      </n-button>
    </div>
  </n-popover>
</template>

<script setup lang="ts">
import { useDialog, useMessage } from "naive-ui";
import SvgIcon from "@/components/Global/SvgIcon.vue";
import { useUserStore } from "@/stores/user";

const userStore = useUserStore();
const dialog = useDialog();
const message = useMessage();
const isDesktop = useMediaQuery("(min-width: 1024px)");
const menuShow = ref(false);

function confirmLogout(): void {
  dialog.warning({
    title: "退出登录",
    content: "确认退出当前用户登录？",
    positiveText: "确认登出",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await userStore.logout();
      } catch (error) {
        message.error(error instanceof Error ? error.message : "退出登录失败");
      }
    },
  });
}
</script>

<style lang="scss" scoped src="@/style/user.scss"></style>
