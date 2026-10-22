# Sonora

Sonora 是以 SPlayer 为界面参照重新实现的音乐播放器。

## 开发

使用 Node.js 22.12+、pnpm 10.28.1 和 [Task](https://taskfile.dev/)；若本机没有 `pnpm` 命令，先运行 `corepack enable pnpm`。项目命令统一收录在 [Taskfile.yml](Taskfile.yml)，运行 `task --list` 查看。

```bash
task install
task run:web
```

浏览器访问地址为 `http://localhost:14559`。`task format` 格式化前端，`task lint` 检查格式、ESLint 和类型，`task build` 生成 Web 构建，`task clean` 清理 `frontend/dist`。`task test` 执行当前全部检查和生产构建；目前尚无单元测试，后端占位目录也不执行 Go 检查。

`frontend/` 使用 SPlayer 的 Web 前端技术栈：Vue 3、TypeScript、Naive UI、Pinia、Vue Router、VueUse、SCSS、Vite；包管理使用 pnpm。**明确舍弃 Electron，不继承原项目的桌面宿主、主进程、预加载、IPC 和打包行为。** 业务、原生音频模块和网易依赖未迁入。`backend/` 为 Go 自研音乐服务的占位目录，尚未初始化 Go 模块或接口。

当前只完成前端展示初始化：原样式主框架、首页空态、菜单导航、明暗主题和窄屏抽屉。搜索仅展示空结果，账号入口显示未登录；没有接入音乐服务、账号或播放功能。保留原标志和 SPlayer 字样以便视觉核对，页面标题为 Sonora。完整页面复刻及逐状态视觉验收仍按路线图推进，不能将此次初始化视为全站复刻完成。复用资源及许可证见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。

验证范围为格式、lint、类型、Web 构建及真实 Chromium 展示检查。

## 项目方向

2026-10-22 确定：重新实现 Sonora，不在原 SPlayer 项目上继续改造。原项目与网易云音乐耦合过深，难以按 Sonora 的需求进行修改。Sonora 舍弃网易云音乐来源，使用我们自己开发的音乐服务；播放器与自研服务通过明确的契约协作。

## 界面要求

**严格复刻 SPlayer 样式是硬性要求，也是每个页面交付的验收条件。** `reference/SPlayer` 的源码与实际渲染共同作为权威基准，布局、配色、字体、字号、间距、尺寸、图标、圆角、阴影、透明度、动效和组件状态必须逐项还原。

所有实现页面及其加载、空态、错误、悬停、选中、展开、弹层和响应式状态均受此约束。接入自研音乐服务、重写架构或分阶段实现功能，都不能成为调整原样式的理由。未经 Eric 明确同意，不得重新设计、简化或替换原有视觉方案。

每个页面交付前必须在相同环境、视口、主题、数据和状态下，与 SPlayer 进行截图对照及交互检查。视觉差异必须修正；无法消除的环境差异应记录证据并交由 Eric 确认，不能自行降低验收标准。具体方法见[路线图的样式验收要求](docs/roadmap.md#6-样式还原与验收方法)。

本机 `http://localhost:25884` 运行的是 `reference/SPlayer`，可用于浏览和对照原项目界面。该地址是当前本地参照环境，不是 Sonora 的固定服务端口。

## 文档

- [重写路线图](docs/roadmap.md)：架构方向、样式验收、分阶段计划和待确定事项。
- [SPlayer 调研记录](docs/splayer-research.md)：原项目的源码证据、耦合点与重写风险。
- [SPlayer 浏览器会话](docs/splayer-session.md)：本机凭据位置和无头浏览器的统一加载方式。
