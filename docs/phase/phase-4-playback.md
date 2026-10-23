# 阶段 4：播放端认证、API 与播放闭环

状态：未开始。前置：阶段 1 的曲库与上传；阶段 3 的播放器展示层。参照路线图：[播放状态与异步规则](../roadmap.md#52-播放状态与异步规则)。

## 4.1 OIDC 认证

已确定（2026-10-23）：播放端使用 OIDC 登录 Authelia，代码完全照搬 `../homepage`（参考提交 `f188285`），不做 forward auth。

- [ ] 照搬 `../homepage/internal/auth`（`auth.go`、`provider.go` 及测试，含 `authelia` 构建标签下的 testcontainers 契约测试与 `testdata/authelia`），以及 `internal/store` 中的用户与会话部分；只替换包路径、Cookie 名称和配置键名，不改认证逻辑。
- [ ] 行为与 homepage 一致：`GET /auth/login`、`GET /auth/callback`、`POST /auth/logout`；授权码流程加 PKCE（S256），scope 为 `openid profile email`；首次使用时才发现 issuer，Authelia 不可用时服务仍能启动；会话长期有效，直到手动退出；`Require` 中间件校验同源与会话。
- [ ] 配置：`public_url`、`oidc.issuer`、`oidc.client_id`、`oidc.client_secret`（建议用环境变量 `SONORA_OIDC_CLIENT_SECRET` 提供）；回调地址 `<public_url>/auth/callback`。
- [ ] 开发模式使用写死的普通用户，跳过 OIDC（对应 homepage 的 `dev_user`）。
- [ ] 契约测试照搬 homepage 的 Taskfile 任务，不包含在 `task test` 中，依赖本地已有镜像。
- [ ] 播放端：未登录时跳转 `/auth/login`；用户入口显示用户名，退出调用 `/auth/logout`。

## 4.2 播放端 API

- [ ] 歌曲、专辑、艺术家、歌单的列表与详情，分页与排序；搜索；封面；歌词。
- [ ] 当前用户的喜爱、歌单、播放历史读写；按用户隔离。
- [ ] 音频流支持 HTTP Range；按阶段 2 的 Chrome 实验结论决定是否需要转码。

## 4.3 Playback

- [ ] 状态机 `idle → resolving → loading → playing / paused → ended`，失败进入 `error`；Playback 是播放状态的唯一所有者。
- [ ] 切歌请求标识与取消：先点 A 后点 B，A 较晚返回也不能覆盖 B；解析中暂停不得自动播放。
- [ ] 队列条目独立身份，允许重复入队；播放模式（顺序、单曲、随机）。
- [ ] 歌词、封面、媒体信息和进度跟随当前条目。
- [ ] 队列与播放位置持久化，重启后恢复为可继续播放的状态，不自动播放。
- [ ] 单元测试覆盖切歌竞态、解析中暂停、队列编辑与恢复。

## 退出条件

- 管理后台上传 → OIDC 登录播放端 → 曲库列表 → 入队 → 播放/暂停/拖动/切歌 → 歌词显示 → 重启恢复，整条链路在 Chrome 中真实可听。
- 两个用户的喜爱、歌单与历史互不可见；Authelia 契约测试通过。
- 快速切歌、播放中暂停、队列编辑和重启不串状态、不丢数据；播放器相关页面完成视觉对照。
