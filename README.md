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
- 官方课程市场（superadmin 创建，所有租户可见）
- 讲师管理自己创建的课程，tenant_admin 管理所有课程

### 学习追踪
- 学员报名课程后学习课时
- 课时学习进度上报（未开始 / 进行中 / 已完成）
- 最近学习记录
- 租户级统计看板（用户数、课程数、学习数据、完成率）

### 资源管理
- 图片（JPEG / PNG / WebP）、视频（MP4 / WebM）、文档（PDF）上传
- 资源分类管理
- 受保护的文件访问接口（需认证 + 租户隔离）
- 运行时切换存储后端（本地 / S3）

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
| 租户管理 | CRUD、自定义域名、生命周期控制 |
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

数据库和上传文件分别保存在 `postgres_data`、`uploads` Docker volume 中。

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
