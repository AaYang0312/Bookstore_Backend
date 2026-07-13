# 博学书城后端

博学书城后端是一个使用 Go 和 Gin 开发的 RESTful API 服务，为书城前端提供用户认证、图书查询、收藏、订单和验证码等功能。

## 技术栈

- Go 1.25
- Gin 1.12
- GORM 1.31
- MySQL 8.x
- Redis 6.x 或更高版本
- JWT

## 功能模块

- 用户：注册、登录、查看及修改个人资料、修改密码
- 图书：首页列表、热销图书、新书、搜索和详情
- 收藏：添加、取消、列表、数量和收藏状态
- 订单：创建订单、订单列表和支付
- 验证码：生成图片验证码，并使用 Redis 保存有效期为 3 分钟的答案

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
| 收藏 | POST | `/favorite/:id` | 是 |
| 收藏 | DELETE | `/favorite/:id` | 是 |
| 收藏 | GET | `/favorite/list` | 是 |
| 收藏 | GET | `/favorite/count` | 是 |
| 收藏 | GET | `/favorite/:id/check` | 是 |
| 订单 | POST | `/order/create` | 是 |
| 订单 | GET | `/order/list` | 是 |
| 订单 | POST | `/order/:id/pay` | 是 |
| 验证码 | GET | `/captcha/generate` | 否 |

需要登录的接口应携带登录后获得的访问令牌：

```http
Authorization: Bearer <access_token>
```

## 与前端联调

1. 启动 MySQL 和 Redis。
2. 启动本服务，确认地址为 `http://localhost:8080`。
3. 再启动 `../bookstore-fronted-master` 中的 React 开发服务器。

后端已配置 CORS，可接受前端开发服务器发起的跨域请求。

## 注意事项

- 配置文件路径是相对于当前工作目录的 `conf/config.yaml`，请从后端根目录启动程序。
- 当前路由只注册了上表中的接口；若前端调用分类、轮播图、退出登录或订单详情接口，需要先在后端实现并注册对应路由。
- 当前项目尚未配置自动化测试命令，修改业务逻辑后建议至少执行 `go test ./...` 和一次前后端联调。

## License

许可证信息见 [LICENSE](LICENSE)。
