import type { Track } from "@/api/types";
import { useStatusStore } from "@/stores";

/**
 * 列表详情逻辑。照搬 SPlayer composables/List/useListDetail.ts，
 * 详情数据的类型由页面决定（专辑、歌单或我喜欢的音乐），不迁入移动端高度。
 */
export const useListDetail = <T>() => {
  const statusStore = useStatusStore();

  const detailData = ref<T | null>(null) as Ref<T | null>;
  const listData = shallowRef<Track[]>([]);
  const loading = ref<boolean>(true);

  /**
   * 计算列表高度
   */
  const getSongListHeight = (listScrolling: boolean) => {
    return statusStore.mainContentHeight - (listScrolling ? 120 : 240);
  };

  /**
   * 重置数据
   */
  const resetData = (resetList: boolean = true) => {
    detailData.value = null;
    if (resetList) {
      listData.value = [];
    }
  };

  /**
   * 设置详情数据
   */
  const setDetailData = (data: T | null) => {
    detailData.value = data;
  };

  /**
   * 设置列表数据
   */
  const setListData = (data: Track[]) => {
    listData.value = data;
  };

  /**
   * 追加列表数据
   */
  const appendListData = (data: Track[]) => {
    listData.value = [...listData.value, ...data];
  };

  /**
   * 设置加载状态
   */
  const setLoading = (value: boolean) => {
    loading.value = value;
  };

  return {
    detailData,
    listData,
    loading,
    getSongListHeight,
    resetData,
    setDetailData,
    setListData,
    appendListData,
    setLoading,
  };
};
