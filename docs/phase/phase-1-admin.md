# 阶段 1：Go 服务基础与管理后台

状态：工作包已实现并通过运行确认（2026-10-24）；真实曲库的大规模上传与浏览器中拖放文件夹上传于 2026-11-03 在线上运行确认，见文末“未完成与待验证”。参照路线图：[运行环境与后端范围](../roadmap.md#42-运行环境与后端范围)。

管理后台是另行开发的 Web 界面，负责上传音乐和维护曲库，使播放端只呈现流媒体体验。SPlayer 没有对应界面，**管理后台不受 SPlayer 样式复刻约束**。

## 已确定（2026-10-23）

- 视觉：Naive UI 默认风格，不做主题定制。
- 工程：独立前端工程 `admin/`，与 `frontend/` 组成 pnpm workspace；技术栈与播放端一致（Vue 3、TypeScript、Naive UI、Vite），依赖版本由同一份锁文件管理。生产环境由 Go 服务托管两份静态产物，管理后台挂在 `/admin/`，同一容器、同一端口。
- 音乐只能通过管理后台上传入库；挂载目录是服务独占的存储卷，不支持手动放入文件。
- 管理员：首次使用服务时，在管理后台页面上输入容器日志中的初始化 token、用户名和密码创建管理员；之后不再显示初始化页面。管理后台不走 OIDC。
- 开发环境写死一个管理员和一个普通用户，跳过初始化页面、初始化 token 与 OIDC。
- 下列工作包作为首发范围的起点，开发中按实际情况灵活调整。

## 证据位置

- 提交：`f2aa41c`（Go 服务）、`20b3405`（管理后台与 Taskfile）。
- 自动测试：`task test:go`（发布构建全部包，另以 `-tags dev` 跑配置测试）；`go test -race ./...` 通过。`task lint` 覆盖 gofmt、两种构建的 `go vet` 与 admin 的 Prettier、ESLint、vue-tsc。
- 运行确认：2026-10-24 用 `task build` 产出的 `release/sonora` 在全新数据卷上以生产配置启动，curl 与无头 Chromium（Playwright）走完初始化、上传、编辑、退出登录、会话过期与服务停止；开发模式用 `task run:server` 与 `task run:admin` 验证。截图在 Git 忽略的 `.local/screens/admin/`。
- 测试样本：[testdata](../../backend/internal/media/testdata/) 由 ffmpeg（`linuxserver/ffmpeg` 镜像，仅拉取运行）从 3 秒 440 Hz 正弦波编码，含中文标签、内嵌封面与歌词；基准时长取自 ffprobe。

## 工作包

### 1.1 Go 服务骨架

- [x] 初始化 `backend/` Go 模块，接入 Taskfile 的 lint、test、run；本地开发直接运行 Go 程序，不在本机构建容器镜像。证据：`task lint:go`、`task test:go`、`task run:server`；源码确认、运行确认。
- [x] 沿用 `../homepage` 的基础选择：标准库 `net/http` 路由、modernc SQLite、YAML 配置。`cmd/`、`internal/config`、`internal/store`（`PRAGMA user_version` 迁移、单连接、WAL）、`internal/server`（静态托管、错误映射）的结构与写法与 homepage 对齐。证据：源码确认。
- [x] 配置：`media_dir`（音频存储卷）与 `data_dir`（SQLite、封面）分开且必须不同；`public_url`；`dev` 开关。证据：[config_test.go](../../backend/internal/config/config_test.go)。
- [x] 开发模式：写死的管理员与普通用户只在开发模式生效，生产配置无法启用。证据：`config_release_test.go`；运行确认 `release/sonora -config config.dev.yaml` 拒绝启动。

偏差：

- “环境变量覆盖密钥”：阶段 1 的配置里没有密钥，暂未实现环境变量覆盖；阶段 4 照搬 OIDC 时按 homepage 加入 `SONORA_OIDC_CLIENT_SECRET`。
- 开发模式比 homepage 的 `dev_user` 更严：`dev` 只被 `-tags dev` 构建接受，且要求监听回环地址，发布二进制无法启用。
- 写死的普通用户 `listener` 目前只是 [devmode](../../backend/internal/devmode/devmode.go) 中的常量，播放端尚无用户表；阶段 4 照搬 homepage 的用户与会话时接入。为此管理员使用独立的 `admins`、`admin_sessions` 表，把 `users`、`sessions` 留给 homepage 代码原样使用。
- Go 模块在 `backend/` 而非仓库根目录，`go:embed` 不能引用上级目录，所以构建脚本先把 `frontend/dist`、`admin/dist` 复制到 `backend/web/` 再嵌入（homepage 直接嵌入 `frontend/dist`）。

### 1.2 管理员初始化与登录

- [x] 服务启动时若不存在管理员，生成 32 字节加密随机 token，写入标准输出并附带初始化页面地址。token 只在内存中；重启后重新生成，旧 token 失效。证据：`TestSetupRejectsMissingWrongAndStaleTokens`；运行确认重启后旧 token 返回 403。
- [x] 未初始化时管理 API 只开放状态与初始化接口（其余返回 403 `not_initialized`，登录也被拒绝）；管理后台检测到未初始化时显示初始化页面，输入 token、用户名、密码与确认密码。证据：`TestUninitialisedServiceOnlyAllowsSetup`；无头浏览器确认任意深链接都进入初始化页。
- [x] 初始化接口先哈希再以常量时间比较 token；错误时拒绝并限流（每分钟 5 次失败）；成功后 token 立即作废。证据：`TestSetupRateLimitsWrongTokens`、`TestSetupSucceedsOnceAndRevokesToken`。
- [x] 初始化只能成功一次，并发请求只有一个生效；之后返回 409 `already_initialized`。证据：`TestConcurrentSetupOnlyOneWins`（HTTP 层）、`TestCreateFirstAdminOnlyOnce`（存储层条件插入）。
- [x] 密码以 bcrypt（cost 12）保存，不写入日志；token 是唯一写入日志的凭据。证据：源码确认；运行确认日志中不含密码。
- [x] 登录、退出、会话过期；Cookie `sonora_admin_session` 限定在 `/admin/api/v1/`，`HttpOnly`、`SameSite=Strict`，与播放端会话互不通用。会话闲置 7 天或登录满 30 天失效。证据：`TestLoginLogoutAndExpiry`、`TestSessionHasHardLifetime`；无头浏览器确认会话失效后回到登录页并提示“登录已过期”。
- [x] 部署说明写明通过 `docker logs` 获取初始化 token：[部署 Sonora](../deployment.md)。

偏差：

- 限流按用途全局计数（初始化每分钟 5 次失败、登录每分钟 10 次失败），不按客户端地址：经反向代理后所有请求同一地址，且 token 为 256 位，全局放慢足以防猜测。代价是他人持续尝试时管理员也要等待窗口结束。
- 登录时未知用户名也做一次 bcrypt 比较，避免用响应时间探测用户名。
- 写请求另做同源检查（`Origin` 必须等于 `public_url`），沿用 homepage 的 `sameOrigin`。

### 1.3 曲库模型与上传入库

- [x] 曲库模型：歌曲、专辑、艺术家；ID 由 `AUTOINCREMENT` 分配、删除后不复用；文件路径只在 store 内部使用，存储卷布局为 `<sha256 前两位>/<sha256>.<格式>`。证据：`TestUploadOutcomes` 断言响应不含路径、`file` 与哈希。
- [x] 上传：多文件上传，读取标题、艺术家、专辑、专辑艺术家、音轨号、碟号、时长、年份、内嵌封面、内嵌歌词与同名外置 `.lrc`；处理重复文件、缺标签与损坏文件。证据：`media_test.go`（五种格式、VBR/CBR/Xing、ID3 USLT、ID3 与 Vorbis 多值字段、截断文件）、`store_test.go`（入库、去重、并发重复上传、GB18030 歌词、跨格式同专辑共用封面、多值艺术家入库）。
- [x] 删除歌曲同时清理存储卷文件、无引用的封面以及空专辑和孤立艺术家。播放端的歌单、喜爱、历史引用在阶段 4 以外键级联移除（`TestDeletingTrackRemovesItFromListenerData`）。证据：`TestDeleteRemovesFilesAndOrphans`。
- [x] 用真实编码器产出的音频样本测试上传、去重与失败路径。证据：见“证据位置”的测试样本说明。

实际做法与偏差：

- 格式按文件内容识别，不看扩展名；只接受桌面 Chrome 能播放的 MP3、FLAC、AAC（M4A）、Ogg Vorbis、Opus。ALAC、Ogg FLAC、WAV 等返回 415。
- 时长由服务自行解析容器（MP3 Xing/VBRI/LAME 或 CBR 估算、FLAC STREAMINFO、MP4 `mdhd`、Ogg 结束页 granule），标签用 `github.com/dhowden/tag`。镜像中不需要 ffmpeg。
- “损坏”指容器结构无法读取或不完整（元数据块越界、缺 `moov`、找不到音频帧或 Ogg 结束页），不解码音频帧，因此音频数据中段损坏不会被发现。
- “重复”指字节完全相同（SHA-256）；同一录音的不同编码或不同标签不视为重复。
- 歌曲与专辑可以有多位艺术家，按标签或编辑时的顺序保存（`track_artists`、`album_artists`）。只拆分标签中真正的多值（ID3 的 NUL 分隔、Vorbis 的重复字段）；单个值即使含“/”“&”“feat.”也原样作为一个艺术家名，因为“AC/DC”这类名字本身含分隔符，由管理后台手动拆分。多值的标题与专辑名以 `, ` 连接。证据：`TestMultipleArtists`、`TestIngestMultiValueArtists`。
- 专辑以“专辑名（不区分大小写）+ 专辑艺术家集合（不计顺序）”识别；专辑艺术家缺失时用歌曲艺术家。缺标题时用文件名；“缺少标签”定义为缺艺术家或缺专辑。
- 结构迁移 2 把单艺术家的旧库转换为关联表，保留全部 ID 与自增序号，删除过的 ID 仍不复用。证据：`TestMigrateKeepsArtistsAndRetiredIDs`，以及开发库副本的迁移结果。
- 内嵌封面：专辑没有封面时成为专辑封面；与专辑封面不同时才保存为单曲封面，以便替换专辑封面后所有歌曲一起变化。
- 封面入库时计算 ThumbHash 存入 `covers` 表，升级前已有的封面在服务启动时补算（本机 303 张约 3 秒）。缩略图 96 与 300 两档，按需生成后缓存在 `covers/thumbs/`，随原图一起清理；本机曲库 96 档平均 3 KB、300 档平均 18 KB，原图平均约 150 KB。证据：`TestEncodeMatchesReference`（与 thumbhash npm 0.1.1 输出逐字节一致）、`TestCoverPlaceholdersAndThumbnails`、`TestPlaceholdersBackfilledOnOpen`。
- 外置歌词在同一次选择中按文件名配对，随音频一起提交；外置歌词优先于内嵌歌词。
- 单文件上限 1 GB；上传先流式写入存储卷内的 `.incoming/` 并计算哈希，确认后改名入库，服务重启时清理残留。

### 1.4 管理 API

- [x] 初始化、登录、退出、会话状态。
- [x] 上传与入库结果，歌曲/专辑/艺术家的列表、详情与修改、合并、删除，封面与歌词的上传替换删除，曲库概况；另加试听用的音频流（支持 Range）。证据：`TestLibraryEndpoints`。
- [x] 统一错误格式与 API 版本约定：[Sonora API 约定](../api.md)。

偏差：

- 没有单独删除艺术家的接口：艺术家随歌曲自动清理。删除专辑会删除其全部歌曲与文件。
- 合并艺术家时，同一首歌或专辑上同时出现的来源与目标只保留目标；合并后艺术家组合与另一张同名专辑相同的专辑一并合并。证据：`TestMergeArtistsCollapsesLinksAndAlbums`。
- 契约验证目前是 Go 的 HTTP 测试，没有独立的契约测试工具；管理后台的类型定义手工维护在 [api.ts](../../admin/src/api.ts)。

### 1.5 管理后台界面

- [x] 初始化 `admin/` 与 pnpm workspace，接入 Taskfile 的 `lint:admin`、`build:admin`、`run:admin`；开发服务器（`127.0.0.1:14561/admin/`）代理 `/admin/api` 到 Go 服务。证据：运行确认。
- [x] 初始化页面、登录页、退出。开发模式登录页预填并提示 admin / admin。
- [x] 上传：多文件与拖放，每个文件显示进度、入库结果与失败原因（损坏、不支持、过大、不可达），重复文件提示“曲库中已有”并链接到已有歌曲，失败可重试。证据：无头浏览器一次上传 6 个文件，结果“已入库 3 · 重复 1 · 失败 2”。
- [x] 元数据：编辑歌曲、专辑，重命名艺术家；合并重复的艺术家与专辑；筛选缺标签的歌曲。证据：无头浏览器把缺标签歌曲补全后筛选结果变为 0。
- [x] 多艺术家：列表中每位艺术家单独链接；歌曲与专辑的艺术家用标签输入编辑，名称含分隔符时提供“按分隔符拆分”。证据：无头浏览器上传多值与“周杰伦/费玉清”单值两首歌，前者入库为两位艺术家；后者拆分保存后并入同一专辑，孤立的“周杰伦/费玉清”被清理。
- [x] 封面与歌词：上传、替换、删除；歌词可直接编辑或上传 `.lrc`。
- [x] 封面先显示 ThumbHash 模糊占位，缩略图加载后淡入；列表用 96 档，编辑区用 300 档。证据：无头浏览器拦截封面请求时专辑列表显示模糊占位，放行后全部淡入；歌曲列表一页 50 首的封面从数 MB 降到 133 KB。
- [x] 曲库概况：歌曲、专辑、艺术家数量，总时长，存储占用，缺少标签数量（链接到筛选）。

偏差：

- 视图不按需加载：服务停止时按需加载的代码块也无法获取，会让导航静默失败；整体打包后切换页面由接口请求显示“无法连接 Sonora 服务”。
- 亮色、暗色跟随系统，均为 Naive UI 默认主题；仅链接颜色取主题主色。

## 退出条件

- [x] 全新数据卷启动后，用容器日志中的 token 能在页面上完成管理员初始化；缺少或错误的 token、重启前的旧 token 和重复初始化均被拒绝。运行确认。
- [x] 管理员能完成上传入库与元数据维护，重启服务后数据保持。运行确认（重启后会话、曲库与文件保持）。
- [x] 上传失败、文件损坏、会话过期和服务不可达有明确状态。运行确认。
- [x] 开发模式下无需初始化与 token，即可用写死的管理员登录。运行确认。

## 未完成与待验证

- **运行确认**（2026-11-03，Eric，桌面 Chrome，线上 1.0.0 的全新数据卷）：用 Eric 的真实曲库做多格式、大曲库上传（路线图风险表“上传入库与标签差异”要求大曲库样本）。在管理后台上传页拖入整个文件夹，共 1,632 首（MP3 1,588、Ogg Vorbis 44），12,425,215,570 字节。页面显示没有失败，也没有重复。入库后用只读 SQLite 查询核对：
  - 每首都有专辑、标题，时长都大于 0；建立专辑 672 张、艺术家 528 位。
  - 封面：671 张专辑有封面，48 首另有单曲封面。唯一没有封面的专辑是《故郷の原风景》，只有 1 首，按源文件没有内嵌封面处理（没有逐一检查源文件）。
  - 歌词：源文件夹中没有 `.lrc`，619 首有歌词，都来自内嵌标签，都是 LRC（MP3 612/1,588，Ogg 7/44）。随机抽取的 10 首无歌词歌曲多为纯音乐与配乐，按源文件本身没有歌词处理。
- **运行确认**（同上）：浏览器中真实拖放文件夹上传（无头测试使用文件选择）。Naive UI 先递归读完整个目录再逐个回调，上传页合成一批后再配对歌词。拖放时不在 `accept` 内的文件被直接忽略。歌词只按文件名（不含目录）配对，多个文件夹有同名文件时可能配错，记入[阶段 6 的已知缺陷与支持限制](phase-6-first-release.md#已知缺陷与支持限制100)；这次的源文件夹没有 `.lrc`，不受影响。
- 容器镜像与 Compose 示例属于阶段 6。
- 同录音不同编码的去重、音频帧级损坏检测未做，需要时再评估。
