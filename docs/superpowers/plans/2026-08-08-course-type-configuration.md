# Course Type Configuration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make “必修课 / 选修课” an explicit required course property, remove per-learner type decisions, and show both course category and course type consistently on PC and H5 learner pages.

**Architecture:** Add a validated `course_type` field to `courses` as the single source of truth. Keep the enrollment field for database compatibility, but copy the course type into newly created enrollments and stop using enrollment type for learner filtering, cards, and required-course statistics. Extend existing admin forms and learner API mappings rather than adding a parallel configuration system.

**Tech Stack:** Go, Gin, GORM, PostgreSQL-compatible migrations, React, TypeScript, Ant Design, Ant Design Mobile, Node test runner, Vitest, Vite.

## Global Constraints

- User-facing values use the complete labels “必修课” and “选修课”.
- New tenant and official courses require an explicit choice; no UI default is selected.
- Existing courses migrate to `required`.
- Course categories remain independent and continue to drive category filtering.
- PC and H5 cards and detail pages show category when present and always show course type.
- No new runtime dependency is introduced.

---

### Task 1: Course type domain model and migration

**Files:**
- Modify: `internal/domain/course.go`
- Modify: `internal/migration/migration.go`
- Test: `internal/migration/migration_test.go`

**Interfaces:**
- Produces: `domain.CourseTypeRequired`, `domain.CourseTypeOptional`, and `Course.CourseType string` serialized as `course_type`.
- Produces: migrated historical rows with `course_type = required`.

- [ ] **Step 1: Write the failing migration test**

Add assertions that migration creates `courses.course_type` and converts a pre-feature course row to `required`:

```go
if !database.Migrator().HasColumn(&domain.Course{}, "CourseType") {
    t.Fatal("courses.course_type missing")
}
var migrated domain.Course
if err := database.First(&migrated, "id = ?", legacyCourse.ID).Error; err != nil {
    t.Fatal(err)
}
if migrated.CourseType != domain.CourseTypeRequired {
    t.Fatalf("course type = %q, want required", migrated.CourseType)
}
```

- [ ] **Step 2: Run the migration test and verify failure**

Run: `go test ./internal/migration -run 'Test.*Migration' -count=1`

Expected: FAIL because `Course.CourseType` and its column do not exist.

- [ ] **Step 3: Add the domain constants, field, and backfill**

Add:

```go
const (
    CourseTypeRequired = "required"
    CourseTypeOptional = "optional"
)

CourseType string `gorm:"not null;default:required" json:"course_type"`
```

After auto-migration, execute a scoped backfill:

```go
database.Model(&domain.Course{}).
    Where("course_type IS NULL OR course_type = ?", "").
    Update("course_type", domain.CourseTypeRequired)
```

- [ ] **Step 4: Re-run migration tests**

Run: `go test ./internal/migration -count=1`

Expected: PASS.

- [ ] **Step 5: Commit the domain and migration change**

```bash
git add internal/domain/course.go internal/migration/migration.go internal/migration/migration_test.go
git commit -m "feat: add course type field"
```

---

### Task 2: Course service and management APIs require course type

**Files:**
- Modify: `internal/service/course.go`
- Modify: `internal/api/course.go`
- Modify: `internal/repository/course_gorm.go`
- Test: `internal/service/course_test.go`
- Test: `internal/service/course_category_test.go`
- Test: `internal/api/course_test.go`
- Test: `internal/api/course_superadmin_test.go`

**Interfaces:**
- Consumes: `domain.CourseTypeRequired` and `domain.CourseTypeOptional`.
- Produces: create/update service calls that validate and persist `courseType string`.
- Produces: management JSON accepting required `course_type`.

- [ ] **Step 1: Add failing service tests for valid and invalid types**

Cover tenant and official course creation and update:

```go
course, err := service.CreateWithCategory(ctx, "Title", "", "", nil, domain.CourseTypeOptional)
if err != nil || course.CourseType != domain.CourseTypeOptional {
    t.Fatalf("course = %#v, err = %v", course, err)
}
if _, err := service.CreateWithCategory(ctx, "Bad", "", "", nil, "recommended"); errorCode(err) != 40000 {
    t.Fatalf("invalid type error = %#v", err)
}
```

- [ ] **Step 2: Add failing API tests for missing and invalid `course_type`**

Use request bodies containing `"course_type":"required"`, omit it once, and send `"course_type":"recommended"` once. Expect valid requests to succeed and invalid requests to return HTTP 400.

- [ ] **Step 3: Run focused backend tests and verify failure**

Run: `go test ./internal/service ./internal/api -run 'Course|Official' -count=1`

Expected: FAIL because service signatures and request fields do not support course type.

- [ ] **Step 4: Implement validation and persistence**

Add one validator:

```go
func validateCourseType(value string) (string, error) {
    if value != domain.CourseTypeRequired && value != domain.CourseTypeOptional {
        return "", errorsx.BadRequest("invalid course type")
    }
    return value, nil
}
```

