# 阶段 4：播放端认证、API 与播放闭环

状态：已完成（2026-10-25，Eric 确认）。输出设备切换转入阶段 5，真实 Authelia 下的浏览器登录转入阶段 6，见“未完成或需要 Eric 确认”。4.1 已完成（播放端的未登录跳转于 2026-10-25 运行确认）；4.2 的播放端 API（曲库、喜爱、歌单、历史）已实现（2026-10-24）；4.3 的服务端补充已实现（2026-10-25）；4.4 Playback 已实现，单元测试与无头 Chromium 的真实播放链路通过（2026-10-25），退出条件中需要 Eric 在桌面 Chrome 上手动确认的部分见下文。前置：阶段 1 的曲库与上传；阶段 3 的播放器展示层。参照路线图：[播放状态与异步规则](../roadmap.md#52-播放状态与异步规则)。

## 4.1 OIDC 认证

已确定（2026-10-23）：播放端使用 OIDC 登录 Authelia，代码完全照搬 `../homepage`（参考提交 `f188285`），不做 forward auth。

- [x] 照搬 `../homepage/internal/auth`（`auth.go`、`provider.go` 及测试，含 `authelia` 构建标签下的 testcontainers 契约测试与 `testdata/authelia`），以及 `internal/store` 中的用户与会话部分；只替换包路径、Cookie 名称和配置键名，不改认证逻辑。证据：[backend/internal/auth/](../../backend/internal/auth/)、[users.go](../../backend/internal/store/users.go)、迁移 4；源码确认。与 homepage 的差异：Cookie 为 `sonora_session` / `sonora_login`；错误页标题为 Sonora；`UpsertUser` 去掉 homepage 特有的“默认分组”插入；`hashToken` 复用管理会话已有的同名函数。契约测试 fixture 的主机名、client id 与测试口令改为 sonora，相应的 pbkdf2 哈希重新生成（先用同一算法复算 homepage 原哈希确认格式一致）。
- [x] 行为与 homepage 一致：`GET /auth/login`、`GET /auth/callback`、`POST /auth/logout`；授权码流程加 PKCE（S256），scope 为 `openid profile email`；首次使用时才发现 issuer，Authelia 不可用时服务仍能启动；会话长期有效，直到手动退出；`Require` 中间件校验同源与会话。证据：`task test:contract` 通过（Authelia 4.39.20，2026-10-24：两个用户登录、重复登录映射到同一用户、退出只影响本浏览器、伪造 state 被拒）；`TestLoginReportsUnavailableProvider`、`TestCallbackWithoutLoginCookieIsRejected`。
- [x] 配置：`public_url`、`oidc.issuer`、`oidc.client_id`、`oidc.client_secret`（建议用环境变量 `SONORA_OIDC_CLIENT_SECRET` 提供）；回调地址 `<public_url>/auth/callback`。非开发模式缺任一项拒绝启动。证据：`TestLoad`、`TestClientSecretFromEnvironment`、`TestLoadRejectsInvalid`；[部署说明](../deployment.md#authelia-客户端)。
- [x] 开发模式使用写死的普通用户，跳过 OIDC（对应 homepage 的 `dev_user`）。证据：`TestDevUserSignsEveryRequestIn`；运行确认：真实曲库副本上 `/api/v1/me` 返回 `listener`，`/auth/login` 跳回 `/`。
- [x] 契约测试照搬 homepage 的 Taskfile 任务，不包含在 `task test` 中，依赖本地已有镜像。证据：`task test:contract`；`task lint:go` 同时 vet `authelia` 标签。
- [x] 播放端：未登录时跳转 `/auth/login`；用户入口显示用户名，退出调用 `/auth/logout`。启动时 [user store](../../frontend/src/stores/user.ts) 读取 `GET /api/v1/me`，401 时跳转登录；播放中的 401 由 Playback（单元测试见 4.4）与 `toLikeSong` 同样处理。用户入口与退出页见[阶段 3.1](phase-3-player-shell.md#31-菜单与路由)（开发实例运行确认）。运行确认（2026-10-25）：非开发模式的临时后端（全新数据目录，`oidc.issuer` 指向不可达地址，验证服务照常启动）加代理到它的临时 Vite，Playwright Chromium 153 无头打开播放端（两次），启动时的接口全部 401 后跳转 `/auth/login`，显示“认证服务暂时不可用”页，页面没有报错；用完删除。登录成功后回到 `/` 而不是原地址，与 homepage 一致。真实 Authelia 下的浏览器登录由 Eric 在退出条件中确认，服务端流程由 `task test:contract` 覆盖。

## 4.2 播放端 API

- [x] 歌曲、专辑、艺术家的列表与详情，分页；搜索（各列表的 `q`）；封面；歌词。数据形态沿用管理端的 `artists` 数组、`cover` 对象和缩略图尺寸，去掉管理字段；歌曲的 `cover` 已解析为实际显示的封面。证据：[player.go](../../backend/internal/server/player.go)、`TestPlayerCatalog`、`TestPlayerAPIRequiresListenerSession`（管理会话与播放端会话互不通用）；[API 约定](../api.md#播放端-api)。运行确认（2026-10-24）：开发模式实例加载 `backend/.data` 的副本（310 张专辑、194 位艺术家），迁移 4 应用成功，列表、详情、歌词、封面正常。
- [x] 歌曲列表排序：`sort` 为默认、标题、歌手、专辑、时长、添加时间，`order` 为升序或降序；适用于全部歌曲、我喜欢、歌单、艺术家的歌曲与搜索结果（2026-10-24 确定，见[阶段 2](phase-2-player-baseline.md#歌曲列表)）。证据：[sort.go](../../backend/internal/store/sort.go)、`TestSortedTrackLists`、`TestPlayerCatalog`（非法取值 400）；[API 约定](../api.md#通用约定)。字段名沿用 SPlayer 的 `sortFieldOptions`（`title`、`artist`、`album`、`duration`、`createTime`）。名称比较用注册到 SQLite 的排序规则 `zh_numeric`（`golang.org/x/text/collate`，简体中文加数字按数值），对应 SPlayer 的 `Intl.Collator("zh-CN", {numeric: true})`；相等时保持列表自身的顺序，与 SPlayer 的稳定排序一致。专辑与艺术家列表的默认顺序也改用这一规则（原为 `NOCASE` 的码点顺序）。运行确认（2026-10-25）：真实曲库副本上标题升序为 `Angel … Young For You, 爱的四季, 爱人错过, 暗涌, 八月照相馆`，337 首按歌手排序 6 ms。
- [x] 当前用户的喜爱、歌单、播放历史读写；按用户隔离；删除歌曲时级联移除。迁移 5 新增 `likes`、`play_history`、`playlists`、`playlist_tracks`。行为参照 SPlayer：喜爱只针对歌曲，播放端用 ID 集合判断喜爱状态（`isLikeSong`）；历史去重、最近在前、上限 500（[data.ts](../../reference/SPlayer/src/stores/data.ts) 的 `setHistory`）。证据：[listeners.go](../../backend/internal/store/listeners.go)、`TestLikesAreIdempotentAndPerUser`、`TestHistoryMovesReplaysToTopAndIsCapped`、`TestPlaylists`（含另一用户全部操作返回 404、整体重排校验）、`TestDeletingTrackRemovesItFromListenerData`、`TestListenerDataIsPerUser`（HTTP 层两个用户互不可见、跨站写入 403、过期重排 409）；[API 约定](../api.md#播放端用户数据)。运行确认（2026-10-24）：真实曲库数据库副本上迁移 5 应用成功，喜爱、记录播放、建歌单并添加歌曲正常。
- 歌单封面取第一首有封面的歌，与现有实现一致（2026-10-24 Eric 确认）。不做歌单自定义封面，不做收藏专辑或艺术家。
- [x] 音频流支持 HTTP Range。证据：`TestPlayerCatalog` 断言 206 与内容；运行确认真实 MP3 返回 206 `audio/mpeg`。
- [x] 按阶段 2 的 Chrome 实验结论决定是否需要转码：不做服务端转码（2026-10-25）。五种格式在 Chromium 中都能直接流式播放与跳转；MP3 起播前 Chrome 要读约 2.4 MB，Eric 确认可以接受。见[阶段 2 的音频实验](phase-2-player-baseline.md#03-chrome-音频实验)。

## 4.3 服务端补充（2026-10-24 讨论确定）

迁移 6 一次加入本节的三处表结构：`tracks.lyrics_format`、`artists.image`、`user_settings`。运行确认（2026-10-25）：开发模式实例加载 `backend/.data/data` 的副本（`user_version` 4，337 首），迁移 5、6 依次应用成功，已有歌词全部为 `lrc`；以下各项的接口与管理后台界面在该实例上逐一操作通过（Playwright，Chrome 1440×1000），用完删除副本。

- [x] 歌词格式：保存原文与格式 `lrc` / `ttml`（迁移新增格式列，已有歌词为 `lrc`）；管理后台接受 `.ttml` 上传与同名外置文件；播放端歌词接口改为 `{format, lyrics}`。证据：[lyrics.go](../../backend/internal/store/lyrics.go)、`TestLyricsFormat`、`TestLyricsKeepTheirFormat`、`TestPlayerCatalog`。格式按内容判断而不看扩展名，因此管理后台直接粘贴编辑的文本也适用：以 `<` 开头的必须是根元素为 `tt` 的合法 XML，否则 400；其余按 LRC。音频内嵌歌词按 LRC。管理后台：歌曲列表的“歌词”列显示 `LRC` / `TTML`，歌曲抽屉显示格式并接受 `.lrc` / `.ttml`；上传页把同名 `.lrc` 或 `.ttml` 与音频配对，两者都有时用 TTML，LRC 标为未使用并说明原因（运行确认）。
- [x] 艺术家图片：管理后台上传、替换、删除，与封面相同的存储、缩略图与 ThumbHash 处理；管理和播放端的艺术家数据增加 `image: {name, thumbhash}`。证据：`PUT/DELETE /admin/api/v1/artists/{id}/image`、`TestArtistImages`（合并时目标沿用来源图片；艺术家随最后一首歌被清理时图片文件一并删除）、`TestPlayerCatalog`。图片与封面共用 `covers` 目录、`covers` 表和 `covers/{name}` 接口；图片相关的错误提示由“封面”改为“图片”。修改专辑可能清理掉艺术家，因此 `UpdateAlbum` 之后也回收无人引用的图片。管理后台的艺术家页增加圆形缩略图列，“重命名”弹窗改为“编辑艺术家”，含名称与图片（运行确认）。
- [x] 设置同步：按用户保存外观与歌词设置，播放端整体读写（例如 `GET/PUT settings`，服务端只校验大小与 JSON 格式，不解释内容）；设备相关设置不上传。证据：[settings.go](../../backend/internal/store/settings.go)、`TestSettingsArePerUserJSONObjects`、`TestSettingsSync`。`GET settings` 在未保存时返回 `{}`；`PUT settings` 整体替换，只接受不超过 256 KB 的 JSON 对象，后写入的覆盖先写入的。
- [x] 首页数据：最近添加的专辑（专辑列表支持按添加时间倒序）；最近播放沿用 `history`。证据：`GET albums?sort=createTime&order=desc`，`TestSortedTrackLists`。“添加时间”是专辑创建的时间（按 ID），给旧专辑补传歌曲不会让它排到前面。

## 4.4 Playback

实现见 [core/player/](../../frontend/src/core/player/) 与 [core/audio-player/](../../frontend/src/core/audio-player/)。[PlayerController.ts](../../frontend/src/core/player/PlayerController.ts) 就是路线图中的 Playback：播放状态、队列与播放模式的唯一所有者，store 中的 `playStatus`、`playLoading` 只由它写入；方法名与流程沿用 SPlayer 的 `PlayerController`、`PlayModeManager`，已照搬的组件没有改动（队列抽屉只把列表项的 key 换成条目身份）。外部能力（播放引擎、取地址、数据源、存储、提示、歌词与取色）由构造函数注入，单元测试换成假的实现。

- **音频链路**照搬 SPlayer 的 Web Audio 引擎（`BaseAudioPlayer`、`AudioElementPlayer`、`AudioEffectManager`、`SharedAudioContext`）：`HTMLAudioElement` → `MediaElementSource` → 10 段均衡器 → 分析器 → 增益 → 限制器 → 输出，与阶段 2 的 Chrome 实验相同。不迁入 MPV、FFmpeg、自动混音的滤波与交叉淡入淡出、倍速、ReplayGain。地址用 `DataSource.audioUrl(trackId)`，不做转码。对 SPlayer 的三处修正：页面从未被点击时 `AudioContext` 无法恢复、`resume()` 不会完成，改为按自动播放限制报 `NotAllowedError`；恢复播放时总是调用 `resume()`（SPlayer 只在 `suspended` 时调用，暂停刚请求挂起时会漏掉）；输出设备选回“默认”时设置为默认设备（SPlayer 直接返回，切走后改不回来）。
- fixture（`?fixture=…`）换成不出声的引擎 [SilentAudioPlayer.ts](../../frontend/src/core/audio-player/SilentAudioPlayer.ts)，播放时间停在跳转点，队列只存在内存中；`?player=` 的“播放中”只是显示状态。`task test:visual` 结果不变（见下文）。

已确定（2026-10-25，Eric 确认与 SPlayer 的差异）：

- 队列条目有独立身份，允许重复入队；“下一首播放”不再按歌曲去重，只有正在播放的那首仍直接播放（与 SPlayer 相同）。随机模式下的增删同时作用于原始顺序（SPlayer 新加入的歌在关闭随机后丢失，删除的歌又回来）。开启随机时当前曲目放在第一位、其余打乱（SPlayer 整体打乱）。
- 不循环时播完最后一首停在末尾；手动“下一首”仍绕回第一首（SPlayer 不循环时也会绕回）。
- 播放失败不重试，直接跳过；连续 3 首失败、只有一首或不循环时已是最后一首，停止并提示，不清空队列（SPlayer 对网络错误先重试 3 次，只剩一首时清空队列）。
- 单元测试引入 vitest。

- [x] 状态机 `idle → resolving → loading → playing / paused → ended`，失败进入 `error`；Playback 是播放状态的唯一所有者。从浏览器恢复的队列处于 `paused`，按播放时才加载音频（与 SPlayer 的 `play()` 在没有音频时调用 `playSong` 相同）。源码确认，单元测试与运行确认见下。
- [x] 切歌请求标识与取消：先点 A 后点 B，A 较晚返回也不能覆盖 B；解析中暂停不得自动播放。每次切歌递增请求标识，解析、加载、开始播放、歌词载入与失败处理的结果都按标识丢弃；暂停另有尝试序号，丢弃暂停前发起的开始播放。用户的播放意图与状态分开记录，解析或加载中暂停，就绪后保持暂停。
- [x] 队列条目独立身份，允许重复入队；播放模式：顺序、列表循环、单曲循环、随机。条目为歌曲加 `entryId`（[queue.ts](../../frontend/src/core/player/queue.ts)），比较一律用 `entryId`。另修正 SPlayer `removeSongIndex` 的一个问题：移除最后一首时总把当前序号设为 0，即使当前播放的不是它，界面上的当前曲目因此跳到第一首而音频不变。循环与随机的提示照搬 SPlayer（“列表循环”“随机播放已开启”等）。
- [x] 点击列表的“播放”：专辑、歌单、我喜欢、最近播放取完整列表入队；全部歌曲与搜索结果只取已加载的部分。阶段 3.4 已按此实现（`fetchAll` 与 `usePagedList`），本阶段接上音频；运行确认：专辑页双击第一首，整张专辑入队并播放。
- [x] 播放超过 min(30 秒, 时长的一半) 才写入最近播放。按实际收听时长累计，跳转产生的位移不计入；每次播放只记一次，单曲循环重播和播完后重新播放算新的一次。
- [x] 播放失败自动跳过，最多连续 3 首，之后停止并提示错误。出错时用 `Range: bytes=0-0` 请求同一地址区分原因：401 回到登录页、不算失败；404 “歌曲已被删除”；不可达“无法连接服务器”；地址正常时按音频元素的错误码提示格式不支持、无法解码或网络错误。播放超过 0.5 秒或用户自己选歌后计数清零（前者照搬 SPlayer）。等待跳过的 0.5 秒内用户切了歌，不再自动跳过。等待期间按钮保持播放中（与 SPlayer 相同，2026-10-25 Eric 确认；此前显示播放图标约 0.3 秒，见[阶段 2 的播放错误对照](phase-2-player-baseline.md#01-播放器视觉基线)），期间暂停则跳到下一首但不播放；停止时才显示为暂停。
- [x] 歌词、封面、媒体信息和进度跟随当前条目。系统媒体会话照搬 SPlayer 的 Web API 部分（[MediaSessionManager.ts](../../frontend/src/core/player/MediaSessionManager.ts)：元数据、封面、进度、播放与上下一首、跳转），另外设置 `playbackState`。频谱与低频音量来自分析器，全屏播放器的频谱和背景低频响应因此有数据。页面标题照搬 SPlayer 网页版：正在播放时为“歌名 - 歌手 | Sonora”，暂停、切歌加载与出错时为“Sonora”（2026-10-25 Eric 确认）。阻止系统息屏用 Screen Wake Lock，设置开启且正在播放时持有（[wakeLock.ts](../../frontend/src/core/player/wakeLock.ts)；SPlayer 只在 Electron 中实现）。均衡器按 status store 的 `eqEnabled`、`eqBands` 应用，输出设备按设置的 `playDevice` 应用；两者的设置界面在阶段 5。
- [x] 队列与播放位置保存在浏览器中，重启后恢复为可继续播放的状态，不自动播放。保存在 IndexedDB（[storage.ts](../../frontend/src/core/player/storage.ts)，不引入 localforage）：队列（含随机前的原始顺序、条目身份与列表来源）在变化后 0.3 秒写入；位置按条目记录，播放中每 5 秒、暂停、跳转、切歌与关闭页面时写入。“记忆上次播放位置”关闭或位置属于别的条目时从头开始。“自动播放”开启时启动后尝试播放，浏览器拒绝时保持暂停。
- [x] 喜欢接上 `PUT/DELETE likes`（[utils/auth.ts](../../frontend/src/utils/auth.ts) 的 `toLikeSong`，照搬 SPlayer 的流程与提示）：服务端确认后才更新喜爱状态，失败时提示并保持原样；会话失效时回到登录页。
- [x] 单元测试覆盖切歌竞态、解析中暂停、队列编辑与恢复。证据：`task test:frontend`（vitest 5.0.5，[PlayerController.test.ts](../../frontend/src/core/player/PlayerController.test.ts)，31 个测试，可控的假引擎与解析），接入 `task test`。覆盖：先点 A 后点 B 且 A 较晚返回、加载中切歌、旧歌词较晚载入；解析中与加载中暂停、暂停后又播放、解析中拖动、自动播放被拒；重复入队、下一首播放当前歌曲、移除最后一首与当前曲目、拖动；不循环停在末尾、列表循环、单曲循环、随机的顺序与恢复；连续失败停止、计数清零、只有一首、401、等待跳过时保持播放中与期间暂停、等待跳过时切歌；页面标题；最近播放的阈值与跳转；保存与恢复（条目身份、位置、随机原始顺序、位置属于别的条目）。临时去掉切歌后的请求标识检查与就绪时的播放意图检查，各有测试失败。

运行确认（2026-10-25）：开发模式临时实例（`backend/.data/data` 的副本，媒体目录只读软链接，端口 14570/14572），Playwright Chromium 153 无头，真实点击，用完删除。专辑页双击歌曲后整张专辑入队并出声（分析器有信号，进度前进）；歌词载入、系统媒体信息为当前歌曲；暂停后时间不再前进；跳到 120 秒后继续；下一首；连续三次切歌只有最后一次生效；喜欢写入服务端后才显示；同一首歌重复入队；不存在的歌曲 404 后提示“歌曲已被删除，已自动跳过”并播放下一首；跳到 200 秒后收听 15 秒未记录、32 秒时记录一次且在最近播放最前；刷新页面后队列、条目身份与当前曲目一致，位置相差 1.6 秒（每 5 秒保存一次，关闭页面时的写入未必完成），处于暂停且没有加载音频，点播放从该位置继续。页面除这首不存在的歌曲的 404 外没有报错。

截图对照（`task test:visual`，2026-10-25）：54 个状态中 30 个一致、23 个属于已接受的例外、1 个不符（`set-yrc-amll`，17946 个像素，最大差值 47）。这是阶段 3.6 记录的 AMLL 偶发差异：改动前的 HEAD 上单独重跑 6 次，同样出现 1 次完全相同的不符；改动后 6 次全部属于例外。

补充（2026-10-25，按钮状态与页面标题）：同样的临时实例上真实播放，标题在播放时为“但愿人长久 - 邓丽君 | Sonora”，暂停后为“Sonora”，继续播放后恢复；拦截音频请求后点“下一首”，三首失败的过程中 `playStatus` 一直为 true（只有加载状态切换），第三首失败后变为 false 并提示停止，页面没有报错。`task test:visual`：54 个状态中 29 个一致、24 个属于已接受的例外、1 个不符（`set-yrc-amll`，19382 个像素，最大差值 45）。单独重跑 6 次都是同一个 17946 个像素的不符；改动前的 HEAD 上单独重跑 6 次，5 次同样不符、1 次属于例外。不是这次改动造成的，但这个偶发差异的出现频率比上次记录时高得多，原因仍未查明。

未完成或需要 Eric 确认：

- [x] 在桌面 Chrome 上实际听：2026-10-25 Eric 在开发模式（`task run:server`、`task run:web`）下手动确认，拖动进度条跳转、页面标题、刷新后恢复、实际听感（含淡入淡出）没有明显问题；细节在日常使用中继续观察。
- 输出设备的列表与切换（阶段 2 的 0.3）：转入[阶段 5](phase-5-player-pages.md)，与设置界面一起确认（2026-10-25 Eric 同意）。
- 真实 Authelia 下的浏览器登录：开发模式跳过 OIDC，未手动确认；服务端流程由 `task test:contract` 覆盖，播放端的未登录跳转见 4.1。转入[阶段 6](phase-6-first-release.md)，首次部署时确认（2026-10-25 Eric 同意）。
- “音频与歌词同步补偿”在 SPlayer 中只在“音频延迟模式”为 playback 时生效，Sonora 没有保留这项设置，补偿目前不起作用；Eric 确认阶段 5 接入设置界面时再定（2026-10-25）。已在阶段 5 解决（2026-10-27）：只保留手动补偿，始终生效，`AudioElementPlayer.currentTime` 加上补偿值，见[阶段 5](phase-5-player-pages.md)。

## 退出条件

现状（2026-10-25）：第一条中播放、暂停、拖动、切歌、歌词与重启恢复已由 Eric 在桌面 Chrome 上确认（开发模式）；真实 Authelia 登录与输出设备切换未手动确认，见上文。第二、三条的证据见 4.1、4.2、4.4 与截图对照。

- 管理后台上传 → OIDC 登录播放端 → 曲库列表 → 入队 → 播放/暂停/拖动/切歌 → 歌词显示 → 重启恢复，整条链路在 Chrome 中真实可听。
- 两个用户的喜爱、歌单与历史互不可见；Authelia 契约测试通过。
- 快速切歌、播放中暂停、队列编辑和重启不串状态、不丢数据；播放器相关页面完成视觉对照。
