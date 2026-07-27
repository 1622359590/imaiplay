# 架构决策记录

## 2026-07-27

### 项目方向

- 从零开发 ImaiPlay 多租户企业培训 SaaS 后端。
- 不参考、不复制 PlayEdu 或其他现有培训系统的代码。
- 仅借鉴「企业内部培训平台」这一产品方向。

### 后端语言与框架

- 语言：Go 1.22+
- Web 框架：Gin
- 原因：生态成熟、性能优秀、团队易招聘。

### 配置管理

- 使用 Viper，支持 `.env` 文件和环境变量。

### 多租户识别策略（第一阶段）

- 共享数据库，业务表带 `tenant_id` / `tenant_code` 字段。
- 请求层级通过子域名或 Header 识别租户。
- 优先子域名，其次 Header `X-Tenant-ID`，再次 `X-Tenant-Code`，最后标记为 `unknown`。

### 数据库策略

- 数据库：PostgreSQL。
- ORM：GORM。
- 启动时连接数据库并自动迁移 Tenant 表，连接或迁移失败则退出。
- 数据库连接池默认最大打开连接数 25、最大空闲连接数 25。
- 第一阶段：单库 + 租户字段隔离。
- 未来可扩展：按租户分 schema 或分库。

### 数据库健康检查

- Server 通过 `func() error` 注入数据库健康检查，不直接依赖 GORM。
- `/health/db` 连通时返回 HTTP 200，断开时返回 HTTP 503。

### 认证与权限

- JWT + RBAC。
- 角色：superadmin / tenant_admin / instructor / learner。
- 公开注册接口 `/api/v1/auth/register` 禁止注册 `superadmin`，仅允许 `tenant_admin`、`instructor`、`learner`。
- `superadmin` 通过数据库初始化或后续专用引导机制创建，不对外开放注册。
- 密码使用 bcrypt `DefaultCost` 哈希，API 永不序列化密码字段。
- JWT 使用 HS256、24 小时有效期及 `imaiplay` issuer。
- API 错误统一为 `code`、`message`、`data`，由 `errorsx` 映射 HTTP 状态。
- 用户按 JWT `tenant_id` 在 Repository 层隔离；superadmin 的租户管理为平台级查询。
- User 通过外键关联 Tenant，删除仍有用户的租户时由数据库拒绝。
- 用户租户邮箱唯一索引由迁移显式创建，避免污染通用 BaseModel。

### 存储策略

- 对象存储抽象层，支持 S3、MinIO、本地文件系统。
- 支持预签名 URL 和分片上传。
