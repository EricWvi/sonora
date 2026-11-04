# 阶段 7：网易云音乐接入与卡片菜单

状态：已立项（2026-11-04，Eric 提出）；卡片与批量操作已实现并运行确认（2026-11-04），网易云部分待确认。前置：阶段 6 退出（2026-11-03）。

本阶段调整“与网易云接口相关的页面和入口全部不做”的原则：只接入下面列出的部分，其他网易云功能仍不做。

## 工作包

### 卡片与批量操作

这部分不依赖网易云，可以先做。

- [x] 卡片右键菜单：专辑、歌单卡片按 SPlayer 的 `CoverMenu.vue`（查看详情、播放或暂停、复制名称与 ID；分享不做），范围包括首页“最近添加的专辑”、专辑页、艺术家的专辑和搜索结果。歌单卡片目前只出现在搜索结果的“歌单”标签。实现：[CoverMenu.vue](../../frontend/src/components/Menu/CoverMenu.vue)，在 [CoverList.vue](../../frontend/src/components/List/CoverList.vue) 中接入，所有用 CoverList 的页面一并生效。`Eye` 图标从 SPlayer 复制（Google Material Icons，已在[视觉资源清单](../visual-assets.md#4-图标)中）。
- [x] 侧边栏歌单项的右键菜单（Eric 确认，2026-11-04）：SPlayer 的侧边栏没有右键菜单，是 Sonora 新增。条目与卡片一致再加管理：查看详情、播放或暂停、编辑歌单、删除歌单、复制歌单名称与 ID；编辑与删除复用歌单详情页的弹窗与确认，删除正在查看的歌单时回到首页。实现见 [SidebarMenu.vue](../../frontend/src/components/Layout/SidebarMenu.vue)：Naive UI 菜单项的 `props` 只作用于分隔线，右键事件经 `n-menu` 的 `node-props` 绑定。
- [x] 首页“最近添加的歌曲”卡片的右键菜单：用歌曲列表现有的 `SongListMenu`（立即播放、下一首播放、添加到歌单、更多操作、查看专辑、查看艺术家、同名搜索），菜单设置 `contextMenuOptions` 同样生效。
- [x] 详情页批量操作（Eric 确认，2026-11-04）：专辑、歌单、我喜欢的音乐三个详情页的“更多 → 批量操作”，弹窗照搬 SPlayer 的 `BatchList.vue`（表格勾选、“高级筛选”按序号范围选择）。与 SPlayer 不同：没有批量下载与删除本地歌曲；所有页面有“添加到歌单”与“下一首播放”（Sonora 新增，按勾选顺序插到当前歌曲之后，[PlayerController.addNextSongs](../../frontend/src/core/player/PlayerController.ts)）；歌单页另有“删除选中的歌曲”；我喜欢的音乐另有“取消喜欢”（Sonora 新增，逐首请求，中途失败时只移除已成功的）。移除后弹窗表格与页面列表、头部数量同步更新。艺术家页不做（歌曲是分页载入）。实现：[BatchList.vue](../../frontend/src/components/Modal/BatchList.vue)、[modal.ts](../../frontend/src/utils/modal.ts) 的 `openBatchList`。

运行确认（2026-11-04，开发模式临时实例，本机 337 首的曲库副本，Playwright Chromium 无头，1440×1000；脚本在 `/tmp`，未入库）：

- 首页专辑卡片右键：查看详情、播放、复制专辑名称、复制专辑 ID；从菜单播放后再右键变为“暂停”。
- 首页歌曲卡片右键：歌曲信息头部与歌曲列表的菜单条目一致。
- 侧边栏歌单项右键：六个条目齐全；查看详情进入歌单页；删除歌单后回到首页，侧边栏不再显示。
- 歌单页批量：按钮为删除选中的歌曲、下一首播放、添加到歌单；勾选 2 首下一首播放提示“已添加 2 首至下一首播放”；删除后弹窗剩 3 行、勾选清零，页面头部数量为 3，服务端歌单为 3 首。
- 专辑页批量：按钮为下一首播放、添加到歌单；高级筛选 1–2 选中 2 首。
- 我喜欢的音乐批量：按钮为取消喜欢、下一首播放、添加到歌单；取消 1 首后服务端由 3 首变为 2 首，弹窗与页面头部同步。
- 全程没有页面错误。`task lint:frontend`、`task test:frontend`（57 个测试，新增批量下一首播放 1 个）、`pnpm build` 通过，产物不含 fixture。

### 网易云音乐

- [ ] 侧边栏第二块改为：
  - 开头加小标题“网易云音乐”；
  - 我喜欢的音乐；
  - 我的收藏（歌单、专辑、歌手）；
  - 最近播放：从 Sonora 的最近播放改成网易云的最近播放。
- [ ] 网易云登录与会话，方式见待决事项。
- [ ] “我喜欢的音乐”“我的收藏”“最近播放”三个页面接入真实数据，含未登录、登录失效和接口失败时的状态。

## 依据

源码确认（`reference/SPlayer`，2026-11-04）：

- 接口包：`@neteasecloudmusicapienhanced/api`，在 SPlayer 的本机服务中调用。
- 登录：`/login/qr/*`、`/login/cellphone`、`/login/status`、`/login/refresh`，见 [login.ts](../../reference/SPlayer/src/api/login.ts)。
- 喜欢与收藏：`/likelist`、`/user/playlist`、`/album/sublist`、`/artist/sublist`，见 [user.ts](../../reference/SPlayer/src/api/user.ts)。
- 最近播放：SPlayer 的 `src/api` 中没有网易云最近播放的请求，页面和接口都需要另行确认。
- 批量操作：`BatchList.vue` 的操作有批量下载（开发者模式）、删除选中的歌曲（歌单）、添加到歌单与删除本地歌曲。

## 待决事项

- 网易云的歌怎么播放：只浏览、能播放，还是只播放曲库里匹配得上的歌。这一项决定播放地址、音质与版权的处理范围，也决定 Sonora 的数据模型要改多少。
- 登录方式（二维码、手机号或 Cookie），以及会话存在服务端还是浏览器；多用户时每人绑定各自的网易云账号，还是全站共用一个。
- 接口调用放在哪里：Go 服务调用接口包需要 Node 运行环境，或者单独部署一个服务，或者在 Go 中实现需要的接口。这一项影响镜像和[部署说明](../deployment.md)。
- 侧边栏“我喜欢的音乐”改成网易云的之后，Sonora 现有的喜爱数据、快捷入口、播放栏的喜欢按钮如何处理。
- Sonora 的最近播放不再出现在侧边栏后，首页“最近播放”栏目、最近播放页和历史记录接口如何处理。
- “我的收藏”是一个页面分三个标签，还是三个入口。

## 退出标准

待上述事项确定后补充。
