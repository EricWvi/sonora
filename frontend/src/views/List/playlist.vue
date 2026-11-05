<!-- 歌单详情。照搬 SPlayer views/List/playlist.vue（用户自己的歌单）：头部为 ListDetail，
歌曲一次载入全部；“编辑歌单”与“更多 → 删除歌单”沿用 SPlayer，弹窗与写入在阶段 5 接入。
默认排序且未搜索时可拖动排序。不迁入收藏、评论、隐私、缓存与分享。“更多 → 批量操作”在阶段 7 加入。 -->
<template>
  <div class="playlist-list">
    <ListDetail
      v-if="!error"
      :detail-data="detailData?.id === playlistId ? detailData : null"
      :list-data="detailData?.id === playlistId ? listData : []"
      :loading="showLoading"
      :list-scrolling="listScrolling"
      :search-value="searchValue"
      :config="listConfig"
      :play-button-text="playButtonText"
      :more-options="moreOptions"
      @update:search-value="handleSearchUpdate"
      @play-all="playAllSongs"
    >
      <template #action-buttons>
        <n-button
          :focusable="false"
          strong
          secondary
          round
          @click="updatePlaylist"
        >
          <template #icon>
            <SvgIcon name="EditNote" />
          </template>
          编辑歌单
        </n-button>
      </template>
    </ListDetail>
    <ListError
      v-if="error"
      :error="error"
      :missing="missing"
      @retry="getPlaylistDetail(playlistId)"
    />
    <!-- 歌曲列表 -->
    <SongList
      v-else-if="!searchValue || searchData?.length"
      :data="detailData?.id === playlistId ? displayData : []"
      :loading="loading"
      :height="songListHeight"
      :play-list-id="`playlist-${playlistId}`"
      :user-playlist-id="playlistId"
      :draggable="canDragSort"
      :doubleClickAction="searchData?.length ? 'add' : 'all'"
      @scroll="handleListScroll"
      @removeSong="removeSong"
      @reorder="handleReorder"
    />
    <n-empty
      v-else
      :description="`搜不到关于 ${searchValue} 的任何歌曲呀`"
      style="margin-top: 60px"
      size="large"
    >
      <template #icon>
        <SvgIcon name="SearchOff" />
      </template>
    </n-empty>
  </div>
</template>

<script setup lang="ts">
import type { DropdownOption } from "naive-ui";
import { source } from "@/api/source";
import { useLoadError } from "@/composables/List/useLoadError";
import { fetchAll } from "@/api/pages";
import { useStatusStore, useUserStore } from "@/stores";
import { renderIcon } from "@/utils/helper";
import { listSortQuery } from "@/utils/meta";
import { openBatchList, openUpdatePlaylist } from "@/utils/modal";
import { deletePlaylist, savePlaylistOrder } from "@/utils/auth";
import { useListDetail } from "@/composables/List/useListDetail";
import { useListSearch } from "@/composables/List/useListSearch";
import { useListScroll } from "@/composables/List/useListScroll";
import { useListActions } from "@/composables/List/useListActions";
import type { ListDetailData } from "@/viewmodels";
import ListDetail from "@/components/List/ListDetail.vue";
import ListError from "@/components/List/ListError.vue";
import SongList from "@/components/List/SongList.vue";

const router = useRouter();
const statusStore = useStatusStore();
const userStore = useUserStore();

// 是否激活
const isActivated = ref<boolean>(false);

const {
  detailData,
  listData,
  loading,
  getSongListHeight,
  resetData,
  setDetailData,
  setListData,
  setLoading,
} = useListDetail<ListDetailData>();
const { searchValue, searchData, displayData, clearSearch, performSearch } =
  useListSearch(listData);
const { listScrolling, handleListScroll, resetScroll } = useListScroll();
const { playAllSongs: playAllSongsAction } = useListActions();

// 加载错误
const { error, missing, setError, clearError } =
  useLoadError("歌单不存在或已被删除");

// 歌单 ID
const playlistId = computed<number>(() =>
  Number(router.currentRoute.value.query.id as string),
);

// 当前正在请求的歌单 ID，用于防止竞态条件
const currentRequestId = ref<number>(0);

// 列表高度
const songListHeight = computed(() => getSongListHeight(listScrolling.value));

// 是否可拖拽排序
const canDragSort = computed(() => {
  return !searchValue.value && statusStore.listSortField === "default";
});

// 列表配置
const listConfig = {
  titleType: "normal" as const,
  showArtist: false,
  showCount: true,
  searchAlign: "center" as const,
};

// 是否显示加载状态
const showLoading = computed(
  () => listData.value.length === 0 && loading.value,
);

// 播放按钮文本
const playButtonText = computed(() => {
  if (showLoading.value) {
    return "加载中...";
  }
  return "播放";
});

