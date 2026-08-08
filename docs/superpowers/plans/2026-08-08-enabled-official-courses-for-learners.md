# Enabled Official Courses for Learners Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every published official course enabled by a tenant immediately visible and accessible to that tenant's learners, defaulting to optional without bulk enrollment writes.

**Architecture:** Add one repository query dedicated to enabled published official courses, then merge that set into learner overview and course-list results with explicit enrollments taking precedence. Learner authorization accepts an enabled official course without enrollment; opening or reporting progress for it lazily creates one optional enrollment so existing progress and recent-learning flows continue to work.

**Tech Stack:** Go 1.x, GORM, SQLite-backed repository/service tests, existing React learner client (no frontend code change).

## Global Constraints

- Enabled official courses without an explicit enrollment use `domain.AssignmentOptional`.
- An existing active enrollment wins during deduplication and preserves its assignment type.
- Tenant-owned courses still require their existing enrollment behavior.
- Disabled, unactivated, draft, deleted, and cross-tenant official courses remain hidden and unauthorized.
- Disabling a course must not delete enrollments or lesson progress.
- Add no database migration and perform no bulk writes when an administrator enables a course.
- Preserve all unrelated dirty-worktree changes, including `internal/service/course_test.go`.

---

### Task 1: Query Enabled Published Official Courses

**Files:**
- Modify: `internal/repository/course.go`
- Modify: `internal/repository/course_gorm.go`
- Test: `internal/repository/course_gorm_test.go`

**Interfaces:**
- Produces: `CourseRepository.FindEnabledOfficialByTenant(ctx context.Context, tenantID string) ([]domain.Course, error)`.
- Ordering contract: title ascending, then ID ascending, so every consumer receives deterministic results.

- [ ] **Step 1: Write the failing repository test**

Extend the official activation fixture with published enabled, disabled, unactivated, draft, and another-tenant activation cases, then assert literal IDs:

```go
items, err := repo.FindEnabledOfficialByTenant(context.Background(), "tenant-a")
if err != nil {
    t.Fatalf("FindEnabledOfficialByTenant() error = %v", err)
}
if got := courseIDs(items); !reflect.DeepEqual(got, []string{"enabled-a", "enabled-b"}) {
    t.Fatalf("enabled official IDs = %#v", got)
}
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `go test ./internal/repository -run TestCourseRepositoryFindEnabledOfficialByTenant -count=1`

Expected: build failure because `FindEnabledOfficialByTenant` is not defined.

- [ ] **Step 3: Add the interface method and minimal GORM query**

Implement a query equivalent to:

```go
func (repo *courseGORMRepository) FindEnabledOfficialByTenant(
    ctx context.Context, tenantID string,
) ([]domain.Course, error) {
    items := make([]domain.Course, 0)
    err := repo.database.WithContext(ctx).
        Where("tenant_id = ? AND is_official = ? AND status = ?", "", true, 1).
        Where("id IN (SELECT course_id FROM tenant_official_courses WHERE tenant_id = ? AND enabled = ?)", tenantID, true).
        Order("title ASC, id ASC").
        Find(&items).Error
    return items, err
}
```

- [ ] **Step 4: Run the focused test and verify GREEN**

Run: `go test ./internal/repository -run TestCourseRepositoryFindEnabledOfficialByTenant -count=1`

Expected: PASS.

- [ ] **Step 5: Commit the repository slice**

```bash
git add internal/repository/course.go internal/repository/course_gorm.go internal/repository/course_gorm_test.go
git commit -m "feat: query enabled official courses"
```

### Task 2: Merge Official Courses into Learner Overview

**Files:**
- Modify: `internal/repository/learner_overview_gorm.go`
- Test: `internal/repository/learner_overview_gorm_test.go`

**Interfaces:**
- Consumes: existing `aggregateCourse(...)` and the `tenant_official_courses` visibility rule.
- Produces: overview rows where unassigned enabled official courses have `AssignmentType == domain.AssignmentOptional`.

- [ ] **Step 1: Write the failing overview tests**

Remove the explicit enrollment for the enabled official fixture and retain these assertions:

```go
officialGot := byID[official.ID]
if officialGot.Course.ID == "" || officialGot.AssignmentType != domain.AssignmentOptional {
    t.Fatalf("automatic official course = %#v", officialGot)
}
```

Add a separately enabled official course with an active required enrollment and assert it occurs once with `domain.AssignmentRequired`. Keep literal absence assertions for disabled, unactivated, draft, and foreign activations.

- [ ] **Step 2: Run the focused test and verify RED**

Run: `go test ./internal/repository -run TestLearnerOverviewRepositoryAggregatesOnlyActiveVisibleAssignments -count=1`

Expected: FAIL because the enabled official course without enrollment is missing.

- [ ] **Step 3: Implement merge and deduplication**

Track every course ID produced from active enrollments. Query remaining published official courses enabled for `tenantID`, aggregate each with `domain.AssignmentOptional`, and append only IDs not already present:

```go
seen := make(map[string]struct{}, len(enrollments))
// Existing enrollment loop aggregates visible rows and records seen[course.ID].
// Then load enabled official rows and aggregate unseen rows as optional.
```

Keep the existing title/ID sort after both sources have been merged.

- [ ] **Step 4: Run the focused repository tests and verify GREEN**

Run: `go test ./internal/repository -run 'TestLearnerOverviewRepository|TestCourseRepositoryFindEnabledOfficialByTenant' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit the overview slice**

