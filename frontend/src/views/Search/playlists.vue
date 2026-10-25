<!-- 搜索结果：歌单。照搬 SPlayer views/Search/playlists.vue，在曲库中搜索，只含自己的歌单，一次载入全部。 -->
<template>
  <div class="search-type">
    <ListError
      v-if="list.error.value"
      :error="list.error.value"
      @retry="list.reload()"
    />
    <Transition v-else name="fade" mode="out-in">
      <CoverList
        v-if="list.items.value.length || !list.loaded.value"
        :data="list.items.value"
        :loading="list.loading.value"
        type="playlist"
        :hiddenCover="settingStore.hiddenCovers.playlist"
      />
      <n-empty
        v-else
        :description="`很抱歉，未能找到与 ${keyword} 相关的任何歌单`"
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
import CoverList from "@/components/List/CoverList.vue";

const props = defineProps<{
  keyword: string;
}>();

const settingStore = useSettingStore();

// 搜索数据
const list = usePagedList(() => source.playlists({ q: props.keyword }), 50);

onMounted(() => list.reload());
</script>
