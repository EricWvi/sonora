# 阶段 7：网易云音乐接入与卡片菜单

状态：进行中。卡片与批量操作已实现并运行确认（2026-11-04）；网易云部分的设计已确定（2026-11-05），实现中。前置：阶段 6 退出（2026-11-03）。

本阶段调整“与网易云接口相关的页面和入口全部不做”的原则：只接入下面列出的部分，其他网易云功能仍不做。

## 工作包

### 卡片与批量操作

这部分不依赖网易云，可以先做。

- [x] 卡片右键菜单：专辑、歌单卡片按 SPlayer 的 `CoverMenu.vue`（查看详情、播放或暂停、复制名称与 ID；分享不做），范围包括首页“最近添加的专辑”、专辑页、艺术家的专辑和搜索结果。歌单卡片目前只出现在搜索结果的“歌单”标签。实现：[CoverMenu.vue](../../frontend/src/components/Menu/CoverMenu.vue)，在 [CoverList.vue](../../frontend/src/components/List/CoverList.vue) 中接入，所有用 CoverList 的页面一并生效。`Eye` 图标从 SPlayer 复制（Google Material Icons，已在[视觉资源清单](../visual-assets.md#4-图标)中）。
- [x] 侧边栏歌单项的右键菜单（Eric 确认，2026-11-04）：SPlayer 的侧边栏没有右键菜单，是 Sonora 新增。条目与卡片一致再加管理：查看详情、播放或暂停、编辑歌单、删除歌单、复制歌单名称与 ID；编辑与删除复用歌单详情页的弹窗与确认，删除正在查看的歌单时回到首页。实现见 [SidebarMenu.vue](../../frontend/src/components/Layout/SidebarMenu.vue)：Naive UI 菜单项的 `props` 只作用于分隔线，右键事件经 `n-menu` 的 `node-props` 绑定。
- [x] 首页“最近添加的歌曲”卡片的右键菜单：用歌曲列表现有的 `SongListMenu`（立即播放、下一首播放、添加到歌单、更多操作、查看专辑、查看艺术家、同名搜索），菜单设置 `contextMenuOptions` 同样生效。
- [x] 详情页批量操作（Eric 确认，2026-11-04）：专辑、歌单、我喜欢的音乐三个详情页的“更多 → 批量操作”，弹窗照搬 SPlayer 的 `BatchList.vue`（表格勾选、“高级筛选”按序号范围选择）。与 SPlayer 不同：没有批量下载与删除本地歌曲；所有页面有“添加到歌单”与“下一首播放”（Sonora 新增，按勾选顺序插到当前歌曲之后，[PlayerController.addNextSongs](../../frontend/src/core/player/PlayerController.ts)）；歌单页另有“删除选中的歌曲”；我喜欢的音乐另有“取消喜欢”（Sonora 新增，逐首请求，中途失败时只移除已成功的）。移除后弹窗表格与页面列表、头部数量同步更新。实现：[BatchList.vue](../../frontend/src/components/Modal/BatchList.vue)、[modal.ts](../../frontend/src/utils/modal.ts) 的 `openBatchList`。

运行确认（2026-11-04，开发模式临时实例，本机 337 首的曲库副本，Playwright Chromium 无头，1440×1000；脚本在 `/tmp`，未入库）：

- 首页专辑卡片右键：查看详情、播放、复制专辑名称、复制专辑 ID；从菜单播放后再右键变为“暂停”。
- 首页歌曲卡片右键：歌曲信息头部与歌曲列表的菜单条目一致。
- 侧边栏歌单项右键：六个条目齐全；查看详情进入歌单页；删除歌单后回到首页，侧边栏不再显示。
- 歌单页批量：按钮为删除选中的歌曲、下一首播放、添加到歌单；勾选 2 首下一首播放提示“已添加 2 首至下一首播放”；删除后弹窗剩 3 行、勾选清零，页面头部数量为 3，服务端歌单为 3 首。
- 专辑页批量：按钮为下一首播放、添加到歌单；高级筛选 1–2 选中 2 首。
- 我喜欢的音乐批量：按钮为取消喜欢、下一首播放、添加到歌单；取消 1 首后服务端由 3 首变为 2 首，弹窗与页面头部同步。
- 全程没有页面错误。`task lint:frontend`、`task test:frontend`（57 个测试，新增批量下一首播放 1 个）、`pnpm build` 通过，产物不含 fixture。
- [x] 歌曲页与艺术家页的批量操作（Eric 确认，2026-11-04；最常用的场景是把最近添加的歌曲批量加入歌单）：这两页的歌曲是分页载入的，点“更多 → 批量操作”时按当前搜索与排序取全部歌曲（每页 500 首），再打开同一个批量弹窗，操作为“下一首播放”“添加到歌单”；取歌期间“更多”按钮显示加载中。歌曲页默认最新入库的在前，用“高级筛选 1–N”即可选中最近添加的一批。不加入库时间列（Eric 决定）。艺术家页新增“更多”按钮，照搬 SPlayer 只在“歌曲”标签下显示；与 SPlayer 不同，SPlayer 只处理已载入的歌曲。实现：[songs.vue](../../frontend/src/views/Library/songs.vue)、[Artist/layout.vue](../../frontend/src/views/Artist/layout.vue)。
- [x] 歌单页头部随别处的编辑同步：从侧边栏右键菜单编辑正在查看的歌单后，详情页头部的名称、简介与更新时间按刷新后的歌单列表更新，不重新读取歌曲（[playlist.vue](../../frontend/src/views/List/playlist.vue)）。

运行确认（2026-11-04，同上的临时实例与浏览器）：

- 歌曲页：“更多”只有批量操作，弹窗共 337 行（与服务端总数一致），第一行是最新入库的歌；高级筛选 1–5 后添加到歌单，服务端歌单为这 5 首，顺序与列表一致。搜索“陈奕迅”后打开批量，弹窗为 12 行，与搜索结果总数一致。
- 艺术家页（18 首的艺术家）：歌曲标签下有“更多”，弹窗 18 行，勾选 1 首下一首播放提示“已添加 1 首至下一首播放”；切到专辑标签后“更多”隐藏。
- 歌单页：从侧边栏编辑名称与简介后，头部立即显示新的名称与简介。
- 全程没有页面错误；`task lint:frontend`、`task test:frontend`（57 个）通过。

### 网易云音乐

已确定（Eric，2026-11-05）：

- **侧边栏歌单分组**（网易云之前的准备改动）：取消“创建的歌单”的折叠，改为始终展开的普通小标题，文字样式不变；照搬 SPlayer 的在线/本地歌单切换按钮（`NPopselect`），在线歌单为网易云上自己创建的歌单，收藏的歌单放进“我的收藏”；切换后记住选择，默认本地歌单（SPlayer 默认在线）。“+”在本地模式下新建 Sonora 歌单，在线模式下新建网易云歌单。
- **侧边栏第二块**：不加小标题，共三项：我喜欢的音乐、我的收藏、最近播放。我喜欢的音乐与最近播放是本地与在线歌曲的混合列表；我的收藏是网易云独有的（歌单、专辑、歌手，不含视频与播客）。没有登录网易云时，混合列表只有本地歌曲。
- **右上角账号**：OIDC 之外可以再登录网易云账号，不做切换账号；两个账号分别退出，退出 OIDC 不清除网易云会话。菜单按钮优先显示网易云的头像、昵称与 VIP 标志；菜单中不放歌单、专辑、歌手的统计。
- **登录方式**：二维码与 Cookie，不做手机号。
- **会话**：存在 Sonora 服务端，按 OIDC 用户隔离（每个用户绑定各自的网易云账号），Cookie 不进入浏览器。
- **接口调用**：`@neteasecloudmusicapienhanced/api` 单独作为一个容器运行（compose 多一个服务），由 Go 服务转发并带上该用户的 Cookie；Sonora 镜像不变，[部署说明](../deployment.md)补充这个服务。不做灰色歌曲解锁（UnblockNeteaseMusic）。
- **最近播放**：用 Sonora 服务端现有的历史记录，扩展为也记录在线歌曲，按时间混合排列。SPlayer 的最近播放同样是自己记录的（浏览器 IndexedDB 的 `historyList`），不调网易云接口。

- **混合列表**：我喜欢的音乐按喜欢时间混排（网易云“我喜欢”歌单的每首歌带喜欢时间）；同一首歌本地与在线都喜欢时只显示一条，优先本地。首页的“我喜欢的音乐”快捷入口与“最近播放”栏目同样混合。没有登录网易云时，“我的收藏”隐藏。
- **喜欢与歌单**：喜欢按钮按歌曲来源分别写入 Sonora 与网易云。本地歌单只能放曲库中的歌曲，在线歌曲要先导入；在线歌曲只能加入网易云歌单。
- **在线歌曲播放时的按钮**：底部栏与全屏播放器在本地歌曲时保持 Sonora 现状，在线歌曲时按 SPlayer；不做评论、MV 等，SPlayer 的“下载”换成“导入到曲库”。播放器菜单改为与 SPlayer 一致的四项（均衡器、自动关闭、AB 循环、播放速度），本地与在线歌曲都有。在线歌曲与本地歌曲走同一条 Web Audio 链路，均衡器与频谱都可用（实测见下面的“实验”，原来“CDN 不允许跨域时不支持均衡器”的预案不再需要）；不经 Go 转发音频。
- **导入单曲**：
  - 只有指定 Authelia 组的成员可以导入，组外成员不显示导入相关的按钮；服务端同样校验。现有 OIDC 客户端的 scope 只有 `openid profile email`，需要在 Authelia 客户端配置中加 `groups`（Eric 修改服务器配置）；可导入的组名写在 Sonora 的配置文件中，不写死在代码里，[部署说明](../deployment.md)同步说明。
  - 由 Go 服务端完成：用服务端保存的该用户网易云会话取播放地址（音质沿用 SPlayer 默认的 `exhigh` 极高），直接下载 mp3 入库，文件不经过浏览器。
  - 元数据同时写入 mp3 的 ID3 标签与数据库：标题、艺术家、专辑、专辑艺术家（与艺术家相同，不用网易云的专辑艺术家）、年份与封面；不写音轨号与碟号。
  - 总是新建专辑，即使曲库已有同名、同艺术家的专辑；封面作为单曲封面，新专辑也用这张封面（新建的专辑不会覆盖已有封面）；新专辑标为“单曲”。专辑合并、改名等到管理后台处理。
  - 歌词：网易云有 yrc 逐字歌词时转为 TTML，没有时用普通 LRC。不用 AMLL TTML 歌词库（Eric 决定，2026-11-05）：它按网易云 ID 收录的 TTML 约 3,751 首（2026-11-05 查 GitHub 仓库），偏日语与 Vocaloid，对这个曲库命中率低，不值得为此依赖 GitHub 访问。个别歌曲的逐字歌词质量不够时，在管理后台人工上传 TTML 或 LRC 替换（现有功能，`PUT tracks/{id}/lyrics`）。翻译与罗马音在网易云的歌词中有就放入，没有就不放；缺少的字段不通过搜索其他歌曲或来源补齐（Eric，2026-11-05）。
  - 歌曲记录网易云歌曲 ID（新增迁移）：导入后即为本地歌曲，据此判断是否已经导入、混合列表去重，也用于以后检测并替换逐字歌词。
  - 入口：在线歌曲的右键菜单“导入到曲库”，以及底部栏与全屏播放器的菜单。
- **本次只做单曲导入**；批量导入、专辑导入，以及检测本地歌曲对应的在线歌曲是否有逐字歌词并替换，作为后续工作包。
- **发布**（Eric，2026-11-05）：网易云开发前不发 v1.5.0，CI 中的版本号注入等下一次发布时验证。
- **接口服务的运行方式**（Eric，2026-11-05）：`node:22-alpine` 运行锁定版本的 `npx -y @neteasecloudmusicapienhanced/api@<版本>`，挂缓存卷，不必每次启动都下载；走 compose 内部网络，不映射端口、不经 Caddy（该服务没有鉴权，不能暴露）；地址写在 Sonora 的配置文件。
- **ID3 写入**（Eric，2026-11-05）：`dhowden/tag` 只能读，写入引入 `bogem/id3v2`；引入前确认许可与维护情况。
- **开发账号**（Eric，2026-11-05）：开发与测试网易云功能时使用 `.local/credentials/splayer/only_session` 中的网易云 Cookie。这是对“凭据只用于参考实例”的明确例外，范围只限本机的开发实例与接口服务：只取 `MUSIC_U`、`__csrf`、`NMTID`，`only_session` 不是网易云的 Cookie，不发送；凭据不输出、不提交、不写入测试产物，临时实例的数据库用完删除。

实验（2026-11-05，`@neteasecloudmusicapienhanced/api` 4.41.1 运行在本机回环，脚本在 `/tmp`，未入库）：

- 地址：“我喜欢”中 25 首歌都返回了完整地址（没有 `freeTrialInfo`）。接口返回的是 `http:` 地址，页面是 https，必须照搬 SPlayer 换成 `https:`；请求 `exhigh` 时返回的 `level` 有 `exhigh` 也有 `standard`，都是 mp3。
- 跨域：m7、m8、m701、m801 四个 CDN 域名带 `Origin` 的 Range 请求都返回 206 与 `Access-Control-Allow-Origin: *`。SPlayer 把 m804、m704 改写成 m801、m701，这次的样本里没有出现这两个域名，改写照搬。
- Web Audio：无头 Chromium 中按 Sonora 的链路（`crossOrigin = "anonymous"` → `MediaElementSource` → 10 段 peaking 滤波 → 分析器 → 输出）播放 3 首（m801、m701），分析器峰值 0.62–1.86，不是静音；跳到 60 秒成功，跳转后信号正常，没有错误事件。SPlayer 网页版也是这条链路，没有别的方案。

实现时要注意（来自 2026-11-05 的代码核对）：

- `history`、`likes`、`playlist_tracks` 的 `track_id` 都是指向 `tracks` 的外键，在线歌曲不能直接写入。历史记录要加来源与歌曲快照；喜欢不写入本地（按来源写网易云）；本地歌单不放在线歌曲。
- 本地与网易云的歌曲 ID 都是整数，可能相同。队列、当前歌曲、喜欢状态与预加载要改用带来源的键，这部分在混合列表之前先做。
- 播放地址带时效：不存入队列与设备存储，播放和预加载时再取；恢复播放时重新取。

工作包：

- [x] 侧边栏歌单分组不能折叠，加在线/本地歌单切换按钮（2026-11-05）。实现见 [SidebarMenu.vue](../../frontend/src/components/Layout/SidebarMenu.vue)：分组的 `expanded-keys` 固定为展开，标题去掉箭头与点击效果（侧边栏收起时照常弹出）；模式存在 status store 的 `playlistMode`（浏览器）；接入网易云登录之前“在线歌单”不可选。原来记录展开状态的设备设置 `menuExpandedKeys` 随之删除。运行确认（临时实例，无头 Chromium，1440×1000）：载入后歌单直接可见，标题为“本地歌单”，没有箭头，点标题不折叠；切换按钮弹出“在线歌单”（不可选）与“本地歌单”（选中）；点歌单进入歌单页；没有页面错误。`task lint:frontend`、`task test:frontend`（57 个）通过。
- [x] a. 接口服务与账号（2026-11-05）：接口服务容器与 Go 的转发，部署说明；服务端会话（迁移 10）；登录（二维码、Cookie）；右上角菜单的双账号。实现：[netease](../../backend/internal/netease/client.go)、[server/netease.go](../../backend/internal/server/netease.go)、[store/netease.go](../../backend/internal/store/netease.go)、[UserMenu.vue](../../frontend/src/components/Layout/UserMenu.vue)、[Modal/Login](../../frontend/src/components/Modal/Login/Login.vue)。

  与 SPlayer 不同：

  - Go 只开放具名接口，不做通用转发，上游响应中的 Cookie 到此为止。
  - Cookie 登录时先校验，无效的 Cookie 当场提示，不保存；SPlayer 先存下，之后才发现。
  - 退出网易云只删除 Sonora 保存的会话，不调用 `/logout`，否则同一个 Cookie 在别的客户端也会失效。
  - 登录弹窗没有标签页（不做手机号），也没有 UID 登录；菜单不放统计与切换账号，也不放等级（Lv）标签。
  - 二维码内容由 Go 返回。[依赖边界检查](../../frontend/boundaries.mjs)不变：播放端不出现网易接口的地址与 Cookie 字段名，模块名也不含 `netease`、`cookie`，所以 Cookie 登录的组件叫 `LoginPaste.vue`。

  **接口服务的响应缓存**（实现中发现）：服务默认缓存响应 2 分钟，缓存键是域名、URL 与请求的 Cookie 头，不含 POST 正文。Cookie 放在正文时所有用户共用同一条缓存，表现为无效 Cookie 的结果被缓存后，紧接着的有效 Cookie 被拒绝，以及无效 Cookie 拿到了真实账号的资料；上线后会把一个用户的数据返回给另一个用户。现在每个请求都带 `x-apicache-bypass`，并在 URL 上带纳秒时间戳，两者都有单元测试（`TestCookieStaysInTheBody`、`TestEveryCallHasItsOwnURL`）。

  证据：Go 测试覆盖两种登录、拒绝无效 Cookie、用户隔离、退出播放端不影响网易云会话、失效与接口服务不可用、3 天续期、响应不含 Cookie；迁移 10 由已有的逐版本升级测试覆盖。运行确认（2026-11-05，开发模式临时实例，本机 337 首的曲库副本，真实接口服务 4.41.1 运行在本机回环，无头 Chromium，1440×1000；脚本在 `/tmp`，未入库）：未登录时菜单显示“登录网易云”；弹窗显示真实二维码与“请打开云音乐 APP 扫码登录”；无效 Cookie 提示“Cookie 无效或已过期”，紧接着用真实 Cookie 登录成功；按钮变为网易云头像（图片已加载）、昵称与 VIP 标志；菜单有“退出网易云”与 OIDC 账号，没有统计；浏览器的 Cookie 与 localStorage 中没有网易云 Cookie；刷新后保持；退出网易云后状态为 `signedOut`，Sonora 登录不受影响。全程没有页面错误，经过浏览器的网易云响应、Sonora 日志与接口服务日志中都没有 Cookie。

  Eric 在临时实例上用手机扫码登录成功（2026-11-05），按钮与菜单显示正常；菜单中的昵称与 VIP 标志按 Eric 的要求改为一行（SPlayer 分两行，因为第二行还有 Lv 标签）。

  待验证：部署说明中 compose 的写法（本机是 `node app.js`，没有在容器里运行）。

- [ ] b. 侧边栏第二块：我喜欢的音乐、我的收藏、最近播放（迁移 11，历史记录支持在线歌曲），在线歌单模式，首页入口，含未登录、登录失效和接口失败时的状态。先把队列改为带来源的键。
- [ ] c. 本地与在线曲库：视觉相同、接口不同；播放队列中混合本地与在线歌曲；喜欢按来源写入。
- [ ] d. 播放器：在线歌曲的按钮，菜单改为四项；运行 `task test:visual`，预期差异记为例外。
- [ ] e. 导入外部歌曲：单曲导入（迁移 12，`tracks` 记录网易云歌曲 ID；OIDC 的 `groups`、配置中的组名）。
- [ ] 后续：批量导入、专辑导入；检测本地歌曲对应的在线歌曲是否有逐字歌词并替换。

## 依据

源码确认（`reference/SPlayer`，2026-11-04）：

- 接口包：`@neteasecloudmusicapienhanced/api`，在 SPlayer 的本机服务中调用。
- 登录：`/login/qr/*`、`/login/cellphone`、`/login/status`、`/login/refresh`，见 [login.ts](../../reference/SPlayer/src/api/login.ts)。
- 喜欢与收藏：`/likelist`、`/user/playlist`、`/album/sublist`、`/artist/sublist`，见 [user.ts](../../reference/SPlayer/src/api/user.ts)。
- 最近播放：SPlayer 的 `src/api` 中没有网易云最近播放的请求，最近播放是浏览器 IndexedDB 中的 `historyList`（[data.ts](../../reference/SPlayer/src/stores/data.ts)）。
- 歌词：`/lyric/new`（含逐字的 yrc、翻译与罗马音）；网页版另从第三方 AMLL TTML 歌词库取 TTML（`songLyricTTML`，[song.ts](../../reference/SPlayer/src/api/song.ts)）。
- 播放地址：`/song/url/v1`，默认音质 `exhigh`（极高）。
- 部署：SPlayer 的镜像是 nginx 加同一容器里的 `npx @neteasecloudmusicapienhanced/api`。
- 批量操作：`BatchList.vue` 的操作有批量下载（开发者模式）、删除选中的歌曲（歌单）、添加到歌单与删除本地歌曲。

## 待决事项

无（2026-11-05）。

## 退出标准

草拟（2026-11-05），待 Eric 确认：

- [ ] 二维码与 Cookie 登录可用；两个账号分别退出；不同 OIDC 用户的网易云会话互相隔离；Cookie 不进入浏览器、日志与接口响应。
- [ ] 未登录、登录失效、接口服务不可达时，本地功能全部不受影响，网易云入口的状态有说明。
- [ ] 混合列表按时间混排并去重；队列中混合本地与在线歌曲，切歌、预加载、刷新恢复与失败跳过正常；均衡器、频谱在线歌曲同样生效。
- [ ] 单曲导入：组外用户看不到入口且服务端拒绝；导入后的文件、数据库、封面与歌词符合“导入单曲”的规定，重复导入被识别。
- [ ] 升级与部署说明、接口说明已同步；`task test:visual` 0 个不符，差异记为例外；Eric 在桌面 Chrome 上确认。
