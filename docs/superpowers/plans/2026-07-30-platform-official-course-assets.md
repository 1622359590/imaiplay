# Platform Official Course Assets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give superadmins complete official-course management with direct platform-owned media uploads, protected lesson delivery, and polished upload controls wherever the three web applications currently accept uploaded images or videos.

**Architecture:** Platform resources remain rows in the existing resources table with an empty `tenant_id`, while repository methods explicitly separate them from tenant resources. The existing authenticated resource stream endpoint handles both tenant files and authorized official-course files; a separate public endpoint serves platform cover images only. The admin frontend uses one reusable uploader for official covers, lesson media, tenant course covers, tenant logos, and resource-library uploads.

**Tech Stack:** Go, Gin, GORM, PostgreSQL/SQLite tests, React, TypeScript, Ant Design, Axios, Vite.

## Global Constraints

- Do not add a second resource table or change the resource ownership schema.
- Platform resources use `tenant_id = ""`; tenant resources keep a non-empty tenant ID.
- Platform storage keys start with `platform/images/`, `platform/videos/`, or `platform/documents/`.
- Official-course videos and PDFs require authentication and authorization on every read.
- Learners must be enrolled in an enabled official course; tenant administrators and instructors may preview an enabled official course.
- Platform covers are public, but the public endpoint serves platform image resources only.
- Existing image and video upload inputs in `web/admin`, `web/pc`, and `web/h5` must not remain URL text inputs.
- Ordinary external hyperlink fields remain text inputs.
- Preserve existing tenant course and tenant resource behavior.
- Do not perform unrelated refactors.

---

## File Structure

### Backend

- `internal/repository/resource.go`: explicit platform-resource and access-query interfaces.
- `internal/repository/resource_gorm.go`: platform-only queries, reference checks, and official-course access query.
- `internal/repository/resource_gorm_test.go`: platform isolation, reference, and authorization tests.
- `internal/service/resource.go`: platform upload/list/delete and protected/public stream decisions.
- `internal/service/resource_test.go`: platform upload, storage-prefix, deletion, and stream tests.
- `internal/service/resource_superadmin_test.go`: replace cross-tenant expectations with platform-only expectations.
- `internal/service/course.go`: status-aware official-course creation.
- `internal/service/course_lesson.go`: ensure official lessons reference platform resources and tenant lessons reference tenant resources.
- `internal/service/course_test.go`: official-course status and deletion behavior.
- `internal/service/course_lesson_test.go`: resource-ownership validation.
- `internal/repository/course_gorm.go`: clean tenant activation and enrollment rows when an official course is deleted.
- `internal/repository/course_gorm_test.go`: official-course cleanup regression test.
- `internal/api/resource.go`: superadmin platform handlers and public cover handler.
- `internal/api/resource_test.go`: role, upload, list, delete, cover, and protected file responses.
- `internal/api/course.go`: accept official-course status.
- `internal/api/course_superadmin_test.go`: complete official-course lifecycle.
- `internal/server/server.go`: register platform upload/list/delete, public cover, and protected file routes.
- `internal/server/server_test.go`: route authorization coverage.
- `cmd/server/main.go`: pass the resource repository into the lesson service.
- `internal/api/test_helpers_test.go`, `internal/test/integration/core_flows_test.go`: update constructors and fixtures.

### Frontend

