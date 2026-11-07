<!-- 主菜单：按 SPlayer Menu.vue 改编，条目见阶段 2 的信息架构。
与 SPlayer 不同（阶段 7，Eric 决定 2026-11-05）：歌单分组不能折叠，始终展开，标题是普通的小标题，
文字样式不变。在线/本地歌单的切换按钮照搬 SPlayer（NPopselect），默认本地歌单；
没有登录网易云时“在线歌单”不可选。与 SPlayer 不同（Eric，2026-11-07）：在线模式的标题是“在线歌单”，不是“创建的歌单”，切换时文字长度不变。“+”在在线模式下新建网易云歌单。
歌单项的右键菜单是 Sonora 新增，SPlayer 的侧边栏没有；在线歌单项的条目与本地相同。 -->
<template>
  <n-menu
    ref="menuRef"
    :value="activeKey"
    :expanded-keys="expandedKeys"
    :class="{ cover: settingStore.menuShowCover, fixed: !collapsed }"
    :indent="0"
    :root-indent="26"
    :collapsed="collapsed"
    :collapsed-width="64"
    :collapsed-icon-size="22"
    :options="options"
    :render-label="renderLabel"
    :node-props="nodeProps"
    @update:value="select"
  />
  <!-- 歌单项右键菜单（Sonora 新增，SPlayer 的侧边栏没有右键菜单） -->
  <n-dropdown
    :x="playlistMenu.x"
    :y="playlistMenu.y"
    :show="playlistMenu.show"
    :options="playlistMenuOptions"
    placement="bottom-start"
    trigger="manual"
    @select="playlistMenu.show = false"
    @clickoutside="playlistMenu.show = false"
  />
</template>

<script setup lang="ts">
import type { DropdownOption } from "naive-ui";
import type { Playlist } from "@/api/types";
import {
  NAvatar,
  NButton,
  NEllipsis,
  NPopselect,
  NText,
  type MenuInst,
  type MenuOption,
  type MenuProps,
} from "naive-ui";
import { RouterLink } from "vue-router";
import { source } from "@/api/source";
import SvgIcon from "@/components/Global/SvgIcon.vue";
import { useSettingStore } from "@/stores/setting";
import { useUserStore } from "@/stores/user";
import { openCreatePlaylist, openUpdatePlaylist } from "@/utils/modal";
import { copyData, renderIcon as renderMenuIcon } from "@/utils/helper";
import { deletePlaylist } from "@/utils/auth";
import { fetchAll } from "@/api/pages";
import { isOnline, listKey } from "@/utils/track";
import { useMusicStore, useStatusStore } from "@/stores";
import { usePlayerController } from "@/core/player/PlayerController";

const props = defineProps<{ collapsed: boolean }>();
const emit = defineEmits<{ "menu-click": [key: string | number] }>();

const route = useRoute();
const router = useRouter();
const settingStore = useSettingStore();
const userStore = useUserStore();
const statusStore = useStatusStore();
// 选择了在线歌单、后来退出网易云时，显示本地歌单，选择保留到下次登录
const onlineMode = computed(
  () => statusStore.playlistMode === "online" && userStore.onlineSignedIn,
);
const menuRef = ref<MenuInst | null>(null);

const renderIcon = (name: string, style?: Record<string, string>) => () =>
  h(SvgIcon, { name, style });

// 本地歌单以数字 ID 为键，在线歌单带前缀（ID 可能相同）
const playlistKey = (playlist: Playlist): string | number =>
  isOnline(playlist) ? `online-${playlist.id}` : playlist.id;

// 前三项与 SPlayer 一样上移 1px，对齐文字基线。
const lift = { transform: "translateY(-1px)" };