// 更多操作
const moreOptions = computed<DropdownOption[]>(() => [
  {
    label: "删除歌单",
    key: "delete",
    props: {
      onClick: toDeletePlaylist,
    },
    icon: renderIcon("Delete"),
  },
  {
    label: "批量操作",
    key: "batch",
    props: {
      onClick: () =>
        openBatchList(displayData.value, {
          playlistId: playlistId.value,
          onRemoved: removeSong,
        }),
    },
    icon: renderIcon("Batch"),
  },
]);

// 获取歌单歌曲（按当前排序）
const getPlaylistSongs = (id: number) =>
  fetchAll((offset, limit) =>
    source.playlistTracks(id, {
      offset,
      limit,
      ...listSortQuery(statusStore.listSortField, statusStore.listSortOrder),
    }),
  );

// 获取歌单详情与歌曲
const getPlaylistDetail = async (id: number) => {
  if (!id) return;
  // 设置当前请求的歌单 ID，用于防止竞态条件
  currentRequestId.value = id;
  setLoading(true);
  clearSearch();
  clearError();
  if (detailData.value?.id !== id) resetData(true);
  try {
    const [playlist, songs] = await Promise.all([
      source.playlist(id),
      getPlaylistSongs(id),
    ]);
    // 检查是否仍然是当前请求的歌单
    if (currentRequestId.value !== id) return;
    setDetailData({
      id: playlist.id,
      name: playlist.name,
      cover: playlist.cover,
      description: playlist.description,
      count: playlist.trackCount,
      updateTime: playlist.updatedAt,
      createTime: playlist.createdAt,
    });
    setListData(songs);
  } catch (e) {
    if (currentRequestId.value !== id) return;
    setError(e);
  } finally {
    if (currentRequestId.value === id) setLoading(false);
  }
};

// 排序更改后重新读取歌曲
watch(
  [() => statusStore.listSortField, () => statusStore.listSortOrder],
  async () => {
    const id = playlistId.value;
    if (!id || detailData.value?.id !== id) return;
    setLoading(true);
    try {
      const songs = await getPlaylistSongs(id);
      if (playlistId.value === id) setListData(songs);
    } finally {
      setLoading(false);
    }
  },
);

// 处理搜索更新
const handleSearchUpdate = (val: string) => {
  searchValue.value = val;
  performSearch(val);
};

// 播放全部歌曲
const playAllSongs = useDebounceFn(() => {
  if (!detailData.value || !displayData.value?.length) return;
  playAllSongsAction(displayData.value, `playlist-${playlistId.value}`);
}, 300);

// 编辑歌单，成功后重新读取
const updatePlaylist = () => {
  if (!detailData.value || !playlistId.value) return;
  const id = playlistId.value;
  openUpdatePlaylist(
    id,
    {
      name: detailData.value.name,
      description: detailData.value.description,
    },
    () => getPlaylistDetail(id),
  );
};

// 删除歌单，成功后返回上一页（与 SPlayer 的本地歌单相同）
const toDeletePlaylist = () => {
  if (!detailData.value || !playlistId.value) return;
  deletePlaylist(playlistId.value, () => router.back());
};

// 删除歌曲：服务端确认后从列表中移除，并更新头部的数量
const removeSong = (keys: (string | number)[]) => {
  const ids = keys.map(Number);
  setListData(listData.value.filter((song) => !ids.includes(song.id)));
  if (detailData.value)
    setDetailData({ ...detailData.value, count: listData.value.length });
};

// 拖拽排序：先更新视图，保存失败时重新读取歌单
const handleReorder = async (fromIndex: number, toIndex: number) => {
  if (fromIndex === toIndex) return;
  const id = playlistId.value;
  const newList = [...listData.value];
  const [moved] = newList.splice(fromIndex, 1);
  newList.splice(toIndex, 0, moved);
  setListData(newList);
  const saved = await savePlaylistOrder(
    id,
    newList.map((song) => song.id),
  );
  if (!saved && playlistId.value === id) getPlaylistDetail(id);
};

// 别处（如侧边栏右键菜单）编辑了这个歌单：按刷新后的歌单列表更新头部，不重新读取歌曲
watch(
  () => userStore.playlists.find((item) => item.id === playlistId.value),
  (playlist) => {
    const detail = detailData.value;
    if (!playlist || detail?.id !== playlist.id) return;
    if (
      detail.name === playlist.name &&
      detail.description === playlist.description
    )
      return;
    setDetailData({
      ...detail,
      name: playlist.name,
      description: playlist.description,
      updateTime: playlist.updatedAt,
    });
  },
);

onBeforeRouteUpdate((to) => {
  if (to.name !== "playlist") return;
  clearSearch();
  const id = Number(to.query.id as string);
  if (id) getPlaylistDetail(id);
});

onActivated(() => {
  if (!isActivated.value) {
    isActivated.value = true;
  } else {
    getPlaylistDetail(playlistId.value);
  }
});

onDeactivated(() => {
  resetScroll();
});

onMounted(() => {
  getPlaylistDetail(playlistId.value);
});
</script>
