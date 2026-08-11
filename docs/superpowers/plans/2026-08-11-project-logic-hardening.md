# Project Logic Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce the audited security and business invariants across authentication, tenant plans, learner access, progress reporting, user lifecycle, and domain provisioning, then publish an auditable release to Gitee and GitHub.

**Architecture:** Security-sensitive creation paths delegate to repository operations that enforce invariants atomically. Learner course access is centralized in `LearnerAccess`, PC and H5 share one heartbeat contract, and domain provisioning state is stored through a repository instead of process memory. Compatibility fields remain readable while one canonical business rule is used for every learner response.

**Tech Stack:** Go 1.24, Gin, GORM, PostgreSQL/SQLite tests, React 18, TypeScript 5.7, Vite 6, npm workspaces.

## Global Constraints

- Preserve `.codex/current-task.md`, `.codex/roadmap.md`, and `.codex/mockups/` as user-owned changes.
- Public registration creates learners only.
- Zero plan limits mean unlimited; missing tenant `plan_id` resolves to the active default plan.
- Required/optional classification comes from `courses.course_type` only.
- User removal deactivates the account and retains historical data.
- Existing API response shapes stay compatible where security does not require rejection.
- Every behavior change starts with a focused failing regression test.
- Push Gitee first and GitHub second; stop if Gitee fails.

---

### Task 1: Harden authentication and runtime secrets

**Files:**
- Modify: `internal/service/auth_credentials.go`
- Modify: `internal/service/auth.go`
- Modify: `internal/service/auth_test.go`
- Modify: `internal/api/auth.go`
- Modify: `internal/api/auth_test.go`
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Modify: `internal/repository/user.go`
- Modify: `internal/repository/user_gorm.go`
- Modify: `internal/repository/user_gorm_test.go`
- Modify: `internal/migration/migration.go`
- Modify: `cmd/server/dependencies.go`
- Modify: `cmd/server/main.go`
- Modify: `docker-compose.yml`

**Interfaces:**
- Produce: `func (service *AuthService) SetBootstrapSecret(secret string)`.
- Produce: `func (service *AuthService) BootstrapSuperadmin(ctx context.Context, email, name, password, bootstrapSecret string) (*domain.User, *TokenPair, error)`.
- Produce: `UserRepository.CreateFirstSuperadmin(context.Context, *domain.User) error`.
- Produce: `config.ValidateRuntimeSecrets(Config) error`.

- [ ] **Step 1: Add failing tests for privileged public registration**

Add table-driven service tests that call `RegisterWithPhone` with `tenant_admin`, `instructor`, and `superadmin` and assert code `40300`; assert `learner` remains creatable.

- [ ] **Step 2: Verify the registration tests fail**

Run:

```bash
go test -count=1 ./internal/service -run 'TestAuthServicePublicRegistration'
```

Expected: tenant administrator and instructor cases create users instead of returning 40300.

- [ ] **Step 3: Restrict public registration to learner**

Replace arbitrary tenant-role acceptance in `RegisterWithPhone` with an exact learner check. Keep staff creation in `UserService.CreateWithPhone` unchanged.

- [ ] **Step 4: Add failing bootstrap-secret and concurrent-create tests**

Add service tests for missing/wrong secret and a repository test that starts two `CreateFirstSuperadmin` calls and asserts exactly one row is created.

- [ ] **Step 5: Verify bootstrap tests fail**

Run:

```bash
go test -count=1 ./internal/service ./internal/repository -run 'Test(AuthServiceBootstrap|UserRepositoryCreateFirstSuperadmin)'
```

Expected: bootstrap does not validate a secret and the repository interface lacks atomic first creation.

- [ ] **Step 6: Implement bootstrap validation and database invariant**

