// 照搬 SPlayer utils/meta.ts 中页面用到的常量。
import type { ImageRenderToolbarProps } from "naive-ui";
import type { SortField, SortOrder } from "@/api/types";

export type ListSortField = SortField;

/**
 * 排序字段选项。只保留服务端支持的字段（阶段 2 的“歌曲列表”），名称与 SPlayer 相同
 */
export const sortFieldOptions = {
  default: { name: "默认" },
  title: { name: "标题" },
  artist: { name: "歌手" },
  album: { name: "专辑" },
  duration: { name: "时长" },
  createTime: { name: "添加时间" },
} as const satisfies Record<SortField, { name: string }>;

/**
 * 排序方式选项
 */
export const sortOrderOptions = {
  default: { name: "默认" },
  asc: { name: "升序" },
  desc: { name: "降序" },
} as const;

/**
 * 列表请求的排序参数；默认排序或默认顺序时省略，使用服务端的默认顺序
 * @param field 排序字段
 * @param order 排序顺序
 */
export const listSortQuery = (
  field: SortField,
  order: "default" | SortOrder,
): { sort?: SortField; order?: SortOrder } =>
  field === "default" || order === "default" ? {} : { sort: field, order };

/**
 * 渲染图片工具栏
 * @param nodes 图片工具栏节点
 * @returns 图片工具栏
 */
export const renderToolbar = ({ nodes }: ImageRenderToolbarProps) => {
  return [
    nodes.prev,
    nodes.next,
    nodes.rotateCounterclockwise,
    nodes.rotateClockwise,
    nodes.resizeToOriginalSize,
    nodes.zoomOut,
    nodes.zoomIn,
    nodes.download,
    nodes.close,
  ];
};
