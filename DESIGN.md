# ImaiPlay 完整系统设计文档

> 多租户企业培训 SaaS 平台 - Go + Gin + PostgreSQL + GORM + React 18 + Ant Design

## 1. 项目概述

### 1.1 定位
ImaiPlay 是一个面向企业的多租户培训 SaaS 平台，支持多个客户（租户）独立入驻、管理学员、发布课程、追踪学习进度。项目包含完整的后端 API、管理后台前端、学员端前端。

### 1.2 技术栈
- **后端语言**：Go 1.22+
- **后端框架**：Gin
- **配置管理**：Viper
- **数据库**：PostgreSQL
- **ORM**：GORM
- **认证**：JWT + bcrypt
- **前端框架**：React 18 + TypeScript + Vite
- **前端 UI**：Ant Design 5（管理后台）+ Ant Design Mobile（H5 学员端）
- **状态管理**：Redux Toolkit
- **对象存储**：S3 / MinIO / 本地文件系统
- **缓存**：第一阶段可选本地缓存，远期 Redis

### 1.3 核心原则
- 从零开发，不参考、不复制 PlayEdu 代码
- 多租户共享数据库 + 租户字段隔离
- 模块边界清晰，便于测试和维护
- RESTful API，JSON 交互
- 前端 UI 参考 PlayEdu 布局和功能结构，但独立实现

### 1.4 SaaS 自助开通流程

ImaiPlay 支持客户完全自助开通租户：

1. **访问注册页**：客户访问 `/admin/register` 或管理后台登录页的「开通租户」入口
2. **填写信息**：公司/组织名称、管理员邮箱、管理员姓名、密码
3. **自动开通**：系统自动创建租户、管理员账号，并初始化演示数据
4. **立即使用**：注册成功后自动登录管理后台；学员可立即访问 `/t/{tenantCode}` 默认门户

#### 演示数据内容

- 1 个示例课程（含 2 个章节、4 个课时）
- 2 个示例学员（learner 角色）
- 1 个示例讲师（instructor 角色）
- 示例资源（1 张图片、1 个视频）

#### 权限规则

- 开通者自动成为 `tenant_admin`，拥有租户内全部权限
- 同一邮箱可在不同租户注册，但同一租户内唯一
- 租户代码由系统自动生成（基于组织名称拼音或随机字符串）

---

## 2. 系统架构

```
cmd/server              # 服务入口
internal/               # 后端代码
  config/               # 配置加载
  context/              # 请求上下文（tenant、user）
  middleware/           # HTTP 中间件
  server/               # HTTP 服务器与路由
  db/                   # 数据库连接
  domain/               # 领域模型
  repository/           # 数据访问接口与实现
  service/              # 业务逻辑
  api/                  # HTTP handler
  security/             # 密码、JWT
  migration/            # 数据库迁移
  storage/              # 对象存储抽象（远期）
  errorsx/              # 统一错误响应
web/                    # 前端项目
  admin/                # 管理后台（React 18 + Ant Design 5）
  pc/                   # PC 学员端（React 18 + Ant Design 5）
  h5/                   # H5 学员端（React 18 + Ant Design Mobile）
pkg/                    # 公共库
.codex/                 # 协作记录
```

---

## 3. 数据库设计

### 3.1 表清单

| 表名 | 说明 |
|------|------|
| tenants | 租户表 |
| users | 用户表 |
| courses | 课程表 |
| course_chapters | 课程章节表 |
| course_lessons | 课时表 |
| course_enrollments | 学员课程报名/指派表 |
| lesson_progress | 课时学习进度表 |
| resources | 资源表（图片、视频、文档） |
| resource_categories | 资源分类表 |
| login_challenges | 统一登录的短期一次性企业选择凭证 |

### 3.2 租户表 tenants

```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY,
    code VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    status SMALLINT DEFAULT 1, -- 1 active, 0 disabled
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 3.3 用户表 users

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'learner',
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);
```

### 3.4 课程表 courses

