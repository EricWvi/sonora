# 部署 Sonora

更新日期：2026-10-24。容器镜像与发布流程属于[阶段 6](phase/phase-6-first-release.md)，尚未提供 Dockerfile；本文记录阶段 1 与阶段 4 已确定的运行方式，镜像就绪后补充 Compose 示例。不在开发机上构建镜像。

## 构建

`task build`（即 `scripts/build.sh`）构建播放端与管理后台，把两份产物复制到 `backend/web/` 并嵌入二进制，输出 `release/sonora`。二进制不依赖 CGO，可直接放入 distroless 等最小镜像。

## 配置

参考 [config.example.yaml](../backend/config.example.yaml)，通过 `-config` 指定文件路径：

| 键 | 说明 |
| --- | --- |
| `listen` | 监听地址，默认 `:14560` |
| `data_dir` | SQLite 数据库与封面，挂载为数据卷 |
| `media_dir` | 音频存储卷，只由 Sonora 写入；手动放入的文件不会入库 |
| `public_url` | 浏览器访问的 origin，例如 `https://music.example.com`；写请求必须来自该 origin，OIDC 回调地址为 `<public_url>/auth/callback` |
| `oidc.issuer`、`oidc.client_id`、`oidc.client_secret` | 播放端登录用的 Authelia；三项必填（开发模式除外），secret 建议用环境变量 `SONORA_OIDC_CLIENT_SECRET` 提供，它会覆盖配置文件中的值 |

两个目录必须不同，并分别挂载持久卷；备份时两者都要包含。

`dev` 只被 `go run -tags dev` 构建接受，发布的二进制遇到 `dev: true` 会拒绝启动。

## Authelia 客户端

播放端认证照搬 `../homepage`。在 Authelia 中注册一个机密客户端：`redirect_uris` 为 `<public_url>/auth/callback`；`require_pkce: true`、`pkce_challenge_method: S256`；scope 为 `openid profile email`；`token_endpoint_auth_method: client_secret_basic`。可参照契约测试的 [configuration.yml](../backend/internal/auth/testdata/authelia/configuration.yml)。Sonora 首次有人登录时才去发现 issuer，Authelia 不可用时服务照常启动，只是暂时无法登录。

## 首次初始化

全新数据卷启动时，服务发现还没有管理员，会在标准输出（即容器日志）中打印一次性初始化 token：

```text
==================== Sonora 初始化 ====================
尚未创建管理员。在浏览器打开 https://music.example.com/admin/
输入下面的初始化 token、用户名和密码创建管理员：

    <token>
...
```

用容器平台的日志查看 token，例如：

```bash
docker logs sonora 2>&1 | grep -A5 'Sonora 初始化'
# 或 docker compose logs sonora
```

然后打开 `<public_url>/admin/`，输入 token、用户名和密码。规则：

- token 只保存在进程内存中，服务重启后重新生成，旧 token 失效。
- 初始化只能成功一次；成功后 token 立即作废，之后启动不再打印。
- token 错误时拒绝，并限制每分钟 5 次失败尝试。
- token 是 Sonora 唯一写入日志的凭据。密码以 bcrypt 保存，不写入日志。

## 访问路径

同一端口提供：播放端 `/`、播放端登录 `/auth/`、播放端 API `/api/v1/`、管理后台 `/admin/`、管理 API `/admin/api/v1/`。管理后台不走 OIDC；需要限制访问来源时在反向代理上处理。