```bash
git add internal/repository/learner_overview_gorm.go internal/repository/learner_overview_gorm_test.go
git commit -m "fix: show enabled official courses in learner overview"
```

### Task 3: Align Course List and Learner Authorization

**Files:**
- Modify: `internal/service/course.go`
- Modify: `internal/service/learner_access.go`
- Create: `internal/service/official_course_visibility_test.go`
- Test: `internal/service/learner_access_test.go`

**Interfaces:**
- Consumes: `CourseRepository.FindEnabledOfficialByTenant`.
- Produces: `ListPublished` returns the union of active enrolled visible courses and enabled published official courses, sorted and paginated after deduplication.
- Produces: `AuthorizeCourse` permits an enabled published official course without enrollment, while tenant courses still require active enrollment.

- [ ] **Step 1: Write failing list and authorization tests**

In the new test file, create an enabled official course without enrollment and assert:

```go
items, total, err := service.ListPublished(learner, 0, 20)
if err != nil || total != 1 || len(items) != 1 || items[0].ID != official.ID {
    t.Fatalf("ListPublished() = %#v, %d, %v", items, total, err)
}
```

Add an active required enrollment for the same official course and assert it remains a single list item. In `learner_access_test.go`, remove the enabled official enrollment and assert course, lesson resource, and material authorization succeed; disabled and unactivated official cases remain 404.

- [ ] **Step 2: Run focused service tests and verify RED**

Run: `go test ./internal/service -run 'TestCourseServiceListsEnabledOfficialWithoutEnrollment|TestLearnerAccessAuthorizationMatrix' -count=1`

Expected: FAIL because the list omits and authorization rejects the unassigned enabled official course.

- [ ] **Step 3: Implement list union and official authorization**

In `ListPublished`, build a `map[string]domain.Course` from active visible enrollments, then merge `FindEnabledOfficialByTenant` results only when absent. Convert to a slice, apply the existing title/ID sort, then paginate.

In `AuthorizeCourse`, load the course through `FindPublishedByID` first. Return it immediately when `course.IsOfficial`; otherwise require an active enrollment using the existing not-found/internal-error mapping.

- [ ] **Step 4: Run focused service tests and verify GREEN**

