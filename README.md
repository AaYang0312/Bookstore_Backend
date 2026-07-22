# 博学书城后端

博学书城后端是一个使用 Go 和 Gin 开发的 RESTful API 服务，为书城用户端和管理后台提供统一接口。目前已包含用户认证、图书目录、分类、轮播图、收藏、订单、验证码，以及图书、分类、订单、用户和轮播图的管理功能。

## 技术栈

- Go 1.25
- Gin 1.12
- GORM 1.31
- MySQL 8.x
- Redis 6.x 或更高版本
- JWT

## 功能模块

- 用户：注册、登录、查看及修改个人资料、修改密码
- 图书：首页列表、热销图书、新书、搜索、分类查询和详情
- 首页：分类导航和轮播图
- 收藏：添加、取消、列表、数量和收藏状态
- 订单：创建订单、订单列表、订单详情和支付
- 验证码：生成图片验证码，并使用 Redis 保存有效期为 3 分钟的答案
- 管理后台：数据看板、图书管理、分类管理、订单管理、用户权限和轮播图管理

## 项目结构

```text
Bookstore_Backend/
├── cache/                  # 图书缓存
├── cmd/bookstore-manager/  # 程序入口
├── conf/config.yaml        # 服务、MySQL 和 Redis 配置
├── config/                 # 配置加载
├── global/                 # MySQL、Redis 客户端初始化
├── jwt/                    # JWT 签发与校验
├── model/                  # 数据模型
├── repository/             # 数据访问层
├── service/                # 业务逻辑层
├── sql/                    # 建表脚本和示例数据
├── web/controller/         # HTTP 控制器
├── web/middleware/         # Gin 中间件
├── web/router/             # 路由注册
├── Makefile
└── go.mod
```

## 环境要求

开始前请安装并启动：

- Go 1.25 或与 `go.mod` 兼容的版本
- MySQL
- Redis

如果只需要快速体验完整系统，也可以只安装 Docker Desktop，直接使用后文的 Docker Compose 方案。

## 本地运行

### 1. 初始化数据库

先执行建库建表脚本，再按需导入示例数据：

```bash
mysql -u root -p < sql/bookstore.sql
mysql -u root -p bookstore < sql/mock.sql
```

### 2. 修改配置

编辑 `conf/config.yaml`，填写本机的服务端口、MySQL 和 Redis 连接信息：

```yaml
server:
  port: 8080

database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: your_password
  name: bookstore

redis:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0
```

请勿将生产环境密码提交到版本库。

### 3. 下载依赖并启动

以下命令需要在后端根目录执行：

```bash
go mod download
go run ./cmd/bookstore-manager
```

服务默认监听 `http://localhost:8080`。程序启动时会连接 MySQL 和 Redis；按 `Ctrl+C` 后会优雅关闭服务并释放连接。

也可以使用 Makefile 构建并运行：

```bash
make bookstore-manager
make run-bookstore-manager
```

修改 Go 源码后需要重新构建并重启，正在运行的 `bin/bookstore-manager` 不会自动加载代码变化：

```bash
# 先在旧进程所在终端按 Ctrl+C
make bookstore-manager
make run-bookstore-manager
```

## Docker 快速部署

### 构建后端镜像

在本目录执行：

```bash
docker build -t bookstore-backend .
```

镜像使用多阶段构建，运行阶段以非 root 用户启动，并内置接口健康检查。单独运行镜像时需要提供可访问的 MySQL 和 Redis；容器里的 `127.0.0.1` 指向容器自身，不能用来访问宿主机服务。

后端支持用环境变量覆盖 `conf/config.yaml`，Docker 部署常用配置如下：