Add `SUPERADMIN_BOOTSTRAP_SECRET` to `Config`, inject it into `AuthService`, accept `bootstrap_secret` in the bootstrap request, compare with `subtle.ConstantTimeCompare`, and call `CreateFirstSuperadmin`. Migration 22 creates a partial unique index allowing only one row where `role = 'superadmin' AND tenant_id IS NULL`.

- [ ] **Step 7: Add failing runtime-secret validation tests**

Test blank, `DefaultJWTSecret`, `imaiplay-compose-change-me`, and secrets shorter than 32 characters as invalid; test a 32+ character random-looking value as valid.

- [ ] **Step 8: Implement runtime validation and remove Compose fallbacks**

Call `config.ValidateRuntimeSecrets` from `main`. Change Compose to `${JWT_SECRET:?JWT_SECRET is required}` and `${DB_PASSWORD:?DB_PASSWORD is required}`.

- [ ] **Step 9: Run focused authentication/configuration tests and commit**

```bash
gofmt -w internal/service/auth_credentials.go internal/service/auth.go internal/service/auth_test.go internal/api/auth.go internal/api/auth_test.go internal/config/config.go internal/config/config_test.go internal/repository/user.go internal/repository/user_gorm.go internal/repository/user_gorm_test.go internal/migration/migration.go cmd/server/dependencies.go cmd/server/main.go
go test -count=1 ./internal/service ./internal/api ./internal/config ./internal/repository
git add internal cmd docker-compose.yml
git commit -m "fix: harden account bootstrap and public registration"
```

---

### Task 2: Centralize and serialize tenant plan limits

**Files:**
- Create: `internal/service/tenant_limits.go`
- Create: `internal/service/tenant_limits_test.go`
- Modify: `internal/service/user.go`
- Modify: `internal/service/user_test.go`
- Modify: `internal/service/course.go`
- Modify: `internal/service/course_test.go`
- Modify: `internal/service/resource.go`
- Modify: `cmd/server/dependencies.go`

**Interfaces:**
- Produce: `type TenantLimitService struct`.
- Produce: `ResolvePlan(context.Context, string) (*domain.Plan, error)`.
- Produce: `WithEmployeeSlot(context.Context, string, func() error) error`.
- Produce: `WithCourseSlot(context.Context, string, func() error) error`.
- Consume: existing tenant, plan, user, course, and resource repositories.

- [ ] **Step 1: Add failing default-plan employee tests**

Create a tenant without `PlanID`, an active default plan with `MaxUsers: 1`, and one active employee. Assert the next employee is rejected. Add a disabled employee case that does not consume a slot.

- [ ] **Step 2: Add failing course-limit tests**

Assert a tenant with `MaxCourses: 1` cannot create a second tenant course, while an official course is not counted.

- [ ] **Step 3: Verify limit tests fail**

```bash
go test -count=1 ./internal/service -run 'Test(TenantLimit|UserService.*DefaultPlan|CourseService.*Limit)'
```

- [ ] **Step 4: Implement a tenant-scoped serialized limit service**

Resolve explicit or default plan in one component. Use a keyed in-process mutex for the current single-server deployment and keep each count-plus-create callback inside the critical section. Count active tenant users only, count non-official tenant courses only, and preserve zero as unlimited.

- [ ] **Step 5: Route user, course, and resource capacity through the component**

Inject the component into `UserService`, `AuthService`, `CourseService`, and `ResourceService`. Keep existing constructors compatible through optional setters while production wiring always supplies the component.

- [ ] **Step 6: Run focused tests and commit**

```bash
gofmt -w internal/service/tenant_limits.go internal/service/tenant_limits_test.go internal/service/user.go internal/service/user_test.go internal/service/course.go internal/service/course_test.go internal/service/resource.go cmd/server/dependencies.go
go test -count=1 ./internal/service -run 'Test(TenantLimit|UserService|CourseService|ResourceService)'
git add internal/service cmd/server/dependencies.go
git commit -m "feat: enforce tenant plan limits consistently"
```

---

### Task 3: Make learner access and course type canonical