Run: `go test ./internal/service -run 'TestCourseServiceListsEnabledOfficialWithoutEnrollment|TestLearnerAccessAuthorizationMatrix' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit list and authorization changes**

```bash
git add internal/service/course.go internal/service/learner_access.go internal/service/official_course_visibility_test.go internal/service/learner_access_test.go
git commit -m "fix: allow learners to access enabled official courses"
```

### Task 4: Lazily Create Optional Enrollment on First Learning Action

**Files:**
- Modify: `internal/service/progress.go`
- Test: `internal/service/progress_test.go`

**Interfaces:**
- Consumes: `domain.Course.IsOfficial`, `domain.AssignmentOptional`, and existing `startCourse` concurrency retry behavior.
- Produces: the first progress read or report for an enabled official course creates one active optional enrollment; existing enrollment types are unchanged.

- [ ] **Step 1: Write failing progress tests**

Create a published enabled official course with one lesson and no enrollment. Call `Get`, then assert the stored enrollment is optional:

```go
_, err := progress.Get(learner, officialLesson.ID)
if err != nil {
    t.Fatalf("Get(enabled official) error = %v", err)
}
enrollment, err := enrollmentRepo.FindByCourseAndUser(learner, official.ID, learnerID)
if err != nil || enrollment.AssignmentType != domain.AssignmentOptional {
    t.Fatalf("official enrollment = %#v, %v", enrollment, err)
}
```

Use a second learner to call `Report` directly and assert it also creates an optional enrollment. Assert reporting an unassigned tenant course remains forbidden and an existing required official enrollment remains required.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `go test ./internal/service -run 'TestProgressServiceStartsEnabledOfficialAsOptional|TestProgressServiceReportsEnabledOfficialWithoutEnrollment' -count=1`

Expected: FAIL because `Get` creates a required enrollment and `Report` rejects the missing enrollment.

- [ ] **Step 3: Implement official lazy enrollment**

Change `startCourse` to receive `*domain.Course`, set `AssignmentType` explicitly, and preserve its conflict re-read:

```go
assignmentType := domain.AssignmentRequired
if course.IsOfficial {
    assignmentType = domain.AssignmentOptional
}
```

In `Report`, keep the active-enrollment check. Only when the enrollment is missing and `course.IsOfficial` is true, call `startCourse(ctx, course, userID, tenantID)` and continue; missing tenant-course enrollment still returns `not enrolled in this course`.

- [ ] **Step 4: Run progress tests and verify GREEN**

Run: `go test ./internal/service -run 'TestProgressService' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit progress behavior**

```bash
git add internal/service/progress.go internal/service/progress_test.go
git commit -m "fix: start official courses as optional"
```

### Task 5: Full Verification

**Files:**
- Verify only; no planned production changes.

**Interfaces:**
- Verifies every acceptance criterion and guards unrelated behavior.

- [ ] **Step 1: Format touched Go files**

Run:

```bash
gofmt -w internal/repository/course.go internal/repository/course_gorm.go internal/repository/course_gorm_test.go internal/repository/learner_overview_gorm.go internal/repository/learner_overview_gorm_test.go internal/service/course.go internal/service/learner_access.go internal/service/learner_access_test.go internal/service/official_course_visibility_test.go internal/service/progress.go internal/service/progress_test.go
```

- [ ] **Step 2: Run all Go tests**

Run: `go test ./... -count=1`

Expected: PASS with zero failures.

- [ ] **Step 3: Run frontend tests and production builds**

Run:

```bash
npm --prefix web/pc test -- --run
npm --prefix web/pc run build
npm --prefix web/admin test -- --run
npm --prefix web/admin run build
```

Expected: all commands exit 0.

- [ ] **Step 4: Check diff scope and whitespace**

Run:

```bash
git diff --check
git status --short
git diff --stat
```

Expected: no whitespace errors; only planned files plus pre-existing user changes appear.

- [ ] **Step 5: Final behavior audit**

Confirm from tests that enabled official courses appear without enrollment, explicit assignments win, disabled/draft/cross-tenant courses remain hidden, first learning creates an optional enrollment, and tenant-course behavior has not changed.