- `web/admin/src/components/MediaUploader.tsx`: reusable uploader with progress, preview, replace, remove, and retry.
- `web/admin/src/api/resource.ts`: tenant/platform upload APIs with progress callbacks and platform preview.
- `web/admin/src/api/officialCourse.ts`: complete official-course CRUD API.
- `web/admin/src/pages/OfficialCourses.tsx`: full official-course management list and form.
- `web/admin/src/pages/CourseDetail.tsx`: official/tenant mode and direct lesson upload.
- `web/admin/src/pages/Courses.tsx`: tenant course cover uploader.
- `web/admin/src/pages/ThemeSettings.tsx`: tenant logo uploader without URL input.
- `web/admin/src/pages/Resources.tsx`: polished uploader state using the shared component.
- `web/admin/src/layout/AdminLayout.tsx`: remove duplicate superadmin course menu.
- `web/admin/src/routes.tsx`: add a superadmin official-course detail route and protect tenant-only course routes.
- `web/admin/src/styles.css`: responsive uploader and official-course presentation styles.
- `web/pc/src` and `web/h5/src`: no upload form exists at the current revision; verify their media reads and builds against the protected resource endpoint.
- `.codex/codex-log.md`: append the completed-task summary, key changes, review feedback, and next recommendation.

---

### Task 1: Isolate Platform Resources in the Repository

**Files:**
- Modify: `internal/repository/resource.go`
- Modify: `internal/repository/resource_gorm.go`
- Create or modify: `internal/repository/resource_gorm_test.go`
- Delete: `internal/repository/resource_find_all_test.go`

**Interfaces:**
- Produces:
  ```go
  FindPlatformByID(ctx context.Context, id string) (*domain.Resource, error)
  FindPlatform(ctx context.Context, offset, limit int) ([]domain.Resource, int64, error)
  DeletePlatform(ctx context.Context, id string) error
  IsPlatformReferenced(ctx context.Context, id string, coverURLs []string) (bool, error)
  CanAccessPlatformResource(
      ctx context.Context, resourceID, tenantID, userID, role string,
  ) (bool, error)
  ```
- Consumes: existing `domain.Resource`, `domain.Course`, `domain.CourseChapter`, `domain.CourseLesson`, `domain.CourseEnrollment`, and `domain.TenantOfficialCourse`.

- [ ] **Step 1: Write failing platform-isolation tests**

  Add tests that create one platform resource and two tenant resources, then assert:

  ```go
  items, total, err := repo.FindPlatform(context.Background(), 0, 20)
  if err != nil || total != 1 || len(items) != 1 || items[0].TenantID != "" {
      t.Fatalf("FindPlatform() = %#v, %d, %v", items, total, err)
  }
  if _, err := repo.FindPlatformByID(context.Background(), tenantResource.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
      t.Fatalf("FindPlatformByID(tenant resource) error = %v", err)
  }
  ```

  Add table cases for `CanAccessPlatformResource`: superadmin succeeds; tenant admin/instructor succeeds only when the official course is enabled; learner additionally needs an active enrollment; disabled tenant and unrelated resource fail.

- [ ] **Step 2: Run the repository tests and confirm failure**

  Run:

  ```bash
  go test ./internal/repository -run 'TestResourceRepository(Platform|Access|Reference)' -count=1
  ```

  Expected: build failure because the new repository methods do not exist.

- [ ] **Step 3: Implement explicit platform queries**

  Replace the cross-tenant `FindAll` method with:

  ```go
  func (repo *resourceGORMRepository) FindPlatform(
      ctx context.Context, offset, limit int,
  ) ([]domain.Resource, int64, error) {
      query := repo.database.WithContext(ctx).Model(&domain.Resource{}).
          Where("tenant_id = ?", "")
      return repo.find(ctx, query, offset, limit)
  }
  ```

  Implement `FindPlatformByID` and `DeletePlatform` with both `id = ?` and
  `tenant_id = ''`. Implement `IsPlatformReferenced` by checking
  `course_lessons.resource_id` and `courses.cover_image IN ?`.

  Implement `CanAccessPlatformResource` as one existence query joining lessons,
  chapters, official courses, and `tenant_official_courses`. For learners,
  require an active `course_enrollments` row for the same tenant, course, and
  user. For `tenant_admin` and `instructor`, require only an enabled official
  course. Return true immediately for `superadmin`.

- [ ] **Step 4: Run repository tests**

  Run:

  ```bash
  go test ./internal/repository -run 'TestResourceRepository(Platform|Access|Reference)' -count=1
  ```

  Expected: PASS.