**Files:**
- Modify: `internal/service/learner_access.go`
- Modify: `internal/service/learner_access_test.go`
- Modify: `internal/service/progress.go`
- Modify: `internal/service/progress_test.go`
- Modify: `internal/service/enrollment.go`
- Modify: `internal/service/enrollment_test.go`
- Modify: `internal/repository/learner_overview_gorm.go`
- Modify: `internal/repository/learner_overview_gorm_test.go`

**Interfaces:**
- Produce: `LearnerAccess.RequireCourseEnrollment(context.Context, *domain.Course) (*domain.CourseEnrollment, error)`.
- Produce: `LearnerAccess.ResolveLessonCourse(context.Context, string) (*domain.Course, *domain.CourseEnrollment, error)`.

- [ ] **Step 1: Replace the existing ordinary-course auto-enrollment expectation**

Change the progress-get regression test to assert `40300` for a published ordinary course without enrollment. Add an enabled official-course case that creates an enrollment and returns zero progress.

- [ ] **Step 2: Verify the access test fails for the expected reason**

```bash
go test -count=1 ./internal/service -run 'TestProgressServiceGet(RejectsUnassignedCourse|StartsEnabledOfficialCourse)'
```

- [ ] **Step 3: Centralize lesson access**

Move the ordinary-versus-official decision into `LearnerAccess`; both `ProgressService.Get` and `ProgressService.Report` call it before reading or writing progress.

- [ ] **Step 4: Add failing course-type consistency tests**

Create an optional course with an enrollment whose legacy assignment type says required. Assert learner overview returns optional. Assert enrollment creation stores the course type regardless of request input.

- [ ] **Step 5: Remove enrollment-level business branching**

Use `course.CourseType` in overview, filters, and enrollment writes. Keep the column and update endpoint response compatible, but prevent it from overriding learner-visible classification.

- [ ] **Step 6: Run focused tests and commit**

```bash
gofmt -w internal/service/learner_access.go internal/service/learner_access_test.go internal/service/progress.go internal/service/progress_test.go internal/service/enrollment.go internal/service/enrollment_test.go internal/repository/learner_overview_gorm.go internal/repository/learner_overview_gorm_test.go
go test -count=1 ./internal/service ./internal/repository -run 'Test(ProgressService|LearnerAccess|EnrollmentService|LearnerOverview)'
git add internal/service internal/repository
git commit -m "fix: enforce learner enrollment and course type rules"
```

---

### Task 4: Unify heartbeat reporting and validate progress

**Files:**
- Create: `web/shared/src/watchHeartbeat.ts`
- Create: `web/shared/src/watchHeartbeat.test.ts`
- Modify: `web/shared/src/index.ts`
- Modify: `web/pc/src/utils/watchHeartbeat.ts`
- Modify: `web/pc/src/pages/LessonPlayerPage.tsx`
- Modify: `web/h5/src/api/progress.ts`
- Modify: `web/h5/src/api/progress.test.ts`
- Modify: `web/h5/src/pages/LessonPlayerPage.tsx`
- Modify: `internal/domain/learning_time.go`
- Modify: `internal/repository/learning_time.go`
- Modify: `internal/repository/learning_time_gorm.go`
- Modify: `internal/repository/learning_time_gorm_test.go`
- Modify: `internal/service/progress.go`
- Modify: `internal/service/progress_test.go`
- Modify: `internal/migration/migration.go`

**Interfaces:**
- Produce: shared `PlaybackLifecycleController` used by PC and H5.
- Extend heartbeat payload with `session_id` while retaining `report_id`.
- Produce: repository validation that bounds accumulated watched time by server-observed session age plus 5 seconds.

- [ ] **Step 1: Add failing H5 heartbeat parity tests**

Assert H5 emits the same 1–60 second heartbeat payload, stable retry report ID, visibility pause, and terminal flush behavior as PC.

