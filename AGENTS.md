# SPlayer 参考实例

使用无头浏览器访问 SPlayer 参考实例时，每次都必须创建独立 BrowserContext，并在打开页面前调用 `scripts/splayer-session.mjs` 的 `prepareSPlayerSession(context, baseURL)`。该函数从本机 `.local/credentials/splayer/only_session` 加载 Cookie；默认实例为 `http://127.0.0.1:25884`。使用方式见 [SPlayer 浏览器会话](docs/splayer-session.md)。

凭据缺失或登录失效时，明确报告状态，请 Eric 更新凭据。凭据仅用于 SPlayer 参考实例，不得用于 Sonora、自研服务或其他站点，也不得输出、提交或写入测试产物。
