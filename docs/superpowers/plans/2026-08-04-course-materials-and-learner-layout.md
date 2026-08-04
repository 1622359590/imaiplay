# 课程学习资料与学院端布局 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 PC 学员端公共容器整体偏右的问题，并让平台官方课程和学院自建课程能够维护多份课程级学习资料，供有课程访问权的 PC/H5 学员下载。

**Architecture:** 新增 `course_materials` 关联表和独立的课程资料仓储、服务与 API；文件实体继续复用 `resources` 与现有存储层，但增加受 200MB 限制的 `attachment` 上传路径。课程详情聚合有序资料元数据，下载时再次校验课程可见性与资料关联；PC 使用一套 1440px 自适应公共容器，H5 只增加紧凑资料列表。

**Tech Stack:** Go 1.x、Gin、GORM、PostgreSQL/SQLite migrations、React 18、TypeScript、Ant Design、Ant Design Mobile、Axios、Vitest/Node test、Vite。

## Global Constraints

- 学习资料属于课程，不绑定章节或课时，也不替代课时学习资源。
- PC 公共内容最大宽度为 1440px，顶栏和主体左右留白必须对称。
- PC 首页宽屏三列、常规桌面两列、窄屏一列；课程不足一行时从容器左侧排列。
- 课程详情信息顺序固定为：课程简介、学习资料、课程目录；无资料时不渲染资料区域。
- 首期附件只支持 PDF、DOC、DOCX、XLS、XLSX、PPT、PPTX、ZIP，单文件最大 200MB。
- 官方课程资料只能由 `superadmin` 管理；学院启用后可下载但不能编辑。
- 学院自建课程资料只能由本租户 `tenant_admin` 管理。
- 学员下载必须重新校验课程已发布、租户范围或官方课程启用状态，以及资料与课程的真实关联。
- 未授权或故意隐藏的资料下载返回 404，下载响应使用安全的 `Content-Disposition: attachment`。
- 不增加在线 Office 预览、资料版本历史、下载统计、学员上传或课时附件。
- 所有行为变更先写失败测试，确认预期失败后再写最小实现。

---

## File Structure

### Backend

- Create `internal/domain/course_material.go`: 课程与资源的有序关联模型。
- Create `internal/repository/course_material.go`: 资料关联仓储接口。
- Create `internal/repository/course_material_gorm.go`: 租户/官方课程可见性、排序和写入实现。
- Create `internal/repository/course_material_gorm_test.go`: 隔离、排序、重复关联和 CRUD 测试。
- Create `internal/service/course_material.go`: 管理权限、资料元数据和学员下载编排。
- Create `internal/service/course_material_test.go`: 普通课程、官方课程、跨租户与下载权限测试。
- Create `internal/api/course_material.go`: 管理、列表、更新、删除和下载处理器。
- Create `internal/api/course_material_test.go`: HTTP 状态、响应结构和附件响应头测试。
- Modify `internal/migration/migration.go`: 注册 v14 并创建资料表与唯一索引。
- Modify `internal/migration/migration_test.go`: 断言第 14 个迁移、表和索引存在且重复迁移幂等。
- Modify `internal/service/resource.go`: 增加附件格式探测、200MB 上传方法和 `attachment` 内容类型。
- Modify `internal/service/resource_test.go`: Office/ZIP 签名、伪造扩展、超限和存储扩展名测试。
- Modify `internal/repository/resource.go`: 增加资料引用检查所需接口。
- Modify `internal/repository/resource_gorm.go`: 平台资料访问和资源删除引用检查包含 `course_materials`。
- Modify `internal/repository/resource_gorm_test.go`: 官方资料访问、未启用租户和引用保护测试。
- Modify `internal/api/resource.go`: 增加租户/平台附件上传处理器方法及 200MB 请求上限。
- Modify `internal/api/resource_test.go`: 附件上传角色、大小和错误响应测试。
- Modify `internal/service/course.go`: `CourseDetail` 聚合 `Materials`，构造器接收资料仓储。
- Modify `internal/service/course_test.go`: 管理端和学员端详情资料返回测试。
- Modify `internal/repository/course_gorm.go`: 删除课程时先清理资料关联。
- Modify `internal/repository/course_gorm_test.go`: 普通和官方课程删除级联测试增加资料断言。
- Modify `internal/api/course.go`: 详情响应接口包含资料元数据。
- Modify `internal/server/server.go`: 注入服务并注册管理、上传和学员下载路由。
- Modify `cmd/server/main.go`: 创建课程资料仓储/服务并注入服务器依赖。
- Modify backend test fixtures that construct `CourseService` or `server.Dependencies`: 传入真实资料仓储/服务。