- [ ] **Step 5: Commit the repository boundary**

  ```bash
  git add internal/repository/resource.go internal/repository/resource_gorm.go internal/repository/resource_gorm_test.go internal/repository/resource_find_all_test.go
  git commit -m "feat: isolate platform resources from tenant resources"
  ```

---

### Task 2: Add Platform Upload and Protected Streaming Services

**Files:**
- Modify: `internal/service/resource.go`
- Modify: `internal/service/resource_test.go`
- Modify: `internal/service/resource_superadmin_test.go`

**Interfaces:**
- Consumes: Task 1 repository methods.
- Produces:
  ```go
  UploadPlatform(ctx context.Context, name string, reader io.Reader, size int64) (*domain.Resource, error)
  ListPlatform(ctx context.Context, offset, limit int) ([]domain.Resource, int64, error)
  DeletePlatform(ctx context.Context, id string) error
  OpenPlatformCover(ctx context.Context, id string) (io.ReadCloser, string, string, error)
  Open(ctx context.Context, id string) (io.ReadCloser, string, string, error)
  ```

- [ ] **Step 1: Write failing service tests**

  Add tests with real local storage that assert:

  ```go
  uploaded, err := service.UploadPlatform(
      courseContext("root", "", "superadmin"),
      "cover.png", bytes.NewReader(pngBytes), int64(len(pngBytes)),
  )
  if err != nil || uploaded.TenantID != "" || uploaded.ResourceType != "image" {
      t.Fatalf("UploadPlatform() = %#v, %v", uploaded, err)
  }
  if uploaded.URL != "/api/v1/platform-covers/"+uploaded.ID {
      t.Fatalf("public cover URL = %q", uploaded.URL)
  }
  ```

  Inspect the local storage tree and require `platform/images/`. Add analogous
  video and PDF cases for `platform/videos/` and `platform/documents/`.
  Assert tenant admin cannot call `UploadPlatform` or `ListPlatform`.

  Add stream cases for superadmin, enabled tenant admin, enabled instructor,
  enrolled learner, unenrolled learner, and disabled tenant. Add a public-cover
  case that rejects a video resource. Add a delete case returning code 40900
  when `IsPlatformReferenced` is true.

- [ ] **Step 2: Run failing service tests**

  Run:

  ```bash
  go test ./internal/service -run 'TestResourceService(Platform|Official|PublicCover)' -count=1
  ```

  Expected: build failure because platform methods do not exist.

- [ ] **Step 3: Extract one validated upload helper**

  Keep `Upload` as the tenant entry point and introduce an internal helper:

  ```go
  func (service *ResourceService) upload(
      ctx context.Context,
      tenantID, userID, name string,
      reader io.Reader,
      size int64,
      platform bool,
  ) (*domain.Resource, error)
  ```

  Reuse the existing content sniffing and 500 MiB maximum. Tenant uploads keep
  quota checks and `<tenantID>/<uuid><ext>`. Platform uploads skip tenant quota
  and select the prefix from the detected type:

  ```go
  prefixes := map[string]string{
      "image": "platform/images",
      "video": "platform/videos",
      "document": "platform/documents",
  }
  ```

  After creating a platform image, return
  `/api/v1/platform-covers/<resource-id>` as its response URL without replacing
  the storage URL persisted in the database.

- [ ] **Step 4: Implement platform list, delete, and read authorization**

  `ListPlatform` returns platform rows only and maps platform image URLs to the
  public cover route. `DeletePlatform` calculates both the public cover URL and
  persisted storage URL, checks references, returns
  `errorsx.Conflict("resource is in use")`, and deletes the database row and
  object with the same rollback behavior as tenant deletion.

  Extend `Open`:

  1. Try the caller's tenant-scoped resource.
  2. If not found, load a platform resource explicitly.
  3. Ask `CanAccessPlatformResource` using user ID, tenant ID, and role.
  4. Stream the object only when authorized.

  `OpenPlatformCover` loads platform rows only, requires
  `ResourceType == "image"`, and streams without user context.

