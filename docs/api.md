# Sonora API 约定

更新日期：2026-10-25。本文记录已实现的管理 API（阶段 1）、播放端登录与播放端 API（阶段 4）及其通用约定。实现见 [server.go](../backend/internal/server/server.go)、[player.go](../backend/internal/server/player.go)、[admin.go](../backend/internal/admin/admin.go) 与 [auth.go](../backend/internal/auth/auth.go)。

## 通用约定

- **版本**：版本号写在路径中。管理 API 前缀为 `/admin/api/v1/`，播放端 API 前缀为 `/api/v1/`。不兼容的修改改用 `v2`，同一版本内只做向后兼容的增加。服务版本见 `GET /api/version`。
- **错误**：所有错误响应的正文为 `{"error": "给人看的说明", "code": "machine_code"}`。客户端先按 HTTP 状态分支，`code` 区分同一状态下的不同情况。照搬自 homepage 的播放端认证（`/auth/` 及播放端 API 的 401/403）只返回 `error`，客户端不能依赖每个错误都带 `code`。
- **身份**：媒体以整数 ID 标识。ID 由 `AUTOINCREMENT` 分配，删除后不复用。文件路径、存储布局和内容哈希不出现在任何响应中；封面以内容哈希文件名引用，只能经 `covers/{name}` 读取。
- **分页**：列表接受 `offset`、`limit`（默认 50，上限 500），返回 `{"items": [...], "total": n}`。
- **排序**：歌曲列表接受 `sort` 与 `order`。`sort` 取 `default`（默认）、`title`、`artist`、`album`、`duration`、`createTime`，与 SPlayer 的 `sortFieldOptions` 同名；`order` 取 `asc`（默认）或 `desc`。`default` 是列表自身的顺序，此时忽略 `order`；按其他字段排序时，相等的项保持列表自身的顺序。`artist` 按第一位艺术家，没有艺术家或专辑的歌曲视为空名。名称按拼音、数字按数值比较（对应 SPlayer 的 `Intl.Collator("zh-CN", {numeric: true})`），专辑与艺术家列表的默认顺序也使用同一规则。专辑列表只接受 `default`（按标题）与 `createTime`。其他取值返回 400 `invalid`。
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
| 413 | `too_large` | 音频超过 1 GB、图片超过 10 MB、歌词超过 512 KB 或设置超过 256 KB |
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
| `GET tracks` | `q`、`missing=true`、`albumId`、`artistId`、`sort`、`order`；默认按专辑时按碟号与音轨号排序，否则新入库在前 |
| `POST tracks` | `multipart/form-data`：一个 `file` 部分，可选 `lyrics` 部分（外置 LRC 或 TTML 歌词文件）。新歌曲 201，重复文件 200，正文为 `{status: "created" \| "duplicate", track}` |
| `GET/PUT/DELETE tracks/{id}` | 详情（含歌词）/ 修改元数据 / 删除歌曲及其音频文件 |
| `PUT/DELETE tracks/{id}/cover` | 请求体为图片原始字节；单曲封面优先于专辑封面 |
| `PUT/DELETE tracks/{id}/lyrics` | 请求体为 LRC 或 TTML 歌词的原始字节，服务识别 UTF-8、UTF-16 与 GB18030 |
| `GET tracks/{id}/audio` | 音频流，支持 Range，用于试听 |
| `GET albums`、`GET/PUT/DELETE albums/{id}` | 列表（`q`、`artistId`、`sort`、`order`）/ 详情 / 修改 `{title, artists, year}` / 删除专辑及其全部歌曲 |
| `POST albums/{id}/merge` | `{sourceIds}` 合并到路径中的专辑 |
| `PUT/DELETE albums/{id}/cover` | 同单曲封面 |
| `GET artists`、`GET/PUT artists/{id}` | 列表（`q`）/ 详情 / 重命名 `{name}` |
| `POST artists/{id}/merge` | `{sourceIds}`；合并后艺术家组合相同的同名专辑一并合并；目标没有图片时沿用来源的图片 |
| `PUT/DELETE artists/{id}/image` | 艺术家图片，请求体为图片原始字节，存储与缩略图同封面 |
| `GET covers/{name}` | 封面或艺术家图片的原图，按内容哈希命名，可长期缓存。`?size=` 取 `96`、`100`、`300` 或 `1024`，返回长边不超过该值的 JPEG 缩略图，首次请求时生成；原图已不大于该尺寸或无法解码时返回原图；其他尺寸 400 |

歌词有两种格式，由内容判断，与文件扩展名无关：以 `<` 开头的是 TTML，必须是根元素为 `tt` 的合法 XML，否则返回 400；其余都按 LRC（含逐字与增强型 LRC，由播放端区分）。音频内嵌的歌词按 LRC 处理。歌曲的 `lyricsFormat` 为 `lrc` 或 `ttml`，没有歌词时为空字符串。

歌曲修改 `PUT tracks/{id}` 的正文为 `{title, artists, album, albumArtists, trackNo, discNo, year}`，`artists` 与 `albumArtists` 是有序的名称数组。艺术家与专辑按名称查找，不存在时创建；名称去掉首尾空白，空名与重复名（不区分大小写）忽略，最多 50 位；`albumArtists` 为空时使用 `artists`。没有歌曲的专辑、既无歌曲也无专辑的艺术家自动删除，因此不提供单独删除艺术家的接口。

