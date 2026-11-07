// 照搬 SPlayer composables/List/useListSearch.ts，歌曲是播放端 API 的 Track。
import type { Track } from "@/api/types";
import { fuzzySearch } from "@/utils/helper";
import { debounce } from "lodash-es";

/**
 * 列表搜索逻辑
 */
export const useListSearch = (listData: Ref<Track[]>) => {
  const searchValue = ref<string>("");
  const searchData = ref<Track[]>([]);

  /**
   * 清除搜索
   */
  const clearSearch = () => {
    searchValue.value = "";
    searchData.value = [];
  };

  /**
   * 执行搜索
   */
  const performSearch = debounce((val: string) => {
    val = val.trim();
    if (!val || val === "") {
      searchData.value = [];
      return;
    }
    const result = fuzzySearch(val, listData.value);
    searchData.value = result;
  }, 300);

  // 与 SPlayer 不同：列表变化时重新搜索，否则删除的歌曲会留在搜索结果里
  watch(listData, () => {
    if (searchValue.value) performSearch(searchValue.value);
  });

  /**
   * 获取显示的数据
   */
  const displayData = computed(() => {
    return searchValue.value ? searchData.value : listData.value;
  });

  return {
    searchValue,
    searchData,
    displayData,
    clearSearch,
    performSearch,
  };
};
