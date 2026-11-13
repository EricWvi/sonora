<!-- 在线歌单、专辑与艺术家的详情，共用一页：视觉与本地的详情页相同，都是 ListDetail 加 SongList，
数据来自网易云。与 SPlayer 不同：不迁入评论、收藏与分享；头部按钮与“更多”照搬 SPlayer 的歌单、专辑与艺术家页（专辑另有“导入专辑到曲库”）；艺术家页是“歌曲（热门）/ 专辑”两个标签，
没有视频标签与全部歌曲（SPlayer 的“全部歌曲”另有页面）。 -->
<template>
  <div :class="['online-list', { 'with-tabs': isArtist && !error }]">
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
      :more-options="moreOptions"
      @update:search-value="handleSearchUpdate"
      @play-all="playAllSongs"
    >
      <!-- 照搬 SPlayer：自己创建的歌单是“编辑歌单”，其他歌单是“收藏歌单”；专辑页“收藏专辑”，艺术家页“关注歌手” -->
      <template #action-buttons>
        <n-button
          v-if="ownPlaylist"
          :focusable="false"
          strong
          secondary
          round
          @click="editPlaylist"
        >
          <template #icon>
            <SvgIcon name="EditNote" />
          </template>
          编辑歌单
        </n-button>
        <n-button
          v-else
          :focusable="false"
          strong
          secondary
          round
          @click="toSubscribeOnline(listType, listIdNumber, !subscribed)"
        >
          <template #icon>
            <SvgIcon :name="subscribed ? 'Favorite' : 'FavoriteBorder'" />
          </template>
          {{ subscribeText }}
        </n-button>
      </template>
    </ListDetail>
    <ListError v-if="error" :error="error" :login="login" @retry="load" />
    <!-- 标签照搬本地艺术家页 -->
    <n-tabs
      v-if="isArtist && !error"
      :value="artistTab"
      class="tabs"
      type="segment"
      @update:value="changeArtistTab"
    >
      <n-tab name="songs"> 歌曲 </n-tab>
      <n-tab name="albums"> 专辑 </n-tab>
    </n-tabs>
    <div
      v-if="isArtist && !error && artistTab === 'albums'"
      class="albums"
      :style="{ height: `${songListHeight}px` }"
      @scroll="handleListScroll"
    >
      <ListError
        v-if="albums.error.value"
        :error="albums.error.value"
        :login="albums.login.value"
        @retry="albums.reload()"
      />
      <CoverList
        v-else
        :data="albums.items.value"
        :loading="albums.loading.value"
        :loadMore="albums.hasMore.value"
        type="album"
        empty-description="没有专辑"
        :hiddenCover="settingStore.hiddenCovers.album"
        @loadMore="albums.loadMore()"
      />
    </div>
    <Transition v-else-if="!error" name="fade" mode="out-in">
      <SongList
        v-if="!searchValue || searchData?.length"
        :data="displayData"
        :loading="loading"
        :height="songListHeight"
        :play-list-id="listId"
        :doubleClickAction="searchData?.length ? 'add' : 'all'"
        :user-playlist-id="ownPlaylist ? listIdNumber : 0"
        @scroll="handleListScroll"
        @removeSong="removeSong"
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
import { ApiError, needsOnlineLogin, source } from "@/api/source";
import { useListDetail } from "@/composables/List/useListDetail";
import { useListSearch } from "@/composables/List/useListSearch";
import { useListScroll } from "@/composables/List/useListScroll";
import { useListActions } from "@/composables/List/useListActions";
import { usePagedList } from "@/composables/List/usePagedList";
import { useSettingStore } from "@/stores";
import { useUserStore } from "@/stores/user";
import { listKey, trackKey } from "@/utils/track";
import { renderIcon } from "@/utils/helper";
import { importAlbum } from "@/utils/albumImport";
import { copyShareLink, openSourcePage } from "@/utils/onlineLink";
import { openBatchList, openUpdatePlaylist } from "@/utils/modal";
import {
  deletePlaylist,
  publishOnlinePlaylist,
  toSubscribeOnline,
} from "@/utils/auth";
import type { DropdownOption } from "naive-ui";
import CoverList from "@/components/List/CoverList.vue";
import ListDetail from "@/components/List/ListDetail.vue";
import ListError from "@/components/List/ListError.vue";
import SongList from "@/components/List/SongList.vue";

