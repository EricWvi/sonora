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
          ref="contentRef"
          id="main-content"
          :native-scrollbar="false"
          :style="{
            '--layout-height': contentHeight,
          }"
          :content-style="{
            display: 'grid',
            gridTemplateRows: '1fr',
            minHeight: '100%',
            padding: isMobile ? '0 16px' : '0 24px',
          }"
          position="absolute"
          embedded
        >
          <!-- 路由页面 -->
          <RouterView v-slot="{ Component }">
            <Transition
              :name="`router-${settingStore.routeAnimation}`"
              mode="out-in"
            >
              <KeepAlive
                v-if="settingStore.useKeepAlive"
                :max="20"
                :exclude="['layout']"
              >
                <component :is="Component" class="router-view" />
              </KeepAlive>
              <component v-else :is="Component" class="router-view" />
            </Transition>
          </RouterView>
          <!-- 回顶 -->
          <n-back-top :right="40" :bottom="120">
            <SvgIcon :size="22" name="Up" />
          </n-back-top>
        </n-layout>
      </n-layout>
    </n-layout>
    <!-- 播放列表 -->
    <SongPlayList />
    <!-- 全局播放器 -->
    <MainPlayer :style="followScroll" />
    <!-- 全屏播放器 -->
    <PlayerProvider>
      <FullPlayer />
    </PlayerProvider>
  </div>
</template>

<script setup lang="ts">
import AppNav from "@/components/Layout/AppNav.vue";
import SidebarMenu from "@/components/Layout/SidebarMenu.vue";
import Logo from "@/components/Layout/Logo.vue";
import MainPlayer from "@/components/Player/MainPlayer.vue";
import SongPlayList from "@/components/List/SongPlayList.vue";
import FullPlayer from "@/components/Player/FullPlayer.vue";
import PlayerProvider from "@/components/Global/PlayerProvider.vue";
import { initPlayer } from "@/core/player/PlayerController";
import { initSettingsSync } from "@/core/settings/sync";
import { useShortcut } from "@/composables/useShortcut";
import {
  useMusicStore,
  useSettingStore,
  useStatusStore,
  useUserStore,
} from "@/stores";

const router = useRouter();
const musicStore = useMusicStore();
const statusStore = useStatusStore();
const settingStore = useSettingStore();

// 主内容
const contentRef = ref<HTMLElement | null>(null);

// 主内容高度
const { height: contentHeight } = useElementSize(contentRef);

watchEffect(() => {
  statusStore.mainContentHeight = contentHeight.value;
});
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
// 设置与服务端同步（阶段 5）
initSettingsSync();
// 页面内快捷键
useShortcut();
</script>

<style lang="scss" scoped src="@/style/layout.scss"></style>
<style lang="scss" scoped src="@/style/sider.scss"></style>
