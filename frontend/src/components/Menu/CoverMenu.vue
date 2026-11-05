<!-- 专辑、歌单卡片的右键菜单。照搬 SPlayer components/Menu/CoverMenu.vue：查看详情、播放或暂停、
复制名称与 ID；不迁入分享（Sonora 没有公开链接）、视频与电台。 -->
<template>
  <n-dropdown
    :x="dropdownX"
    :y="dropdownY"
    :show="dropdownShow"
    :options="dropdownOptions"
    class="search-inp-menu"
    placement="bottom-start"
    trigger="manual"
    @select="dropdownShow = false"
    @clickoutside="dropdownShow = false"
  />
</template>

<script setup lang="ts">
import type { DropdownOption } from "naive-ui";
import { listKey, listRoute } from "@/utils/track";
import type { Album, Playlist } from "@/api/types";
import { renderIcon, copyData } from "@/utils/helper";
import { useMusicStore, useStatusStore } from "@/stores";

const emit = defineEmits<{
  // 播放或暂停
  toPlay: [item: Album | Playlist];
}>();

const router = useRouter();
const musicStore = useMusicStore();
const statusStore = useStatusStore();

// 右键菜单数据
const dropdownX = ref<number>(0);
const dropdownY = ref<number>(0);
const dropdownShow = ref<boolean>(false);
const dropdownOptions = ref<DropdownOption[]>([]);

// 开启右键菜单
const openDropdown = (
  e: MouseEvent,
  item: Album | Playlist,
  type: "playlist" | "album",
) => {
  try {
    e.preventDefault();
    dropdownShow.value = false;
    const typeName = type === "playlist" ? "歌单" : "专辑";
    const name = "title" in item ? item.title : item.name;
    // 播放列表来源（见 music store 的 playPlaylistId）
    const playing =
      musicStore.playPlaylistId === listKey(type, item) &&
      statusStore.playStatus;
    nextTick().then(() => {
      dropdownOptions.value = [
        {
          key: "open",
          label: "查看详情",
          props: {
            // 在线专辑、歌单的详情页是 online-list，ID 可能与本地的相同
            onClick: () => router.push(listRoute(type, item)),
          },
          icon: renderIcon("Eye"),
        },
        {
          key: "play",
          label: "播放",
          show: !playing,
          props: { onClick: () => emit("toPlay", item) },
          icon: renderIcon("Play"),
        },
        {
          key: "pause",
          label: "暂停",
          show: playing,
          props: { onClick: () => emit("toPlay", item) },
          icon: renderIcon("Pause"),
        },
        { key: "line", type: "divider" },
        {
          key: "code-name",
          label: `复制${typeName}名称`,
          props: { onClick: () => copyData(name) },
          icon: renderIcon("Copy", { size: 18 }),
        },
        {
          key: "code-id",
          label: `复制${typeName} ID`,
          props: { onClick: () => copyData(item.id) },
          icon: renderIcon("Copy", { size: 18 }),
        },
      ];
      // 显示菜单
      dropdownX.value = e.clientX;
      dropdownY.value = e.clientY;
      dropdownShow.value = true;
    });
  } catch (error) {
    console.error("右键菜单出现异常：", error);
    window.$message.error("右键菜单出现异常");
  }
};

defineExpose({ openDropdown });
</script>