```sql
CREATE TABLE courses (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    cover_image VARCHAR(500),
    status SMALLINT DEFAULT 0, -- 0 draft, 1 published
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 3.5 课程章节表 course_chapters

```sql
CREATE TABLE course_chapters (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    course_id UUID NOT NULL REFERENCES courses(id),
    title VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 3.6 课时表 course_lessons

```sql
CREATE TABLE course_lessons (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    chapter_id UUID NOT NULL REFERENCES course_chapters(id),
    title VARCHAR(255) NOT NULL,
    content_type VARCHAR(32) NOT NULL, -- video, document, text
    content_url VARCHAR(500),
    duration_seconds INT DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 3.7 课程报名/指派表 course_enrollments

```sql
CREATE TABLE course_enrollments (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    course_id UUID NOT NULL REFERENCES courses(id),
    user_id UUID NOT NULL REFERENCES users(id),
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, course_id, user_id)
);
```

### 3.8 课时学习进度表 lesson_progress

```sql
CREATE TABLE lesson_progress (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    lesson_id UUID NOT NULL REFERENCES course_lessons(id),
    progress_percent SMALLINT DEFAULT 0,
    status SMALLINT DEFAULT 0, -- 0 not_started, 1 in_progress, 2 completed
    last_position_seconds INT DEFAULT 0,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, user_id, lesson_id)
);
```

### 3.9 资源表 resources

```sql
CREATE TABLE resources (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    category_id UUID REFERENCES resource_categories(id),
    name VARCHAR(255) NOT NULL,
    resource_type VARCHAR(32) NOT NULL, -- image, video, document
    url VARCHAR(500) NOT NULL,
    size_bytes BIGINT DEFAULT 0,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 3.10 资源分类表 resource_categories

```sql
CREATE TABLE resource_categories (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES resource_categories(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## 4. 领域模型

### 4.1 BaseModel

所有模型共享以下字段：

```go
type BaseModel struct {
    ID        string `gorm:"primaryKey"`
    TenantID  string `gorm:"index;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (m *BaseModel) BeforeCreate(_ *gorm.DB) error {
    if m.ID == "" {
        m.ID = uuid.NewString()
    }
    return nil
}
```

### 4.2 Tenant

```go
type Tenant struct {
    BaseModel
    Code   string `gorm:"uniqueIndex;not null"`
    Name   string `gorm:"not null"`
    Status int    `gorm:"default:1"`
}
```

### 4.3 User

```go
type User struct {
    BaseModel
    Email    string `gorm:"not null"`
    Password string `gorm:"not null"`
    Name     string `gorm:"not null"`
    Role     string `gorm:"not null;default:'learner'"`
    Status   int    `gorm:"default:1"`
}
```

表级唯一索引：`uniqueIndex:idx_users_tenant_email`

### 4.4 Course

```go
type Course struct {
    BaseModel
    Title       string `gorm:"not null"`
    Description string
    CoverImage  string
    Status      int    `gorm:"default:0"` // 0 draft, 1 published
    CreatedBy   string `gorm:"not null"`
}
```

### 4.5 CourseChapter

```go
type CourseChapter struct {
    BaseModel
    CourseID  string `gorm:"index;not null"`
    Title     string `gorm:"not null"`
    SortOrder int    `gorm:"default:0"`
}
```

### 4.6 CourseLesson

```go
type CourseLesson struct {
    BaseModel
    ChapterID        string `gorm:"index;not null"`
    Title            string `gorm:"not null"`
    ContentType      string `gorm:"not null"` // video, document, text
    ContentURL       string
    DurationSeconds  int    `gorm:"default:0"`
    SortOrder        int    `gorm:"default:0"`
}
```

### 4.7 CourseEnrollment

```go
type CourseEnrollment struct {
    BaseModel
    CourseID string `gorm:"index;not null"`
    UserID   string `gorm:"index;not null"`
    Status   int    `gorm:"default:1"`
}
```

唯一索引：`uniqueIndex:idx_enrollments_tenant_course_user`

### 4.8 LessonProgress

```go
type LessonProgress struct {
    BaseModel
    UserID              string     `gorm:"index;not null"`
    LessonID            string     `gorm:"index;not null"`
    ProgressPercent     int        `gorm:"default:0"`
    Status              int        `gorm:"default:0"` // 0 not_started, 1 in_progress, 2 completed
    LastPositionSeconds int        `gorm:"default:0"`
    CompletedAt         *time.Time
}
```

唯一索引：`uniqueIndex:idx_progress_tenant_user_lesson`

### 4.9 Resource

```go
type Resource struct {
    BaseModel
    CategoryID   *string
    Name         string `gorm:"not null"`
    ResourceType string `gorm:"not null"` // image, video, document
    URL          string `gorm:"not null"`
    SizeBytes    int64  `gorm:"default:0"`
    CreatedBy    string `gorm:"not null"`
}
```

### 4.10 ResourceCategory

```go
type ResourceCategory struct {
    BaseModel
    Name     string  `gorm:"not null"`
    ParentID *string
}
```

---

## 5. 配置

### 5.1 环境变量

```env
SERVER_PORT=8080
APP_NAME=imaiplay
APP_VERSION=0.1.0

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=imaiplay
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25

JWT_SECRET=change-me-in-production
```

### 5.2 Config 结构

```go
type Config struct {
    ServerPort     string
    AppName        string
    AppVersion     string
    DBHost         string
    DBPort         int
    DBUser         string
    DBPassword     string
    DBName         string
    DBSSLMode      string
    DBMaxOpenConns int
    DBMaxIdleConns int
    JWTSecret      string
}
```

---

## 6. 认证与授权

### 6.1 角色

| 角色 | 说明 |
|------|------|
| superadmin | 平台超级管理员，管理租户 |
| tenant_admin | 租户管理员，管理租户内所有资源 |
| instructor | 讲师，可创建课程、查看学习数据 |
| learner | 学员，学习课程、查看个人进度 |

### 6.2 JWT Claims

```go
type Claims struct {
    UserID   string `json:"user_id"`
    TenantID string `json:"tenant_id"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}
```

### 6.3 认证流程

1. 用户访问统一登录 `/login`，或某个租户的 `/t/{tenantCode}/login`、自定义域名登录页。
2. 租户门户登录按当前租户验证账号；平台统一登录跨租户查找同一邮箱或手机号的候选账号。
3. 系统先验证密码和账号/租户可用状态；验证失败时不返回任何企业名称。
4. 只有一个匹配企业时直接签发 JWT；多个匹配企业时返回企业列表和五分钟有效的一次性选择凭证。
5. `/api/v1/auth/select-tenant` 原子消费选择凭证，签发包含不可变 `tenant_id` 的 JWT；凭证不能重放。
6. 后续请求携带 `Authorization: Bearer <token>`，Auth 中间件将用户信息写入 context。

### 6.4 会话边界

- 管理后台、PC 门户和 H5 门户分别使用独立的 access/refresh token 键，避免不同角色互相覆盖。
- 旧版通用 token 只在角色和有效期符合目标应用时迁移一次。
- access token 过期后由 refresh token 单飞刷新；刷新失败才清理当前应用会话并跳转登录。

### 6.5 权限控制（第一阶段简化版）

- `superadmin`：可访问平台级接口（租户管理）
- `tenant_admin`：可访问所有租户内管理接口
- `instructor`：可创建课程、管理自己创建的课程
- `learner`：只能访问学员端接口

具体权限校验在 Handler 层通过 role 判断，不实现复杂的 ACL。

---

## 7. API 设计

### 7.1 响应格式

成功：
```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

失败：
```json
{
  "code": 10001,
  "message": "error description",
  "data": null
}
```

### 7.2 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 40000 | 请求参数错误 |
| 40100 | 未认证 |
| 40300 | 无权限 |
| 40400 | 资源不存在 |
| 40900 | 资源冲突 |
| 50000 | 服务器内部错误 |

### 7.3 公共接口（无需认证）

```
GET  /health
GET  /health/db
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### 7.4 认证接口

#### POST /api/v1/auth/register

请求：
```json
{
  "email": "admin@acme.com",
  "password": "password123",
  "name": "Admin User",
  "role": "tenant_admin"
}
```

响应：
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "uuid",
    "tenant_id": "uuid",
    "email": "admin@acme.com",
    "name": "Admin User",
    "role": "tenant_admin",
    "status": 1,
    "created_at": "2026-07-27T10:00:00Z"
  }
}
```

#### POST /api/v1/auth/login

请求：
```json
{
  "email": "admin@acme.com",
  "password": "password123"
}
```

响应：
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-07-28T10:00:00Z"
  }
}
```

### 7.5 租户管理接口（superadmin）

```
POST   /backend/v1/tenants
GET    /backend/v1/tenants
GET    /backend/v1/tenants/:id
PUT    /backend/v1/tenants/:id
DELETE /backend/v1/tenants/:id
```

### 7.6 用户管理接口（tenant_admin）

```
POST   /backend/v1/users
GET    /backend/v1/users
GET    /backend/v1/users/:id
PUT    /backend/v1/users/:id
DELETE /backend/v1/users/:id
```

### 7.7 课程管理接口（tenant_admin / instructor）

```
POST   /backend/v1/courses
GET    /backend/v1/courses
GET    /backend/v1/courses/:id
PUT    /backend/v1/courses/:id
DELETE /backend/v1/courses/:id

