<!-- 艺术家的专辑。照搬 SPlayer views/Artist/albums.vue，分页载入。 -->
<template>
  <div class="artist-type">
    <ListError
      v-if="list.error.value"
      :error="list.error.value"
      @retry="list.reload()"
    />
    <CoverList
      v-else
      :data="list.items.value"
      :loading="list.loading.value"
      :loadMore="list.hasMore.value"
      type="album"
      :hiddenCover="settingStore.hiddenCovers.album"
      @loadMore="list.loadMore()"
    />
  </div>
</template>

<script setup lang="ts">
import { source } from "@/api/source";
import { useSettingStore } from "@/stores";
import { usePagedList } from "@/composables/List/usePagedList";
import CoverList from "@/components/List/CoverList.vue";
import ListError from "@/components/List/ListError.vue";

const props = defineProps<{
  id: number;
}>();

const settingStore = useSettingStore();

// 专辑数据
const list = usePagedList(
  (offset, limit) => source.albums({ artistId: props.id, offset, limit }),
  50,
);

watch(
  () => props.id,
  () => list.reload(),
  { immediate: true },
);
</script>
