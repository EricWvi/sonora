<!-- 搜索结果：歌曲。照搬 SPlayer views/Search/songs.vue，在曲库中搜索，分页载入。
“播放”只放入已经载入的部分；双击把歌曲加入下一首并播放（与 SPlayer 相同）。 -->
<template>
  <div class="search-type">
    <ListError
      v-if="list.error.value"
      :error="list.error.value"
      @retry="list.reload()"
    />
    <Transition v-else name="fade" mode="out-in">
      <SongList
        v-if="list.items.value.length || !list.loaded.value"
        :data="list.items.value"
        :loading="list.loading.value"
        doubleClickAction="add"
        :loadMore="list.hasMore.value"
        disabledSort
        @reachBottom="list.loadMore()"
      />
      <n-empty
        v-else
        :description="`很抱歉，未能找到与 ${keyword} 相关的任何歌曲`"
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
import { usePagedList } from "@/composables/List/usePagedList";
import SongList from "@/components/List/SongList.vue";
import ListError from "@/components/List/ListError.vue";

const props = defineProps<{
  keyword: string;
}>();

// 搜索数据
const list = usePagedList(
  (offset, limit) => source.tracks({ q: props.keyword, offset, limit }),
  50,
);

onMounted(() => list.reload());
</script>
