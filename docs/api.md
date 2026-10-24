# Sonora API 约定

更新日期：2026-10-24。本文记录已实现的管理 API（阶段 1）、播放端登录与播放端 API（阶段 4）及其通用约定。实现见 [server.go](../backend/internal/server/server.go)、[player.go](../backend/internal/server/player.go)、[admin.go](../backend/internal/admin/admin.go) 与 [auth.go](../backend/internal/auth/auth.go)。

## 通用约定

- **版本**：版本号写在路径中。管理 API 前缀为 `/admin/api/v1/`，播放端 API 前缀为 `/api/v1/`。不兼容的修改改用 `v2`，同一版本内只做向后兼容的增加。服务版本见 `GET /api/version`。
- **错误**：所有错误响应的正文为 `{"error": "给人看的说明", "code": "machine_code"}`。客户端先按 HTTP 状态分支，`code` 区分同一状态下的不同情况。照搬自 homepage 的播放端认证（`/auth/` 及播放端 API 的 401/403）只返回 `error`，客户端不能依赖每个错误都带 `code`。
- **身份**：媒体以整数 ID 标识。ID 由 `AUTOINCREMENT` 分配，删除后不复用。文件路径、存储布局和内容哈希不出现在任何响应中；封面以内容哈希文件名引用，只能经 `covers/{name}` 读取。
- **分页**：列表接受 `offset`、`limit`（默认 50，上限 500），返回 `{"items": [...], "total": n}`。
- **写请求**：非 GET 请求若带 `Origin` 头，必须等于配置的 `public_url`，否则返回 403 `bad_origin`。

| 状态 | code | 场景 |
| --- | --- | --- |
| 400 | `bad_request` / `invalid` | 请求格式错误 / 输入不合法（说明可直接展示） |
| 401 | `unauthenticated` | 没有管理会话或会话已过期（播放端为不带 `code` 的 401，见下文） |
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
| `GET covers/{name}` | 封面原图，按内容哈希命名，可长期缓存。`?size=96` 或 `?size=300` 返回长边不超过该值的 JPEG 缩略图，首次请求时生成；原图已不大于该尺寸或无法解码时返回原图；其他尺寸 400 |

歌曲修改 `PUT tracks/{id}` 的正文为 `{title, artists, album, albumArtists, trackNo, discNo, year}`，`artists` 与 `albumArtists` 是有序的名称数组。艺术家与专辑按名称查找，不存在时创建；名称去掉首尾空白，空名与重复名（不区分大小写）忽略，最多 50 位；`albumArtists` 为空时使用 `artists`。没有歌曲的专辑、既无歌曲也无专辑的艺术家自动删除，因此不提供单独删除艺术家的接口。

歌曲、专辑及歌曲所属专辑的封面为 `cover: {name, thumbhash}`，没有封面时为 `null`。`thumbhash` 是 [ThumbHash](https://evanw.github.io/thumbhash/) 的标准 base64，客户端解码为模糊占位图，在缩略图加载完成前显示；图片无法解码时为空字符串。

歌曲、专辑响应中的艺术家是有序数组 `artists: [{id, name}]`，歌曲所属专辑为 `album: {id, title, artists, cover}`；没有艺术家时为空数组。按 `artistId` 筛选歌曲时，包括该艺术家参与的歌曲和其专辑中的歌曲。

## 播放端登录

照搬 `../homepage` 的 OIDC 实现（参考提交 `f188285`），只改了包路径、Cookie 名称、配置键名和页面标题。

| 方法与路径 | 说明 |
| --- | --- |
| `GET /auth/login` | 跳转到 Authelia；授权码流程加 PKCE（S256），scope `openid profile email`。Authelia 不可达时返回 503 页面 |
| `GET /auth/callback` | 校验 state、ID token（签名、issuer、audience、有效期、nonce）后建立会话，跳回 `/` |
| `POST /auth/logout` | 只结束本浏览器的会话，不退出 Authelia，204 |

会话 Cookie 为 `sonora_session`，`Path=/`、`HttpOnly`、`SameSite=Lax`，https 时带 `Secure`。服务端会话不过期，直到退出；使用中每天续发一次 Cookie，避免浏览器 400 天上限。用户按 `(issuer, subject)` 识别，名称与邮箱每次登录时刷新。开发模式不走 OIDC，所有请求都以写死的用户 `listener` 登录，`/auth/login` 直接跳回 `/`。

未登录访问播放端 API 返回 401 `{"error": "未登录"}`，播放端应跳转 `/auth/login`；跨站写请求返回 403。管理会话与播放端会话互不通用。

## 播放端 API

全部需要播放端会话。播放端看不到只与管理有关的字段：原文件名、文件大小、入库时间、歌词正文（另有接口）都不出现在列表与详情中。列表的分页、`q` 搜索和排序与管理 API 相同：歌曲新入库在前，按专辑筛选时按碟号与音轨号；专辑按标题；艺术家按名称。

| 方法与路径 | 说明 |
| --- | --- |
| `GET me` | 当前用户 `{name, email}` |
| `GET tracks` | `q`（匹配标题、专辑名、艺术家名）、`albumId`、`artistId` |
| `GET tracks/{id}` | 歌曲详情 |
| `GET tracks/{id}/lyrics` | `{lyrics}`，原样返回 LRC 或纯文本，没有歌词时为空字符串；解析由播放端完成 |
| `GET tracks/{id}/audio` | 音频流，支持 Range |
| `GET albums`、`GET albums/{id}` | 列表（`q`、`artistId`）/ 详情 |
| `GET artists`、`GET artists/{id}` | 列表（`q`）/ 详情；艺术家的歌曲与专辑用 `tracks?artistId=`、`albums?artistId=` 获取 |
| `GET covers/{name}` | 与管理 API 相同，支持 `?size=96` 与 `?size=300` |

数据形态：

- 歌曲 `{id, title, artists, album, trackNo, discNo, year, durationMs, format, cover, hasLyrics}`。`album` 为 `{id, title, artists}` 或 `null`；`cover` 是歌曲实际显示的封面，单曲封面优先，否则为专辑封面，都没有时为 `null`。
- 专辑 `{id, title, artists, year, cover, trackCount, durationMs}`。
- 艺术家 `{id, name, trackCount, albumCount}`。
- `artists` 与 `cover` 的形态和管理 API 一致：`[{id, name}]` 与 `{name, thumbhash}`。

喜爱、歌单与播放历史尚未实现。
