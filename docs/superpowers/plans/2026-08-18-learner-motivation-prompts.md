# Learner Motivation Prompts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build server-driven, once-per-account welcome prompts and once-per-day learner summaries for PC and H5, using real tenant-scoped learning data and direct course continuation actions.

**Architecture:** Add a tenant-scoped learner engagement state plus a read model that aggregates yesterday's learning. A dedicated motivation service returns one discriminated prompt model and accepts idempotent presentation acknowledgements. Shared TypeScript validates the contract; PC and H5 render platform-specific accessible surfaces without blocking their home-page data.

**Tech Stack:** Go, Gin, GORM, SQLite/PostgreSQL-compatible migrations, React 18, TypeScript 5.7, Ant Design 5, Ant Design Mobile 5, Node test runner, Vitest, CSS custom properties.

**Spec:** `docs/superpowers/specs/2026-08-14-learner-motivation-prompts-design.md`

## Global Constraints

- New learner means the account's first successful tenant login; historical learners are backfilled as already welcomed.
- Welcome appears once per account; daily summary or re-engagement appears at most once per Asia/Shanghai date across PC, H5, and devices.
- Password, SMS-code, and organization-selection login paths share one first-login write path.
- Peer comparison is tenant-local, names no peers, and appears only with at least 10 active learners yesterday.
- Every displayed metric comes from persisted learning data; do not invent streaks, points, badges, certificates, or percentages.
- Motivation failures never block login, home-page data, or course navigation.
- PC and H5 colors must use existing learner CSS variables so admin theme changes propagate.
- Do not add npm dependencies or bypass the existing bundle budgets.
- Preserve the user's unrelated `.codex/` worktree changes and never stage them.

## File Structure

### Backend

- Create `internal/domain/learner_engagement.go`: persisted prompt state.
- Create `internal/repository/learner_motivation.go`: repository interfaces and snapshot types.
- Create `internal/repository/learner_motivation_gorm.go`: state transitions, statistics, recommendations, prompt keys.
- Create `internal/repository/learner_motivation_gorm_test.go`: tenant isolation, prompt frequency, aggregate and percentile tests.
- Create `internal/service/learner_motivation.go`: prompt selection and copy rules.
- Create `internal/service/learner_motivation_test.go`: deterministic state-machine tests.
- Create `internal/api/learner_motivation.go`: GET and acknowledgement handlers.
- Create `internal/api/learner_motivation_test.go`: handler role, envelope, and invalid-key tests.
- Modify `internal/migration/migration.go` and `internal/migration/migration_test.go`: v25 schema plus historical-account backfill.
- Modify `internal/repository/user_gorm.go` and `internal/repository/user_gorm_test.go`: create engagement state in the learner-account transaction.
- Modify `internal/service/auth.go`, `internal/service/auth_login.go`, and `internal/service/auth_test.go`: mark first successful tenant login once.
- Modify `internal/server/server.go`, `internal/server/routes.go`, `internal/server/routes_learner.go`, `cmd/server/dependencies.go`, and server/integration fixtures: dependency and route wiring.

### Shared web contract

- Create `web/shared/src/learning/learnerMotivation.ts`: strict raw-to-view-model normalization and formatting helpers.
- Create `web/shared/tests/learnerMotivation.test.ts`: invalid payload, duration, comparison, and route-target tests.
- Modify `web/shared/package.json` and `web/shared/src/index.ts`: export the shared module.

### PC

- Create `web/pc/src/api/motivation.ts` and `web/pc/src/api/motivation.test.ts`: GET/ack adapter.
- Create `web/pc/src/components/LearnerMotivationPrompt.tsx`: accessible Ant Design modal and navigation.
- Create `web/pc/tests/motivationPromptWiring.test.ts`: source-level component wiring and accessibility contract.
- Modify `web/pc/src/pages/HomePage.tsx`: mount only after overview success.
- Modify `web/pc/src/styles/dashboard.css`, `web/pc/src/styles/responsive.css`, and `web/pc/tests/learnerStyles.test.ts`: minimal themed modal styles and reduced motion.

### H5

- Create `web/h5/src/api/motivation.ts` and `web/h5/src/api/motivation.test.ts`: GET/ack adapter.
- Create `web/h5/src/components/LearnerMotivationPrompt.tsx`: bottom popup and navigation.
- Modify `web/h5/src/pages/HomePage.tsx`: mount only after required course data succeeds.
- Modify `web/h5/src/styles.css`, `web/h5/tests/learnerStyles.test.ts`, and `web/h5/tests/learnerPageWiring.test.ts`: safe-area layout, themed styles, wiring, reduced motion.

