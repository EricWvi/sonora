<!-- 我喜欢的音乐。照搬 SPlayer views/List/liked.vue：头部为 ListDetail，封面为 SPlayer 的
默认封面，信息只有歌曲数（阶段 2 的“详情页头部”）；歌曲一次载入全部，默认最近喜欢的在前。
本地与在线歌曲混合（阶段 7），服务端按喜欢时间混排。不迁入歌单编辑、拖动排序、评论与分享。“更多”只有批量操作（阶段 7），其中“取消喜欢”是 Sonora 新增。 -->
<template>
  <div class="liked-list">
    <ListDetail
      v-if="!error"
      :detail-data="detailData"
      :list-data="listData"
      :loading="showLoading"
      :list-scrolling="listScrolling"
      :search-value="searchValue"
      :config="listConfig"
      :play-button-text="playButtonText"
      :more-options="moreOptions"
      title-text="我喜欢的音乐"
      @update:search-value="handleSearchUpdate"
      @play-all="playAllSongs"
    />
    <ListError v-if="error" :error="error" @retry="getLikedSongs" />
    <Transition v-else name="fade" mode="out-in">
      <SongList
        v-if="!searchValue || searchData?.length"
        :data="displayData"
        :loading="loading"
        :height="songListHeight"
        play-list-id="likes"
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
import { ApiError, source } from "@/api/source";
import { fetchAll } from "@/api/pages";
import { useStatusStore } from "@/stores";
import { sortTracks, trackKey } from "@/utils/track";
import { useListDetail } from "@/composables/List/useListDetail";
import { useListSearch } from "@/composables/List/useListSearch";
import { useListScroll } from "@/composables/List/useListScroll";
import { useListActions } from "@/composables/List/useListActions";
import type { DropdownOption } from "naive-ui";
import { renderIcon } from "@/utils/helper";
import { openBatchList } from "@/utils/modal";
import ListDetail from "@/components/List/ListDetail.vue";
import ListError from "@/components/List/ListError.vue";
import SongList from "@/components/List/SongList.vue";

const statusStore = useStatusStore();

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
const { listScrolling, handleListScroll, resetScroll } = useListScroll();
const { playAllSongs: playAllSongsAction } = useListActions();

// 是否激活
const isActivated = ref<boolean>(false);

// 加载错误
const error = ref<string>("");

// 列表高度
const songListHeight = computed(() => getSongListHeight(listScrolling.value));

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
    label: "批量操作",
    key: "batch",
    props: {
      onClick: () =>
        openBatchList(displayData.value, {
          liked: true,
          onRemoved: removeSongs,
        }),
    },
    icon: renderIcon("Batch"),
  },
]);

// 取消喜欢后从列表移除，并更新头部的数量
const removeSongs = (keys: string[]) => {
  setListData(listData.value.filter((song) => !keys.includes(trackKey(song))));
  if (detailData.value)
    setDetailData({ ...detailData.value, count: listData.value.length });
};

// 获取喜欢的歌曲（按当前排序）
const getLikedSongs = async () => {
  setLoading(true);
  error.value = "";
  try {
    let onlineFailed = false;
    const mixed = await fetchAll(async (offset, limit) => {
      const page = await source.mixedLikes({ offset, limit });
      onlineFailed ||= !!page.onlineFailed;
      return page;
    });
    if (onlineFailed)
      window.$message.warning("网易云暂时无法读取，只显示本地歌曲");
    // 在线歌曲不在数据库中，混合列表在浏览器中排序
    const songs = sortTracks(
      mixed,
      statusStore.listSortField,
      statusStore.listSortOrder,
    );
    setDetailData({
      id: 0,
      name: "我喜欢的音乐",
      cover: null,
      count: songs.length,
    });
    setListData(songs);
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : "加载失败";
  } finally {
    setLoading(false);
  }
};

// 排序更改后重新读取
watch(
  [() => statusStore.listSortField, () => statusStore.listSortOrder],
  getLikedSongs,
);

// 处理搜索更新
const handleSearchUpdate = (val: string) => {
  searchValue.value = val;
  performSearch(val);
};

// 播放全部歌曲
const playAllSongs = useDebounceFn(() => {
  if (!displayData.value?.length) return;
  playAllSongsAction(displayData.value, "likes");
}, 300);

onActivated(() => {
  if (!isActivated.value) {
    isActivated.value = true;
  } else {
    clearSearch();
    getLikedSongs();
  }
});

onDeactivated(() => {
  resetScroll();
});

onMounted(getLikedSongs);
</script>
