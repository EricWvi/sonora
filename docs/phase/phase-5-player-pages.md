# 阶段 5：播放端功能页面

状态：进行中（2026-10-25）。前置：阶段 4 的认证、API 与播放闭环（已完成，2026-10-25）。

## 工作包

- [x] 歌曲、专辑、艺术家，以及专辑、艺术家详情接入真实曲库数据。阶段 3.4 已按播放端 API 实现，阶段 4 接上音频。运行确认（2026-10-25）：开发模式临时实例（`backend/.data/data` 的副本，337 首歌、310 张专辑、194 位艺术家），无头 Chromium。
  - 三个列表页的数量与 API 的 `total` 相同。数量用 `n-number-animation` 从 0 滚动到总数，与 SPlayer 相同，动画结束前读到的是中间值。
  - 专辑详情（《邓丽君精选1》5 首、1982、19:14）、艺术家详情（Aki阿杰的歌曲与专辑标签页各 3 项）与 API 一致。
  - 页面没有报错，也没有失败的请求。
- [x] 首页接入最近添加的专辑与最近播放。运行确认（2026-10-25，同一实例）：快捷入口显示曲库的 337 首歌；“最近添加的专辑”的 12 张与 `albums?sort=createTime&order=desc` 的顺序相同；“最近播放”与 `history` 相同。
- [x] 歌单：新建、编辑、删除、添加/移除歌曲、拖动排序（服务端 409 时刷新）；由 Go 服务持久化，写入后重新读取验证。实现（2026-10-25）：
  - `DataSource` 增加歌单与最近播放的写方法（[source.ts](../../frontend/src/api/source.ts)），fixture 在内存中实现同样的规则（重算封面与数量，顺序不一致时 409）。
  - 弹窗照搬 SPlayer 的 [CreatePlaylist.vue](../../frontend/src/components/Modal/CreatePlaylist.vue)、[UpdatePlaylist.vue](../../frontend/src/components/Modal/UpdatePlaylist.vue)、[PlaylistAdd.vue](../../frontend/src/components/Modal/PlaylistAdd.vue)。新建用其本地歌单的表单（名称、描述）；编辑去掉网易的歌单分类；添加到歌单不列出“我喜欢的音乐”（Eric 确认，与 SPlayer 的本地歌单相同），所选歌曲都已在歌单中时提示“所选歌曲已在歌单中”（沿用其本地歌单）。
  - 删除歌曲、删除歌单的确认与提示照搬 SPlayer（[utils/auth.ts](../../frontend/src/utils/auth.ts)）。删除歌单后返回上一页，与 SPlayer 的本地歌单相同；SPlayer 的网易歌单删除后仍停在原页面。
  - 拖动排序先更新视图，再保存完整顺序。多次拖动按先后依次保存，避免较早的请求后到。保存失败时重新读取歌单：409 提示“歌单已在别处修改，已重新载入”，其他错误照搬 SPlayer 的“保存排序失败，请重试”。
  - 写入失败时展示服务端的说明，会话失效时回到登录页；成功后重新读取侧栏的歌单（封面、数量）。侧栏的“+”打开新建歌单。
  - 运行确认（2026-10-25，开发模式临时实例，无头 Chromium，真实点击）：
    - 侧栏“+”新建（空名称提示“请填写必要信息”），右键“添加到歌单”加入 3 首，重复添加时提示已在歌单中；
    - 编辑名称与描述后，详情头部与侧栏同时更新；
    - 拖动第 1 首到第 3 首之后，页面与服务端顺序一致；
    - 经 API 移除一首后再拖动，服务端返回 409，页面提示后重新载入，与服务端一致；
    - 右键“从歌单中删除”确认后，列表与数量更新；删除歌单后回到上一页，侧栏不再显示。
    - 页面没有报错，控制台只有浏览器记录的那次 409。fixture（`library`）中删除与添加同样正常。
