# ImaiPlay SaaS 培训系统

> 一款面向多租户的企业培训 SaaS 平台，包含完整的前后端系统。

## 项目定位

ImaiPlay 是一个 **多租户企业培训 SaaS 平台**，支持客户完全自助开通租户、管理学员、发布课程、追踪学习进度。项目包含：

- **后端 API**：Go + Gin + PostgreSQL + GORM
- **管理后台**：React 18 + Ant Design 5
- **PC 学员端**：React 18 + Ant Design 5
- **H5 学员端**：React 18 + Ant Design Mobile 5

## 核心功能

- **SaaS 自助开通**：客户访问注册页即可创建租户，自动初始化演示数据
- **多租户隔离**：共享数据库 + 租户字段隔离，数据安全隔离
- **课程管理**：课程 → 章节 → 课时三级结构，支持视频、文档、文本
- **学习追踪**：课时学习进度上报、最近学习记录、学习统计
- **资源管理**：图片、视频、文档上传与分类管理
- **权限体系**：superadmin / tenant_admin / instructor / learner 四级角色

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端语言 | Go 1.22+ |
| 后端框架 | Gin |
| 配置管理 | Viper |
| 数据库 | PostgreSQL |
| ORM | GORM |
| 认证 | JWT + bcrypt |
| 前端框架 | React 18 + TypeScript + Vite |
| 前端 UI | Ant Design 5（管理后台/PC）+ Ant Design Mobile 5（H5） |
| 状态管理 | Redux Toolkit |
| 对象存储 | 本地文件系统（S3 / MinIO 远期支持） |
| 缓存 | 第一阶段本地缓存，远期 Redis |
| 部署 | Docker / Docker Compose（远期 Kubernetes） |

## 项目结构

```
imaiplay-go/
├── cmd/server/            # 服务入口
├── internal/              # 后端代码
│   ├── config/            # 配置加载
│   ├── context/           # 请求上下文（租户、用户）
│   ├── middleware/        # HTTP 中间件
│   ├── server/            # HTTP 服务器
│   ├── db/                # 数据库连接
│   ├── domain/            # 领域模型
│   ├── service/           # 业务逻辑
│   ├── repository/        # 数据访问
│   ├── api/               # HTTP handler
│   ├── security/          # 密码、JWT
│   ├── migration/         # 数据库迁移
│   ├── storage/           # 对象存储抽象
│   └── errorsx/           # 统一错误响应
├── web/                   # 前端项目
│   ├── admin/             # 管理后台（React 18 + Ant Design 5）
│   ├── pc/                # PC 学员端（React 18 + Ant Design 5）
│   └── h5/                # H5 学员端（React 18 + Ant Design Mobile）
├── pkg/                   # 公共库
├── .codex/                # 协作记录
│   ├── current-task.md    # 当前任务指令
│   ├── codex-log.md       # 协作日志
│   ├── decisions.md       # 架构决策
│   ├── progress.md        # 开发进度
│   ├── issues.md          # 问题与待决定事项
│   └── knowledge-graph.md # 模块关系图谱
├── DESIGN.md              # 完整系统设计文档
├── Makefile
└── README.md
```

## 多租户策略

采用 **共享数据库 + 租户字段隔离**：

- 每张业务表包含 `tenant_id` 字段
- 请求通过子域名或 Header `X-Tenant-Code` 识别租户
- Repository 层自动注入租户过滤条件
- 未来可扩展为按租户分 schema 或分库

## 角色体系

| 角色 | 说明 |
|------|------|
| superadmin | 平台超级管理员，管理租户 |
| tenant_admin | 租户管理员，管理租户内所有资源 |
| instructor | 讲师，可创建课程、管理自己创建的课程 |
| learner | 学员，学习课程、查看个人进度 |

## 核心领域

- **Tenant（租户）**：SaaS 客户单位
- **User（用户）**：系统用户，属于单个租户
- **Course（课程）**：培训内容，包含章节和课时
- **CourseChapter（章节）**：课程的章节
- **CourseLesson（课时）**：最小学习单元（视频、文档、文本）
- **CourseEnrollment（报名）**：学员与课程的关系
- **LessonProgress（进度）**：学员对每个课时的学习进度
- **Resource（资源）**：图片、视频、文档
- **ResourceCategory（资源分类）**：资源分类管理

## SaaS 自助开通流程

1. 客户访问注册页，填写组织名称、管理员邮箱、姓名、密码
2. 系统自动创建租户（基于组织名称生成租户代码）
3. 自动创建管理员账号（tenant_admin 角色）
4. 自动初始化演示数据（示例课程、学员、讲师、资源）
5. 自动签发 JWT，客户直接进入管理后台

## 开发节奏

按小步快跑方式推进，每个任务边界清楚、可验证。详细任务见 `.codex/progress.md`。

## 协作记录规范

所有协作记录保存在 `.codex/` 目录下：

- `current-task.md`：当前任务指令（Codex 直接阅读执行）
- `codex-log.md`：协作日志（历史决策与评审记录）
- `decisions.md`：架构决策与原因
- `progress.md`：任务执行进度
- `issues.md`：问题与待决定事项
- `knowledge-graph.md`：模块与依赖关系图谱

## 作者

ImaiWork
