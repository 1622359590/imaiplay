# Learner and Station Workspaces Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (- [ ]) syntax for tracking.

**Goal:** Build the screenshot-matched learner workspace and role-aware station administration workspace on real enrollment, progress, learning-time, resource, plan, and tenant data.

**Architecture:** First integrate the tested 97fee3e mainline, which already contains course materials, secure media playback, portal sessions, domain binding, and the light admin baseline. Add narrowly scoped domain/repository/service/API units for categories, assignment type, learning-time aggregation, demo-record tracking, learner authorization, and discriminated dashboards; then consume those contracts from independent PC learner and admin React workspaces. Keep H5 backward-compatible and finish with automated, browser-interaction, responsive, and reference-image comparison gates.

**Tech Stack:** Go 1.22, Gin 1.10, GORM 1.31, SQLite/PostgreSQL, React 18, TypeScript 5.7, Vite 6, React Router 6/7, Ant Design 5, ECharts, Vitest, Node test runner.

## Global Constraints

- Execute in a clean isolated worktree created through superpowers:using-git-worktrees; do not modify the dirty source worktree.
- Integrate commit 97fee3e before feature work and retain commits cdd05d9 and 6b40df2.
- Preserve lazyWithReload and RouteErrorPage behavior for stale admin chunks.
- Preserve the lesson player max-width of 1000px and its left edge alignment with the shared learner container.
- Never stage the source worktree's .codex files, internal/service/course_test.go, web/admin/package-lock.json, untracked server, or other unrelated user changes.
- Station owner is backend role tenant_admin and UI label 站长; superadmin and instructor remain distinct roles.
- Learner course content, lesson resources, playback URLs, and material downloads require an active enrollment and a currently visible course; authorization misses return 404.
- Learning time comes only from active playback heartbeat deltas; seeking, pausing, and background suspension do not count.
- The default station shell is a 48px header, 200px white sidebar, #f6f6f6 content background, 24px outer/grid spacing, and 12px card radius.
- The station sidebar supports multiple simultaneously open groups, sessionStorage persistence, route-forced expansion, a 76px compact mode, and a drawer below 960px.
- The PC learner content container has a 1600px maximum width and 24px desktop safe margins.
- Use tenant logo, name, welcome text, and primary color; do not copy PLAYEDU or hard-code screenshot business data.
- Do not create fake departments, product-document pages, CSS charts, hand-drawn SVGs, placeholder art, or fabricated dashboard metrics.
- Use @ant-design/icons for UI icons and ECharts for the resource donut.
- Keep PC TypeScript semicolons and admin TypeScript's existing no-semicolon style.
- All domain changes are test-first and every task ends with focused tests plus a commit.

---

## File Map

### Integration baseline

- Merge 97fee3e: course materials, portal/auth sessions, secure playback, official-course management, domain binding, media upload/preview, plan/storage status, and light admin styles.
- Preserve current branch files: web/admin/src/utils/lazyWithReload.ts, web/admin/src/components/RouteErrorPage.tsx, web/admin/src/routes.tsx recovery wiring, and web/pc/src/styles.css player alignment.

### Backend domain and persistence

- Create internal/domain/course_category.go: platform/tenant single-level course category.
- Modify internal/domain/course.go: optional CategoryID relation.
- Modify internal/domain/course_enrollment.go: required/optional AssignmentType.
- Create internal/domain/learning_time.go: daily aggregate and idempotent heartbeat receipt.
- Create internal/domain/tenant_demo_record.go: explicit seeded-record registry.
- Modify internal/migration/migration.go and internal/migration/migration_test.go: migration v13 and indexes.
- Create internal/repository/course_category.go and internal/repository/course_category_gorm.go: scoped category CRUD.
- Modify internal/repository/course.go and internal/repository/course_gorm.go: owner-scoped and assigned-course queries.
- Modify internal/repository/course_enrollment.go and internal/repository/course_enrollment_gorm.go: assignment updates and active-enrollment lookup.
- Create internal/repository/learning_time.go and internal/repository/learning_time_gorm.go: transactional receipt plus daily upsert.
- Create internal/repository/learner_overview.go and internal/repository/learner_overview_gorm.go: learner summaries and course-level recent learning.
- Create internal/repository/tenant_demo_record.go and internal/repository/tenant_demo_record_gorm.go: registered seed IDs.
- Modify internal/repository/dashboard.go and internal/repository/dashboard_gorm.go: tenant, instructor, and platform aggregate DTOs.

### Backend services and APIs

- Create internal/service/course_category.go and internal/api/course_category.go: normalized category CRUD.
- Modify internal/service/course.go, internal/service/course_chapter.go, and internal/service/course_lesson.go: instructor ownership and learner enrollment gates.
- Modify internal/service/enrollment.go and internal/api/enrollment.go: assignment type create/update.
- Create internal/service/learner_access.go: one authorization policy for course, lesson resource, and material access.
- Create internal/service/learner_overview.go and internal/api/learner_overview.go: one learner-home payload.
- Modify internal/service/progress.go and internal/api/progress.go: heartbeat, daily time, and course-level recent items.
- Modify internal/service/course_material.go and internal/api/course_material.go: enrollment-gated download and instructor read-only list.
- Modify internal/service/resource.go, internal/api/resource.go, and internal/security/playback.go: lesson relation checks and course-bound playback tickets.
- Modify internal/service/dashboard.go and internal/api/dashboard.go: scope-discriminated dashboard responses.
- Modify internal/service/auth.go and internal/api/auth.go: current-user response.
- Modify internal/service/tenant_register.go: registered demo data and safe clearing.
- Modify internal/server/server.go and cmd/server/main.go: handlers, repositories, dependencies, and routes.

### PC learner app

- Create web/pc/src/api/learner.ts: overview and course-summary contracts.
- Modify web/pc/src/api/course.ts and web/pc/src/api/progress.ts: safe detail, material, resource, and heartbeat calls.
- Create web/pc/src/utils/learnerCourses.ts and web/pc/src/utils/watchHeartbeat.ts: pure filtering and heartbeat state.
- Create web/pc/src/api/learner.test.ts, web/pc/src/utils/learnerCourses.test.ts, and web/pc/src/utils/watchHeartbeat.test.ts.
- Modify web/pc/src/router.tsx and web/pc/src/components/AppLayout.tsx: Home/Recent navigation.
- Replace web/pc/src/pages/HomePage.tsx and create web/pc/src/pages/RecentPage.tsx.
- Create web/pc/src/components/LearningSummary.tsx and web/pc/src/components/LearnerFilters.tsx.
- Modify web/pc/src/components/CourseCard.tsx, web/pc/src/components/CourseGrid.tsx, and web/pc/src/components/CourseMaterials.tsx.
- Modify web/pc/src/pages/CourseDetailPage.tsx and web/pc/src/pages/LessonPlayerPage.tsx.
- Modify web/pc/src/styles.css: exact desktop and responsive learner layout.

### Admin app

- Create web/admin/src/config/adminNavigation.ts: role menus, route permissions, open-group helpers, and role labels.
- Create web/admin/tests/adminNavigation.test.ts.
- Modify web/admin/src/api/auth.ts, web/admin/src/store/userSlice.ts, and web/admin/src/routes.tsx: /auth/me bootstrap and shared RoleRoute.
- Create web/admin/src/context/AdminThemeContext.tsx: resolved tenant brand and accessible primary-color derivatives.
- Modify web/admin/src/layout/AdminLayout.tsx and web/admin/src/components/AdminThemeProvider.tsx: screenshot shell and tenant brand.
- Create web/admin/src/api/courseCategory.ts and web/admin/src/pages/CourseCategories.tsx.
- Modify web/admin/src/api/course.ts, web/admin/src/pages/Courses.tsx, and web/admin/src/pages/CourseDetail.tsx: category, assignment type, and role-safe materials.
- Modify web/admin/src/pages/Users.tsx and web/admin/src/pages/Resources.tsx: one-shot quick actions and instructor restrictions.
- Create web/admin/src/utils/oneShotAction.ts and web/admin/tests/oneShotAction.test.ts.
- Replace web/admin/src/api/dashboard.ts and web/admin/src/pages/Dashboard.tsx.
- Create web/admin/src/utils/dashboardViewModel.ts, web/admin/src/components/ResourceDonut.tsx, and web/admin/tests/dashboardViewModel.test.ts.
- Modify web/admin/package.json and web/admin/package-lock.json: ECharts dependency while preserving existing lock changes.
- Modify web/admin/src/styles.css: station cards, quick actions, ranking, chart, and responsive behavior.

