<!-- 全部歌曲。标题与操作栏沿用 SPlayer 本地音乐页面；列表分页载入，排序与搜索由服务端完成。
“播放”只放入已经载入的部分（阶段 2 的“播放行为”）。 -->
<template>
  <div class="library-songs">
    <ListPageHeader title="歌曲">
      <template #status>
        <n-text class="item">
          <SvgIcon name="Music" :depth="3" />
          <!-- 与 SPlayer 不同：总数直接显示，不用 n-number-animation 从 0 滚动 -->
          {{ list.total.value }} 首歌曲
        </n-text>
      </template>
      <template #left>
        <n-button
          :focusable="false"
          :disabled="!list.items.value.length"
          type="primary"
          strong
          secondary
          round
          v-debounce="playAll"
        >
          <template #icon>
            <SvgIcon name="Play" />
          </template>
          播放
        </n-button>
        <n-button
          :disabled="list.loading.value"
          :loading="list.loading.value"
          :focusable="false"
          class="more"
          strong
          secondary
          circle
          @click="list.reload()"
        >
          <template #icon>
            <SvgIcon name="Refresh" />
          </template>
        </n-button>
      </template>
      <template #right>
        <n-input
          v-model:value="searchValue"
          :input-props="{ autocomplete: 'off' }"
          class="search"
          placeholder="模糊搜索"
          clearable
          round
        >
          <template #prefix>
            <SvgIcon name="Search" />
          </template>
        </n-input>
      </template>
    </ListPageHeader>
    <ListError
      v-if="list.error.value"
      :error="list.error.value"
      @retry="list.reload()"
    />
    <n-empty
      v-else-if="searchValue && list.loaded.value && !list.items.value.length"
      :description="`搜不到关于 ${searchValue} 的任何歌曲呀`"
      style="margin-top: 60px"
      size="large"
    >
      <template #icon>
        <SvgIcon name="SearchOff" />
      </template>
    </n-empty>
    <SongList
      v-else
      :data="list.items.value"
      :loading="list.loading.value"
      :load-more="list.hasMore.value"
      :list-version="searchKeyword"
      :double-click-action="searchKeyword ? 'add' : 'all'"
      class="song-list"
      @reach-bottom="list.loadMore()"
    />
  </div>
</template>

<script setup lang="ts">
import { source } from "@/api/source";
import { useStatusStore } from "@/stores";
import { listSortQuery } from "@/utils/meta";
import { usePlayerController } from "@/core/player/PlayerController";
import { usePagedList } from "@/composables/List/usePagedList";
import ListPageHeader from "@/components/List/ListPageHeader.vue";
import ListError from "@/components/List/ListError.vue";
import SongList from "@/components/List/SongList.vue";

const statusStore = useStatusStore();
const player = usePlayerController();

// 搜索：输入停顿后按服务端的 q 重新载入
const searchValue = ref<string>("");
const searchKeyword = refDebounced(
  computed(() => searchValue.value.trim()),
  300,
);

const list = usePagedList((offset, limit) =>
  source.tracks({
    q: searchKeyword.value,
    offset,
    limit,
    ...listSortQuery(statusStore.listSortField, statusStore.listSortOrder),
  }),
);

watch(
  [
    searchKeyword,
    () => statusStore.listSortField,
    () => statusStore.listSortOrder,
  ],
  () => list.reload(),
  { immediate: true },
);

// 播放已经载入的歌曲
const playAll = () => {
  if (!list.items.value.length) return;
  player.updatePlayList(list.items.value);
};
</script>

<style lang="scss" scoped>
.library-songs {
  display: flex;
  flex-direction: column;
  height: 100%;
  .song-list {
    flex: 1;
    overflow: hidden;
  }
}
</style>
