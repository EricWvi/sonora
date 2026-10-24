# SPlayer 参考资源

前端初始化使用了 [SPlayer](https://github.com/SPlayer-Dev/SPlayer) 的部分视觉实现，原作者为 imsyy 及 SPlayer 贡献者。参考提交：`28b76bad9eb800e0c90dac8475d4769acfe2450b`。

保留其 AGPL-3.0 许可证全文于 [LICENSE](LICENSE)。复制或改编内容包括：

- `frontend/src/style/`：原全局样式、动画及布局/导航/菜单/搜索/用户组件样式。
- `frontend/src/components/Layout/Logo.vue`、`frontend/src/components/Global/SvgIcon.vue`：原标志与图标展示实现。
- `frontend/src/assets/icons/`、`frontend/public/fonts/`：参考项目对应 SVG 图标与 Logo 字体。
- 主题生成、主布局、导航和首页展示结构按原项目对应实现改编，业务逻辑重新实现。

2026-10-23 确定：标志图形暂时复用；界面中的 SPlayer 名称与文字均替换为 Sonora。Sonora 不迁入网易来源、认证或业务接口。各依赖的许可证仍以各包声明为准。

# ThumbHash

`backend/internal/thumbhash/` 的编码器移植自 Evan Wallace 的 [ThumbHash](https://github.com/evanw/thumbhash) 参考实现，MIT 许可证：

> Copyright (c) 2023 Evan Wallace
>
> Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
>
> The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
