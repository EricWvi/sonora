import type { Page } from "@/api/types";

/** 服务端单页上限（docs/api.md 的“分页”） */
export const PAGE_LIMIT = 500;

/**
 * 逐页读取整个列表，用于“播放全部”与一次载入的详情页（专辑、歌单、我喜欢、最近播放）
 * @param fetchPage 读取一页
 */
export async function fetchAll<T>(
  fetchPage: (offset: number, limit: number) => Promise<Page<T>>,
): Promise<T[]> {
  const items: T[] = [];
  for (;;) {
    const page = await fetchPage(items.length, PAGE_LIMIT);
    items.push(...page.items);
    if (!page.items.length || items.length >= page.total) return items;
  }
}
