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
      <SearchInp />
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
import SearchInp from "@/components/Search/SearchInp.vue";
import SvgIcon from "@/components/Global/SvgIcon.vue";
import type { DropdownOption } from "naive-ui";
import { useSettingStore } from "@/stores/setting";
import { openSetting, openThemeConfig } from "@/utils/modal";

const router = useRouter();
const settingStore = useSettingStore();
const isDesktop = useMediaQuery("(min-width: 1024px)");
const isSmallScreen = useMediaQuery("(max-width: 767.98px)");
const showAside = ref(false);
// 设置菜单：沿用 SPlayer 网页版（明暗模式、主题配置、全局设置）。
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
  {
    label: "主题配置",
    key: "themeConfig",
    icon: () => h(SvgIcon, { name: "Palette" }),
  },
  {
    key: "divider-1",
    type: "divider",
  },
  {
    key: "setting",
    label: "全局设置",
    icon: () => h(SvgIcon, { name: "Settings" }),
  },
]);

function setSelect(key: string): void {
  switch (key) {
    case "themeMode":
      settingStore.setThemeMode();
      break;
    case "themeConfig":
      openThemeConfig();
      break;
    case "setting":
      openSetting();
      break;
  }
}
</script>

<style lang="scss" scoped src="@/style/nav.scss"></style>
