# ImaiPlay SaaS 培训系统

> 面向多租户的企业培训 SaaS 平台，支持自助开通、课程管理、学习追踪、资源管理。

## 项目定位

ImaiPlay 是一个多租户企业培训 SaaS 平台。客户自助注册即可开通独立租户，管理学员、发布课程、追踪学习进度。系统包含完整的后端 API 和三端前端应用。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM / PostgreSQL |
| 认证 | JWT + bcrypt + Refresh Token 轮换 |
| 前端 | React 18 / TypeScript / Vite / Ant Design 5 |
| 存储 | 本地文件系统 + S3 兼容存储（阿里云 OSS / MinIO） |
| 短信 | 阿里云短信通道 |
| 部署 | Docker Compose / Nginx 反向代理 |
| 文档 | Swagger/OpenAPI（swaggo 自动生成） |

## 核心功能

### 租户管理
- SaaS 自助注册开通，自动初始化演示数据
- 超级管理员手动创建租户
- 租户生命周期控制（试用 / 正常 / 暂停 / 删除）
- 自定义域名绑定
- 品牌主题定制（品牌色、Logo、欢迎语）
- 套餐与存储配额管理

### 课程体系
- 课程 → 章节 → 课时三级结构
- 支持视频、文档、文本三种课时类型
- superadmin 可完整维护官方课程、章节、课时与平台资源
- 租户启用官方课程后可供管理员预览、学员报名学习
- 讲师管理自己创建的课程，tenant_admin 管理所有课程

### 学习追踪
- 学员报名课程后学习课时
- 课时学习进度上报（未开始 / 进行中 / 已完成）
- 最近学习记录
- 租户级统计看板（用户数、课程数、学习数据、完成率）

### 资源管理
- 图片（JPEG / PNG / WebP）、视频（MP4 / WebM）、文档（PDF）上传
- 封面、Logo、视频和 PDF 使用可视化上传控件，支持进度、预览、替换、删除与失败重试
- 平台资源与租户资源隔离，官方课程可复用 superadmin 上传的平台资源
- 资源分类管理
- 视频和 PDF 使用受保护的文件访问接口；平台封面图片可公开展示
- 官方课程资源按租户启用状态、用户角色和学员报名状态授权访问
- 运行时切换存储后端（本地 / S3）

## 本次功能与优化

- 完善 superadmin 官方课程 CRUD，并支持维护章节、视频课时、PDF 课时和文本课时。
- 新增平台资源上传、列表、预览和删除接口，阻止删除仍被官方课程引用的资源。
- 官方课程删除时同步清理租户启用、报名和学习进度等关联数据。
- 管理后台只保留一个「官方课程」入口，避免 superadmin 看到重复的租户课程菜单。
- 课程封面、主题 Logo 和资源管理页取消媒体 URL 文本框，统一改为更直观的上传交互。
- PC 与 H5 学员端继续通过受保护接口播放或下载课程资源，不暴露存储地址。
- 租户自定义域名支持一键自动绑定：DNS 验证、宝塔建站、反向代理、HTTPS 证书和 `/admin` 访问限制由系统完成。

### 认证与安全
- 邮箱密码注册登录
- 手机短信验证码登录
- 手机密码找回（短信验证码）
- JWT + Refresh Token 轮换与吊销
- 登录 / 注册接口限流
- bcrypt 密码哈希
- 四级角色：superadmin / tenant_admin / instructor / learner

### 运维支撑
- 操作审计日志（自动记录非 GET 请求，敏感字段过滤）
- 结构化日志 + 请求 ID
- 数据库迁移版本管理（schema_migrations）
- PostgreSQL 备份恢复脚本
- Docker Compose 一键部署

## 角色体系

| 角色 | 说明 |
|------|------|
| superadmin | 平台超级管理员，管理所有租户、套餐、官方课程 |
| tenant_admin | 租户管理员，管理租户内全部资源 |
| instructor | 讲师，创建和管理自己发布的课程 |
| learner | 学员，学习课程、查看个人进度 |

## 项目结构