### Verification

- Modify design-qa.md: five-reference comparison log and final result.
- Run all Go, PC, H5, and admin tests/builds plus in-app browser interaction and visual checks.

---

### Task 1: Integrate the Tested Mainline in an Isolated Worktree

**Files:**
- Merge source: commit 97fee3e
- Resolve: web/admin/src/routes.tsx
- Resolve: web/admin/src/styles.css
- Resolve: web/pc/src/styles.css

**Interfaces:**
- Consumes: current branch at 0387aec or later, plus mainline 97fee3e.
- Produces: one clean feature branch containing both material/media/portal work and current stale-chunk/player fixes.

- [ ] **Step 1: Create and verify the isolated worktree**

Use superpowers:using-git-worktrees, then run:

~~~bash
git status --short --branch
git rev-parse --verify 97fee3e
~~~

Expected: the new worktree is clean and 97fee3e resolves.

- [ ] **Step 2: Start the merge without committing**

~~~bash
git merge --no-ff --no-commit 97fee3e
git status --short
~~~

Expected: only real committed-tree conflicts appear; no source-worktree untracked files are present.

- [ ] **Step 3: Resolve the known behavior contracts**

In web/admin/src/routes.tsx, keep lazyWithReload and RouteErrorPage while retaining all mainline portal/admin routes. In web/admin/src/styles.css, keep the mainline light admin rules plus route-error styles. In web/pc/src/styles.css, retain mainline responsive styles and this invariant:

~~~css
.player-content {
  width: min(100%, 1000px);
  margin-inline: 0;
}
~~~

- [ ] **Step 4: Verify the integrated baseline**

~~~bash
go test ./...
(cd web/pc && npm test && npm run build)
(cd web/h5 && npm test && npm run build)
(cd web/admin && npm test && npm run build)
~~~

Expected: all commands pass before feature code is added.

- [ ] **Step 5: Commit the baseline merge**

~~~bash
git add web/admin/src/routes.tsx web/admin/src/styles.css web/pc/src/styles.css
git commit
~~~

Expected commit subject: merge mainline learner materials and portal updates.

---

### Task 2: Add the Course, Assignment, Learning-Time, and Demo Schemas

**Files:**
- Create: internal/domain/course_category.go
- Modify: internal/domain/course.go
- Modify: internal/domain/course_enrollment.go
- Create: internal/domain/learning_time.go
- Create: internal/domain/tenant_demo_record.go
- Modify: internal/migration/migration.go
- Modify: internal/migration/migration_test.go

**Interfaces:**
- Produces: domain.AssignmentRequired, domain.AssignmentOptional, Course.CategoryID, CourseEnrollment.AssignmentType, LearningDailyStat, LearningTimeReport, CourseCategory, and TenantDemoRecord.
- Produces: migration v13 with all unique and query indexes.

- [ ] **Step 1: Write failing migration tests**

Add tests named TestMigrationV13AddsLearningWorkspaceSchema and TestMigrationV13IsIdempotent. The first migrates a v12 database and asserts the new columns/tables plus these unique failures:

~~~go
category := &domain.CourseCategory{
    BaseModel: domain.BaseModel{TenantID: "tenant-1"},
    Name: "销售", NormalizedName: "销售",
}
if err := database.Create(category).Error; err != nil {
    t.Fatal(err)
}
duplicateCategory := *category
duplicateCategory.ID = ""
if err := database.Create(&duplicateCategory).Error; err == nil {
    t.Fatal("expected duplicate category to fail")
}

report := &domain.LearningTimeReport{
    BaseModel: domain.BaseModel{TenantID: "tenant-1"},
    UserID: "learner-1", ReportID: "report-1",
}
if err := database.Create(report).Error; err != nil {
    t.Fatal(err)
}
duplicateReport := *report
duplicateReport.ID = ""
if err := database.Create(&duplicateReport).Error; err == nil {
    t.Fatal("expected duplicate report to fail")
}
~~~

Also assert an existing enrollment reads AssignmentType as required after migration.

- [ ] **Step 2: Run the migration tests and verify failure**

~~~bash
go test ./internal/migration -run 'TestMigrationV13' -count=1
~~~

Expected: FAIL because v13 and the domain types do not exist.

- [ ] **Step 3: Define the exact domain types**

~~~go
const (
    AssignmentRequired = "required"
    AssignmentOptional = "optional"
)

type CourseCategory struct {
    BaseModel
    Name           string `gorm:"not null" json:"name"`
    NormalizedName string `gorm:"not null" json:"-"`
    SortOrder      int    `gorm:"default:0" json:"sort_order"`
    Status         int    `gorm:"default:1" json:"status"`
}

type LearningDailyStat struct {
    BaseModel
    UserID          string `gorm:"index;not null" json:"user_id"`
    StudyDate       string `gorm:"size:10;index;not null" json:"study_date"`
    DurationSeconds int64  `gorm:"not null;default:0" json:"duration_seconds"`
}

type LearningTimeReport struct {
    BaseModel
    UserID              string `gorm:"index;not null"`
    LessonID            string `gorm:"index;not null"`
    ReportID            string `gorm:"not null"`
    WatchedSecondsDelta int    `gorm:"not null"`
}

type TenantDemoRecord struct {
    BaseModel
    BatchID    string `gorm:"index;not null"`
    RecordType string `gorm:"not null"`
    RecordID   string `gorm:"not null"`
}
~~~

Add CategoryID *string to Course and AssignmentType string with default required to CourseEnrollment.

- [ ] **Step 4: Implement migration v13**

AutoMigrate the four new models and changed Course/CourseEnrollment, backfill empty assignment values to required, then create:

