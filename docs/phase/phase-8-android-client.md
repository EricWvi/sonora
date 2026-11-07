# 阶段 8：安卓客户端

状态：进行中（2026-11-07 Eric 提出并开始）：参照源码已加入并在 Mac 上构建成功；7 个待决事项已全部确定，工作包已建立，第 1、2 项已完成。前置：阶段 7 退出（2026-11-06）。

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
- `MainActivity` 是临时的单按钮页面，第 4 项替换。暂时没有应用图标（新视觉方案须 Eric 同意）。

## 构建

```sh
export JAVA_HOME="/Applications/Android Studio.app/Contents/jbr/Contents/Home"
export ANDROID_HOME="$HOME/Library/Android/sdk"
cd android && ./gradlew :app:assembleDebug :app:testDebugUnitTest
```

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
- [ ] **3. 登录与 API 层**：系统浏览器登录、自定义 scheme 回调、token 保存、API 客户端。验证：模拟器连开发服务器。
- [ ] **4. 曲库与播放**：首页、歌曲、专辑、艺术家、搜索、播放器、队列、后台播放和媒体通知。验证：模拟器运行。
- [ ] **5. 用户数据**：喜欢、歌单、最近播放、歌词。验证：模拟器运行，并与 Web 端核对数据。
- [ ] **6. 漫游 FM**。验证：模拟器运行。
- [ ] **7. 歌曲缓存**：手动下载、播放缓存（默认 5 GB，超出后删最久没播放的）、下载管理页、已删除歌曲的清理、离线播放。验证：模拟器断网测试。
- [ ] **8. 发布**：签名密钥、`android-v*` 工作流、应用内更新。验证：PKB110 上从 Release 安装并覆盖升级，再走一遍全部功能。