### Admin

- Create `web/admin/src/components/CourseMaterialsManager.tsx`: 多文件上传、重试、重命名、排序、替换和删除。
- Modify `web/admin/src/api/course.ts`: 资料类型和管理接口。
- Modify `web/admin/src/api/resource.ts`: 普通/平台附件上传接口。
- Modify `web/admin/src/pages/CourseDetail.tsx`: 在课程简介和目录之间挂载资料管理器。
- Modify `web/admin/src/styles.css`: 资料列表、上传队列和窄屏布局。
- Create `web/admin/tests/course-materials.test.ts`: API 路径、payload 和文件类型契约测试。

### PC learner

- Create `web/pc/src/components/CourseMaterials.tsx`: 有资料时渲染下载列表并处理 Blob 下载。
- Modify `web/pc/src/api/course.ts`: 映射 `materials` 并实现资料下载。
- Modify `web/pc/src/pages/CourseDetailPage.tsx`: 在简介与目录之间渲染资料列表。
- Modify `web/pc/src/components/CourseGrid.tsx`: 使用稳定的三/二/一列网格类。
- Modify `web/pc/src/styles.css`: 合并重复公共容器规则并实现 1440px 居中布局。
- Modify `web/pc/src/api/course.test.ts`: 资料映射与下载接口测试。

### H5 learner

- Create `web/h5/src/components/CourseMaterials.tsx`: 触控友好的资料列表与 Blob 下载。
- Modify `web/h5/src/api/course.ts`: 映射资料并实现下载。
- Modify `web/h5/src/types/course.ts`: 增加 `CourseMaterial` 和 `Course.materials`。
- Modify `web/h5/src/pages/CourseDetailPage.tsx`: 在简介和目录之间渲染资料。
- Modify `web/h5/src/styles.css`: 资料列表、长名称和 44px 下载触控区域。
- Modify `web/h5/src/api/course.test.ts`: 资料映射与下载请求测试。

---

### Task 1: 课程资料模型、迁移与仓储

**Files:**
- Create: `internal/domain/course_material.go`
- Create: `internal/repository/course_material.go`
- Create: `internal/repository/course_material_gorm.go`
- Create: `internal/repository/course_material_gorm_test.go`
- Modify: `internal/migration/migration.go`
- Modify: `internal/migration/migration_test.go`

**Interfaces:**
- Produces: `domain.CourseMaterial`, `repository.CourseMaterialRepository`。
- Produces methods: `Create`, `FindByID`, `FindByCourse`, `Update`, `Delete`, `DeleteByCourse`。

- [ ] **Step 1: Write failing migration and repository tests**

Add migration assertions:

```go
if !database.Migrator().HasTable(&domain.CourseMaterial{}) {
    t.Fatal("AutoMigrate() did not create course_materials table")
}
if !database.Migrator().HasIndex(&domain.CourseMaterial{}, "idx_course_materials_course_resource") {
    t.Fatal("missing course material uniqueness index")
}
if count != 14 {
    t.Fatalf("schema migrations count = %d, want 14", count)
}
```

Add a repository test that creates two tenant materials with sort orders `2` and `1`, an official material with empty `tenant_id`, and a foreign tenant material. Assert literal order `[second.ID, first.ID]`, cross-tenant `FindByID` returns `gorm.ErrRecordNotFound`, duplicate `(course_id, resource_id)` fails, and update/delete affect only the scoped row.

- [ ] **Step 2: Run tests and verify the expected failure**

Run:

```bash
go test ./internal/migration ./internal/repository -run 'TestAutoMigrate|TestCourseMaterialRepository' -count=1
```

Expected: FAIL because `CourseMaterial` and its repository/migration do not exist.

- [ ] **Step 3: Add the model and repository interface**

