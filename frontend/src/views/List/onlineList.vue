<!-- 在线歌单、专辑与艺术家的详情，共用一页：视觉与本地的详情页相同，都是 ListDetail 加 SongList，
数据来自网易云。与 SPlayer 不同：不迁入评论、收藏与分享；艺术家页只有热门歌曲。 -->
<template>
  <div class="online-list">
    <ListDetail
      v-if="!error"
      :detail-data="detailData"
      :list-data="listData"
      :loading="showLoading"
      :list-scrolling="listScrolling"
      :search-value="searchValue"
      :config="listConfig"
      :play-button-text="showLoading ? '加载中...' : '播放'"
      :title-text="detailData?.name ?? ''"
      @update:search-value="handleSearchUpdate"
      @play-all="playAllSongs"
    />
    <ListError v-if="error" :error="error" @retry="load" />
    <Transition v-else name="fade" mode="out-in">
      <SongList
        v-if="!searchValue || searchData?.length"
        :data="displayData"
        :loading="loading"
        :height="songListHeight"
        :play-list-id="listId"
        :doubleClickAction="searchData?.length ? 'add' : 'all'"
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
    </Transition>
  </div>
</template>

<script setup lang="ts">
import type { ListDetailData } from "@/viewmodels";
import type { OnlineListType } from "@/api/types";
import { ApiError, source } from "@/api/source";
import { useListDetail } from "@/composables/List/useListDetail";
import { useListSearch } from "@/composables/List/useListSearch";
import { useListScroll } from "@/composables/List/useListScroll";
import { useListActions } from "@/composables/List/useListActions";
import ListDetail from "@/components/List/ListDetail.vue";
import ListError from "@/components/List/ListError.vue";
import SongList from "@/components/List/SongList.vue";

const route = useRoute();
const {
  detailData,
  listData,
  loading,
  getSongListHeight,
  setDetailData,
  setListData,
  setLoading,
} = useListDetail<ListDetailData>();
const { searchValue, searchData, displayData, clearSearch, performSearch } =
  useListSearch(listData);
const { listScrolling, handleListScroll } = useListScroll();
const { playAllSongs: playAllSongsAction } = useListActions();

const error = ref<string>("");
const songListHeight = computed(() => getSongListHeight(listScrolling.value));
const showLoading = computed(
  () => listData.value.length === 0 && loading.value,
);
const listConfig = {
  titleType: "normal" as const,
  showArtist: false,
  showCount: true,
  searchAlign: "center" as const,
};

const listType = computed(
  () => (route.query.type as OnlineListType | undefined) ?? "playlist",
);
const listId = computed(() => `online-${listType.value}-${route.query.id}`);

const load = async () => {
  const id = Number(route.query.id);
  if (route.name !== "online-list" || !id) return;
  setLoading(true);
  error.value = "";
  try {
    const { playlist, tracks } = await source.onlineList(listType.value, id);
    setDetailData({
      id,
      name: playlist.name,
      cover: playlist.cover,
      description: playlist.description,
      count: tracks.length,
    });
    setListData(tracks);
  } catch (e) {
    setListData([]);
    error.value = e instanceof ApiError ? e.message : "加载失败";
  } finally {
    setLoading(false);
  }
};

const handleSearchUpdate = (val: string) => {
  searchValue.value = val;
  performSearch(val);
};

const playAllSongs = useDebounceFn(() => {
  if (!displayData.value?.length) return;
  playAllSongsAction(displayData.value, listId.value);
}, 300);

watch(
  () => [route.query.type, route.query.id],
  () => {
    clearSearch();
    void load();
  },
);
onMounted(load);
</script>
