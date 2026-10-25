<!-- 艺术家的歌曲。照搬 SPlayer views/Artist/songs.vue，分页载入，排序由服务端完成。 -->
<template>
  <div class="artist-songs">
    <ListError
      v-if="list.error.value"
      :error="list.error.value"
      @retry="list.reload()"
    />
    <SongList
      v-else
      :data="list.items.value"
      :loading="list.loading.value"
      :load-more="list.hasMore.value"
      disableHeightTransition
      @reachBottom="list.loadMore()"
      @scroll="emit('scroll', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { source } from "@/api/source";
import { useStatusStore } from "@/stores";
import { listSortQuery } from "@/utils/meta";
import { debounce } from "lodash-es";
import { usePlayerController } from "@/core/player/PlayerController";
import { usePagedList } from "@/composables/List/usePagedList";
import SongList from "@/components/List/SongList.vue";
import ListError from "@/components/List/ListError.vue";

const props = defineProps<{
  id: number;
}>();

const emit = defineEmits<{
  scroll: [e: Event];
}>();

const player = usePlayerController();
const statusStore = useStatusStore();

// 歌曲数据
const list = usePagedList((offset, limit) =>
  source.tracks({
    artistId: props.id,
    offset,
    limit,
    ...listSortQuery(statusStore.listSortField, statusStore.listSortOrder),
  }),
);

// 播放已经载入的歌曲
const playAllSongs = debounce(() => {
  if (!list.items.value.length) return;
  player.updatePlayList(list.items.value);
}, 300);

defineExpose({ playAllSongs, songData: list.items });

watch(
  [
    () => props.id,
    () => statusStore.listSortField,
    () => statusStore.listSortOrder,
  ],
  () => list.reload(),
  { immediate: true },
);
</script>
