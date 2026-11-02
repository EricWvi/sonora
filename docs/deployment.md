# 部署 Sonora

更新日期：2026-11-02。镜像由 CI 构建（[阶段 6](phase/phase-6-first-release.md)），不在开发机上构建镜像。

## 构建

`task build`（即 `scripts/build.sh`）构建播放端与管理后台，把两份产物复制到 `backend/web/` 并嵌入二进制，输出 `release/sonora`。二进制不依赖 CGO，可直接放入 distroless 等最小镜像。

## 镜像

推送 `v*` tag（例如 `v1.0.0`）后，GitHub Actions（[release.yml](../.github/workflows/release.yml)）按 [Dockerfile](../Dockerfile) 构建并推送 `ghcr.io/ericwvi/sonora:<版本>`；正式版同时更新 `latest`，预发布版（如 `1.1.0-rc.1`）不更新。也可以在 Actions 页面手动运行并填写版本号。只构建 `linux/amd64`。

构建步骤与 `scripts/build.sh` 相同，另加两项检查：两份前端的 ESLint，以及产物中不含开发 fixture。运行层是 `gcr.io/distroless/static-debian12:nonroot`（用户 65532）：

| 路径 | 用途 |
| --- | --- |
| `/app/config.yaml` | 配置文件，只读挂载 |
| `/data` | 数据卷：`sonora.db` 与 `covers/` |
| `/media` | 存储卷：上传的音频 |
| 端口 `14560` | HTTP，放在反向代理之后 |

两个卷在宿主机上须属于 `65532:65532`，否则服务无法写入。

## Compose 示例

```yaml
services:
  sonora:
    image: ghcr.io/ericwvi/sonora:1.0.0
    restart: unless-stopped
    volumes:
      - ./sonora/data:/data
      - ./sonora/media:/media
      - ./sonora/config.yaml:/app/config.yaml:ro
    environment:
      - TZ=Asia/Shanghai
      - SONORA_OIDC_CLIENT_SECRET=${SONORA_OIDC_CLIENT_SECRET}
```

`config.yaml`：

```yaml
listen: ":14560"
data_dir: "/data"
media_dir: "/media"
public_url: "https://music.example.com"
oidc:
  issuer: "https://auth.example.com"
  client_id: "sonora"
```

反向代理把整个域名转发到 `sonora:14560`，不要加 forward_auth：播放端自己走 OIDC，管理后台有自己的登录（密码 bcrypt，失败全局每分钟至多 10 次）。

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

播放端认证照搬 `../homepage`。在 Authelia 中注册一个机密客户端：`redirect_uris` 为 `<public_url>/auth/callback`；`require_pkce: true`、`pkce_challenge_method: S256`；scope 为 `openid profile email`；`token_endpoint_auth_method: client_secret_basic`。可参照契约测试的 [configuration.yml](../backend/internal/auth/testdata/authelia/configuration.yml)，在 `identity_providers.oidc.clients` 下加入：

```yaml
- client_id: sonora
  client_name: Sonora
  # authelia crypto hash generate pbkdf2 --variant sha512 生成；明文给 SONORA_OIDC_CLIENT_SECRET
  client_secret: '$pbkdf2-sha512$...'
  public: false
  authorization_policy: two_factor  # 按自己的策略
  require_pkce: true
  pkce_challenge_method: S256
  redirect_uris:
    - https://music.example.com/auth/callback
  scopes: [openid, profile, email]
  grant_types: [authorization_code]
  response_types: [code]
  token_endpoint_auth_method: client_secret_basic
```

Sonora 首次有人登录时才去发现 issuer，Authelia 不可用时服务照常启动，只是暂时无法登录。

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

## 升级与恢复

数据库的表结构随版本升级：启动时按顺序执行尚未执行的迁移，每个迁移在单独的事务中完成，失败时回滚并拒绝启动，数据库停留在上一个完整的版本。Sonora 不自动备份，升级前手动备份：

1. 停止容器（`docker compose stop sonora`），确保数据库没有写入。
2. 复制数据卷与存储卷，例如 `cp -a /path/to/data /path/to/data.bak-<版本>`。数据卷中的 `sonora.db`、`sonora.db-wal` 与 `covers/` 要一起复制；数据库记录的音频文件在存储卷中，两者要来自同一时刻。
3. 换成新镜像启动，查看日志确认启动成功。

恢复：停止容器，用备份替换数据卷（含 `-wal`、`-shm` 文件一并替换或删除），换回旧版镜像启动。数据库的版本比当前程序支持的更新时（例如升级后直接换回旧镜像），服务拒绝启动并在日志中说明，不会在新表结构上运行旧程序。

## 访问路径

同一端口提供：播放端 `/`、播放端登录 `/auth/`、播放端 API `/api/v1/`、管理后台 `/admin/`、管理 API `/admin/api/v1/`。管理后台不走 OIDC；需要限制访问来源时在反向代理上处理。
