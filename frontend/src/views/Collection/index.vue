<!-- 我的收藏：网易云账号收藏的歌单、专辑与歌手。对应 SPlayer 的 views/Like，不含视频与播客；
每种内容的卡片与本地列表相同。 -->
<template>
  <div class="collection">
    <div class="title">
      <n-text class="keyword">我的收藏</n-text>
    </div>
    <ListError v-if="error" :error="error" @retry="load" />
    <n-tabs v-else v-model:value="tab" type="segment" animated>
      <n-tab-pane name="playlists" tab="歌单">
        <CoverList
          :data="playlists"
          type="playlist"
          :loading="loading"
          empty-description="还没有收藏歌单"
        />
      </n-tab-pane>
      <n-tab-pane name="albums" tab="专辑">
        <CoverList
          :data="albums"
          type="album"
          :loading="loading"
          empty-description="还没有收藏专辑"
        />
      </n-tab-pane>
      <n-tab-pane name="artists" tab="歌手">
        <ArtistList :data="artists" :loading="loading" />
        <n-empty
          v-if="!loading && !artists.length"
          description="还没有收藏歌手"
          style="margin-top: 60px"
          size="large"
        />
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script setup lang="ts">
import type { Album, Artist, Playlist } from "@/api/types";
import { ApiError, source } from "@/api/source";
import { useUserStore } from "@/stores/user";
import CoverList from "@/components/List/CoverList.vue";
import ArtistList from "@/components/List/ArtistList.vue";
import ListError from "@/components/List/ListError.vue";

const userStore = useUserStore();
const tab = ref("playlists");
const playlists = computed<Playlist[]>(
  () => userStore.onlinePlaylists.collected,
);
const albums = shallowRef<Album[]>([]);
const artists = shallowRef<Artist[]>([]);
const loading = ref(true);
const error = ref("");

const load = async () => {
  loading.value = true;
  error.value = "";
  try {
    const [albumPage, artistPage] = await Promise.all([
      source.onlineAlbums(),
      source.onlineArtists(),
    ]);
    albums.value = albumPage.items;
    artists.value = artistPage.items;
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
};

onMounted(load);
</script>

<style lang="scss" scoped>
.collection {
  .title {
    display: flex;
    align-items: flex-end;
    margin: 20px 0 16px;
    .keyword {
      font-size: 30px;
      font-weight: bold;
    }
  }
}
</style>