Thread `courseType` through tenant and official create/update paths, assign `course.CourseType`, and include `course_type` with `binding:"required"` in create and update request structs. Update repository field maps so updates persist it.

- [ ] **Step 5: Run focused backend tests**

Run: `go test ./internal/service ./internal/api ./internal/repository -run 'Course|Official' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit management API changes**

```bash
git add internal/service/course.go internal/api/course.go internal/repository/course_gorm.go internal/service/course_test.go internal/service/course_category_test.go internal/api/course_test.go internal/api/course_superadmin_test.go
git commit -m "feat: require course type in management APIs"
```

---

### Task 3: Make course type authoritative for learner data and enrollment creation

**Files:**
- Modify: `internal/repository/learner_overview.go`
- Modify: `internal/repository/learner_overview_gorm.go`
- Modify: `internal/service/learner_overview.go`
- Modify: `internal/service/course.go`
- Modify: `internal/service/progress.go`
- Test: `internal/repository/learner_overview_gorm_test.go`
- Test: `internal/service/learner_overview_test.go`
- Test: `internal/service/progress_test.go`
- Test: `internal/api/learner_overview_test.go`
- Test: `internal/api/student_test.go`

**Interfaces:**
- Produces: learner overview `course.course_type` and category object.
- Produces: learner course detail `course_type` and category name.
- Keeps `assignment_type` only as compatibility output during transition.

- [ ] **Step 1: Write failing learner overview tests**

Create a course where `CourseType` differs from enrollment `AssignmentType` and assert filtering/statistics follow the course:

```go
course.CourseType = domain.CourseTypeOptional
enrollment.AssignmentType = domain.AssignmentRequired
// RequiredTotal must remain zero and returned type must be optional.
```

Also assert the learner detail response contains `course_type` and a populated category summary.

- [ ] **Step 2: Write a failing automatic enrollment test**

Start an optional course without a prior enrollment and assert the created compatibility enrollment has `AssignmentType == domain.CourseTypeOptional`.

- [ ] **Step 3: Run learner-focused tests and verify failure**

Run: `go test ./internal/repository ./internal/service ./internal/api -run 'Learner|Overview|Progress|Published' -count=1`

Expected: FAIL because enrollment type is still authoritative and H5 detail lacks category/type metadata.

- [ ] **Step 4: Switch learner presentation to `Course.CourseType`**

Populate `LearnerCourseSummary.AssignmentType` from `item.Course.CourseType`, calculate required totals using `course.Course.CourseType`, and expose `course_type` in the H5 published list/detail response. Resolve category display name using the existing course category repository and return it with detail data.

- [ ] **Step 5: Make automatic enrollment inherit course type**

In `ProgressService.startCourse`, load the course before creating the compatibility enrollment and set:

```go
AssignmentType: course.CourseType,
```

Do not change authorization or duplicate-enrollment handling.

- [ ] **Step 6: Run learner-focused tests**

Run: `go test ./internal/repository ./internal/service ./internal/api -run 'Learner|Overview|Progress|Published' -count=1`

Expected: PASS.

- [ ] **Step 7: Commit learner data-flow changes**

```bash
git add internal/repository/learner_overview.go internal/repository/learner_overview_gorm.go internal/service/learner_overview.go internal/service/course.go internal/service/progress.go internal/repository/learner_overview_gorm_test.go internal/service/learner_overview_test.go internal/service/progress_test.go internal/api/learner_overview_test.go internal/api/student_test.go
git commit -m "fix: derive learner course type from course"
```

---

### Task 4: Update tenant and official course admin forms

**Files:**
- Modify: `web/admin/src/api/course.ts`
- Modify: `web/admin/src/api/officialCourse.ts`
- Modify: `web/admin/src/pages/Courses.tsx`
- Modify: `web/admin/src/pages/OfficialCourses.tsx`
- Modify: `web/admin/src/pages/CourseDetail.tsx`
- Modify: `web/admin/src/utils/adminFormValues.ts`
- Test: `web/admin/tests/officialCourses.test.ts`
- Test: `web/admin/tests/courseDetailModel.test.ts`
- Create: `web/admin/tests/courseTypeForm.test.ts`

**Interfaces:**
- Consumes: management API field `course_type: 'required' | 'optional'`.
- Produces: mandatory unselected-on-create course type form values.
- Produces: learner assignment requests containing only `user_id`.

- [ ] **Step 1: Add failing static and model tests**

Assert both course editors contain a required `course_type` field with labels “必修课” and “选修课”, new-course defaults leave it undefined, and `CourseDetail.tsx` no longer contains the “分配类型” column or form item.

- [ ] **Step 2: Run admin tests and verify failure**

Run: `cd web/admin && npm test -- --runInBand`

Expected: FAIL because course type fields are absent and enrollment type UI remains.

- [ ] **Step 3: Add API types and form controls**

Define:

```ts
export type CourseType = 'required' | 'optional'
```

Add `course_type` to `Course` and input payloads. In tenant and official editors render:

```tsx
<Form.Item
  label="课程类型"
  name="course_type"
  rules={[{ required: true, message: '请选择课程类型' }]}
