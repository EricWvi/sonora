<!-- 专辑详情。照搬 SPlayer views/List/album.vue：头部为 ListDetail，歌曲一次载入全部。
不迁入收藏专辑、评论、缓存、批量操作与分享；“更多”只有添加到歌单（阶段 2 的“详情页头部”）。 -->
<template>
  <div class="album-list">
    <ListDetail
      v-if="!error"
      :detail-data="detailData?.id === albumId ? detailData : null"
      :list-data="detailData?.id === albumId ? listData : []"
      :loading="showLoading"
      :list-scrolling="listScrolling"
      :search-value="searchValue"
      :config="listConfig"
      :play-button-text="playButtonText"
      :more-options="moreOptions"
      @update:search-value="handleSearchUpdate"
      @play-all="playAllSongs"
    />
    <ListError
      v-if="error"
      :error="error"
      :missing="missing"
      @retry="getAlbumDetail(albumId)"
    />
    <!-- 歌曲列表 -->
    <SongList
      v-else-if="!searchValue || searchData?.length"
      :data="detailData?.id === albumId ? displayData : []"
      :loading="loading"
      :height="songListHeight"
      :play-list-id="`album-${albumId}`"
      :doubleClickAction="searchData?.length ? 'add' : 'all'"
      hidden-album
      group-by-disc
      @scroll="handleListScroll"
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
import { useStatusStore } from "@/stores";
import { renderIcon } from "@/utils/helper";
import { listSortQuery } from "@/utils/meta";
import { openPlaylistAdd } from "@/utils/modal";
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
  useLoadError("专辑不存在或已被删除");

// 专辑 ID
const albumId = computed<number>(() =>
  Number(router.currentRoute.value.query.id as string),
);

// 当前正在请求的专辑 ID，用于防止竞态条件
const currentRequestId = ref<number>(0);

// 列表高度
const songListHeight = computed(() => getSongListHeight(listScrolling.value));

// 列表配置
const listConfig = {
  titleType: "ellipsis" as const,
  showArtist: true,
  showCount: true,
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
    label: "添加到歌单",
    key: "playlist-add",
    props: {
      onClick: () => openPlaylistAdd(displayData.value),
    },
    icon: renderIcon("AddList"),
  },
]);

// 获取专辑歌曲（按当前排序）
const getAlbumSongs = (id: number) =>
  fetchAll((offset, limit) =>
    source.tracks({
      albumId: id,
      offset,
      limit,
      ...listSortQuery(statusStore.listSortField, statusStore.listSortOrder),
    }),
  );

// 获取专辑详情与歌曲
const getAlbumDetail = async (id: number) => {
  if (!id) return;
  // 设置当前请求的专辑 ID，用于防止竞态条件
  currentRequestId.value = id;
  setLoading(true);
  clearSearch();
  clearError();
  if (detailData.value?.id !== id) resetData(true);
  try {
    const [album, songs] = await Promise.all([
      source.album(id),
      getAlbumSongs(id),
    ]);
    // 检查是否仍然是当前请求的专辑
    if (currentRequestId.value !== id) return;
    setDetailData({
      id: album.id,
      name: album.title,
      cover: album.cover,
      description: album.description,
      artists: album.artists,
      count: album.trackCount,
      year: album.year,
      durationMs: album.durationMs,
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
    const id = albumId.value;
    if (!id || detailData.value?.id !== id) return;
    setLoading(true);
    try {
      const songs = await getAlbumSongs(id);
      if (albumId.value === id) setListData(songs);
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
  playAllSongsAction(displayData.value, `album-${albumId.value}`);
}, 300);

onBeforeRouteUpdate((to) => {
  if (to.name !== "album") return;
  clearSearch();
  const id = Number(to.query.id as string);
  if (id) getAlbumDetail(id);
});

onActivated(() => {
  if (!isActivated.value) {
    isActivated.value = true;
  } else {
    getAlbumDetail(albumId.value);
  }
});

onDeactivated(() => {
  resetScroll();
});

onMounted(() => {
  getAlbumDetail(albumId.value);
});
</script>
