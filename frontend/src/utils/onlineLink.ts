// 在线内容在网易云网页上的地址，用于“分享链接”与“打开源页面”（照搬 SPlayer 的 getShareUrl）。
// 依赖边界不允许播放端写网易云的地址，模板由服务端在账号状态中给出（`links`）。
// 与 SPlayer 不同：只有网页版格式，没有“分享链接格式”设置。
import type { OnlineLinkType } from "@/api/types";
import { useUserStore } from "@/stores/user";
import { copyData } from "@/utils/helper";

/** 在线内容的网页地址；服务器没有配置网易云时为空。 */
export const onlineLink = (type: OnlineLinkType, id: number): string =>
  useUserStore().netease.links?.[type]?.replace("{id}", String(id)) ?? "";

/** 复制分享链接，提示照搬 SPlayer。 */
export const copyShareLink = (
  type: OnlineLinkType,
  id: number,
  message = "已复制分享链接到剪贴板",
) => {
  const link = onlineLink(type, id);
  if (!link) return window.$message.error("网易云暂时不可用，无法生成链接");
  void copyData(link, message);
};

/** 在新标签页打开网易云的页面。 */
export const openSourcePage = (type: OnlineLinkType, id: number) => {
  const link = onlineLink(type, id);
  if (!link) return window.$message.error("网易云暂时不可用，无法打开页面");
  window.open(link);
};
