<!-- 专辑网格。标题与操作栏沿用 SPlayer 本地音乐页面，网格为 CoverList；分页载入，
排序（按标题或最近添加）与搜索由服务端完成。排序按钮沿用 SPlayer 窄窗口下的标签页下拉按钮。 -->
<template>
  <div class="library-albums">
    <ListPageHeader title="专辑">
      <template #status>
        <n-text class="item">
          <SvgIcon name="Album" :depth="3" />
          <n-number-animation :from="0" :to="list.total.value" /> 张专辑
        </n-text>
      </template>
      <template #left>
        <n-dropdown
          :options="sortOptions"
          :value="sort"
          trigger="click"
          placement="bottom-start"
          @select="(key: AlbumSort) => (sort = key)"
        >
          <n-button :focusable="false" strong secondary round>
            {{ sortOptions.find((option) => option.key === sort)?.label }}
            <template #icon>
              <SvgIcon name="Down" />
            </template>
          </n-button>
        </n-dropdown>
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
    <CoverList
      v-else
      :data="list.items.value"
      :loading="list.loading.value"
      :load-more="list.hasMore.value"
      :hidden-cover="settingStore.hiddenCovers.album"
      :empty-description="
        searchKeyword ? `搜不到关于 ${searchKeyword} 的任何专辑呀` : undefined
      "
      type="album"
      @load-more="list.loadMore()"
    />
  </div>
</template>

<script setup lang="ts">
import type { DropdownOption } from "naive-ui";
import { source } from "@/api/source";
import { useSettingStore } from "@/stores";
import { usePagedList } from "@/composables/List/usePagedList";
import ListPageHeader from "@/components/List/ListPageHeader.vue";
import ListError from "@/components/List/ListError.vue";
import CoverList from "@/components/List/CoverList.vue";

type AlbumSort = "default" | "createTime";

const settingStore = useSettingStore();

// 排序：专辑列表只接受按标题（default）与入库时间
const sort = ref<AlbumSort>("default");
const sortOptions: (DropdownOption & { key: AlbumSort })[] = [
  { label: "按标题", key: "default" },
  { label: "最近添加", key: "createTime" },
];

// 搜索：输入停顿后按服务端的 q 重新载入
const searchValue = ref<string>("");
const searchKeyword = refDebounced(
  computed(() => searchValue.value.trim()),
  300,
);

const list = usePagedList(
  (offset, limit) =>
    source.albums({
      q: searchKeyword.value,
      offset,
      limit,
      ...(sort.value === "createTime"
        ? { sort: "createTime", order: "desc" }
        : {}),
    }),
  50,
);

watch([searchKeyword, sort], () => list.reload(), { immediate: true });
</script>

<style lang="scss" scoped>
.library-albums {
  display: flex;
  flex-direction: column;
}
</style>