Create:

```go
type CourseMaterial struct {
    BaseModel
    CourseID   string   `gorm:"index;not null" json:"course_id"`
    ResourceID string   `gorm:"index;not null" json:"resource_id"`
    DisplayName string  `gorm:"not null" json:"display_name"`
    SortOrder  int      `gorm:"default:0" json:"sort_order"`
    CreatedBy  string   `gorm:"not null" json:"created_by"`
    Resource   Resource `gorm:"foreignKey:ResourceID" json:"resource"`
}
```

Define the repository contract with context-scoped reads. `FindByCourse` must order by `sort_order ASC, created_at ASC`; `Update` may change only `display_name`, `sort_order`, and `resource_id`.

- [ ] **Step 4: Register migration v14 and implement GORM scope**

Append `{Version: 14, Up: migrateV14}`. `migrateV14` must call `AutoMigrate(&domain.CourseMaterial{})` and create:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS idx_course_materials_course_resource
ON course_materials (course_id, resource_id)
```

The repository read scope must allow:

- tenant-owned rows when `material.tenant_id` matches context tenant;
- platform rows for `superadmin`;
- platform rows for a tenant only when the related official course is published and enabled in `tenant_official_courses`;
- learner reads only through published visible courses.

Write operations retain exact `tenant_id` scoping so tenant users cannot mutate platform rows.

- [ ] **Step 5: Run focused tests**

```bash
go test ./internal/migration ./internal/repository -run 'TestAutoMigrate|TestCourseMaterialRepository' -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/domain/course_material.go internal/repository/course_material.go internal/repository/course_material_gorm.go internal/repository/course_material_gorm_test.go internal/migration/migration.go internal/migration/migration_test.go
git commit -m "feat: add course material persistence"
```

### Task 2: 附件格式、上传限制与资源引用保护

**Files:**
- Modify: `internal/service/resource.go`
- Modify: `internal/service/resource_test.go`
- Modify: `internal/repository/resource.go`
- Modify: `internal/repository/resource_gorm.go`
- Modify: `internal/repository/resource_gorm_test.go`
- Modify: `internal/api/resource.go`
- Modify: `internal/api/resource_test.go`

**Interfaces:**
- Produces: `ResourceService.UploadAttachment(ctx, name, reader, size)`。
- Produces: `ResourceService.UploadPlatformAttachment(ctx, name, reader, size)`。
- Produces API handlers: `UploadAttachment`, `UploadPlatformAttachment`。
- Consumes: `domain.CourseMaterial` from Task 1 for reference checks.

- [ ] **Step 1: Write failing attachment validation tests**

Use real byte signatures and literal extensions:

```go
tests := []struct {
    name string
    body []byte
    suffix string
}{
    {"guide.pdf", []byte("%PDF-1.7\n"), ".pdf"},
    {"guide.doc", append([]byte{0xd0, 0xcf, 0x11, 0xe0}, make([]byte, 508)...), ".doc"},
    {"sheet.xlsx", append([]byte{'P', 'K', 3, 4}, make([]byte, 508)...), ".xlsx"},
    {"slides.pptx", append([]byte{'P', 'K', 3, 4}, make([]byte, 508)...), ".pptx"},
    {"bundle.zip", append([]byte{'P', 'K', 3, 4}, make([]byte, 508)...), ".zip"},
}
```

Assert every upload returns `ResourceType == "attachment"` and preserves the expected suffix. Add rejection cases for `malware.exe`, ZIP bytes named `guide.pdf`, OLE bytes named `guide.zip`, and size `200*1024*1024 + 1`.

Add repository tests proving a resource referenced by `course_materials` cannot be deleted and a platform attachment is readable only when its official course is enabled.

- [ ] **Step 2: Run tests and verify they fail**

```bash
go test ./internal/service ./internal/repository ./internal/api -run 'Attachment|CourseMaterialReference' -count=1
```

Expected: FAIL because attachment upload methods and material reference queries do not exist.

- [ ] **Step 3: Implement strict attachment detection**

Add:

```go
const maxAttachmentSize int64 = 200 * 1024 * 1024