- [ ] **Step 2: Verify H5 tests fail**

```bash
npm --prefix web/h5 test -- --run src/api/progress.test.ts
```

Expected: H5 omits heartbeat fields and lifecycle reporting.

- [ ] **Step 3: Move lifecycle control to shared and integrate both clients**

Export the existing tested PC controller from shared, replace the PC wrapper with a re-export, and wire H5 player events, visibility changes, periodic flush, pause, completion, and page hide to the controller.

- [ ] **Step 4: Add failing forged-time and progress-regression tests**

Assert a new session cannot report 60 seconds immediately, duplicate report IDs are idempotent, video percentage cannot exceed duration-derived progress, and a lower later position does not reduce saved progress.

- [ ] **Step 5: Implement server-side heartbeat bounds**

Migration 23 adds `session_id` and an index on tenant/user/session. Repository reads the first session report timestamp and rejects cumulative watched time beyond elapsed server time plus 5 seconds. Service clamps video progress using lesson duration and persists only monotonic maxima. Legacy position-only reports remain accepted but add no learning time.

- [ ] **Step 6: Run focused backend/frontend tests and commit**

```bash
gofmt -w internal/domain/learning_time.go internal/repository/learning_time.go internal/repository/learning_time_gorm.go internal/repository/learning_time_gorm_test.go internal/service/progress.go internal/service/progress_test.go internal/migration/migration.go
go test -count=1 ./internal/service ./internal/repository -run 'Test(ProgressService|LearningTime)'
npm --prefix web/shared test
npm --prefix web/pc test
npm --prefix web/h5 test
git add internal web/shared web/pc web/h5
git commit -m "feat: validate cross-device learning heartbeats"
```

---

### Task 5: Preserve history when removing users

**Files:**
- Modify: `internal/service/user.go`
- Modify: `internal/service/user_test.go`
- Modify: `internal/repository/user.go`
- Modify: `internal/repository/user_gorm.go`
- Modify: `internal/repository/user_gorm_test.go`

**Interfaces:**
- Produce: `UserRepository.Deactivate(context.Context, string) error`.
- Keep: `UserService.Delete(context.Context, string) error` API signature.

- [ ] **Step 1: Add a failing historical-retention test**

Create a learner, enrollment, and progress record; call `UserService.Delete`; assert the user row remains with `status = 0` and both learning records remain.

- [ ] **Step 2: Verify the test fails because the user row is deleted**

```bash
go test -count=1 ./internal/service ./internal/repository -run 'TestUserServiceDeleteDeactivatesAndRetainsHistory'
```

- [ ] **Step 3: Replace hard delete with deactivation**

Implement a tenant-scoped status update in the repository and have `UserService.Delete` call it. Authentication already rejects status other than 1.

- [ ] **Step 4: Run focused tests and commit**

```bash
gofmt -w internal/service/user.go internal/service/user_test.go internal/repository/user.go internal/repository/user_gorm.go internal/repository/user_gorm_test.go
go test -count=1 ./internal/service ./internal/repository -run 'Test(UserService|UserRepository)'
git add internal/service internal/repository
git commit -m "fix: retain learning history when removing users"
```

---

### Task 6: Persist resumable domain-binding jobs

**Files:**
- Create: `internal/domain/domain_bind_job.go`
- Create: `internal/repository/domain_bind_job.go`
- Create: `internal/repository/domain_bind_job_gorm.go`
- Create: `internal/repository/domain_bind_job_gorm_test.go`
- Modify: `internal/service/domain_bind.go`
- Modify: `internal/service/domain_bind_status.go`
- Modify: `internal/service/domain_bind_orchestrator.go`
- Modify: `internal/service/domain_bind_test.go`
- Modify: `internal/migration/migration.go`
- Modify: `cmd/server/repositories.go`
- Modify: `cmd/server/dependencies.go`