const route = useRoute();
const router = useRouter();
const settingStore = useSettingStore();
const userStore = useUserStore();
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
const login = ref<boolean>(false);
// 标签的高度算在歌曲列表之外
const ARTIST_TABS_HEIGHT = 54;
const songListHeight = computed(
  () =>
    getSongListHeight(listScrolling.value) -
    (isArtist.value ? ARTIST_TABS_HEIGHT : 0),
);
const showLoading = computed(
  () => listData.value.length === 0 && loading.value,
);
const listConfig = {
  titleType: "normal" as const,
  showArtist: false,
  showCount: true,
  searchAlign: "center" as const,
};

// 收藏与关注写入网易云；状态按收藏的歌单、专辑与关注的艺术家判断（同 SPlayer）
const listIdNumber = computed(() => Number(route.query.id));
const ownPlaylist = computed(
  () =>
    listType.value === "playlist" &&
    userStore.onlinePlaylists.created.some(
      (item) => item.id === listIdNumber.value,
    ),
);
const subscribed = computed(() => {
  const id = listIdNumber.value;
  switch (listType.value) {
    case "album":
      return userStore.onlineCollected.albums.includes(id);
    case "artist":
      return userStore.onlineCollected.artists.includes(id);
    default:
      return userStore.onlinePlaylists.collected.some((item) => item.id === id);
  }
});
const subscribeText = computed(() => {
  if (listType.value === "artist")
    return subscribed.value ? "取消关注" : "关注歌手";
  if (subscribed.value) return "取消收藏";
  return listType.value === "album" ? "收藏专辑" : "收藏歌单";
});

// 编辑与删除复用侧边栏的弹窗与确认；编辑后头部按刷新后的歌单列表更新（见下面的 watch）
const editPlaylist = () => {
  const detail = detailData.value;
  if (!detail) return;
  void openUpdatePlaylist(
    detail.id,
    { name: detail.name, description: detail.description },
    async () => {},
    true,
  );
};

// 服务端确认删除后从列表中移除，并更新头部的数量
const removeSongs = (keys: string[]) => {
  setListData(listData.value.filter((song) => !keys.includes(trackKey(song))));
  if (detailData.value)
    setDetailData({ ...detailData.value, count: listData.value.length });
};

// 歌曲右键菜单的“从歌单中删除”；这个页面的歌都是在线歌曲
const removeSong = (ids: number[]) =>
  removeSongs(ids.map((id) => trackKey({ id, source: "online" })));

const afterPublished = () => {
  if (detailData.value) setDetailData({ ...detailData.value, private: false });
};

