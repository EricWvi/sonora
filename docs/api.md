# Sonora API 约定

更新日期：2026-11-05。本文记录已实现的管理 API（阶段 1）、播放端登录与播放端 API（阶段 4）及其通用约定。实现见 [server.go](../backend/internal/server/server.go)、[player.go](../backend/internal/server/player.go)、[admin.go](../backend/internal/admin/admin.go) 与 [auth.go](../backend/internal/auth/auth.go)。

## 通用约定

- **版本**：版本号写在路径中。管理 API 前缀为 `/admin/api/v1/`，播放端 API 前缀为 `/api/v1/`。不兼容的修改改用 `v2`，同一版本内只做向后兼容的增加。服务版本见 `GET /api/version`。
- **错误**：所有错误响应的正文为 `{"error": "给人看的说明", "code": "machine_code"}`。客户端先按 HTTP 状态分支，`code` 区分同一状态下的不同情况。照搬自 homepage 的播放端认证（`/auth/` 及播放端 API 的 401/403）只返回 `error`，客户端不能依赖每个错误都带 `code`。
- **身份**：媒体以整数 ID 标识。ID 由 `AUTOINCREMENT` 分配，删除后不复用。文件路径、存储布局和内容哈希不出现在任何响应中；封面以内容哈希文件名引用，只能经 `covers/{name}` 读取。
- **分页**：列表接受 `offset`、`limit`（默认 50，上限 500），返回 `{"items": [...], "total": n}`。
- **排序**：歌曲列表接受 `sort` 与 `order`。`sort` 取 `default`（默认）、`title`、`artist`、`album`、`duration`、`createTime`，与 SPlayer 的 `sortFieldOptions` 同名；`order` 取 `asc`（默认）或 `desc`。`default` 是列表自身的顺序，此时忽略 `order`；按其他字段排序时，相等的项保持列表自身的顺序。`artist` 按第一位艺术家，没有艺术家或专辑的歌曲视为空名。名称按拼音、数字按数值比较（对应 SPlayer 的 `Intl.Collator("zh-CN", {numeric: true})`），专辑与艺术家列表的默认顺序也使用同一规则。专辑列表与管理端的艺术家列表只接受 `default`（按标题或名称）与 `createTime`（2026-11-06 起艺术家列表可排序）。其他取值返回 400 `invalid`。
- **搜索**：`q` 去掉首尾空白后按子串匹配（ASCII 不区分大小写，`%`、`_` 按字面匹配）。`sort` 为 `default` 时结果按相关度排列（2026-10-27）：先比较主要字段（歌曲与专辑的标题、艺术家与歌单的名称），再比较次要字段（歌曲的专辑名与艺术家名、专辑的艺术家名）；同一字段内与 `q` 完全相同的在前，其次是以 `q` 开头的，最后是只包含 `q` 的；同一档内保持列表自身的顺序。指定其他 `sort` 时按该字段排序。
- **写请求**：非 GET 请求若带 `Origin` 头，必须等于配置的 `public_url`，否则返回 403 `bad_origin`。

| 状态 | code                      | 场景                                                             |
| ---- | ------------------------- | ---------------------------------------------------------------- |
| 400  | `bad_request` / `invalid` | 请求格式错误 / 输入不合法（说明可直接展示）                      |
| 400  | `netease_rejected`        | 网易云拒绝了写操作，`message` 是网易云的说明                     |
| 401  | `unauthenticated`         | 没有管理会话或会话已过期（播放端为不带 `code` 的 401，见下文）   |
| 401  | `invalid_credentials`     | 用户名或密码错误                                                 |
| 403  | `not_initialized`         | 尚未创建管理员时访问其他管理接口                                 |
| 403  | `invalid_token`           | 初始化 token 缺失或错误                                          |
| 403  | `bad_origin`              | 跨站写请求                                                       |
| 403  | `import_forbidden`        | 不在可以导入歌曲的组内                                           |
| 404  | `not_found`               | 内容不存在或已删除                                               |
| 409  | `already_initialized`     | 重复初始化                                                       |
| 409  | `conflict`                | 重命名与现有专辑或艺术家冲突，应改用合并                         |
| 413  | `too_large`               | 音频超过 1 GB、图片超过 10 MB、歌词超过 512 KB 或设置超过 256 KB |
| 415  | `unsupported_format`      | 不是 MP3、FLAC、AAC（M4A）、Ogg Vorbis 或 Opus                   |
| 422  | `corrupt_file`            | 格式可识别但结构损坏或不完整                                     |
| 429  | `netease_throttled`       | 网易云认为请求过于频繁（网易云的 405 或 -460），稍后再试         |
| 502  | `netease_unavailable`     | 网易云接口服务不可达或出错                                       |
| 503  | `netease_disabled`        | 服务器没有配置网易云接口服务                                     |
| 429  | `rate_limited`            | 初始化或登录失败次数过多，带 `Retry-After`                       |

