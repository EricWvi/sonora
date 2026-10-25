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
| 播放器主题覆盖（封面色） | `src/components/Global/PlayerProvider.vue` | 78 | 照搬 | 已照搬（阶段 3.5，2026-10-25） |
| 封面取色与配色方案 | `src/utils/color.ts`、`src/assets/data/themeColor.json` | 325 | 照搬 | 已照搬，去掉任务栏歌词取色；生成结果与原项目一致（阶段 3.2） |
| 字体大小工具 | `src/utils/style.ts` | 12 | 照搬 | 已照搬，另照搬歌词字体配置 `utils/lyric/lyricFontConfig.ts`（2026-10-25） |
| 播放器：底部播放栏、全屏播放器、封面、歌词、背景、控制栏、频谱 | `src/components/Player/` | 6097 | 严格复刻，去掉评论、音质与网易相关元素（[阶段 2 决定](phase/phase-2-player-baseline.md#播放器里与-sonora-不对应的元素)） | 底部播放栏与全屏播放器（`FullPlayer.vue`、`PlayerMenu.vue`、`PlayerControl.vue`、`PlayerMeta/`、`PlayerLyric/`、`PlayerSpectrum.vue`）已照搬，截图对照见阶段 3.5（2026-10-25）。不迁入 `FullPlayerMobile.vue`、`PlayerComment.vue`、`PersonalFM.vue` |
| AMLL 封装：歌词播放器、流动背景 | `src/components/AMLL/` | 512 | 严格复刻 | 已照搬（阶段 3.5，2026-10-25） |
| 音频分析：频谱与低频音量（全屏播放器的频谱、背景随低频响应） | `src/core/audio-player/` 中的 `AudioEffectManager.ts`、`BaseAudioPlayer.ts`、`AudioElementPlayer.ts`、`IPlaybackEngine.ts` | 1320 | 照搬 Web Audio 引擎，去掉 MPV、FFmpeg 与自动混音 | 已照搬（阶段 4.4，2026-10-25），分析器参数与低频音量的计算不变 |
| 播放队列抽屉 | `src/components/List/SongPlayList.vue`、`src/composables/List/useDragSort.ts` | 786 | 严格复刻 | 已照搬，去掉电台分支；截图对照逐像素一致（阶段 3.5，2026-10-25） |
| 歌词解析 | `src/utils/lyric/` | 1914 | 照搬 LRC、TTML 相关部分；去掉 YRC、QRC 与在线歌词源 | 已照搬 `parseLrc.ts`、`lyricParser.ts`（去掉 QRC 与 `lyricLinesToTTML`）、`parseTTML.ts`、`lyricStripper.ts`、`lyricFormat.ts`，处理流程见 `core/player/LyricManager.ts`（2026-10-25） |
| 歌曲列表、详情页头部 | `src/components/List/SongList.vue`、`ListDetail.vue` | 1425 | 视觉语言一致，结构按阶段 2 决定 | 已照搬（阶段 3.4，2026-10-25）：排序改由服务端完成，去掉电台、大小列、移动端菜单、播放量、标签与评论；增加碟号分组标题与专辑年份、总时长、歌单创建与更新时间等信息项（待 Eric 确认）。另照搬右键菜单 `SongListMenu.vue`、`useSongMenu.ts`（条目按阶段 2）与 `CoverList.vue`、`ArtistList.vue` |
| 卡片 | `src/components/Card/` | 831 | 视觉语言一致 | 已照搬 `SongCard.vue`（歌曲行，去掉音质、原唱、特权、云盘、MV、脏标标签）与 `SongListCard.vue`（首页快捷入口）；`SongDataCard.vue` 不迁入（阶段 3.4，2026-10-25） |
| 布局、导航、侧栏、顶栏 | `src/components/Layout/`、`src/layout/` | 1505 | 视觉语言一致 | 已改编主框架、导航与主题初始化 |
| 通用组件：图片、输入框、虚拟滚动 | `src/components/UI/` | 737 | 照搬 | 已照搬 `VirtualScroll.vue`、`s-image.vue`；`s-input.vue` 未用到（搜索框与列表内搜索在 SPlayer 中用的是 `n-input`） |
| 搜索框、搜索建议与搜索历史 | `src/components/Search/` | 816 | 照搬，建议改为曲库搜索 | 已照搬（阶段 3.4，2026-10-25），去掉热搜榜、推荐关键词、分享链接识别与搜索框右键菜单 |
| 页面：最近播放、专辑与歌单详情、我喜欢、艺术家详情、搜索结果、首页 | `src/views/History.vue`、`views/List/`、`views/Artist/`、`views/Search/`、`views/Home/` | — | 视觉语言一致，结构按阶段 2 决定 | 已改编（阶段 3.4，2026-10-25），列表组合函数 `useListDetail`、`useListSearch`、`useListScroll`、`useListActions` 一并照搬 |
| 弹窗：新建、编辑歌单与添加到歌单 | `src/components/Modal/CreatePlaylist.vue`、`UpdatePlaylist.vue`、`PlaylistAdd.vue`、`src/utils/rules.ts` | 567 | 照搬 | 已照搬（阶段 5，2026-10-25）：新建用其本地歌单的表单（名称、描述），编辑去掉网易的歌单分类，添加到歌单列出自己的歌单、不含“我喜欢的音乐”；`rules.ts` 只取 `textRule` |
| 弹窗：跳转到歌手、复制歌词、均衡器 | `src/components/Modal/JumpArtist.vue`、`CopyLyrics.vue`、`Equalizer.vue` | 509 | 照搬 | 已照搬（阶段 5，2026-10-25）：跳转到歌手去掉按名称在线搜索，图片取艺术家图片；均衡器去掉只在 Electron 中启用的限制。`CopySongInfo.vue`（歌曲详情复制）不迁入，Eric 决定不要这个功能 |
| 外观与歌词设置项 | `src/components/Setting/config/appearance.ts`、`lyric.ts`、`components/LyricPreview.vue` | 1735 | 照搬保留的设置项（[阶段 2 的设置表](phase/phase-2-player-baseline.md#设置)） | 未开始 |

## 3. 运行时依赖

只列与视觉相关的依赖。版本是 SPlayer 锁定的版本；Sonora 引入时使用相同的主版本，升级要先确认视觉不变。

| 包 | 用途 | SPlayer 锁定版本 | 许可证 | Sonora |
| --- | --- | --- | --- | --- |
| `naive-ui` | 组件库 | 2.43.2（固定） | MIT | 已用，同版本 |
| `@material/material-color-utilities` | 主题色与封面取色（`themeFromSourceColor`、`QuantizerCelebi`、`Score`） | 0.4.0 | Apache-2.0 | 已用，同版本 |
| `@imsyy/color-utils` | `rgbToHex` | 1.0.2 | package.json 写 ISC，包内 LICENSE 是 MIT 全文 | 已用，同版本（2026-10-25） |
| `lodash-es` | `chunk`（封面取色）及 SPlayer 各处的 `debounce`、`cloneDeep` 等 | 4.18.1 | MIT | 已用，同版本（2026-10-25） |
| `@applemusic-like-lyrics/core` | AMLL 歌词播放器与流动背景 | 0.5.1 | AGPL-3.0-only | 已用，同版本（2026-10-25）；最新为 0.6.0。它的依赖 `@ungap/structured-clone` 解析为 1.4.0（SPlayer 锁定 1.3.1，ISC），`bezier-easing`、`deep-freeze`、`gl-matrix` 与 SPlayer 相同 |
| `@applemusic-like-lyrics/lyric` | LRC、TTML 等歌词解析 | 1.0.1 | AGPL-3.0-only | 已用，同版本（2026-10-25），播放端用它解析 TTML，fixture 脚本用它转换歌词。其依赖 `@applemusic-like-lyrics/ttml` 1.0.1（AGPL-3.0-only）与 SPlayer 相同，`pako` 解析为 2.2.0（SPlayer 锁定 2.1.0，MIT AND Zlib，用于 QRC 解压，与视觉无关） |
| `@pixi/app`、`@pixi/core`、`@pixi/display`、`@pixi/sprite`、`@pixi/filter-blur`、`@pixi/filter-color-matrix` | AMLL 的对等依赖 | 7.4.3 | MIT | 已用，同版本（2026-10-25） |
| `@pixi/filter-bulge-pinch` | AMLL 的对等依赖 | 5.1.1 | MIT | 已用，同版本（2026-10-25） |
| `sortablejs` | 设置中首页栏目的拖动排序（经 `@vueuse/integrations` 的 `useSortable`） | 1.15.7 | MIT | 阶段 5 接入设置时引入。播放队列与歌曲列表的拖动排序用 SPlayer 自己的 `useDragSort`，不依赖它 |
| `@vueuse/core` | 组合式工具（自动隐藏计时、媒体查询等） | 14.3.0 | MIT | 已用 |
| `dayjs` | 播放时间显示（`msToTime`） | 1.11.21 | MIT | 已用，同版本（2026-10-25） |
| `fuse.js` | 列表内模糊搜索（`fuzzySearch`） | 7.4.1 | Apache-2.0 | 已用，同版本（阶段 3.4，2026-10-25） |

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

`frontend/src/assets/icons/` 现有 58 个图标，与 SPlayer 的文件逐字节一致（2026-10-25 阶段 3.4 页面模板后，`cmp` 核对）：Google Material Icons 22 个（`Add`、`DarkTheme`、`Delete`、`EditNote`、`Favorite`、`FavoriteBorder`、`FormatList`、`LightTheme`、`List`、`Menu`、`Music`、`NavigateBefore`、`NavigateNext`、`Pause`、`Person`、`Play`、`Refresh`、`Right`、`SkipNext`、`SkipPrev`、`Time`、`Update`）、Material Symbols 12 个（`AddList`、`Copy`、`DeleteSweep`、`Eq`、`Expand`、`Home`、`MusicList`、`PlayList`、`Search`、`VolumeDown`、`VolumeMute`、`VolumeUp`）、Material Symbols Light 1 个（`DropDown`）、Material Design Icons 1 个（`TextPlay`）、Fluent UI System Icons 1 个（`Calendar-Empty`）、IconaMoon 3 个（`Repeat`、`RepeatSong`、`Shuffle`），近似项 10 个（`Artist`、`AutoTheme`、`History`、`PlayNext`、`PlaylistAdd`、`Power`、`Replay5`、`SearchOff`、`Settings`、`VolumeOff`），未确认 8 个（`Album`、`Controls`、`Down`、`Forward5`、`Fullscreen`、`FullscreenExit`、`Location`、`Up`）。侧栏的歌曲、专辑、艺术家图标沿用 SPlayer 流媒体页面标签的 `Music`、`Album`、`Artist`。

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
| `record.png` | 60×60 | 无（SPlayer 源码中没有引用，2026-10-25 核对） | 不使用 |
| `pointer.png` | 251×409 | 唱片播放器的唱针 | 使用；已复制 |
| `avatar.jpg` | 100×100 | 默认头像 | 不使用：顶栏用户入口沿用 SPlayer 未登录时的 `Person` 图标头像（OIDC 不提供头像） |
| `video.jpg`、`vip.png`、`annual.png`、`avatar/` | — | 视频、网易会员、年度报告、关于页的贡献者头像 | 不使用 |

`public/icons/` 是应用图标与 favicon，属于 SPlayer 标志。阶段 2 已决定标志图形暂时复用。

## 7. 与 THIRD_PARTY_NOTICES.md 的同步

2026-10-25 已补：Logo 字体 Poppins Medium 的 OFL 声明，全文放在 `frontend/public/fonts/OFL.txt`；`frontend/src/assets/icons/` 中 Google Material Icons、Material Symbols 与 Phosphor 的版权与许可证。阶段 3.1 删除 `HeartBit` 后不再含 Phosphor 图标，该声明已移除，复制 `SortDown`、`SortUp` 时再加回。

`@imsyy/color-utils` 与 `lodash-es` 已在阶段 3.2 加入；IconaMoon 的署名在阶段 3.5 底部播放栏时加入；AMLL core、PixiJS 与 Material Design Icons 在阶段 3.5 全屏播放器时加入（2026-10-25）。Fluent UI System Icons（专辑详情的年份图标 `Calendar-Empty`）与 Fuse.js 在阶段 3.4 页面模板时加入（2026-10-25）。SortableJS 在阶段 5 接入设置时加入。