// “更多”照搬 SPlayer：公开隐私歌单、删除歌单（自己创建的）、批量操作、复制分享链接、打开源页面。
// 不做“刷新缓存”（在线列表不缓存）。艺术家页的批量操作只在歌曲标签下。
// 专辑另有“导入专辑到曲库”，批量操作里另有“导入到曲库”（Sonora 新增，可以导入的用户才有）。
const moreOptions = computed<DropdownOption[]>(() => {
  const type = listType.value;
  const id = listIdNumber.value;
  return [
    {
      label: "公开隐私歌单",
      key: "privacy",
      show: ownPlaylist.value && !!detailData.value?.private,
      props: { onClick: () => publishOnlinePlaylist(id, afterPublished) },
      icon: renderIcon("ListLockOpen"),
    },
    {
      label: "删除歌单",
      key: "delete",
      show: ownPlaylist.value,
      props: {
        onClick: () => deletePlaylist(id, () => router.back(), true),
      },
      icon: renderIcon("Delete"),
    },
    {
      label: "批量操作",
      key: "batch",
      show: type !== "artist" || artistTab.value === "songs",
      props: {
        onClick: () =>
          openBatchList(displayData.value, {
            ...(ownPlaylist.value && {
              playlistId: id,
              onlinePlaylist: true,
              onRemoved: removeSongs,
            }),
            ...(type === "album" &&
              userStore.user?.canImport && { importAlbumId: id }),
          }),
      },
      icon: renderIcon("Batch"),
    },
    {
      label: "复制分享链接",
      key: "copy",
      props: { onClick: () => copyShareLink(type, id) },
      icon: renderIcon("Share"),
    },
    {
      label: "打开源页面",
      key: "open",
      props: { onClick: () => openSourcePage(type, id) },
      icon: renderIcon("Link"),
    },
    {
      label: "导入专辑到曲库",
      key: "import-album",
      show: type === "album" && !!userStore.user?.canImport,
      props: { onClick: () => importAlbum(id) },
      icon: renderIcon("Download"),
    },
  ];
});

const listType = computed(
  () => (route.query.type as OnlineListType | undefined) ?? "playlist",
);
const listId = computed(() =>
  listKey(listType.value, { id: Number(route.query.id), source: "online" }),
);
const isArtist = computed(() => listType.value === "artist");

// 标签放在地址里：从专辑详情返回时回到专辑标签
const artistTab = computed(() =>
  route.query.tab === "albums" ? "albums" : "songs",
);
const changeArtistTab = (tab: string) => {
  listScrolling.value = false;
  void router.replace({
    query: { ...route.query, tab: tab === "albums" ? "albums" : undefined },
  });
};

// 专辑在第一次打开标签时才读取
const albums = usePagedList(
  (offset, limit) =>
    source.onlineArtistAlbums(Number(route.query.id), { offset, limit }),
  50,
);
const loadedAlbumsOf = ref<unknown>(null);
watch(
  () => [route.name, isArtist.value, artistTab.value, route.query.id] as const,
  ([name, artist, tab, id]) => {
    if (name !== "online-list" || !artist || tab !== "albums") return;
    if (loadedAlbumsOf.value === id) return;
    loadedAlbumsOf.value = id;
    void albums.reload();
  },
  { immediate: true },
);

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
      private: playlist.private,
    });
    setListData(tracks);
  } catch (e) {
    setListData([]);
    error.value = e instanceof ApiError ? e.message : "加载失败";
    login.value = needsOnlineLogin(e);
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

// 别处（侧边栏右键菜单）编辑了这个歌单：按刷新后的歌单列表更新头部，同本地歌单页
watch(
  () =>
    isArtist.value || listType.value !== "playlist"
      ? undefined
      : userStore.onlinePlaylists.created.find(
          (item) => item.id === Number(route.query.id),
        ),
  (playlist) => {
    const detail = detailData.value;
    if (!playlist || detail?.id !== playlist.id) return;
    setDetailData({
      ...detail,
      name: playlist.name,
      description: playlist.description,
    });
  },
);
</script>

<style lang="scss" scoped>
// 列表页的头部是绝对定位的，内容靠 padding 让开（style/main.scss）。艺术家页让开的是标签。
.online-list.with-tabs {
  .tabs {
    flex-shrink: 0;
    margin-top: 240px;
    margin-bottom: 12px;
    transition: margin 0.3s;
  }
  .list-detail.small ~ .tabs {
    margin-top: 120px;
  }
  // 虚拟网格按所在滚动容器计算可视范围，所以这里要有确定的高度
  .albums {
    overflow: auto;
  }
  .list-detail ~ :deep(.song-list),
  .list-detail.small ~ :deep(.song-list),
  .list-detail ~ :deep(.n-empty),
  .list-detail.small ~ :deep(.n-empty) {
    padding-top: 0;
  }
}
</style>