>
  <Select options={[
    { value: 'required', label: '必修课' },
    { value: 'optional', label: '选修课' },
  ]} />
</Form.Item>
```

For new courses set `course_type: undefined`; for edits prefill the saved value.

- [ ] **Step 4: Remove per-learner type editing**

Remove `assignment_type` from the enrollment form, table, and create/update UI calls. Keep removal and learner-selection behavior unchanged.

- [ ] **Step 5: Run admin tests and build**

Run: `cd web/admin && npm test && npm run build`

Expected: PASS.

- [ ] **Step 6: Commit admin changes**

```bash
git add web/admin/src web/admin/tests
git commit -m "feat: configure course type in admin"
```

---

### Task 5: Show category and course type on PC learner pages

**Files:**
- Modify: `web/pc/src/api/learner.ts`
- Modify: `web/pc/src/components/CourseCard.tsx`
- Modify: `web/pc/src/pages/CourseDetailPage.tsx`
- Modify: `web/pc/src/styles/course.css`
- Test: `web/pc/src/api/learner.test.ts`
- Test: `web/pc/src/utils/learnerCourses.test.ts`
- Modify: `web/pc/tests/selectionTheme.test.ts`

**Interfaces:**
- Consumes: learner course `course_type` and `category`.
- Produces: card/detail labels with `course-category-tag` and `course-type-tag`.

- [ ] **Step 1: Add failing PC mapping and rendering assertions**

Update fixtures to return `course_type`. Assert the mapper rejects unsupported values and maps valid values. Add source/CSS assertions that category and type labels are both rendered and use separate classes.

- [ ] **Step 2: Run PC tests and verify failure**

Run: `cd web/pc && npm test`

Expected: FAIL because mapping still reads enrollment type and the category tag is not rendered.

- [ ] **Step 3: Implement PC mapping and labels**

Rename the UI model property to `courseType`, parse `course.course_type`, update filtering to use it, and render:

```tsx
{course.category && <Tag className="course-category-tag">{course.category.name}</Tag>}
<Tag className="course-type-tag">
  {course.courseType === 'required' ? '必修课' : '选修课'}
</Tag>
```

Use the same labels on the detail page and keep the compact card layout.

- [ ] **Step 4: Run PC tests and build**

Run: `cd web/pc && npm test && npm run build`

Expected: PASS.

- [ ] **Step 5: Commit PC changes**

```bash
git add web/pc/src web/pc/tests
git commit -m "feat: show learner course category and type"
```

---

### Task 6: Show category and course type on H5 learner pages

**Files:**
- Modify: `web/h5/src/types/course.ts`
- Modify: `web/h5/src/api/course.ts`
- Modify: `web/h5/src/components/CourseCard.tsx`
- Modify: `web/h5/src/pages/CourseDetailPage.tsx`
- Modify: `web/h5/src/styles.css`
- Test: `web/h5/src/api/course.test.ts`
- Create: `web/h5/src/components/CourseCard.test.tsx`

**Interfaces:**
- Consumes: published course API `course_type` and category summary.
- Produces: the same complete Chinese labels as PC.

- [ ] **Step 1: Add failing H5 API and card tests**

Assert mapped courses preserve `courseType` and category name, then render a card fixture and assert both visible labels are present.

- [ ] **Step 2: Run H5 tests and verify failure**

Run: `cd web/h5 && npm test`

Expected: FAIL because the H5 model hardcodes category and has no course type.

- [ ] **Step 3: Implement H5 mapping and rendering**

Add `courseType: 'required' | 'optional'` and optional category data to the course model. Remove the hardcoded `category: '企业课程'`. Render category when present and always render the complete course type label on card and detail pages.

- [ ] **Step 4: Run H5 tests and build**

Run: `cd web/h5 && npm test && npm run build`

Expected: PASS.

- [ ] **Step 5: Commit H5 changes**

```bash
git add web/h5/src
git commit -m "feat: show mobile course category and type"
```

---

### Task 7: Full regression verification

**Files:**
- Verify only; modify a file only when a failing regression demonstrates a defect in the preceding tasks.

**Interfaces:**
- Consumes: all preceding task outputs.
- Produces: a clean, deployable branch.

- [ ] **Step 1: Run backend validation**

Run: `go test ./... -count=1 && go vet ./...`

Expected: PASS.

- [ ] **Step 2: Run shared and frontend validation**

Run each command:

```bash
(cd web/shared && npm test && npm run build)
(cd web/admin && npm test && npm run build)
(cd web/pc && npm test && npm run build)
(cd web/h5 && npm test && npm run build)
```

Expected: all commands PASS; existing Vite chunk-size warnings are non-blocking.

- [ ] **Step 3: Check formatting and worktree state**

Run:

```bash
gofmt -w internal/domain internal/migration internal/service internal/api internal/repository
git diff --check
git status --short
```

Expected: no whitespace errors and only intended files changed.

- [ ] **Step 4: Present the completed branch for integration**

Report the commit list, test results, and exact deployment implication: historical courses start as “必修课” and can be edited afterward.
