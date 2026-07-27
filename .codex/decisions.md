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
- 首阶段仅实现本地文件系统存储驱动。
- 允许上传类型：图片（JPEG、PNG、WebP）、视频（MP4、WebM）、文档（PDF）。
- 单文件最大 500 MB。
- `resource_type` 依据实际 MIME 类型识别，不信任文件扩展名。
- 不支持的类型或超限返回 HTTP 400。
- 删除资源时先删元数据再删文件；文件删除失败则恢复原记录以便重试。
- 有子分类时拒绝删除父分类，分类删除时清空资源的分类引用。
- 支持预签名 URL 和分片上传（远期）。

### 前端技术栈

- 框架：React 18 + TypeScript + Vite。
- UI 库：管理后台和 PC 学员端使用 Ant Design 5；H5 学员端使用 Ant Design Mobile 5。
- 状态管理：Redux Toolkit。
- HTTP 客户端：axios，Token 存储于 localStorage。
- 项目位置：与后端同一仓库，`web/admin`、`web/pc`、`web/h5`。
- UI 风格：参考 PlayEdu 管理后台、PC 学员端、H5 学员端的布局和配色，独立实现，不复制其代码。

### 课程结构与权限

- 课程内容采用 Course → CourseChapter → CourseLesson 三层结构。
- 课程状态保持 `int`：0 为草稿，1 为已发布。
- `tenant_admin` 管理租户全部课程；`instructor` 仅管理自己创建的课程。
- 学员端展示租户内已发布课程，无需报名即可浏览列表和详情。
- 学员上报课时进度前必须已报名该课程，且 `CourseEnrollment.status = 1`，否则返回 HTTP 403。
- 课程、章节和课时查询均在 Repository 层强制使用 JWT `tenant_id`。
- Server 仅注入课程 Service 接口，不依赖 GORM 实现。
- 学习进度按 JWT `user_id` 隔离，并通过 `CourseEnrollment` 校验课程访问权。
- 学习进度按租户、用户、课时唯一，0/1-99/100 分别映射未开始、进行中、已完成。
- 最近学习按进度更新时间倒序返回课程、课时与进度。

### 前端运行边界

- 三端独立构建，开发端口依次为 5173、5174、5175。
- 后端 CORS 仅允许上述 localhost 来源及任务规定的请求头和方法。
