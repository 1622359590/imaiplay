# 知识图谱

## 模块关系

```mermaid
graph TD
    Client[客户端] -->|HTTP| Server[internal/server]
    Server --> Router[Router]
    Router --> Middleware[internal/middleware]
    Middleware --> Context[internal/context]
    Middleware --> Handler[Handler]
    Server --> DBHealth[数据库健康检查函数]
    DBHealth --> DB[internal/db]
    Handler --> Service[internal/service]
    Service --> Repository[internal/repository]
    Repository --> Domain[internal/domain]
    Repository --> DB
    Migration[internal/migration] --> Domain
    Migration --> DB
    DB --> PostgreSQL[(PostgreSQL)]
    Service --> Storage[Storage 抽象]
    Storage --> S3[S3/MinIO]
    Storage --> Local[本地文件]
```

## 当前已规划模块

| 模块 | 职责 |
|------|------|
| cmd/server | 服务入口 |
| internal/config | 配置加载 |
| internal/context | 请求上下文（租户、用户） |
| internal/middleware | HTTP 中间件（租户识别、日志、恢复） |
| internal/server | HTTP 服务器与路由注册 |
| internal/db | PostgreSQL 连接、连接池与健康检查 |
| internal/domain | Tenant 等领域模型 |
| internal/migration | GORM 自动迁移 |
| internal/service | 业务逻辑 |
| internal/repository | Tenant Repository 接口与 GORM 实现 |
| internal/api | HTTP handler |
| internal/security | 认证与权限 |
| pkg | 公共库 |
| .codex | 协作记录 |

## 数据流

1. 请求到达 Nginx / Server
2. Middleware 识别 tenant
3. Context 携带 tenant 进入 Handler
4. Handler 调用 Service
5. Service 调用 Repository，Repository 通过 GORM 访问 PostgreSQL
6. 启动时 Migration 自动创建或更新 Tenant 表
7. `/health/db` 通过注入函数检查数据库连通性
8. 返回响应
