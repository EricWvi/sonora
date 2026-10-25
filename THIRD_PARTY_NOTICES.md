# SPlayer 参考资源

前端初始化使用了 [SPlayer](https://github.com/SPlayer-Dev/SPlayer) 的部分视觉实现，原作者为 imsyy 及 SPlayer 贡献者。参考提交：`28b76bad9eb800e0c90dac8475d4769acfe2450b`。

保留其 AGPL-3.0 许可证全文于 [LICENSE](LICENSE)。复制或改编内容包括：

- `frontend/src/style/`：原全局样式、动画及布局/导航/菜单/用户组件样式。
- `frontend/src/components/Layout/Logo.vue`、`frontend/src/components/Global/SvgIcon.vue`：原标志与图标展示实现。
- `frontend/src/assets/icons/`、`frontend/public/fonts/`：参考项目对应 SVG 图标与 Logo 字体。
- `frontend/public/images/`：默认封面与艺术家图片（`album.jpg`、`song.jpg`、`artist.jpg`），以及唱片样式的唱针（`pointer.png`）。
- `frontend/src/views/Status/404.vue`、`500.vue`：原状态页。
- `frontend/src/components/Global/Provider.vue`、`frontend/src/utils/color.ts`、`frontend/src/assets/data/themeColor.json`：原全局主题与封面取色，照搬后去掉 Sonora 不做的功能。
- `frontend/src/stores/setting.ts`、`frontend/src/stores/status.ts`、`frontend/src/stores/music.ts`、`frontend/src/stores/data.ts`：原设置、状态、当前歌曲与播放列表的字段定义和默认值（保留部分）。
- `frontend/src/components/Player/`、`frontend/src/components/Global/TextContainer.vue`、`frontend/src/composables/useTimeFormat.ts`：原播放器组件，照搬后去掉 Sonora 不做的功能。
- `frontend/src/components/List/SongPlayList.vue`、`frontend/src/components/UI/VirtualScroll.vue`、`frontend/src/composables/List/useDragSort.ts`：原播放队列抽屉、虚拟滚动与拖动排序。
- `frontend/src/components/AMLL/`、`frontend/src/components/Global/PlayerProvider.vue`、`frontend/src/components/UI/s-image.vue`、`frontend/src/utils/style.ts`、`frontend/src/utils/lyric/lyricFontConfig.ts`：原 AMLL 歌词与背景封装、播放器主题覆盖、图片组件与歌词字体工具。
- `frontend/src/utils/lyric/`、`frontend/src/assets/data/exclude.ts`、`frontend/src/utils/calc.ts`、`time.ts`、`format.ts`、`helper.ts`、`instruction.ts`、`modal.ts`：原歌词解析与排除规则、歌词行计算及播放器用到的工具函数（保留部分）。
- `frontend/src/components/List/SongList.vue`、`ListDetail.vue`、`CoverList.vue`、`ArtistList.vue`、`frontend/src/components/Card/SongCard.vue`、`SongListCard.vue`、`frontend/src/components/Menu/SongListMenu.vue`、`frontend/src/composables/useSongMenu.ts`、`frontend/src/composables/List/useListDetail.ts`、`useListSearch.ts`、`useListScroll.ts`、`useListActions.ts`：原歌曲列表、详情页头部、封面与艺术家网格、卡片、歌曲右键菜单与列表组合函数。
- `frontend/src/components/Search/`：原搜索框、搜索建议与搜索历史，建议改为曲库搜索。
- `frontend/src/views/History.vue`、`frontend/src/views/List/`、`frontend/src/views/Artist/`、`frontend/src/views/Search/`、`frontend/src/views/Home/`：原最近播放、专辑与歌单详情、我喜欢、艺术家详情、搜索结果与首页，按 Sonora 的页面清单改编。
- `frontend/src/utils/meta.ts`、`auth.ts`：原排序选项、图片工具栏与喜欢等入口（保留部分）。
- `frontend/src/core/player/`：按原播放控制器、播放模式管理、媒体会话与歌词管理改编，方法名与处理流程沿用原项目。
- `frontend/src/core/audio-player/`：原 Web Audio 播放引擎（`BaseAudioPlayer`、`AudioElementPlayer`、`AudioEffectManager`、`IPlaybackEngine` 与 `core/automix/SharedAudioContext.ts`），去掉 Sonora 不用的引擎与自动混音。
- `frontend/src/utils/TypedEventTarget.ts`：原项目收录的类型安全 EventTarget，来自 [typescript-event-target](https://github.com/DerZade/typescript-event-target)（Copyright (c) 2022 Jonas "DerZade" Schade，MIT 许可证），未作修改。
- 主布局、导航和首页展示结构按原项目对应实现改编，业务逻辑重新实现。

2026-10-23 确定：标志图形暂时复用；界面中的 SPlayer 名称与文字均替换为 Sonora。Sonora 不迁入网易来源、认证或业务接口。各依赖的许可证仍以各包声明为准。

# Poppins

`frontend/public/fonts/logo.woff2` 是 SPlayer 随附的 Logo 字体，经字体名称表确认为 Poppins Medium（Indian Type Foundry），SIL Open Font License 1.1。许可证全文随字体放在 [frontend/public/fonts/OFL.txt](frontend/public/fonts/OFL.txt)：

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

# ThumbHash

`backend/internal/thumbhash/` 的编码器移植自 Evan Wallace 的 [ThumbHash](https://github.com/evanw/thumbhash) 参考实现，MIT 许可证：

> Copyright (c) 2023 Evan Wallace
>
> Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
>
> The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