## 管理 API

会话 Cookie 为 `sonora_admin_session`，`Path=/admin/api/v1/`、`HttpOnly`、`SameSite=Strict`，`public_url` 为 https 时带 `Secure`；与播放端会话互不通用。会话闲置 7 天或登录满 30 天后失效。

无需会话：

| 方法与路径    | 说明                                                                 |
| ------------- | -------------------------------------------------------------------- |
| `GET status`  | `{initialized, devMode, admin}`；`admin` 为当前会话的管理员或 `null` |
| `POST setup`  | `{token, username, password}`，成功后返回状态并设置会话              |
| `POST login`  | `{username, password}`，用户名不区分大小写                           |
| `POST logout` | 结束当前会话，204                                                    |

需要会话：

| 方法与路径                                 | 说明                                                                                                                                                                                                                                                                                                                                                           |
| ------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET overview`                             | 歌曲、专辑、艺术家数量，缺标签数量，总时长，音频与封面占用字节                                                                                                                                                                                                                                                                                                 |
| `GET tracks`                               | `q`、`missing=true`、`albumId`、`artistId`、`sort`、`order`；默认按专辑时按碟号与音轨号排序，否则新入库在前                                                                                                                                                                                                                                                    |
| `POST tracks`                              | `multipart/form-data`：一个 `file` 部分，可选 `lyrics` 部分（外置 LRC 或 TTML 歌词文件）。新歌曲 201，重复文件 200，正文为 `{status: "created" \| "duplicate", track}`                                                                                                                                                                                         |
| `GET/PUT/DELETE tracks/{id}`               | 详情（含歌词）/ 修改元数据 / 删除歌曲及其音频文件                                                                                                                                                                                                                                                                                                              |
| `PUT/DELETE tracks/{id}/cover`             | 请求体为图片原始字节；单曲封面优先于专辑封面                                                                                                                                                                                                                                                                                                                   |
| `PUT/DELETE tracks/{id}/lyrics`            | 请求体为 LRC 或 TTML 歌词的原始字节，服务识别 UTF-8、UTF-16 与 GB18030                                                                                                                                                                                                                                                                                         |
| `GET tracks/{id}/audio`                    | 音频流，支持 Range，用于试听                                                                                                                                                                                                                                                                                                                                   |
| `GET albums`、`GET/PUT/DELETE albums/{id}` | 列表（`q`、`artistId`、`sort`、`order`）/ 详情 / 修改 `{title, artists, year, single, description, neteaseId}` / 删除专辑及其全部歌曲。`single` 标记只含单曲的专辑（单曲自动建的专辑，或只是同名的几首单曲），省略时保持原值；新入库的专辑为 `false`。`description` 为专辑简介（2026-11-04 起），去掉首尾空白后最多 10000 个字符，省略时保持原值，空字符串清除 |
| `POST albums/{id}/merge`                   | `{sourceIds, keepCovers}` 合并到路径中的专辑。目标没有封面、年份或网易云 ID 时，沿用第一个有这些信息的来源专辑。`keepCovers` 默认 `false`，合并后只留目标的封面；为 `true` 时，来源专辑中没有单曲封面的歌曲以原专辑的封面作为单曲封面（与目标封面相同或原专辑没有封面时不设）                                                                                  |
| `PUT/DELETE albums/{id}/cover`             | 同单曲封面                                                                                                                                                                                                                                                                                                                                                     |
| `GET artists`、`GET/PUT artists/{id}`      | 列表（`q`、`sort`、`order`）/ 详情 / 修改 `{name, description, neteaseId}`。`description` 为歌手简介（2026-11-04 起），去掉首尾空白后最多 10000 个字符，省略时保持原值，空字符串清除                                                                                                                                                                           |
| `POST artists/{id}/merge`                  | `{sourceIds}`；合并后艺术家组合相同的同名专辑一并合并；目标没有图片、简介或网易云 ID 时沿用来源的                                                                                                                                                                                                                                                              |
| `PUT/DELETE artists/{id}/image`            | 艺术家图片，请求体为图片原始字节，存储与缩略图同封面                                                                                                                                                                                                                                                                                                           |
| `GET covers/{name}`                        | 封面或艺术家图片的原图，按内容哈希命名，可长期缓存。`?size=` 取 `96`、`100`、`300` 或 `1024`，返回长边不超过该值的 JPEG 缩略图，首次请求时生成；原图已不大于该尺寸或无法解码时返回原图；其他尺寸 400                                                                                                                                                           |

歌词有两种格式，由内容判断，与文件扩展名无关：以 `<` 开头的是 TTML，必须是根元素为 `tt` 的合法 XML，否则返回 400；其余都按 LRC（含逐字与增强型 LRC，由播放端区分）。音频内嵌的歌词按 LRC 处理。歌曲的 `lyricsFormat` 为 `lrc` 或 `ttml`，没有歌词时为空字符串。

歌曲修改 `PUT tracks/{id}` 的正文为 `{title, artists, album, albumArtists, trackNo, discNo, year, neteaseId}`，`artists` 与 `albumArtists` 是有序的名称数组。艺术家与专辑按名称查找，不存在时创建；名称去掉首尾空白，空名与重复名（不区分大小写）忽略，最多 50 位；`albumArtists` 为空时使用 `artists`。没有歌曲的专辑、既无歌曲也无专辑的艺术家自动删除，因此不提供单独删除艺术家的接口。

歌曲、专辑及歌曲所属专辑的封面为 `cover: {name, thumbhash}`，没有封面时为 `null`；艺术家的图片 `image` 形态相同。`thumbhash` 是 [ThumbHash](https://evanw.github.io/thumbhash/) 的标准 base64，客户端解码为模糊占位图，在缩略图加载完成前显示；图片无法解码时为空字符串。

歌曲、专辑与艺术家带 `neteaseId`，是从网易云导入时记下的歌曲、专辑与艺术家 ID，其余为 `null`；只在管理端返回，播放端接口不含此字段。三者的修改接口都接受 `neteaseId`（2026-11-06 起）：正整数设为该值，`null` 清除，省略时保持原值，其他取值返回 400 `invalid`；歌曲的 `neteaseId` 不能与另一首歌曲相同，否则返回 409，导入时据此判断歌曲是否已在曲库中，专辑与艺术家的可以重复。专辑与艺术家的 `neteaseId` 从 2026-11-06 起记录，此前导入的为 `null`。导入时艺术家按名称匹配已有艺术家，已有 `neteaseId` 的保持不变，所以同名的两位网易云艺术家共用第一次导入时的 ID；每次导入都新建专辑，同一张网易云专辑的歌曲各自的专辑 `neteaseId` 相同。

歌曲、专辑响应中的艺术家是有序数组 `artists: [{id, name}]`，歌曲所属专辑为 `album: {id, title, artists, cover}`；没有艺术家时为空数组。按 `artistId` 筛选歌曲时，只包括该艺术家作为歌曲艺术家参与的歌曲，与艺术家的 `trackCount` 一致（2026-11-04 起；原来还包括其专辑中的歌曲）。

## 播放端登录

照搬 `../homepage` 的 OIDC 实现（参考提交 `f188285`），只改了包路径、Cookie 名称、配置键名和页面标题。

| 方法与路径           | 说明                                                                                                     |
| -------------------- | -------------------------------------------------------------------------------------------------------- |
| `GET /auth/login`    | 跳转到 Authelia；授权码流程加 PKCE（S256），scope `openid profile email`。Authelia 不可达时返回 503 页面 |
| `GET /auth/callback` | 校验 state、ID token（签名、issuer、audience、有效期、nonce）后建立会话，跳回 `/`                        |
| `POST /auth/logout`  | 只结束本浏览器的会话，不退出 Authelia，204                                                               |

会话 Cookie 为 `sonora_session`，`Path=/`、`HttpOnly`、`SameSite=Lax`，https 时带 `Secure`。服务端会话不过期，直到退出；使用中每天续发一次 Cookie，避免浏览器 400 天上限。用户按 `(issuer, subject)` 识别，名称与邮箱每次登录时刷新。开发模式不走 OIDC，所有请求都以写死的用户 `listener` 登录，`/auth/login` 直接跳回 `/`。

未登录访问播放端 API 返回 401 `{"error": "未登录"}`，播放端应跳转 `/auth/login`；跨站写请求返回 403。Go 服务的错误响应都带 `{"error"}` 正文，播放端把没有这个正文的 5xx（来自服务前面的代理）视为服务不可达。管理会话与播放端会话互不通用。

## 播放端 API

全部需要播放端会话。播放端看不到只与管理有关的字段：原文件名、文件大小、入库时间、歌词正文（另有接口）都不出现在列表与详情中。列表的分页、`q` 搜索和[排序](#通用约定)与管理 API 相同：歌曲默认新入库在前，按专辑筛选时按碟号与音轨号；专辑默认按标题；艺术家按名称。

| 方法与路径                        | 说明                                                                                                                                                                                                                                                                                      |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET me`                          | 当前用户 `{name, email, canImport}`；`canImport` 表示可以导入歌曲                                                                                                                                                                                                                         |
| `GET tracks`                      | `q`（匹配标题、专辑名、艺术家名）、`albumId`、`artistId`、`sort`、`order`；默认新入库在前，首页“最近添加的歌曲”取前 6 首                                                                                                                                                                  |
| `GET tracks/{id}`                 | 歌曲详情                                                                                                                                                                                                                                                                                  |
| `GET tracks/{id}/lyrics`          | `{format, lyrics}`：`format` 为 `lrc` 或 `ttml`，`lyrics` 为原文；没有歌词时都是空字符串。解析由播放端完成                                                                                                                                                                                |
| `GET tracks/{id}/audio`           | 音频流，支持 Range                                                                                                                                                                                                                                                                        |
| `GET albums`、`GET albums/{id}`   | 列表（`q`、`artistId`、`sort`、`order`、`hideSingles`）/ 详情；`hideSingles=true` 时不含标记为单曲的专辑，专辑页与首页使用；首页“最近添加的专辑”用 `sort=createTime&order=desc&hideSingles=true`；专辑另有 `onlineId`：关联的在线专辑（导入或管理后台填写的网易云 ID），没有时为 `null`。 |
| `GET artists`、`GET artists/{id}` | 列表（`q`）/ 详情；艺术家的歌曲与专辑用 `tracks?artistId=`、`albums?artistId=` 获取；艺术家另有 `onlineId`，含义同专辑。                                                                                                                                                                  |
| `GET covers/{name}`               | 封面与艺术家图片，与管理 API 相同，支持同样的 `?size=`。播放端按 SPlayer 的封面尺寸取 `100`（小图：底部栏、背景与取色）、`300`、`1024`（全屏播放器的大封面），2026-10-25 增加 `100` 与 `1024`                                                                                             |