- [ ] **Step 5: Run service tests**

  Run:

  ```bash
  go test ./internal/service -run 'TestResourceService(Platform|Official|PublicCover)' -count=1
  ```

  Expected: PASS.

- [ ] **Step 6: Commit platform resource services**

  ```bash
  git add internal/service/resource.go internal/service/resource_test.go internal/service/resource_superadmin_test.go
  git commit -m "feat: add platform resource upload and access controls"
  ```

---

### Task 3: Expose Platform Resource APIs

**Files:**
- Modify: `internal/api/resource.go`
- Modify: `internal/api/resource_test.go`
- Modify: `internal/server/server.go`
- Modify: `internal/server/server_test.go`

**Interfaces:**
- Consumes: Task 2 service methods.
- Produces routes:
  - `POST /backend/v1/admin/resources/upload`
  - `GET /backend/v1/admin/resources`
  - `DELETE /backend/v1/admin/resources/:id`
  - `GET /backend/v1/admin/resources/:id/file`
  - `GET /api/v1/platform-covers/:id`
  - existing authenticated `GET /api/v1/resources/:id/file` extended to platform resources.

- [ ] **Step 1: Extend handler tests first**

  Add a `platformResourceStub` with counters for upload, list, delete, protected
  open, and cover open. Test:

  ```go
  router.POST("/admin/resources/upload", handler.UploadPlatform)
  router.GET("/admin/resources", handler.ListPlatform)
  router.DELETE("/admin/resources/:id", handler.DeletePlatform)
  router.GET("/platform-covers/:id", handler.PlatformCover)
  ```

  Require HTTP 200 for superadmin, HTTP 403 for tenant admin on platform
  mutations, HTTP 409 for a referenced delete, and image bytes plus
  `Content-Type: image/png` from the cover route.

- [ ] **Step 2: Run handler tests and confirm failure**

  Run:

  ```bash
  go test ./internal/api -run 'TestResourceHandlerPlatform' -count=1
  ```

  Expected: build failure because the handler methods are absent.

- [ ] **Step 3: Implement handlers and interface**

  Add `UploadPlatform`, `ListPlatform`, `DeletePlatform`, and `PlatformCover`.
  Share multipart parsing between tenant and platform upload handlers so both
  retain `http.MaxBytesReader` and the existing localized invalid-file error.
  Keep role checks explicit:

  ```go
  if !requireHandlerRole(c, "superadmin") {
      return
  }
  ```

  `PlatformCover` has no role check and calls only `OpenPlatformCover`.

- [ ] **Step 4: Register routes in the correct middleware groups**

  Register the public cover route on the root router before protected groups:

  ```go
  router.GET("/api/v1/platform-covers/:id", resourceHandler.PlatformCover)
  ```

  Register platform management and preview under the authenticated backend
  group. Keep the student file route under tenant, JWT, and tenant-access
  middleware.

- [ ] **Step 5: Run API and server tests**

  Run:

  ```bash
  go test ./internal/api ./internal/server -run 'Resource|BackendRoutes' -count=1
  ```

  Expected: PASS.

- [ ] **Step 6: Commit platform HTTP APIs**

  ```bash
  git add internal/api/resource.go internal/api/resource_test.go internal/server/server.go internal/server/server_test.go
  git commit -m "feat: expose platform resource management APIs"
  ```

---

### Task 4: Complete Official Course Domain Rules

**Files:**
- Modify: `internal/api/course.go`
- Modify: `internal/api/course_superadmin_test.go`
- Modify: `internal/service/course.go`
- Modify: `internal/service/course_test.go`
- Modify: `internal/service/course_lesson.go`
- Modify: `internal/service/course_lesson_test.go`
- Modify: `internal/repository/course_gorm.go`
- Modify: `internal/repository/course_gorm_test.go`
- Modify: `cmd/server/main.go`
- Modify: `internal/api/test_helpers_test.go`
- Modify: `internal/server/server_test.go`
- Modify: `internal/test/integration/core_flows_test.go`

