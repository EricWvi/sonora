<!-- 主菜单：按 SPlayer Menu.vue 改编，条目见阶段 2 的信息架构 -->
<template>
  <n-menu
    ref="menuRef"
    :value="activeKey"
    v-model:expanded-keys="settingStore.menuExpandedKeys"
    :class="{ cover: settingStore.menuShowCover }"
    :indent="0"
    :root-indent="26"
    :collapsed="collapsed"
    :collapsed-width="64"
    :collapsed-icon-size="22"
    :options="options"
    :render-label="renderLabel"
    @update:value="select"
  />
</template>

<script setup lang="ts">
import {
  NAvatar,
  NButton,
  NEllipsis,
  NText,
  type MenuInst,
  type MenuOption,
} from "naive-ui";
import { RouterLink } from "vue-router";
import { source } from "@/api/source";
import SvgIcon from "@/components/Global/SvgIcon.vue";
import { useSettingStore } from "@/stores/setting";
import { useUserStore } from "@/stores/user";
import { openCreatePlaylist } from "@/utils/modal";

const props = defineProps<{ collapsed: boolean }>();
const emit = defineEmits<{ "menu-click": [key: string | number] }>();

const route = useRoute();
const router = useRouter();
const settingStore = useSettingStore();
const userStore = useUserStore();
const menuRef = ref<MenuInst | null>(null);

const renderIcon = (name: string, style?: Record<string, string>) => () =>
  h(SvgIcon, { name, style });

// 前三项与 SPlayer 一样上移 1px，对齐文字基线。
const lift = { transform: "translateY(-1px)" };

const options = computed<MenuOption[]>(() => [
  { key: "home", link: "home", label: "首页", icon: renderIcon("Home", lift) },
  {
    key: "songs",
    link: "songs",
    label: "歌曲",
    icon: renderIcon("Music", lift),
  },
  {
    key: "albums",
    link: "albums",
    label: "专辑",
    icon: renderIcon("Album", lift),
  },
  {
    key: "artists",
    link: "artists",
    label: "艺术家",
    icon: renderIcon("Artist"),
  },
  { key: "divider", type: "divider" },
  {
    key: "like-songs",
    link: "like-songs",
    label: "我喜欢的音乐",
    icon: renderIcon("Favorite"),
  },
  {
    key: "history",
    link: "history",
    label: "最近播放",
    show: !settingStore.sidebarHide.hideHistory,
    icon: renderIcon("History"),
  },
  { key: "divider-two", type: "divider" },
  {
    key: "user-playlists",
    show: !settingStore.sidebarHide.hideUserPlaylists,
    icon: props.collapsed ? renderIcon("PlaylistAdd") : undefined,
    label: () =>
      h("div", { class: "user-list" }, [
        h(NText, { depth: 3 }, () => "创建的歌单"),
        h(NButton, {
          type: "tertiary",
          round: true,
          strong: true,
          secondary: true,
          renderIcon: renderIcon("Add"),
          "aria-label": "新建歌单",
          onClick: (event: Event) => {
            event.stopPropagation();
            openCreatePlaylist();
          },
        }),
      ]),
    children: userStore.playlists.map((playlist) => ({
      key: playlist.id,
      label: () =>
        settingStore.menuShowCover
          ? h("div", { class: "pl-cover" }, [
              h(NAvatar, {
                src: playlist.cover
                  ? source.coverUrl(playlist.cover, 96)
                  : "/images/album.jpg",
                fallbackSrc: "/images/album.jpg",
                lazy: true,
              }),
              h(NEllipsis, null, () => playlist.name),
            ])
          : h(NEllipsis, null, () => playlist.name),
      icon: settingStore.menuShowCover ? undefined : renderIcon("PlayList"),
    })),
  },
]);

const renderLabel = (option: MenuOption) => {
  if ("link" in option)
    return h(
      RouterLink,
      { to: { name: option.link as string } },
      () => option.label as string,
    );
  return typeof option.label === "function"
    ? option.label()
    : (option.label as string);
};

// 歌单以数字 ID 为键；其他菜单项由 RouterLink 导航。
function select(key: string | number): void {
  emit("menu-click", key);
  if (typeof key === "number")
    void router.push({ name: "playlist", query: { id: key } });
}

// 当前路由对应的菜单项：详情页高亮所属的列表，自己的歌单高亮该歌单。
const activeKey = computed<string | number | null>(() => {
  const name = String(route.matched[0]?.name ?? route.name ?? "");
  switch (name) {
    case "album":
      return "albums";
    case "artist":
      return "artists";
    case "playlist": {
      const id = Number(route.query.id);
      return userStore.playlists.some((playlist) => playlist.id === id)
        ? id
        : null;
    }
    default:
      return name;
  }
});

// 歌单可能晚于菜单加载，等菜单渲染出该项后再展开所在分组。
watch(
  activeKey,
  (key) => {
    if (key !== null) menuRef.value?.showOption(key);
  },
  { flush: "post" },
);
</script>

<style lang="scss" scoped src="@/style/menu.scss"></style>
<style lang="scss" src="@/style/menu-labels.scss"></style>
