# 阶段 3：播放端视觉外壳

状态：进行中，3.1 已完成（2026-10-25）。前置：阶段 2 的页面清单（已确定）；播放器部分需要阶段 2 的播放器参考图。参照路线图：[样式还原与验收方法](../roadmap.md#6-样式还原与验收方法)。

## 当前事实

`frontend/` 已有主框架（[AppLayout.vue](../../frontend/src/layout/AppLayout.vue)）、侧栏菜单、顶部导航、明暗主题和窄屏抽屉；界面名称已改为 Sonora，标志图形沿用 SPlayer。菜单、路由与顶栏用户入口已按首发范围重建（3.1）。各页面仍是占位（[EmptyView.vue](../../frontend/src/views/EmptyView.vue)）；没有播放器、卡片、歌曲列表或视觉比较工具。

数据只经 [api/source.ts](../../frontend/src/api/source.ts) 的 `DataSource` 读取。开发服务器上以 `?fixture=<场景>` 打开时换成 [fixtures](../../frontend/src/fixtures/index.ts)，不需要 Go 服务，例如 `http://127.0.0.1:14559/?fixture=library#/`；场景有 `library`、`empty`、`loading`、`unreachable`。生产构建不含 fixture（已用 `vite build` 检查产物）。3.3 在这一层上补齐 ViewModel 与其余场景。

## 工作包

### 3.1 菜单与路由

- [x] 按[阶段 2 的信息架构](phase-2-player-baseline.md#信息架构)重建路由与侧栏：首页、歌曲、专辑、艺术家、我喜欢的音乐、最近播放、创建的歌单分组（新建按钮，去掉在线/本地切换）。移除在线页面、心动模式按钮、收藏的歌单和网易登录入口。运行确认（2026-10-25）：
  - 路由名沿用 SPlayer（`home`、`songs`、`albums`、`album`、`artists`、`artist-songs`/`artist-albums`、`like-songs`、`history`、`playlist`、`search-*`、`404`、`500`），详情页以 `?id=` 指定，缺少 `id` 或 `keyword` 时转到 404，未知路径转到 404。
  - 侧栏由 SPlayer [Menu.vue](../../reference/SPlayer/src/components/Layout/Menu.vue) 改编，样式不变；图标沿用 SPlayer 流媒体页面标签的 `Music`、`Album`、`Artist`。歌单显示封面（SPlayer 默认 `menuShowCover: true`），没有封面时用 `album.jpg`；歌单分组默认折叠，打开自己的歌单时自动展开并高亮。专辑、艺术家详情页高亮所属列表，SPlayer 在这两处不高亮任何项。新建按钮暂无动作，弹窗在阶段 5 接入。
  - 检查方式：临时开发实例（数据为曲库副本、端口 14570/14572）与 fixture 下用 Playwright Chromium 截图，覆盖展开、折叠侧栏、歌单高亮、`990×800` 抽屉菜单、404 与重定向；截图在 `/tmp`，已删除。
- [x] 顶栏用户入口显示登录用户名，提供退出（fixture 提供用户）。由 SPlayer [User.vue](../../reference/SPlayer/src/components/Layout/User.vue) 改编：头像沿用未登录时的 `Person` 图标；弹层只保留用户名、邮箱（替换等级标签的位置）与“退出登录”，去掉喜欢数量与多账号。退出沿用 SPlayer 的确认对话框，调用 `POST /auth/logout` 后显示“已退出登录”页（[SignedOut.vue](../../frontend/src/views/Status/SignedOut.vue)），不直接跳转登录，否则 Authelia 会以现有会话立即重新登录（与 homepage 一致）。运行确认：对临时开发实例退出，请求为 `POST /auth/logout`，随后显示退出页。会话失效（401）时跳转 `/auth/login`，开发模式下无法触发，待 OIDC 环境验证。邮箱与退出页属于新增视觉元素，待 Eric 确认。

### 3.2 主题与视觉相关代码

- [ ] 对照 [Provider.vue](../../reference/SPlayer/src/components/Global/Provider.vue) 补齐 Naive UI 主题覆盖、字体与全局配置。
- [ ] 最小窗口宽度 990px（2026-10-25 确定）：窗口更窄时页面保持 990px 宽、出现横向滚动；不引入 SPlayer 在 `max-width: 989.98px` 下的移动版全屏播放器与移动布局。
- [ ] 复刻主题色与封面取色链路（[color.ts](../../reference/SPlayer/src/utils/color.ts)），以固定封面验证生成的 CSS 变量与原项目一致。
- [ ] 复刻外观与歌词设置项对视觉的作用（设置界面在阶段 5 接入），用 fixture 切换验证。

### 3.3 Fixture 与 ViewModel

- [ ] 定义歌曲行、专辑/艺术家/歌单卡片、详情头部、播放器状态的 ViewModel，字段沿用播放端 API 的数据形态（`artists` 数组、`cover` 对象），不含文件路径，不使用网易字段。
- [ ] fixture 可切换空曲库、有数据、加载、服务不可达、播放中/暂停；播放器 fixture 复现阶段 2 参考图中的歌曲数据。

### 3.4 页面模板（视觉语言一致）

- [ ] 首页：问候语、快捷入口、最近添加的专辑、最近播放。
- [ ] 歌曲列表（列、排序菜单、右键菜单、滚动）；专辑、艺术家网格；专辑详情（含碟号分组）、艺术家详情、歌单详情、我喜欢的音乐、最近播放。
- [ ] 搜索建议面板与结果页（歌曲、专辑、艺术家、歌单）。
- [ ] 新增的视觉元素（碟号分组标题等）交 Eric 确认。

### 3.5 播放器展示层（严格复刻）

- [ ] 底部播放栏、播放队列抽屉、全屏播放器（各样式与背景、AMLL 歌词渲染含逐字/翻译/音译）。只接 fixture。
- [ ] 去掉评论按钮、`Official` 标签、音质标签与音质切换；有逐字歌词时显示逐字标记。

### 3.6 依赖边界与视觉比较

- [ ] 增加检查，禁止前端导入网易 API、Cookie 工具或 `reference/` 下的运行时代码。
- [ ] 引入 Playwright 截图对照流程，复用阶段 2 的环境参数；输出参考图、实际图、差异图与结论；接入 Taskfile（例如 `task test:visual`）。

## 退出条件

- 没有 Go 服务时也能启动，并能用 fixture 切换上述状态。
- 播放器的截图对照通过，证据已留存；主题与取色结果与原项目一致。
- 其他页面经 Eric 看过，新增视觉元素已确认。