---

### Task 1: Persist Learner Engagement State and Backfill Historical Accounts

**Files:**
- Create: `internal/domain/learner_engagement.go`
- Modify: `internal/migration/migration.go`
- Modify: `internal/migration/migration_test.go`
- Modify: `internal/repository/user_gorm.go`
- Modify: `internal/repository/user_gorm_test.go`

**Interfaces:**
- Produces: `domain.LearnerEngagementState` with `FirstLoginAt`, `WelcomeSeenAt`, `LastDailyPromptDate`, `PendingPromptKey`, `PendingPromptKind`, `PendingPromptDate`, and `PendingPromptExpiresAt`.
- Produces: v25 table `learner_engagement_states`, unique index `(tenant_id, user_id)`, and historical learner backfill.
- Consumes later: repository and auth tasks query this state by tenant/user.

- [ ] **Step 1: Write the failing migration and user-creation tests**

Add assertions equivalent to:

```go
if !database.Migrator().HasTable(&domain.LearnerEngagementState{}) {
    t.Fatal("learner engagement state table missing")
}
if !database.Migrator().HasIndex(&domain.LearnerEngagementState{}, "idx_learner_engagement_user") {
    t.Fatal("learner engagement user index missing")
}
```

Seed one existing learner before v25, run `AutoMigrate`, and assert `WelcomeSeenAt != nil`. In `user_gorm_test.go`, create a new learner and assert one state row with `WelcomeSeenAt == nil`; create an instructor and assert no state row.

- [ ] **Step 2: Run the focused tests and verify RED**

Run:

```bash
go test ./internal/migration ./internal/repository -run 'Test.*(Migration|UserRepository).*Engagement' -count=1
```

Expected: FAIL because `LearnerEngagementState` and v25 do not exist.

- [ ] **Step 3: Add the domain model and v25 migration**

Create:

```go
type LearnerEngagementState struct {
    BaseModel
    UserID               string     `gorm:"index;not null"`
    FirstLoginAt         *time.Time
    WelcomeSeenAt        *time.Time
    LastDailyPromptDate  string     `gorm:"size:10;not null;default:''"`
    PendingPromptKey     string     `gorm:"size:80;not null;default:''"`
    PendingPromptKind    string     `gorm:"size:24;not null;default:''"`
    PendingPromptDate    string     `gorm:"size:10;not null;default:''"`
    PendingPromptExpiresAt *time.Time
}
```

Register v25, `AutoMigrate` the model, create `idx_learner_engagement_user`, then backfill existing `role='learner'` users with both `FirstLoginAt` and `WelcomeSeenAt` set to the migration time. Use GORM iteration rather than dialect-specific UUID SQL.

- [ ] **Step 4: Make learner account creation transactional**

In `userGORMRepository.Create`, wrap tenant-user creation in `database.Transaction`. After `tx.Create(user)`, create a blank engagement state only when `user.Role == "learner"`. Preserve the existing special superadmin `NULL tenant_id` branch.

- [ ] **Step 5: Run tests and commit**

Run:

```bash
go test ./internal/migration ./internal/repository -count=1
git diff --check
```

Commit:

```bash
git add internal/domain/learner_engagement.go internal/migration/migration.go internal/migration/migration_test.go internal/repository/user_gorm.go internal/repository/user_gorm_test.go
git commit -m "feat(learner): persist motivation prompt state"
```

### Task 2: Track the First Successful Tenant Login

**Files:**
- Create: `internal/repository/learner_motivation.go`
- Create: `internal/repository/learner_motivation_gorm.go`
- Create: `internal/repository/learner_motivation_gorm_test.go`
- Modify: `internal/service/auth.go`
- Modify: `internal/service/auth_login.go`
- Modify: `internal/service/auth_test.go`
- Modify: `cmd/server/dependencies.go`

**Interfaces:**
- Produces: `LearnerMotivationRepository.MarkFirstLogin(ctx context.Context, userID string, at time.Time) (bool, error)`.
- Produces: `AuthService.SetLearnerMotivationRepository(repository.LearnerMotivationRepository)`.
- Consumes: the engagement row created in Task 1.