**Interfaces:**
- Consumes: Task 1 `ResourceRepository.FindPlatformByID` and existing tenant-scoped `FindByID`.
- Produces:
  ```go
  CreateOfficial(
      ctx context.Context,
      title, description, coverImage string,
      status int,
  ) (*domain.Course, error)

  NewCourseLessonService(
      lessons repository.CourseLessonRepository,
      chapters repository.CourseChapterRepository,
      courses repository.CourseRepository,
      resources repository.ResourceRepository,
  ) *CourseLessonService
  ```

- [ ] **Step 1: Add failing official-course tests**

  Extend the superadmin API lifecycle to create a draft official course with
  `"status":0`, update it to status 1, add a chapter, add a lesson backed by a
  platform resource, retrieve detail, and delete it.

  Add service cases:

  ```go
  _, err := fixture.lessons.CreateWithResource(
      superadminCtx, officialChapter.ID, "Intro", "video",
      tenantResource.ID, "", 60, 1,
  )
  if errorCode(err) != 40000 {
      t.Fatalf("official lesson with tenant resource error = %#v", err)
  }
  ```

  Add the inverse case for a tenant lesson using a platform resource. Add a
  repository test proving official deletion removes
  `tenant_official_courses`, enrollments for all tenants, lesson progress,
  chapters, and lessons.

- [ ] **Step 2: Run focused tests and confirm failure**

  Run:

  ```bash
  go test ./internal/api ./internal/service ./internal/repository -run 'OfficialCourse|OfficialLesson' -count=1
  ```

  Expected: FAIL because create status and resource ownership are not enforced.

- [ ] **Step 3: Make official creation status-aware**

  Add `status` to the request struct and validate it exactly as update does:

  ```go
  if status != 0 && status != 1 {
      return nil, errorsx.BadRequest("invalid course status")
  }
  ```

  Keep `IsOfficial: true` and `TenantID: ""`. Public registration and tenant
  course creation remain unchanged.

- [ ] **Step 4: Validate lesson resource ownership**

  Pass `ResourceRepository` into `CourseLessonService`. In
  `CreateWithResource` and `UpdateWithResource`, load the course through the
  chapter and require:

  - official course plus non-empty resource ID: `FindPlatformByID` succeeds;
  - tenant course plus non-empty resource ID: tenant-scoped `FindByID` succeeds;
  - text lesson: resource ID is empty.

  Return `errorsx.BadRequest("resource does not belong to this course")` on a
  mismatch. Update every constructor call listed in this task.

- [ ] **Step 5: Clean all official-course references on deletion**

  In the existing transaction, branch on `course.IsOfficial`. For official
  courses, delete progress and enrollments by `course_id`/official lesson IDs
  across tenant IDs, delete tenant activation rows, then delete platform-owned
  lessons, chapters, and the course. Keep the existing tenant-scoped branch
  unchanged.

- [ ] **Step 6: Run focused tests**

  Run:

  ```bash
  go test ./internal/api ./internal/service ./internal/repository -run 'OfficialCourse|OfficialLesson' -count=1
  ```

  Expected: PASS.

- [ ] **Step 7: Commit official-course domain changes**

  ```bash
  git add internal/api/course.go internal/api/course_superadmin_test.go internal/service/course.go internal/service/course_test.go internal/service/course_lesson.go internal/service/course_lesson_test.go internal/repository/course_gorm.go internal/repository/course_gorm_test.go cmd/server/main.go internal/api/test_helpers_test.go internal/server/server_test.go internal/test/integration/core_flows_test.go
  git commit -m "feat: enforce official course resource ownership"
  ```

---

### Task 5: Build the Shared Visual Uploader