数据形态：

- 歌曲 `{id, title, artists, album, trackNo, discNo, year, durationMs, format, cover, hasLyrics}`。`album` 为 `{id, title, artists}` 或 `null`；`cover` 是歌曲实际显示的封面，单曲封面优先，否则为专辑封面，都没有时为 `null`。
- 专辑 `{id, title, artists, year, cover, description, trackCount, durationMs}`。`description` 为管理后台填写的简介，没有时为空字符串。
- 艺术家 `{id, name, image, description, trackCount, albumCount}`，`image` 是管理后台上传的图片，没有时为 `null`；`description` 为管理后台填写的简介，没有时为空字符串。
- `artists` 与 `cover` 的形态和管理 API 一致：`[{id, name}]` 与 `{name, thumbhash}`。

## 网易云账号

网易云账号与播放端登录相互独立，按当前登录的用户隔离，每个用户绑定各自的账号。Cookie 只保存在服务端，任何响应都不含 Cookie。Go 服务只开放下面这些具名接口，不做通用转发，因为接口服务的响应会回传 Cookie。

| 方法与路径                   | 说明                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ---------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET netease/status`         | `{state, profile}`。`state` 为 `disabled`（服务器没有配置）、`signedOut`（没有登录）、`signedIn`、`expired`（登录已失效，重新登录即可）或 `unavailable`（已登录但接口服务不可达）。只有 `signedIn` 时 `profile` 为 `{userId, nickname, avatarUrl, vipType}`，否则为 `null`；`vipType` 为 0 表示不是会员。会话保存超过 3 天时顺带续期，与 SPlayer 相同另有 `links`：网易云网页上 `song`、`album`、`artist`、`playlist` 的地址模板，`{id}` 换成 ID，用于分享链接与打开源页面（播放端源码不写网易云的地址）；`disabled` 时没有 |
| `POST netease/login/qr`      | 开始二维码登录，返回 `{key, url}`，`url` 是二维码的内容                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `GET netease/login/qr/{key}` | 查询进度，返回 `{code, nickname, avatarUrl}`；`code` 为 800（已过期）、801（等待扫码）、802（待确认，带昵称与头像）或 803（成功，服务端已保存会话）。播放端每秒查询一次                                                                                                                                                                                                                                                                                                                                                     |
| `POST netease/login/cookie`  | `{cookie}`，服务端校验后保存，返回与 `status` 相同的形态；`cookie` 可以是从浏览器复制的整段请求头。无效时返回 400 `invalid`                                                                                                                                                                                                                                                                                                                                                                                                 |
| `POST netease/logout`        | 只删除 Sonora 保存的会话，幂等，204。与 SPlayer 不同，不调用接口服务的 `/logout`，以免让同一个 Cookie 在别的客户端失效                                                                                                                                                                                                                                                                                                                                                                                                      |

### 网易云内容

下面的接口需要已登录网易云：没有登录时返回 403 `netease_signed_out`，登录失效时返回 403 `netease_expired`。用 403 而不是 401，因为失效的只是网易云，401 表示播放端会话失效。

在线内容的 JSON 形态与本地相同，列表可以混用，另有 `source: "online"`；本地内容没有 `source`。封面为 `{name: "", thumbhash: "", url}`，`url` 是网易云的图片地址，由浏览器直接加载。在线歌曲另有 `playable`，`false` 表示网易云不提供这首歌；没有 `trackNo`、`discNo`，`year` 取发行时间。本地与在线的 ID 都是整数，可能相同，比较时要同时比较来源。

| 方法与路径                                                                                                                   | 说明                                                                                                                                                                                                                                                                                                                                                                                                                |
| ---------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET netease/playlists`                                                                                                      | `{created, collected}`：自己创建的与收藏的歌单，不含“我喜欢的音乐”                                                                                                                                                                                                                                                                                                                                                  |
| `GET netease/playlists/{id}`、`netease/albums/{id}`、`netease/artists/{id}`                                                  | `{playlist, tracks}`：详情都以歌单的形态返回，全部歌曲一次返回；艺术家为热门歌曲；不可播放的歌 `playable` 为 `false`                                                                                                                                                                                                                                                                                                |
| `GET netease/albums`、`GET netease/artists`                                                                                  | 收藏的专辑与歌手，`{items, total}`                                                                                                                                                                                                                                                                                                                                                                                  |
| `GET netease/artists/{id}/albums`                                                                                            | 艺术家的专辑，分页，`{items, total}`                                                                                                                                                                                                                                                                                                                                                                                |
| `PUT`、`DELETE netease/playlists/{id}/subscription`、`netease/albums/{id}/subscription`、`netease/artists/{id}/subscription` | 在网易云上收藏 / 取消收藏歌单（别人的歌单）与专辑、关注 / 取消关注艺术家，204；被网易云拒绝时 400 `netease_rejected`。是否已收藏从 `GET netease/playlists` 的 `collected`、`GET netease/albums`、`GET netease/artists` 判断                                                                                                                                                                                         |
| `POST netease/playlists`                                                                                                     | `{name, private}` 在网易云账号上新建歌单，返回歌单                                                                                                                                                                                                                                                                                                                                                                  |
| `PUT netease/playlists/{id}`                                                                                                 | `{name, description}` 只修改有变化的字段，不影响标签，204；不是该账号自己创建的歌单返回 404                                                                                                                                                                                                                                                                                                                         |
| `DELETE netease/playlists/{id}`                                                                                              | 删除，204；不是该账号自己创建的歌单（包括“我喜欢”与收藏的歌单）返回 404                                                                                                                                                                                                                                                                                                                                             |
| `POST netease/playlists/{id}/tracks`                                                                                         | `{ids}` 把在线歌曲加入该账号自己创建的歌单，204；已在歌单中等被网易云拒绝时返回 400 `netease_rejected`                                                                                                                                                                                                                                                                                                              |
| `GET netease/recommend/songs`                                                                                                | 每日推荐，`{items, total}`，全部歌曲一次返回；网易云每天 6:00 更新，服务端不缓存；不可播放的歌 `playable` 为 `false`                                                                                                                                                                                                                                                                                                |
| `GET netease/songs/{id}/url`                                                                                                 | `{url}`：音质 `exhigh`，地址带时效，每次播放时获取；只有试听或没有地址时返回 404 `netease_no_url`                                                                                                                                                                                                                                                                                                                   |
| `GET netease/songs/{id}/lyrics`                                                                                              | 与 `tracks/{id}/lyrics` 相同的 `{format, lyrics}`：有逐字歌词时为 TTML，否则 LRC，翻译与罗马音有就放入                                                                                                                                                                                                                                                                                                              |
| `POST netease/songs/{id}/import`                                                                                             | 导入曲库，需要在 `netease.import_group` 指定的组内，否则 403 `import_forbidden`。由服务端下载，返回 `{status, track}`，`status` 为 `created` 或 `duplicate`（已导入过，不再下载）；账号没有完整版本时 404 `netease_no_url`，不是 MP3 时 415 `unsupported_format`。导入新建的艺术家（专辑导入还包括专辑艺术家）同时写入网易云的图片与简介                                                                            |
| `GET netease/albums/{id}/import`                                                                                             | 导入专辑前的预览（组外 403 `import_forbidden`）：`{title, total, toImport, inLibrary, moving, unplayable, target, sameTitle}`。`toImport` 是可以播放、曲库中还没有的歌；`inLibrary` 是已导入过的，其中 `moving` 首还独占单曲导入时建的专辑，导入时会移入这张专辑；`target` 是歌曲并入的曲库专辑（关联了这张网易云专辑、且不是单曲导入建的），为 `null` 时新建；`sameTitle` 是其他同名专辑的数量，导入后不会自动合并 |
| `POST netease/albums/{id}/import`                                                                                            | 开始导入专辑，202，返回任务（形态同 `imports`）；同一用户同一张专辑已在排队或进行中时返回那个任务。服务端在后台一次一张、逐首下载，每首单独入库；向网易云取歌的间隔至少 5 秒；服务重启后任务丢失，已入库的歌保留，再导入一次即可补齐                                                                                                                                                                                |
| `GET netease/imports`                                                                                                        | 当前用户的专辑导入任务，新的在前，`{items}`；每项 `{id, albumId, title, state, total, done, created, moved, existing, unplayable, failed, error}`。`state` 为 `queued`、`running`、`done` 或 `failed`（登录失效、退出网易云、网易云限制请求频率或接口服务不可用时中止，`error` 说明原因）；`failed` 是没有导入的歌 `[{title, reason}]`（账号没有完整版本、不是 MP3、下载失败），这些歌不影响其余的歌                |
| `GET netease/likes/ids`                                                                                                      | `{ids}`：喜欢的在线歌曲                                                                                                                                                                                                                                                                                                                                                                                             |
| `PUT netease/likes/{id}`、`DELETE netease/likes/{id}`                                                                        | 在网易云上喜欢 / 取消，204                                                                                                                                                                                                                                                                                                                                                                                          |

