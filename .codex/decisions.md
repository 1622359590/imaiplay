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

- 第一阶段：单库 + 租户字段隔离。
- 未来可扩展：按租户分 schema 或分库。

### 认证与权限

- JWT + RBAC。
- 角色：superadmin / tenant_admin / instructor / learner。

### 存储策略

- 对象存储抽象层，支持 S3、MinIO、本地文件系统。
- 支持预签名 URL 和分片上传。
