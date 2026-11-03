<!-- 艺术家网格。标题与操作栏沿用 SPlayer 本地音乐页面，网格为 ArtistList；按名称排列，
分页载入，搜索由服务端完成。 -->
<template>
  <div class="library-artists">
    <ListPageHeader title="艺术家">
      <template #status>
        <n-text class="item">
          <SvgIcon name="Artist" :depth="3" />
          <!-- 与 SPlayer 不同：总数直接显示，不用 n-number-animation 从 0 滚动 -->
          {{ list.total.value }} 位艺术家
        </n-text>
      </template>
      <template #left>
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
    <ArtistList
      v-else
      :data="list.items.value"
      :loading="list.loading.value"
      :load-more="list.hasMore.value"
      :hidden-cover="settingStore.hiddenCovers.artist"
      @load-more="list.loadMore()"
    />
  </div>
</template>

<script setup lang="ts">
import { source } from "@/api/source";
import { useSettingStore } from "@/stores";
import { usePagedList } from "@/composables/List/usePagedList";
import ListPageHeader from "@/components/List/ListPageHeader.vue";
import ListError from "@/components/List/ListError.vue";
import ArtistList from "@/components/List/ArtistList.vue";

const settingStore = useSettingStore();

// 搜索：输入停顿后按服务端的 q 重新载入
const searchValue = ref<string>("");
const searchKeyword = refDebounced(
  computed(() => searchValue.value.trim()),
  300,
);

const list = usePagedList(
  (offset, limit) => source.artists({ q: searchKeyword.value, offset, limit }),
  50,
);

watch(searchKeyword, () => list.reload(), { immediate: true });
</script>

<style lang="scss" scoped>
.library-artists {
  display: flex;
  flex-direction: column;
}
</style>