### 混合列表

本地与在线歌曲混合的列表，分页与其他列表相同，不接受 `sort`（在线歌曲不在数据库中，排序由播放端完成）。没有登录网易云时只有本地歌曲。

| 方法与路径                   | 说明                                                                                                                                                             |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET likes/mixed`            | 按喜欢时间混排。同一首歌本地与在线都喜欢时只有一条，优先本地（按 `tracks.netease_id` 判断），时间取较晚的。网易云读取失败时返回本地歌曲并带 `onlineFailed: true` |
| `GET history/mixed`          | `{items: [{track, playedAt}], total}`，按时间混排；本地与在线共用 500 条的上限                                                                                   |
| `POST history/online`        | `{song}` 记录一次在线歌曲的播放，服务端保存歌曲的快照（JSON 不超过 16 KB），204                                                                                  |
| `DELETE history/online/{id}` | 删除一条，204；`DELETE history` 同时清空本地与在线的记录                                                                                                         |

## 播放端用户数据

设置、喜爱、歌单与播放历史属于当前登录的用户，同样需要播放端会话。别人的歌单一律返回 404，不区分“不存在”与“无权访问”。管理端删除歌曲时，它会从所有人的喜爱、歌单与历史中消失。

| 方法与路径                                      | 说明                                                                                                                                                                                                                                                                                                    |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET settings`                                  | 当前用户保存的设置，原样返回播放端保存的 JSON 对象；从未保存时为 `{}`                                                                                                                                                                                                                                   |
| `PUT settings`                                  | 请求体为完整的设置对象，整体替换，204。服务端只校验是 JSON 对象且不超过 256 KB，不解释内容。播放端保存 `{setting, shortcuts}`：setting store 中除输出设备、同步补偿与侧栏展开状态外的全部字段，以及快捷键的组合键；均衡器与音量不上传（规则见[阶段 5](phase/phase-5-player-pages.md#已确定2026-10-25)） |
| `GET likes`                                     | 喜爱的歌曲，默认最近喜爱的在前，分页，可排序                                                                                                                                                                                                                                                            |
| `GET likes/ids`                                 | `{ids}`：全部喜爱歌曲的 ID，供播放端在任意列表中标记喜爱状态                                                                                                                                                                                                                                            |
| `PUT likes/{trackId}`、`DELETE likes/{trackId}` | 喜爱 / 取消，幂等，204；歌曲不存在时喜爱返回 404                                                                                                                                                                                                                                                        |
| `GET history`                                   | `{items: [{track, playedAt}], total}`，最近播放的在前；同一首歌只保留一条，重播移到最前；每人最多保留 500 首，与 SPlayer 一致                                                                                                                                                                           |
| `POST history`                                  | `{trackId}` 记录一次播放，204。什么时候算一次播放由播放端决定                                                                                                                                                                                                                                           |
| `DELETE history`、`DELETE history/{trackId}`    | 清空 / 删除一条，204                                                                                                                                                                                                                                                                                    |
| `GET playlists`                                 | 自己的歌单，新建的在前；`q` 匹配歌单名，按相关度排列                                                                                                                                                                                                                                                    |
| `POST playlists`                                | `{name, description}` 新建，201。名称 1–100 个字符，简介最多 1000 个字符                                                                                                                                                                                                                                |
| `GET/PUT/DELETE playlists/{id}`                 | 详情 / 修改 `{name, description}` / 删除                                                                                                                                                                                                                                                                |
| `GET playlists/{id}/tracks`                     | 歌单中的歌曲，默认按歌单顺序，分页，可排序                                                                                                                                                                                                                                                              |
| `POST playlists/{id}/tracks`                    | `{trackIds}` 按给定顺序插到歌单开头（2026-11-04 起，原为追加到末尾），已在歌单中的跳过；返回 `{added, playlist}`。任一歌曲不存在时整体失败（404）；每个歌单最多 10000 首                                                                                                                                |
| `DELETE playlists/{id}/tracks/{trackId}`        | 移出歌单，返回歌单                                                                                                                                                                                                                                                                                      |
| `PUT playlists/{id}/order`                      | `{trackIds}` 为歌单的全部歌曲的新顺序。与服务端现有歌曲不一致时返回 409 `conflict`，说明客户端的副本已过期，应刷新后重试                                                                                                                                                                                |

歌单为 `{id, name, description, cover, trackCount, durationMs, createdAt, updatedAt}`。同一首歌在一个歌单中只出现一次。`cover` 取歌单中第一首有封面的歌曲的封面，都没有时为 `null`；歌单暂不支持自定义封面。