- [ ] **Step 1: Write repository atomicity and auth-path tests**

Repository test: call `MarkFirstLogin` twice and assert results `true`, then `false`, with the original timestamp preserved. Add auth tests for scoped password login, SMS login, and `SelectTenant`; each must leave exactly one non-null `first_login_at`. Assert tenant admins do not create learner prompt behavior.

- [ ] **Step 2: Run focused tests and verify RED**

```bash
go test ./internal/repository ./internal/service -run 'Test.*FirstLogin' -count=1
```

Expected: FAIL because the repository and setter are missing.

- [ ] **Step 3: Implement tenant-bound atomic marking**

Define the repository interface with `MarkFirstLogin` plus the future methods declared in Tasks 3–4. Implement `MarkFirstLogin` as one tenant/user-scoped conditional update:

```go
result := db.Model(&domain.LearnerEngagementState{}).
    Where("tenant_id = ? AND user_id = ? AND first_login_at IS NULL", tenantID, userID).
    Update("first_login_at", at.UTC())
return result.RowsAffected == 1, result.Error
```

Reject missing learner identity; do not accept tenant or user identifiers from request bodies.

- [ ] **Step 4: Wire all tenant login paths through one call**

Add an optional motivation repository to `AuthService`. In `completeTenantLogin`, before issuing tokens, call `MarkFirstLogin` only for `role == "learner"`. An injected repository error fails the login with `errorsx.Internal("record learner login failed")`; retry remains safe because the update is idempotent. Configure it in `buildServerDependencies`.

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/repository ./internal/service -count=1
git diff --check
git add internal/repository/learner_motivation.go internal/repository/learner_motivation_gorm.go internal/repository/learner_motivation_gorm_test.go internal/service/auth.go internal/service/auth_login.go internal/service/auth_test.go cmd/server/dependencies.go
git commit -m "feat(auth): record learner first login"
```

### Task 3: Build the Motivation Snapshot and Prompt State Machine

**Files:**
- Modify: `internal/repository/learner_motivation.go`
- Modify: `internal/repository/learner_motivation_gorm.go`
- Modify: `internal/repository/learner_motivation_gorm_test.go`
- Create: `internal/service/learner_motivation.go`
- Create: `internal/service/learner_motivation_test.go`

**Interfaces:**
- Produces: `LearnerMotivationRepository.Load(ctx, today, yesterday, dayBefore string, yesterdayStart, todayStart time.Time) (repository.LearnerMotivationSnapshot, error)`.
- Produces: `IssuePrompt(ctx, kind, promptDate string, expiresAt time.Time) (string, error)` and `AcknowledgePrompt(ctx, key string, at time.Time) error`.
- Produces: `LearnerMotivationService.Get(context.Context) (LearnerMotivation, error)` and `Acknowledge(context.Context, string) error`.

- [ ] **Step 1: Write snapshot tests using real GORM rows**

Seed two tenants, yesterday/day-before daily stats, duplicate lesson reports, completion timestamps, accessible and inaccessible courses, and 9/10/11 active learner cohorts. Assert:

```go
if got.YesterdayLessonCount != 2 || got.YesterdayCompletedLessonCount != 1 {
    t.Fatalf("yesterday lesson metrics = %#v", got)
}
if got.ActiveLearnerCount == 10 && got.ExceededPercent != 70 {
    t.Fatalf("percentile = %d", got.ExceededPercent)
}
```

Include equal-duration peers and assert ties are not counted as exceeded.

- [ ] **Step 2: Write service state-machine tests**

Use a stub repository to cover `welcome`, `daily_summary`, `reengagement`, and `none`. Fix `service.now` to `2026-08-18 09:00 Asia/Shanghai`; assert queried dates are `2026-08-18`, `2026-08-17`, `2026-08-16`. Assert ranking is absent below 10 active learners and copy does not mention zero-learning guilt.

- [ ] **Step 3: Run focused tests and verify RED**

```bash
go test ./internal/repository ./internal/service -run 'TestLearnerMotivation' -count=1
```

Expected: FAIL because snapshot and service methods are absent.

- [ ] **Step 4: Implement snapshot queries and recommendation reuse**

Define snapshot fields for state, yesterday/day-before seconds, distinct lesson count, completed lesson/course counts, active cohort count, exceeded percent, required totals, and one `RecommendedCourse`. Filter all queries by context tenant/user. Join lessons → chapters → currently visible courses; exclude disabled/unassigned content. Select recommendations in the exact priority from the spec and return the first unfinished lesson plus last position.

- [ ] **Step 5: Implement deterministic prompt selection and copy**

Define:

```go
type LearnerMotivation struct {
    Kind string `json:"kind"`
    PromptKey string `json:"prompt_key,omitempty"`
    StudyDate string `json:"study_date,omitempty"`
    Title string `json:"title,omitempty"`
    Message string `json:"message,omitempty"`
    Metrics *LearnerMotivationMetrics `json:"metrics,omitempty"`
    Comparison *LearnerMotivationComparison `json:"comparison,omitempty"`
    Course *LearnerMotivationCourse `json:"course,omitempty"`
}
```

For `welcome`, require `FirstLoginAt != nil && WelcomeSeenAt == nil`. For later states, return `none` when `LastDailyPromptDate == today`. Issue/reuse a UUID prompt key for the same kind/date with a 24-hour expiry. Use strictly derived Chinese copy and omit zero/unknown comparison fields.

- [ ] **Step 6: Implement acknowledgement semantics**

`AcknowledgePrompt` must match current tenant/user, non-empty key, pending kind/date, and non-expired key. For `welcome`, set `welcome_seen_at`; for daily kinds set `last_daily_prompt_date`. Leave the matching key available so repeated acknowledgement succeeds; replace it only when issuing a later prompt.

- [ ] **Step 7: Run tests and commit**

```bash
go test ./internal/repository ./internal/service -count=1
git diff --check
git add internal/repository/learner_motivation.go internal/repository/learner_motivation_gorm.go internal/repository/learner_motivation_gorm_test.go internal/service/learner_motivation.go internal/service/learner_motivation_test.go
git commit -m "feat(learner): generate daily motivation prompts"
```

### Task 4: Expose and Wire the Learner Motivation API

**Files:**
- Create: `internal/api/learner_motivation.go`
- Create: `internal/api/learner_motivation_test.go`
- Modify: `internal/server/server.go`
- Modify: `internal/server/routes.go`
- Modify: `internal/server/routes_learner.go`
- Modify: `cmd/server/dependencies.go`
- Modify: `internal/server/server_test.go`
- Modify: `internal/test/integration/core_flows_test.go`

**Interfaces:**
- Produces: `GET /api/v1/learner/motivation`.
- Produces: `POST /api/v1/learner/motivation/ack` with `{ "prompt_key": string }`.
- Consumes: `LearnerMotivationService.Get` and `.Acknowledge` from Task 3.

- [ ] **Step 1: Write handler and route contract tests**

Assert learner role receives the response envelope, tenant admin receives 403, malformed/blank acknowledgement returns 400, and a valid acknowledgement returns success. Extend server route smoke tests to require both paths.

- [ ] **Step 2: Run focused tests and verify RED**

```bash
go test ./internal/api ./internal/server -run 'Test.*(Motivation|Routes)' -count=1
```

Expected: FAIL because the handler and routes are missing.

- [ ] **Step 3: Implement handler and dependency wiring**

Create an API interface with `Get` and `Acknowledge`. Both handlers call `requireHandlerRole(c, "learner")`; acknowledgement uses a binding-required `PromptKey string`. Add `LearnerMotivationService` to `server.Dependencies`, construct one repository instance in `newRepositories`, inject it into Auth and the motivation service, and register both learner routes.

- [ ] **Step 4: Add end-to-end tenant contract coverage**

In integration tests, create a new learner, perform login, GET `welcome`, acknowledge it, then assert the next GET is not `welcome`. Seed yesterday stats for an existing learner and assert `daily_summary`; use another tenant's larger stats to prove isolation.

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/api ./internal/server ./internal/test/integration -count=1
git diff --check
git add internal/api/learner_motivation.go internal/api/learner_motivation_test.go internal/server/server.go internal/server/routes.go internal/server/routes_learner.go cmd/server/dependencies.go internal/server/server_test.go internal/test/integration/core_flows_test.go
git commit -m "feat(api): expose learner motivation prompts"
```

