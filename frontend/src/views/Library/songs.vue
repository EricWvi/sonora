<!-- 全部歌曲。标题与操作栏沿用 SPlayer 本地音乐页面；列表分页载入，排序与搜索由服务端完成。
“播放”只放入已经载入的部分（阶段 2 的“播放行为”）。
“更多 → 批量操作”（阶段 7，Eric 确认 2026-11-04）：点开时按当前搜索与排序取全部歌曲，
再打开与详情页相同的批量弹窗；SPlayer 的本地音乐页批量操作只处理已载入的列表。 -->
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
        <!-- 更多操作 -->
        <n-dropdown
          :options="moreOptions"
          trigger="click"
          placement="bottom-start"
        >
          <n-button
            :focusable="false"
            :disabled="!list.items.value.length"
            :loading="batchLoading"
            class="more"
            circle
            strong
            secondary
          >
            <template #icon>
              <SvgIcon name="List" />
            </template>
          </n-button>
        </n-dropdown>
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
import type { DropdownOption } from "naive-ui";
import { source } from "@/api/source";
import { fetchAll } from "@/api/pages";
import { renderIcon } from "@/utils/helper";
import { openBatchList } from "@/utils/modal";
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

// 批量操作：取全部歌曲（与列表相同的搜索与排序）后打开弹窗
const batchLoading = ref<boolean>(false);
const openBatch = async () => {
  if (batchLoading.value) return;
  batchLoading.value = true;
  try {
    const q = searchKeyword.value;
    const songs = await fetchAll((offset, limit) =>
      source.tracks({
        q,
        offset,
        limit,
        ...listSortQuery(statusStore.listSortField, statusStore.listSortOrder),
      }),
    );
    await openBatchList(songs);
  } catch (error) {
    console.error("获取全部歌曲失败：", error);
    window.$message.error("获取歌曲失败，请重试");
  } finally {
    batchLoading.value = false;
  }
};

// 更多操作
const moreOptions = computed<DropdownOption[]>(() => [
  {
    label: "批量操作",
    key: "batch",
    props: { onClick: openBatch },
    icon: renderIcon("Batch"),
  },
]);

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
