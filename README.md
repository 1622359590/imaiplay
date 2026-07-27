# ImaiPlay SaaS 培训系统

> 一款面向多租户的企业培训 SaaS 后端系统，采用 Go 语言开发。

## 项目定位

ImaiPlay 是一个 **多租户企业培训 SaaS 平台**，让每个客户（租户）都能独立入驻、自定义培训内容、管理学员、追踪学习进度。

## 核心原则

- **从零开发**：不参考、不复制 PlayEdu 或其他现有系统的代码。
- **SaaS 优先**：多租户隔离、可扩展、可配置。
- **简单可维护**：模块边界清晰，代码易于理解和测试。

## 技术栈

| 层级 | 技术 |
|------|------|
| 语言 | Go 1.22+ |
| Web 框架 | Gin |
| 配置管理 | Viper |
| 数据库 | MySQL 8 / PostgreSQL（待定） |
| ORM | GORM（待定） |
| 缓存 | Redis（待定） |
| 消息队列 | 待定 |
| 对象存储 | S3 / MinIO / 本地（抽象层） |
| 认证 | JWT + RBAC |
| 部署 | Docker / Kubernetes（远期） |

## 模块结构（规划）

```
imaiplay-go/
├── cmd/server/            # 服务入口
├── internal/
│   ├── config/            # 配置加载
│   ├── context/           # 请求上下文（租户、用户）
│   ├── middleware/        # HTTP 中间件
│   ├── server/            # HTTP 服务器
│   ├── domain/            # 领域模型
│   ├── service/           # 业务逻辑
│   ├── repository/        # 数据访问
│   ├── api/               # HTTP handler
│   └── security/          # 认证与权限
├── pkg/                   # 公共库
├── .codex/                # 协作记录
├── Makefile
└── README.md
```

## 多租户策略

采用 **共享数据库 + 租户字段隔离** 作为第一阶段方案：

- 每张业务表包含 `tenant_id` / `tenant_code` 字段。
- 请求上下文通过子域名或 Header 解析租户。
- Repository 层自动注入租户过滤条件。

未来可扩展为按租户分 schema 或分库。

## 角色体系

| 角色 | 说明 |
|------|------|
| superadmin | 平台超级管理员，管理租户 |
| tenant_admin | 租户管理员，管理租户内所有资源 |
| instructor | 讲师，可创建课程、查看学习数据 |
| learner | 学员，学习课程、查看个人进度 |

## 核心领域

- **Tenant（租户）**：SaaS 客户单位。
- **User（用户）**：系统用户，可属于一个或多个租户。
- **Course（课程）**：培训内容，包含章节和课时。
- **Chapter（章节）**：课程的章节。
- **Lesson（课时）**：最小学习单元，支持视频、文档等类型。
- **Enrollment（学习记录）**：学员与课程的关系。
- **Progress（进度）**：学员对每个课时的学习进度。

## 协作记录规范

所有协作记录保存在 `.codex/` 目录下：

- `decisions.md`：架构决策与原因
- `progress.md`：任务执行进度
- `issues.md`：问题与待决定事项
- `knowledge-graph.md`：模块与依赖关系图谱

## 开发节奏

按小步快跑方式推进，每个任务边界清楚、可验证。详细任务见 `.codex/progress.md`。

## 作者

ImaiWork