~~~sql
CREATE UNIQUE INDEX IF NOT EXISTS idx_course_categories_scope_name
  ON course_categories (tenant_id, normalized_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_learning_daily_user_date
  ON learning_daily_stats (tenant_id, user_id, study_date);
CREATE UNIQUE INDEX IF NOT EXISTS idx_learning_report_idempotency
  ON learning_time_reports (tenant_id, user_id, report_id);
CREATE INDEX IF NOT EXISTS idx_learning_daily_ranking
  ON learning_daily_stats (tenant_id, study_date, duration_seconds);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_demo_record
  ON tenant_demo_records (tenant_id, record_type, record_id);
~~~

- [ ] **Step 5: Run focused and full migration tests**

~~~bash
go test ./internal/migration -count=1
~~~

Expected: PASS on a fresh database, upgraded v12 database, and repeated AutoMigrate.

- [ ] **Step 6: Commit the schema**

~~~bash
git add internal/domain internal/migration
git commit -m "feat: add learning workspace schema"
~~~

---

### Task 3: Implement Scoped Course Categories and Registered Demo Data

**Files:**
- Create: internal/repository/course_category.go
- Create: internal/repository/course_category_gorm.go
- Create: internal/repository/course_category_gorm_test.go
- Create: internal/repository/tenant_demo_record.go
- Create: internal/repository/tenant_demo_record_gorm.go
- Create: internal/service/course_category.go
- Create: internal/service/course_category_test.go
- Create: internal/api/course_category.go
- Create: internal/api/course_category_test.go
- Modify: internal/service/tenant_register.go
- Modify: internal/service/tenant_register_test.go
- Modify: internal/server/server.go
- Modify: cmd/server/main.go

**Interfaces:**
- Produces: NormalizeCourseCategoryName(name) (display, normalized string, err error).
- Produces: CourseCategoryService Create/List/Update/Delete with tenant or platform scope.
- Produces: TenantDemoRecordRepository RegisterBatch, HasRecords, ListByTenant, DeleteBatch.

- [ ] **Step 1: Write failing category and demo safety tests**

Cover whitespace folding, Unicode case folding, tenant isolation, platform-only mutations, a 409 when a category is referenced, and no name-based demo deletion. Use these assertions:

~~~go
display, normalized, err := NormalizeCourseCategoryName("  Sales   Enablement ")
if err != nil || display != "Sales Enablement" || normalized != "sales enablement" {
    t.Fatalf("normalized category = %q/%q err=%v", display, normalized, err)
}

businessCourse := &domain.Course{
    BaseModel: domain.BaseModel{TenantID: "tenant-1"},
    Title: demoCourseTitle, CreatedBy: "admin-1",
}
if err := database.Create(businessCourse).Error; err != nil {
    t.Fatal(err)
}
adminCtx := usercontext.WithUser(
    context.Background(), "admin-1", "tenant-1", "admin@example.com", "tenant_admin",
)
if err := registration.ClearDemoData(adminCtx); err != nil {
    t.Fatal(err)
}
var remaining int64
if err := database.Model(&domain.Course{}).
    Where("id = ?", businessCourse.ID).Count(&remaining).Error; err != nil || remaining != 1 {
    t.Fatalf("unregistered business course count=%d error=%v", remaining, err)
}
~~~

- [ ] **Step 2: Run focused tests and verify failure**

~~~bash
go test ./internal/repository ./internal/service ./internal/api -run 'CourseCategory|DemoData' -count=1
~~~

Expected: FAIL because repositories and services are missing.

- [ ] **Step 3: Implement normalization and scoped CRUD**

Use norm.NFKC plus cases.Fold from golang.org/x/text, strings.Fields for whitespace, and explicit role rules:

~~~go
func NormalizeCourseCategoryName(value string) (string, string, error) {
    display := strings.Join(strings.Fields(norm.NFKC.String(value)), " ")
    if display == "" || len([]rune(display)) > 64 {
        return "", "", errorsx.BadRequest("invalid course category name")
    }
    return display, cases.Fold().String(display), nil
}
~~~

GET tenant categories allows tenant_admin and instructor; tenant writes allow tenant_admin; /admin/course-categories allows only superadmin and uses empty tenant scope.

- [ ] **Step 4: Register and clear demo records by ID**

During tenant creation, generate one BatchID and register every seeded course, chapter, lesson, learner, and resource ID in the same transaction. ClearDemoData loads only registered IDs for the current tenant, deletes in dependency order, and removes registration rows in the same transaction. Never query by fixed title, email, or resource name.

- [ ] **Step 5: Wire routes and dependencies**

Add:

~~~go
backend.GET("/course-categories", categoryHandler.List)
backend.POST("/course-categories", categoryHandler.Create)
backend.PUT("/course-categories/:id", categoryHandler.Update)
backend.DELETE("/course-categories/:id", categoryHandler.Delete)
backend.GET("/admin/course-categories", categoryHandler.ListPlatform)
backend.POST("/admin/course-categories", categoryHandler.CreatePlatform)
backend.PUT("/admin/course-categories/:id", categoryHandler.UpdatePlatform)
backend.DELETE("/admin/course-categories/:id", categoryHandler.DeletePlatform)
~~~

- [ ] **Step 6: Run category, demo, and integration tests**

~~~bash
go test ./internal/repository ./internal/service ./internal/api ./internal/test/integration -count=1
~~~

Expected: PASS with no legacy-name deletions.

- [ ] **Step 7: Commit category and demo behavior**

~~~bash
git add internal cmd/server/main.go
git commit -m "feat: manage course categories and registered demo data"
~~~

---

### Task 4: Enforce Assignment Type and Instructor Ownership

**Files:**
- Modify: internal/repository/course.go
- Modify: internal/repository/course_gorm.go
- Modify: internal/repository/course_gorm_test.go
- Modify: internal/repository/course_enrollment.go
- Modify: internal/repository/course_enrollment_gorm.go
- Modify: internal/repository/course_enrollment_gorm_test.go
- Modify: internal/service/course.go
- Modify: internal/service/course_test.go
- Modify: internal/service/course_chapter.go
- Modify: internal/service/course_chapter_test.go
- Modify: internal/service/course_lesson.go
- Modify: internal/service/course_lesson_test.go
- Modify: internal/service/enrollment.go
- Modify: internal/service/enrollment_test.go
- Modify: internal/service/resource.go
- Modify: internal/service/resource_test.go
- Modify: internal/api/enrollment.go
- Modify: internal/api/enrollment_test.go
- Modify: internal/api/resource.go
- Modify: internal/server/server.go

**Interfaces:**
- Produces: EnrollmentService.Enroll(ctx, courseID, userID, assignmentType).
- Produces: EnrollmentService.UpdateAssignment(ctx, enrollmentID, assignmentType).
- Produces: CourseRepository FindByTenantAndCreator and ownership-safe manager services.
- Produces: instructor resource list/upload but no delete/category/material mutations.

- [ ] **Step 1: Write failing role and assignment tests**

Seed an instructor-owned course and a second instructor's foreign course, chapter, lesson, and resource. Table-drive required/optional validation and every direct-ID path:

~~~go
appErrorCode := func(err error) int {
    var appErr *errorsx.AppError
    if errors.As(err, &appErr) {
        return appErr.Code
    }
    return 0
}
tests := []struct {
    name string
    call func() error
}{
    {"get foreign course", func() error {
        _, err := courseService.Get(instructorContext, foreignCourse.ID)
        return err
    }},
    {"update foreign chapter", func() error {
        _, err := chapterService.Update(instructorContext, foreignChapter.ID, "changed", 1)
        return err
    }},
    {"delete foreign lesson", func() error {
        return lessonService.Delete(instructorContext, foreignLesson.ID)
    }},
    {"delete any resource", func() error {
        return resourceService.Delete(instructorContext, uploadedResource.ID)
    }},
    {"mutate material", func() error {
        _, err := materialService.Add(instructorContext, ownCourse.ID, materialInput)
        return err
    }},
}
for _, test := range tests {
    t.Run(test.name, func(t *testing.T) {
        if err := test.call(); appErrorCode(err) != 40300 {
            t.Fatalf("error = %v", err)
        }
    })
}
~~~

Also assert tenant_admin can change required to optional and an invalid value returns 400.

- [ ] **Step 2: Run focused tests and verify failure**

~~~bash
go test ./internal/service ./internal/api -run 'Assignment|Instructor|CourseManager' -count=1
~~~

Expected: FAIL on current broad courseManager/resourceManager behavior.

- [ ] **Step 3: Centralize manager-course ownership**

Add requireManageableCourse(ctx, id) that returns:

~~~go
switch role {
case "superadmin":
    allow = course.IsOfficial && course.TenantID == ""
case "tenant_admin":
    allow = !course.IsOfficial && course.TenantID == tenantID
case "instructor":
    allow = !course.IsOfficial && course.TenantID == tenantID &&
        course.CreatedBy == userID
default:
    allow = false
}
~~~

Chapter and lesson services must resolve their parent course and call the same policy before list/create/update/delete.

- [ ] **Step 4: Add assignment create/update**

Validate only domain.AssignmentRequired and domain.AssignmentOptional. Add PUT /backend/v1/enrollments/:id with:

~~~json
{"assignment_type":"optional"}
~~~

Existing enrollments remain required.

- [ ] **Step 5: Split instructor resource capability**

Allow tenant_admin and instructor through resource upload/list. Keep delete and resource-category writes tenant_admin-only. Do not permit instructor course-material writes; allow a read-only material list only after requireManageableCourse confirms ownership.

- [ ] **Step 6: Run role, repository, and API tests**

~~~bash
go test ./internal/repository ./internal/service ./internal/api -count=1
~~~

Expected: PASS, including direct-ID ownership tests.

- [ ] **Step 7: Commit role and assignment behavior**

~~~bash
git add internal
git commit -m "feat: enforce course ownership and assignment types"
~~~

---

### Task 5: Record Real Learning Time and Serve Learner Overview

**Files:**
- Create: internal/repository/learning_time.go
- Create: internal/repository/learning_time_gorm.go
- Create: internal/repository/learning_time_gorm_test.go
- Create: internal/repository/learner_overview.go
- Create: internal/repository/learner_overview_gorm.go
- Create: internal/repository/learner_overview_gorm_test.go
- Create: internal/service/learner_overview.go
- Create: internal/service/learner_overview_test.go
- Create: internal/api/learner_overview.go
- Create: internal/api/learner_overview_test.go
- Modify: internal/service/progress.go
- Modify: internal/service/progress_test.go
- Modify: internal/api/progress.go
- Modify: internal/api/progress_test.go
- Modify: internal/server/server.go
- Modify: cmd/server/main.go

**Interfaces:**
- Produces: LearningTimeRepository.Record(ctx, report, studyDate) (recorded bool, err error).
- Produces: LearnerOverviewService.Get(ctx) (LearnerOverview, error).
- Produces: ProgressService.Report(ctx, lessonID, position, percent, watchedDelta, reportID).
- Produces: GET /api/v1/learner/overview and course-level GET /api/v1/recent-learning.

- [ ] **Step 1: Write failing heartbeat and aggregation tests**

Cover 1, 15, and 60 second deltas; reject 0, negative, 61+, or missing report ID when delta is positive. Assert duplicate report IDs do not increment:

~~~go
newReport := func() *domain.LearningTimeReport {
    return &domain.LearningTimeReport{
        BaseModel: domain.BaseModel{TenantID: "tenant-1"},
        UserID: "learner-1", LessonID: "lesson-1",
        ReportID: "report-1", WatchedSecondsDelta: 15,
    }
}
first, err := repo.Record(ctx, newReport(), "2026-08-05")
second, err2 := repo.Record(ctx, newReport(), "2026-08-05")
if err != nil || err2 != nil || !first || second {
    t.Fatalf("recorded = %v/%v errors=%v/%v", first, second, err, err2)
}
var stat domain.LearningDailyStat
if err := database.Where(
    "tenant_id = ? AND user_id = ? AND study_date = ?",
    "tenant-1", "learner-1", "2026-08-05",
).First(&stat).Error; err != nil || stat.DurationSeconds != 15 {
    t.Fatalf("daily stat=%#v error=%v", stat, err)
}
~~~

Test Asia/Shanghai midnight, required totals, zero-lesson incompletion, official enabled courses, category summaries, weighted course progress, and one recent item per course.

- [ ] **Step 2: Run focused tests and verify failure**

~~~bash
go test ./internal/repository ./internal/service ./internal/api -run 'LearningTime|LearnerOverview|RecentLearning' -count=1
~~~

Expected: FAIL because the repository and endpoint do not exist.

- [ ] **Step 3: Implement transactional heartbeat recording**

Insert LearningTimeReport first and upsert LearningDailyStat in the same transaction. A unique-constraint duplicate returns recorded=false without changing the daily total. Compute StudyDate with time.LoadLocation("Asia/Shanghai").

- [ ] **Step 4: Implement the fixed overview DTO**

~~~go
type CourseCategorySummary struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type CourseSummary struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    CoverImage  string                 `json:"cover_image"`
    Category    *CourseCategorySummary `json:"category,omitempty"`
}

