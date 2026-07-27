# 当前任务：SaaS 质量与生产可用性加固（任务 16–20 打包）

> 本文件由 Claude 生成，Codex 请直接阅读并执行。
> 执行前请先阅读 `DESIGN.md` 和 `.codex/codex-log.md` 了解完整设计和项目状态。
> 当前分支：`codex/first-core-features`。
> 前置：任务 15 已完成，`/uploads` 直链已移除，资源访问改走受保护接口
> `GET /api/v1/resources/:id/file` 与 `GET /backend/v1/resources/:id/file`。

## 总目标

一次性完成五个加固子任务，把系统从「功能可用」推进到「生产可用」。
**这是打包任务，但必须严格遵守下面的执行纪律，否则打回重做。**

## 执行纪律（最高优先级，违反即打回）

1. **五个子任务，五个独立 git commit**。每个子任务完成、测试通过后单独提交，
   commit message 以 `feat(task-N):` 开头。禁止把多个子任务揉进一个 commit。
2. **严格按顺序执行**：16 → 17 → 18 → 19 → 20。后面的任务可以依赖前面的成果。
3. **文件范围不重叠**：每个子任务只动自己「允许修改的范围」里列的文件。
   如果某文件被多个子任务需要（如 `server.go`、`config.go`），在**最后一个**
   需要它的子任务里统一改，并在该子任务的汇报中说明。
4. 每个子任务都要满足自己的验收标准，并保证 `go test ./...` 在该 commit 时通过。
5. 任何超出范围的想法，记录到「遇到的问题」里，**不要顺手实现**。

---

## 子任务 16：学员端资源 ID 贯通

### 目标
课程课时引用的资源从「存 URL」升级为「存资源 ID」，恢复 PC/H5 学员端课程播放。

### 允许修改的范围
- `internal/domain` 课时（Lesson）模型（加 `resource_id`，保留 `url`）
- `internal/repository/course_lesson*.go`、`internal/service/course_lesson.go`、`internal/api/course_lesson.go`
- 课程 Detail 返回（`internal/api/course.go`、`internal/service/course.go`）
- `web/pc/src`、`web/h5/src` 课程播放组件；`web/admin/src` 课时创建/编辑
- 对应 `*_test.go`
- **不要动** `internal/api/resource.go`、`internal/service/resource.go` 的访问控制逻辑（任务 15 已实现，只复用）

### 具体要求
1. 课时模型加 `resource_id`（可空 string，兼容旧数据），保留旧 `url` 字段不删。
2. 学员端 `GET /api/v1/courses/:id`（PublishedDetail）返回课时携带 `resource_id`、`resource_type`。
3. PC/H5 播放组件：有 `resource_id` 时用带 token 的请求调 `/api/v1/resources/:id/file` 取 blob → objectURL 播放；无 `resource_id` 旧数据给出降级处理（说明取舍）。
4. 管理端创建/编辑课时时从资源库选资源并保存 `resource_id`。
5. `resource_id` 字段的添加可先靠 AutoMigrate（版本管理在任务 20 统一做）。

### 验收标准
- 学员凭 token 能取本租户课时资源（200），跨租户 404。
- 课时 Detail 正确返回 `resource_id`/`resource_type`。
- PC/H5 能播放已发布课程视频/查看文档。
- 旧 `url` 数据不报错。

---

## 子任务 17：认证加固（refresh token + 登出吊销 + 登录/注册限流）

### 目标
解决 JWT 无法吊销、无刷新机制的问题，并给登录/注册接口加内存限流防暴力破解。

### 允许修改的范围
- `internal/security/`（refresh token 生成/校验）
- `internal/service/auth.go`、`internal/api/auth.go`
- 新增 `internal/repository/refresh_token*.go`、`internal/domain` refresh_token 模型
- 新增 `internal/middleware/ratelimit.go`（内存限流）
- `internal/server/server.go`（挂载限流中间件到登录/注册路由；refresh/logout 路由）
- 对应 `*_test.go`
- `web/admin/src/api/auth.ts`（如 refresh 流程需要前端配合）

### 已确认决策
- **refresh token 服务端存表**：新增 `refresh_tokens` 表，存 token 哈希、user_id、tenant_id、过期时间、吊销标记。支持轮换（每次刷新签发新 refresh 并作废旧）与真正吊销（登出时置吊销位）。
- **限流用内存方案**：进程内计数器/令牌桶，不引入 Redis。按「IP + 租户」维度限制登录/注册频率。

### 具体要求
1. 登录/注册成功后除签发 24h access token 外，签发更长有效期（如 7 天）的 refresh token 并存表（存哈希不存明文）。
2. 新增 `POST /api/v1/auth/refresh`：校验 refresh token 有效且未吊销 → 签发新 access token + 新 refresh token（旧 refresh 作废）。
3. 新增 `POST /api/v1/auth/logout`（需认证）：吊销当前用户的 refresh token。
4. 登录 `POST /api/v1/auth/login`、注册 `POST /api/v1/tenants/register`、`POST /api/v1/auth/register` 加内存限流，超限返回 429。限流阈值可配置（放 config，给默认值）。
5. 内存限流要考虑并发安全（mutex/atomic），并提供清理过期计数的机制防内存泄漏。
6. 保持旧 access token 校验逻辑不变（中间件不动）。

### 验收标准
- 完整流程：登录 → refresh 换新 token（旧 refresh 再用失败）→ logout 后 refresh 失败。
- 限流：连续超阈值请求返回 429，窗口过后恢复。
- 已有认证测试不受影响，`go test ./...` 通过。

