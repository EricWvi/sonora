// 播放端 API 的数据形态，见 docs/api.md。字段与服务端响应一致，不含文件路径，也不含网易字段。

export type User = {
  name: string;
  email: string;
};

/** 封面与艺术家图片：内容哈希文件名与 ThumbHash 占位。 */
export type Cover = {
  name: string;
  thumbhash: string;
};

export type Page<T> = {
  items: T[];
  total: number;
};

export type Playlist = {
  id: number;
  name: string;
  description: string;
  cover: Cover | null;
  trackCount: number;
  durationMs: number;
  createdAt: string;
  updatedAt: string;
};
