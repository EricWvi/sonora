<!-- 搜索结果：艺术家。照搬 SPlayer views/Search/artists.vue，在曲库中搜索，分页载入。 -->
<template>
  <div class="search-type">
    <ListError
      v-if="list.error.value"
      :error="list.error.value"
      @retry="list.reload()"
    />
    <Transition v-else name="fade" mode="out-in">
      <ArtistList
        v-if="list.items.value.length || !list.loaded.value"
        :data="list.items.value"
        :loading="list.loading.value"
        :loadMore="list.hasMore.value"
        :hiddenCover="settingStore.hiddenCovers.artist"
        @loadMore="list.loadMore()"
      />
      <n-empty
        v-else
        :description="`很抱歉，未能找到与 ${keyword} 相关的任何艺术家`"
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
import { source } from "@/api/source";
import { useSettingStore } from "@/stores";
import { usePagedList } from "@/composables/List/usePagedList";
import ListError from "@/components/List/ListError.vue";
import ArtistList from "@/components/List/ArtistList.vue";

const props = defineProps<{
  keyword: string;
}>();

const settingStore = useSettingStore();

// 搜索数据
const list = usePagedList(
  (offset, limit) => source.artists({ q: props.keyword, offset, limit }),
  50,
);

onMounted(() => list.reload());
</script>