POST   /backend/v1/courses/:id/chapters
GET    /backend/v1/courses/:id/chapters
PUT    /backend/v1/chapters/:id
DELETE /backend/v1/chapters/:id

POST   /backend/v1/chapters/:id/lessons
GET    /backend/v1/chapters/:id/lessons
PUT    /backend/v1/lessons/:id
DELETE /backend/v1/lessons/:id
```

### 7.8 课程指派接口

```
POST   /backend/v1/courses/:id/enrollments
GET    /backend/v1/courses/:id/enrollments
DELETE /backend/v1/enrollments/:id
```

### 7.9 学员端接口

```
GET    /api/v1/courses              # 我的课程列表
GET    /api/v1/courses/:id          # 课程详情
GET    /api/v1/courses/:id/progress # 课程整体进度
POST   /api/v1/lessons/:id/progress # 上报学习进度
GET    /api/v1/lessons/:id/progress # 获取课时进度
GET    /api/v1/recent-learning      # 最近学习
```

### 7.10 资源管理接口

```
POST   /backend/v1/resource-categories
GET    /backend/v1/resource-categories
PUT    /backend/v1/resource-categories/:id
DELETE /backend/v1/resource-categories/:id

POST   /backend/v1/resources
GET    /backend/v1/resources
GET    /backend/v1/resources/:id
DELETE /backend/v1/resources/:id
```

---

## 8. 模块设计

### 8.1 internal/context

```go
package context

