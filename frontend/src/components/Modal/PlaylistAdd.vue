<!-- 添加到歌单 -->
<template>
  <div class="playlist-add">
    <n-scrollbar style="max-height: 70vh">
      <n-list class="playlists-list" hoverable clickable>
        <!-- 新建歌单 -->
        <n-list-item class="playlist add" @click="openCreatePlaylist">
          <template #prefix>
            <SvgIcon name="Add" :size="20" />
          </template>
          <n-thing title="创建新歌单" />
        </n-list-item>
        <!-- 已有歌单 -->
        <n-list-item
          v-for="item in userStore.playlists"
          :key="item.id"
          class="playlist"
          @click="addToPlaylist(item.id)"
        >
          <template #prefix>
            <n-image
              :src="
                item.cover
                  ? source.coverUrl(item.cover, 100)
                  : '/images/album.jpg?asset'
              "
              class="cover"
              preview-disabled
              lazy
              @load="coverLoaded"
            >
              <template #placeholder>
                <div class="cover-loading">
                  <img
                    class="loading-img"
                    src="/images/album.jpg?asset"
                    alt="loading-img"
                  />
                </div>
              </template>
            </n-image>
          </template>
          <n-thing :title="item.name">
            <template #description>
              <n-text depth="3" class="size"
                >{{ item.trackCount }} 首音乐</n-text
              >
            </template>
          </n-thing>
        </n-list-item>
      </n-list>
    </n-scrollbar>
  </div>
</template>

<script setup lang="ts">
// 照搬 SPlayer 的 PlaylistAdd.vue。列表是自己的歌单，不含“我喜欢的音乐”
// （Sonora 的喜爱不是歌单，与 SPlayer 的本地歌单相同）。
import type { Track } from "@/api/types";
import type { MessageReactive } from "naive-ui";
import { ApiError, loginUrl, source } from "@/api/source";
import { useUserStore } from "@/stores";
import { coverLoaded } from "@/utils/helper";
import { debounce } from "lodash-es";
import { openCreatePlaylist } from "@/utils/modal";

const props = defineProps<{
  data: Track[];
}>();

const emit = defineEmits<{
  close: [];
}>();

const userStore = useUserStore();

// 加载提示
const loadingMsg = ref<MessageReactive>();

// 添加到歌单
const addToPlaylist = debounce(
  async (id: number) => {
    loadingMsg.value = window.$message.loading("正在添加歌曲至歌单", {
      duration: 0,
    });
    let added: number;
    try {
      ({ added } = await source.addToPlaylist(
        id,
        props.data.map((item) => item.id),
      ));
    } catch (error) {
      loadingMsg.value?.destroy();
      if (error instanceof ApiError && error.status === 401) {
        location.assign(loginUrl);
        return;
      }
      window.$message.error(
        error instanceof ApiError ? error.message : "添加失败，请重试",
      );
      return;
    }
    loadingMsg.value?.destroy();
    emit("close");
    if (added > 0) window.$message.success("添加歌曲至歌单成功");
    else window.$message.info("所选歌曲已在歌单中");
    await userStore.loadPlaylists();
  },
  500,
  { leading: true, trailing: false },
);
</script>

<style lang="scss" scoped>
.playlists-list {
  .playlist {
    border-radius: 8px;
    :deep(.n-list-item__prefix) {
      display: flex;
      align-items: center;
      justify-content: center;
      min-width: 50px;
      height: 50px;
      border-radius: 8px;
      background-color: var(--n-border-color);
      overflow: hidden;
      transition: background-color 0.3s;
    }
  }
}
.n-empty {
  padding: 40px 0;
}
</style>