```
imaiplay-go/
├── cmd/server/                # 服务入口
├── internal/                  # 后端代码
│   ├── api/                   # HTTP handler（含 Swagger 注解）
│   ├── config/                # 配置加载（Viper）
│   ├── context/               # 请求上下文（租户、用户）
│   ├── middleware/            # 中间件（认证、租户、审计、限流、日志、请求ID）
│   ├── server/                # HTTP 服务器与路由注册
│   ├── db/                    # PostgreSQL 连接与健康检查
│   ├── domain/                # 领域模型
│   ├── migration/             # 数据库迁移（版本管理）
│   ├── service/               # 业务逻辑
│   ├── repository/            # 数据访问层（接口 + GORM 实现）
│   ├── security/              # 密码哈希、JWT 签发与验证
│   ├── storage/               # 存储抽象（本地文件 + S3 兼容）
│   ├── sms/                   # 短信服务（阿里云）
│   ├── baota/                 # 宝塔 API 客户端（建站、反代、证书与 Nginx）
│   ├── errorsx/               # 统一错误码与 HTTP 错误响应
│   └── test/integration/      # 集成测试
├── web/                       # 前端项目
│   ├── admin/                 # 管理后台（React 18 + Ant Design 5）
│   ├── pc/                    # PC 学员端（React 18 + Ant Design 5）
│   └── h5/                    # H5 学员端（React 18 + Ant Design Mobile 5）
├── docs/                      # Swagger 文档（自动生成）
├── docker/nginx/conf/         # Nginx 反向代理配置
├── scripts/                   # 备份恢复脚本
├── .codex/                    # AI 协作记录
├── Dockerfile                 # 多阶段构建
├── docker-compose.yml         # PostgreSQL + 应用服务
├── Makefile                   # 构建 / 测试 / 运行 / Docker
├── .env.example               # 环境变量模板
├── DESIGN.md                  # 完整系统设计文档
└── README.md
```

## 多租户架构

共享数据库 + 租户字段隔离：

- 每张业务表包含 `tenant_id` 字段
- 请求通过子域名或 Header `X-Tenant-Code` 识别租户
- Repository 层自动注入租户过滤条件
- 未来可扩展为按租户分 schema 或分库

## API 概览

### 公开接口
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/health/db` | 数据库连通性检查 |
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 邮箱密码登录 |
| POST | `/api/v1/auth/login-code/send` | 发送短信验证码 |
| POST | `/api/v1/auth/login-code` | 短信验证码登录 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |
| POST | `/api/v1/auth/forgot-password` | 密码找回 |
| POST | `/api/v1/auth/reset-password` | 重置密码 |
| POST | `/api/v1/tenants/register` | 租户自助注册 |
| POST | `/api/v1/bootstrap/superadmin` | 初始化超级管理员 |
| GET | `/swagger/index.html` | API 文档（需启用） |

### 管理后台接口（需认证）
| 模块 | 接口 |
|------|------|
| 租户管理 | CRUD、生命周期控制 |
| 域名设置 | CNAME/DNS 验证、自动绑定、状态查询、解绑 |
| 用户管理 | CRUD、角色管理 |
| 课程管理 | CRUD、章节、课时、报名 |
| 资源管理 | 上传、分类、文件访问 |
| 套餐管理 | CRUD、分配、使用量查询 |
| 主题设置 | 获取、更新品牌主题 |
| 审计日志 | 按租户/用户/时间筛选 |
| 短信配置 | 获取、保存、测试 |
| 存储配置 | 获取、保存、测试 |
| 统计看板 | 用户/课程/学习/完成率统计 |
| 官方课程 | 创建、列表、启用/禁用 |

### 学员端接口（需认证）
| 接口 | 说明 |
|------|------|
| 课程列表 / 详情 | 浏览已发布课程 |
| 课时进度上报 | 上报学习进度 |
| 最近学习记录 | 按更新时间倒序 |
| 资源文件访问 | 受保护的文件下载 |

## 快速开始

### 本地开发

```bash
# 1. 安装依赖
go mod download
cd web/admin && npm install
cd web/pc && npm install
cd web/h5 && npm install

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env，配置数据库连接等

# 3. 启动后端
make run