var attachmentExtensions = map[string]string{
    ".pdf": "application/pdf",
    ".doc": "application/msword",
    ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    ".xls": "application/vnd.ms-excel",
    ".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    ".ppt": "application/vnd.ms-powerpoint",
    ".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    ".zip": "application/zip",
}
```

`detectAttachmentFormat(name, prefix)` must require both an allowed suffix and its matching family signature: PDF `%PDF`, ZIP-based formats `PK\x03\x04`, and legacy Office formats OLE `d0 cf 11 e0`. Store all as `resource_type = "attachment"`; use the sanitized original suffix in the storage key. Platform keys use `platform/attachments/<resource-id><suffix>`.

- [ ] **Step 4: Add attachment upload handlers**

Add methods that reuse the storage/quota path but pass `maxAttachmentSize` and strict attachment detection. Add routes later consumed by Task 4:

```text
POST /backend/v1/resources/attachments/upload
POST /backend/v1/admin/resources/attachments/upload
```

The first requires `tenant_admin`; the second requires `superadmin`. Set multipart request limit to `maxAttachmentSize + 1MiB` rather than the existing 1GiB limit.

- [ ] **Step 5: Include course materials in resource reference checks**

Extend platform reference detection with `EXISTS (SELECT 1 FROM course_materials WHERE resource_id = ?)`. Extend platform access with a second authorized path joining `course_materials`, published official courses, and enabled tenant courses. Regular tenant resource deletion must return conflict while a course material still references the resource.

- [ ] **Step 6: Run focused tests**

```bash
go test ./internal/service ./internal/repository ./internal/api -run 'Attachment|CourseMaterialReference' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/service/resource.go internal/service/resource_test.go internal/repository/resource.go internal/repository/resource_gorm.go internal/repository/resource_gorm_test.go internal/api/resource.go internal/api/resource_test.go
git commit -m "feat: support course material attachments"
```

### Task 3: 课程资料管理服务与 API

**Files:**
- Create: `internal/service/course_material.go`
- Create: `internal/service/course_material_test.go`
- Create: `internal/api/course_material.go`
- Create: `internal/api/course_material_test.go`
- Modify: `internal/server/server.go`
- Modify: `cmd/server/main.go`
- Modify: `internal/api/test_helpers_test.go`
- Modify: `internal/server/server_test.go`
- Modify: `internal/test/integration/core_flows_test.go`

**Interfaces:**
- Consumes: `CourseMaterialRepository` and `ResourceRepository`.
- Produces: `CourseMaterialService.Add`, `Update`, `Remove`, `ListForManager`。
- Produces request type: `{resource_id, display_name, sort_order}`。

- [ ] **Step 1: Write failing service permission tests**

Create real database fixtures for a tenant course, foreign tenant course, official course, tenant attachment, and platform attachment. Assert:

- tenant admin adds a tenant attachment to its own course;
- tenant admin cannot add to an official or foreign course;
- instructor and learner receive 403;
- superadmin adds a platform attachment to an official course;
- resource ownership/type mismatch returns 400 or 404 without creating a row;
- duplicate association returns 409;
- update trims `display_name`, validates non-empty text, and changes literal sort order;
- remove deletes only the association, not the resource record.

- [ ] **Step 2: Run service tests and verify failure**

```bash
go test ./internal/service -run TestCourseMaterialService -count=1
```

Expected: FAIL because `CourseMaterialService` does not exist.

- [ ] **Step 3: Implement the service boundary**

Define:

```go
type CourseMaterialInput struct {
    ResourceID  string `json:"resource_id"`
    DisplayName string `json:"display_name"`
    SortOrder   int    `json:"sort_order"`
}

