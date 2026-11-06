<!-- 每日推荐。照搬 SPlayer views/DailySongs.vue，歌曲来自网易云账号。与 SPlayer 不同：
不缓存到本地，每次进入都向接口服务读取，所以提示只写“每日 6:00 更新”，不写上次更新时间；
不迁入“不感兴趣”（会改动网易云账号的推荐）。 -->
<template>
  <div class="daily-songs">
    <div class="title">
      <n-text class="name">每日推荐</n-text>
      <div class="tip">
        <n-text depth="3">根据你的音乐口味 · 每日 6:00 更新</n-text>
      </div>
      <n-flex class="menu">
        <n-button
          :focusable="false"
          :disabled="!songs.length"
          type="primary"
          size="large"
          strong
          secondary
          round
          v-debounce="() => player.updatePlayList(songs)"
        >
          <template #icon>
            <SvgIcon name="Play" />
          </template>
          播放全部
        </n-button>
        <!-- 更多 -->
        <n-dropdown
          :options="moreOptions"
          trigger="click"
          placement="bottom-start"
        >
          <n-button
            :focusable="false"
            size="large"
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
      </n-flex>
    </div>
    <ListError v-if="error" :error="error" @retry="load()" />
    <!-- 列表 -->
    <SongList
      v-else
      :data="songs"
      :loading="loading"
      height="auto"
      disabledSort
    />
  </div>
</template>

<script setup lang="ts">
import type { DropdownOption } from "naive-ui";
import type { Track } from "@/api/types";
import { ApiError, source } from "@/api/source";
import { usePlayerController } from "@/core/player/PlayerController";
import { renderIcon } from "@/utils/helper";
import { openBatchList } from "@/utils/modal";
import ListError from "@/components/List/ListError.vue";
import SongList from "@/components/List/SongList.vue";

const player = usePlayerController();

const songs = shallowRef<Track[]>([]);
const loading = ref<boolean>(true);
const error = ref<string>("");

const load = async (refresh = false) => {
  loading.value = true;
  error.value = "";
  try {
    songs.value = (await source.onlineDailySongs()).items;
    if (refresh) window.$message.success("每日推荐更新成功");
  } catch (e) {
    songs.value = [];
    error.value = e instanceof ApiError ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
};

// 更多操作
const moreOptions = computed<DropdownOption[]>(() => [
  {
    label: "更新日推",
    key: "refresh",
    props: { onClick: () => load(true) },
    icon: renderIcon("Refresh"),
  },
  {
    label: "批量操作",
    key: "batch",
    disabled: !songs.value.length,
    props: { onClick: () => openBatchList(songs.value) },
    icon: renderIcon("Batch"),
  },
]);

// 每次进入都重新读取；keep-alive 首次激活与挂载同时发生，跳过
const isActivated = ref<boolean>(false);
onActivated(() => {
  if (isActivated.value) load();
  isActivated.value = true;
});
onMounted(() => load());
</script>

<style lang="scss" scoped>
.daily-songs {
  .title {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    height: 300px;
    margin-bottom: 20px;
    .name {
      font-size: 55px;
      font-weight: bold;
      animation: fade-spacing 1s cubic-bezier(0.4, 0, 0.2, 1) forwards;
    }
    .tip {
      font-size: 16px;
      margin-top: 6px;
      opacity: 0;
      animation: fade-down 0.5s cubic-bezier(0.4, 0, 0.2, 1) forwards;
      animation-delay: 0.5s;
    }
    .menu {
      margin-top: 30px;
    }
  }
  .song-list {
    height: auto;
  }
}
</style>