const (
    SourceSubdomain  = "subdomain"
    SourceHeaderID   = "header_id"
    SourceHeaderCode = "header_code"
    SourceUnknown    = "unknown"
    UnknownTenant    = "unknown"
)

func WithTenant(ctx context.Context, code, source string) context.Context
func TenantFromContext(ctx context.Context) (code string, source string)

func WithUser(ctx context.Context, userID, tenantID, email, role string) context.Context
func UserFromContext(ctx context.Context) (userID, tenantID, email, role string, ok bool)
```

### 8.2 internal/middleware

```go
// tenant.go
func Tenant() gin.HandlerFunc

// auth.go
func Auth(jwtSecret string) gin.HandlerFunc

// require_role.go（可选）
func RequireRole(roles ...string) gin.HandlerFunc
```

### 8.3 internal/security

```go
// password.go
func HashPassword(password string) (string, error)
func CheckPassword(password, hash string) bool

// jwt.go
type Claims struct { ... }
func GenerateToken(userID, tenantID, email, role, secret string) (string, error)
func ValidateToken(tokenString, secret string) (*Claims, error)
```

### 8.4 internal/repository

```go
// tenant.go
type TenantRepository interface {
    Create(ctx context.Context, tenant *domain.Tenant) error
    FindByID(ctx context.Context, id string) (*domain.Tenant, error)
    FindByCode(ctx context.Context, code string) (*domain.Tenant, error)
    FindAll(ctx context.Context) ([]domain.Tenant, error)
    Update(ctx context.Context, tenant *domain.Tenant) error
    Delete(ctx context.Context, id string) error
}

