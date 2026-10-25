<template>
  <n-layout-header class="nav">
    <n-flex class="page-control">
      <Logo v-if="!isDesktop" :size="40" @click="router.push('/')" />
      <template v-if="!isSmallScreen">
        <n-button
          :focusable="false"
          tertiary
          circle
          aria-label="后退"
          @click="router.go(-1)"
        >
          <template #icon
            ><SvgIcon name="NavigateBefore" :size="26"
          /></template>
        </n-button>
        <n-button
          :focusable="false"
          tertiary
          circle
          aria-label="前进"
          @click="router.go(1)"
        >
          <template #icon><SvgIcon name="NavigateNext" :size="26" /></template>
        </n-button>
      </template>
    </n-flex>
    <n-flex :wrap="false" justify="end" class="nav-main">
      <div class="search">
        <n-input
          v-model:value="keyword"
          class="search-input"
          placeholder="搜索音乐"
          round
          clearable
          :input-props="{ 'aria-label': '搜索音乐' }"
          @keyup.enter="search"
        >
          <template #prefix><SvgIcon :size="18" name="Search" /></template>
        </n-input>
      </div>
      <div v-if="isDesktop" class="nav-drag" />
      <n-flex align="center">
        <UserMenu />
        <n-dropdown :options="setOptions" trigger="click" @select="setSelect">
          <n-button
            :focusable="false"
            title="设置"
            aria-label="设置"
            tertiary
            circle
          >
            <template #icon><SvgIcon name="Settings" /></template>
          </n-button>
        </n-dropdown>
        <n-button
          v-if="!isDesktop"
          :focusable="false"
          aria-label="打开菜单"
          tertiary
          circle
          @click="showAside = true"
        >
          <template #icon><SvgIcon name="Menu" /></template>
        </n-button>
        <n-drawer v-model:show="showAside" :width="240" placement="left">
          <n-drawer-content
            :body-content-style="{ padding: 0 }"
            :native-scrollbar="false"
          >
            <template #header
              ><n-flex align="center" justify="center" class="aside-logo"
                ><Logo /><n-text>Sonora</n-text></n-flex
              ></template
            >
            <SidebarMenu :collapsed="false" @menu-click="showAside = false" />
          </n-drawer-content>
        </n-drawer>
      </n-flex>
    </n-flex>
  </n-layout-header>
</template>

<script setup lang="ts">
import Logo from "./Logo.vue";
import SidebarMenu from "./SidebarMenu.vue";
import UserMenu from "./UserMenu.vue";
import SvgIcon from "@/components/Global/SvgIcon.vue";
import type { DropdownOption } from "naive-ui";
import { useSettingStore } from "@/stores/setting";

const router = useRouter();
const settingStore = useSettingStore();
const isDesktop = useMediaQuery("(min-width: 1024px)");
const isSmallScreen = useMediaQuery("(max-width: 767.98px)");
const keyword = ref("");
const showAside = ref(false);
// 设置菜单：沿用 SPlayer，主题配置与全局设置的弹窗在阶段 5 接入。
const setOptions = computed<DropdownOption[]>(() => [
  {
    label:
      settingStore.themeMode === "auto"
        ? "浅色模式"
        : settingStore.themeMode === "light"
          ? "深色模式"
          : "跟随系统",
    key: "themeMode",
    icon: () =>
      h(SvgIcon, {
        name:
          settingStore.themeMode === "auto"
            ? "LightTheme"
            : settingStore.themeMode === "light"
              ? "DarkTheme"
              : "AutoTheme",
      }),
  },
]);

function setSelect(key: string): void {
  if (key === "themeMode") settingStore.setThemeMode();
}

// 搜索先展示空结果，不模拟服务端数据。
function search(): void {
  if (keyword.value.trim())
    void router.push({
      name: "search",
      query: { keyword: keyword.value.trim() },
    });
}
</script>

<style lang="scss" scoped src="@/style/nav.scss"></style>
<style lang="scss" scoped src="@/style/search.scss"></style>