# 4. 启动前端（三个终端分别运行）
cd web/admin && npm run dev    # 管理后台 http://localhost:5173
cd web/pc && npm run dev       # PC 学员端 http://localhost:5174
cd web/h5 && npm run dev       # H5 学员端 http://localhost:5175
```

### Docker 部署

```bash
cp .env.example .env
# 编辑 .env，修改生产环境密钥

make docker-up
```

服务启动后：

- 管理后台：`http://localhost/`
- PC 学员端：`http://localhost/pc/`
- H5 学员端：`http://localhost/h5/`
- 健康检查：`http://localhost:8080/health`
- API 文档：`http://localhost:8080/swagger/index.html`（需 `SWAGGER_ENABLED=true`）

```bash
make docker-down   # 停止服务
```

数据库、上传文件和运行时配置分别保存在 `postgres_data`、`uploads`、`app_config`
Docker volume 中。OSS 与短信密钥会写入 `app_config`，重建服务容器后仍会保留。
请保持 `.env` 中的 `JWT_SECRET` 不变，否则已加密的运行时密钥无法解密。

### 租户自定义域名自动绑定

租户管理员在管理后台的「域名设置」中填写域名，按页面提示将域名配置为 CNAME 指向平台域名，然后依次点击「验证域名」和「自动绑定」。系统会自动完成：

1. 校验域名格式、保留域名和重复绑定。
2. 查询 CNAME 及最终 A/AAAA 解析，确认指向服务器公网 IP。
3. 调用宝塔 API 创建站点并配置反向代理。
4. 写入租户站点的 Nginx 配置，禁止访问 `/admin`，执行配置检查和 reload。
5. 申请 Let's Encrypt 证书并轮询证书订单，成功后保存租户域名。

宝塔 API 自动化需要先在宝塔「面板设置 → API 接口」中开启接口并生成 API Key，然后在 `.env` 中配置：

```dotenv
BAOTA_PANEL_URL=http://host.docker.internal:8888
BAOTA_API_KEY=替换为宝塔生成的 API Key
BAOTA_SERVER_IP=你的服务器公网 IP
BAOTA_PROXY_TARGET=http://127.0.0.1:18080
```

Docker 环境中 `BAOTA_PANEL_URL` 应使用 `host.docker.internal` 访问宿主机宝塔面板，不要填写容器内的 `127.0.0.1`。同时在宝塔 API 白名单中放行服务器本机或 Docker 网桥来源。`BAOTA_PROXY_TARGET` 应填写宝塔 Nginx 能访问到的 ImaiPlay 后端地址；使用 `docker-compose.bt.yml` 时通常是 `http://127.0.0.1:18080`。

域名绑定失败时系统会回滚已创建的宝塔站点；解绑时会先删除宝塔站点，再清除租户域名。证书续期由宝塔的 Let's Encrypt 自动续期任务负责。完整 DNS、宝塔 API 和 HTTPS 说明见 [`docs/域名配置指南.md`](docs/域名配置指南.md)。

## 测试

```bash
# 后端单元测试
go test ./...

# 查看覆盖率
go test -cover ./internal/service/... ./internal/repository/...

# 集成测试
go test ./internal/test/integration/...

# 前端构建验证
cd web/admin && npm run build
cd web/pc && npm run build
cd web/h5 && npm run build
```

## 数据库表

| 表名 | 说明 |
|------|------|
| tenants | 租户（含主题、套餐、域名、生命周期） |
| users | 用户 |
| courses | 课程 |
| course_chapters | 章节 |
| course_lessons | 课时 |
| course_enrollments | 课程报名 |
| lesson_progress | 学习进度 |
| resources | 资源 |
| resource_categories | 资源分类 |
| refresh_tokens | Refresh Token |
| password_resets | 密码重置 |
| audit_logs | 审计日志 |
| plans | 套餐 |
| schema_migrations | 迁移版本 |

## 后续规划

- 消息通知（站内信 / 邮件）
- 考试/测验系统
- 证书发放
- 学习路径/学习计划
- 第三方登录（微信 / 钉钉 / 飞书）
- 分片上传 / 断点续传
- Redis 缓存
- 视频加密 / 防盗链
- Kubernetes 部署

## 作者

ImaiWork