// 歌单分组始终展开（不能折叠）
const expandedKeys = ["user-playlists"];

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
    key: "collection",
    link: "collection",
    label: "我的收藏",
    // 网易云独有，没有登录时隐藏
    show: userStore.onlineSignedIn,
    icon: renderIcon("Star"),
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
        h(NText, { depth: 3 }, () =>
          onlineMode.value ? "在线歌单" : "本地歌单",
        ),
        h(
          NPopselect,
          {
            options: [
              // 没有登录网易云时不可选
              {
                label: "在线歌单",
                value: "online",
                disabled: !userStore.onlineSignedIn,
              },
              { label: "本地歌单", value: "local" },
            ],
            value: onlineMode.value ? "online" : "local",
            trigger: "click",
            onUpdateValue: (value: "online" | "local") => {
              statusStore.playlistMode = value;
            },
          },
          () =>
            h(NButton, {
              type: "tertiary",
              round: true,
              strong: true,
              secondary: true,
              renderIcon: renderIcon("Menu"),
              "aria-label": "切换歌单",
              onClick: (event: Event) => event.stopPropagation(),
            }),
        ),
        h(NButton, {
          type: "tertiary",
          round: true,
          strong: true,
          secondary: true,
          renderIcon: renderIcon("Add"),
          "aria-label": "新建歌单",
          onClick: (event: Event) => {
            event.stopPropagation();
            openCreatePlaylist(onlineMode.value);
          },
        }),
      ]),
    children: (onlineMode.value
      ? userStore.onlinePlaylists.created
      : userStore.playlists
    ).map((playlist) => ({
      key: playlistKey(playlist),
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

// 歌单项右键菜单：条目与专辑、歌单卡片的 CoverMenu 一致（查看详情、播放或暂停、复制名称与 ID），
// 另加编辑歌单、删除歌单，复用歌单详情页的弹窗与确认（Eric 确认，2026-11-04）
const musicStore = useMusicStore();
const player = usePlayerController();
const playlistMenu = reactive({ x: 0, y: 0, show: false });
const playlistMenuOptions = ref<DropdownOption[]>([]);

const playPlaylist = async (playlist: Playlist) => {
  const pid = listKey("playlist", playlist);
  if (musicStore.playPlaylistId === pid) return player.playOrPause();
  try {
    const list = isOnline(playlist)
      ? (await source.onlineList("playlist", playlist.id)).tracks
      : await fetchAll((offset, limit) =>
          source.playlistTracks(playlist.id, { offset, limit }),
        );
    if (!list.length) return window.$message.warning("歌单中没有歌曲");
    await player.updatePlayList(list, undefined, pid);
  } catch (error) {
    console.error("播放歌单失败：", error);
    window.$message.error("播放歌单失败，请重试");
  }
};

// 正在查看这个歌单
const viewing = (playlist: Playlist) =>
  isOnline(playlist)
    ? route.name === "online-list" &&
      route.query.type === "playlist" &&
      Number(route.query.id) === playlist.id
    : route.name === "playlist" && Number(route.query.id) === playlist.id;

const openPlaylistMenu = (e: MouseEvent, playlist: Playlist) => {
  e.preventDefault();
  playlistMenu.show = false;
  const online = isOnline(playlist);
  const playing =
    musicStore.playPlaylistId === listKey("playlist", playlist) &&
    statusStore.playStatus;
  nextTick().then(() => {
    playlistMenuOptions.value = [
      {
        key: "open",
        label: "查看详情",
        props: { onClick: () => select(playlistKey(playlist)) },
        icon: renderMenuIcon("Eye"),
      },
      {
        key: playing ? "pause" : "play",
        label: playing ? "暂停" : "播放",
        props: { onClick: () => playPlaylist(playlist) },
        icon: renderMenuIcon(playing ? "Pause" : "Play"),
      },
      { key: "line", type: "divider" },
      {
        key: "edit",
        label: "编辑歌单",
        props: {
          onClick: () =>
            openUpdatePlaylist(
              playlist.id,
              { name: playlist.name, description: playlist.description },
              async () => {},
              online,
            ),
        },
        icon: renderMenuIcon("EditNote"),
      },
      {
        key: "delete",
        label: "删除歌单",
        props: {
          onClick: () =>
            deletePlaylist(
              playlist.id,
              () => {
                // 正在查看这个歌单时回到首页
                if (viewing(playlist)) void router.push({ name: "home" });
              },
              online,
            ),
        },
        icon: renderMenuIcon("Delete"),
      },
      { key: "line-two", type: "divider" },
      {
        key: "code-name",
        label: "复制歌单名称",
        props: { onClick: () => copyData(playlist.name) },
        icon: renderMenuIcon("Copy", { size: 18 }),
      },
      {
        key: "code-id",
        label: "复制歌单 ID",
        props: { onClick: () => copyData(playlist.id) },
        icon: renderMenuIcon("Copy", { size: 18 }),
      },
    ];
    playlistMenu.x = e.clientX;
    playlistMenu.y = e.clientY;
    playlistMenu.show = true;
  });
};

// 菜单项的 props 只作用于分隔线，事件要经 node-props 绑定到歌单项
const nodeProps = (option: MenuOption) => {
  const playlist = [
    ...userStore.playlists,
    ...userStore.onlinePlaylists.created,
  ].find((item) => playlistKey(item) === option.key);
  // MenuNodeProps 的类型只允许字符串与数字属性，事件处理函数在运行时照常绑定
  return (
    playlist
      ? { onContextmenu: (e: MouseEvent) => openPlaylistMenu(e, playlist) }
      : {}
  ) as ReturnType<NonNullable<MenuProps["nodeProps"]>>;
};

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
  else if (key.startsWith("online-"))
    void router.push({
      name: "online-list",
      query: { type: "playlist", id: key.slice("online-".length) },
    });
}

// 当前路由对应的菜单项：详情页高亮所属的列表，自己的歌单高亮该歌单。
const activeKey = computed<string | number | null>(() => {
  const name = String(route.matched[0]?.name ?? route.name ?? "");
  switch (name) {
    case "online-list":
      return route.query.type === "playlist"
        ? `online-${route.query.id}`
        : null;
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
