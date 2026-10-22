<template>
  <div id="app-layout">
    <n-layout id="main" has-sider>
      <n-layout-sider
        v-if="isDesktop"
        id="main-sider"
        :content-style="{ overflow: 'hidden', height: '100%', padding: '0' }"
        :native-scrollbar="false"
        :collapsed="appearance.menuCollapsed"
        :collapsed-width="64"
        :width="240"
        collapse-mode="width"
        show-trigger="bar"
        bordered
        @collapse="appearance.menuCollapsed = true"
        @expand="appearance.menuCollapsed = false"
      >
        <div class="sider-all">
          <div
            :class="['logo', { collapsed: appearance.menuCollapsed }]"
            @click="router.push('/')"
          >
            <Logo /><n-text>SPlayer</n-text>
          </div>
          <n-scrollbar style="max-height: calc(100vh - 70px)"
            ><SidebarMenu :collapsed="appearance.menuCollapsed"
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
  </div>
</template>

<script setup lang="ts">
import AppNav from "@/components/Layout/AppNav.vue";
import SidebarMenu from "@/components/Layout/SidebarMenu.vue";
import Logo from "@/components/Layout/Logo.vue";
import { useAppearanceStore } from "@/stores/appearance";

const router = useRouter();
const appearance = useAppearanceStore();
const isDesktop = useMediaQuery("(min-width: 1024px)");
const isMobile = useMediaQuery("(max-width: 639.98px)");
</script>

<style lang="scss" scoped src="@/style/layout.scss"></style>
<style lang="scss" scoped src="@/style/sider.scss"></style>