| 环境变量 | 说明 | Compose 中的值 |
| --- | --- | --- |
| `BOOKSTORE_SERVER_HOST` | 服务监听地址 | `0.0.0.0` |
| `BOOKSTORE_SERVER_PORT` | 服务端口 | `8080` |
| `BOOKSTORE_DATABASE_HOST` | MySQL 主机 | `mysql` |
| `BOOKSTORE_DATABASE_PORT` | MySQL 端口 | `3306` |
| `BOOKSTORE_DATABASE_USER` | MySQL 用户 | `root` |
| `BOOKSTORE_DATABASE_PASSWORD` | MySQL 密码 | 从部署环境文件读取 |
| `BOOKSTORE_DATABASE_NAME` | 数据库名 | `bookstore` |
| `BOOKSTORE_REDIS_HOST` | Redis 主机 | `redis` |
| `BOOKSTORE_REDIS_PORT` | Redis 端口 | `6379` |
| `BOOKSTORE_REDIS_PASSWORD` | Redis 密码 | 默认为空 |
| `BOOKSTORE_REDIS_DB` | Redis DB | `0` |

### 一键启动完整书城

完整编排文件位于前端仓库。它会同时构建前端、Go 后端和 Agent，并启动 MySQL、Redis：

```bash
cd ../bookstore-fronted-master
cp .env.docker.example .env.docker
# 编辑 .env.docker，填写真实模型密钥
docker compose --env-file .env.docker up --build -d
docker compose ps
```

页面地址为 `http://localhost:3000`，Go API 为 `http://localhost:8080/api/v1`。数据库建表脚本和示例数据只会在 MySQL 数据卷首次创建时自动导入。

## 开发管理员账户

当前本机开发数据库已经配置以下管理员账户：

```text
用户名：admin
密码：12345678
管理后台：http://localhost:3000/admin
```

该账户属于本机数据库状态，`sql/bookstore.sql` 和 `sql/mock.sql` 不会自动创建它。初始化一套全新数据库后，可以先通过前端注册用户，再在 MySQL 中授予管理员权限：

```sql
UPDATE users SET is_admin = TRUE WHERE username = 'admin';
```

管理权限由 `users.is_admin` 控制。登录成功后，返回的 `user_info.is_admin` 会供前端显示管理入口；后端仍会在每次管理请求中通过 JWT 和数据库角色进行二次校验。

> `admin / 12345678` 仅用于本地开发。部署到共享或生产环境前必须修改密码。当前项目密码实现采用 Base64 编码，不属于安全的密码哈希方案，生产使用前应改为 bcrypt 或 Argon2。

## API 概览

所有接口均以 `/api/v1` 开头。

| 模块 | 方法 | 路径 | 是否需要登录 |
| --- | --- | --- | --- |
| 用户 | POST | `/user/register` | 否 |
| 用户 | POST | `/user/login` | 否 |
| 用户 | GET | `/user/profile` | 是 |
| 用户 | PUT | `/user/profile` | 是 |
| 用户 | PUT | `/user/password` | 是 |
| 图书 | GET | `/book/hot` | 否 |
| 图书 | GET | `/book/new` | 否 |
| 图书 | GET | `/book/list` | 否 |
| 图书 | GET | `/book/search` | 否 |
| 图书 | GET | `/book/detail/:id` | 否 |
| 图书 | GET | `/book/category/:category` | 否 |
| 分类 | GET | `/category/list` | 否 |
| 轮播图 | GET | `/carousel/list` | 否 |
| 收藏 | POST | `/favorite/:id` | 是 |
| 收藏 | DELETE | `/favorite/:id` | 是 |
| 收藏 | GET | `/favorite/list` | 是 |
| 收藏 | GET | `/favorite/count` | 是 |
| 收藏 | GET | `/favorite/:id/check` | 是 |
| 订单 | POST | `/order/create` | 是 |
| 订单 | GET | `/order/list` | 是 |
| 订单 | GET | `/order/:id` | 是 |
| 订单 | POST | `/order/:id/pay` | 是 |
| 验证码 | GET | `/captcha/generate` | 否 |

需要登录的接口应携带登录后获得的访问令牌：

```http
Authorization: Bearer <access_token>
```

### 管理后台 API

以下接口除访问令牌外，还要求当前用户的 `is_admin` 为 `true`。普通登录用户访问时返回 `403`。