---

## 子任务 18：superadmin 初始化（一次性引导接口）+ 租户删除 409 语义

### 目标
提供首次部署的 superadmin 创建入口；修正租户删除在有用户时返回 500 的错误语义。

### 允许修改的范围
- `internal/api/auth.go` 或新增 `internal/api/bootstrap.go`（一次性引导接口）
- `internal/service/auth.go` 或新增 `internal/service/bootstrap.go`
- `internal/service/tenant.go`、`internal/api/tenant.go`（删除语义 409）
- `internal/server/server.go`（挂载引导路由）
- 对应 `*_test.go`

### 已确认决策
- **一次性引导接口**：`POST /api/v1/bootstrap/superadmin`。仅当系统中**尚无任何 superadmin** 时可用，创建成功后自动失效（再有 superadmin 存在时返回 403/409）。

### 具体要求
1. 引导接口无需认证，但内部先检查是否已存在 superadmin，存在则拒绝。
2. 请求体含 email、name、password；创建 role=`superadmin` 的用户（superadmin 不属于任何租户，tenant_id 处理与现有逻辑一致，参考 `internal/service/auth.go:34`）。
3. 创建成功返回用户信息（可顺便签发 token，与现有登录一致）。
4. 租户删除：当租户仍有用户被外键拒绝时，返回 409 Conflict 并给出明确信息，而非 500。在 repository/service 层把外键约束错误映射为 409。
5. superadmin 创建走 bcrypt 哈希（复用现有逻辑）。

### 验收标准
- 空库首次调用引导接口成功创建 superadmin；再次调用被拒绝。
- 删除仍有用户的租户返回 409 而非 500。
- `go test ./...` 通过。

---

## 子任务 19：结构化日志 + 请求 ID

### 目标
替换 gin 默认日志为结构化日志，每个请求注入并输出 request ID，便于生产排障与链路追踪。

### 允许修改的范围
- `internal/server/server.go`（替换 `gin.Logger()`）
- 新增 `internal/middleware/logging.go`、`internal/middleware/requestid.go`
- 如需要，`internal/config/config.go`（日志级别/格式配置）
- 对应 `*_test.go`

### 具体要求
1. 新增请求 ID 中间件：每个请求生成 UUID（或复用客户端 `X-Request-ID` 头），写入 context 与响应头 `X-Request-ID`。
2. 用结构化日志（Go 标准库 `log/slog`，不引第三方）替换 `gin.Logger()`，每条请求日志输出：request_id、method、path、status、耗时、client_ip、tenant_id（若有）、user_id（若有）。
3. 日志级别/JSON 或 text 格式可通过 config 配置，给合理默认值。
4. 保留 `gin.Recovery()`，panic 时也输出 request_id。
5. 不要改动业务逻辑，只动日志/中间件层。

### 验收标准
- 请求后日志含 request_id 且响应头带回 `X-Request-ID`。
- 日志为结构化格式，关键字段齐全。
- `go test ./...` 通过。

---

## 子任务 20：数据库迁移版本管理（自研版本表）

### 目标
在保留 AutoMigrate 的基础上，加入版本化管理，让 schema 变更可追溯、可按序执行。

### 允许修改的范围
- `internal/migration/migration.go` 及新增迁移文件
- `cmd/server/main.go`（启动时执行迁移）
- `internal/config/config.go`（如需要）
- 对应 `*_test.go`

### 已确认决策
- **自研版本表**：新增 `schema_migrations` 表（version、applied_at）。迁移以 Go 代码形式按版本号登记，启动时按序执行未应用的迁移，并记录到版本表。

### 具体要求
1. 设计 `schema_migrations` 表与迁移注册机制（版本号 + 迁移函数）。
2. 首个版本（v1）执行现有 `AutoMigrate` 的全量建表，作为基线。
3. 后续 schema 变更（如任务 16 的 `resource_id` 字段）以新迁移版本加入，**幂等**（重复执行安全）。
4. 启动时自动执行未应用的迁移，记录版本；已应用的跳过。
5. 提供测试：全新库执行到最新版本、重复执行不产生重复/报错。
6. 不要引入 golang-migrate 等外部依赖。

### 验收标准
- 全新数据库启动后所有表（含 `resource_id` 字段、`refresh_tokens` 表、`schema_migrations` 表）就绪。
- 重复启动不重复执行已应用迁移，不报错。
- `go test ./...` 通过。

---

## 不要做（全任务通用）

- 不要引入 Redis、S3/MinIO、golang-migrate、第三方日志库、第三方限流库。
- 不要做 signed URL、分片上传、CDN、复杂 ACL/RBAC。
- 不要删除课时旧 `url` 字段。
- 不要修改 `.codex/` 下协作文档（由 Claude 维护）。
- 不要把多个子任务合并成一个 commit。

## 总验收标准

1. 五个独立 commit，每个对应一个子任务，`git log` 可清晰区分。
2. 每个 commit 时点 `go test ./...` 通过。
3. `web/pc`、`web/h5`、`web/admin` 三端 `npm run build` 全部通过。
4. 全部子任务的验收标准逐条满足。

## 完成后需要返回

1. 五个 commit 的 hash + message 列表。
2. 每个子任务的：修改文件列表、该子任务的 git diff、关键设计取舍。
3. `go test ./...` 最终结果 + 三端 `npm run build` 结果。
4. 任务 16 的「无 resource_id 旧数据降级方案」说明、任务 20 的迁移机制说明。
5. 遇到的问题与超出范围但未做的事项清单。