func (s *CourseMaterialService) Add(ctx context.Context, courseID string, input CourseMaterialInput) (*domain.CourseMaterial, error)
func (s *CourseMaterialService) Update(ctx context.Context, courseID, materialID string, input CourseMaterialInput) (*domain.CourseMaterial, error)
func (s *CourseMaterialService) Remove(ctx context.Context, courseID, materialID string) error
func (s *CourseMaterialService) ListForManager(ctx context.Context, courseID string) ([]domain.CourseMaterial, error)
```

Use `CourseRepository.FindByID` to establish course scope, then enforce `superadmin + official` or `tenant_admin + non-official`. Validate resource ownership matches `course.TenantID`, `ResourceType == "attachment"`, and `display_name` is trimmed and non-empty.

- [ ] **Step 4: Write failing handler tests**

Register test routes and assert:

```text
GET    /backend/v1/courses/:id/materials
POST   /backend/v1/courses/:id/materials
PUT    /backend/v1/courses/:id/materials/:materialID
DELETE /backend/v1/courses/:id/materials/:materialID
```

Use actual JWT-backed handler fixtures where available. Assert 200 response envelopes for valid requests, 400 for missing display name, 403 for wrong role, 404 for cross-tenant IDs, and 409 for duplicate resource association.

- [ ] **Step 5: Implement handler, dependency injection and routes**

Add `CourseMaterialService api.CourseMaterialService` to `server.Dependencies`, create the handler in `registerRoutes`, and wire the repository/service in `cmd/server/main.go`. Update every test dependency constructor to provide the real service rather than `nil` so route tests fail for authorization, not panic.

- [ ] **Step 6: Run service, API and server tests**

```bash
go test ./internal/service ./internal/api ./internal/server ./internal/test/integration -run 'CourseMaterial|Route|Core' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/service/course_material.go internal/service/course_material_test.go internal/api/course_material.go internal/api/course_material_test.go internal/server/server.go cmd/server/main.go internal/api/test_helpers_test.go internal/server/server_test.go internal/test/integration/core_flows_test.go
git commit -m "feat: add course material management API"
```

### Task 4: 课程详情聚合、学员权限与安全下载

**Files:**
- Modify: `internal/service/course.go`
- Modify: `internal/service/course_test.go`
- Modify: `internal/api/course.go`
- Modify: `internal/service/course_material.go`
- Modify: `internal/service/course_material_test.go`
- Modify: `internal/api/course_material.go`
- Modify: `internal/api/course_material_test.go`
- Modify: `internal/repository/course_gorm.go`
- Modify: `internal/repository/course_gorm_test.go`
- Modify: `internal/server/server.go`
- Modify all `NewCourseService(...)` call sites found by `rg -n 'NewCourseService\('`.

**Interfaces:**
- Produces: `CourseDetail.Materials []domain.CourseMaterial`。
- Produces: `CourseMaterialService.OpenForLearner(ctx, materialID)` returning `(io.ReadCloser, contentType, fileName, error)`。
- Produces route: `GET /api/v1/course-materials/:id/download`。

- [ ] **Step 1: Write failing course-detail tests**

Extend course fixtures with two materials inserted in reverse creation/sort order. Assert manager and learner details return exactly two materials in sort order, a course without materials returns `[]` rather than `null`, a disabled official course remains 404, and a tenant cannot see foreign material metadata.

Update `CourseDetail` expected shape:

```go
type CourseDetail struct {
    Course    domain.Course          `json:"course"`
    Chapters  []CourseChapterDetail  `json:"chapters"`
    Materials []domain.CourseMaterial `json:"materials"`
}
```

- [ ] **Step 2: Run detail tests and verify failure**

```bash
go test ./internal/service ./internal/api -run 'Course.*Detail.*Material' -count=1
```

Expected: FAIL because details do not contain materials.

- [ ] **Step 3: Inject the material repository into course details**

Change `NewCourseService(courses, chapters, lessons, materials)` and update all call sites. In `detail`, load materials after the course has already passed manager or published visibility checks. Initialize `Materials` to an empty slice before returning.

- [ ] **Step 4: Write failing download authorization tests**

Use stored bytes and assert:

- learner in the owning tenant downloads a published tenant course material;
- learner in another tenant receives 404;
- enabled official course material downloads successfully;
- disabled official course returns 404;
- a resource ID not associated through `course_materials` cannot be downloaded through a guessed material ID;
- filename containing `../`, quotes, CR, or LF is sanitized;
- response header starts with `attachment;` and body equals the stored bytes.

- [ ] **Step 5: Implement learner download**

Add:

```go
func (s *CourseMaterialService) OpenForLearner(
    ctx context.Context, materialID string,
) (io.ReadCloser, string, string, error)
```

The method must load the scoped material, verify the related course through `FindPublishedByID`, then call the resource streaming boundary. The handler sets:

```go
c.Header("Content-Type", contentType)
c.Header("Content-Disposition", attachmentContentDisposition(fileName))
c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
```

Register only under the authenticated student group.

- [ ] **Step 6: Add course-delete cleanup tests and implementation**

Extend both ordinary and official course deletion tests with a `CourseMaterial`; assert its count is zero after deletion while the resource row remains. Delete material associations before deleting the course so foreign keys do not fail.

- [ ] **Step 7: Run focused backend tests**

```bash
go test ./internal/repository ./internal/service ./internal/api ./internal/server -run 'CourseMaterial|Course.*Delete|Course.*Detail' -count=1
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/service/course.go internal/service/course_test.go internal/api/course.go internal/service/course_material.go internal/service/course_material_test.go internal/api/course_material.go internal/api/course_material_test.go internal/repository/course_gorm.go internal/repository/course_gorm_test.go internal/server/server.go cmd/server/main.go internal/api/test_helpers_test.go internal/server/server_test.go internal/test/integration/core_flows_test.go
git commit -m "feat: expose secure learner course materials"
```

### Task 5: 管理后台课程资料管理器

**Files:**
- Create: `web/admin/src/components/CourseMaterialsManager.tsx`
- Create: `web/admin/tests/course-materials.test.ts`
- Modify: `web/admin/src/api/course.ts`
- Modify: `web/admin/src/api/resource.ts`
- Modify: `web/admin/src/pages/CourseDetail.tsx`
- Modify: `web/admin/src/styles.css`

**Interfaces:**
- Consumes backend routes from Tasks 2–4.
- Produces TypeScript type `CourseMaterial` and component props `{courseId, officialMode, initialMaterials, onChange}`.

- [ ] **Step 1: Write failing API contract tests**

Test the exact URLs and request bodies using the existing Node test interception pattern:

```ts
await courseApi.addMaterial('course-1', {
  resource_id: 'resource-1',
  display_name: '入门手册.pdf',
  sort_order: 2,
})
```

Assert attachment uploads call `/backend/v1/resources/attachments/upload` or `/backend/v1/admin/resources/attachments/upload`, and material update/delete include both course and material IDs.

- [ ] **Step 2: Run admin tests and verify failure**

```bash
node --test web/admin/tests/course-materials.test.ts
```

Expected: FAIL because the API methods do not exist.

- [ ] **Step 3: Add API types and methods**

Define:

```ts
export interface CourseMaterial {
  id: string
  course_id: string
  resource_id: string
  display_name: string
  sort_order: number
  resource: Resource
}
```

Add `listMaterials`, `addMaterial`, `updateMaterial`, and `removeMaterial`; add `uploadAttachment` and `uploadPlatformAttachment` with `timeout: 0` and per-file upload progress.

- [ ] **Step 4: Build the material manager**

The component must:

- accept multiple files through Ant Design `Upload`;
- filter to `.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.zip` and 200MB before upload;
- maintain one queue row per file with `waiting/uploading/success/error` state;
- upload a resource, then call `addMaterial` with the original filename;
- keep a failed row and expose a single-file retry action;
- use inline edit for `display_name`;
- reorder by exchanging adjacent `sort_order` values through two update requests;
- replace by uploading a new resource and updating `resource_id`;
- remove only the association after confirmation.

Do not reuse `MediaUploader`, because it models a single image/video/PDF value and cannot represent a multi-file queue.

- [ ] **Step 5: Mount it in the course content page**

Render `CourseMaterialsManager` after `.course-summary` and before the “课程目录” heading. Pass `officialMode` so superadmin uses platform attachment upload; tenant course uses tenant upload. Tenant admins viewing the separate official-course selector never enter this editable content page, so no false edit controls are shown.

- [ ] **Step 6: Add compact admin styling and build**

Use one card, file rows, restrained coral primary actions, 8–12px gaps, and wrap actions below metadata under 760px. Run:

```bash
npm --prefix web/admin test
npm --prefix web/admin run build
```

Expected: all tests and TypeScript/Vite build PASS.

- [ ] **Step 7: Commit**

```bash
git add web/admin/src/components/CourseMaterialsManager.tsx web/admin/tests/course-materials.test.ts web/admin/src/api/course.ts web/admin/src/api/resource.ts web/admin/src/pages/CourseDetail.tsx web/admin/src/styles.css
git commit -m "feat: manage materials inside courses"
```

### Task 6: PC 学员端公共布局与课程资料下载

**Files:**
- Create: `web/pc/src/components/CourseMaterials.tsx`
- Modify: `web/pc/src/api/course.ts`
- Modify: `web/pc/src/api/course.test.ts`
- Modify: `web/pc/src/pages/CourseDetailPage.tsx`
- Modify: `web/pc/src/components/CourseGrid.tsx`
- Modify: `web/pc/src/styles.css`

**Interfaces:**
- Consumes: `materials` on course detail and `GET /api/v1/course-materials/:id/download`.
- Produces: shared CSS classes `.learner-container`, `.course-materials`, `.course-material-row`.

- [ ] **Step 1: Write failing PC mapping/download tests**

Add a raw detail fixture with one material and assert the mapped value uses literal fields:

```ts
expect(course.materials).toEqual([{
  id: 'material-1',
  displayName: '入门手册.pdf',
  sizeBytes: 4096,
  resourceType: 'attachment',
}])
```

Mock the client and assert `downloadCourseMaterial('material-1')` sends a GET to `/api/v1/course-materials/material-1/download` with `responseType: 'blob'`.

- [ ] **Step 2: Run PC tests and verify failure**

```bash
npm --prefix web/pc test
```

Expected: FAIL because materials mapping/download are absent.

- [ ] **Step 3: Implement the learner material component**

Render nothing for an empty array. For each item show an icon plus extension, display name, formatted size, and a text `下载` button. On click, disable that row, fetch the Blob, create an object URL, click a temporary `<a download={displayName}>`, then revoke the URL in `finally`. Use Ant Design `message.error('资料下载失败，请稍后重试')` without leaving the course page.

- [ ] **Step 4: Insert materials into course detail**

Render `<CourseMaterials materials={course.materials} />` after `.detail-hero` and before `.chapter-card`. Preserve the existing course directory behavior.

- [ ] **Step 5: Consolidate the public layout CSS**

Replace the repeated final definitions of `.top-header`, `.main-content`, `.page-section`, `.detail-hero`, and `.chapter-card` with one final contract:

```css
.learner-container,
.main-content {
  width: min(1440px, calc(100% - 48px));
  margin-inline: auto;
}