type LessonSummary struct {
    ID                  string `json:"id"`
    Title               string `json:"title"`
    DurationSeconds     int    `json:"duration_seconds"`
    LastPositionSeconds int    `json:"last_position_seconds"`
}

type LearnerCourseSummary struct {
    Course                CourseSummary  `json:"course"`
    AssignmentType        string         `json:"assignment_type"`
    LessonCount           int            `json:"lesson_count"`
    CompletedLessonCount  int            `json:"completed_lesson_count"`
    ProgressPercent       int            `json:"progress_percent"`
    LastLearnedAt         *time.Time     `json:"last_learned_at,omitempty"`
    RecentLesson          *LessonSummary `json:"recent_lesson,omitempty"`
}

type LearnerOverview struct {
    RequiredCompleted    int                     `json:"required_completed"`
    RequiredTotal        int                     `json:"required_total"`
    TodayLearningSeconds int64                   `json:"today_learning_seconds"`
    TotalLearningSeconds int64                   `json:"total_learning_seconds"`
    Categories           []CourseCategorySummary `json:"categories"`
    Courses              []LearnerCourseSummary  `json:"courses"`
}
~~~

Course progress is the rounded average of all lesson percentages; no lessons means 0 and incomplete.

- [ ] **Step 5: Restrict legacy course list/detail to active assignments**

Change CourseService.ListPublished and GetPublishedDetail to use active enrollments and the same tenant/official visibility rule as the overview. This keeps H5 and any older PC client aligned with the assigned-course policy even when they do not call /learner/overview.

- [ ] **Step 6: Return course-level recent learning**

Query the newest progress row per course for active assigned, visible courses. Sort last_learned_at descending and return course, recent lesson, course progress, and last position. Do not paginate lesson rows before deduplication.

- [ ] **Step 7: Wire the API and backward-compatible heartbeat body**

~~~go
var request struct {
    PositionSeconds     int    `json:"position_seconds"`
    ProgressPercent     int    `json:"progress_percent"`
    WatchedSecondsDelta int    `json:"watched_seconds_delta"`
    ReportID            string `json:"report_id"`
}
~~~

Old clients may omit both new fields and still update position/progress without adding time.

- [ ] **Step 8: Run focused and full backend tests**

~~~bash
go test ./internal/repository ./internal/service ./internal/api ./internal/test/integration -count=1
~~~

Expected: PASS with idempotent time and correct overview/recent results.

- [ ] **Step 9: Commit learning aggregation**

~~~bash
git add internal cmd/server/main.go
git commit -m "feat: aggregate learner progress and study time"
~~~

---

### Task 6: Close Course Material and Lesson Resource Authorization Gaps

**Files:**
- Create: internal/service/learner_access.go
- Create: internal/service/learner_access_test.go
- Modify: internal/service/course.go
- Modify: internal/service/course_material.go
- Modify: internal/service/course_material_test.go
- Modify: internal/service/resource.go
- Modify: internal/service/resource_test.go
- Modify: internal/api/course_material_test.go
- Modify: internal/api/resource.go
- Modify: internal/api/resource_test.go
- Modify: internal/security/playback.go
- Modify: internal/security/playback_test.go
- Modify: cmd/server/main.go

**Interfaces:**
- Produces: LearnerAccess.AuthorizeCourse(ctx, courseID) (*domain.Course, error), AuthorizeMaterial(ctx, materialID) (*domain.CourseMaterial, *domain.Course, error), and AuthorizeLessonResource(ctx, resourceID) (*domain.Course, error), all returning NotFound on policy misses.
- Produces: playback claims containing ResourceID, CourseID, UserID, TenantID, Role, and expiry.

- [ ] **Step 1: Write the authorization matrix tests**

Test active enrollment, disabled enrollment, draft course, disabled official course, other-course material, unrelated same-tenant resource, cross-tenant resource, and expired/mismatched playback claims:

