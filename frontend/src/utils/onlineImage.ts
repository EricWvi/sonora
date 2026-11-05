/**
 * 网易云图片地址：照搬 SPlayer，换成 https，并用 `param` 取合适的尺寸。
 * 这是图片资源的地址，不是接口请求，由浏览器直接加载。
 */
export function onlineImage(url: string | undefined, size: number): string {
  if (!url) return "";
  return `${url.replace(/^http:/, "https:")}?param=${size}y${size}`;
}