.top-header {
  padding-inline: max(24px, calc((100vw - 1440px) / 2));
}

.course-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px;
}
```

At `max-width: 1599px`, use two columns; at `max-width: 680px`, use one column and 14px side margins. This keeps the agreed 2048px viewport at three columns and the 1440px viewport at two columns. Ensure `.player-page` remains centered inside `.main-content` with its existing readable maximum width. Remove obsolete duplicated width declarations rather than adding another override at the end.

- [ ] **Step 6: Run PC tests and build**

```bash
npm --prefix web/pc test
npm --prefix web/pc run build
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add web/pc/src/components/CourseMaterials.tsx web/pc/src/api/course.ts web/pc/src/api/course.test.ts web/pc/src/pages/CourseDetailPage.tsx web/pc/src/components/CourseGrid.tsx web/pc/src/styles.css
git commit -m "feat: add learner materials and center pc layout"
```

### Task 7: H5 课程资料下载

**Files:**
- Create: `web/h5/src/components/CourseMaterials.tsx`
- Modify: `web/h5/src/types/course.ts`
- Modify: `web/h5/src/api/course.ts`
- Modify: `web/h5/src/api/course.test.ts`
- Modify: `web/h5/src/pages/CourseDetailPage.tsx`
- Modify: `web/h5/src/styles.css`

**Interfaces:**
- Consumes: the same learner detail and download routes as PC.
- Produces: H5 `CourseMaterial` type and `.mobile-course-materials` UI.

- [ ] **Step 1: Write failing H5 API tests**

Use the same literal material fixture as PC, mapped to H5 naming. Assert an absent backend `materials` field becomes `[]`, not `undefined`, and download requests use `responseType: 'blob'`.

- [ ] **Step 2: Run H5 tests and verify failure**

```bash
npm --prefix web/h5 test
```

Expected: FAIL because H5 types and mapping lack materials.

- [ ] **Step 3: Implement the H5 materials component**

Place the component after `.detail-summary` and before `.chapter-section`. Each row must have at least 44px height, a file icon/extension, a two-line clamped name, size text, and a `下载` button. Use `Toast.show` for failure and prevent duplicate taps while one row is downloading.

- [ ] **Step 4: Add H5 styles and build**

Use the existing white card, `var(--learner-line)` border, 10px radius, and coral focus/action color. Do not reintroduce gradients, statistics, or bottom navigation. Run:

```bash
npm --prefix web/h5 test
npm --prefix web/h5 run build
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/h5/src/components/CourseMaterials.tsx web/h5/src/types/course.ts web/h5/src/api/course.ts web/h5/src/api/course.test.ts web/h5/src/pages/CourseDetailPage.tsx web/h5/src/styles.css
git commit -m "feat: add h5 course material downloads"
```

### Task 8: 全量回归、视觉验证与交付

**Files:**
- Modify only files required by failures found in this task.

**Interfaces:**
- Consumes all previous task outputs.
- Produces a verified, merge-ready feature branch.

- [ ] **Step 1: Run full backend verification**

```bash
gofmt -w internal/domain/course_material.go internal/repository/course_material.go internal/repository/course_material_gorm.go internal/service/course_material.go internal/api/course_material.go
go test ./...
go build ./...
```

Expected: all packages PASS and build exits 0.

- [ ] **Step 2: Run all frontend verification**

```bash
npm --prefix web/admin test
npm --prefix web/admin run build
npm --prefix web/pc test
npm --prefix web/pc run build
npm --prefix web/h5 test
npm --prefix web/h5 run build
```

Expected: all tests and three production builds PASS.

- [ ] **Step 3: Verify responsive layout with real pages**

Use the in-app browser against the local application and capture these states after loading real course data:

- PC 2048×1080: header and content left/right whitespace differ by no more than 2px; course grid is three columns.
- PC 1440×900: course grid is two columns and content has 24px minimum gutters.
- PC 768×900: no horizontal scrollbar; course grid is one column.
- Course detail with materials: intro, materials, directory order is correct.
- Course detail without materials: no empty material card or extra gap.
- H5 390×844: names clamp to two lines and each download target is at least 44px high.

Inspect the reference screenshots and implementation screenshots side by side. Reject screenshots showing blank/loading states, shifted content, cropped cards, low-contrast text, or broken spacing.

- [ ] **Step 4: Verify authorization through HTTP integration**

Run focused integration requests for tenant admin, superadmin, enabled learner, disabled tenant learner, and foreign tenant learner. Confirm only the intended roles can manage or download and that denied downloads return 404 without resource metadata.

- [ ] **Step 5: Check the final diff and commit verification fixes**

```bash
git diff --check
git status --short
```

If verification required code changes, rerun the affected test/build command and commit only those files:

```bash
git commit -m "fix: complete course material verification"
```

Before that commit, stage each changed path explicitly from `git status --short`; do not use `git add .` or stage unrelated files. If a verification fix belongs to an earlier task, prefer amending that task's explicit file list and rerunning its focused verification.

If no files changed, do not create an empty commit.

## Final Acceptance Checklist

- [ ] PC 首页、课程详情和课时页使用同一 1440px 居中容器，无固定左侧占位。
- [ ] PC 三/二/一列断点与卡片左对齐通过真实视口检查。
- [ ] 课程详情只在有资料时显示资料区，位置在简介和目录之间。
- [ ] tenant admin 可维护本租户自建课程资料，不能维护官方或跨租户资料。
- [ ] superadmin 可维护官方课程资料。
- [ ] 已启用官方课程和已发布租户课程的学员可下载资料。
- [ ] 未启用、未发布、跨租户和猜测 ID 的下载均被拒绝。
- [ ] PDF、Word、Excel、PowerPoint、ZIP 的签名、扩展名和 200MB 限制均有测试。
- [ ] 删除课程清理资料关联但保留资源；删除被引用资源返回冲突。
- [ ] 后端全量测试/构建和三个前端测试/构建全部通过。
- [ ] 最终工作树只包含本功能相关变更。