~~~go
appErrorCode := func(err error) int {
    var appErr *errorsx.AppError
    if errors.As(err, &appErr) {
        return appErr.Code
    }
    return 0
}
for _, test := range []struct {
    name string
    invoke func() error
    wantNotFound bool
}{
    {"active enrollment", func() error {
        _, _, err := access.AuthorizeMaterial(enrolledContext, assignedMaterial.ID)
        return err
    }, false},
    {"not enrolled", func() error {
        _, _, err := access.AuthorizeMaterial(unassignedContext, assignedMaterial.ID)
        return err
    }, true},
    {"wrong lesson resource", func() error {
        _, err := access.AuthorizeLessonResource(enrolledContext, unrelatedResource.ID)
        return err
    }, true},
    {"official disabled", func() error {
        _, err := access.AuthorizeLessonResource(enrolledContext, disabledOfficialResource.ID)
        return err
    }, true},
} {
    t.Run(test.name, func(t *testing.T) {
        err := test.invoke()
        if test.wantNotFound != (appErrorCode(err) == 40400) {
            t.Fatalf("error = %v", err)
        }
    })
}
~~~

- [ ] **Step 2: Run focused tests and verify the existing leak**

~~~bash
go test ./internal/service ./internal/api ./internal/security -run 'LearnerAccess|CourseMaterial|Playback' -count=1
~~~

Expected: at least the unassigned material and unrelated tenant resource tests FAIL.

- [ ] **Step 3: Implement one learner access policy**

Resolve the material/resource through chapter/lesson/course relations, require an active CourseEnrollment for the current learner, and use CourseRepository visibility checks for tenant or enabled official courses. Convert every policy miss to errorsx.NotFound.

- [ ] **Step 4: Separate manager and learner streaming**

Keep admin resource preview scoped by admin role. For learner /resources/:id/file and /playback-url, call AuthorizeLessonResource before opening. CourseMaterialService.OpenForLearner calls AuthorizeMaterial before storage.

- [ ] **Step 5: Bind playback tickets to the authorized course**

Extend GeneratePlaybackToken and ValidatePlaybackToken with CourseID. Playback re-runs AuthorizeLessonResource under the token identity and rejects a course mismatch or expired ticket.

- [ ] **Step 6: Remove protected raw URLs from learner DTOs**

Published course detail returns resource_id and safe content metadata, never the stored object URL. Public external content URLs pass existing URL validation and remain allowed.

- [ ] **Step 7: Run security and full backend tests**

~~~bash
go test ./internal/security ./internal/service ./internal/api ./internal/server ./internal/test/integration -count=1
~~~

Expected: PASS; unauthorized variants consistently return 404.

- [ ] **Step 8: Commit the authorization fix**

~~~bash
git add internal cmd/server/main.go
git commit -m "fix: require enrollment for learner course resources"
~~~

---

### Task 7: Return Current User and Role-Scoped Dashboards

**Files:**
- Modify: internal/repository/dashboard.go
- Modify: internal/repository/dashboard_gorm.go
- Modify: internal/repository/dashboard_gorm_test.go
- Modify: internal/service/dashboard.go
- Modify: internal/service/dashboard_test.go
- Modify: internal/api/dashboard.go
- Modify: internal/api/dashboard_test.go
- Modify: internal/service/auth.go
- Modify: internal/service/auth_test.go
- Modify: internal/api/auth.go
- Modify: internal/api/auth_test.go
- Modify: internal/server/server.go
- Modify: cmd/server/main.go

**Interfaces:**
- Produces: GET /api/v1/auth/me with safe AuthUser.
- Produces: dashboard responses discriminated by scope=platform, scope=tenant, or scope=instructor.

- [ ] **Step 1: Write failing DTO and metric tests**

Assert exact JSON key sets per role. Tenant metrics count only active learners, include yesterday delta, active manager count, resource type zeros, at most ten ranked learners, and HasDemoData from registrations. Instructor metrics filter CreatedBy. Platform metrics retain the existing five recent tenants.

~~~go
if got.Scope != "tenant" || got.LearnerCount != 2 ||
    got.ResourceTypeCounts.Attachment != 0 ||
    len(got.TodayLearningRanking) > 10 {
    t.Fatalf("tenant dashboard = %#v", got)
}
~~~

Test the Asia/Shanghai boundary by injecting service.now.

- [ ] **Step 2: Run focused tests and verify failure**

~~~bash
go test ./internal/repository ./internal/service ./internal/api -run 'Dashboard|AuthMe|CurrentUser' -count=1
~~~

Expected: FAIL on inaccurate current metrics and missing /auth/me.

- [ ] **Step 3: Implement the three fixed response shapes**

~~~go
type ResourceTypeCounts struct {
    Video      int64 `json:"video"`
    Image      int64 `json:"image"`
    Document   int64 `json:"document"`
    Attachment int64 `json:"attachment"`
}

type LearningRankItem struct {
    UserID          string `json:"user_id"`
    DisplayName     string `json:"display_name"`
    DurationSeconds int64  `json:"duration_seconds"`
}

type TenantDashboard struct {
    Scope                      string             `json:"scope"`
    TodayLearningUserCount     int64              `json:"today_learning_user_count"`
    YesterdayLearningUserCount int64              `json:"yesterday_learning_user_count"`
    TodayLearningUserDelta     int64              `json:"today_learning_user_delta"`
    LearnerCount               int64              `json:"learner_count"`
    TodayNewLearnerCount       int64              `json:"today_new_learner_count"`
    PublishedCourseCount       int64              `json:"published_course_count"`
    CourseCount                int64              `json:"course_count"`
    ResourceCategoryCount      int64              `json:"resource_category_count"`
    ResourceCount              int64              `json:"resource_count"`
    ManagerCount               int64              `json:"manager_count"`
    HasDemoData                bool               `json:"has_demo_data"`
    ResourceTypeCounts         ResourceTypeCounts `json:"resource_type_counts"`
    TodayLearningRanking       []LearningRankItem `json:"today_learning_ranking"`
}

type PlatformTenant struct {
    ID              string    `json:"id"`
    Name            string    `json:"name"`
    Code            string    `json:"code"`
    Status          int       `json:"status"`
    LifecycleStatus string    `json:"lifecycle_status,omitempty"`
    CreatedAt       time.Time `json:"created_at"`
}

type PlatformDashboard struct {
    Scope             string           `json:"scope"`
    TenantCount       int64            `json:"tenant_count"`
    ActiveTenantCount int64            `json:"active_tenant_count"`
    LearnerCount      int64            `json:"learner_count"`
    CourseCount       int64            `json:"course_count"`
    RecentTenants     []PlatformTenant `json:"recent_tenants"`
}

type InstructorCourse struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Status    int       `json:"status"`
    UpdatedAt time.Time `json:"updated_at"`
}

type InstructorDashboard struct {
    Scope                  string             `json:"scope"`
    CourseCount            int64              `json:"course_count"`
    PublishedCourseCount   int64              `json:"published_course_count"`
    TodayLearningUserCount int64              `json:"today_learning_user_count"`
    RecentCourses          []InstructorCourse `json:"recent_courses"`
}
~~~

Each response contains only its own scope fields.

- [ ] **Step 4: Use real aggregate sources**

Use LearningDailyStat for today/yesterday/ranking, role=learner and status=1 for learners, roles tenant_admin/instructor and status=1 for managers, resources grouped by type, TenantDemoRecord for HasDemoData, and tenant-visible own plus enabled-official courses for course counts.

- [ ] **Step 5: Implement /auth/me**

Load the authenticated user ID, verify status=1 and tenant consistency, and return only id, tenant_id, name, email, phone, and role. Password and code login responses reuse the same presenter.

- [ ] **Step 6: Run focused, server, and integration tests**