- [x] 我喜欢与最近播放：按用户隔离，由 Go 服务持久化，写入未确认时不显示为已保存。隔离与持久化见[阶段 4.2](phase-4-playback.md)，喜欢的写入见 4.4。
  - 清空最近播放接上 `DELETE history`（2026-10-25），确认对话框照搬 SPlayer；服务端确认后才清空列表，运行确认同上。
  - “我喜欢”页保持现状（Eric 确认，2026-10-25）：页面内取消喜欢时保留这一行（与 SPlayer 相同），每次重新进入都从服务端重新读取。SPlayer 只在数量变化时刷新。
- [x] 弹窗（先后顺序的第 ③ 步）：照搬 SPlayer 的 [JumpArtist.vue](../../frontend/src/components/Modal/JumpArtist.vue)、[CopyLyrics.vue](../../frontend/src/components/Modal/CopyLyrics.vue)、[Equalizer.vue](../../frontend/src/components/Modal/Equalizer.vue)，入口在 [utils/modal.ts](../../frontend/src/utils/modal.ts)（2026-10-25）。
  - 跳转到歌手：去掉按名称在线搜索的分支（Sonora 的歌手都带 ID），图片取管理后台上传的艺术家图片，没有时为 SPlayer 的默认图片。
  - 歌曲详情复制（“复制更多信息”）：照搬并运行确认后，Eric 决定不要这个功能（2026-10-25）：去掉别名、链接后只剩 Sonora 内部的 ID 与列表中已有的信息，没有用处。弹窗与右键菜单、底部栏“更多操作”中的入口一并删除，“更多操作”只剩“复制歌曲名称”。
  - 复制歌词：歌名取 `Track.title`，其余不变。
  - 均衡器：SPlayer 只在 Electron 中启用（网页版开关与滑块禁用），Sonora 按阶段 2 的决定在网页中启用（Eric 确认，2026-10-25）；status store 补上 `setEqEnabled`、`setEqBands`、`setEqPreset`，变化由 `initPlayer` 应用到播放引擎。实际听感待 Eric 在桌面 Chrome 上确认（与输出设备一起）。
  - 运行确认（2026-10-25，开发模式临时实例，无头 Chromium，真实点击）：三位歌手的歌曲“查看艺术家”列出三人，点击后跳到艺术家页并关闭弹窗（曲库中没有艺术家上传过图片，三张都是默认图）；全屏播放器“复制歌词”列出 53 行，复制所选两行与“——《你的答案》 - 阿冗”；均衡器关闭时预设不可点，开启并选“重低音”后播放引擎 10 段滤波的增益为 `6, 6, 8, 2, 0…`，关闭后回到 0，store 保留预设与频段。页面没有报错。与 SPlayer 的上述差异 Eric 已确认（2026-10-25）。
  - 删除“复制更多信息”后 `task test:visual`：54 个状态中 30 个一致、23 个属于例外、1 个不符，是已知的 `set-yrc-amll` 偶发差异（17946 个像素，最大差值 47）。
