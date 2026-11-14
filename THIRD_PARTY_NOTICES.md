# SPlayer 参考资源

前端初始化使用了 [SPlayer](https://github.com/SPlayer-Dev/SPlayer) 的部分视觉实现，原作者为 imsyy 及 SPlayer 贡献者。参考提交：`28b76bad9eb800e0c90dac8475d4769acfe2450b`。

保留其 AGPL-3.0 许可证全文于 [LICENSE](LICENSE)。复制或改编内容包括：

- `frontend/src/style/`：原全局样式、动画及布局/导航/菜单/用户组件样式。
- `frontend/src/components/Layout/Logo.vue`、`frontend/src/components/Global/SvgIcon.vue`：原标志与图标展示实现。
- `frontend/src/assets/icons/`、`frontend/public/fonts/`：参考项目对应 SVG 图标与 Logo 字体。
- `frontend/public/images/`：默认封面与艺术家图片（`album.jpg`、`song.jpg`、`artist.jpg`），以及唱片样式的唱针（`pointer.png`）。
- `frontend/src/views/Status/404.vue`、`500.vue`：原状态页。
- `frontend/src/components/Global/Provider.vue`、`frontend/src/utils/color.ts`、`frontend/src/assets/data/themeColor.json`：原全局主题与封面取色，照搬后去掉 Sonora 不做的功能。
- `frontend/src/stores/setting.ts`、`frontend/src/stores/status.ts`、`frontend/src/stores/music.ts`、`frontend/src/stores/data.ts`、`frontend/src/stores/shortcut.ts`：原设置、状态、当前歌曲、播放列表与快捷键的字段定义和默认值（保留部分）。
- `frontend/src/components/Setting/`、`frontend/src/types/settings.ts`、`frontend/src/components/Modal/ThemeConfig.vue`、`frontend/src/components/Modal/Setting/`、`frontend/src/components/UI/s-input.vue`、`frontend/src/composables/useShortcut.ts`：原设置界面（设置项配置、渲染、搜索、歌词预览、快捷键录制）、主题配置与各设置弹窗、输入框组件及页面内快捷键，照搬后去掉 Sonora 不做的设置项。
- `frontend/src/components/Player/`、`frontend/src/components/Global/TextContainer.vue`、`frontend/src/composables/useTimeFormat.ts`：原播放器组件，照搬后去掉 Sonora 不做的功能。
- `frontend/src/components/List/SongPlayList.vue`、`frontend/src/components/UI/VirtualScroll.vue`、`frontend/src/composables/List/useDragSort.ts`：原播放队列抽屉、虚拟滚动与拖动排序。
- `frontend/src/components/AMLL/`、`frontend/src/components/Global/PlayerProvider.vue`、`frontend/src/components/UI/s-image.vue`、`frontend/src/utils/style.ts`、`frontend/src/utils/lyric/lyricFontConfig.ts`：原 AMLL 歌词与背景封装、播放器主题覆盖、图片组件与歌词字体工具。
- `frontend/src/utils/lyric/`、`frontend/src/assets/data/exclude.ts`、`frontend/src/utils/calc.ts`、`time.ts`、`format.ts`、`helper.ts`、`instruction.ts`、`modal.ts`：原歌词解析与排除规则、歌词行计算及播放器用到的工具函数（保留部分）。
- `frontend/src/components/List/SongList.vue`、`ListDetail.vue`、`CoverList.vue`、`ArtistList.vue`、`frontend/src/components/Card/SongCard.vue`、`SongListCard.vue`、`frontend/src/components/Menu/SongListMenu.vue`、`frontend/src/composables/useSongMenu.ts`、`frontend/src/composables/List/useListDetail.ts`、`useListSearch.ts`、`useListScroll.ts`、`useListActions.ts`：原歌曲列表、详情页头部、封面与艺术家网格、卡片、歌曲右键菜单与列表组合函数。
- `frontend/src/components/Modal/`（`CreatePlaylist.vue`、`UpdatePlaylist.vue`、`PlaylistAdd.vue`、`JumpArtist.vue`、`CopyLyrics.vue`、`Equalizer.vue`）、`frontend/src/utils/rules.ts`：原新建与编辑歌单、添加到歌单、跳转到歌手、复制歌词与均衡器的弹窗及表单规则（保留部分）。
- `frontend/src/components/Search/`：原搜索框、搜索建议与搜索历史，建议改为曲库搜索。
- `frontend/src/views/History.vue`、`frontend/src/views/List/`、`frontend/src/views/Artist/`、`frontend/src/views/Search/`、`frontend/src/views/Home/`：原最近播放、专辑与歌单详情、我喜欢、艺术家详情、搜索结果与首页，按 Sonora 的页面清单改编。
- `frontend/src/utils/meta.ts`、`auth.ts`：原排序选项、图片工具栏与喜欢等入口（保留部分）。
- `frontend/src/core/player/`：按原播放控制器、播放模式管理、媒体会话与歌词管理改编，方法名与处理流程沿用原项目。
- `frontend/src/core/audio-player/`：原 Web Audio 播放引擎（`BaseAudioPlayer`、`AudioElementPlayer`、`AudioEffectManager`、`IPlaybackEngine` 与 `core/automix/SharedAudioContext.ts`），去掉 Sonora 不用的引擎与自动混音。
- `frontend/src/utils/TypedEventTarget.ts`：原项目收录的类型安全 EventTarget，来自 [typescript-event-target](https://github.com/DerZade/typescript-event-target)（Copyright (c) 2022 Jonas "DerZade" Schade，MIT 许可证），未作修改。
- 主布局、导航和首页展示结构按原项目对应实现改编，业务逻辑重新实现。

2026-10-23 确定：标志图形暂时复用；界面中的 SPlayer 名称与文字均替换为 Sonora。Sonora 不迁入网易来源、认证或业务接口。各依赖的许可证仍以各包声明为准。

# Poppins

`frontend/public/fonts/logo.woff2` 是 Logo 字体 Poppins Medium（Indian Type Foundry，4.004），SIL Open Font License 1.1。SPlayer 随附的子集只有 “SPlayer” 的 7 个字母；2026-11-03 改为从 [google/fonts](https://github.com/google/fonts/tree/main/ofl/poppins) 的 `Poppins-Medium.ttf` 重新取子集，加入 “Sonora” 用到的 `n`、`o`（`pyftsubset Poppins-Medium.ttf --text=PSaelryon --layout-features='' --flavor=woff2`），原有字形的轮廓与字宽不变。许可证全文随字体放在 [frontend/public/fonts/OFL.txt](frontend/public/fonts/OFL.txt)：

> Copyright 2020 The Poppins Project Authors (https://github.com/itfoundry/Poppins)

# 图标

`frontend/src/assets/icons/` 的图标复制自 SPlayer，其中可确认来源的属于以下图标集（核对方法见[视觉资源与许可清单](docs/visual-assets.md#4-图标)）：

- [Google Material Icons](https://github.com/google/material-design-icons)、[Material Symbols](https://github.com/google/material-design-icons)：Copyright Google LLC，Apache License 2.0。
- [Material Design Icons](https://github.com/Templarian/MaterialDesign)：`TextPlay.svg`，Copyright Pictogrammers，Apache License 2.0。
- [Fluent UI System Icons](https://github.com/microsoft/fluentui-system-icons)：`Calendar-Empty.svg`，Copyright (c) 2020 Microsoft Corporation，MIT 许可证。
- [IconaMoon](https://github.com/dariushhpg1/IconaMoon)：`Repeat.svg`、`RepeatSong.svg`、`Shuffle.svg`，作者 Dariush Habibpour，[CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)。经 SPlayer 收录，Sonora 未作修改。

其余图标来源尚未确认，按 SPlayer 的 AGPL-3.0 处理。

# 运行时依赖

播放端随产物分发以下与视觉相关的 npm 包，许可证全文见各包：

- [Naive UI](https://github.com/tusen-ai/naive-ui)：MIT。
- [@material/material-color-utilities](https://github.com/material-foundation/material-color-utilities)：Copyright Google LLC，Apache License 2.0。
- [@imsyy/color-utils](https://github.com/imsyy/color-utils)：Copyright (c) 2024 imsyy，MIT 许可证（package.json 标为 ISC，以包内 LICENSE 为准）。
- [lodash-es](https://github.com/lodash/lodash)：Copyright OpenJS Foundation and other contributors，MIT 许可证。
- [@applemusic-like-lyrics/lyric](https://github.com/amll-dev/applemusic-like-lyrics) 及其依赖 `@applemusic-like-lyrics/ttml`：歌词解析，AGPL-3.0-only。
- [@applemusic-like-lyrics/core](https://github.com/amll-dev/applemusic-like-lyrics)：歌词播放器与流动背景，AGPL-3.0-only。
- [PixiJS](https://github.com/pixijs/pixijs)（`@pixi/*`）：AMLL 流动背景的渲染，Copyright (c) 2013-2023 Mathew Groves, Chad Engler，MIT 许可证。其中 `@pixi/filter-bulge-pinch` 来自 [pixi-filters](https://github.com/pixijs/filters)，Copyright (c) 2013-2017 Mathew Groves, Chad Engler，MIT 许可证。
- [Day.js](https://github.com/iamkun/dayjs)：Copyright (c) 2018-present, iamkun，MIT 许可证。
- [Fuse.js](https://github.com/krisk/Fuse)：Copyright 2017 Kirollos Risk，Apache License 2.0。
- [SortableJS](https://github.com/SortableJS/Sortable)：设置中首页栏目的拖动排序，Copyright (c) 2019 All contributors to Sortable，MIT 许可证。
- [@vueuse/integrations](https://github.com/vueuse/vueuse)：`useSortable`，Copyright (c) 2019-PRESENT Anthony Fu，MIT 许可证。

# ThumbHash

`backend/internal/thumbhash/` 的编码器移植自 Evan Wallace 的 [ThumbHash](https://github.com/evanw/thumbhash) 参考实现，MIT 许可证：

> Copyright (c) 2023 Evan Wallace
>
> Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
>
> The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

# PixelPlayerOSS

`android/` 是 Sonora 的安卓客户端，从 [PixelPlayerOSS](https://github.com/PixelPlayerHQ/PixelPlayerOSS) 复制部分代码二次开发，参考 tag `v0.4.0-alpha.53`（提交 `a1f8fce42e3b81a65e3a5eda005d2bda654fff67`）。`android/` 按 GPL-3.0 发布，许可证全文见 [android/LICENSE](android/LICENSE)；仓库其余部分仍是 AGPL-3.0。

> PixelPlayerOSS
> Copyright (C) 2026 Theo Vilardo

复制或改编内容包括（包名都改为 `top.onlyquant.sonora`）：

- `data/service/player/`：`DualPlayerEngine`（去掉云端流代理与 URI 解析）、`QueueWindowPolicy`、`AudioDecoderPolicy`、`HiFiCapabilityChecker`、`HiResSampleRateCapAudioProcessor`、`SurroundDownmixProcessor`、`UnsupportedAudioListener`，以及对应的单元测试。
- `data/service/MusicService.kt`：按原版缩减，只保留会话部分。
- `data/service/ReplayGainProcessor.kt`（阶段 8 第 4 项，音量均衡）：按原版改编，增益取自服务端的 `gainDb`，不读文件标签，去掉专辑增益开关。
- `data/model/AudioOutputMode.kt`、`data/model/Transition.kt`（去掉 `TransitionRule`）、`utils/Envelope.kt`。
- `android/app/build.gradle.kts`、`android/gradle/`：构建配置与依赖版本取自原版，Gradle Wrapper 复制自原版。
- `ui/theme/`：`Color.kt`、`Shape.kt`、`Type.kt` 照搬，`Theme.kt` 去掉 AMOLED 预设与正在播放取色的覆盖。
- `res/mipmap-*/`、`res/drawable/new_monochrome.xml`、`pixelplayer_base_monochrome.xml`、`monochrome_player.xml`：PixelPlayer 的启动图标、标志与通知图标，照搬（Eric 决定沿用 PixelPlayer 的图标，名称仍是 Sonora）。
- `res/font/gflex_variable.ttf`：Google Sans Flex（Copyright 2015 Google LLC，SIL Open Font License 1.1，许可证全文见 [frontend/public/fonts/OFL.txt](frontend/public/fonts/OFL.txt)），照搬。
- 播放器面板与队列（阶段 8 第 4 项）：`presentation/components/` 下的 `UnifiedPlayerSheet*`、`QueueBottomSheet`、`player/`、`scoped/`、`subcomps/PlayingEqIcon` 等，`presentation/viewmodel/` 下的播放、队列、睡眠定时、主题与音频输出的 StateHolder，`data/service/` 的睡眠定时与 `CoilBitmapLoader`，`utils/` 的 `MediaItemBuilder`、`QueueUtils` 等，`ui/theme/ColorRoles.kt`，以及 `res/values/strings_player.xml`（取自原版 zh-CN 文案）与对应图标，照搬；`PlayerViewModel`、`PlayerUiState`、`UserPreferencesRepository`、`AlbumArtUtils`、`AudioMetaUtils` 按原版缩减；`presentation/main/MainUI.kt` 按原版 `MainActivity.MainUI` 缩减。
- 曲库（阶段 8 第 4 项）：`presentation/screens/` 下的 `LibraryScreen`、`LibrarySongsTab`、`LibraryMediaTabs`、`LibraryEmptyState`、`LibraryPlaybackAwareSongItem`、`LibraryScrollRestore`、`LibrarySortScrollReset`、`TabAnimation`、`search/components/` 的 `GenreTypography` 与 `GenreTitleAutoSize`，`presentation/components/` 下的 `LibrarySortBottomSheet`、`MultiSelectionBottomSheet`、`SongInfoBottomSheet`、`ShimmerBox`、`PlayerSheetAwareBackHandler`、`ExpressiveScrollBarLabelResolvers` 与 `subcomps/` 的 `EnhancedSongListItem`、`LibraryActionRow`、`SelectionActionRow`、`AutoSizingText`、`TightWrapText`，`presentation/selection/`、`presentation/library/KeepUnscrolledListAtTop`、`presentation/utils/` 的 `SegmentedItemShape` 与 `SelectionScale`、`ui/theme/ShapeCache`，`data/model/` 的 `SortOption`、`LibraryTabId`、`StorageFilter`，`presentation/viewmodel/` 的 `MultiSelectionStateHolder`、`LibraryTabsStateHolder`、`LibraryStateHolder`，以及 `res/values/strings_library.xml`（取自原版 zh-CN 文案）与对应图标；`LibraryScreen`、`SongInfoBottomSheet`、`MultiSelectionBottomSheet`、`LibraryActionRow`、`LibraryStateHolder` 等按原版缩减。
- 专辑与艺术家详情页（阶段 8 第 4 项）：`presentation/screens/` 下的 `AlbumDetailScreen`、`ArtistDetailScreen`，`presentation/components/` 下的 `MediaDetailHero`、`CollapsibleCommonTopBar`、`ExpressiveTopBarContent`、`ExpressiveDropdownMenu`、`ScreenWrapper`，`presentation/navigation/` 的 `Transitions`、`MainRootRoutes`，`presentation/viewmodel/` 的 `AlbumDetailViewModel`、`ArtistDetailViewModel`、`ArtistAlbumSectionExpansion`，以及 `res/values/strings_detail.xml`（取自原版 zh-CN 文案）；两个详情页与 ViewModel 按原版缩减，`presentation/navigation/AppNavigation.kt` 按原版缩减。
- 首页（阶段 8 第 4 项）：`presentation/screens/HomeScreen`、`presentation/components/` 下的 `AlbumArtCollage`、`CollagePatterns`、`RecentlyPlayedSection`、`GradientTopBar`（只留 `HomeGradientTopBar`）、`presentation/model/RecentlyPlayedSongUi`、`data/preferences/CollagePattern`、`utils/shapes/RoundedStarShape`、`res/drawable/rounded_music_note_24.xml`，以及 `res/values/strings_home.xml` 中取自原版 zh-CN 的文案；`HomeScreen`、`GradientTopBar`、`RecentlyPlayedSongUi` 按原版缩减。
- 搜索（阶段 8 第 4 项）：`presentation/screens/SearchScreen`、`presentation/viewmodel/SearchStateHolder`、`data/model/` 的 `SearchFilterType` 与 `SearchResultItem`、`res/drawable/rounded_check_circle_24.xml`，以及 `res/values/strings_search.xml`（取自原版 zh-CN 文案）；`SearchScreen` 去掉流派浏览、搜索历史与专辑、歌单、流派的多选，`SearchStateHolder` 去掉搜索历史。
- `presentation/auth/LoginScreen.kt`：按原版 `presentation/navidrome/auth/NavidromeLoginActivity.kt` 的 `NavidromeLoginScreen` 改编；文案参照原版 `values-zh-rCN/strings_auth.xml`。

`backend/internal/auth/assets/gflex.woff2` 是安卓设备登录确认页的字体，取自 PixelPlayerOSS 的 `app/src/main/res/font/gflex_variable.ttf`（Google Sans Flex 3.007，Copyright 2015 Google LLC，SIL Open Font License 1.1，许可证全文见 [frontend/public/fonts/OFL.txt](frontend/public/fonts/OFL.txt)）。与原版的 `RoundedSans` 一样把 ROND 轴固定为 100，其余轴取默认值，只保留 300–700 的字重，并只取拉丁字符子集（`fonttools varLib.instancer … wdth=100 opsz=drop GRAD=0 ROND=100 slnt=0 wght=300:700`，再 `pyftsubset --unicodes='U+0020-007E,U+00A0-00FF,U+2018-201D,U+2026' --layout-features='*' --flavor=woff2`）。确认页的图标是 Material Icons Rounded（Apache-2.0）的 `PhoneAndroid`、`Person`、`Check`、`Close`。
