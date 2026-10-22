# SPlayer 参考实例的浏览器会话

本机凭据位置为 `.local/credentials/splayer/only_session`，已由 Git 忽略。目录权限为 `700`，文件权限为 `600`；需要更新登录状态时替换该文件内容，保持权限不变。该凭据仅用于查看 SPlayer 参考界面。

每次无头访问都使用新的独立 BrowserContext，并在导航前调用 [prepareSPlayerSession](../scripts/splayer-session.mjs)。函数读取最新凭据，将 Cookie 加入会话，并按 SPlayer 的读取方式准备当前实例 origin 下的 localStorage。不要与访问 Sonora 或其他站点的会话共用 BrowserContext。

当前可以使用 `../palace` 已安装的 Playwright。从该目录运行：

```bash
node --input-type=module <<'JS'
import { chromium } from '@playwright/test';
import { prepareSPlayerSession } from '../sonora/scripts/splayer-session.mjs';

const browser = await chromium.launch({ headless: true });
try {
  const context = await browser.newContext();
  await prepareSPlayerSession(context);
  const page = await context.newPage();
  await page.goto('http://127.0.0.1:25884', { waitUntil: 'networkidle' });
  // 在此检查登录状态或执行界面对照。
} finally {
  await browser.close();
}
JS
```

如果实例端口改变，向函数传入新的本机 `baseURL`，并使用相同地址导航。加载完成后检查页面或登录状态接口；凭据缺失、失效或服务不可达时，报告实际状态，请 Eric 更新凭据或实例地址。

浏览器脚本不输出凭据，不保存包含凭据的 storageState、HAR、trace 或请求日志。截图用于界面对照时注意账号个人信息；凭据值不进入仓库或测试产物。