// user.go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id string) (*domain.User, error)
    FindByEmailAndTenant(ctx context.Context, email, tenantID string) (*domain.User, error)
    FindByTenant(ctx context.Context, tenantID string) ([]domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
}

// course.go
type CourseRepository interface { ... }

// course_chapter.go
type CourseChapterRepository interface { ... }

// course_lesson.go
type CourseLessonRepository interface { ... }

// course_enrollment.go
type CourseEnrollmentRepository interface { ... }

// lesson_progress.go
type LessonProgressRepository interface { ... }

// resource.go
type ResourceRepository interface { ... }

// resource_category.go
type ResourceCategoryRepository interface { ... }
```

### 8.5 internal/service

```go
// auth.go
type AuthService interface {
    Register(ctx context.Context, tenantID, email, password, name, role string) (*domain.User, error)
    Login(ctx context.Context, tenantID, email, password string) (string, error)
}

// tenant.go
type TenantService interface { ... }

// user.go
type UserService interface { ... }

// course.go
type CourseService interface { ... }

// enrollment.go
type EnrollmentService interface { ... }

// progress.go
type ProgressService interface { ... }

// resource.go
type ResourceService interface { ... }
```

### 8.6 internal/api

```go
// auth.go
type AuthHandler struct { ... }
func (h *AuthHandler) Register(c *gin.Context)
func (h *AuthHandler) Login(c *gin.Context)

// tenant.go
type TenantHandler struct { ... }

// user.go
type UserHandler struct { ... }

// course.go
type CourseHandler struct { ... }

// enrollment.go
type EnrollmentHandler struct { ... }

// progress.go
type ProgressHandler struct { ... }

// resource.go
type ResourceHandler struct { ... }
```

### 8.7 internal/errorsx

```go
type AppError struct {
    Code    int
    Message string
}

func (e *AppError) Error() string

func BadRequest(message string) *AppError
func Unauthorized(message string) *AppError
func Forbidden(message string) *AppError
func NotFound(message string) *AppError
func Conflict(message string) *AppError
func Internal(message string) *AppError

func Response(c *gin.Context, err error)
```

---

## 9. 多租户数据隔离

### 9.1 门户与租户解析

- `play.imai.work` 是受保护的平台主域名，永远不绑定给租户。
- 默认门户为 `https://play.imai.work/t/{tenantCode}`；前端从路径解析租户编码并在 API 请求中发送 `X-Tenant-Code`。
- 自定义域名是可选别名。后端解析优先级为：已绑定自定义 Host、兼容子域名、`X-Tenant-Code`、可解析的 `X-Tenant-ID`。
- 公共 `GET /api/v1/portal` 按租户编码或 Host 返回名称、Logo、品牌色、欢迎语及两个门户地址，不返回私密配置。

### 9.2 Repository 层隔离

所有 Repository 查询必须带上 `tenant_id` 过滤：

```go
db.WithContext(ctx).Where("tenant_id = ?", tenantID)
```

### 9.3 Service 层获取租户

Service 从 context 获取 tenant code，再查询 tenant_id：

```go
tenantCode, _ := context.TenantFromContext(ctx)
tenant, err := tenantRepo.FindByCode(ctx, tenantCode)
if err != nil { ... }
```

