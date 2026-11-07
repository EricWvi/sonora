# 阶段 8：安卓客户端

状态：计划中，尚未开始（2026-11-07 Eric 提出）。前置：阶段 7 退出（2026-11-06）。

## 目标

以 PixelPlayerOSS 为基础二次开发 Sonora 的安卓客户端：连接 Sonora 服务，支持现在 Web 播放端的大部分功能，并增加歌曲缓存（离线播放）。Web 播放端仍只支持桌面版 Chrome，不做移动端适配；手机上使用这个客户端。

## 已确定（Eric，2026-11-07）

- **基础项目**：[PixelPlayerHQ/PixelPlayerOSS](https://github.com/PixelPlayerHQ/PixelPlayerOSS)，不是主仓库 PixelPlayerHQ/PixelPlayer。
- **参照源码**：作为 Git submodule 放在 `reference/` 下（与 `reference/SPlayer` 相同），固定在 tag `v0.4.0-alpha.53`。
- **开发路线**：从参照源码中复制需要的大部分代码出来二次开发，不在原版上删除多余的功能。与 Web 播放端的做法一致：`reference/` 只用于对照，客户端代码不依赖它。
- **范围**：Web 播放端的大部分功能，加上歌曲缓存。具体哪些功能进入、哪些不做，见“待决事项”。

## 参照项目的事实（2026-11-07 从 GitHub 核对，未在本机构建）

- **许可证**：PixelPlayerOSS 是 GPL-3.0；主仓库 PixelPlayer 是专有许可，只允许个人非商业的学习与使用，禁止分发衍生作品，所以不能作为基础。复制 PixelPlayerOSS 代码的客户端须按 GPL-3.0 发布，并保留原版权声明；Sonora 现在是 AGPL-3.0，客户端的许可证与 [THIRD_PARTY_NOTICES.md](../../THIRD_PARTY_NOTICES.md) 的写法要在开始前定下来。
- **技术栈**：Kotlin、Jetpack Compose、Material 3、Media3（含 FFmpeg、无缝播放、交叉淡化）；Android 11+（API 30），编译与目标 SDK 37，JDK 21。包名 `com.lostf1sh.pixelplayeross`，需要改为 Sonora 自己的包名。
- **与 Sonora 最接近的部分**：已经接入 Navidrome/Subsonic 与 Jellyfin，有登录、同步、流式播放和应用私有目录的离线下载。相关代码在 `data/navidrome`、`data/jellyfin`、`data/network/*`、`data/worker/*SyncWorker.kt`，离线下载在 `data/offline/CloudOfflineRepository.kt`、`data/worker/CloudTrackDownloadWorker.kt`、`OfflineTrackDao`，以及下载管理页 `CloudDownloadsScreen`。这些是实现 Sonora 曲库同步与歌曲缓存的主要参照。
- **已经去掉的集成**：Telegram、网易云、QQ 音乐、Google Drive、Gemini、Cast、Wear OS、Play 商店计费、Firebase 与 Google Play 服务。
- **开发环境**（Eric，2026-11-07）：安卓客户端在 Mac 上开发，那里已有环境，做法参考 `../asr` 项目（听写的安卓与 Mac 客户端）。这台 Linux 开发机没有 JDK 与 Android SDK，不在这里构建。

## 待决事项

开始实施前逐项与 Eric 确认：

1. **客户端代码放在哪里**：Sonora 仓库内的新目录（如 `android/`），还是单独的仓库；影响许可证、CI 与发布方式。
2. **登录方式**：播放端现在走 Authelia OIDC，会话在浏览器 Cookie 中。客户端需要一种适合原生应用的方式（例如 OIDC 授权码 + PKCE 加 Authelia 新客户端，或 Sonora 签发的令牌），可能要改 Go 服务与 Authelia 配置。可参照 asr 已实现的原生客户端登录（`../asr/docs/todos.md`）：客户端打开系统浏览器走同一套 OIDC，按 RFC 8252 带 `code_challenge`（S256），安卓用自定义 scheme 回调，服务端签发设备 token。
3. **功能清单**：Web 播放端哪些功能进入客户端。候选包括曲库浏览、搜索、歌单、喜欢、最近播放、歌词、漫游 FM、网易云的混合列表与在线歌曲、导入到曲库、设置同步；网易云相关功能仍由 Go 服务转发，客户端不直接访问网易云。
4. **界面**：沿用 PixelPlayerOSS 的 Material 3 界面，还是向 Web 播放端（SPlayer）的视觉靠拢；不受 Web 播放端“严格复刻 SPlayer”的约束，视觉方案仍须 Eric 同意。
5. **歌曲缓存的规则**：手动下载、播放即缓存还是两者都有；容量上限与清理；曲库中歌曲被替换或删除时如何失效；在线（网易云）歌曲是否缓存。服务端可能需要配合，例如音频响应的缓存头（[阶段 9](phase-9-advanced.md) 的“音频 immutable 缓存”）。
6. **本地文件播放**：PixelPlayerOSS 的本地扫描、文件夹浏览、元数据编辑等是否保留；Sonora 的原则是音乐只通过管理后台入库。
7. **验收方式**：真机或模拟器，测试设备与安卓版本；客户端的发布与安装方式（GitHub Release、自行安装 APK 等）。

## 工作包

待上面的事项确定后建立。下一个会话的第一步：把 PixelPlayerOSS 作为 submodule 加到 `reference/PixelPlayerOSS`，检出 `v0.4.0-alpha.53`。
