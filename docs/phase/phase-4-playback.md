# 阶段 4：播放端认证、API 与播放闭环

状态：进行中。4.1 的服务端部分与 4.2 的播放端 API（曲库、喜爱、歌单、历史）已实现（2026-10-24）；4.3 的服务端补充已实现（2026-10-25）；Playback 与播放端接入未开始。前置：阶段 1 的曲库与上传；阶段 3 的播放器展示层。参照路线图：[播放状态与异步规则](../roadmap.md#52-播放状态与异步规则)。

## 4.1 OIDC 认证

已确定（2026-10-23）：播放端使用 OIDC 登录 Authelia，代码完全照搬 `../homepage`（参考提交 `f188285`），不做 forward auth。

- [x] 照搬 `../homepage/internal/auth`（`auth.go`、`provider.go` 及测试，含 `authelia` 构建标签下的 testcontainers 契约测试与 `testdata/authelia`），以及 `internal/store` 中的用户与会话部分；只替换包路径、Cookie 名称和配置键名，不改认证逻辑。证据：[backend/internal/auth/](../../backend/internal/auth/)、[users.go](../../backend/internal/store/users.go)、迁移 4；源码确认。与 homepage 的差异：Cookie 为 `sonora_session` / `sonora_login`；错误页标题为 Sonora；`UpsertUser` 去掉 homepage 特有的“默认分组”插入；`hashToken` 复用管理会话已有的同名函数。契约测试 fixture 的主机名、client id 与测试口令改为 sonora，相应的 pbkdf2 哈希重新生成（先用同一算法复算 homepage 原哈希确认格式一致）。
- [x] 行为与 homepage 一致：`GET /auth/login`、`GET /auth/callback`、`POST /auth/logout`；授权码流程加 PKCE（S256），scope 为 `openid profile email`；首次使用时才发现 issuer，Authelia 不可用时服务仍能启动；会话长期有效，直到手动退出；`Require` 中间件校验同源与会话。证据：`task test:contract` 通过（Authelia 4.39.20，2026-10-24：两个用户登录、重复登录映射到同一用户、退出只影响本浏览器、伪造 state 被拒）；`TestLoginReportsUnavailableProvider`、`TestCallbackWithoutLoginCookieIsRejected`。
- [x] 配置：`public_url`、`oidc.issuer`、`oidc.client_id`、`oidc.client_secret`（建议用环境变量 `SONORA_OIDC_CLIENT_SECRET` 提供）；回调地址 `<public_url>/auth/callback`。非开发模式缺任一项拒绝启动。证据：`TestLoad`、`TestClientSecretFromEnvironment`、`TestLoadRejectsInvalid`；[部署说明](../deployment.md#authelia-客户端)。
- [x] 开发模式使用写死的普通用户，跳过 OIDC（对应 homepage 的 `dev_user`）。证据：`TestDevUserSignsEveryRequestIn`；运行确认：真实曲库副本上 `/api/v1/me` 返回 `listener`，`/auth/login` 跳回 `/`。
- [x] 契约测试照搬 homepage 的 Taskfile 任务，不包含在 `task test` 中，依赖本地已有镜像。证据：`task test:contract`；`task lint:go` 同时 vet `authelia` 标签。
- [ ] 播放端：未登录时跳转 `/auth/login`；用户入口显示用户名，退出调用 `/auth/logout`。服务端已就绪：未登录返回 401，`GET /api/v1/me` 返回用户名；播放端 Vite 开发服务器已代理 `/api` 与 `/auth`。

## 4.2 播放端 API

- [x] 歌曲、专辑、艺术家的列表与详情，分页；搜索（各列表的 `q`）；封面；歌词。数据形态沿用管理端的 `artists` 数组、`cover` 对象和缩略图尺寸，去掉管理字段；歌曲的 `cover` 已解析为实际显示的封面。证据：[player.go](../../backend/internal/server/player.go)、`TestPlayerCatalog`、`TestPlayerAPIRequiresListenerSession`（管理会话与播放端会话互不通用）；[API 约定](../api.md#播放端-api)。运行确认（2026-10-24）：开发模式实例加载 `backend/.data` 的副本（310 张专辑、194 位艺术家），迁移 4 应用成功，列表、详情、歌词、封面正常。
- [x] 歌曲列表排序：`sort` 为默认、标题、歌手、专辑、时长、添加时间，`order` 为升序或降序；适用于全部歌曲、我喜欢、歌单、艺术家的歌曲与搜索结果（2026-10-24 确定，见[阶段 2](phase-2-player-baseline.md#歌曲列表)）。证据：[sort.go](../../backend/internal/store/sort.go)、`TestSortedTrackLists`、`TestPlayerCatalog`（非法取值 400）；[API 约定](../api.md#通用约定)。字段名沿用 SPlayer 的 `sortFieldOptions`（`title`、`artist`、`album`、`duration`、`createTime`）。名称比较用注册到 SQLite 的排序规则 `zh_numeric`（`golang.org/x/text/collate`，简体中文加数字按数值），对应 SPlayer 的 `Intl.Collator("zh-CN", {numeric: true})`；相等时保持列表自身的顺序，与 SPlayer 的稳定排序一致。专辑与艺术家列表的默认顺序也改用这一规则（原为 `NOCASE` 的码点顺序）。运行确认（2026-10-25）：真实曲库副本上标题升序为 `Angel … Young For You, 爱的四季, 爱人错过, 暗涌, 八月照相馆`，337 首按歌手排序 6 ms。
- [x] 当前用户的喜爱、歌单、播放历史读写；按用户隔离；删除歌曲时级联移除。迁移 5 新增 `likes`、`play_history`、`playlists`、`playlist_tracks`。行为参照 SPlayer：喜爱只针对歌曲，播放端用 ID 集合判断喜爱状态（`isLikeSong`）；历史去重、最近在前、上限 500（[data.ts](../../reference/SPlayer/src/stores/data.ts) 的 `setHistory`）。证据：[listeners.go](../../backend/internal/store/listeners.go)、`TestLikesAreIdempotentAndPerUser`、`TestHistoryMovesReplaysToTopAndIsCapped`、`TestPlaylists`（含另一用户全部操作返回 404、整体重排校验）、`TestDeletingTrackRemovesItFromListenerData`、`TestListenerDataIsPerUser`（HTTP 层两个用户互不可见、跨站写入 403、过期重排 409）；[API 约定](../api.md#播放端用户数据)。运行确认（2026-10-24）：真实曲库数据库副本上迁移 5 应用成功，喜爱、记录播放、建歌单并添加歌曲正常。
- 歌单封面取第一首有封面的歌，与现有实现一致（2026-10-24 Eric 确认）。不做歌单自定义封面，不做收藏专辑或艺术家。
- [x] 音频流支持 HTTP Range。证据：`TestPlayerCatalog` 断言 206 与内容；运行确认真实 MP3 返回 206 `audio/mpeg`。
- [ ] 按阶段 2 的 Chrome 实验结论决定是否需要转码。

## 4.3 服务端补充（2026-10-24 讨论确定）

迁移 6 一次加入本节的三处表结构：`tracks.lyrics_format`、`artists.image`、`user_settings`。运行确认（2026-10-25）：开发模式实例加载 `backend/.data/data` 的副本（`user_version` 4，337 首），迁移 5、6 依次应用成功，已有歌词全部为 `lrc`；以下各项的接口与管理后台界面在该实例上逐一操作通过（Playwright，Chrome 1440×1000），用完删除副本。

- [x] 歌词格式：保存原文与格式 `lrc` / `ttml`（迁移新增格式列，已有歌词为 `lrc`）；管理后台接受 `.ttml` 上传与同名外置文件；播放端歌词接口改为 `{format, lyrics}`。证据：[lyrics.go](../../backend/internal/store/lyrics.go)、`TestLyricsFormat`、`TestLyricsKeepTheirFormat`、`TestPlayerCatalog`。格式按内容判断而不看扩展名，因此管理后台直接粘贴编辑的文本也适用：以 `<` 开头的必须是根元素为 `tt` 的合法 XML，否则 400；其余按 LRC。音频内嵌歌词按 LRC。管理后台：歌曲列表的“歌词”列显示 `LRC` / `TTML`，歌曲抽屉显示格式并接受 `.lrc` / `.ttml`；上传页把同名 `.lrc` 或 `.ttml` 与音频配对，两者都有时用 TTML，LRC 标为未使用并说明原因（运行确认）。
- [x] 艺术家图片：管理后台上传、替换、删除，与封面相同的存储、缩略图与 ThumbHash 处理；管理和播放端的艺术家数据增加 `image: {name, thumbhash}`。证据：`PUT/DELETE /admin/api/v1/artists/{id}/image`、`TestArtistImages`（合并时目标沿用来源图片；艺术家随最后一首歌被清理时图片文件一并删除）、`TestPlayerCatalog`。图片与封面共用 `covers` 目录、`covers` 表和 `covers/{name}` 接口；图片相关的错误提示由“封面”改为“图片”。修改专辑可能清理掉艺术家，因此 `UpdateAlbum` 之后也回收无人引用的图片。管理后台的艺术家页增加圆形缩略图列，“重命名”弹窗改为“编辑艺术家”，含名称与图片（运行确认）。
- [x] 设置同步：按用户保存外观与歌词设置，播放端整体读写（例如 `GET/PUT settings`，服务端只校验大小与 JSON 格式，不解释内容）；设备相关设置不上传。证据：[settings.go](../../backend/internal/store/settings.go)、`TestSettingsArePerUserJSONObjects`、`TestSettingsSync`。`GET settings` 在未保存时返回 `{}`；`PUT settings` 整体替换，只接受不超过 256 KB 的 JSON 对象，后写入的覆盖先写入的。
- [x] 首页数据：最近添加的专辑（专辑列表支持按添加时间倒序）；最近播放沿用 `history`。证据：`GET albums?sort=createTime&order=desc`，`TestSortedTrackLists`。“添加时间”是专辑创建的时间（按 ID），给旧专辑补传歌曲不会让它排到前面。

## 4.4 Playback

- [ ] 状态机 `idle → resolving → loading → playing / paused → ended`，失败进入 `error`；Playback 是播放状态的唯一所有者。
- [ ] 切歌请求标识与取消：先点 A 后点 B，A 较晚返回也不能覆盖 B；解析中暂停不得自动播放。
- [ ] 队列条目独立身份，允许重复入队；播放模式：顺序、列表循环、单曲循环、随机。
- [ ] 点击列表的“播放”：专辑、歌单、我喜欢、最近播放取完整列表入队；全部歌曲与搜索结果只取已加载的部分。
- [ ] 播放超过 min(30 秒, 时长的一半) 才写入最近播放。
- [ ] 播放失败自动跳过，最多连续 3 首，之后停止并提示错误。
- [ ] 歌词、封面、媒体信息和进度跟随当前条目。
- [ ] 队列与播放位置保存在浏览器中，重启后恢复为可继续播放的状态，不自动播放。
- [ ] 单元测试覆盖切歌竞态、解析中暂停、队列编辑与恢复。

## 退出条件

- 管理后台上传 → OIDC 登录播放端 → 曲库列表 → 入队 → 播放/暂停/拖动/切歌 → 歌词显示 → 重启恢复，整条链路在 Chrome 中真实可听。
- 两个用户的喜爱、歌单与历史互不可见；Authelia 契约测试通过。
- 快速切歌、播放中暂停、队列编辑和重启不串状态、不丢数据；播放器相关页面完成视觉对照。
