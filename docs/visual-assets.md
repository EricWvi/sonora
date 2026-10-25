# 视觉资源与许可清单

更新日期：2026-10-25。本文是[阶段 2 工作包 0.2](phase/phase-2-player-baseline.md#02-视觉资源与许可清单) 的交付物：清点 SPlayer 中与视觉有关的代码、依赖、图标、字体与图片，记录来源、许可证和 Sonora 的复用方式。对外的许可声明以 [THIRD_PARTY_NOTICES.md](../THIRD_PARTY_NOTICES.md) 为准，两者需同步更新。

参考源码：`reference/SPlayer`，提交 `28b76bad9eb800e0c90dac8475d4769acfe2450b`。依赖版本取自其 `pnpm-lock.yaml`，许可证取自 npm 注册表的包声明（2026-10-25 查询）。

## 1. 许可证结论

- **SPlayer 本身是 AGPL-3.0**。Sonora 播放端复刻其视觉代码，播放端因此按 AGPL-3.0 分发，[LICENSE](../LICENSE) 已是 AGPL-3.0 全文。AGPL 第 13 条要求向通过网络使用服务的用户提供对应源码。
- **歌词渲染库 AMLL 是 AGPL-3.0-only**，与上一条一致，不额外增加限制。
- 其余视觉依赖为 MIT、ISC 或 Apache-2.0，只需保留版权与许可声明。
- **需要特别处理的三项**：
  - 底部播放栏的随机、列表循环、单曲循环图标来自 IconaMoon，许可证是 **CC-BY-4.0**，复制时必须署名（见第 4 节）。
  - Logo 字体 `logo.woff2` 是 **Poppins Medium，SIL OFL 1.1**，随字体分发时须附 OFL 全文。`frontend/public/fonts/` 已复制该字体，全文已补在同一目录。
  - 45 个图标的来源尚未确认，在确认前按 SPlayer 的 AGPL-3.0 处理（见第 4 节）。

## 2. 视觉相关代码

“复用方式”按 [README 的界面要求](../README.md#界面要求)：播放器严格复刻；SPlayer 所有视觉相关代码都复刻；其他页面只沿用视觉语言。行数是 SPlayer 中对应文件的总行数，用来估计工作量。

| 部分 | SPlayer 位置 | 行数 | 复用方式 | Sonora 现状 |
| --- | --- | --- | --- | --- |
| 全局样式与动画 | `src/style/`（`main.scss`、`animate.scss`） | 795 | 照搬 | 已复制到 `frontend/src/style/`，另拆出布局、导航、菜单等文件 |
| Naive UI 主题覆盖、明暗与主题色 | `src/components/Global/Provider.vue` | 324 | 照搬 | 已照搬，去掉自定义背景图与代码注入；生成结果与原项目一致（阶段 3.2） |
| 播放器主题覆盖（封面色） | `src/components/Global/PlayerProvider.vue` | 78 | 照搬 | 未开始 |
| 封面取色与配色方案 | `src/utils/color.ts`、`src/assets/data/themeColor.json` | 325 | 照搬 | 已照搬，去掉任务栏歌词取色；生成结果与原项目一致（阶段 3.2） |
| 字体大小工具 | `src/utils/style.ts` | 12 | 照搬 | 未开始 |
| 播放器：底部播放栏、全屏播放器、封面、歌词、背景、控制栏、频谱 | `src/components/Player/` | 6097 | 严格复刻，去掉评论、音质与网易相关元素（[阶段 2 决定](phase/phase-2-player-baseline.md#播放器里与-sonora-不对应的元素)） | 底部播放栏已照搬（`MainPlayer.vue`、`PlayerRightMenu.vue`、`PlayerSlider.vue`），截图对照见阶段 3.5（2026-10-25）；其余未开始 |
| AMLL 封装：歌词播放器、流动背景 | `src/components/AMLL/` | 512 | 严格复刻 | 未开始 |
| 播放队列抽屉 | `src/components/List/SongPlayList.vue`、`src/composables/List/useDragSort.ts` | 786 | 严格复刻 | 已照搬，去掉电台分支；截图对照逐像素一致（阶段 3.5，2026-10-25） |
| 歌词解析 | `src/utils/lyric/` | 1914 | 照搬 LRC、TTML 相关部分；去掉 YRC、QRC 与在线歌词源 | 已照搬 `parseLrc.ts`、`lyricParser.ts`（去掉 QRC 与 `lyricLinesToTTML`）、`parseTTML.ts`、`lyricStripper.ts`、`lyricFormat.ts`，处理流程见 `core/player/LyricManager.ts`（2026-10-25） |
| 歌曲列表、详情页头部 | `src/components/List/SongList.vue`、`ListDetail.vue` | 1425 | 视觉语言一致，结构按阶段 2 决定 | 未开始 |
| 卡片 | `src/components/Card/` | 831 | 视觉语言一致 | 未开始 |
| 布局、导航、侧栏、顶栏 | `src/components/Layout/`、`src/layout/` | 1505 | 视觉语言一致 | 已改编主框架、导航与主题初始化 |
| 通用组件：图片、输入框、虚拟滚动 | `src/components/UI/` | 737 | 照搬 | 已照搬 `VirtualScroll.vue`（播放队列），其余未开始 |
| 外观与歌词设置项 | `src/components/Setting/config/appearance.ts`、`lyric.ts`、`components/LyricPreview.vue` | 1735 | 照搬保留的设置项（[阶段 2 的设置表](phase/phase-2-player-baseline.md#设置)） | 未开始 |

## 3. 运行时依赖

只列与视觉相关的依赖。版本是 SPlayer 锁定的版本；Sonora 引入时使用相同的主版本，升级要先确认视觉不变。

| 包 | 用途 | SPlayer 锁定版本 | 许可证 | Sonora |
| --- | --- | --- | --- | --- |
| `naive-ui` | 组件库 | 2.43.2（固定） | MIT | 已用，同版本 |
| `@material/material-color-utilities` | 主题色与封面取色（`themeFromSourceColor`、`QuantizerCelebi`、`Score`） | 0.4.0 | Apache-2.0 | 已用，同版本 |
| `@imsyy/color-utils` | `rgbToHex` | 1.0.2 | package.json 写 ISC，包内 LICENSE 是 MIT 全文 | 已用，同版本（2026-10-25） |
| `lodash-es` | `chunk`（封面取色）及 SPlayer 各处的 `debounce`、`cloneDeep` 等 | 4.18.1 | MIT | 已用，同版本（2026-10-25） |
| `@applemusic-like-lyrics/core` | AMLL 歌词播放器与流动背景 | 0.5.1 | AGPL-3.0-only | 待引入；最新为 0.6.0，先锁定 0.5.1 |
| `@applemusic-like-lyrics/lyric` | LRC、TTML 等歌词解析 | 1.0.1 | AGPL-3.0-only | 已用，同版本（2026-10-25），播放端用它解析 TTML，fixture 脚本用它转换歌词。其依赖 `@applemusic-like-lyrics/ttml` 1.0.1（AGPL-3.0-only）与 SPlayer 相同，`pako` 解析为 2.2.0（SPlayer 锁定 2.1.0，MIT AND Zlib，用于 QRC 解压，与视觉无关） |
| `@pixi/app`、`@pixi/core`、`@pixi/display`、`@pixi/sprite`、`@pixi/filter-blur`、`@pixi/filter-color-matrix` | AMLL 的对等依赖 | 7.4.3 | MIT | 随 AMLL 引入 |
| `@pixi/filter-bulge-pinch` | AMLL 的对等依赖 | 5.1.1 | MIT | 随 AMLL 引入 |
| `sortablejs` | 设置中首页栏目的拖动排序（经 `@vueuse/integrations` 的 `useSortable`） | 1.15.7 | MIT | 阶段 5 接入设置时引入。播放队列与歌曲列表的拖动排序用 SPlayer 自己的 `useDragSort`，不依赖它 |
| `@vueuse/core` | 组合式工具（自动隐藏计时、媒体查询等） | 14.3.0 | MIT | 已用 |
| `dayjs` | 播放时间显示（`msToTime`） | 1.11.21 | MIT | 已用，同版本（2026-10-25） |

## 4. 图标

SPlayer 的 147 个图标在 `src/assets/icons/`，由 `SvgIcon.vue` 内联渲染。来源核对方法：取每个 SVG 的全部 `path` 数据，与 Iconify 发布的 33 个常见图标集（`@iconify-json/*`，2026-10-25 下载）逐一比对。

**已确认（路径完全一致），102 个：**

| 图标集与许可证 | 数量 | 图标 |
| --- | --- | --- |
| Google Material Icons，Apache-2.0 | 46 | `Add`、`AutoFix`、`Batch`、`Cloud`、`Code`、`DarkTheme`、`Delete`、`Download`、`EditNote`、`Eye`、`Favorite`、`FavoriteBorder`、`Folder`、`FormatList`、`Info`、`Keyboard`、`LightTheme`、`Link`、`List`、`Lock`、`Lyrics`、`Menu`、`Message`、`Music`、`NavigateBefore`、`NavigateNext`、`Paste`、`Pause`、`PauseCircle`、`Person`、`Phone`、`Play`、`PlayCircle`、`PlayFilled`、`Radio`、`Refresh`、`Right`、`Share`、`SkipNext`、`SkipPrev`、`SnippetFolder`、`Star`、`Storage`、`ThumbUp`、`Time`、`Update` |
| Material Symbols，Apache-2.0 | 24 | `AddList`、`Chat`、`Check`、`Close`、`Collapse`、`Copy`、`DeleteSweep`、`DownloadDone`、`Edit`、`Eq`、`Expand`、`Home`、`More`、`MusicList`、`PlayList`、`PlayRate`、`PlaylistAddCheck`、`Search`、`Sort`、`VolumeDown`、`VolumeMute`、`VolumeUp`、`Zoom`、`ZoomIn` |
| Material Design Icons（Pictogrammers），Apache-2.0 | 22 | `Earth`、`ExitToApp`、`EyeLock`、`FolderMusic`、`FolderPlus`、`Github`、`ListLockOpen`、`SortAZ`、`SortClockDown`、`SortClockUp`、`SortDateDown`、`SortDateUp`、`SortZA`、`TextPlay`、`TextSizeAdd`、`TextSizeReduce`、`Video`、`WindowClose`、`WindowHide`、`WindowMaximize`、`WindowMinimize`、`WindowRestore` |
| Phosphor，MIT | 3 | `HeartBit`、`SortDown`、`SortUp` |
| **IconaMoon，CC-BY-4.0** | 3 | `Repeat`、`RepeatSong`、`Shuffle` |
| Fluent UI System Icons，MIT | 2 | `Calendar-Empty`、`Disconnect` |
| Material Symbols Light，Apache-2.0 | 1 | `DropDown` |
| Remix Icon，Apache-2.0 | 1 | `QQ` |

`Github`、`QQ` 是品牌标志，Sonora 不使用。

**近似，25 个，待人工确认。** 相似度 0.93 以上的 21 个多半只是坐标精度或图标集版本不同： `Artist`（mdi:account-music）、`AutoTheme`（ic:round-brightness-auto）、`Fire`（mdi:fire）、`FolderCog`（mdi:folder-cog）、`History`（ic:round-history）、`IP`（mdi:ip）、`Password`（ic:round-password）、`PlaylistAdd`（material-symbols:playlist-add-rounded）、`Podcast`（mdi:podcast）、`Power`（ic:round-power-settings-new）、`Record`（mdi:podcast）、`Replay5`（material-symbols:replay-5-rounded）、`Restart`（ic:round-restart-alt）、`RssFeed`（ic:round-rss-feed）、`SearchOff`（ic:round-search-off）、`Settings`（ic:round-settings）、`SettingsOther`（ic:baseline-miscellaneous-services）、`Tag`（mdi:tag-multiple）、`ThumbUpOff`（ic:baseline-thumb-up-off-alt）、`TimeAuto`（material-symbols:time-auto-rounded）、`VolumeOff`（material-symbols:volume-off-rounded）。相似度 0.79–0.87 的 4 个：`CloudLockOpen`、`PlayNext`、`ThumbDown`、`DesktopLyric`。

**未确认，20 个：** `Album`、`Controls`、`DesktopLyric2`、`Discover`、`Down`、`Extension`、`Forward5`、`Fullscreen`、`FullscreenExit`、`HeartBroken`、`Help`、`HiRes`、`Location`、`LockOpen`、`Logo`、`Palette`、`SPlayer`、`SettingsLine`、`Stream`、`Up`。其中 `Logo`、`SPlayer` 是 SPlayer 自己的标志。确认来源前按 SPlayer 的 AGPL-3.0 处理。

**播放器用到的图标**（底部栏、全屏播放器、队列抽屉）中，未确认或只是近似的有：`Controls`、`Down`、`Fullscreen`、`FullscreenExit`、`Forward5`、`Location`、`Replay5`、`ThumbDown`、`TimeAuto`、`Album`、`Artist`。

**署名要求：** IconaMoon 的 CC-BY-4.0 要求注明作者、许可证与修改情况。Sonora 复制 `Repeat`、`RepeatSong`、`Shuffle` 时，在 THIRD_PARTY_NOTICES.md 加入 IconaMoon（作者 Dariush Habibpour，<https://github.com/dariushhpg1/IconaMoon>）的署名。Apache-2.0 与 MIT 的图标集同样在该文件中列出版权与许可证。

`frontend/src/assets/icons/` 现有 41 个图标，与 SPlayer 的文件逐字节一致（2026-10-25 阶段 3.5 队列抽屉后，`cmp` 核对）：Google Material Icons 16 个（`Add`、`DarkTheme`、`Delete`、`Favorite`、`FavoriteBorder`、`FormatList`、`LightTheme`、`Menu`、`Music`、`NavigateBefore`、`NavigateNext`、`Pause`、`Person`、`Play`、`SkipNext`、`SkipPrev`）、Material Symbols 11 个（`AddList`、`Copy`、`DeleteSweep`、`Eq`、`Expand`、`Home`、`PlayList`、`Search`、`VolumeDown`、`VolumeMute`、`VolumeUp`）、Material Symbols Light 1 个（`DropDown`）、IconaMoon 3 个（`Repeat`、`RepeatSong`、`Shuffle`），近似项 7 个（`Artist`、`AutoTheme`、`History`、`PlaylistAdd`、`Power`、`Settings`、`VolumeOff`），未确认 3 个（`Album`、`Controls`、`Location`）。侧栏的歌曲、专辑、艺术家图标沿用 SPlayer 流媒体页面标签的 `Music`、`Album`、`Artist`。

## 5. 字体

| 字体 | 位置 | 来源与许可证 | 用途 | Sonora |
| --- | --- | --- | --- | --- |
| `logo` | `public/fonts/logo.woff2`、`font.css` | Poppins Medium（Indian Type Foundry），SIL OFL 1.1。由 `fc-scan` 读取字体名称表确认 | 侧栏标志文字 | 已复制到 `frontend/public/fonts/`，附 `OFL.txt` |
| 界面字体 | `Provider.vue` 的 `fontFamily` | 不随项目分发 | 全局 | 见下文 |

界面字体栈为 `v-sans, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"`（设置为“默认”时）。SPlayer 没有引入定义 `v-sans` 的 `vfonts`，因此实际字体由 `system-ui` 决定，随系统变化。视觉对照必须在同一环境中进行；[基线采集脚本](../scripts/baseline/splayer-player.mjs)在 manifest 中记录了 Chrome 实际使用的平台字体。

## 6. 图片

`public/images/` 下是 SPlayer 自带的图片，随 SPlayer 以 AGPL-3.0 分发。

| 文件 | 尺寸 | SPlayer 用途 | Sonora |
| --- | --- | --- | --- |
| `song.jpg` | 500×500 | 歌曲默认封面与封面加载占位 | 使用（歌单、我喜欢的默认封面也用它，见阶段 2 详情页头部）；已复制到 `frontend/public/images/` |
| `album.jpg` | 633×633 | 专辑默认封面 | 使用（侧栏歌单没有封面时也用它，与 SPlayer 一致）；已复制 |
| `artist.jpg` | 300×300 | 艺术家默认图片 | 使用（没有上传艺术家图片时）；已复制 |
| `record.png` | 60×60 | 唱片播放器样式 | 使用 |
| `pointer.png` | 251×409 | 唱片播放器的唱针 | 使用 |
| `avatar.jpg` | 100×100 | 默认头像 | 不使用：顶栏用户入口沿用 SPlayer 未登录时的 `Person` 图标头像（OIDC 不提供头像） |
| `video.jpg`、`vip.png`、`annual.png`、`avatar/` | — | 视频、网易会员、年度报告、关于页的贡献者头像 | 不使用 |

`public/icons/` 是应用图标与 favicon，属于 SPlayer 标志。阶段 2 已决定标志图形暂时复用。

## 7. 与 THIRD_PARTY_NOTICES.md 的同步

2026-10-25 已补：Logo 字体 Poppins Medium 的 OFL 声明，全文放在 `frontend/public/fonts/OFL.txt`；`frontend/src/assets/icons/` 中 Google Material Icons、Material Symbols 与 Phosphor 的版权与许可证。阶段 3.1 删除 `HeartBit` 后不再含 Phosphor 图标，该声明已移除，复制 `SortDown`、`SortUp` 时再加回。

阶段 3 引入播放器时还需加入：AMLL（AGPL-3.0-only）、IconaMoon（CC-BY-4.0，含署名）、Material Design Icons、Fluent UI System Icons（若使用）、PixiJS 与 SortableJS。`@imsyy/color-utils` 与 `lodash-es` 已在阶段 3.2 加入。