**Files:**
- Create: `web/admin/src/components/MediaUploader.tsx`
- Modify: `web/admin/src/api/resource.ts`
- Modify: `web/admin/src/styles.css`

**Interfaces:**
- Consumes: Task 3 platform APIs.
- Produces:
  ```ts
  export type UploadedMedia = Pick<
    Resource,
    'id' | 'name' | 'resource_type' | 'url' | 'size_bytes'
  >

  export interface MediaUploaderProps {
    value?: UploadedMedia
    onChange?: (value?: UploadedMedia) => void
    accept: 'image' | 'video' | 'document' | 'all'
    upload: (file: File, onProgress: (percent: number) => void) => Promise<Resource>
    disabled?: boolean
  }
  ```

- [ ] **Step 1: Add progress-capable API methods**

  Add:

  ```ts
  upload: (file: File, onProgress?: (percent: number) => void) =>
    uploadTo('/backend/v1/resources/upload', file, onProgress),
  uploadPlatform: (file: File, onProgress?: (percent: number) => void) =>
    uploadTo('/backend/v1/admin/resources/upload', file, onProgress),
  listPlatform: (offset = 0, limit = 100) =>
    client.get<PageResult<Resource>>('/backend/v1/admin/resources', { params: { offset, limit } }),
  platformFile: (id: string) =>
    client.get<Blob>(`/backend/v1/admin/resources/${id}/file`, { responseType: 'blob' }),
  removePlatform: (id: string) =>
    client.delete(`/backend/v1/admin/resources/${id}`),
  ```

  `uploadTo` uses Axios `onUploadProgress` and reports a rounded 0–100 value
  only when `event.total` is available.

- [ ] **Step 2: Implement `MediaUploader`**

  Use Ant Design `Upload.Dragger`, `Progress`, `Image`, `Button`, and
  `Typography`. Keep the selected file in the component while uploading.
  On success call `onChange(resource)`. On failure retain the selected file,
  display the normalized API error, and expose a **Retry** button.

  Render image thumbnail for images and an icon/name/size card for videos and
  documents. Provide **Preview**, **Replace**, and **Remove** actions. Revoke
  every `URL.createObjectURL` in effect cleanup.

- [ ] **Step 3: Add responsive styling**

  Add scoped classes beginning with `.media-uploader`. Desktop uses a dashed
  drag target and side-by-side preview/actions; widths below 768 px stack the
  preview and buttons. Use the existing primary color and Ant Design tokens;
  do not introduce a second CSS framework.

- [ ] **Step 4: Build the admin frontend**

  Run:

  ```bash
  npm run build
  ```

  Working directory: `web/admin`.

  Expected: TypeScript and Vite build succeed.

- [ ] **Step 5: Commit the reusable uploader**

  ```bash
  git add web/admin/src/components/MediaUploader.tsx web/admin/src/api/resource.ts web/admin/src/styles.css
  git commit -m "feat: add reusable visual media uploader"
  ```

---

### Task 6: Upgrade the Superadmin Official Course Workflow

**Files:**
- Modify: `web/admin/src/api/officialCourse.ts`
- Modify: `web/admin/src/pages/OfficialCourses.tsx`
- Modify: `web/admin/src/pages/CourseDetail.tsx`
- Modify: `web/admin/src/layout/AdminLayout.tsx`
- Modify: `web/admin/src/routes.tsx`
- Modify: `web/admin/src/styles.css`

**Interfaces:**
- Consumes: Task 3 platform resource APIs, Task 4 official course APIs, and Task 5 `MediaUploader`.
- Produces routes:
  - `/official-courses`
  - `/official-courses/:id`

- [ ] **Step 1: Complete the official-course API client**

  Add typed `detail`, `update`, and `remove` methods using the existing generic
  course endpoints. Make `create` accept:

  ```ts
  {
    title: string
    description?: string
    cover_image?: string
    status: 0 | 1
  }
  ```