**Interfaces:**
- Produce: `DomainBindJobRepository.FindByTenant(context.Context, string) (*domain.DomainBindJob, error)`.
- Produce: `DomainBindJobRepository.Reserve(context.Context, *domain.DomainBindJob) error` with unique tenant and domain constraints.
- Produce: `DomainBindJobRepository.UpdateStatus(context.Context, string, string, string, int, string) error`.
- Produce: `DomainBindJobRepository.Delete(context.Context, string) error`.

- [ ] **Step 1: Add failing repository persistence tests**

Assert a job survives repository reconstruction, a domain cannot be reserved by two tenants, and status/step/error updates are durable.

- [ ] **Step 2: Verify repository tests fail**

```bash
go test -count=1 ./internal/repository -run 'TestDomainBindJobRepository'
```

- [ ] **Step 3: Add model, migration, and repository**

Migration 24 creates `domain_bind_jobs` with unique indexes on `tenant_id` and `domain`. Store state, step, safe error message, external site ID, attempt count, and timestamps.

- [ ] **Step 4: Add failing restart and retry service tests**

Start a job, reconstruct `DomainBindService` with the same repository, and assert status is restored. Simulate failure after site creation, then retry and assert the service checks the existing site and continues instead of creating a duplicate.

- [ ] **Step 5: Replace in-memory status ownership with the repository**

Keep only a narrow mutex preventing duplicate goroutines in one process. Every status transition writes the job. `Status` reads the job first; retry increments attempts and resumes from the last confirmed step. Unbind removes the external site and durable job after clearing the tenant domain.

- [ ] **Step 6: Run focused tests and commit**

```bash
gofmt -w internal/domain/domain_bind_job.go internal/repository/domain_bind_job.go internal/repository/domain_bind_job_gorm.go internal/repository/domain_bind_job_gorm_test.go internal/service/domain_bind.go internal/service/domain_bind_status.go internal/service/domain_bind_orchestrator.go internal/service/domain_bind_test.go internal/migration/migration.go cmd/server/repositories.go cmd/server/dependencies.go
go test -count=1 ./internal/repository ./internal/service -run 'TestDomainBind'
git add internal cmd/server
git commit -m "feat: persist resumable domain binding jobs"
```

---

### Task 7: Update release record, verify, and publish

**Files:**
- Create: `docs/updates/2026-08-11-project-logic-hardening.md`
- Modify: `README.md`

**Interfaces:**
- Produce: deployment notes for bootstrap secret, required Compose variables, migrations, behavior changes, and rollback considerations.

- [ ] **Step 1: Write the update record**

Document security changes, plan enforcement, learner access, heartbeat semantics, user deactivation, domain job persistence, new environment variables, and deployment order. Link it from the README update section.

- [ ] **Step 2: Run patch and migration hygiene checks**

```bash
git diff --check
git status --short
```

- [ ] **Step 3: Run fresh full backend verification**

```bash
go build ./cmd/server/
go test -count=1 ./...
go vet ./...
```

- [ ] **Step 4: Run fresh full frontend verification**

```bash
npm run test:all --prefix web
npm run build:all --prefix web
```

- [ ] **Step 5: Commit documentation**

```bash
git add README.md docs/updates/2026-08-11-project-logic-hardening.md
git commit -m "docs: record project logic hardening update"
```

- [ ] **Step 6: Verify only user-owned files remain uncommitted**

`git status --short` must list only `.codex/current-task.md`, `.codex/roadmap.md`, and `.codex/mockups/`.

- [ ] **Step 7: Push and verify Gitee before GitHub**

```bash
git push gitee main
git push origin main
LOCAL=$(git rev-parse main)
test "$LOCAL" = "$(git ls-remote gitee refs/heads/main | awk '{print $1}')"
test "$LOCAL" = "$(git ls-remote origin refs/heads/main | awk '{print $1}')"
```

Expected: local, Gitee, and GitHub `main` all point to the same final commit.
