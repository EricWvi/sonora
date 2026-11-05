<!-- 顶栏用户入口：按 SPlayer User.vue 改编。与 SPlayer 不同：网易云之外还有 OIDC 账号，
两者分别退出；按钮优先显示网易云的头像、昵称与 VIP 标志；不放歌单、专辑、歌手的统计，
不做切换账号。 -->
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
          <n-avatar
            v-if="profile"
            :src="onlineImage(profile.avatarUrl, 100)"
            fallback-src="/images/avatar.jpg"
            round
          />
          <n-avatar v-else round>
            <SvgIcon name="Person" :depth="3" :size="26" />
          </n-avatar>
        </div>
        <n-flex v-if="isDesktop" :wrap="false" class="user-data" size="small">
          <n-text class="name text-hidden">
            {{ profile?.nickname ?? userStore.user?.name ?? "" }}
          </n-text>
          <img v-if="profile?.vipType" class="vip-img" src="/images/vip.png" />
          <SvgIcon
            :class="['down', { open: menuShow }]"
            name="DropDown"
            :depth="3"
          />
        </n-flex>
      </div>
    </template>
    <div class="user-menu" @click="menuShow = false">
      <template v-if="online.state !== 'disabled'">
        <n-flex class="user-info" align="center" justify="center" vertical>
          <n-flex
            v-if="profile"
            :wrap="false"
            align="center"
            justify="center"
            size="small"
            class="online-name"
          >
            <n-text class="nickname text-hidden">{{ profile.nickname }}</n-text>
            <img v-if="profile.vipType" class="vip-img" src="/images/vip.png" />
          </n-flex>
          <n-text v-else class="online-state" :depth="3">
            {{ onlineStateText }}
          </n-text>
        </n-flex>
        <n-button
          v-if="profile"
          :focusable="false"
          class="logout online-action"
          strong
          secondary
          round
          @click="confirmOnlineLogout"
        >
          <template #icon>
            <SvgIcon name="Power" />
          </template>
          退出网易云
        </n-button>
        <n-button
          v-else-if="online.state !== 'unavailable'"
          :focusable="false"
          class="logout online-action"
          strong
          secondary
          round
          @click="openOnlineLogin"
        >
          登录网易云
        </n-button>
        <n-divider />
      </template>
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
import { openOnlineLogin } from "@/utils/modal";
import { onlineImage } from "@/utils/onlineImage";

const userStore = useUserStore();
const dialog = useDialog();
const message = useMessage();
const isDesktop = useMediaQuery("(min-width: 1024px)");
const menuShow = ref(false);

const online = computed(() => userStore.netease);
const profile = computed(() => online.value.profile);
const onlineStateText = computed(() => {
  switch (online.value.state) {
    case "expired":
      return "网易云登录已失效";
    case "unavailable":
      return "网易云服务暂时不可用";
    default:
      return "未登录网易云";
  }
});

function confirmOnlineLogout(): void {
  dialog.warning({
    title: "退出网易云",
    content: "确认退出网易云账号？Sonora 的登录不受影响。",
    positiveText: "确认退出",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await userStore.logoutNetease();
      } catch (error) {
        message.error(error instanceof Error ? error.message : "退出失败");
      }
    },
  });
}

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