歌曲、专辑及歌曲所属专辑的封面为 `cover: {name, thumbhash}`，没有封面时为 `null`；艺术家的图片 `image` 形态相同。`thumbhash` 是 [ThumbHash](https://evanw.github.io/thumbhash/) 的标准 base64，客户端解码为模糊占位图，在缩略图加载完成前显示；图片无法解码时为空字符串。

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

全部需要播放端会话。播放端看不到只与管理有关的字段：原文件名、文件大小、入库时间、歌词正文（另有接口）都不出现在列表与详情中。列表的分页、`q` 搜索和[排序](#通用约定)与管理 API 相同：歌曲默认新入库在前，按专辑筛选时按碟号与音轨号；专辑默认按标题；艺术家按名称。

| 方法与路径 | 说明 |
| --- | --- |
| `GET me` | 当前用户 `{name, email}` |
| `GET tracks` | `q`（匹配标题、专辑名、艺术家名）、`albumId`、`artistId`、`sort`、`order` |
| `GET tracks/{id}` | 歌曲详情 |
| `GET tracks/{id}/lyrics` | `{format, lyrics}`：`format` 为 `lrc` 或 `ttml`，`lyrics` 为原文；没有歌词时都是空字符串。解析由播放端完成 |
| `GET tracks/{id}/audio` | 音频流，支持 Range |
| `GET albums`、`GET albums/{id}` | 列表（`q`、`artistId`、`sort`、`order`）/ 详情；首页“最近添加的专辑”用 `sort=createTime&order=desc` |
| `GET artists`、`GET artists/{id}` | 列表（`q`）/ 详情；艺术家的歌曲与专辑用 `tracks?artistId=`、`albums?artistId=` 获取 |
| `GET covers/{name}` | 封面与艺术家图片，与管理 API 相同，支持同样的 `?size=`。播放端按 SPlayer 的封面尺寸取 `100`（小图：底部栏、背景与取色）、`300`、`1024`（全屏播放器的大封面），2026-10-25 增加 `100` 与 `1024` |

数据形态：

- 歌曲 `{id, title, artists, album, trackNo, discNo, year, durationMs, format, cover, hasLyrics}`。`album` 为 `{id, title, artists}` 或 `null`；`cover` 是歌曲实际显示的封面，单曲封面优先，否则为专辑封面，都没有时为 `null`。
- 专辑 `{id, title, artists, year, cover, trackCount, durationMs}`。
- 艺术家 `{id, name, image, trackCount, albumCount}`，`image` 是管理后台上传的图片，没有时为 `null`。
- `artists` 与 `cover` 的形态和管理 API 一致：`[{id, name}]` 与 `{name, thumbhash}`。

## 播放端用户数据

设置、喜爱、歌单与播放历史属于当前登录的用户，同样需要播放端会话。别人的歌单一律返回 404，不区分“不存在”与“无权访问”。管理端删除歌曲时，它会从所有人的喜爱、歌单与历史中消失。

| 方法与路径 | 说明 |
| --- | --- |
| `GET settings` | 当前用户保存的外观与歌词设置，原样返回播放端保存的 JSON 对象；从未保存时为 `{}` |
| `PUT settings` | 请求体为完整的设置对象，整体替换，204。服务端只校验是 JSON 对象且不超过 256 KB，不解释内容；输出设备、音量等设备相关设置不上传 |
| `GET likes` | 喜爱的歌曲，默认最近喜爱的在前，分页，可排序 |
| `GET likes/ids` | `{ids}`：全部喜爱歌曲的 ID，供播放端在任意列表中标记喜爱状态 |
| `PUT likes/{trackId}`、`DELETE likes/{trackId}` | 喜爱 / 取消，幂等，204；歌曲不存在时喜爱返回 404 |
| `GET history` | `{items: [{track, playedAt}], total}`，最近播放的在前；同一首歌只保留一条，重播移到最前；每人最多保留 500 首，与 SPlayer 一致 |
| `POST history` | `{trackId}` 记录一次播放，204。什么时候算一次播放由播放端决定 |
| `DELETE history`、`DELETE history/{trackId}` | 清空 / 删除一条，204 |
| `GET playlists` | 自己的歌单，新建的在前；`q` 匹配歌单名 |
| `POST playlists` | `{name, description}` 新建，201。名称 1–100 个字符，简介最多 1000 个字符 |
| `GET/PUT/DELETE playlists/{id}` | 详情 / 修改 `{name, description}` / 删除 |
| `GET playlists/{id}/tracks` | 歌单中的歌曲，默认按歌单顺序，分页，可排序 |
| `POST playlists/{id}/tracks` | `{trackIds}` 按给定顺序追加到末尾，已在歌单中的跳过；返回 `{added, playlist}`。任一歌曲不存在时整体失败（404）；每个歌单最多 10000 首 |
| `DELETE playlists/{id}/tracks/{trackId}` | 移出歌单，返回歌单 |
| `PUT playlists/{id}/order` | `{trackIds}` 为歌单的全部歌曲的新顺序。与服务端现有歌曲不一致时返回 409 `conflict`，说明客户端的副本已过期，应刷新后重试 |

歌单为 `{id, name, description, cover, trackCount, durationMs, createdAt, updatedAt}`。同一首歌在一个歌单中只出现一次。`cover` 取歌单中第一首有封面的歌曲的封面，都没有时为 `null`；歌单暂不支持自定义封面。
