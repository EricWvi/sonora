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
- `MusicService`：只保留围绕引擎的会话部分（原版 2584 行）。小组件、均衡器、统计、队列持久化、睡眠定时、自定义会话命令与自定义通知都没有复制，需要时再加。
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
  - 曲库数据（Eric，2026-11-08）：整库同步到本机 Room，页面从本机读，与原版 Navidrome/Jellyfin 的 SyncWorker 相同；打开 App 时与下拉刷新时同步；搜索在线调接口，离线时退回本机的子串匹配。
  - 增量变更（Eric，2026-11-08）：原来的接口没有增量变更（源码确认：只有 `tracks` 有 `updated_at`，删除不留墓碑，写入路径有入库、导入、补全、编辑、合并、删除与 `pruneOrphans`），服务端新增迁移 16：变更日志表 `library_changes`，每个条目一行、每次变化取新的递增序号，由 SQLite 触发器维护并按依赖扩散（艺术家改名记他的歌曲与专辑，歌曲进出专辑记该专辑），删除也记录；新增 `GET library/changes?since=` 按序号返回变更的完整对象与删除的 ID。客户端本机排序用服务端给的排序键（与服务端 `zh_numeric` 同一规则），顺序与 Web 端一致。接口见 [api.md](../api.md#曲库变更)。
    - 源码确认与运行确认（2026-11-08，Mac）：`go test ./...` 通过；新测试覆盖首次同步、艺术家改名扩散到歌曲与专辑、专辑封面扩散到歌曲、歌词只改歌曲、歌曲换专辑与艺术家、`MergeArtists`（经 `UPDATE OR IGNORE`，触发器因此用先删后插而不是 `INSERT OR REPLACE`）、删除与清理孤立条目、分页、`reset`、从第 15 版带数据迁移，以及排序键与 SQL 排序顺序一致。
    - 待验证：开发服务器上的运行与真实曲库规模下的耗时。
  - 顺序：播放器面板与队列 → 曲库 → 详情页 → 首页 → 搜索。曲库只留歌曲、专辑、艺术家三个标签。
- [ ] **5. 用户数据**：喜欢、歌单、最近播放、歌词。验证：模拟器运行，并与 Web 端核对数据。
- [ ] **6. 漫游 FM**。验证：模拟器运行。
- [ ] **7. 歌曲缓存**：手动下载、播放缓存（默认 5 GB，超出后删最久没播放的）、下载管理页、已删除歌曲的清理、离线播放。验证：模拟器断网测试。
- [ ] **8. 发布**：签名密钥、`android-v*` 工作流、应用内更新。验证：PKB110 上从 Release 安装并覆盖升级，再走一遍全部功能。