### Task 5: Add the Shared TypeScript Motivation Contract

**Files:**
- Create: `web/shared/src/learning/learnerMotivation.ts`
- Create: `web/shared/tests/learnerMotivation.test.ts`
- Modify: `web/shared/package.json`
- Modify: `web/shared/src/index.ts`

**Interfaces:**
- Produces: `normalizeLearnerMotivation(raw: unknown): LearnerMotivation`.
- Produces: discriminated union `kind: 'none' | 'welcome' | 'daily_summary' | 'reengagement'`.
- Produces: `formatLearningDuration(seconds: number): string` and `motivationTargetPath(prompt): string | undefined`.
- Produces: `acknowledgeAndContinue(acknowledge, continuation): Promise<void>`, which always runs the continuation after the acknowledgement settles.

- [ ] **Step 1: Write strict normalizer tests**

Cover all four kinds and reject unknown kind, blank prompt key for visible prompts, negative seconds, progress outside 0–100, comparison outside 0–99, and a visible primary action missing course/lesson IDs. Assert `none` normalizes to `{ kind: 'none' }`.

- [ ] **Step 2: Run the shared tests and verify RED**

```bash
cd web && npm run test -w @imaiplay/shared -- learnerMotivation
```

Expected: FAIL because the module export is missing.

- [ ] **Step 3: Implement the union and helpers**

