# 阶段 8：安卓客户端

状态：进行中（2026-11-07 Eric 提出并开始）：参照源码已加入并在 Mac 上构建成功；7 个待决事项已全部确定，工作包已建立，第 1–3 项已完成。前置：阶段 7 退出（2026-11-06）。

## 目标

以 PixelPlayerOSS 为基础二次开发 Sonora 的安卓客户端：连接 Sonora 服务，支持现在 Web 播放端的大部分功能，并增加歌曲缓存（离线播放）。Web 播放端仍只支持桌面版 Chrome，不做移动端适配；手机上使用这个客户端。

## 已确定（Eric，2026-11-07）

- **基础项目**：[PixelPlayerHQ/PixelPlayerOSS](https://github.com/PixelPlayerHQ/PixelPlayerOSS)，不是主仓库 PixelPlayerHQ/PixelPlayer。
- **参照源码**：作为 Git submodule 放在 `reference/` 下（与 `reference/SPlayer` 相同），固定在 tag `v0.4.0-alpha.53`。
- **开发路线**：从参照源码中复制需要的大部分代码出来二次开发，不在原版上删除多余的功能。与 Web 播放端的做法一致：`reference/` 只用于对照，客户端代码不依赖它。
- **范围**：Web 播放端的大部分功能，加上歌曲缓存。具体哪些功能进入、哪些不做，见“待决事项”。
- **许可证**：客户端按 GPL-3.0 发布，在 `android/LICENSE` 单独声明，保留 PixelPlayerOSS 的版权声明并更新 [THIRD_PARTY_NOTICES.md](../../THIRD_PARTY_NOTICES.md)；仓库其余部分仍是 AGPL-3.0。
- **代码位置**：Sonora 仓库的 `android/`。`.gitignore` 已加入对应规则。
- **登录**：参照 asr 的原生客户端登录。客户端打开系统浏览器走同一套 OIDC，按 RFC 8252 带 S256 的 `code_challenge`，安卓用自定义 scheme 回调，服务端签发设备 token；API 同时接受 Cookie 会话与 `Authorization: Bearer`。Go 服务需要新增设备表（迁移 15 起，`scripts/synth-library.mjs` 的 `SCHEMA_VERSION` 同步）与相关接口，接口写入 [api.md](../api.md)。按 asr 的设计，回调仍是 `/auth/callback`，Authelia 不用另外登记；实现时用 Sonora 的配置核实。
- **界面**（2026-11-07 Eric 调整）：完全复刻 PixelPlayerOSS 的视觉风格，标准与 Web 播放端复刻 SPlayer 相同：`reference/PixelPlayerOSS` 的源码与实际渲染是权威基准，布局、配色、字体、字号、间距、尺寸、图标、圆角、阴影、动效和组件状态逐项还原，交付前在相同模拟器、主题、数据和状态下与原版截图对照。不向 SPlayer 靠拢。与原版不同的地方须 Eric 同意，并写在代码注释和本文件中。服务端为客户端渲染的设备登录确认页也按 PixelPlayerOSS 的视觉风格。PixelPlayerOSS 在安卓 12 及以上用壁纸动态取色，网页读不到，所以客户端打开登录页时带上当前主题色（`colors` 参数），确认页照此着色（Eric 选择，2026-11-07）；确认页保留，不去掉（Eric，2026-11-07）。
- **歌曲缓存方式**：手动下载加播放缓存。手动下载的歌曲一直保留，只能手动删除；播放过的歌曲自动缓存，超过容量上限后删除最久没有播放的。
- **功能清单**：曲库浏览、搜索、歌单、喜欢、最近播放、歌词，加漫游 FM。移动端不加入网易云有关的功能（Eric，2026-11-07）。喜欢与最近播放使用现有的 `likes`、`likes/ids`、`history`：它们只读 `tracks` 表，只有曲库歌曲（源码确认，`backend/internal/store/listeners.go`），不需要在客户端过滤，也不需要新增接口。不使用 `likes/mixed`、`history/mixed`，客户端也没有在线歌曲。不包含设置同步，客户端设置只保存在手机上，不写 `PUT settings`（它整体替换，内容是 Web 端的字段，写入会覆盖 Web 端设置）。
- **缓存细节**：只缓存曲库歌曲。播放缓存默认上限 5 GB（可在设置中修改），手动下载的歌曲不计入上限。曲库歌曲的音频按 ID 不变（源码确认，2026-11-07：管理 API 只能修改元数据，没有替换音频的接口；重新上传是新歌曲，ID 不复用），所以缓存不需要校验内容，只需处理歌曲被删除（接口返回 404）；元数据以服务端为准。音频响应现在是 `private, no-cache`，不影响客户端自己管理的缓存，阶段 9 的“音频 immutable 缓存”不是前置条件。
- **发布**：GitHub Release 上的签名 APK，应用内更新，做法照搬 asr。用独立的 `android-v*` tag 触发单独的工作流，客户端发版不重建服务端镜像。需要新建签名密钥与 secrets，由 Eric 处理或确认。
- **验收**：开发中先用模拟器；验收的真机是 PKB110，安卓 16。
- **本地文件播放**：不保留。曲库只来自 Sonora 服务，不申请媒体读取权限，不复制本地扫描、文件夹浏览与元数据编辑的代码。

## 参照项目的事实（2026-11-07 核对）

- **许可证**：PixelPlayerOSS 是 GPL-3.0；主仓库 PixelPlayer 是专有许可，只允许个人非商业的学习与使用，禁止分发衍生作品，所以不能作为基础。复制 PixelPlayerOSS 代码的客户端须按 GPL-3.0 发布，并保留原版权声明；Sonora 现在是 AGPL-3.0，客户端的许可证与 [THIRD_PARTY_NOTICES.md](../../THIRD_PARTY_NOTICES.md) 的写法要在开始前定下来。
- **技术栈**：Kotlin、Jetpack Compose、Material 3、Media3（含 FFmpeg、无缝播放、交叉淡化）；Android 11+（API 30），编译与目标 SDK 37，JDK 21。包名 `com.lostf1sh.pixelplayeross`，需要改为 Sonora 自己的包名。
- **与 Sonora 最接近的部分**：已经接入 Navidrome/Subsonic 与 Jellyfin，有登录、同步、流式播放和应用私有目录的离线下载。相关代码在 `data/navidrome`、`data/jellyfin`、`data/network/*`、`data/worker/*SyncWorker.kt`，离线下载在 `data/offline/CloudOfflineRepository.kt`、`data/worker/CloudTrackDownloadWorker.kt`、`OfflineTrackDao`，以及下载管理页 `CloudDownloadsScreen`。这些是实现 Sonora 曲库同步与歌曲缓存的主要参照。
- **已经去掉的集成**：Telegram、网易云、QQ 音乐、Google Drive、Gemini、Cast、Wear OS、Play 商店计费、Firebase 与 Google Play 服务。
- **构建**（运行确认，2026-11-07，Mac）：`reference/PixelPlayerOSS` 在 `a1f8fce`（`v0.4.0-alpha.53`，与 `v0.4.0-alpha.52` 是同一提交），用 Android Studio 自带的 JBR 21 执行 `./gradlew :app:assembleDebug`，5 分 37 秒成功，产出 arm64-v8a 与 armeabi-v7a 两个 debug APK，各约 155 MB；没有用到 NDK 或 build-tools 37。未安装运行。
- **规模**（运行确认，2026-11-07）：`app/src/main/java` 下 638 个 Kotlin 文件，约 18.2 万行。
- **开发环境**（Eric，2026-11-07）：安卓客户端在 Mac 上开发，那里已有环境，做法参考 `../asr` 项目（听写的安卓与 Mac 客户端）。这台 Linux 开发机没有 JDK 与 Android SDK，不在这里构建。

## 待决事项

全部已确定，见“已确定”。

## 与 PixelPlayerOSS 的差异

- `DualPlayerEngine`：去掉 Navidrome、Jellyfin 流代理、云端 URI 解析与离线查找，媒体 URI 直接交给数据源。第 3、7 项会按 Sonora 的认证与缓存补上数据源。
- `MusicService`：只保留围绕引擎的会话部分（原版 2584 行）。小组件、均衡器、统计、队列持久化、睡眠定时、自定义会话命令与自定义通知都没有复制，需要时再加。后来加上了睡眠定时与 ReplayGain（增益来自服务端，见第 4 项的“音量均衡”）。
- 不带 FFmpeg 解码器：Sonora 只入库 MP3、FLAC、AAC、Vorbis 与 Opus，由系统解码。
- `Transition.kt`：去掉依赖 Room 的 `TransitionRule`。
- `MainActivity` 登录后是临时页面（用户名、测试音播放、退出登录），第 4 项替换。应用名是 Sonora，图标照搬 PixelPlayer 的（Eric，2026-11-07）：启动图标（各密度的自适应图标与单色版）、登录页徽标用 `pixelplayer_base_monochrome`（原版“关于”页的标志），媒体通知的小图标用 `monochrome_player`。确认页上的设备名用手机型号（`Build.MODEL`），Eric 同意。
- 主题（`ui/theme/`）：`Color`、`Shape`、`Type` 照搬，`Theme.kt` 去掉 AMOLED 预设与正在播放取色的覆盖，复制对应设置时再加；字体 `gflex_variable.ttf` 照搬。
- 登录页按原版 `NavidromeLoginScreen` 的布局：只有服务器地址一个输入框（账号在浏览器里输入），没有返回按钮（它是第一个页面），一个圆形徽标放 PixelPlayer 的标志（原版是 Navidrome、Subsonic 两个标志）。文案参照原版的 zh-CN 译文。
- 网络：OkHttp 访问回环地址时不走系统代理，见第 3 项。

## 构建

```sh
export JAVA_HOME="/Applications/Android Studio.app/Contents/jbr/Contents/Home"
export ANDROID_HOME="$HOME/Library/Android/sdk"
cd android && ./gradlew :app:assembleDebug :app:testDebugUnitTest
```

在模拟器上连开发后端：开发服务器上 `task run:server`（`127.0.0.1:14560`），Mac 上 `ssh -f -N -L 127.0.0.1:14560:127.0.0.1:14560 eric`，再 `adb reverse tcp:14560 tcp:14560`，登录页填 `http://127.0.0.1:14560`。

## 工作包

顺序 Eric 已确认（2026-11-07）。实际从第 2 项开始，这样不用先等后端环境；第 3 项连开发服务器时需要第 1 项已部署到那里。

- [x] 把 PixelPlayerOSS 作为 submodule 加到 `reference/PixelPlayerOSS`，检出 `v0.4.0-alpha.53`，在 Mac 上构建原版（运行确认，见“参照项目的事实”）。
- [x] **1. Go 设备 token**：设备表（迁移 15）、设备登录流程、Bearer 认证、设备管理接口，同步 [api.md](../api.md) 和 `SCHEMA_VERSION`（提交 `2b6ba49`）。
  - 运行确认（2026-11-07）：Mac 与开发服务器上 `go test ./...` 通过；开发服务器上 `go test -tags authelia ./internal/auth` 通过，其中 `TestAutheliaDeviceLogin` 在真实 Authelia 4.39.20 下走完设备登录、确认、换 token 并用 Bearer 访问。
  - 与 asr 的差异：只支持安卓（回调 `sonora://auth/callback`），token 前缀 `sonora_d_`，确认页没有头像，没有推送通道；Bearer 请求的用户带 Authelia 组，`canImport` 与网页一致；设备的时间按其他 `createdAt` 字段用 RFC 3339 字符串返回。
  - 不做（Eric，2026-11-07）：网页端的设备管理与吊销。`GET devices` 与吊销自己的 `DELETE devices/{id}` 保留，客户端退出登录时用后者。
- [x] **2. 工程骨架**：建 `android/`，包名 `top.onlyquant.sonora`，加 `android/LICENSE`，更新 [THIRD_PARTY_NOTICES.md](../../THIRD_PARTY_NOTICES.md)，只复制播放内核，能启动并播放一首固定的歌（debug 版打包的 30 秒测试音）。
  - 运行确认（2026-11-07，模拟器 `Medium_Phone`，API 36，无音频输出）：`dumpsys media_session` 显示点击播放后进入 `PLAYING`；切到后台后进度从 3059 ms 走到 9078 ms；媒体键暂停后两次读数都停在 9214 ms；恢复后走到 12324 ms；服务在前台并有媒体通知；logcat 没有崩溃或 `ExoPlaybackException`。
  - 运行确认（2026-11-07）：`./gradlew :app:assembleDebug :app:testDebugUnitTest` 成功，debug APK 约 41 MB，复制来的 74 个单元测试全部通过。
  - 待验证：实际听感（模拟器没有音频输出）；PKB110 真机；除 MP3 外的格式实际播放（模拟器上有 AAC、FLAC、Opus、Vorbis 的系统解码器）。
- [x] **3. 登录与 API 层**：系统浏览器登录、自定义 scheme 回调、token 保存、API 客户端。
  - 运行确认（2026-11-07，模拟器 `Medium_Phone`，API 36，连开发服务器的开发模式后端）：登录页用 Custom Tab 打开设备登录，确认页带上 App 当前的动态配色；点“允许”后 Chrome 跳回 `sonora://auth/callback`，App 换到 token 并用 Bearer 调 `GET me` 得到 `listener`；重启 App 后保持登录；“退出登录”吊销自己，服务端 `devices` 表清空；确认页点“取消”回到登录页；服务端删掉设备后重启 App，`GET me` 返回 401，App 回到登录页并提示“登录已失效，请重新登录”。`./gradlew :app:testDebugUnitTest` 通过（PKCE、服务器地址、配色参数与代理选择的单元测试）。
  - 待验证：真实 Authelia 下的登录（开发模式跳过 OIDC）；PKB110 真机上 Custom Tab 的跳回；登录页与原版的逐项截图对照（布局已与原版 Subsonic 登录页并排看过，未逐像素对照）。
  - 访问回环地址（`localhost`、`127.x`、`::1`）不走系统代理，与 Chrome 一致。模拟器的全局代理（`10.0.2.2`）会把 `127.0.0.1` 的请求转到别处，开发时曾因此连到一台使用真实 Authelia 的服务，没有在那个页面上输入或操作。
- [ ] **4. 曲库与播放**：首页、歌曲、专辑、艺术家、搜索、播放器、队列、后台播放和媒体通知。验证：模拟器运行。
  - 首页（Eric，2026-11-08）：沿用原版 `HomeScreen` 的版式，换内容：封面拼贴放最近添加的专辑（`albums?sort=createTime&order=desc&hideSingles=true`），Daily Mix 卡片放最近添加的歌曲（前 6 首，`onlySingles=true`，与 Web 端相同），最近播放用 `history`；去掉年度回顾与听歌统计（原版依赖本机播放统计，服务端没有）；标题区的随机播放保留，播放范围做首页时再定。
  - 首页的调整（Eric，2026-11-11，取代上一条的封面拼贴与 Daily Mix 卡片）：标题区和封面拼贴改为漫游 FM：大标题“漫游电台”（照原版两行排）、副标题“Your Roaming Mix”，标题右侧的按钮图标由随机播放换成收音机（`Icons.Rounded.Radio`），点了开始播放漫游 FM（第 6 项因此提前到这里做）；封面拼贴显示 FM 接下来真正要放的专辑，每格一张（原版默认的拼贴样式有 5 格，取 5 张）（服务端 `GET fm` 新增 `count` 参数），点封面打开专辑页；不做每日精选卡片；顶栏只留设置按钮（设置页做好前点了没有反应），不要侧边栏（原版的侧边栏菜单、构建版本标签、云端串流、更新日志都去掉）。最近播放照原版，用 `GET history`。
  - 曲库数据（Eric，2026-11-08）：整库同步到本机 Room，页面从本机读，与原版 Navidrome/Jellyfin 的 SyncWorker 相同；打开 App 时与下拉刷新时同步；搜索在线调接口，离线时退回本机的子串匹配。
  - 增量变更（Eric，2026-11-08）：原来的接口没有增量变更（源码确认：只有 `tracks` 有 `updated_at`，删除不留墓碑，写入路径有入库、导入、补全、编辑、合并、删除与 `pruneOrphans`），服务端新增迁移 16：变更日志表 `library_changes`，每个条目一行、每次变化取新的递增序号，由 SQLite 触发器维护并按依赖扩散（艺术家改名记他的歌曲与专辑，歌曲进出专辑记该专辑），删除也记录；新增 `GET library/changes?since=` 按序号返回变更的完整对象与删除的 ID。客户端本机排序用服务端给的排序键（与服务端 `zh_numeric` 同一规则），顺序与 Web 端一致。接口见 [api.md](../api.md#曲库变更)。
    - 源码确认与运行确认（2026-11-08，Mac）：`go test ./...` 通过；新测试覆盖首次同步、艺术家改名扩散到歌曲与专辑、专辑封面扩散到歌曲、歌词只改歌曲、歌曲换专辑与艺术家、`MergeArtists`（经 `UPDATE OR IGNORE`，触发器因此用先删后插而不是 `INSERT OR REPLACE`）、删除与清理孤立条目、分页、`reset`、从第 15 版带数据迁移，以及排序键与 SQL 排序顺序一致。
    - 运行确认（2026-11-08，开发服务器，提交 `70833eb`）：迁移 16 在开发库上执行成功，`go test ./internal/store ./internal/server` 通过；开发库的 369 首歌、311 张专辑、215 位艺术家按每页 500 条分 2 页取完，cursor 为 895（三者之和），再以该 cursor 请求为空且 cursor 不变；`since=-1` 返回 400。Mac 上用 2,000 首合成数据临时测得：全量同步约 19 ms，艺术家改名约 0.5 ms，之后的增量读取约 0.3 ms。
    - 待验证：线上曲库（1,632 首）上的迁移与同步。
  - 顺序：播放器面板与队列 → 曲库 → 详情页 → 首页 → 搜索。五项都已复制（搜索 2026-11-14）。曲库有歌曲、专辑、艺术家、歌单四个标签（2026-11-11 先定三个，同日 Eric 加上歌单，对应 Web 端的本地歌单）。
  - 播放器面板与队列（Eric，2026-11-10）：保留原版的睡眠定时（含播完本首、播放 N 首后停止）、播放速度与音频输出选择；去掉音频书签、歌曲信息页、手动导入歌词与“队列存为歌单”（歌单在第 5 项）。
    - 已复制（提交 `b124494`、`fd392d9`）：`UnifiedPlayerSheetV2` 及其 `components/scoped/`、`components/player/`、`QueueBottomSheet`、睡眠定时（`SleepTimer*`，在 `MusicService` 里执行）、播放速度、音频输出选择，以及承载面板的底部导航栏与外壳（原版 `MainActivity.MainUI`，在 `presentation/main/MainUI.kt`）。`PlayerViewModel` 精简为约 1,100 行（原版 5,191 行），方法体照原版。音频经 `media3-datasource-okhttp`、封面经 Coil 与通知的 `CoilBitmapLoader` 都走同一个 OkHttpClient，由拦截器给本服务器的请求加 device token。
    - 与原版的差异：歌词按钮在，点了没有反应（第 5 项接上）；喜欢按钮不生效（第 5 项）；去掉书签按钮、队列里每行的“更多”按钮（只用来打开歌曲信息页）与“存为歌单”；全屏播放器标题旁的云朵图标（原版标在流媒体歌曲上，Sonora 的歌都是流，Eric 2026-11-10 决定去掉）；封面取色只缓存在内存（原版存 Room）；外壳没有宽屏的导航栏、侧边抽屉、按页面隐藏导航栏与 SmartMix；导航栏的设置项取原版默认值。首页标签暂时是一个按入库时间排的歌曲列表（`GET tracks` 前 500 首，点一首用整个列表作队列），搜索和音乐库标签暂时空白，做曲库时替换。
    - 运行确认（2026-11-10，模拟器 `Medium_Phone`，开发服务器 `2f4b8eb`）：`./gradlew :app:assembleDebug :app:testDebugUnitTest` 通过；登录后列表显示开发库的歌，点第一首后 `dumpsys media_session` 为 `PLAYING`、缓冲到 76 s，迷你播放器和全屏播放器显示封面并按封面取色，全屏播放器显示 `48.0 kHz • 320 kbps • MP3`；队列显示 369 首，点队列里的歌切换播放；播放速度选 1.5× 后会话的 `speed=1.5`，改回 1×；“当前歌曲播放完毕”开关打开后服务端（`MusicService`）提示“将在当前歌曲播放完毕后停止”。
    - 截图对照（2026-11-10，同一模拟器、深色、原版应用语言设为 zh-CN；Eric 从管理后台下载的 3 首歌导入原版，Sonora 播放开发库里的同 3 首，队列顺序相同，暂停后截图，逐像素比较）：迷你播放器一致；全屏播放器除了去掉的书签按钮（音频输出按钮因此右移一格）和暂停时的进度外一致；队列除了去掉的每行“更多”按钮（当前歌曲的播放动画因此右移）和封面图片本身外一致。封面不同是因为原版读文件里嵌入的封面，Sonora 用服务端的封面。原版曲库按码位排标题（中学时代、藏心、韶华渡），Sonora 服务端按拼音排（藏心、韶华渡、中学时代），曲库沿用服务端顺序（2026-11-08 的决定）。
    - 运行确认（2026-11-10，模拟器 `Medium_Phone`，提交 `fd392d9` 的 debug 版，队列为中学时代、藏心、韶华渡，状态取自 `dumpsys media_session` 与 logcat 的 `SleepTimerController`）：
      - 拖动排序：把“韶华渡”拖到“藏心”上方后，队列显示新顺序，按“下一首”切到“韶华渡”。队列只显示当前歌曲及之后的歌。
      - 滑动移除：左滑一行后会话队列从 3 首变为 2 首，底部提示“藏心 已移除 · 撤销”；4 秒内点“撤销”，队列恢复为 3 首，歌曲回到原位置。提示 4 秒后消失（`QueueUndoStateHolder`，照原版）。
      - 清空队列：“⋯”菜单里点“清空队列”先弹确认框（“确定要清空队列内除正在播放以外的全部歌曲吗？”），确认后队列只剩当前歌曲。
      - 迷你播放器滑动关闭与撤销：右滑迷你播放器后会话变为 `NONE`、队列清空，底部提示“歌单已撤销 · 撤销”；4 秒内点“撤销”，3 首队列恢复，进度回到关闭前的 3840 ms，提示“播放列表已恢复”。恢复后直接开始播放（照原版 `undoDismissPlaylist` 的 `playSongs`）。文案取自原版 zh-CN。
      - 音频输出：音箱按钮打开“播放设备”面板，列出“本机”（已选中）和“更多设备”；点“更多设备”打开系统输出选择器（`MediaOutputDialog`），里面显示的是 Sonora 的会话与当前歌曲。模拟器只有本机输出，没有测切换到其他设备。
      - 定时停止：“当前歌曲播放完毕”在“中学时代”结尾暂停，停在下一首“藏心”的 0 ms，不自动播放；“播放次数”设为 2 次时先单曲循环一遍，第二遍结尾暂停并回到 0 ms，循环按钮恢复为原来的关闭状态；自定义时长设为 1 分钟，23:18:34.540 设定，23:19:34.539 到期并暂停。
      - 整个过程 logcat 的 crash 缓冲区为空，App 进程没有错误日志。
    - 待验证：切换到其他音频输出设备（蓝牙等，PKB110 上）；实际听感。
  - 曲库（Eric，2026-11-11）：照原版 `LibraryScreen`，有歌曲、专辑、艺术家、歌单四个标签（歌单对应 Web 端的本地歌单，内容在第 5 项），数据来自本机 Room（用 `library/changes` 同步）。
    - 歌曲的“⋮”菜单只留播放、喜欢、添加到队列、下一首、加入歌单（喜欢与歌单在第 5 项接上）；去掉分享、删除、设置提示音、在 MusicBrainz 上查找、编辑与“详情”页。
    - 长按多选保留，批量操作只留添加到队列、下一首、加入歌单、喜欢，去掉删除与批量编辑。
    - 标题右边去掉书签按钮（音频书签已去掉），保留设置按钮，设置页做好之前点了没有反应。
    - 标签行去掉最后的“编辑标签”铅笔按钮（调整标签顺序），标签固定为默认顺序。
    - 操作行的存储筛选按钮（全部 / 云端 / 本机）保留，暂时点了没有反应。
    - 多选面板保留“全部播放”。
    - 排序面板去掉“云端 · 仅云端”开关（Sonora 的歌都在服务器上）；“默认顺序”用服务端列表的默认顺序：歌曲新入库在前，专辑按标题，艺术家按名称，与 Web 端一致。
    - “专辑”标签与 Web 端的专辑页一样隐藏标记为单曲的专辑（`hideSingles`，Eric，2026-11-11）。
    - 本机曲库（提交 `f799103`）：Room 数据库 `SonoraDatabase`，表名沿用原版（`songs`、`albums`、`artists`、`song_artist_cross_ref`），列按 Sonora 的接口，另有 `library_state` 记录服务器地址与 cursor。没有外键（一页变更里歌曲可能先于它的专辑与艺术家）；艺术家与专辑的 `sortKey` 只存在各自的表里，按艺术家、专辑排序歌曲时联表取。`LibraryRepository.sync()` 逐页取 `library/changes`，每页连同 cursor 在一个事务里写入；服务器地址变了或服务端返回 `reset` 时清空重来。打开 App（已登录）时同步一次。排序照原版 `SortOption` 的键在 SQL 里排，名称按 `sortKey`，“添加时间”按 ID（服务端按入库顺序分配、不复用；播放端 API 没有入库时间）。上线前改表结构直接删库重新同步（`fallbackToDestructiveMigration`），发布后照原版写迁移。
      - 运行确认（2026-11-11，模拟器，开发服务器 `2f4b8eb`）：首次同步后本机有 369 首歌、311 张专辑、215 位艺术家、420 条歌曲与艺术家的关系，cursor 897，与开发库的 `tracks`、`albums`、`artists`、`track_artists` 行数和 `library_changes` 最大序号相同；本机 SQL 排出的顺序与服务端列表逐项相同：歌曲的 `default`、`title`（升降序）、`artist`、`album`、`duration`、`createTime`，专辑的 `default`、`hideSingles=true`、`createTime` 降序，艺术家的 `default`。`./gradlew :app:testDebugUnitTest` 81 个测试通过，其中新增 3 个测试变更到本机行与 `Song` 的转换。
      - 待验证：增量同步（服务端改名、删除后再同步）、`reset`、断网时同步失败的提示。
    - 曲库界面（提交 `d9b53b2`）：整份复制原版 `LibraryScreen`（只留主函数、同步提示、`AlbumGridItemRedesigned`、`ArtistListItem`、`AlbumListItem`）、`LibrarySongsTab`、`LibraryMediaTabs`（专辑、艺术家）、`LibraryEmptyState`、`LibraryPlaybackAwareSongItem`、`LibraryScrollRestore`、`LibrarySortScrollReset`、`TabAnimation`、`LibrarySortBottomSheet`、`MultiSelectionBottomSheet`、`SongInfoBottomSheet`、`EnhancedSongListItem`、`LibraryActionRow`、`SelectionActionRow`、多选（`presentation/selection/`）、`MultiSelectionStateHolder`、`LibraryTabsStateHolder`、`SortOption`、`LibraryTabId` 等；`LibraryStateHolder` 按原版缩减，排序选项照原版的键保存到 `UserPreferencesRepository`，分页与列表从 `LibraryRepository` 取；`PlayerViewModel` 加上原版的曲库部分（点歌以排序后的整个曲库为队列、随机播放、随机专辑、随机艺术家、定位当前歌曲、批量操作）。文案在 `res/values/strings_library.xml`（原版 zh-CN）。缩减处在代码里写了“Sonora: …”。
      - 与原版的差异：没有文件夹、我喜欢、流派标签和紧凑导航（原版的设置项，默认是标签行）；歌单标签暂时是空白页（原版的 `PlaylistContainer` 与空状态在第 5 项），“新建”点了没有反应，没有 M3U“导入”按钮，排序面板没有歌单的布局切换；歌曲菜单只有选项页，没有分享、删除、设置提示音、MusicBrainz、编辑按钮和详情页，因此也没有底部的“选项 / 详情”切换条（原版为它留的 80dp 空白保留，菜单高度不变）；菜单和多选面板第一行去掉分享后，“播放”和“喜欢”各占一半，喜欢按钮改成和播放按钮一样带文字（“喜欢”）的胶囊形（Eric，2026-11-11；原版只有图标）；多选面板没有批量编辑、下载与“全部删除”，“播放列表”一行只剩一个按钮；同步提示只显示“正在同步”（原版按扫描阶段显示），首次同步的遮罩没有进度条（`library/changes` 没有总数）；喜欢与加入歌单点了没有反应（第 5 项）。
      - 运行确认（2026-11-11，模拟器 `Medium_Phone`，开发服务器 `2f4b8eb`）：`./gradlew :app:assembleDebug :app:testDebugUnitTest` 通过（81 个测试）；四个标签显示，专辑标签只显示非单曲专辑；点“Angel”后会话为 `PLAYING`、队列 369 首、位于第 2 首（与标题排序一致）；当前歌曲滚出视野后出现定位按钮，点后滚回；排序改为“默认顺序”后前 3 首与 `GET tracks` 相同，再改回“标题”；歌曲标签“随机播放”开始 369 首的随机队列；长按多选，返回键退出多选；两张专辑多选后“⋮”面板显示 34 首（32 + 2）；下拉刷新触发同步（logcat `Library synced to cursor 897`）；全程 crash 缓冲区为空。
      - 截图对照（同一模拟器、深色、原版 zh-CN，原版曲库 3 首、Sonora 369 首）：专辑标签的头部（标题、标签行、操作行）只有去掉的书签按钮和原版第 5 个标签不同；歌曲排序面板（拖动条到“时长”）与专辑排序面板（拖动条到“视图”切换）按面板顶部对齐后逐像素一致，原版下面的“云端”一节已按决定去掉；歌曲菜单的拖动条和“添加到队列 / 下一首”一行、多选面板的“添加到队列 / 下一首”一行逐像素一致。列表宽度不同是因为 Sonora 的列表可滚动、显示滚动条（原版 3 首不滚动），原版按是否可滚动留出滚动条的位置。
      - 待验证：专辑列表视图、艺术家排序、PKB110 真机；相同数据下列表的逐像素对照（原版只有 3 首，Sonora 不能只显示这 3 首）。
  - 专辑与艺术家详情页（提交 `530dbd9`）：整份复制原版 `AlbumDetailScreen`、`ArtistDetailScreen`、`AlbumDetailViewModel`、`ArtistDetailViewModel`、`ArtistAlbumSectionExpansion`、`MediaDetailHero`、`CollapsibleCommonTopBar`、`ExpressiveTopBarContent`、`ExpressiveDropdownMenu`、`ScreenWrapper`、`Transitions`、`MainRootRoutes`；`AppNavigation` 只留首页、搜索、音乐库与两个详情页，转场照原版；`MainUI` 照原版在详情页隐藏底部导航栏。数据从本机 Room 取（`LibraryRepository.album`、`artist`、`songsForAlbumFlow`、`songsForArtistFlow`，同步后自动更新）。文案在 `res/values/strings_detail.xml`（原版 zh-CN）。
    - 与原版的差异：艺术家页去掉右上角的编辑按钮（“更换照片 / 恢复默认”），艺术家图只用服务端的（Eric，2026-11-11），收起后标题右侧因此只留 24dp；专辑页的“下载专辑”按钮保留，点了没有反应，第 7 项接上（Eric，2026-11-11），Sonora 的歌都在服务器上，所以每张专辑都显示；没有流派（服务端没有），专辑页的流派与年份一行只有年份，艺术家页的专辑行没有流派；歌曲“⋮”菜单与曲库相同；“添加到歌单”与喜欢点了没有反应（第 5 项）；专辑内与艺术家的歌曲按碟号、曲号，之后按服务端 sortKey（原版按小写标题）；原版两个页面里留着的旧顶栏（`UseSharedCollapsibleTopBarProbe` 恒为 true 时不用）没有复制。
    - 运行确认（2026-11-11，模拟器 `Medium_Phone`，开发服务器 `2f4b8eb`）：`./gradlew :app:assembleDebug :app:testDebugUnitTest` 通过（81 个测试）；从专辑标签点“F1 The Album (Cinematic Edition)”进入专辑页，按碟分组（第 1 碟 17 首、第 2 碟 15 首，与本机库相同）；点“播放”后会话为 `PLAYING`、队列 32 首；上滑后头部收起、标题停靠到顶栏；“更多操作”菜单显示添加到歌单、下一首播放、添加到队列；从艺术家标签进入“李健”，11 张专辑分节显示，“全部收起”后歌曲行收起、按钮变为“全部展开”，点某张专辑的播放按钮后从该专辑播放；从全屏播放器点封面进入专辑页；返回键回到上一页；详情页上底部导航栏隐藏；全程 crash 缓冲区为空。
    - 截图对照（同一模拟器、深色、原版 zh-CN，数据相同：周笔畅的“锦绣南歌 电视剧原声带”，1 首“藏心”，两边都播放“藏心”后暂停）：艺术家页逐像素比较，只有状态栏时间、去掉的编辑按钮和专辑行末尾的“• C-Pop”不同；专辑页只有状态栏时间、“下载专辑”按钮（原版本机歌曲不显示）、流派与年份一行（原版“C-Pop · 2020”）和标题字形不同。标题字形不同是因为原版按专辑 ID 选可变字体的参数（`GenreTypography.getGenreStyle`），两边的 ID 不同；封面相同。
    - 专辑页的艺术家胶囊（标题下方）的头像和点击跳转改用专辑艺术家（专辑的第一位艺术家，Eric，2026-11-11）；原版取第一首歌的艺术家，合辑上名字是专辑艺术家、点进去却是第一首歌的艺术家。运行确认（2026-11-11，模拟器）：“F1 The Album (Cinematic Edition)”点“Hans Zimmer”进入 Hans Zimmer 的页面（该专辑 15 首）。
    - 待验证：有艺术家图时的头部（开发库 215 位艺术家都没有图）、PKB110 真机。
  - 首页与漫游 FM（2026-11-11，服务端提交 `9ff653f`，客户端提交 `2244697`）：
    - 服务端：`GET fm?count=N`（1–10，默认 2，超出范围 400），返回从当前专辑起的 N 张，这些专辑是真正排进计划的；没有分组歌单时一次抽取只排一张大专辑，所以抽取次数随 N 增加。`store.FMAhead`，`FM` 不变；测试 `TestFMAheadLinesUpMoreAlbums` 与 `player_test` 的参数校验。接口见 [api.md](../api.md#漫游-fm)。
    - 客户端：复制原版 `HomeScreen`（按上面的决定缩减）、`AlbumArtCollage`、`CollagePatterns`、`CollagePattern`、`RecentlyPlayedSection`、`GradientTopBar` 的 `HomeGradientTopBar`、`RecentlyPlayedSongUiModel`、`RoundedStarShape`；首页数据在 `HomeViewModel`（`GET fm?count=5` 与 `GET history`），首页显示时和 FM 换专辑时重新读取。最近播放长按用曲库相同的缩减菜单；“全部”箭头暂时没有反应（最近播放页在第 5 项）；首页的空状态副标题用 Web 端的“还没有可以播放的专辑”。临时的歌曲列表页（`TrackListPlaceholder`）去掉；退出登录暂时放在首页的设置按钮后面（点了弹出确认框），设置页做好后移过去（Eric，2026-11-11 同意）。
    - 漫游 FM 播放照 Web 端（`PlayerController` 的漫游 FM 部分）：FM 的歌放进普通队列，用队列标签 `sonora_fm` 标出，用户插入的“下一首播放”不带标签；离开一张专辑的最后一首去下一首时 `finish`，并把服务端排好的后续专辑接到队列后面；每换到一首 FM 的歌就 `PUT fm/position`，请求按顺序发出。队列后面没有歌（服务端没能及时排出下一张）时，按“下一首”或播完都先等服务端再切过去，取不到时提示“暂时取不到下一张专辑”（Web 端的文案）。开始 FM 时关掉随机播放与循环（Web 端在 FM 里没有这两个按钮，这里按钮还在）。纯函数在 `data/fm/RoamingFm.kt`，有单元测试。
    - 运行确认（2026-11-11，模拟器，开发服务器上用临时目录编译的未提交后端 `/tmp/sonora-fm/sonora-dev`）：`./gradlew :app:assembleDebug :app:testDebugUnitTest` 通过（84 个测试）；首页显示漫游电台标题区、拼贴与最近播放（服务端历史）；点收音机按钮后会话为 `PLAYING`、队列为 FM 的专辑；换歌后服务端 `fm_plan.track_id` 随之更新；从服务端记录的位置继续播放；在专辑最后一首按“下一首”后 `fm_albums.plays` 加 1、队列接上后续专辑并切到它的第一首；FM 播放时按钮变为高亮色；点拼贴封面进入专辑页，返回回到首页；crash 缓冲区为空。
    - 运行确认（2026-11-11，同上，开发库给听者 `listener` 建了分组歌单“漫游分组（开发测试）”，含 10 张小专辑各一首，并设为 FM 分组）：拼贴显示 5 张不同的专辑；F1 最后一首按“下一首”后 F1 的 `plays` 加 1，切到下一张“音乐傲骨”的第一首，队列接上再下一张“我去2000年”的 3 首；设置按钮弹出退出登录确认框，取消后关闭。
    - 待验证：播放到队列末尾自然结束时的衔接（按“下一首”的路径已验证）；PKB110 真机。
  - 搜索（Eric，2026-11-14，提交 `8c4eac4`）：复制原版 `SearchScreen`（替换占位的 `BlankTab`）与 `SearchStateHolder`，在 `AppNavigation` 的 Search 路由接上。
    - 搜索框为空时原版显示流派网格，Sonora 没有流派，留空，只有搜索框；不记录搜索历史（原版提交时记录，但界面上没有地方显示）。
    - 筛选标签照原版：全部、歌曲、专辑、艺术家、歌单。“歌单”保留，歌单结果在第 5 项接上 `GET playlists?q=`，在那之前这一分区始终为空。
    - 只有歌曲能长按多选，“⋮”菜单与多选面板和曲库相同；去掉专辑、歌单与流派的多选。专辑结果包括单曲专辑（与 Web 端搜索相同）。
    - 搜索框右边的设置按钮与首页相同，弹出“退出登录”确认框，设置页做好后移过去。
    - 双击底部导航的“搜索”聚焦输入框并弹出键盘，照原版在 `MainUI` 接上。
    - 数据：在线时并发请求 `tracks`、`albums`、`artists` 的 `q`（`limit=500`），按服务端的相关度；请求失败时退回本机 Room，`rankLocalMatches` 照服务端 `search.go` 的规则匹配与排序（各层字段依次比较相等、开头、包含，ASCII 不区分大小写，同档保持列表默认顺序）。点歌曲时以全部歌曲结果为队列播放（照原版）。
    - 运行确认（2026-11-14，模拟器 `Medium_Phone`，开发服务器 `e683d3c`）：`./gradlew :app:assembleDebug :app:testDebugUnitTest` 通过（94 个测试）；搜索框为空时只显示搜索框；搜“hans”得到 15 首歌、1 张专辑、1 位艺术家，顺序与 `curl tracks?q=hans` 一致；断开 Mac 上的 ssh 转发后再搜，logcat 记录“无法连接服务器”并退回本机，歌曲顺序与在线时相同，专辑与艺术家标签也有结果；点 Red Flag 后播放，会话 `active item id=2` 与它在歌曲结果中的位置一致；专辑结果打开专辑页；艺术家的播放按钮开始播放该艺术家的歌；“⋮”菜单与曲库相同（播放、喜欢、添加到队列、下一首，以及加入歌单的“播放列表”）；长按进入多选，返回键清除选择；设置按钮弹出退出登录确认框，取消后关闭；双击“搜索”后键盘弹出；crash 缓冲区为空。
    - 截图对照（同一模拟器、深色、原版应用语言 zh-CN）：搜不到结果（“zzzq”）时，筛选标签、空状态图标与文案与原版逐像素比较，差值不超过 3（RGB 各通道）；不同只有原版底部恢复的迷你播放器（原版有队列持久化）。
    - 待验证：结果列表与原版的逐像素对照。原版曲库只有三首中文歌，adb 只能输入 ASCII，而唯一含 ASCII 的“闻人听書_”在原版里被输入法转成了全角“＿”，两边找不到相同的结果；结果行的组件照原版原样复制。
  - 音量均衡（2026-11-14，提交 `407a0fa`；服务端与 Web 端见提交 `4e9cf69`、`68c8cda`，接口见 [api.md](../api.md#音量均衡)）：歌曲带服务端算好的 `gainDb`，`library/changes` 同步到本机 `songs.gain_db`（`SonoraDatabase` 第 2 版，升级时删库重新同步），经 MediaItem 的 extras 交给播放器。
    - 施加方式（已定）：用 `player.volume`，不用 AudioProcessor（引擎默认开 audio offload，开启后处理器不运行；处理器在缓冲之后，增益会偏离歌曲边界）。增益大于 0 dB 时截到 1（开发库 369 首里 10 首需要调大，最多 +4.5 dB）。在线歌曲不做均衡，安卓还没有在线歌曲，`normalizationVolume` 留了这个参数。
    - 复制原版 `ReplayGainProcessor`，增益来源由读文件标签（`ReplayGainManager`，异步读取、缓存、预取）换成 extras 里的 `gainDb`，同步取得；去掉“使用专辑增益”开关（服务端每首只有一个增益）；关掉开关时清掉上次的增益，原版在关掉后恢复播放会把旧增益设回去。`MusicService` 的调用点照原版：换播放器、交叉淡入淡出开始与结束、`onVolumeChanged`、开始播放、跳转、换歌、元数据变化、开关变化。
    - 开关（Eric，2026-11-14）：偏好项 `replaygain_enabled`，默认开启（原版默认关闭，与 Web 端的“音量平衡”一致），这次不做界面。设置页做好时照原版放进“音量标准化(ReplayGain)”分组，开关“启用 ReplayGain”取原版文案，副标题改为说明按服务端测得的响度统一音量，不放专辑增益开关。
    - 运行确认（2026-11-14，模拟器 `Medium_Phone`，开发服务器 `68c8cda`）：`./gradlew :app:assembleDebug :app:testDebugUnitTest` 通过（89 个测试）；升级后本机库为第 2 版，369 首都有 `gain_db`，与服务端一致；播放 Red Flag 时 logcat `ReplayGain` 为 volume 0.9354，按“下一首”到 Three Laps Is A Lifetime 为 0.7047，`dumpsys media.audio_flinger` 里这条 AudioTrack 的 L/R dB 为 -3；暂停后继续、回到本曲开头仍为 -3；把 `replaygain_enabled` 写成 false 后重启 App，播放 Red Flag 时 L/R dB 为 0；恢复设置后 Three Laps 回到 -3。
    - 待验证：交叉淡入淡出结束时音量不跳变（`TransitionController` 还没复制，目前没有交叉淡入淡出，钩子已照原版接好，只是源码确认）；播放中切换开关立即生效（没有界面，只源码确认）；播放器重建（audio offload 回退）后的音量；实际听感。
- [ ] **5. 用户数据**：喜欢、歌单、最近播放、歌词。验证：模拟器运行，并与 Web 端核对数据。
- [ ] **6. 漫游 FM**。验证：模拟器运行。播放部分随首页提前做了（见第 4 项“首页与漫游 FM”）；安卓不做跳过专辑（Eric，2026-11-11）；还没有分组歌单的设置。
- [ ] **7. 歌曲缓存**：手动下载、播放缓存（默认 5 GB，超出后删最久没播放的）、下载管理页、已删除歌曲的清理、离线播放。验证：模拟器断网测试。
- [ ] **8. 发布**：签名密钥、`android-v*` 工作流、应用内更新。验证：PKB110 上从 Release 安装并覆盖升级，再走一遍全部功能。
