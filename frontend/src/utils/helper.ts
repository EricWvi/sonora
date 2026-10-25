// 照搬 SPlayer utils/helper.ts 中播放器与页面用到的函数。
import { h } from "vue";
import Fuse from "fuse.js";
import type { Track } from "@/api/types";
import SvgIcon from "@/components/Global/SvgIcon.vue";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyObject = { [key: string]: any };

/**
 * 渲染图标
 * @param iconName 图标名称
 * @param option 图标选项（大小和样式）
 * @returns 图标组件
 */
export const renderIcon = (
  iconName: string,
  option: {
    size?: number;
    style?: AnyObject;
  } = {},
) => {
  const { size, style } = option;
  return () => {
    return h(SvgIcon, { name: iconName, size, style });
  };
};

/**
 * 封面加载完成时，设置透明度为 1
 * @param e 事件对象
 */
export const coverLoaded = (e: Event) => {
  const target = e.target as HTMLElement | null;
  if (target && target.nodeType === Node.ELEMENT_NODE) {
    target.style.opacity = "1";
  }
};

/**
 * 复制数据到剪贴板（原生实现）
 * @param text 要复制的数据
 * @param message 复制成功提示消息（可选）
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const copyData = async (text: any, message?: string) => {
  if (!text) return;
  const content =
    typeof text === "string"
      ? text.trim()
      : Array.isArray(text)
        ? text.join("\n")
        : JSON.stringify(text, null, 2);
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(content);
      window.$message.success(message ?? "已复制到剪贴板");
      return;
    } catch (err) {
      console.error("clipboard.writeText 失败，尝试降级方案", err);
    }
  }
  // 降级方案
  try {
    const textarea = document.createElement("textarea");
    textarea.value = content;
    // 避免页面滚动
    textarea.style.position = "fixed";
    textarea.style.top = "-9999px";
    textarea.style.left = "-9999px";
    // 添加到页面
    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();
    // 执行复制
    const success = document.execCommand("copy");
    document.body.removeChild(textarea);
    if (success) {
      window.$message.success(message ?? "已复制到剪贴板");
    } else {
      throw new Error("execCommand 返回 false");
    }
  } catch (error) {
    window.$message.error("复制出错，请重试");
    console.error("复制出错：", error);
  }
};

/**
 * 模糊搜索。字段按 Sonora 的歌曲改为标题、艺术家与专辑，权重与 SPlayer 相同
 * @param keyword 搜索关键词
 * @param data 歌曲列表
 */
export const fuzzySearch = (keyword: string, data: Track[]): Track[] => {
  try {
    if (!keyword || !data || !Array.isArray(data)) return [];

    const fuse = new Fuse(data, {
      // 针对歌曲可读字段进行索引
      keys: [
        { name: "title", weight: 0.5 },
        { name: "artists.name", weight: 0.15 },
        { name: "album.title", weight: 0.1 },
      ],
      threshold: 0.35, // 0 精确匹配 ~ 1 完全模糊
      ignoreLocation: true, // 不要求关键词位置接近
    });

    return fuse.search(keyword).map((result) => result.item);
  } catch (error) {
    console.error("模糊搜索出现错误：", error);
    return [];
  }
};
