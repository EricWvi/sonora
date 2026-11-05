// 照搬 SPlayer utils/meta.ts 中页面用到的常量。
import { h } from "vue";
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

/**
 * 渲染图片工具栏，下载按钮改为执行 onDownload。
 * Naive UI 的下载按钮只会跳转到图片地址，也没有提供替换点击行为的入口，
 * 所以在捕获阶段拦下点击；保留原节点是为了留住图标和提示。
 * @param onDownload 点击下载按钮时执行
 * @returns 图片工具栏的渲染函数
 */
export const renderDownloadToolbar =
  (onDownload: () => void) =>
  ({ nodes }: ImageRenderToolbarProps) =>
    renderToolbar({
      nodes: {
        ...nodes,
        download: h(
          "span",
          {
            style: { display: "contents" },
            onClickCapture: (e: MouseEvent) => {
              e.stopPropagation();
              onDownload();
            },
          },
          nodes.download,
        ),
      },
    });
