# 阶段 3：播放端视觉外壳

状态：3.1–3.6 已完成，新增视觉元素 Eric 已确认（2026-10-25）。前置：阶段 2 的页面清单（已确定）；播放器部分需要阶段 2 的播放器参考图。参照路线图：[样式还原与验收方法](../roadmap.md#6-样式还原与验收方法)。

## 当前事实

`frontend/` 已有主框架（[AppLayout.vue](../../frontend/src/layout/AppLayout.vue)）、侧栏菜单、顶部导航、明暗主题和窄屏抽屉；界面名称已改为 Sonora，标志图形沿用 SPlayer。菜单、路由与顶栏用户入口已按首发范围重建（3.1）。页面模板（3.4）已完成：首页、歌曲、专辑、艺术家、各详情页与搜索都由 `DataSource` 取数，fixture 与 Go 服务都可用；写操作（歌单编辑与排序、清空最近播放）只提示，在阶段 5 接入；喜欢与播放在阶段 4.4 接入（2026-10-25）。播放器展示层（底部播放栏、播放队列抽屉、全屏播放器，3.5）已完成，由 fixture 驱动、不出声；截图对照工具见 3.6。

数据只经 [api/source.ts](../../frontend/src/api/source.ts) 的 `DataSource` 读取，它覆盖播放端 API 的全部读取接口。开发服务器上以 `?fixture=<场景>` 打开时换成 [fixtures](../../frontend/src/fixtures/index.ts)，不需要 Go 服务；`&player=<状态>` 再指定播放器的初始状态（`source.ts` 导出的 `fixturePlayer`），例如 `http://127.0.0.1:14559/?fixture=baseline&player=yoru-paused#/`。场景与状态见 3.3。生产构建不含 fixture（已用 `vite build` 检查产物）。

## 工作包

### 3.1 菜单与路由

- [x] 按[阶段 2 的信息架构](phase-2-player-baseline.md#信息架构)重建路由与侧栏：首页、歌曲、专辑、艺术家、我喜欢的音乐、最近播放、创建的歌单分组（新建按钮，去掉在线/本地切换）。移除在线页面、心动模式按钮、收藏的歌单和网易登录入口。运行确认（2026-10-25）：
  - 路由名沿用 SPlayer（`home`、`songs`、`albums`、`album`、`artists`、`artist-songs`/`artist-albums`、`like-songs`、`history`、`playlist`、`search-*`、`404`、`500`），详情页以 `?id=` 指定，缺少 `id` 或 `keyword` 时转到 404，未知路径转到 404。
  - 侧栏由 SPlayer [Menu.vue](../../reference/SPlayer/src/components/Layout/Menu.vue) 改编，样式不变；图标沿用 SPlayer 流媒体页面标签的 `Music`、`Album`、`Artist`。歌单显示封面（SPlayer 默认 `menuShowCover: true`），没有封面时用 `album.jpg`；歌单分组默认折叠，打开自己的歌单时自动展开并高亮。专辑、艺术家详情页高亮所属列表，SPlayer 在这两处不高亮任何项（Eric 已确认，2026-10-25）。新建按钮暂无动作，弹窗在阶段 5 接入。
  - 检查方式：临时开发实例（数据为曲库副本、端口 14570/14572）与 fixture 下用 Playwright Chromium 截图，覆盖展开、折叠侧栏、歌单高亮、`990×800` 抽屉菜单、404 与重定向；截图在 `/tmp`，已删除。
- [x] 顶栏用户入口显示登录用户名，提供退出（fixture 提供用户）。由 SPlayer [User.vue](../../reference/SPlayer/src/components/Layout/User.vue) 改编：头像沿用未登录时的 `Person` 图标；弹层只保留用户名、邮箱（替换等级标签的位置）与“退出登录”，去掉喜欢数量与多账号。退出沿用 SPlayer 的确认对话框，调用 `POST /auth/logout` 后显示“已退出登录”页（[SignedOut.vue](../../frontend/src/views/Status/SignedOut.vue)），不直接跳转登录，否则 Authelia 会以现有会话立即重新登录（与 homepage 一致）。运行确认：对临时开发实例退出，请求为 `POST /auth/logout`，随后显示退出页。会话失效（401）时跳转 `/auth/login`，开发模式下无法触发，待 OIDC 环境验证。邮箱与退出页属于新增视觉元素，Eric 已确认（2026-10-25）。

### 3.2 主题与视觉相关代码

- [x] 对照 [Provider.vue](../../reference/SPlayer/src/components/Global/Provider.vue) 补齐 Naive UI 主题覆盖、字体与全局配置。[Provider.vue](../../frontend/src/components/Global/Provider.vue) 照搬原文件，只去掉自定义背景图（阶段 2 保留的主题配置只有主题色与封面取色）与自定义代码注入；`window.$message` 等全局工具一并挂载。设置与状态迁入 [setting.ts](../../frontend/src/stores/setting.ts)、[status.ts](../../frontend/src/stores/status.ts)，字段名与默认值照搬 SPlayer，只保留阶段 2 的设置表中的项；顶栏设置菜单的明暗切换照搬（跟随系统 → 浅色 → 深色），主题配置与全局设置的弹窗在阶段 5 接入。原项目的 `watch` 不监听 `themeVariant`，Sonora 保持一致。
- [x] 最小窗口宽度 990px（2026-10-25 确定）：窗口更窄时页面保持 990px 宽、出现横向滚动；不引入 SPlayer 在 `max-width: 989.98px` 下的移动版全屏播放器与移动布局。实现：`#app` 横向滚动，`#app-layout` 最小宽度 990px。运行确认：`900×800` 时 `#app` 的 `scrollWidth` 为 990、`clientWidth` 为 900，`990×800` 无滚动。底部播放栏已确认（3.5）：窄于 990px 时保持 990px 宽并随页面滚动。全屏播放器与队列抽屉窄于 990px 时铺满视口、不随页面滚动：它们盖住页面，页面的横向滚动条在下面，跟随滚动会让右侧 90px 无法到达（Eric 确认，2026-10-25）。
- [x] 复刻主题色与封面取色链路（[color.ts](../../reference/SPlayer/src/utils/color.ts)），以固定封面验证生成的 CSS 变量与原项目一致。运行确认（2026-10-25）：对照脚本 [scripts/baseline/theme.mjs](../../scripts/baseline/theme.mjs) 在参考实例与 Sonora 中各跑同一组 118 个用例，比较 body 上的 CSS 变量、`songCoverTheme`，以及 Naive UI `ConfigProvider` 实际收到的 `theme` 与 `themeOverrides`。用例覆盖 20 个预设、纯色与自定义颜色（明暗 × 是否全局着色），6 个主题变体，4 张封面（两首基线歌曲的封面，以及生成的灰色图和渐变图）的封面取色，关闭“播放器跟随封面主色”和自定义全局字体。结果：118 个用例全部一致，重复一轮仍一致；结果在 `.local/baseline/splayer/2026-10-25/theme/`。
  - 取色本身有随机性：material-color-utilities 的 `QuantizerWsmeans` 在初始簇不足时用 `Math.random()` 补点，同一张照片封面在 SPlayer 中多次取色，某个通道会差 1（三轮测得两种结果；纯色图与渐变图不受影响）。脚本在取色前把两边的 `Math.random` 换成同一个种子的序列，再做精确比较。实际使用中，原项目与 Sonora 有同样的 ±1 波动。
  - 参考实例取色时直接调用播放控制器的 `afterPlaySetup`，封面请求由 Playwright 拦截并返回固定图片；Sonora 从 Vite 开发服务器导入 `utils/color.ts` 调用 `getCoverColor`。封面尺寸已在 3.5 确定：服务端增加 `100` 与 `1024` 两档缩略图，播放器按 SPlayer 的 `coverSize` 取 s=100（底部栏、背景与取色）、m=300、l=1024，xl 用原图（Eric 确认，2026-10-25）。
- [x] 复刻播放器相关的外观与歌词设置项对视觉的作用（设置界面在阶段 5 接入），用 fixture 切换验证。运行确认（2026-10-25）：
  - 设置的字段名与默认值与 SPlayer 逐项相同（脚本比对两个 store 的默认值）。
  - 状态表加入 `set-*` 状态，每个状态把多项设置改成非默认值，与参考实例同条件截图对照（参考为 `2026-10-25-sonora`，`1440×1000`）。逐像素一致的有：`set-bar`（时间格式、播放列表数量、隐藏迷你封面、底栏歌词）、`set-lrc-layout`（封面/歌词占比、不跟随封面主色、歌词位置居中、固定字号与三种字号、字重、调换翻译与音译、歌词模糊、混合模式、水平偏移、滚动位置、歌词字体与英语歌词字体）、`set-lrc-right`（歌词靠右、靠右位置、隐藏括号内容、全屏播放器元素全部关闭）、`set-fullscreen-color`（全屏封面渐变位置、纯色背景）、`set-yrc`（逐字歌词靠右、固定字号、日语歌词字体）。
  - AMLL 的 `set-yrc-amll`（隐藏已播放行、渐变宽度、逐字音译、靠右、滚动位置、字体字重字号）与 `set-lrc-amll`（调换翻译与音译、关闭音译、倒计时）只剩评论层造成的 1–2 个像素（见 3.5 的例外），封面取色随机时前景差 1–2。
  - `set-yrc-off`（关闭逐字歌词）的歌词内容不同：参考实例改用网易另给的 LRC，Sonora 只有一份 TTML，由逐字歌词生成普通歌词（与 SPlayer 只有 TTML 时相同，见 3.3）。这是数据差异，渲染一致。
  - 流动背景的帧率、流速、渲染比例、暂停时停止流动（`set-animation`）按渲染器输入比较，两边相同（方法见 3.6）。
  - 未在这里验证的：动态封面（视频来自网易接口，Sonora 没有来源，设置保留但不起作用）；音乐频谱与背景响应低频音量（需要音频，4.4）；展开动画（过程动画，未采集）。
- [x] 页面相关的设置项（歌曲列表显示、侧边栏管理、首页栏目、页面切换动画）随 3.4 页面模板验证。运行确认（2026-10-25，临时 Vite 实例与 Playwright Chromium，`fixture=library`，每项在新的 BrowserContext 中写入 `sonora-setting` 后检查页面）：
  - 歌曲列表显示：`showSongAlbum`、`showSongDuration`、`showSongOperations`、`showSongArtist` 关闭时，表头与每行的对应列消失；`hiddenCovers` 的 `list`（歌曲行与详情头部封面）、`album`（专辑网格换成无封面卡片）、`artist`（艺术家网格）、`artistDetail`（艺术家详情图片）、`home`（首页卡片、专辑与歌曲行）都生效。`hiddenCovers` 补回 SPlayer 的 `artist` 与 `playlist`（搜索结果的歌单）两项，字段名与默认值相同。`hideBracketedContent` 在 SongCard、ListDetail、CoverList、ArtistList 中调用照搬的 `removeBrackets`，fixture 没有带括号的标题，只做了源码确认。
  - 侧边栏管理：`sidebarHide.hideHistory`、`hideUserPlaylists` 隐藏“最近播放”与“创建的歌单”，与 SPlayer 的 `Menu.vue` 相同（此前侧栏没有读取这两项）。
  - 首页栏目：`homePageSections` 字段形态与 SPlayer 相同，栏目换成 Sonora 的“最近添加的专辑”（`album`）与“最近播放”（`history`）；调换 `order` 后顺序随之改变，`visible: false` 的栏目不显示；`showHomeGreeting` 关闭时没有问候语。
  - 页面切换动画：AppLayout 按 SPlayer 加入 `RouterView` 的 `Transition`（`router-${routeAnimation}`）、`KeepAlive`（`max: 20`，排除 `layout`）与回到顶部按钮。8 种 `routeAnimation` 切换页面时都出现对应的 `router-*-enter-active` 与 `-leave-active`，`none` 不出现动画类；`useKeepAlive` 开启时离开再回到歌曲页保留列表滚动位置，关闭时回到顶部。动画样式（`animate.scss`）与 SPlayer 相同。

### 3.3 Fixture 与 ViewModel

- [x] 定义歌曲行、专辑/艺术家/歌单卡片、详情头部、播放器状态的 ViewModel，字段沿用播放端 API 的数据形态（`artists` 数组、`cover` 对象），不含文件路径，不使用网易字段。源码确认（2026-10-25）：
  - [api/types.ts](../../frontend/src/api/types.ts) 补齐播放端 API 的数据形态（歌曲、专辑、艺术家、历史、歌词、列表参数），与 [player.go](../../backend/internal/server/player.go) 的响应结构逐字段对应；`DataSource` 增加曲库、喜爱、历史、歌单与歌词的读取，以及 `audioUrl`。写操作（喜爱、歌单编辑、记录播放）在阶段 4、5 接入时再加。
  - [viewmodels/index.ts](../../frontend/src/viewmodels/index.ts)：`SongRow` 是歌曲加喜爱状态（列按阶段 2 的“歌曲列表”）；`CoverCard` 与 `DetailHeader` 直接携带 API 对象，按类型区分，详情头部的四种对应阶段 2 的“详情页头部”表；`PlayerState` 是队列、当前位置、播放位置与是否在播放。页面模板与播放器组件在 3.4、3.5 按需要再细化，不另起字段名。
- [x] fixture 可切换空曲库、有数据、加载、服务不可达、播放中/暂停；播放器 fixture 复现阶段 2 参考图中的歌曲数据。运行确认（2026-10-25，临时 Vite 实例与 Playwright Chromium，逐个场景调用 `DataSource`）：

  | 场景 `fixture=` | 内容 | `player=` |
  | --- | --- | --- |
  | `library` | 虚构曲库：5 位艺术家、5 张专辑（含双碟专辑、多位艺术家合作、超长标题）、34 首歌（含无专辑的歌、两首有 LRC 歌词）、4 个歌单、喜爱与最近播放。没有封面 | `paused`、`playing` |
  | `baseline` | 阶段 2 基线的歌曲：《THE BOOK》9 首与《紅蓮華》，含歌词原文与封面；喜爱为空（参考账号没有喜欢这两首歌） | `yoru-start`（0 秒）、`yoru-paused`、`yoru-playing`（54.0 秒，队列为整张专辑，当前第 8 首）、`guren-paused`、`guren-playing`（25.6 秒） |
  | `empty` | 空曲库，没有歌单 | — |
  | `loading` | 请求一直不返回 | — |
  | `unreachable` | 请求都以“无法连接服务器”失败 | — |

  列表的搜索、筛选、排序与分页在 fixture 中按服务端规则近似实现（拼音与数值比较、默认新入库在前、按专辑时按碟号与曲目序号）。未知场景、未知播放器状态或缺少 baseline 数据时，页面报错并说明原因。
  - **baseline 的数据不进入仓库**：[scripts/fixtures/baseline.mjs](../../scripts/fixtures/baseline.mjs) 从 `.local/baseline/splayer/2026-10-25/data/` 生成 `.local/fixtures/baseline/`（曲库 JSON 与封面），fixture 以 `import.meta.glob` 读取。封面按服务端规则命名（SHA-256 前 16 字节）；没有 ThumbHash 编码器，占位为空字符串（与服务端无法解码的图片相同）。
  - **歌词按 Sonora 保存的格式转换**：《夜に駆ける》的网易 LRC、翻译与音译合成一份 LRC，翻译与音译写在同一时间戳下。《紅蓮華》的 YRC 转为 TTML，因为只有 TTML 能保留每个字的起止时间与翻译、音译；逐字 LRC 只记录字的开始时间，SPlayer 解析逐字 LRC 时也不对齐翻译。转换用 SPlayer 自己的解析代码（`parseLrc`、`alignLyrics`）与 AMLL 的 `parseYrc`、`stringifyTTML`；SPlayer 的 `lyricLinesToTTML` 不写 `itunes:key`，AMLL 1.0.1 会跳过这样的行，因此不用它。
  - **转换的校验**：脚本在 Chromium 中按 Sonora 的解析路径读回（TTML：`cleanTTMLTranslations` 后 `parseTTML`；LRC：`parseSmartLrc` 后 `alignLyricLines`），与 SPlayer 从网易文件得到的歌词逐行比较，再与参考实例播放器 store 中的最终歌词（`data/<歌曲>-player.json`，见阶段 2 的 0.1）比较。比较每行与每个字的起止时间、文本、翻译、音译、背景人声与对唱标记。两首歌两项比较全部一致。

  以 baseline 对照截图时已知的数据差异，在 3.5、3.6 处理：
  - SPlayer 全屏播放器在歌名下显示网易的歌曲别名（`alia`：“向夜晚奔去”、“TV动画《鬼灭之刃》片头曲”）。Eric 确定 Sonora 不要歌曲别名（2026-10-25），这一行不显示，对照时作为已知差异。
  - 歌名下的标签行（参考图中的 `HQ`、`YRC`、`Official`）：Eric 确定歌词类型标签也不显示（2026-10-25），音质与音源标签原本就去掉，整行不再出现，对照时作为已知差异。
  - SPlayer 只有 TTML 时，用逐字歌词生成普通歌词（`lrcData`），参考实例中《紅蓮華》的普通歌词来自网易的另一份 LRC，行数相同。现有参考图中逐字模式下不显示普通歌词。

### 3.4 页面模板（视觉语言一致）

- [x] 公共组件照搬 SPlayer，数据换成播放端 API 的对象。运行确认（2026-10-25）：
  - [SongCard.vue](../../frontend/src/components/Card/SongCard.vue)（歌曲行，去掉音质、原唱翻唱、特权、云盘、MV、脏标标签与电台、移动端分支）、[SongList.vue](../../frontend/src/components/List/SongList.vue)（虚拟列表、排序菜单、右键菜单、拖动排序、定位当前歌曲与回到顶部）、[ListDetail.vue](../../frontend/src/components/List/ListDetail.vue)（详情页头部，滚动时收起）、[CoverList.vue](../../frontend/src/components/List/CoverList.vue) 与 [ArtistList.vue](../../frontend/src/components/List/ArtistList.vue)（封面网格，悬停播放按钮）、[SongListCard.vue](../../frontend/src/components/Card/SongListCard.vue)（首页快捷入口）、[SongListMenu.vue](../../frontend/src/components/Menu/SongListMenu.vue) 与 [useSongMenu.ts](../../frontend/src/composables/useSongMenu.ts)、列表组合函数 `useListDetail`、`useListSearch`（Fuse.js 模糊搜索，同 SPlayer 版本 7.4.1）、`useListScroll`、`useListActions`。
  - **排序由服务端完成**（阶段 2 的“歌曲列表”）：SongList 的排序菜单照搬，只修改 status store 的 `listSortField`、`listSortOrder`（与 SPlayer 同名，同样保存在浏览器中），页面监听后带上 `sort`、`order` 重新请求。排序字段只保留服务端支持的六项，名称与 SPlayer 相同。SPlayer 的单选组同时用 `v-model` 与 `@update:value`，选中新字段时顺序会从“默认”变为“升序”，Sonora 保持一致。
  - **加载方式**：全部歌曲、艺术家的歌曲与专辑、专辑与艺术家网格、搜索结果分页载入（[usePagedList.ts](../../frontend/src/composables/List/usePagedList.ts)，触底或“加载更多”时追加，过期请求丢弃）；专辑、歌单、我喜欢的音乐、最近播放一次载入全部（[fetchAll](../../frontend/src/api/pages.ts)，每页 500）。“播放”按阶段 2 的“播放行为”：专辑、歌单、我喜欢、最近播放放入整个列表，全部歌曲、艺术家的歌曲与搜索结果只放入已经载入的部分。
  - **播放**：控制器按 SPlayer 补上 `updatePlayList` 与 `addNextSong`（data store 的 `setNextPlaySong`），只改界面状态；随机模式下打乱队列仍在 4.4。`playPlaylistId` 改为带类型的字符串（`album-3`、`playlist-5`、`likes`），因为 Sonora 的专辑与歌单 ID 会重复，网格上的播放按钮据此显示暂停。
  - 写操作（喜欢、添加到歌单、编辑与删除歌单、保存歌单顺序、从歌单中删除、清空最近播放、复制更多信息、选择歌手）只提示，分别在阶段 4.4 与 5 接入（[utils/auth.ts](../../frontend/src/utils/auth.ts)、[utils/modal.ts](../../frontend/src/utils/modal.ts)）。
  - 加载失败时显示错误说明与“重试”（[ListError.vue](../../frontend/src/components/List/ListError.vue)，沿用 SPlayer 流媒体页面的空状态：`n-empty` 加主色按钮），详情页出错时不显示头部骨架。
- [x] 首页：问候语、快捷入口、最近添加的专辑、最近播放。运行确认（2026-10-25）：问候语照搬 SPlayer（带用户名）；快捷入口是两张 SongListCard（“我喜欢的音乐”沿用 SPlayer 的说明文字，“歌曲”显示曲库歌曲数）；栏目标题与列表沿用 SPlayer 首页（`CoverList`，以及 SPlayer 每日推荐用的 `SongList height="auto"` 显示最近 10 首），标题指向专辑页与最近播放页。首页结构是新的，Eric 已确认（2026-10-25）。
- [x] 歌曲列表（列、排序菜单、右键菜单、滚动）；专辑、艺术家网格；专辑详情（含碟号分组）、艺术家详情、歌单详情、我喜欢的音乐、最近播放。运行确认（2026-10-25）：
  - 歌曲、专辑、艺术家三个列表页的标题与操作栏沿用 SPlayer 本地音乐页面（标题、数量、播放、刷新、模糊搜索；[ListPageHeader.vue](../../frontend/src/components/List/ListPageHeader.vue)），搜索走服务端的 `q`。专辑页的排序（按标题、最近添加）用 SPlayer 窄窗口下标签页的下拉按钮样式。
  - 专辑详情、歌单详情、我喜欢的音乐照搬 SPlayer 的 `album.vue`、`playlist.vue`、`liked.vue`，头部内容按阶段 2 的“详情页头部”表：专辑的“更多”只有添加到歌单；歌单有“编辑歌单”与“更多 → 删除歌单”，默认排序且未搜索时可拖动排序；我喜欢的音乐用默认封面。艺术家详情照搬 `Artist/layout.vue`（圆形图片、歌曲数与专辑数、“歌曲 / 专辑”标签页），去掉关注、视频、简介与“更多”。最近播放照搬 `History.vue`。
  - 右键菜单条目按阶段 2：立即播放、下一首播放、添加到歌单、更多操作（复制歌曲名称、复制更多信息）、从歌单中删除（在自己的歌单中）、查看专辑、查看艺术家、同名搜索。“查看专辑”“查看艺术家”是 SPlayer 没有的条目，沿用 `Album`、`Artist` 图标；多位艺术家时打开选择（弹窗在阶段 5）。
  - 交互在 fixture 中逐项点过：双击歌曲放入整张专辑并从该曲播放、网格播放按钮放入整张专辑、艺术家页“播放”、右键“下一首播放”插到当前曲目之后、专辑内模糊搜索、按时长排序后重新请求、详情页滚动收起头部；页面没有报错。空曲库、加载中与服务不可达三个场景都检查过。
- [x] 搜索建议面板与结果页（歌曲、专辑、艺术家、歌单）。运行确认（2026-10-25）：顶栏搜索框照搬 SPlayer 的 `SearchInp.vue`（遮罩、搜索历史、搜索建议、`searchInputBehavior`），搜索历史保存在浏览器中（`sonora-data`，与 SPlayer 的 data-store 一样）。建议来自曲库搜索：歌曲 4 首、艺术家、专辑与自己的歌单各 2 项，点击歌曲加入下一首并播放，点击其他项进入详情。结果页照搬 `Search/layout.vue` 与各类型页面，标签页为歌曲、专辑、艺术家、歌单；歌单搜索用 `GET playlists?q=`（`DataSource.playlists` 增加 `q`，fixture 同步）。去掉热搜榜、推荐关键词、分享链接识别与搜索框右键菜单（复制、粘贴）。
- [x] 新增的视觉元素交 Eric 确认，以下各项 Eric 已确认（2026-10-25，截图在 `.local/visual/pages-3.4/`）：
  - 专辑详情的碟号分组标题：多碟专辑、默认排序时显示，沿用表头的高度与透明度，`Album` 图标对齐序号列，文字“第 N 碟”对齐标题列，序号按碟重新开始；排序或搜索时不分组。
  - 专辑详情头部的年份（`Calendar-Empty`）与总时长（`Time`），歌单头部同时显示更新时间（`Update`）与创建时间（`Time`）；SPlayer 只显示其中一个时间。
  - 专辑卡片在名称下显示艺术家（沿用 SPlayer 视频卡片的艺术家样式，单行省略），再显示“N首 · 年份”（SPlayer 是发行日期）。
  - 首页的结构（上文）、列表页的排序下拉按钮与加载失败的“重试”。
  - 窄于 990px 的页面没有单独处理，按最小宽度 990px 横向滚动；`1200px` 以下详情页头部右侧的模糊搜索隐藏，与 SPlayer 相同。
- 回归：页面改动后 `task test:visual` 全量 54 个状态 30 一致、24 例外、0 不符（AppLayout 加入路由动画与 KeepAlive 不影响播放器截图）。`task lint:frontend` 与 `vite build` 通过，产物中没有 fixture。

### 3.5 播放器展示层（严格复刻）

- [x] 底部播放栏。运行确认（2026-10-25）：
  - 照搬 [MainPlayer.vue](../../frontend/src/components/Player/MainPlayer.vue)、[PlayerRightMenu.vue](../../frontend/src/components/Player/PlayerRightMenu.vue)、[PlayerSlider.vue](../../frontend/src/components/Player/PlayerComponents/PlayerSlider.vue)、TextContainer、useTimeFormat 与 v-debounce 等指令；status、music、data 三个 store 的播放器字段照搬 SPlayer，歌曲是 API 的 `Track`。去掉倍速、定时关闭、音质、桌面歌词、电台与私人 FM；“其他控制”只留均衡器，“更多”菜单按 Eric 的决定删减（见[阶段 2](phase-2-player-baseline.md#播放器里与-sonora-不对应的元素)）。均衡器、添加到歌单、复制更多信息、选择歌手的弹窗在阶段 5 接入，喜欢在阶段 4.4 接入，目前只提示。
  - 歌词按 SPlayer 处理本地歌词的方式解析（[LyricManager.ts](../../frontend/src/core/player/LyricManager.ts)：TTML 为逐字歌词，LRC 自动识别格式，再做排除、括号替换与规范化），歌词行按 SPlayer 的 `calculateLyricIndex` 计算。[PlayerController.ts](../../frontend/src/core/player/PlayerController.ts) 沿用原方法名（播放暂停、上下一首、跳转、音量、静音、随机与循环），这一阶段只改界面状态、不播放音频；随机只切换模式，打乱队列在 4.4 接入。启动时载入 fixture 的播放器状态（`?player=`）与喜爱列表。
  - `index.html` 去掉 `lang="zh-CN"`，与 SPlayer 一致：有 `lang` 时 Chrome 为中文选简体中文字形，SPlayer 没有 `lang`，中日文都用日文字形（基线环境记录的 Noto Sans CJK JP），歌名“夜”字因此不同。这一改动影响播放端所有中文的字形。
  - 窄于 990px：栏保持 SPlayer 的 `position: fixed`，加 `min-width: 990px`，并按 `#app` 的横向滚动量平移；`900×800` 时宽 990、滚动 90px 后左移 90px，不滚动时不设 transform。改成相对页面定位或单独建合成层都会改变文字的渲染（与参考图出现数百个像素的差异），所以没有采用。
  - 截图对照（`1440×1000` 的 7 个底部栏状态，参考为 `2026-10-25-sonora`，见 3.6）：`bar-paused`、`bar-paused-dark`、`bar-playing`、`bar-progress-hover`、`bar-progress-drag` 逐像素一致。两处残余差异：
    - `bar-paused-start`：“00:00 /”在参考图中低 1 像素（333 像素）。这是参考图采集步骤造成的：参考实例换一种步骤（播放、切歌、暂停后跳到 0 秒）截出的图与 Sonora 完全一致，与 Sonora 的实现无关。Eric 确认作为例外接受（2026-10-25）。
    - `bar-volume-hover`：进度条填充末端 2 个像素，最大差值 9，属于阶段 2 噪声表中“进度的亚像素位置”一类，Eric 已接受。
  - 窄窗口（`1280`、`1025`、`1024`、`990` × 800 的暂停与音量弹层）：`1280` 暂停态一致；其余每张 1–9 个像素、最大差值 2–9，位于进度条末端与播放按钮的圆形边缘。两边各元素的位置逐一测量完全相同，推测与背后页面不同导致的合成分层有关，未证实；Eric 已接受（2026-10-25）。
  - 按钮交互（播放暂停、上下一首及循环到队尾、随机、循环、静音、点击进度条跳转、“更多”菜单）在 fixture 中逐项点过，状态变化正确。
- [x] 播放队列抽屉。运行确认（2026-10-25）：
  - 照搬 [SongPlayList.vue](../../frontend/src/components/List/SongPlayList.vue)、[VirtualScroll.vue](../../frontend/src/components/UI/VirtualScroll.vue) 与 [useDragSort.ts](../../frontend/src/composables/List/useDragSort.ts)，挂在 AppLayout 中与 SPlayer 相同的位置。歌名取 `Track.title`；Sonora 的歌曲都带艺术家数组，去掉电台与艺术家为字符串的分支。拖动排序用 SPlayer 自己的 `useDragSort`，不需要 sortablejs（SPlayer 只在设置的首页栏目管理中用它）。
  - 控制器补上 `cleanPlayList`、`removeSongIndex`、`moveSong`，data 与 status store 补上 `setPlayList`、`originalPlayList` 与 `resetPlayStatus`，流程照搬 SPlayer，只改界面状态；队列的保存在 4.4 接入。
  - **修正 SPlayer 的一个功能问题**：`v-debounce` 指令在挂载时绑定点击函数，之后不更新。队列项按歌曲 id 复用元素，删除或移动歌曲后点击某一项，播放的是它原来序号上的歌。参考实例上复现：删除第 3 首后点“群青”（序号 4），播放的是“ハルカ”。Sonora 的指令改为点击时调用最新绑定的函数，防抖行为不变；与视觉无关。
  - 截图对照（参考为 `2026-10-25-sonora`）：`1440×1000` 的 `queue-open`、`queue-hover`、`queue-empty`、`queue-drag`，以及 `1280`、`1025`、`1024`、`990` × 800 的 `queue-open`，全部逐像素一致；`1440×1000` 重复三轮仍一致。`queue-drag` 是新加的状态（拖住第 2 首的手柄移到第 5、6 首之间，显示落点线、拖动标签与半透明的原项）。同一轮回归中底部栏的结果与 3.5 记录的例外相同。
  - 交互在 fixture 中逐项点过：拖动排序（含拖动当前曲目，当前序号随之更新）、删除当前曲目之前、之后与当前曲目本身、点击切歌并关闭抽屉、“当前播放”滚动到当前曲目、清空列表（确认对话框后底部栏隐藏、提示“播放列表已清空”），页面没有报错。
  - 窄于 990px 时抽屉固定在视口右侧，不随页面横向滚动：它是带遮罩的浮层，SPlayer 中同样贴着视口；`900×800` 滚动前后都在 `x=500–900`。
- [x] 全屏播放器（各样式与背景、AMLL 歌词渲染含逐字/翻译/音译），只接 fixture。运行确认（2026-10-25）：
  - 照搬 FullPlayer、PlayerMenu、PlayerControl、PlayerMeta（PlayerBackground、PlayerCover、PlayerData）、PlayerLyric（入口、DefaultLyric、AMLyric）、PlayerSpectrum、components/AMLL（LyricPlayer、BackgroundRender）、PlayerProvider、`s-image`、`utils/style.ts` 与歌词字体配置，挂在 AppLayout 中与 SPlayer 相同的位置。引入 `@applemusic-like-lyrics/core` 0.5.1 与 pixi 对等依赖，版本与 SPlayer 锁定的相同。
  - 去掉：评论（含只在评论时出现的顶部实时歌词）、移动版播放器、下载、私人 FM、电台、自动混音、云盘与解锁提示、本地封面读取、动态封面（视频来自网易接口）、Electron 的阻止息屏与歌词滚动消息。歌名取 `title`，专辑取 `album.title`。歌名下的别名与整行标签不显示（Eric 的决定）。喜欢在 4.4、复制歌词与设置弹窗在阶段 5 接入，目前只提示；频谱与低频音量在 4.4 接入音频后才有数据。
  - fixture 的“原图”改用网易 1920 档缩略图：SPlayer 的 `coverSize.xl`（全屏封面样式）请求的是它，而不是网易原图；两边显示的原本是不同的 JPEG。数据快照补采 `data/<歌曲>-cover-1920.jpg`，fixture 有它时用作原图。
  - 截图对照（参考为 `2026-10-25-sonora`，`1440×1000`）：`full-cover-blur`、`full-cover-color`、`full-cover-none`、`full-cover-blur-dark`、`full-cover-blur-playing`、`full-cover-blur-hidden`、`full-lrc-plain`、`full-yrc` 逐像素一致。残余差异：
    - **评论层（Eric 确认作为例外，2026-10-25）**：SPlayer 一直挂着透明度为 0 的评论组件，Chrome 因它改变合成分层，Sonora 没有这一层。影响 `full-record-blur`（唱片纹理与封面，1.9% 的像素、最大差值 70）、`full-lrc-amll` 与 `full-yrc-amll`（封面阴影上 1 个像素）。证据：在参考实例上同样步骤截图，截图前删除 `.player-comment` 后，三张都与 Sonora 逐像素一致；保留时差异重现。
    - **封面取色 ±1**：`full-fullscreen-blur` 前景（文字与图标）差 1–2。参考图的封面主色为 `255,218,217`，Sonora 为 `255,218,216`；material-color-utilities 的随机补点造成（见 3.2）。参考实例另一轮取到 `216` 时，截图与 Sonora 逐像素一致。
    - 流动背景（`full-cover-animation-t0`、`-t2000`）按渲染器输入比较，两边相同（见 3.6）。
  - 窄窗口（`1280`、`1025`、`1024`、`990` × 800 的 `full-cover-blur`、`full-yrc`）：每张 100–150 个像素、最大差值 21–31，位于底部控制栏的图标，同样是评论层造成（删除参考实例的评论后 `1280×800` 一致）。
  - 交互在 fixture 中逐项点过：从底部栏打开、播放暂停、上下一首、点击歌词跳转并播放、歌词偏移 ±0.5 秒、纯净歌词模式、从全屏播放器打开队列（抽屉带 `full-player` 样式）、点击专辑跳转并收起、收起按钮、元素自动隐藏与移动鼠标后恢复，页面没有报错。
- [x] 去掉评论按钮、歌曲别名、歌名下的整行标签（音质、歌词类型 `LRC`/`YRC`/`TTML` 及歌词来源切换、`Official` 等音源标签）与音质切换。源码确认（2026-10-25），截图对照见上。

### 3.6 依赖边界与视觉比较

- [x] 增加检查，禁止前端导入网易 API、Cookie 工具或 `reference/` 下的运行时代码。运行确认（2026-10-25）：规则集中在 [frontend/boundaries.mjs](../../frontend/boundaries.mjs)，分两层：
  - ESLint（`task lint:frontend`）：`import`、`export … from`、动态 `import()` 与 `import.meta.glob` 的路径中不得出现 `reference/`、`netease`、`cookie`（后两项不区分大小写）；源码字符串中不得出现 `music.163.com`、`api/netease`、`MUSIC_U`。
  - 构建插件（`vite build`，因此 `task build:frontend` 与 `task test` 都会执行）：加载每个模块时检查实际路径（`reference/` 目录下的文件，以及路径含 `netease`、`cookie` 的源码与依赖包），构建结束时检查 SCSS 等引入的文件，生成产物时检查网易接口地址。ESLint 看不到的样式引入与间接依赖由这一层拦截。
  - 检查方式：临时加入 10 种违规写法（`js-cookie`、`@/utils/cookie`、导入与动态导入 `reference/` 下的文件、`import.meta.glob`、网易 API 包、`.vue` 中的网易请求模块、网易地址与 `MUSIC_U` 字符串），ESLint 全部报错，正常的动态导入不受影响；构建时临时引入 `reference/` 下的 TS 文件、SCSS `@use` 与网易地址，三种都以“依赖边界检查失败”中止。现有代码两层都通过，产物中没有 fixture。
  - 管理后台（`admin/`）不复制 SPlayer 代码，没有接入这项检查。
- [x] 引入 Playwright 截图对照流程，复用阶段 2 的环境参数；输出参考图、实际图、差异图与结论；接入 Taskfile（`task test:visual`）。运行确认（2026-10-25）：
  - 状态表（歌曲、跳转点、交互步骤、裁剪区域）从采集脚本抽到 [player-states.mjs](../../scripts/baseline/player-states.mjs)，参考实例与 Sonora 共用。
  - **对照基线 `--sonora`**：[splayer-player.mjs](../../scripts/baseline/splayer-player.mjs) `--sonora` 用 SPlayer 自己的设置去掉 Eric 决定不显示的元素（`showPlayerQuality`、`showPlayMeta` 关闭，`fullscreenPlayerElements` 的评论与下载关闭），并清空歌曲别名；播放中的状态也停在跳转点，只把 store 的 `playStatus` 设为 true，避免截图时间不定；截图区域包含背后页面的状态隐藏 `#main` 与歌曲列表的浮动按钮。产物在 `.local/baseline/splayer/2026-10-25-sonora/`，原基线 `2026-10-25` 不变，仍是 SPlayer 原样的记录。
  - [scripts/visual/player.mjs](../../scripts/visual/player.mjs) 在同样的环境与步骤下截取 Sonora（`?fixture=baseline&player=…`），产物在 `.local/visual/<run>/`，用 [diff.mjs](../../scripts/baseline/diff.mjs) 比较。fixture 的封面使用网易按 100、300、1024 缩放的原图（基线快照补采），播放器状态带上参考实例音频的实际时长。
  - 只重采部分状态（`--only`）时，两个脚本的 manifest 保留其余状态的记录，按状态表排序（2026-10-25；此前会覆盖整个 manifest，`2026-10-25-sonora/1440x1000` 的 manifest 因此缺少四个底部栏状态的记录，截图仍在）。
  - 没有歌曲时没有底部栏，背后的页面会从队列抽屉的圆角透出，`queue-empty` 因此也隐藏 `#main`（`hideMain`）。
  - 无背景样式（`full-cover-none`）同样会透出背后的页面（参考实例专辑页右下角的浮动按钮），也隐藏 `#main`。
  - 状态的设置在基础设置上合并，对象型设置（如 `fullscreenPlayerElements`）逐个字段合并，`--sonora` 关掉的评论与下载按钮不会被状态覆盖。
  - **流动背景的对照方法**（2026-10-25）：流动背景按真实时间渲染，不能逐像素比较。两边运行同一版本的 AMLL 渲染器，因此比较它的输入：`BackgroundRender` 实际收到的 props（帧率、流速、渲染比例、是否有歌词、低频音量、播放状态）与封面图片内容的 SHA-256。两个采集脚本在 `animations: "allow"` 的状态把它们写入 manifest 的 `background`；截图只用于人工对照。`full-cover-animation-t0`、`-t2000` 与 `set-animation` 两边相同。
  - **`task test:visual`**（[compare.mjs](../../scripts/visual/compare.mjs)）：在 `127.0.0.1:14571` 起一个临时 Vite（`--port` 可改，`--base` 用已有的服务），按参考目录中每个视口已有的状态截取 Sonora、逐像素比较，再按已接受的差异判定，最后停掉 Vite。结论逐个状态列为“一致”“例外”（附原因）或“不符”，有不符、缺图、页面报错或流动背景的渲染输入不同时以 1 退出。截图、差异图与 `summary.json` 在 `.local/visual/test-visual/`，每次运行先清空。可用 `-- --viewport 1440x1000 --only '^bar-'` 只跑一部分。
  - 只能在本机运行：参考图在 `.local/baseline/splayer/2026-10-25-sonora/`（由参考实例采集），baseline fixture 在 `.local/fixtures/baseline/`，Playwright 来自 `../palace`；缺少任何一项时说明如何补齐后退出。任务本身不访问参考实例，也不读取凭据；参考图需要重采时仍手动运行 `splayer-player.mjs --sonora`。
  - 判定规则写在 `compare.mjs` 中，只收录 Eric 已接受的差异，每条限定状态、视口、像素数与最大差值：`bar-paused-start` 的采集步骤、`bar-volume-hover` 的亚像素位置（`1440×1000` 至多 2 个像素，窄窗口的底部栏按 3.5 的记录至多 9 个像素、最大差值 9）、评论层（`full-record-blur`、窄窗口 `full-cover-blur` 与 `full-yrc`）、`set-yrc-off` 的数据差异。全屏播放器与设置状态先去掉差值不超过 2 的像素（取色 ±1 与评论层的 1–2 个像素），剩下的再按上述规则判定；取色偏移与评论层图标同时出现时（例如 `1024×800` 的 `full-yrc`，6820 个像素）也能分开处理。`diff.mjs` 的报告为此增加了按差值统计的像素数（`deltas`）。
  - 结果：全量（`1440×1000` 34 个状态与 4 个窄视口共 20 个状态）54 个状态中 30 个一致、24 个属于例外、0 个不符，约 3 分半钟。第一轮全量中 `1440×1000` 的 `full-yrc-amll` 出现过一次 AMLL 歌词整行错位（20279 个像素，最大差值 99），单独重跑三次都只差评论层的 1 个像素。之后又跑了三轮全量，其中一轮 `set-yrc-amll` 的当前行逐字高亮与参考图不同（去掉取色偏移后 9945 个像素，最大差值 45），另一轮窄窗口 `bar-volume-hover` 出现 5 个像素（在 3.5 已接受的范围内，规则随之放宽到 9）。AMLL 的两个状态（`full-yrc-amll`、`set-yrc-amll`）偶尔截到不同的歌词帧，单独重跑都能对上，原因未查明，没有为此放宽判定；遇到时重跑该状态确认（**待查**）。

## 退出条件

- 没有 Go 服务时也能启动，并能用 fixture 切换上述状态。
- 播放器的截图对照通过，证据已留存；主题与取色结果与原项目一致。
- 其他页面经 Eric 看过，新增视觉元素已确认。
