import { ApiError } from "@/api/source";

/**
 * 详情页的加载错误，交给 ListError 显示。404 时说明为 `missingText`，
 * 并标记 `missing`：内容已被删除，不再提供重试（阶段 5 的异常状态）。
 * @param missingText 内容不存在时的说明，例如“专辑不存在或已被删除”
 * @param fallback 不是接口错误时的说明
 */
export const useLoadError = (missingText: string, fallback = "加载失败") => {
  const error = ref<string>("");
  const missing = ref<boolean>(false);

  /** 记录加载失败 */
  const setError = (e: unknown) => {
    missing.value = e instanceof ApiError && e.status === 404;
    error.value = missing.value
      ? missingText
      : e instanceof ApiError
        ? e.message
        : fallback;
  };

  /** 开始重新加载时清除 */
  const clearError = () => {
    error.value = "";
    missing.value = false;
  };

  return { error, missing, setError, clearError };
};
