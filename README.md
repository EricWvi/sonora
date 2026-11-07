# Sonora

Sonora 是以 SPlayer 为界面参照重新实现的音乐播放器。

## 开发

使用 Node.js 22.12+、pnpm 10.28.1、Go 1.26 和 [Task](https://taskfile.dev/)；若本机没有 `pnpm` 命令，先运行 `corepack enable pnpm`。项目命令统一收录在 [Taskfile.yml](Taskfile.yml)，运行 `task --list` 查看。

```bash
task install
task run:server   # Go 服务，开发模式，127.0.0.1:14560
task run:admin    # 管理后台，http://127.0.0.1:14561/admin/ ，账号 admin / admin
task run:web      # 播放端，http://localhost:14559
```

开发模式只在 `go run -tags dev` 构建中可用：跳过初始化 token，使用写死的管理员 `admin` / `admin`，数据写在 `backend/.data/`。`task format` 格式化全部代码；`task lint` 检查前端与管理后台的格式、ESLint、类型以及 Go 的 gofmt 与 `go vet`，播放端的 ESLint 与构建同时检查依赖边界（不导入网易 API、Cookie 工具与 `reference/`）；`task test` 在此之上运行 Go 测试并构建两份前端；`task build` 生成嵌入两份前端的发布二进制 `release/sonora`；`task clean` 清理构建产物。`task test:visual` 把播放器与 SPlayer 参考图逐像素对照，依赖本机 `.local/` 中的参考图与 baseline fixture，只能在本机运行，见[阶段 3 的 3.6](docs/phase/phase-3-player-shell.md#36-依赖边界与视觉比较)。

工程组成：

- `frontend/`：播放端，使用 SPlayer 的 Web 前端技术栈：Vue 3、TypeScript、Naive UI、Pinia、Vue Router、VueUse、SCSS、Vite。**明确舍弃 Electron，不继承原项目的桌面宿主、主进程、预加载、IPC 和打包行为。** 业务、原生音频模块和网易依赖未迁入。
- `admin/`：管理后台，Vue 3、TypeScript、Naive UI 默认风格，与 `frontend/` 组成 pnpm workspace。
- `backend/`：Go 服务，标准库 `net/http`、modernc SQLite 与 YAML 配置，与 `../homepage` 一致。

当前进度：阶段 1 已完成 Go 服务的管理员初始化、上传入库与管理 API，以及管理后台界面，见[阶段 1](docs/phase/phase-1-admin.md)。播放端只完成展示初始化：原样式主框架、首页空态、菜单导航、明暗主题和窄屏抽屉；尚未接入音乐服务、账号或播放功能。标志图形暂时复用 SPlayer，界面名称已改为 Sonora。完整页面复刻及逐状态视觉验收仍按路线图推进。复用资源及许可证见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。

## 项目方向

2026-10-22 确定：重新实现 Sonora，不在原 SPlayer 项目上继续改造。原项目与网易云音乐耦合过深，难以按 Sonora 的需求进行修改。Sonora 舍弃网易云音乐来源，使用我们自己开发的音乐服务；播放器与自研服务通过明确的契约协作。

2026-10-23 确定首发范围：

- 只支持桌面版 Chrome，不考虑移动端。2026-11-07 调整：Web 播放端仍是如此；手机上使用[阶段 8](docs/phase/phase-8-android-client.md) 的安卓客户端。
- Sonora 是自托管的音乐流媒体服务，“自研音乐服务”即指它，由三部分组成：
  - 播放端：Web 前端，播放器严格复刻 SPlayer，其他页面沿用其视觉语言。体验是流媒体服务，不接触文件和目录，没有文件夹浏览。
  - 管理后台：独立工程 `admin/`，Naive UI 默认风格；首次使用时凭容器日志中的一次性初始化 token 在页面上创建管理员；音乐只能在这里上传入库。
  - Go 服务：运行在容器中，音乐文件存放在挂载卷，为播放端和管理后台提供 API 与音频流。
- 播放端通过 OIDC 登录 Authelia，代码照搬 `../homepage`；喜爱、歌单与历史按用户隔离。开发环境写死一个管理员和一个普通用户。
- 先做管理后台。播放端页面清单与设计取舍已于 2026-10-24 确定，见[阶段 2](docs/phase/phase-2-player-baseline.md#已确定2026-10-24)。
- 播放端只做流媒体服务的常见播放功能。与网易云接口相关的页面和入口全部不做（[阶段 7](docs/phase/phase-7-netease-and-menus.md) 列出的部分除外），原项目的大部分在线页面因此不在范围内。
- 标志图形暂时复用；所有 SPlayer 名称与文字换成 Sonora。

## 界面要求

**复刻 SPlayer 的视觉是硬性要求，也是播放端交付的验收条件。** 2026-10-24 Eric 明确了范围：

- **播放器严格复刻**：全屏播放器（及其各种模式与背景）、底部播放栏、播放队列抽屉。`reference/SPlayer` 的源码与实际渲染是权威基准，布局、配色、字体、字号、间距、尺寸、图标、圆角、阴影、透明度、动效和组件状态逐项还原；交付前在相同环境、视口、主题、数据和状态下与 SPlayer 截图对照。
- **SPlayer 所有视觉相关代码都复刻**：全局主题、封面动态取色、全局与组件样式、动效，以及歌词样式、播放器外观等设置项。
- **其他页面视觉语言一致**：使用同一套颜色、字体、圆角、阴影和组件，页面结构按 Sonora 的流媒体体验自行设计。

接入自研音乐服务、重写架构或分阶段实现功能，都不能成为降低上述要求的理由。未经 Eric 明确同意，不得改动视觉方案；无法消除的环境差异应记录证据并交由 Eric 确认。管理后台在 SPlayer 中没有对应界面，使用 Naive UI 默认风格。具体方法见[路线图的样式验收要求](docs/roadmap.md#6-样式还原与验收方法)。

**2026-11-02 Eric 确认：对标 SPlayer 的开发已基本完善，此后可以进行不照搬 SPlayer 的改动**，例如性能与交互上的改进。改动处在代码注释与阶段文件中写明与 SPlayer 的差异；改动视觉方案仍须 Eric 同意，播放器相关改动后仍运行截图对照，预期的差异记为例外。首批这类改动见[阶段 6](docs/phase/phase-6-first-release.md#已确定2026-11-02)：去掉进度条的过渡，专辑与艺术家网格改为虚拟网格。

本机 `http://localhost:25884` 运行的是 `reference/SPlayer`，可用于浏览和对照原项目界面。该地址是当前本地参照环境，不是 Sonora 的固定服务端口。

## 文档

- [重写路线图](docs/roadmap.md)：架构方向、样式验收、分阶段计划和待确定事项。
- [阶段计划](docs/phase/README.md)：各阶段的当前状态、工作包与退出证据。
- [SPlayer 调研记录](docs/splayer-research.md)：原项目的源码证据、耦合点与重写风险。
- [SPlayer 浏览器会话](docs/splayer-session.md)：本机凭据位置和无头浏览器的统一加载方式。
- [Sonora API 约定](docs/api.md)：管理 API、错误格式与版本约定。
- [部署 Sonora](docs/deployment.md)：配置、数据卷与首次初始化。