| 模块 | 方法 | 路径 | 功能 |
| --- | --- | --- | --- |
| 看板 | GET | `/admin/dashboard` | 获取统计数据、近期订单和畅销图书 |
| 图书 | GET | `/admin/books` | 查询图书，可使用 `keyword`、`status`、`page`、`page_size` |
| 图书 | POST | `/admin/books` | 新增图书 |
| 图书 | PUT | `/admin/books/:id` | 编辑图书 |
| 图书 | PATCH | `/admin/books/:id/status` | 上架或下架图书 |
| 图书 | PATCH | `/admin/books/:id/stock` | 修改库存 |
| 分类 | GET | `/admin/categories` | 获取全部分类 |
| 分类 | POST | `/admin/categories` | 新增分类 |
| 分类 | PUT | `/admin/categories/:id` | 编辑分类 |
| 分类 | PATCH | `/admin/categories/:id/status` | 启用或停用分类 |
| 订单 | GET | `/admin/orders` | 查询订单，可使用 `keyword`、`status`、`page`、`page_size` |
| 订单 | GET | `/admin/orders/:id` | 获取订单详情 |
| 订单 | PATCH | `/admin/orders/:id/status` | 修改订单状态 |
| 用户 | GET | `/admin/users` | 查询用户，可使用 `keyword`、`page`、`page_size` |
| 用户 | PATCH | `/admin/users/:id/role` | 授予或取消管理员权限 |
| 轮播图 | GET | `/admin/carousel` | 获取全部轮播图 |
| 轮播图 | POST | `/admin/carousel` | 新增轮播图 |
| 轮播图 | PUT | `/admin/carousel/:id` | 编辑轮播图 |
| 轮播图 | DELETE | `/admin/carousel/:id` | 删除轮播图 |
| 轮播图 | PATCH | `/admin/carousel/:id/status` | 启用或停用轮播图 |

状态修改接口使用 `PATCH`。例如下架图书：

```http
PATCH /api/v1/admin/books/1/status
Authorization: Bearer <access_token>
Content-Type: application/json

{"status": 0}
```

## 与前端联调

1. 启动 MySQL 和 Redis。
2. 启动本服务，确认地址为 `http://localhost:8080`。
3. 再启动 `../bookstore-fronted-master` 中的 React 开发服务器。

管理前端默认使用 `http://localhost:8080/api/v1`。如需修改接口地址，在前端环境变量中设置：

```bash
REACT_APP_API_BASE_URL=http://localhost:8080/api/v1
```

管理前端默认请求真实后端接口。只有显式设置 `REACT_APP_ADMIN_DEMO_MODE=true` 时才会使用演示数据回退。

后端已配置 CORS，可接受前端开发服务器发起的 `GET`、`POST`、`PUT`、`PATCH`、`DELETE` 和 `OPTIONS` 请求。若修改后端代码后浏览器仍显示旧行为，请先确认已经重新构建并重启后端进程。

首页分类和轮播图依赖数据库中的 `categories`、`carousel` 数据。首次运行时，请先执行 `sql/bookstore.sql` 创建表，再执行 `sql/mock.sql` 写入示例数据。

## 注意事项

- 配置文件路径是相对于当前工作目录的 `conf/config.yaml`，请从后端根目录启动程序。
- 本地默认只监听 `localhost`；Docker Compose 会通过 `BOOKSTORE_SERVER_HOST=0.0.0.0` 让服务可被其他容器访问。
- MySQL 和 Redis 都是启动必需依赖；Redis 用于图书缓存和验证码答案存储。
- 分类列表、按分类查询图书、首页轮播图、用户订单详情和管理后台接口均已注册。
- 当前没有独立的后端退出登录接口。前端退出时删除本地令牌；如需服务端令牌撤销，需要继续实现黑名单或会话机制。
- 路由和 Repository 已包含自动化测试。修改后端后应执行：

```bash
GOCACHE=/private/tmp/bookstore-go-cache go test ./...
git diff --check
```

## License

许可证信息见 [LICENSE](LICENSE)。
