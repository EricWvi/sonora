<template>
  <div id="app-layout">
    <n-layout
      id="main"
      :class="{
        'show-player': musicStore.isHasPlayer && statusStore.showPlayBar,
        'show-full-player': statusStore.showFullPlayer,
      }"
      has-sider
    >
      <n-layout-sider
        v-if="isDesktop"
        id="main-sider"
        :style="{
          height:
            musicStore.isHasPlayer && statusStore.showPlayBar
              ? 'calc(100vh - 80px)'
              : '100vh',
        }"
        :content-style="{ overflow: 'hidden', height: '100%', padding: '0' }"
        :native-scrollbar="false"
        :collapsed="statusStore.menuCollapsed"
        :collapsed-width="64"
        :width="240"
        collapse-mode="width"
        show-trigger="bar"
        bordered
        @collapse="statusStore.menuCollapsed = true"
        @expand="statusStore.menuCollapsed = false"
      >
        <div class="sider-all">
          <div
            :class="['logo', { collapsed: statusStore.menuCollapsed }]"
            @click="router.push('/')"
          >
            <Logo /><n-text>Sonora</n-text>
          </div>
          <n-scrollbar style="max-height: calc(100vh - 70px)"
            ><SidebarMenu :collapsed="statusStore.menuCollapsed"
          /></n-scrollbar>
        </div>
      </n-layout-sider>
      <n-layout id="main-layout">
        <AppNav id="main-header" />
        <n-layout
          id="main-content"
          :native-scrollbar="false"
          :content-style="{
            display: 'grid',
            gridTemplateRows: '1fr',
            minHeight: '100%',
            padding: isMobile ? '0 16px' : '0 24px',
          }"
          position="absolute"
          embedded
        >
          <router-view class="router-view" />
        </n-layout>
      </n-layout>
    </n-layout>
    <!-- 播放列表 -->
    <SongPlayList />
    <!-- 全局播放器 -->
    <MainPlayer :style="followScroll" />
  </div>
</template>

<script setup lang="ts">
import AppNav from "@/components/Layout/AppNav.vue";
import SidebarMenu from "@/components/Layout/SidebarMenu.vue";
import Logo from "@/components/Layout/Logo.vue";
import MainPlayer from "@/components/Player/MainPlayer.vue";
import SongPlayList from "@/components/List/SongPlayList.vue";
import { initPlayer } from "@/core/player/PlayerController";
import { useMusicStore, useStatusStore, useUserStore } from "@/stores";

const router = useRouter();
const musicStore = useMusicStore();
const statusStore = useStatusStore();
const isDesktop = useMediaQuery("(min-width: 1024px)");

// 窗口窄于 990px 时 #app 横向滚动，固定定位的播放器随之平移；不滚动时不设 transform。
const { x: scrollX } = useScroll(() =>
  document.querySelector<HTMLElement>("#app"),
);
const followScroll = computed(() =>
  scrollX.value > 0
    ? { transform: `translateX(${-scrollX.value}px)` }
    : undefined,
);
const isMobile = useMediaQuery("(max-width: 639.98px)");

// 服务不可达时保留空的用户与歌单，页面各自显示错误状态（阶段 3.4）。
void useUserStore()
  .load()
  .catch((error: unknown) => console.error("加载用户数据失败", error));
void initPlayer().catch((error: unknown) =>
  console.error("载入播放器失败", error),
);
</script>

<style lang="scss" scoped src="@/style/layout.scss"></style>
<style lang="scss" scoped src="@/style/sider.scss"></style>