~~~bash
go test ./internal/repository ./internal/service ./internal/api ./internal/server ./internal/test/integration -count=1
~~~

Expected: PASS with exact role shapes and safe current-user data.

- [ ] **Step 7: Commit dashboards and profile**

~~~bash
git add internal cmd/server/main.go
git commit -m "feat: add role-scoped dashboards and current user"
~~~

---

### Task 8: Add Typed Learner Client Data and Pure View Models

**Files:**
- Create: web/pc/src/api/learner.ts
- Create: web/pc/src/api/learner.test.ts
- Modify: web/pc/src/api/course.ts
- Modify: web/pc/src/api/progress.ts
- Create: web/pc/src/utils/learnerCourses.ts
- Create: web/pc/src/utils/learnerCourses.test.ts
- Create: web/pc/src/utils/watchHeartbeat.ts
- Create: web/pc/src/utils/watchHeartbeat.test.ts
- Modify: web/pc/package.json

**Interfaces:**
- Produces: getLearnerOverview(), getRecentLearning(), filterLearnerCourses(), courseStatus(), WatchHeartbeat.
- Consumes: fixed backend DTOs from Tasks 5–7.

- [ ] **Step 1: Write failing normalization, filter, and heartbeat tests**

~~~ts
const courses: LearnerCourse[] = [
  {
    id: 'course-1',
    title: '销售基础',
    assignmentType: 'required',
    category: { id: 'sales', name: '销售' },
    lessonCount: 2,
    completedLessonCount: 0,
    progressPercent: 25,
  },
  {
    id: 'course-2',
    title: '文化选修',
    assignmentType: 'optional',
    category: { id: 'culture', name: '文化' },
    lessonCount: 1,
    completedLessonCount: 1,
    progressPercent: 100,
  },
];
expect(filterLearnerCourses(courses, {
  tab: 'required',
  categoryId: 'sales',
})).toEqual([courses[0]]);

const heartbeat = new WatchHeartbeat(() => 'report-1');
heartbeat.play();
heartbeat.addPlayedSeconds(15);
expect(heartbeat.flush()).toEqual({
  watched_seconds_delta: 15,
  report_id: 'report-1',
});
expect(heartbeat.flush()).toBeNull();
~~~

Cover all/required/optional/completed/incomplete, category composition, zero lessons, percent clamping, pause/background, 60-second cap, and retry retaining the same report ID.

- [ ] **Step 2: Run PC tests and verify failure**

~~~bash
(cd web/pc && npm test)
~~~

Expected: FAIL because the modules are missing.

- [ ] **Step 3: Implement exact TypeScript contracts**

Use assignment_type as 'required' | 'optional', numeric progress_percent, safe material metadata, recent lesson/position, and category IDs. Preserve backend snake_case at the API boundary and map once into camelCase view models:

~~~ts
export interface LearnerCourse {
  id: string;
  title: string;
  description?: string;
  coverImage?: string;
  assignmentType: 'required' | 'optional';
  category?: { id: string; name: string };
  lessonCount: number;
  completedLessonCount: number;
  progressPercent: number;
  lastLearnedAt?: string;
  recentLesson?: {
    id: string;
    title: string;
    durationSeconds: number;
    lastPositionSeconds: number;
  };
}

export interface LearnerOverview {
  requiredCompleted: number;
  requiredTotal: number;
  todayLearningSeconds: number;
  totalLearningSeconds: number;
  categories: Array<{ id: string; name: string }>;
  courses: LearnerCourse[];
}
~~~

- [ ] **Step 4: Implement pure filters and heartbeat state**

WatchHeartbeat only accumulates while playing and visible. failed(payload) restores the same payload; acknowledged(reportID) clears only that report. Use crypto.randomUUID through an injected factory for tests.

- [ ] **Step 5: Run PC tests and build**

~~~bash
(cd web/pc && npm test && npm run build)
~~~

Expected: PASS with no UI changes yet.

- [ ] **Step 6: Commit learner client contracts**

~~~bash
git add web/pc
git commit -m "feat(pc): add learner workspace data models"
~~~

---

### Task 9: Build the Learner Home, Navigation, and Recent Learning

**Files:**
- Modify: web/pc/src/router.tsx
- Modify: web/pc/src/components/AppLayout.tsx
- Modify: web/pc/src/pages/HomePage.tsx
- Create: web/pc/src/pages/RecentPage.tsx
- Create: web/pc/src/components/LearningSummary.tsx
- Create: web/pc/src/components/LearnerFilters.tsx
- Modify: web/pc/src/components/CourseCard.tsx
- Modify: web/pc/src/components/CourseGrid.tsx
- Modify: web/pc/src/styles.css

**Interfaces:**
- Consumes: getLearnerOverview, getRecentLearning, filterLearnerCourses, and portalRoutePath.
- Produces: screenshot-matched home and course-level recent pages.

- [ ] **Step 1: Add a failing route contract test**

Extend a pure portal-routing/navigation test to assert both / and /recent remain inside default and custom-domain tenant portals and legacy /courses redirects to /.

- [ ] **Step 2: Run the focused route test and verify failure**

~~~bash
(cd web/pc && npm test)
~~~

Expected: FAIL because /recent still redirects home.

- [ ] **Step 3: Build the shell and summary cards**

AppLayout renders tenant brand, Home and Recent nav links, learner role label, and logout. Home performs one overview request, shows a shared skeleton while loading, and replaces a failed response with a retry action that preserves no stale totals. It renders Course progress and Study time cards from the overview:

~~~tsx
<LearningSummary
  completed={overview.requiredCompleted}
  required={overview.requiredTotal}
  todaySeconds={overview.todayLearningSeconds}
  totalSeconds={overview.totalLearningSeconds}
/>
~~~

- [ ] **Step 4: Build filters and course cards**

Render five mutually exclusive tabs plus a category Select. Completed cards show the medal icon and congratulation copy; incomplete cards show a numeric Progress. Cards keep Enter/Space behavior.

- [ ] **Step 5: Build course-level recent learning**

RecentPage displays one card per course, the recent lesson, course percentage, last learned time, and a continue link to the recent lesson. Empty state links back home; network failure shows retry without discarding the shell.

- [ ] **Step 6: Apply measured learner layout**

At 1543×793 use the shared container, two summary cards, two course columns, left-aligned sparse rows, and screenshot typography. At 1600px and above allow three course columns; below 760px stack.
Add visible focus rings and a prefers-reduced-motion rule that disables nonessential card and progress transitions.

- [ ] **Step 7: Run PC tests and production build**

~~~bash
(cd web/pc && npm test && npm run build)
~~~

Expected: PASS and no route/type errors.

- [ ] **Step 8: Commit learner home and recent**

~~~bash
git add web/pc
git commit -m "feat(pc): build learner dashboard and recent learning"
~~~

---

### Task 10: Build Course Tabs, Secure Downloads, and Playback Heartbeats

**Files:**
- Modify: web/pc/src/pages/CourseDetailPage.tsx
- Modify: web/pc/src/components/CourseMaterials.tsx
- Modify: web/pc/src/pages/LessonPlayerPage.tsx
- Modify: web/pc/src/api/progress.ts
- Modify: web/pc/src/styles.css

**Interfaces:**
- Consumes: safe course detail, lesson statuses, material download, authorized playback URL, and WatchHeartbeat.
- Produces: course overview/progress ring, Catalog/Materials tabs, continue state, and actual-time reporting.

- [ ] **Step 1: Write failing progress request tests**

Assert reportLessonProgress serializes:

~~~ts
{
  position_seconds: 42,
  progress_percent: 35,
  watched_seconds_delta: 15,
  report_id: 'report-1',
}
~~~

and omits time fields for a position-only compatibility update.

- [ ] **Step 2: Run the focused test and verify failure**

~~~bash
(cd web/pc && npm test)
~~~