- [ ] **Step 2: Replace the one-line official-course page**

  Build a normal paginated table with cover, title/description, status,
  creation time, and actions for **Content**, **Edit**, and **Delete**.

  In create/edit modal state keep:

  ```ts
  type OfficialCourseForm = {
    title: string
    description?: string
    status: 0 | 1
    cover?: UploadedMedia
  }
  ```

  Use `MediaUploader` with `resourceApi.uploadPlatform`. Translate
  `cover?.url` to `cover_image` before saving. Do not render an `is_official`
  switch or a cover URL input. After successful creation, offer a button that
  navigates directly to `/official-courses/<id>`.

- [ ] **Step 3: Add official mode to course detail**

  Determine mode from the route prefix:

  ```ts
  const officialMode = location.pathname.startsWith('/official-courses/')
  ```

  In official mode load `resourceApi.listPlatform`; tenant mode continues to
  load `resourceApi.list`. For video/PDF lesson forms, show `MediaUploader`
  plus an existing-resource selector. Upload with `uploadPlatform` in official
  mode and `upload` in tenant mode, then set `resource_id` automatically.
  Text lessons show the text-content editor and no uploader.

- [ ] **Step 4: Remove duplicate superadmin navigation**

  Remove `/courses` from `superadminMenus`. Protect `/courses` and
  `/courses/:id` with `TenantAdminOnly`. Add:

  ```tsx
  {
    path: '/official-courses/:id',
    element: <SuperadminOnly><CourseDetail /></SuperadminOnly>,
  }
  ```

- [ ] **Step 5: Build and manually inspect**

  Run:

  ```bash
  npm run build
  ```

  Working directory: `web/admin`.

  Expected: build succeeds. Inspect desktop and narrow viewport layouts for
  the list, course modal, image preview, lesson upload, resource reuse, retry,
  replace, and remove states.

- [ ] **Step 6: Commit official-course frontend**

  ```bash
  git add web/admin/src/api/officialCourse.ts web/admin/src/pages/OfficialCourses.tsx web/admin/src/pages/CourseDetail.tsx web/admin/src/layout/AdminLayout.tsx web/admin/src/routes.tsx web/admin/src/styles.css
  git commit -m "feat: add complete superadmin official course workflow"
  ```

---

### Task 7: Remove Remaining Uploaded-Media URL Inputs

**Files:**
- Modify: `web/admin/src/pages/Courses.tsx`
- Modify: `web/admin/src/pages/ThemeSettings.tsx`
- Modify: `web/admin/src/pages/Resources.tsx`
- Modify: `web/admin/src/styles.css`

**Interfaces:**
- Consumes: Task 5 `MediaUploader` and tenant `resourceApi.upload`.
- Produces: no image/video upload URL inputs in the current admin application.

- [ ] **Step 1: Replace the tenant course cover input**

  Remove:

  ```tsx
  <Form.Item label="封面地址" name="cover_image">
    <Input placeholder="https://..." />
  </Form.Item>
  ```

  Use `MediaUploader accept="image"` and translate its resource URL back into
  `CourseInput.cover_image` on save. Preserve edit preview from the existing
  `cover_image`.

- [ ] **Step 2: Replace the tenant logo URL input**

  Remove both the **Logo 地址** text field and the separate basic upload button.
  Use one `MediaUploader accept="image"` bound to `logo_url`. Preserve brand
  color and welcome text behavior.

- [ ] **Step 3: Polish the resource-library upload**

  Replace the single upload button in `Resources.tsx` with the shared uploader
  configured as `accept="all"`. Refresh the table after success, clear the
  uploader value, and retain existing open/delete actions.

- [ ] **Step 4: Run the media-input inventory**

  Run:

  ```bash
  rg -n '封面地址|Logo 地址|视频地址|image.*url|video.*url|name="(cover_image|logo_url)"' web/admin/src web/pc/src web/h5/src --glob '*.{ts,tsx}'
  ```

  Expected: no upload form represented as a URL text input. Read-only rendering
  fields and API model names may still match; inspect every match and confirm it
  is not an upload input.