- [ ] 搜索：歌曲、专辑、艺术家、歌单，以及建议与无结果状态。
- [ ] 歌曲右键菜单，条目见[阶段 2](phase-2-player-baseline.md#歌曲列表)。
- [x] 设置（先后顺序的第 ④ 步）：条目见[阶段 2](phase-2-player-baseline.md#设置)与下文“已确定”中的取舍。已实现并运行确认（2026-10-27）；输出设备的列表与切换（由阶段 4 转入，见[阶段 2 的 0.3](phase-2-player-baseline.md#03-chrome-音频实验)）和均衡器的实际听感，Eric 已在桌面 Chrome 上确认（2026-10-27）。
  - **设置界面**照搬 SPlayer 的 [components/Setting/](../../frontend/src/components/Setting/)（`MainSetting`、`UniversalSetting`、`SettingItemRenderer`、`SettingSearch`、`LyricPreview`、`ShortcutRecorder` 与五个分类的配置），入口为顶栏设置菜单的“全局设置”（菜单按 SPlayer 网页版补上“主题配置”）和全屏播放器的歌词设置按钮（[utils/modal.ts](../../frontend/src/utils/modal.ts) 的 `openSetting`）。子弹窗照搬 `ThemeConfig` 与 [Modal/Setting/](../../frontend/src/components/Modal/Setting/) 中的字体、封面显示、侧边栏、首页栏目、歌单界面、全屏播放器、右键菜单、歌词排除。新增依赖 `sortablejs` 1.15.7 与 `@vueuse/integrations` 14.3.0（首页栏目拖动，与 SPlayer 同版本）。
  - 与 SPlayer 的差异（除“已确定”中的取舍）：去掉移动端分支（最小宽度 990px）；主题配置没有自定义背景；字体用 Chrome 的 Local Font Access 列出（阶段 2 的决定），已授权时列出，未询问过时第一次展开下拉框请求权限，拒绝时与 SPlayer 网页版一样手动输入；“展示播放状态信息”控制的歌名下整行标签已不显示，这一项去掉；各管理弹窗的条目换成 Sonora 中有的项（右键菜单增加“查看专辑”“查看艺术家”，关掉“复制歌曲名称”时“更多操作”一并隐藏）；说明文字中 Sonora 没有的功能（自定义图片、下载、评论、歌单广场等）换成 Sonora 的对应项，“下一首歌曲预载”“搜索关键词建议”“清除全部数据”的说明按 Sonora 的行为修改。
  - **服务端同步**：[core/settings/sync.ts](../../frontend/src/core/settings/sync.ts)，规则见“已确定”；服务端只存 `{setting, shortcuts}`（[api.md](../api.md)）。`DataSource` 增加 `settings`、`saveSettings`（`keepalive`），fixture 在内存中保存。单元测试（[sync.test.ts](../../frontend/src/core/settings/sync.test.ts)，10 个）：上传内容不含设备项、服务端为 `{}` 时上传本地、服务端优先且设备项不变、缺少与类型不符的字段用默认值、1 秒防抖、只改设备项不上传、读取失败后的补传、重新可见时重读与待上传时不读、`keepalive` 补发、上传失败后重传。
  - **快捷键**：[stores/shortcut.ts](../../frontend/src/stores/shortcut.ts) 与 [composables/useShortcut.ts](../../frontend/src/composables/useShortcut.ts)（照搬 SPlayer `useInit.ts` 的键盘事件，含 100 ms 防抖）；控制器补上 `seekBy`。此前 Sonora 没有快捷键，全屏播放器按 Escape 关不掉，现在可以。
  - **同步补偿**：引擎增加 `setAudioDelayCompensation`，`AudioElementPlayer.currentTime` 为元素的播放位置加上补偿（跳转中仍返回目标时间，与 SPlayer 相同），`initPlayer` 跟随设置更新。单元测试 [AudioElementPlayer.test.ts](../../frontend/src/core/audio-player/AudioElementPlayer.test.ts)（3 个）。
  - **下一首歌曲预载**：实验 [preload-next.mjs](../../scripts/experiments/preload-next.mjs)（Chromium 153，8 Mbit/s，曲库中 12–14 MB 的 MP3，结果在 `.local/experiments/preload-next/`）：不预加载时切歌后 2548 ms 才 `loadedmetadata`；用 `fetch` 预取开头 3 MiB 后仍要 2582 ms，主元素重新请求、没有复用（音频是 `private, no-cache`）；用一个隐藏的 `Audio` 元素（同样 `crossOrigin = "anonymous"`）预加载后只要 7 ms，主元素直接复用它下载的数据。隐藏元素的 `preload` 为 `metadata` 时只下载约 2.5 MB（等 30 秒），为 `auto` 时下载约 5.9 MB，切歌都是 7 ms，因此用 `metadata`。实现：引擎的 `preload(url)`（隐藏元素保留到下一次预加载），控制器在当前曲开始播放后按自动切歌的顺序预加载下一首（`nextAutoEntry`：单曲循环、不循环时的最后一首与只有一首时没有），队列、播放模式或设置变化时重新计算；当前曲还没开始时不释放，避免切到预加载的那首之前被释放。单元测试 5 个（[PlayerController.test.ts](../../frontend/src/core/player/PlayerController.test.ts)）。
  - **运行确认**（2026-10-27，开发模式临时实例，真实曲库副本，Playwright Chromium 153 无头，真实点击；截图在 `.local/visual/settings-5/`）：
    - 设置弹窗五个分类与设置搜索正常，与参考实例同视口的截图除上述取舍外一致；九个子弹窗都能打开。左栏底部去掉作者一行后，按 Eric 的要求补上作者行所占的高度（26.39px），“Sonora v0.1.0”与 SPlayer 的版本行位置相同（834.38 与 834.39）。设置中的下拉菜单挂到 body：留在弹窗内时会被 `.main-setting` 的 `overflow: hidden` 裁掉，Eric 在输出设备下拉框上发现（SPlayer 结构相同）。以上差异、说明文字的修改与上传失败时不提示（只记录并在下次修改或页面重新可见时重传），Eric 已确认（2026-10-27）。
    - 同步：首次打开时服务端为 `{}`，上传本地设置（87 个字段、10 个快捷键）；改主题模式后 0.4 秒服务端未变、1.4 秒已更新；改同步补偿后服务端没有这个字段；清掉浏览器中的设置后刷新，从服务端恢复；服务端被别处改动后，页面重新可见时读到新值。
    - 快捷键：P 打开、Escape 关闭全屏播放器，L 切换播放列表，→ 快进约 5 秒，Ctrl+Space 暂停。右键菜单关掉“查看专辑”后不显示；歌单界面关掉“时间”后专辑年份隐藏。授予 `local-fonts` 后字体下拉框列出本机字体；授予麦克风权限（假设备）后输出设备列出三个，切换后 `sinkId` 为所选设备，选回“默认”后为空（默认设备）。
    - 真实播放（1 MB/s 限速）：第一首开始播放后预加载第二首；切到第二首用 22 ms，主元素只从第 2555904 字节继续请求；切到没有预加载的第五首用 4574 ms。同步补偿 500 ms 时引擎的播放位置比元素超前 0.500 秒。
    - “清除全部数据”后服务端设置回到默认值，重载后搜索历史与播放队列为空。页面没有报错。
  - Eric 在桌面 Chrome（Windows，155）上确认（2026-10-27）：输出设备的列表与切换正常。两处修改：
    - 输出设备的下拉菜单被设置弹窗的下边缘挡住，见上文，设置中的下拉菜单改为挂到 body。
    - “其他控制”只剩均衡器一项，悬停出的菜单像提示，点按钮没有反应（Eric 以为均衡器打不开）。改为 `Eq` 图标的按钮，点击直接打开均衡器（Eric 确认），`Controls` 图标删除。底部栏与全屏播放器的控制栏都受影响，截图对照新增遮罩：[player.mjs](../../scripts/visual/player.mjs) 把这个按钮的区域记入 manifest，[diff.mjs](../../scripts/baseline/diff.mjs) 的 `--masks` 不计其中的像素，[compare.mjs](../../scripts/visual/compare.mjs) 把有遮罩像素的状态记为例外并注明原因，其余区域照原规则判定。修改后 `task test:visual`：54 个状态中 12 个一致、41 个属于例外、1 个不符（`set-yrc-amll` 的已知偶发差异）。
    - 均衡器开启并选择预设后，听感有变化（Eric 确认）。
  - 回归：`task lint:frontend`、`task test:frontend`（49 个测试）、`vite build` 通过，产物中没有 fixture。`task test:visual`：54 个状态中 29 个一致、23 个属于例外、2 个不符，都是 AMLL 的偶发差异：`set-yrc-amll`（17946 个像素，与以前相同）与 `full-yrc-amll`（20727 个像素，最大差值 37）。单独重跑 `full-yrc-amll`：改动后 10 次中 6 次不符，改动前的 HEAD 上 10 次中 5 次不符，差异的像素数相同（18541、20727、21733），不是这次改动造成的。
- [ ] 异常状态：空曲库、服务不可达与恢复、歌曲已被管理后台删除、登录会话失效。
- [ ] 播放器相关状态完成截图对照；其他页面经 Eric 看过。

## 已确定（2026-10-25）

- **先后顺序**（Eric 确认）：① 核对已接入的真实数据（工作包 1、2）→ ② 歌单写操作、清空最近播放、“我喜欢”页刷新 → ③ 照搬 SPlayer 的弹窗（添加到歌单、选择歌手、复制更多信息、复制歌词、均衡器；复制更多信息后来去掉）→ ④ 设置界面、服务端同步与输出设备 → ⑤ 搜索建议与无结果、右键菜单 → ⑥ 异常状态 → ⑦ 截图对照，Eric 看页面。
- **设置的同步**（Eric 确认，2026-10-27）：setting store 与快捷键中除设备相关项外全部上传（外观、歌词、播放、常规、快捷键）；输出设备、音频与歌词同步补偿、侧栏展开状态，以及 status store 中的均衡器与音量留在浏览器。最后一次改动后 1 秒整体 `PUT`，关闭页面时用 `keepalive` 补发。打开页面时先用浏览器中的缓存渲染，`GET` 返回后以服务端为准（缺少的字段用默认值）；服务端为 `{}` 时上传浏览器中的当前设置；`GET` 失败时只用本地，本地改动等下次读取成功后再上传。冲突时后写者胜，页面重新可见且没有待上传的改动时重新读取。
- **输出设备**（Eric 确认，2026-10-27）：Chrome 要有麦克风权限才列出设备，首次展开下拉框且设备名为空时请求一次，拿到后立即停用麦克风；拒绝时只有“默认”。
- **设置条目的取舍**（Eric 确认，2026-10-27，补充阶段 2 的设置表）：
  - 增加：全屏播放器元素、歌单界面元素（去掉“显示标签”）、右键菜单管理、侧边栏显示歌单封面、显示播放列表数量、歌词排除配置、搜索框行为。
  - 不做：关于与鸣谢、网络与连接、自定义背景、备份与恢复；阶段 2 中的动态封面（视频来自网易）与音量平衡（曲库没有 ReplayGain 数据）去掉。
  - “下一首歌曲预载”改为预加载下一首的音频（SPlayer 预取的是网易播放地址）：当前曲开始播放后，请求下一首的开头约 3 MB；下一首按队列实际顺序，单曲循环时不预加载。先实验确认 Chrome 会复用缓存。
  - 快捷键：SPlayer 网页版在页面内响应全部快捷键，设置里只能改“页面内快捷键”三项。Sonora 去掉桌面歌词，其余 10 个都在页面内生效并可修改，不显示“开启全局快捷键”。
  - “清除全部数据”：重置设置（上传默认值），清空浏览器中的播放队列、搜索历史、均衡器与音量后重载；服务端的歌单、喜爱与历史不动，不退出登录（SPlayer 会退出登录）。
  - 设置弹窗左栏底部只显示“Sonora”与版本号（`frontend/package.json`），去掉 GitHub 作者一行。
- **音频与歌词同步补偿只保留手动补偿**（Eric 确认，与 SPlayer 不同）：设置项照搬（±1000 ms，步长 10），`latencyHint` 仍固定为 `interactive`，补偿始终生效、设置项始终显示。SPlayer 只在“Web Audio 延迟策略”为 `playback` 时显示并应用这项（同时自动减去 `outputLatency + baseLatency`），Sonora 不保留延迟策略，也不做自动补偿。

## 待决事项

- 阶段 2 没覆盖、阶段 3.4 暂未照搬的 SPlayer 功能，Eric 决定以后再议（2026-10-25）：搜索框右键菜单（复制、粘贴）、专辑与歌单卡片的右键菜单（`CoverMenu.vue`）、详情页“更多”中的批量操作。

## 退出条件

- 首发页面清单中的每个页面都接入真实数据，功能与视觉证据齐全。
- 管理后台修改曲库后，歌单、喜爱和历史仍保持正确。
