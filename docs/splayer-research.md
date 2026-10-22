# SPlayer 源码调研：Sonora 重写与样式复刻依据

调研日期：2026-10-22。范围：工作区 `reference/SPlayer` 源码，子模块提交 `28b76bad9eb800e0c90dac8475d4769acfe2450b`；这是对当前参考副本的静态阅读，不代表上游最新版本，也不代表各平台运行验证已完成。本文不包含账号凭据，不依赖真实登录 Cookie。

## 1. 结论与边界

**已确定的产品约束**：Sonora 采用 Web 前端与 Go 后端，明确舍弃 Electron 及原项目的桌面运行行为；界面样式完全按照 SPlayer 还原；舍弃网易云音乐来源，使用我们自己开发的音乐服务。下文网易能力描述是原项目的源码事实，不是 Sonora 的接入计划。功能目录用于识别工作量与依赖，不等于全部列入首发。详细阶段安排见 [重写路线图](roadmap.md)。

**样式硬性要求**：必须严格复刻 SPlayer，所有实现页面及其视觉状态均以参考源码和实际渲染为准，逐项还原布局、尺寸、字体、配色、图标、效果、动效与响应式行为。更换音乐服务与重写业务不改变这一要求；未经 Eric 明确同意，不得重新设计、简化或替换原样式。交付必须提供原项目与 Sonora 的同条件截图对照及交互检查，按[强制验收标准](roadmap.md#6-样式还原与验收方法)处理差异。

**源码事实**：SPlayer 不只有网易云音乐能力。它同时实现本地音乐、多个自建流媒体协议、多个音频后端和大量桌面集成；但网易身份、数字歌曲 ID、音质及权益字段仍深入公共类型、播放器调度、组件和状态。把网易请求地址改掉无法完成解耦。证据见 [公共歌曲类型](../reference/SPlayer/src/types/main.ts#L39)、[歌曲管理器](../reference/SPlayer/src/core/player/SongManager.ts#L1)、[菜单](../reference/SPlayer/src/components/Layout/Menu.vue#L19)、[路由守卫](../reference/SPlayer/src/router/index.ts#L35)。

**建议**：把原项目作为视觉和交互规范，以 Sonora 自己的领域模型与自研音乐服务契约重新承载；移除网易认证、接口包和平台字段，不实现网易适配器。样式复刻和能力重建分开验收，避免视觉进度被多引擎、系统歌词、音源解锁等原项目重功能绑住。

## 2. 技术栈与运行拓扑

| 层 | 当前源码事实 | 对重写的影响 |
| --- | --- | --- |
| 前端 | Vue `^3.5.35`、TypeScript `^6.0.3`、Vue Router `^5.1.0`、Pinia `^3.0.4` | Eric 已确定前端技术栈完全按照 SPlayer；Sonora 初始化沿用这些框架 |
| 组件与样式 | Naive UI `2.43.2`、Sass、SVG 图标、Material Color Utilities | 视觉并非一套独立 CSS，包含组件库主题覆盖、动态配色和 scoped 样式 |
| 桌面 | Electron `^41.7.1`、electron-vite、electron-builder | Web 与桌面功能存在明确差异，不能用浏览器页面代表完整桌面客户端 |
| 本机服务 | Fastify，网易 API、QQ 歌词、音源匹配、控制接口 | 既是静态页面入口，也是本机 API 网关，不只是开发服务器 |
| 原生 | Rust workspace、NAPI 模块、OpenCC WASM | 本地扫描、音频分析、下载、媒体控件与任务栏歌词带来跨平台编译和打包成本 |
| 持久化 | Pinia 持久化、localforage、electron-store、better-sqlite3、文件缓存 | 数据散布在浏览器存储和主进程，迁移应按数据职责设计 |

版本和脚本来自 [package.json](../reference/SPlayer/package.json)，Rust 成员来自 [Cargo.toml](../reference/SPlayer/Cargo.toml)。这些是依赖声明范围，不是已安装版本或上游最新版本。

运行链路：`src/main.ts` 挂载应用、Pinia、Router 和全局样式；`App.vue` 为普通页面挂载全局 Provider；`AppLayout.vue` 初始化应用、挂载侧栏、顶部导航、路由页、队列、底部播放器与全屏播放器。参见 [入口](../reference/SPlayer/src/main.ts)、[根组件](../reference/SPlayer/src/App.vue)、[布局](../reference/SPlayer/src/layout/AppLayout.vue)、[初始化](../reference/SPlayer/src/composables/useInit.ts)。

本机 Fastify 默认监听 `127.0.0.1:25884`，可由 `VITE_SERVER_PORT` 覆盖；生产模式还服务 renderer 静态文件。源码能够解释用户提供的端口，但仅凭默认值不能证明正在运行的进程与该副本完全一致。参见 [服务入口](../reference/SPlayer/electron/server/index.ts#L14)。

## 3. 路由与用户功能清单

以下按路由配置列举，**不是首发承诺**。设置、登录、主题、队列和全屏播放器有大量弹窗或全局组件，不会全部出现在路由表中。

| 功能域 | 已有入口 | 限制与依赖 |
| --- | --- | --- |
| 推荐与发现 | 首页、歌单发现、排行榜、歌手、新歌 | 主要在线内容入口；具体数据仍由页面对应 API 决定 |
| 搜索 | 歌曲、歌单、歌手、专辑、视频、播客 | 必须有 `keyword` 查询参数 |
| 内容详情 | 歌单、专辑、歌手歌曲/专辑/视频、视频、歌曲百科、评论 | 多数要求 `id` 查询参数 |
| 播客 | 热门播客、分类、播客详情 | 分类还要求名称参数 |
| 账号音乐 | 喜欢歌曲、云盘、每日推荐、收藏歌单/专辑/歌手/视频/播客 | 路由标记 `needLogin`，守卫直接依赖全局 `isLogin()` |
| 本地音乐 | 歌曲、歌手、专辑、文件夹、本地歌单 | 路由标记 `needApp`，浏览器会进入 403 |
| 下载 | 已下载、下载中 | 路由标记 `needApp` |
| 自建流媒体 | 歌曲、歌手、专辑、歌单、流媒体歌单详情 | 有独立服务器配置和来源 API |
| 其他 | 最近播放、403/404/500 | 最近播放不要求在线登录 |
| 独立歌词窗 | `/desktop-lyric` | 不经普通 AppLayout，要求桌面环境 |

证据：[完整路由表](../reference/SPlayer/src/router/routes.ts#L8)、[客户端与登录守卫](../reference/SPlayer/src/router/index.ts#L35)。路由采用 hash history；抓取原页面时应记录 `/#/...` 路由和查询参数，不能只留首页 URL。

菜单还承载私人漫游、心动模式、账号歌单和显示开关，不能从路由数量推断全部交互。参见 [Menu.vue](../reference/SPlayer/src/components/Layout/Menu.vue)。顶部搜索和用户入口受 `useOnlineService` 控制，客户端窗口按钮受运行环境和无边框设置控制，参见 [Nav.vue](../reference/SPlayer/src/components/Layout/Nav.vue#L34)。

自建服务类型明确列出 `navidrome`、`jellyfin`、`subsonic`、`opensubsonic`、`emby`，实现入口分别位于 [流媒体 API](../reference/SPlayer/src/api/streaming/index.ts) 和 [流媒体模型](../reference/SPlayer/src/types/streaming.ts#L10)。这是可用能力的源码证据，并非已测试所有服务器版本。

## 4. 播放、音频与歌词

### 4.1 当前分层

- [PlayerController](../reference/SPlayer/src/core/player/PlayerController.ts#L30)：应用级播放控制，协调状态、队列、歌词、音频、播放记录和桌面消息；已有 Last.fm 逻辑，不能当成纯音频引擎照搬。
- [SongManager](../reference/SPlayer/src/core/player/SongManager.ts#L49)：获取和预取播放地址、缓存、音质降级、私人漫游；直接导入网易歌曲和推荐 API，同时识别本地和流媒体。
- [AudioManager](../reference/SPlayer/src/core/player/AudioManager.ts#L21)：统一调用播放后端，处理跨曲渐变和事件转发。
- [IPlaybackEngine](../reference/SPlayer/src/core/audio-player/IPlaybackEngine.ts)：播放/暂停/定位、音量、倍速、输出设备、增益、频谱和效果能力契约；部分方法可选。
- [LyricManager](../reference/SPlayer/src/core/player/LyricManager.ts#L1)：歌词获取、缓存、预取、来源优先级和解析处理；直接依赖网易与 QQ API。
- [MediaSessionManager](../reference/SPlayer/src/core/player/MediaSessionManager.ts)：媒体会话；[PlayerIpc](../reference/SPlayer/src/core/player/PlayerIpc.ts)：桌面交互。

### 4.2 三种后端有能力差异

| 后端 | 源码行为 | 重写需要验证的点 |
| --- | --- | --- |
| AudioElementPlayer | 浏览器音频元素后端；声明倍速、设备切换、均衡器、频谱能力 | 编解码器、自动播放、音频上下文恢复、跨域媒体 |
| FFmpegAudioPlayer | Worker/WASM 后端；声明相同四类能力 | 隔离环境、SharedArrayBuffer、解码内存、进度和定位 |
| MpvPlayer | 仅 Electron 可选；声明倍速、设备切换，不支持当前接口的均衡器和频谱 | mpv 发现和生命周期、平台安装、进程故障恢复 |

证据：[音频选择与回退](../reference/SPlayer/src/core/player/AudioManager.ts#L42)、[元素后端能力](../reference/SPlayer/src/core/audio-player/AudioElementPlayer.ts#L28)、[FFmpeg 能力](../reference/SPlayer/src/core/audio-player/ffmpeg-engine/FFmpegAudioPlayer.ts#L80)、[mpv 能力](../reference/SPlayer/src/core/audio-player/MpvPlayer.ts#L54)、[mpv 主进程服务](../reference/SPlayer/electron/main/services/MpvService.ts)。FFmpeg 所需环境由 [checkIsolationSupport](../reference/SPlayer/src/utils/env.ts#L34) 检查安全上下文、跨源隔离及 SharedArrayBuffer；条件不满足会回退到元素后端。mpv 的 crossfade 也回退到普通切换，不能承诺三个后端完全相同。

自动混音另有 [AutomixManager](../reference/SPlayer/src/core/automix/AutomixManager.ts)、[AudioScheduler](../reference/SPlayer/src/core/automix/AudioScheduler.ts)、分析 Worker 和 Rust 分析工具。它依赖时序、预取、双播放器与音频分析，是独立复杂度来源，建议在基本播放稳定后单独决策。

### 4.3 歌词不是一个文本字段

源码包含 LRC/YRC/QRC/TTML 解析、逐字时间、翻译、音译、背景声和对唱标记；还有简繁转换、文本清洗、对齐与多来源缓存。证据：[歌词结构](../reference/SPlayer/src/types/main.ts#L193)、[歌词管理器](../reference/SPlayer/src/core/player/LyricManager.ts)、[解析目录](../reference/SPlayer/src/utils/lyric)、[AMLL 歌词显示](../reference/SPlayer/src/components/AMLL/LyricPlayer.vue)、[OpenCC WASM](../reference/SPlayer/native/ferrous-opencc-wasm/src/lib.rs)。

**建议**：领域层保留可选的行/字时间和语言信息；来源负责下载和转换原始格式，显示组件只接收统一结构。歌词同步应以播放时钟为准，拖动、倍速和切歌取消必须纳入验收。完整逐字歌词和高级特效是否首发另行决策。

## 5. 桌面与原生能力边界

| 能力 | 主要实现位置 | 规划含义 |
| --- | --- | --- |
| 主窗、登录窗、桌面歌词、任务栏歌词窗 | [windows](../reference/SPlayer/electron/main/windows) | 多窗口状态恢复、焦点、穿透、位置和生命周期都需要独立测试 |
| 托盘、全局快捷键、Windows 缩略图工具条 | [tray](../reference/SPlayer/electron/main/tray/index.ts)、[shortcut](../reference/SPlayer/electron/main/shortcut/index.ts)、[thumbar](../reference/SPlayer/electron/main/thumbar/index.ts) | 浏览器外观检查覆盖不到这些能力 |
| 系统媒体控制、Discord RPC | [external-media-integration](../reference/SPlayer/native/external-media-integration/index.d.ts) | 存在 Linux/macOS/Windows 分平台实现 |
| Windows 任务栏歌词 | [taskbar-lyric](../reference/SPlayer/native/taskbar-lyric/src/lib.rs) | 有 Win10/Win11 策略、任务栏和注册表监听，不是普通 DOM 布局 |
| 本地扫描、元数据、下载、音频分析 | [tools 类型](../reference/SPlayer/native/tools/index.d.ts)、[LocalMusicService](../reference/SPlayer/electron/main/services/LocalMusicService.ts) | Rust NAPI 与主进程数据库配合，带来各平台产物要求 |
| 歌词 Socket、本机控制 API | [SocketService](../reference/SPlayer/electron/main/services/SocketService.ts)、[control API](../reference/SPlayer/electron/server/control/index.ts) | 需要定义监听范围、事件与命令边界 |
| 自动更新、协议处理 | [update](../reference/SPlayer/electron/main/update/index.ts)、[protocol](../reference/SPlayer/electron/main/utils/protocol.ts) | 分发阶段问题，不能仅通过 Web 页面验收 |

现有 [preload](../reference/SPlayer/electron/preload/index.ts) 暴露 electron 工具 API、store 与 logger。窗口默认配置明确包含 `sandbox: false`、`webSecurity: false`、`nodeIntegration: true`，见 [窗口工厂](../reference/SPlayer/electron/main/windows/index.ts#L19)。这证明参考实现选择了宽松的桌面配置；并不意味着样式复刻必须复制该权限模型。

**Sonora 范围**：Web 能力和自研 Go 服务职责需重新划分，不能以原项目 IPC 接口作为新实现的契约。以上桌面能力只记录原项目事实。Sonora 已明确舍弃 Electron，不迁移桌面宿主、IPC、原生模块和窗口行为；采用 Web 前端与 Go 后端，当前后端只建占位目录。

## 6. 数据存储与迁移面

| 载体 | 已观察的数据职责 | 证据 |
| --- | --- | --- |
| localStorage | settings/status/music/shortcuts/data 的 Pinia 持久化；Cookie 的备用存储 | [setting](../reference/SPlayer/src/stores/setting.ts#L831)、[music](../reference/SPlayer/src/stores/music.ts#L134)、[cookie](../reference/SPlayer/src/utils/cookie.ts) |
| localforage | 播放列表、历史、云盘列表、用户收藏、背景图、本地歌单及流媒体数据 | [data](../reference/SPlayer/src/stores/data.ts#L60)、[local](../reference/SPlayer/src/stores/local.ts)、[streaming](../reference/SPlayer/src/stores/streaming.ts#L18) |
| electron-store | 窗口/歌词位置、任务栏配置、代理、缓存路径与上限、WebSocket、下载设置 | [主进程配置](../reference/SPlayer/electron/main/store/index.ts) |
| SQLite | 本地曲目、音频分析、meta；缓存键值、类型、字节数据、大小、访问时间 | [LocalMusicDB](../reference/SPlayer/electron/main/database/LocalMusicDB.ts#L15)、[CacheDB](../reference/SPlayer/electron/main/database/CacheDB.ts#L27) |
| 文件系统 | 本地音乐、封面、歌曲缓存和下载结果 | [LocalMusicService](../reference/SPlayer/electron/main/services/LocalMusicService.ts#L25)、[MusicCacheService](../reference/SPlayer/electron/main/services/MusicCacheService.ts)、[DownloadService](../reference/SPlayer/electron/main/services/DownloadService.ts) |

流媒体服务器配置的 `password` 字段在类型注释中明确写为明文存储，见 [StreamingServerConfig](../reference/SPlayer/src/types/streaming.ts#L15)。Cookie 工具会将凭据写入 document.cookie 与 localStorage，并包含输出其键值的日志语句，见 [cookie.ts](../reference/SPlayer/src/utils/cookie.ts)。重写不应延续凭据日志；本文未读取任何真实凭据文件。

**建议**：为设置、来源账号、曲库、队列、历史、收藏、下载和可删除缓存分别定义所有权、schemaVersion 与导入导出边界。SPlayer 到 Sonora 的兼容迁移属于独立可选工作，不应未经用户确认就读取或自动导入旧账号数据。至少保证 Sonora 自身重启恢复、数据升级和缓存清理不误删用户音乐。

## 7. 网易耦合位置与替换策略

| 耦合点 | 源码事实 | 建议替代边界 |
| --- | --- | --- |
| 请求客户端 | 默认 `/api/netease`；拦截器按登录状态注入 `MUSIC_U`、IP 和代理参数 | 移除网易客户端，按自研服务契约实现请求和鉴权 |
| 后端网关 | 动态枚举网易 API 包导出，按路径调用；另在网易前缀下提供 TTML | 不迁移网易网关，直接定义自研音乐服务及歌词契约 |
| 身份与路由 | `needLogin` 用全局 `isLogin()` 判断，菜单直接检查登录 | 使用自研服务会话与具体能力判断 |
| 歌曲标识 | `SongType.id: number`；流媒体额外补 `originalId/serverId/serverType` | 内部实体引用包含来源实例、实体类型、来源原始 ID；避免来源间 ID 碰撞 |
| 版权与音质 | `free/pc/mv/mark`、网易音质等级进入公共类型 | 通用可播放性、试听限制、可选格式信息；原始来源信息留在适配器 |
| 播放地址 | SongManager 直接调用网易 songUrl/quality、私人漫游和音源匹配 | resolvePlayback 返回可失效的播放资源及限制，播放引擎只消费资源 |
| 数据转换 | formatSongsList 处理网易 `ar/al/dt/fee` 等字段与兼容数据结构 | 按自研服务契约映射实体，不保留网易字段兼容层 |
| 歌词与喜欢 | LyricManager 直接导入网易/QQ API，播放器喜欢按钮调用通用 auth 中的网易操作 | LyricsProvider 与 Library/Collection 服务，来源不支持时给出明确能力状态 |

证据：[request.ts](../reference/SPlayer/src/utils/request.ts#L1)、[网易本机网关](../reference/SPlayer/electron/server/netease/index.ts#L1)、[路由守卫](../reference/SPlayer/src/router/index.ts#L35)、[公共类型](../reference/SPlayer/src/types/main.ts#L39)、[SongManager](../reference/SPlayer/src/core/player/SongManager.ts#L1)、[formatSongsList](../reference/SPlayer/src/utils/format.ts#L49)、[LyricManager](../reference/SPlayer/src/core/player/LyricManager.ts#L1)、[底部播放器](../reference/SPlayer/src/components/Player/MainPlayer.vue#L62)、[auth 工具](../reference/SPlayer/src/utils/auth.ts)。

来源能力建议区分：浏览、搜索、读取歌单、修改歌单、收藏、歌词、下载、登录、推荐。不要假设本地目录或任意流媒体服务必须提供网易式推荐、播客、评论或云盘；缺失能力时的界面行为需明确规定，同时保留已确认的原样式。

## 8. 样式复刻的实际来源

### 8.1 必须一起研究的文件

| 视觉部分 | 权威源码 |
| --- | --- |
| 全局配置、Naive UI 字体/主题/组件覆盖 | [Provider.vue](../reference/SPlayer/src/components/Global/Provider.vue#L1) |
| 动态色板、明暗模式、封面取色、CSS 变量 | [color.ts](../reference/SPlayer/src/utils/color.ts)、[themeColor.json](../reference/SPlayer/src/assets/data/themeColor.json) |
| 全局 reset、背景透明、Naive 组件补充样式 | [main.scss](../reference/SPlayer/src/style/main.scss) |
| 动效名称与曲线 | [animate.scss](../reference/SPlayer/src/style/animate.scss) |
| 主布局与背景图/视频层 | [AppLayout.vue](../reference/SPlayer/src/layout/AppLayout.vue) |
| 侧栏、导航、账号、搜索 | [Layout](../reference/SPlayer/src/components/Layout)、[Search](../reference/SPlayer/src/components/Search) |
| 底栏、全屏/移动播放器、进度和控件 | [Player](../reference/SPlayer/src/components/Player)、[PlayerProvider](../reference/SPlayer/src/components/Global/PlayerProvider.vue) |
| 卡片、列表、更多菜单、弹窗、设置 | [components](../reference/SPlayer/src/components)、各 [views](../reference/SPlayer/src/views) 的 scoped 样式 |
| 图标机制与文件 | [SvgIcon.vue](../reference/SPlayer/src/components/Global/SvgIcon.vue)、[assets/icons](../reference/SPlayer/src/assets/icons) |
| 图片、Logo 字体、应用图标 | [public/images](../reference/SPlayer/public/images)、[public/fonts](../reference/SPlayer/public/fonts)、[public/icons](../reference/SPlayer/public/icons) |
| 响应式断点 | [useMobile.ts](../reference/SPlayer/src/composables/useMobile.ts) |

可直接核对的布局数值：侧栏宽 `240px`、折叠宽 `64px`，主内容顶偏移 `70px`，有底栏时底偏移 `80px`，内容左右内边距移动端 `16px`、其他 `24px`。来源：[AppLayout.vue](../reference/SPlayer/src/layout/AppLayout.vue#L68)。这些不是所有页面完整设计 token，不能据此忽略卡片、文字、阴影、边框和滚动条。

断点同时存在 `<512`、`<640`、`<768`、`<990`、`>=1024`、`>=1280`；注释有不一致之处，应以实际表达式为准。桌面侧栏在 `>=1024` 出现，较窄窗口改用顶部菜单抽屉。来源：[useMobile.ts](../reference/SPlayer/src/composables/useMobile.ts)、[Nav.vue](../reference/SPlayer/src/components/Layout/Nav.vue#L52)。

### 8.2 建议的视觉证据与验收方法

1. 固定参考源码版本、运行环境、视口、缩放、系统字体、主题、侧栏状态和样例数据。
2. 按页面和状态采集：未登录/已登录、加载/空/错误、有无播放底栏、暂停/播放、长标题、失效封面、侧栏折叠、抽屉/菜单/设置弹窗。
3. 先把 Provider、CSS 变量、布局尺寸、图标和通用组件复刻到固定数据页面，再接来源服务。
4. 截图比较需对齐字体、图片、动画时刻；动态推荐数据和真实账号不可作为唯一稳定测试素材。对可动画元素固定时间或关闭动画另作静态对比，并补充真实动效检查。
5. 至少检查参考窗口 `1280×800`（源码窗口默认值）和断点两侧的窄屏；额外视口是否属于正式支持范围由产品决策确定。

严格复刻和交付前对照检查是硬性要求；截图基线、噪声校准和浏览器覆盖尚待实施。视觉一致也包含交互：悬停、焦点、展开、滚动、弹层位置和全屏切换。已知视觉偏差必须修正，环境差异例外须经 Eric 明确确认；不能只通过首页截图签收。

## 9. 许可证与资产记录

**可核验事实**：`package.json` 声明 `AGPL-3.0`，根目录 [LICENSE](../reference/SPlayer/LICENSE) 为 GNU Affero General Public License v3 正文；[README](../reference/SPlayer/README.md#L33) 另列保留版权和原链接等声明，在 [后部说明](../reference/SPlayer/README.md#L291) 还写有学习用途、商业用途限制等文字。这些内容应分别保留记录，不将 README 的所有表述等同于标准许可证条文。

**待核验事项**：直接复制 Vue 模板、CSS、SVG、字体、Logo、图片或二进制 WASM 时，各文件及第三方依赖的授权、署名和分发义务；视觉参考与源码复制两种路径的适用边界；Sonora 自身许可证与发布方式。本文是源码清单，不给出法律结论。公开发布前需完成授权清单，不能把“重写”自动理解为无需检查被复制的资产。

## 10. 可复用与需重写的清单

| 类别 | 建议处理 | 原因 |
| --- | --- | --- |
| 布局、主题和组件样式 | 作为视觉规范；确认授权后选择性复用或逐项实现 | 与“完全按照原项目样式”的要求直接相关 |
| SVG、占位图、字体 | 建立资产台账后决策，保留来源和授权 | 资产可与业务分离，但不是天然无版权 |
| 时间/歌词解析等纯函数 | 审查依赖后可考虑复用，附输入输出用例 | 避免把 Pinia、来源 API 或全局副作用一起带入 |
| 引擎接口、后端能力描述 | 借鉴设计，按首发后端缩小实现 | 现有接口包含很多高级音频能力，不必首发全部实现 |
| 来源 API、认证、数据映射 | 根据自研音乐服务契约重新实现，舍弃网易来源 | 当前共享请求/类型默认网易语义 |
| PlayerController、SongManager、LyricManager | 以明确职责重写编排与来源访问 | 已混合全局状态、来源获取、缓存和桌面 IPC |
| 页面和菜单的数据依赖 | 保留外观，改为应用服务/展示模型 | 原组件会直接调认证、API、播放器单例与 store |
| 存储与迁移 | 设计 Sonora schema 和仓储边界 | 当前多个存储载体与来源字段交错 |
| 桌面桥接 | 舍弃，不迁移到 Sonora | 已确定采用 Web 前端与 Go 后端，不继承 Electron 行为 |
| 原生、mpv、任务栏歌词 | 不纳入 Sonora 计划 | 属于原项目桌面实现；混音等可用于 Web 的能力另行评估 |

本表均为工程建议，不表示已经决定拷贝原实现，也不表示已经批准取消任何用户可见功能。

## 11. 风险与待回答问题

- **范围膨胀**：外观完整与功能完整是两条工作量曲线；应由 roadmap 单独列出首发范围和候选能力，而不是照着路由表逐页全做。
- **来源标识污染**：不同服务器、账号、协议出现相同原始 ID 时，队列、缓存、歌词、喜欢和历史必须仍能区分。
- **异步竞态**：快速切歌、拖动、来源断线和过期播放 URL 应避免旧请求覆盖新状态；原 LyricManager 已有请求序列标记，可作为风险线索。
- **运行方式差异**：Sonora 已确定采用 Web，原桌面扫描、多窗口和原生模块不迁移。文件管理按浏览器权限与 Go 服务职责重新定义。
- **安全与凭据**：源码存在前述宽松窗口配置、浏览器凭据存储及日志；重写应明确凭据所有权、日志脱敏与受限宿主 API。
- **高级音频成本**：FFmpeg、mpv、Automix 各有能力和环境差异；首发后端及音频格式矩阵待决定。
- **视觉基准漂移**：真实首页内容、封面取色、背景、自定义字体、协议弹窗和登录状态会改变截图；需冻结可重复基线。
- **迁移成本**：不确认数据范围就承诺无损迁移，会把旧存储结构和 Cookie 语义重新带入 Sonora。
- **许可与分发**：源码、资产、第三方服务和原生二进制需分别列清单；本文只完成源码层面的发现。
- **现有自动验证不足以证明行为**：package 脚本中能看到 lint/typecheck/build，没有测试脚本声明；不能由此断言仓库完全没有测试，但不应假设已有一套可直接继承的完整行为回归体系。

音乐服务采用自研方案已经确定。仍需明确：首发浏览器与部署方式；自研服务的仓库、现状与接口契约；首发必须可操作的页面；本地扫描是否需要；完整逐字歌词是否首发；设置项和账号数据迁移边界；复用资产的授权路径。这些事项可以在完成视觉基线和自研服务契约验证的同时逐步收敛。