Expected: FAIL because the request does not accept heartbeat payloads.

- [ ] **Step 3: Build the detail hero and accessible tabs**

Render cover/name/type/description/completion copy on the left and an Ant Design progress circle on the right. Use Tabs for 课程目录 and 课程附件; keep the material tab visible when empty.

- [ ] **Step 4: Render exact lesson states**

Completed rows show 已学完; partial rows show 上次学习到 mm:ss and 继续学习; untouched rows show 开始学习. Duration is formatted mm:ss from seconds, not rounded minutes.

- [ ] **Step 5: Keep safe material download behavior**

Show file name, size, loading per material, and inline error messaging. Never render a storage URL.

- [ ] **Step 6: Connect the heartbeat to the video lifecycle**

Start accumulation on playing, stop on pause/waiting/ended/document hidden, sample real played wall seconds, flush every 15 seconds, and flush on pause/ended/pagehide. Retry failed payloads with their original report_id. Seeking changes position only.

- [ ] **Step 7: Preserve player alignment and responsive detail layout**

Keep player-content at max-width 1000px and margin-inline 0. Stack the detail hero and tabs below 760px without horizontal overflow.

- [ ] **Step 8: Run PC tests/build and H5 compatibility build**

~~~bash
(cd web/pc && npm test && npm run build)
(cd web/h5 && npm test && npm run build)
~~~

Expected: PASS; H5 remains compatible with optional heartbeat fields.

- [ ] **Step 9: Commit learner detail and heartbeat**

~~~bash
git add web/pc web/h5
git commit -m "feat(pc): add course progress tabs and study heartbeat"
~~~

---

### Task 11: Build Shared Admin Role Navigation and Profile Bootstrap

**Files:**
- Create: web/admin/src/config/adminNavigation.ts
- Create: web/admin/tests/adminNavigation.test.ts
- Create: web/admin/src/context/AdminThemeContext.tsx
- Modify: web/admin/src/api/auth.ts
- Modify: web/admin/src/store/userSlice.ts
- Modify: web/admin/src/routes.tsx
- Modify: web/admin/src/layout/AdminLayout.tsx
- Modify: web/admin/src/components/AdminThemeProvider.tsx
- Modify: web/admin/src/styles.css

**Interfaces:**
- Produces: navigationForRole(role), pathsForRole(role), allowedRolesForPath(path), requiredOpenGroups(path), roleLabel(role).
- Produces: getCurrentUser() and an authenticated profile bootstrap.
- Preserves: lazyWithReload and RouteErrorPage.

- [ ] **Step 1: Write failing pure navigation tests**

~~~ts
assert.deepEqual(
  pathsForRole('instructor'),
  ['/', '/courses', '/resources'],
);
assert.deepEqual(requiredOpenGroups('/course-categories'), ['course-center']);
assert.equal(roleLabel('tenant_admin'), '站长');
assert.deepEqual(allowedRolesForPath('/theme-settings'), ['tenant_admin']);
~~~

Also assert station groups can all be open simultaneously and unknown roles get no routes.

- [ ] **Step 2: Run admin tests and verify failure**

~~~bash
(cd web/admin && npm test)
~~~

Expected: FAIL because adminNavigation does not exist.

- [ ] **Step 3: Implement one role configuration**

The same config produces menu groups and RoleRoute permissions. Station order is Home, Resources, Courses, Learners, Site settings, Security. Instructor sees only Workbench, My courses, and Resources. Superadmin sees platform groups.

- [ ] **Step 4: Bootstrap the real profile**

Add getCurrentUser() calling /api/v1/auth/me. ProtectedRoute waits for profile restoration when a token exists, dispatches setProfile on success, and clears the session on 401/403. Never display a guessed name from JWT.

- [ ] **Step 5: Resolve tenant brand and accessible theme tokens**

AdminThemeContext loads the existing backend theme endpoint after the station session is known, exposes logoURL, brandName, primaryColor, selectedMenuColor, and focusColor, and falls back to ImaiPlay plus #ff4e4f. Use a contrast helper to darken selectedMenuColor until white text reaches at least 4.5:1; superadmin uses the platform fallback without a tenant request.

- [ ] **Step 6: Replace role wrappers with RoleRoute**

~~~tsx
<RoleRoute allow={allowedRolesForPath('/users')}>
  <Users />
</RoleRoute>
~~~

Every protected route uses the shared config. Keep lazyWithReload imports and the root errorElement.

- [ ] **Step 7: Build the measured shell**

Use a 200px white Sider, 48px Header, independent menu scroll, tenant logo/name, Chinese identity, visible keyboard focus, and an accessible deep selected color. Persist openKeys in sessionStorage; union them with requiredOpenGroups(location.pathname). Use inlineCollapsed at 960–1199 and Drawer below 960.

- [ ] **Step 8: Run admin tests and build**

~~~bash
(cd web/admin && npm test && npm run build)
~~~

Expected: PASS; lazy route chunks remain present in dist/assets.

- [ ] **Step 9: Commit admin navigation**

~~~bash
git add web/admin
git commit -m "feat(admin): add role-aware grouped navigation"
~~~

---

### Task 12: Add Admin Category, Assignment, Material, and Quick Actions

**Files:**
- Create: web/admin/src/api/courseCategory.ts
- Create: web/admin/src/pages/CourseCategories.tsx
- Create: web/admin/src/utils/oneShotAction.ts
- Create: web/admin/tests/oneShotAction.test.ts
- Modify: web/admin/src/api/course.ts
- Modify: web/admin/src/pages/Courses.tsx
- Modify: web/admin/src/pages/CourseDetail.tsx
- Modify: web/admin/src/pages/Users.tsx
- Modify: web/admin/src/pages/Resources.tsx
- Modify: web/admin/src/routes.tsx
- Modify: web/admin/src/styles.css

**Interfaces:**
- Produces: courseCategoryApi CRUD, consumeOneShotAction(search, key), category_id course forms, assignment_type enrollment forms.
- Consumes: role permissions and backend Tasks 3–4.

- [ ] **Step 1: Write failing one-shot and form-mapping tests**

~~~ts
assert.deepEqual(consumeOneShotAction('?create=1', 'create'), {
  active: true,
  remainingSearch: '',
});
assert.deepEqual(consumeOneShotAction('?create=1&page=2', 'create'), {
  active: true,
  remainingSearch: '?page=2',
});
~~~

Test category_id and assignment_type survive edit normalization.

- [ ] **Step 2: Run admin tests and verify failure**

~~~bash
(cd web/admin && npm test)
~~~

Expected: FAIL because the helpers and category API are missing.

- [ ] **Step 3: Build course-category management**

CourseCategories lists active/inactive categories, creates/renames/reorders, and surfaces 409 reference conflicts. Station writes tenant endpoints; official-course editors use platform endpoints for superadmin.

- [ ] **Step 4: Add category and assignment controls**

Courses form uses a searchable category Select. CourseDetail enrollment table shows learner and assignment type, lets station owners change required/optional, and keeps instructor enrollment/material controls read-only or hidden according to shared permissions.

- [ ] **Step 5: Consume dashboard quick actions once**

Users opens the learner create modal for ?create=1, Courses opens create for ?create=1, and Resources focuses/opens upload for ?upload=1. Immediately replace the URL with remaining query parameters before opening UI.

- [ ] **Step 6: Apply resource type filters and instructor restrictions**

Resources adds All/Video/Image/Document/Attachment filters. Instructor sees list/upload/preview but no delete. Tenant admin sees full controls.

- [ ] **Step 7: Run admin tests and build**

~~~bash
(cd web/admin && npm test && npm run build)
~~~

Expected: PASS with the new lazy CourseCategories route.

- [ ] **Step 8: Commit admin maintenance features**

~~~bash
git add web/admin
git commit -m "feat(admin): manage course categories and assignments"
~~~