Use runtime type guards over `unknown`; do not cast server payloads directly. Map snake_case fields to camelCase. Format `<60s` as `不足 1 分钟`, exact minutes as `N 分钟`, and longer values as `H 小时 M 分钟`. Return `/courses/{courseId}/lessons/{lessonId}` only when both IDs are present. Implement `acknowledgeAndContinue` with `try { await acknowledge() } finally { continuation() }` and test both resolved and rejected acknowledgements.

- [ ] **Step 4: Export, test, and commit**

```bash
cd web && npm run test -w @imaiplay/shared && npm run build -w @imaiplay/shared
cd .. && git diff --check
git add web/shared/src/learning/learnerMotivation.ts web/shared/tests/learnerMotivation.test.ts web/shared/package.json web/shared/src/index.ts
git commit -m "feat(shared): validate learner motivation prompts"
```

### Task 6: Render the Accessible PC Motivation Modal

**Files:**
- Create: `web/pc/src/api/motivation.ts`
- Create: `web/pc/src/api/motivation.test.ts`
- Create: `web/pc/src/components/LearnerMotivationPrompt.tsx`
- Create: `web/pc/tests/motivationPromptWiring.test.ts`
- Modify: `web/pc/src/pages/HomePage.tsx`
- Modify: `web/pc/src/styles/dashboard.css`
- Modify: `web/pc/src/styles/responsive.css`
- Modify: `web/pc/tests/learnerStyles.test.ts`

**Interfaces:**
- Consumes: shared `LearnerMotivation`, normalizer, duration formatter, and target path.
- Produces: `getLearnerMotivation()` and `acknowledgeLearnerMotivation(promptKey)`.
- Produces: `<LearnerMotivationPrompt enabled />` mounted after overview success.

- [ ] **Step 1: Write API and component wiring tests**

API tests assert GET/POST paths and shared normalization. The Node source contract asserts that the component: returns `null` for `none`; acknowledges from Ant Design `afterOpenChange`; keeps close independent from acknowledgement success; runs CTA through shared `acknowledgeAndContinue`; routes through `portalRoutePath`; catches GET failures without rendering an error result; and exposes a labelled modal title and close control. This avoids adding a DOM test dependency.

- [ ] **Step 2: Write style contract tests**

Assert `.learner-motivation-modal` uses only `var(--learner-*)`, width remains in the 520–620px range, metric rows are not nested cards, primary action uses tenant contrast variables, and reduced motion disables prompt translation.

- [ ] **Step 3: Run PC tests and verify RED**

```bash
cd web/pc && npm test
```

Expected: FAIL because API/component/style selectors do not exist.

- [ ] **Step 4: Implement the API and modal state machine**

Fetch once when `enabled` becomes true. Keep prompt request errors local. Render Ant Design `Modal` with `afterOpenChange(open => open && ackOnce())`; close sets local hidden state even if ack rejects. CTA derives the course route, attempts acknowledgement, and always navigates in `finally`.

- [ ] **Step 5: Implement the three PC templates and styling**

Use one 560px surface with title/message, at most four inline metrics, optional comparison sentence, compact course row, primary CTA, and text close action. Use existing icons from `@ant-design/icons`; all colors and shadows come from learner variables. Add focus-visible and reduced-motion rules.

