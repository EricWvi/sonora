# Sonora API 约定

更新日期：2026-10-24。本文记录阶段 1 已实现的管理 API 及其约定；播放端 API（阶段 4）沿用同一套约定。实现见 [server.go](../backend/internal/server/server.go) 与 [admin.go](../backend/internal/admin/admin.go)。

## 通用约定

- **版本**：版本号写在路径中。管理 API 前缀为 `/admin/api/v1/`，播放端 API 计划使用 `/api/v1/`。不兼容的修改改用 `v2`，同一版本内只做向后兼容的增加。服务版本见 `GET /api/version`。
- **错误**：所有错误响应的正文为 `{"error": "给人看的说明", "code": "machine_code"}`。客户端先按 HTTP 状态分支，`code` 区分同一状态下的不同情况。阶段 4 照搬的 homepage 认证代码只返回 `error`，客户端不能依赖每个错误都带 `code`。
- **身份**：媒体以整数 ID 标识。ID 由 `AUTOINCREMENT` 分配，删除后不复用。文件路径、存储布局和内容哈希不出现在任何响应中；封面以内容哈希文件名引用，只能经 `covers/{name}` 读取。
- **分页**：列表接受 `offset`、`limit`（默认 50，上限 500），返回 `{"items": [...], "total": n}`。
- **写请求**：非 GET 请求若带 `Origin` 头，必须等于配置的 `public_url`，否则返回 403 `bad_origin`。

| 状态 | code | 场景 |
| --- | --- | --- |
| 400 | `bad_request` / `invalid` | 请求格式错误 / 输入不合法（说明可直接展示） |
| 401 | `unauthenticated` | 没有会话或会话已过期 |
| 401 | `invalid_credentials` | 用户名或密码错误 |
| 403 | `not_initialized` | 尚未创建管理员时访问其他管理接口 |
| 403 | `invalid_token` | 初始化 token 缺失或错误 |
| 403 | `bad_origin` | 跨站写请求 |
| 404 | `not_found` | 内容不存在或已删除 |
| 409 | `already_initialized` | 重复初始化 |
| 409 | `conflict` | 重命名与现有专辑或艺术家冲突，应改用合并 |
| 413 | `too_large` | 音频超过 1 GB、封面超过 10 MB 或歌词超过 512 KB |
| 415 | `unsupported_format` | 不是 MP3、FLAC、AAC（M4A）、Ogg Vorbis 或 Opus |
| 422 | `corrupt_file` | 格式可识别但结构损坏或不完整 |
| 429 | `rate_limited` | 初始化或登录失败次数过多，带 `Retry-After` |

## 管理 API

会话 Cookie 为 `sonora_admin_session`，`Path=/admin/api/v1/`、`HttpOnly`、`SameSite=Strict`，`public_url` 为 https 时带 `Secure`；与播放端会话互不通用。会话闲置 7 天或登录满 30 天后失效。

无需会话：

| 方法与路径 | 说明 |
| --- | --- |
| `GET status` | `{initialized, devMode, admin}`；`admin` 为当前会话的管理员或 `null` |
| `POST setup` | `{token, username, password}`，成功后返回状态并设置会话 |
| `POST login` | `{username, password}`，用户名不区分大小写 |
| `POST logout` | 结束当前会话，204 |

需要会话：

| 方法与路径 | 说明 |
| --- | --- |
| `GET overview` | 歌曲、专辑、艺术家数量，缺标签数量，总时长，音频与封面占用字节 |
| `GET tracks` | `q`、`missing=true`、`albumId`、`artistId`；按专辑时按碟号与音轨号排序，否则新入库在前 |
| `POST tracks` | `multipart/form-data`：一个 `file` 部分，可选 `lyrics` 部分（外置歌词文件）。新歌曲 201，重复文件 200，正文为 `{status: "created" \| "duplicate", track}` |
| `GET/PUT/DELETE tracks/{id}` | 详情（含歌词）/ 修改元数据 / 删除歌曲及其音频文件 |
| `PUT/DELETE tracks/{id}/cover` | 请求体为图片原始字节；单曲封面优先于专辑封面 |
| `PUT/DELETE tracks/{id}/lyrics` | 请求体为歌词原始字节，服务识别 UTF-8、UTF-16 与 GB18030 |
| `GET tracks/{id}/audio` | 音频流，支持 Range，用于试听 |
| `GET albums`、`GET/PUT/DELETE albums/{id}` | 列表（`q`、`artistId`）/ 详情 / 修改 `{title, artists, year}` / 删除专辑及其全部歌曲 |
| `POST albums/{id}/merge` | `{sourceIds}` 合并到路径中的专辑 |
| `PUT/DELETE albums/{id}/cover` | 同单曲封面 |
| `GET artists`、`GET/PUT artists/{id}` | 列表（`q`）/ 详情 / 重命名 `{name}` |
| `POST artists/{id}/merge` | `{sourceIds}`；合并后艺术家组合相同的同名专辑一并合并 |
| `GET covers/{name}` | 封面图片，按内容哈希命名，可长期缓存 |

歌曲修改 `PUT tracks/{id}` 的正文为 `{title, artists, album, albumArtists, trackNo, discNo, year}`，`artists` 与 `albumArtists` 是有序的名称数组。艺术家与专辑按名称查找，不存在时创建；名称去掉首尾空白，空名与重复名（不区分大小写）忽略，最多 50 位；`albumArtists` 为空时使用 `artists`。没有歌曲的专辑、既无歌曲也无专辑的艺术家自动删除，因此不提供单独删除艺术家的接口。

歌曲、专辑响应中的艺术家是有序数组 `artists: [{id, name}]`，歌曲所属专辑为 `album: {id, title, artists, cover}`；没有艺术家时为空数组。按 `artistId` 筛选歌曲时，包括该艺术家参与的歌曲和其专辑中的歌曲。