---

### Task 13: Build the Station Dashboard and Role-Specific Workbenches

**Files:**
- Modify: web/admin/src/api/dashboard.ts
- Create: web/admin/src/utils/dashboardViewModel.ts
- Create: web/admin/tests/dashboardViewModel.test.ts
- Create: web/admin/src/components/ResourceDonut.tsx
- Modify: web/admin/src/pages/Dashboard.tsx
- Modify: web/admin/package.json
- Modify: web/admin/package-lock.json
- Modify: web/admin/src/styles.css

**Interfaces:**
- Produces: DashboardResponse discriminated union, stationDashboardCards(data), and resourceSeries(data).
- Consumes: dashboard, plan/current, theme, and domain-bind/status APIs.

- [ ] **Step 1: Inspect and preserve the pre-existing lockfile diff**

~~~bash
git diff -- web/admin/package-lock.json
~~~

Expected: understand any existing feature-worktree difference before installing ECharts; never copy the dirty source lockfile wholesale.

- [ ] **Step 2: Write failing dashboard view-model tests**

~~~ts
const response: TenantDashboard = {
  scope: 'tenant',
  today_learning_user_count: 4,
  yesterday_learning_user_count: 6,
  today_learning_user_delta: -2,
  learner_count: 12,
  today_new_learner_count: 1,
  published_course_count: 3,
  course_count: 4,
  resource_category_count: 2,
  resource_count: 0,
  manager_count: 2,
  has_demo_data: false,
  resource_type_counts: { video: 0, image: 0, document: 0, attachment: 0 },
  today_learning_ranking: [],
};
assert.deepEqual(stationDashboardCards(response)[0].comparison, {
  direction: 'down',
  value: 2,
});
assert.equal(resourceSeries(response).reduce((sum, item) => sum + item.value, 0),
  response.resource_count);
~~~

Cover positive/zero delta, fixed zero resource types, empty ranking, plan unlimited quota, and type narrowing for all three scopes.

- [ ] **Step 3: Install ECharts**

~~~bash
(cd web/admin && npm install echarts)
~~~

Expected: package.json and package-lock.json add only the resolved ECharts dependency tree plus preserved baseline changes.

- [ ] **Step 4: Implement the discriminated API union**

~~~ts
export interface TenantDashboard {
  scope: 'tenant'
  today_learning_user_count: number
  yesterday_learning_user_count: number
  today_learning_user_delta: number
  learner_count: number
  today_new_learner_count: number
  published_course_count: number
  course_count: number
  resource_category_count: number
  resource_count: number
  manager_count: number
  has_demo_data: boolean
  resource_type_counts: Record<'video' | 'image' | 'document' | 'attachment', number>
  today_learning_ranking: Array<{
    user_id: string
    display_name: string
    duration_seconds: number
  }>
}

export interface PlatformDashboard {
  scope: 'platform'
  tenant_count: number
  active_tenant_count: number
  learner_count: number
  course_count: number
  recent_tenants: PlatformTenant[]
}

export interface InstructorDashboard {
  scope: 'instructor'
  course_count: number
  published_course_count: number
  today_learning_user_count: number
  recent_courses: Array<{
    id: string
    title: string
    status: number
    updated_at: string
  }>
}

export type DashboardResponse =
  | PlatformDashboard
  | TenantDashboard
  | InstructorDashboard
~~~

PlatformTenant is the existing typed id/name/code/status/lifecycle_status/created_at DTO. Do not keep optional platform fields on a tenant-shaped object.

- [ ] **Step 5: Build station dashboard cards**

Render two equal columns: metric groups; quick operations and Site/Plan; Today learning ranking and Resource statistics. Quick links are /users?create=1, /courses?create=1, /resources?upload=1, and /official-courses. Show Clear demo data only when has_demo_data is true.

- [ ] **Step 6: Build the real ECharts donut**

ResourceDonut initializes/disposes an ECharts instance, observes container resize, uses video #fe8650, image #ffb501, document #00cc66, and an accessible distinct attachment color. It renders center total, tooltip, labels, and textual legend; empty data uses an Ant Design Empty instead of blank sectors.

- [ ] **Step 7: Keep platform and instructor scope-specific**

Platform uses its existing four metrics plus recent tenants. Instructor shows own course total, published total, today's learners of own courses, and recent edited courses; it never renders tenant plan, users, categories, or demo actions.

- [ ] **Step 8: Handle partial failures**

Load dashboard as the primary request and plan/domain data independently. A secondary failure renders retry copy only in its card. Chart failure falls back to a four-row textual resource list.

- [ ] **Step 9: Run admin tests and build**

~~~bash
(cd web/admin && npm test && npm run build)
~~~

Expected: PASS; ECharts is code-split with Dashboard rather than the login chunk.

- [ ] **Step 10: Commit station dashboard**

~~~bash
git add web/admin/package.json web/admin/package-lock.json web/admin/src
git commit -m "feat(admin): build station operations dashboard"
~~~

---

### Task 14: Complete Automated, Browser, and Visual Verification

**Files:**
- Create or Modify: design-qa.md
- Modify only when defects are found: files from Tasks 2–13.

**Interfaces:**
- Consumes: completed backend, PC, H5, and admin workspaces.
- Produces: all green automated checks and design-qa.md ending with final result: passed.

- [ ] **Step 1: Run the complete automated suite**

~~~bash
go test ./...
(cd web/pc && npm test && npm run build)
(cd web/h5 && npm test && npm run build)
(cd web/admin && npm test && npm run build)
~~~

Expected: every command exits 0.

- [ ] **Step 2: Start the verified local stack**

Use the repository's existing development configuration and a disposable test database. Seed one superadmin, one station owner, one instructor with an own course, one second instructor with a foreign course, and learners covering required/optional/completed/in-progress/unstarted states.

- [ ] **Step 3: Exercise core flows in the in-app Browser**

Use browser:control-in-app-browser. Verify learner filters, recent learning, course tabs, material download, continue position, heartbeat increments, station group persistence, quick actions, categories, assignment changes, resource filters, and all three admin roles. Directly enter forbidden instructor routes and confirm the permission page without a business API call.

- [ ] **Step 4: Capture the five fixed reference states**

Use deviceScaleFactor 2 and these CSS viewports:

~~~text
Learner home:      1543 × 793
Course catalog:    1497 × 875
Course materials:  1489 × 790
Station collapsed: 1649 × 825
Station expanded:  1675 × 875
~~~

- [ ] **Step 5: Compare references and implementation together**

For each state, place the provided reference and implementation screenshot side by side in one comparison image. Inspect container edges, 48px/200px admin shell dimensions, 24px gaps, typography, card sizes, selected/open menu states, shadows, colors, and horizontal overflow. Fix every P0/P1/P2 mismatch and recapture.

- [ ] **Step 6: Check responsive states**

At 1440, 1024, and 390 CSS pixels, verify no horizontal overflow, the 76px compact sidebar, mobile drawer, single-column cards, 2×2 quick actions, readable resource legend, and left-aligned 1000px player.

- [ ] **Step 7: Write the QA record**

design-qa.md must list each viewport, interaction result, comparison iteration, resolved issues, automated commands, and end exactly with:

~~~text
final result: passed
~~~

- [ ] **Step 8: Inspect scope and commit QA fixes**

~~~bash
git status --short
git diff --check
git diff --stat
~~~

Expected: only planned feature and QA files. Commit:

~~~bash
git add design-qa.md
git commit -m "test: verify learner and station workspaces"
~~~

- [ ] **Step 9: Run the final verification once more**

~~~bash
go test ./...
(cd web/pc && npm test && npm run build)
(cd web/h5 && npm test && npm run build)
(cd web/admin && npm test && npm run build)
git status --short --branch
~~~

Expected: all checks pass and the isolated feature worktree is clean.
