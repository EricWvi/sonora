<!-- 最近播放。照搬 SPlayer views/History.vue，记录来自服务端（GET history），一次载入全部。 -->
<template>
  <div class="history">
    <div class="title">
      <n-text class="keyword">最近播放</n-text>
      <n-text class="size" depth="3"
        >共 {{ historyList.length || 0 }} 首</n-text
      >
    </div>
    <n-flex class="menu">
      <n-button
        :focusable="false"
        :disabled="!historyList.length"
        type="primary"
        strong
        secondary
        round
        v-debounce="() => player.updatePlayList(historyList)"
      >
        <template #icon>
          <SvgIcon name="Play" />
        </template>
        播放
      </n-button>
      <n-button
        :focusable="false"
        :disabled="!historyList.length"
        class="more"
        strong
        secondary
        round
        @click="cleanHistory"
      >
        <template #icon>
          <SvgIcon name="Delete" />
        </template>
        清空列表
      </n-button>
    </n-flex>
    <ListError v-if="error" :error="error" @retry="load" />
    <Transition v-else name="fade" mode="out-in">
      <SongList
        v-if="historyList.length > 0 || loading"
        :data="historyList"
        :loading="loading"
        disabled-sort
      />
      <n-empty
        v-else
        description="暂无记录，快去播放一些歌曲吧"
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
import type { Track } from "@/api/types";
import { ApiError, source } from "@/api/source";
import { fetchAll } from "@/api/pages";
import { usePlayerController } from "@/core/player/PlayerController";
import { clearHistory } from "@/utils/auth";
import SongList from "@/components/List/SongList.vue";
import ListError from "@/components/List/ListError.vue";

const player = usePlayerController();

// 最近播放列表
const historyList = shallowRef<Track[]>([]);
const loading = ref<boolean>(true);
const error = ref<string>("");

const load = async () => {
  loading.value = true;
  error.value = "";
  try {
    const entries = await fetchAll((offset, limit) =>
      source.history({ offset, limit }),
    );
    historyList.value = entries.map((entry) => entry.track);
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
};

// 清空最近播放
const cleanHistory = () => {
  window.$dialog.warning({
    title: "清空列表",
    content: "确认清空最近播放列表？该操作不可撤销！",
    positiveText: "确认",
    negativeText: "取消",
    onPositiveClick: async () => {
      if (!(await clearHistory())) return;
      historyList.value = [];
      window.$message.success("最近播放列表已清空");
    },
  });
};

// 每次进入都重新读取；keep-alive 首次激活与挂载同时发生，跳过
const isActivated = ref<boolean>(false);
onActivated(() => {
  if (isActivated.value) load();
  isActivated.value = true;
});
onMounted(load);
</script>

<style lang="scss" scoped>
.history {
  display: flex;
  flex-direction: column;
  height: 100%;
  .title {
    display: flex;
    align-items: flex-end;
    line-height: normal;
    margin-top: 12px;
    margin-bottom: 20px;
    .keyword {
      font-size: 30px;
      font-weight: bold;
      margin-right: 8px;
      line-height: normal;
    }
    .size {
      font-size: 15px;
      font-weight: normal;
      line-height: 30px;
    }
  }
  .menu {
    width: 100%;
    margin-bottom: 12px;
    .n-button {
      height: 40px;
      transition: all 0.3s var(--n-bezier);
    }
  }
  .song-list {
    flex: 1;
    overflow: hidden;
  }
}
</style>