### 9.4 JWT 租户一致性

业务请求以已签名 JWT 的 `tenant_id` 作为数据隔离依据，不接受客户端直接指定业务 `tenant_id`。TenantMatch 中间件会将当前门户解析出的租户与 JWT `tenant_id` 对比，不一致时返回 403；自定义域名与默认门户因为解析到同一个租户 ID，所以可共享同一租户数据。

### 9.5 超级管理员

`superadmin` 角色不绑定具体租户，context 中 tenantID 为空字符串。平台级接口跳过租户过滤。

### 9.6 账号模型决策

本阶段保留现有“每个租户一条 users 记录”的模型，不合并历史账号。统一登录通过凭证验证后的候选用户和一次性企业选择凭证完成切换；全局 `accounts + tenant_memberships` 模型延后，待观察多企业登录选择率和旧账号迁移成本后再评估。

---

## 10. 实现顺序

建议 Codex 按以下顺序实现，但可以根据完整设计文档一次性规划：

1. **统一错误响应** (`internal/errorsx`)
2. **JWT 与密码安全** (`internal/security`)
3. **用户上下文与认证中间件** (`internal/context/user.go`, `internal/middleware/auth.go`)
4. **用户模型与认证接口** (`domain/user`, `repository/user`, `service/auth`, `api/auth`)
5. **租户管理接口** (`domain/tenant` 扩展，service/tenant, api/tenant）
6. **用户管理接口** (`service/user`, `api/user`)
7. **课程/章节/课时管理** (`domain/course`, `domain/course_chapter`, `domain/course_lesson`)
8. **课程指派与学习进度**
9. **资源管理**
10. **统计看板**
11. **superadmin 初始化机制**
12. **API 文档与部署脚本**
13. **测试完善与代码优化**

---

## 11. 后续规划

### 任务 10：统计看板

- 租户级 Dashboard：学员总数、课程总数、已发布课程数、今日新增学员、今日学习人次
- 学习时长 Top10 学员、课程完成率排行
- 新增接口：`GET /backend/v1/dashboard`
- 管理后台首页数据展示

### 任务 11：superadmin 初始化机制

- 命令行初始化脚本 `cmd/init/main.go`
- 创建默认租户和首个 superadmin 用户
- 通过环境变量设置初始邮箱密码

### 任务 12：API 文档与部署脚本

- 接入 Swagger/OpenAPI（`swaggo/gin-swagger`）
- Dockerfile 多阶段构建
- Docker Compose（PostgreSQL + 后端 + Nginx）
- Nginx 统一入口配置
- README 部署说明

### 任务 13：测试完善与代码优化

- 补充集成测试
- 前端路由懒加载，解决 500KB 打包警告
- 统一前端依赖版本
- 性能优化（数据库索引、慢查询）

### 任务 14：可选扩展功能

- 学习路径/学习计划
- 考试/测验系统
- 证书发放
- 消息通知
- 第三方登录（企业微信、钉钉、飞书）
- 视频加密/防盗链
- 分片上传/断点续传

---

## 12. 测试要求

- 每个 `internal/security` 函数必须有单元测试
- 每个 `internal/repository` 实现必须有 SQLite 内存数据库测试
- 每个 `internal/service` 必须有单元测试（可 mock repository）
- 每个 `internal/api` handler 必须有 HTTP 测试
- `go test ./...` 必须全部通过
- `make build` 必须成功

---

## 13. 协作记录

Codex 每完成一个阶段，必须更新：
- `.codex/progress.md`
- `.codex/issues.md`
- `.codex/decisions.md`
- `.codex/knowledge-graph.md`
- `.codex/codex-log.md`

---

## 14. 注意事项

- 不要参考、复制 PlayEdu 代码
- 所有模型 ID 使用 UUID 字符串
- 所有时间字段使用 UTC
- 密码必须 bcrypt 哈希存储
- API 不返回密码字段
- Repository 查询必须带租户过滤
- 测试使用 SQLite 内存数据库