- [ ] **Step 6: Mount after overview success, test, and commit**

Mount beside the successful home content with `enabled={!loading && !error && Boolean(overview)}`. Run:

```bash
cd web/pc && npm test && npm run build
cd ../.. && git diff --check
git add web/pc/src/api/motivation.ts web/pc/src/api/motivation.test.ts web/pc/src/components/LearnerMotivationPrompt.tsx web/pc/src/pages/HomePage.tsx web/pc/src/styles/dashboard.css web/pc/src/styles/responsive.css web/pc/tests/learnerStyles.test.ts web/pc/tests/motivationPromptWiring.test.ts
git commit -m "feat(learner): add PC motivation prompts"
```

### Task 7: Render the H5 Motivation Bottom Popup

**Files:**
- Create: `web/h5/src/api/motivation.ts`
- Create: `web/h5/src/api/motivation.test.ts`
- Create: `web/h5/src/components/LearnerMotivationPrompt.tsx`
- Modify: `web/h5/src/pages/HomePage.tsx`
- Modify: `web/h5/src/styles.css`
- Modify: `web/h5/tests/learnerStyles.test.ts`
- Modify: `web/h5/tests/learnerPageWiring.test.ts`

**Interfaces:**
- Consumes: the same shared prompt contract and target path as PC.
- Produces: an Ant Design Mobile `Popup` adapted to safe areas and touch navigation.

- [ ] **Step 1: Write API, wiring, and style tests**

Assert GET/POST normalization; Home mounts the prompt only after required courses load; `.learner-motivation-popup` applies `env(safe-area-inset-bottom)`; CTA has at least 44px height; prompt motion is disabled in reduced-motion; all new colors are CSS variables.

- [ ] **Step 2: Run H5 tests and verify RED**

```bash
cd web/h5 && npm test
```

Expected: FAIL because H5 prompt files and selectors are missing.

- [ ] **Step 3: Implement the H5 adapter and popup**

Use `Popup` with `position="bottom"`, rounded top corners, scrollable content, safe-area padding, a two-column metric grid, and a full-width primary button. The close/back action and CTA use the same acknowledge-once function as PC; ack failure never prevents closing or navigation through `theme.routePath`.

- [ ] **Step 4: Mount after required data succeeds, test, and commit**

Do not mount during loading or `loadError`. Run:

```bash
cd web/h5 && npm test && npm run build
cd ../.. && git diff --check
git add web/h5/src/api/motivation.ts web/h5/src/api/motivation.test.ts web/h5/src/components/LearnerMotivationPrompt.tsx web/h5/src/pages/HomePage.tsx web/h5/src/styles.css web/h5/tests/learnerStyles.test.ts web/h5/tests/learnerPageWiring.test.ts
git commit -m "feat(learner): add H5 motivation prompts"
```

### Task 8: Run Cross-Platform Release Verification

**Files:**
- Modify only if a verification failure exposes an in-scope defect in Tasks 1–7.

**Interfaces:**
- Consumes: all backend, shared, PC, and H5 deliverables.
- Produces: release evidence with no ignored failing command.

- [ ] **Step 1: Run backend tests from a clean command**

```bash
go test -count=1 ./...
```

Expected: exit 0.

- [ ] **Step 2: Run all web tests, builds, and bundle budgets**

```bash
cd web && npm run verify:all
```

Expected: Shared, Admin, PC, and H5 tests pass; all builds pass; every JavaScript chunk stays within the existing 500000-byte budget; no circular-chunk or size warning appears.

- [ ] **Step 3: Run source and worktree checks**

```bash
cd ..
git diff --check
rg -n '#[0-9a-fA-F]{3,8}\b|rgba?\(|hsla?\(' web/pc/src/styles web/h5/src/styles.css
git status --short
```

Expected: diff check is empty; new prompt CSS contains no color literal; only intentionally preserved user `.codex/` changes remain uncommitted.

- [ ] **Step 4: Review the full feature diff against the spec**

```bash
git diff b295806..HEAD --stat
git log --oneline b295806..HEAD
```

Confirm every spec section maps to implementation/tests and no admin setting, gamification, notification, or unrelated refactor was added.

Do not create an empty verification commit. If verification exposes an in-scope defect, return to the task that owns the defect, repeat its focused red/green cycle, and use that task's explicit commit scope.
