# SPlayer 参考资源

前端初始化使用了 [SPlayer](https://github.com/SPlayer-Dev/SPlayer) 的部分视觉实现，原作者为 imsyy 及 SPlayer 贡献者。参考提交：`28b76bad9eb800e0c90dac8475d4769acfe2450b`。

保留其 AGPL-3.0 许可证全文于 [LICENSE](LICENSE)。复制或改编内容包括：

- `frontend/src/style/`：原全局样式、动画及布局/导航/菜单/搜索/用户组件样式。
- `frontend/src/components/Layout/Logo.vue`、`frontend/src/components/Global/SvgIcon.vue`：原标志与图标展示实现。
- `frontend/src/assets/icons/`、`frontend/public/fonts/`：参考项目对应 SVG 图标与 Logo 字体。
- 主题生成、主布局、导航和首页展示结构按原项目对应实现改编，业务逻辑重新实现。

2026-10-23 确定：标志图形暂时复用；界面中的 SPlayer 名称与文字均替换为 Sonora。Sonora 不迁入网易来源、认证或业务接口。各依赖的许可证仍以各包声明为准。