- [ ] **Step 5: Build all three frontends**

  Run:

  ```bash
  (cd web/admin && npm run build)
  (cd web/pc && npm run build)
  (cd web/h5 && npm run build)
  ```

  Expected: all three builds succeed.

- [ ] **Step 6: Commit remaining upload UI changes**

  ```bash
  git add web/admin/src/pages/Courses.tsx web/admin/src/pages/ThemeSettings.tsx web/admin/src/pages/Resources.tsx web/admin/src/styles.css
  git commit -m "feat: replace media URL inputs with upload controls"
  ```

---

### Task 8: Full Verification and Collaboration Record

**Files:**
- Modify: `.codex/codex-log.md`
- Modify only if the current task requires status changes: `.codex/progress.md`, `.codex/issues.md`

**Interfaces:**
- Consumes: all previous tasks.
- Produces: verified build, test evidence, and concise collaboration history.

- [ ] **Step 1: Run formatting and focused tests**

  Run:

  ```bash
  gofmt -w internal/repository/resource.go internal/repository/resource_gorm.go internal/repository/resource_gorm_test.go internal/service/resource.go internal/service/resource_test.go internal/service/resource_superadmin_test.go internal/service/course.go internal/service/course_test.go internal/service/course_lesson.go internal/service/course_lesson_test.go internal/repository/course_gorm.go internal/repository/course_gorm_test.go internal/api/resource.go internal/api/resource_test.go internal/api/course.go internal/api/course_superadmin_test.go internal/server/server.go internal/server/server_test.go cmd/server/main.go internal/api/test_helpers_test.go internal/test/integration/core_flows_test.go
  go test ./internal/repository ./internal/service ./internal/api ./internal/server ./internal/test/integration -count=1
  ```

  Expected: PASS.

- [ ] **Step 2: Run complete backend verification**

  Run:

  ```bash
  go build ./...
  go test ./... -count=1
  ```

  Expected: both commands exit 0.

- [ ] **Step 3: Run complete frontend verification**

  Run:

  ```bash
  (cd web/admin && npm run build)
  (cd web/pc && npm run build)
  (cd web/h5 && npm run build)
  ```

  Expected: all commands exit 0.

- [ ] **Step 4: Verify the production-critical flows locally**

  Exercise with `httptest` or the local server:

  1. superadmin uploads a cover, creates a draft official course, publishes it,
     adds a chapter, uploads a video/PDF lesson, previews it, and deletes only
     an unreferenced platform resource;
  2. tenant enables the official course;
  3. tenant admin and instructor can preview;
  4. enrolled learner can read the lesson file;
  5. unenrolled learner and disabled tenant receive denial;
  6. tenant resource listings never contain platform or other-tenant rows.

- [ ] **Step 5: Append the collaboration log**

  Add a concise entry to `.codex/codex-log.md` with exactly these headings:

  ```markdown
  ## YYYY-MM-DD Superadmin official course media
  - 任务执行摘要：
  - 关键修改：
  - 评审反馈：
  - 下一步建议：
  ```

  Record actual test output and any unresolved issue. Do not claim deployment
  or push unless those actions were completed.

- [ ] **Step 6: Review the final diff**

  Run:

  ```bash
  git status --short
  git diff --check
  git diff --stat HEAD~7..HEAD
  git log --oneline --decorate -8
  ```

  Expected: only scoped files are changed, no whitespace errors, and no build
  outputs or secrets are staged.

- [ ] **Step 7: Commit the collaboration record**

  ```bash
  git add .codex/codex-log.md .codex/progress.md .codex/issues.md
  git commit -m "docs: record official course media completion"
  ```

  If `.codex/progress.md` or `.codex/issues.md` did not require a change, omit
  those exact paths from `git add`.
