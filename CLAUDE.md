# portal — 公共 API 及账户认证服务

希波万象平台的 Go 后端服务，提供公共内容查询接口和账户认证体系（注册、登录、会话管理、WebAuthn）。

## 技术栈

- **语言**: Go 1.24
- **HTTP 框架**: Gin v1.11
- **数据库**: PostgreSQL + sqlx（原生 SQL，无 ORM）
- **认证**: JWT RS256 + WebAuthn (FIDO2)
- **缓存/消息队列**: Redis v9
- **机器人防护**: Cloudflare Turnstile
- **运行端口**: 8001

## 常用命令

```bash
go run . --config file://config/host.yml          # 启动 portal 服务
go run . --svcrole syncer --config file://...      # 启动同步进程
go build -o ./portal .                             # 编译
docker build --progress=plain -t portal .          # 构建 Docker 镜像
```

## 目录结构

```
portal/
├── main.go             # 入口：解析 --svcrole 参数，分发到 portal 或 syncer
├── server.go           # HTTP 服务器：路由注册、CORS 配置、中间件
├── business/           # 业务逻辑层
│   ├── account/        # 账户模块（signup/signin/signout/session/userinfo/auth）
│   ├── articles/       # 文章查询
│   ├── channels/       # 频道查询
│   ├── comments/       # 评论（含 Redis 消息队列）
│   ├── images/         # 图片查询
│   ├── viewers/        # 浏览记录统计
│   ├── cloudflare/     # Turnstile 验证
│   └── userinfo.go     # 认证工具（从 Cookie/Header 提取用户身份）
├── cloud/
│   └── files/          # 云端文件管理（完整 CRUD）
├── models/             # 数据模型 + 数据库操作（Account、Session、Application 等）
├── services/           # 工具服务（JSON、Base58、文件系统、Git）
├── handlers/           # 特殊处理器（WebAuthn 硬件密钥认证）
├── host/               # 本地存储（notebook、album）
├── syncer/             # 后台同步进程（文章资产同步）
├── config/             # 配置示例（host.yml）
└── docs/               # API 文档（api.md）
```

## API 路由（前缀 `/portal`）

### 认证接口
```
POST /portal/account/signup              注册（可选 Turnstile 验证）
POST /portal/account/signin              登录（可选 Turnstile 验证）
POST /portal/account/signout             登出
GET  /portal/account/userinfo            获取当前用户信息（需登录）
GET  /portal/account/session             查询会话
GET  /portal/account/auth/app            查询应用信息
POST /portal/account/auth/permit         应用授权
GET  /portal/account/auth/userinfo       用户授权信息（被 stargate 调用）
```

### 公开内容
```
GET /portal/articles                     文章列表（支持分页/搜索/频道筛选）
GET /portal/articles/:uid                文章详情
GET /portal/channels                     频道列表
GET /portal/images                       图片列表
GET /portal/comments                     评论列表
POST /portal/comments                    发布评论（需登录）
```

### 云端文件（需登录）
```
GET  /portal/cloud/files                 文件列表
POST /portal/cloud/files/:uid            创建/更新文件
GET  /portal/cloud/files/path            文件路径层级树
```

## 认证机制

### 登录流程
1. 验证 Turnstile（取决于 `SERVE_MODE`）
2. 查询账户、校验密码（bcrypt/argon2）
3. 创建 `SessionModel`，签发 JWT Token（RS256，使用 `JWT_PRIVATE_KEY`）
4. 通过 HttpOnly Cookie（字段名 `PT`）和响应体同时返回 Token

### 鉴权方式（二选一）
- **Cookie**: `Cookie: PT=<jwt_token>`（前端浏览器默认方式）
- **Header**: `Portal-Authorization: Bearer <jwt_token>`

### 鉴权函数
- `FindAccountFromCookie()` — 从请求 Cookie 提取用户
- `FindAccountFromToken()` — 从 Authorization Header 提取用户
- 公开接口无需鉴权，返回匿名用户

## 数据库访问

**使用原生 SQL + sqlx，不使用 ORM。**

```go
// 查询示例
sqlText := `select * from accounts where username = :username limit 1;`
rows, err := datastore.NamedQuery(sqlText, map[string]interface{}{"username": username})
var results []*AccountModel
sqlx.StructScan(rows, &results)

// 写入示例
datastore.NamedExec(sqlText, sqlParams)
```

**主要数据表**（均在 `portal` 数据库）：
- `public.accounts` — 用户账户
- `public.sessions` — 会话
- `community.articles` — 文章
- `community.files` — 文件
- `community.channels` — 频道
- `community.comments` — 评论
- `community.images` — 图片

## 配置参数

| 参数 | 说明 |
|------|------|
| `DATABASE` | PostgreSQL 连接串（host= user= password= dbname= port=） |
| `INTERNAL_PORTAL_URL` | 内部访问 URL（供 stargate 调用） |
| `PUBLIC_PORTAL_URL` | 公开访问 URL |
| `JWT_PRIVATE_KEY` | RSA 私钥（PEM 格式，用于签发 JWT） |
| `JWT_PUBLIC_KEY` | RSA 公钥（PEM 格式，用于验证 JWT） |
| `SERVE_MODE` | `SELFHOST`（跳过 Turnstile）/ 默认 WAN 模式 |
| `STORAGE_URL` | 文件存储路径（`file:///...`） |
| `REDIS_URL` | Redis 连接 URL（可选） |
| `RPID` / `RPOrigins` | WebAuthn 配置 |
| `CLOUDFLARE_TURNSTILE_SECRET` | Cloudflare Turnstile 密钥 |

环境变量：`PORT`（默认 8001）、`CONFIG`（配置文件路径）

## 运行模式

```bash
--svcrole portal   # 默认：HTTP API 服务器
--svcrole syncer   # 后台同步进程（文章资产同步）
```

## CORS 配置

- 允许来源：`*.huable.xyz` 域名（debug 模式下放开）
- 允许方法：PUT、PATCH、POST、GET
- 允许 Header：`Origin`、`Content-Type`、`Portal-Authorization`
- 允许 Credentials：`true`

## 重要约定

- 所有数据库查询使用命名参数（`:param`），防止 SQL 注入
- `SERVE_MODE=SELFHOST` 时跳过 Cloudflare Turnstile 验证（本地开发用）
- Docker 运行时以非 root 用户 `portal:golang` 运行
