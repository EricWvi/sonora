import type { Page } from "@/api/types";
import { ApiError } from "@/api/source";

/**
 * 分页列表：首次载入一页，触底或点击“加载更多”时追加下一页。
 * 用于全部歌曲、专辑与艺术家网格、艺术家的歌曲与专辑、搜索结果。
 * 请求参数变化时由页面调用 `reload`；过期请求的结果会被丢弃。
 * @param fetchPage 读取一页
 * @param pageSize 每页数量
 */
export const usePagedList = <T>(
  fetchPage: (offset: number, limit: number) => Promise<Page<T>>,
  pageSize: number = 100,
) => {
  const items = shallowRef<T[]>([]);
  const total = ref<number>(0);
  const loading = ref<boolean>(true);
  const loaded = ref<boolean>(false);
  /** 最近一次请求的错误说明；成功后清空 */
  const error = ref<string>("");
  const hasMore = computed(() => items.value.length < total.value);

  // 请求序号，丢弃过期的结果
  let requestId = 0;

  const load = async (reset: boolean) => {
    const id = ++requestId;
    loading.value = true;
    error.value = "";
    try {
      const offset = reset ? 0 : items.value.length;
      const page = await fetchPage(offset, pageSize);
      if (id !== requestId) return;
      items.value = reset ? page.items : [...items.value, ...page.items];
      total.value = page.total;
      loaded.value = true;
    } catch (e) {
      if (id !== requestId) return;
      if (reset) {
        items.value = [];
        total.value = 0;
      }
      error.value = e instanceof ApiError ? e.message : "加载失败";
      console.error("列表加载失败", e);
    } finally {
      if (id === requestId) loading.value = false;
    }
  };

  /** 重新载入第一页 */
  const reload = () => load(true);

  /** 追加下一页 */
  const loadMore = () => {
    if (loading.value || !hasMore.value) return;
    return load(false);
  };

  return { items, total, loading, loaded, error, hasMore, reload, loadMore };
};
