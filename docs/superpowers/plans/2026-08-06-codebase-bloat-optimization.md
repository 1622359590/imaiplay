# Codebase Bloat Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reduce repository, frontend, and backend maintenance bloat without changing public APIs, database state, page visuals, or user workflows.

**Architecture:** Execute the refactor in green phases: repository hygiene, a framework-neutral frontend shared package, frontend hotspot decomposition, backend composition decomposition, and service decomposition. Existing public modules remain compatibility adapters while implementation moves behind focused files.

**Tech Stack:** Go 1.22, Gin, GORM, React 18, TypeScript 5.7, Vite 6.2, npm workspaces, Node test runner, Vitest.

## Global Constraints

- Keep API paths, parameters, response shapes, `errorsx` codes, database schema, visual output, and workflows unchanged.
- Keep PC, H5, and admin UI separate; share only framework-neutral infrastructure and types.
- Preserve auth validation timing, token rotation, tenant privacy, logout invalidation, and 401/403 behavior.
- Preserve domain-binding locks, async behavior, SSL policy, ownership checks, rollback, and audit semantics.
- Preserve all 114 backend routes, methods, optional-service guards, and middleware ordering.
- Finish every task with focused tests and a dedicated commit.
- Work only in `/Users/imaiwork/Documents/imaiplay-go/.worktrees/learner-station-workspaces` on `codex/codebase-bloat-optimization`.

---

### Task 1: Remove process artifacts and prevent root binaries

**Files:**
- Modify: `.gitignore`, `design-qa.md`
- Delete: `.qa-artifacts/login-layout-fix/comparison.png`, `.qa-artifacts/login-layout-fix/source-normalized-1672x848.png`
- Delete: `.qa-artifacts/unified-coral/admin-palette-comparison.png`, `.qa-artifacts/unified-coral/admin-workspace-mobile-before.png`, `.qa-artifacts/unified-coral/detail-comparison.png`, `.qa-artifacts/unified-coral/h5-detail-mobile-before.png`
- Delete: every pre-existing `docs/superpowers/plans/*.md` except this plan

**Interfaces:**
- Consumes: current ignore rules and final QA references.
- Produces: root-only `/server` ignore and final-only QA evidence.

- [ ] **Step 1: Verify the checks fail before cleanup**

```bash
git check-ignore server
rg -n 'before|comparison|source-normalized' design-qa.md .qa-artifacts
```

Expected: `server` is not ignored and process artifacts are reported.

- [ ] **Step 2: Add the exact ignore rule**

```gitignore
# Local Go binary produced from the repository root.
/server
```

- [ ] **Step 3: Update QA references and delete process files**

Keep only these login references in `design-qa.md`:

```markdown
- Fixed desktop implementation: `.qa-artifacts/login-layout-fix/pc-login-fixed-1672x848.png`.
- Fixed mobile implementation: `.qa-artifacts/login-layout-fix/pc-login-fixed-390x844.png`.
```

Delete the six process images listed above. Keep login `fixed`, official-course-picker `final`, and final unified-coral desktop/mobile images.

- [ ] **Step 4: Delete completed plans and verify**

```bash
git ls-files 'docs/superpowers/plans/*.md' | grep -v '2026-08-06-codebase-bloat-optimization.md'
git check-ignore server
test -z "$(rg -l 'before|comparison|source-normalized' design-qa.md || true)"
git diff --check
```

Delete the files printed by the first command. Expected: subsequent checks pass and `docs/superpowers/specs/` is unchanged.

- [ ] **Step 5: Commit**

```bash
git add .gitignore design-qa.md .qa-artifacts docs/superpowers/plans
git commit -m "chore: remove obsolete repository artifacts"
```

---

### Task 2: Establish the npm workspace and shared error package

**Files:**
- Create: `web/package.json`, `web/package-lock.json`
- Create: `web/shared/package.json`, `web/shared/tsconfig.json`, `web/shared/src/index.ts`, `web/shared/src/api/errors.ts`, `web/shared/tests/errors.test.ts`
- Modify: `web/admin/package.json`, `web/pc/package.json`, `web/h5/package.json`, `Dockerfile`
- Delete: the three application-level package lockfiles

**Interfaces:**
- Produces: `@imaiplay/shared`, `responseStatus(error): number | undefined`, `responseMessage(error): string | undefined`, `test:all`, and `build:all`.

- [ ] **Step 1: Write the failing shared test**

```ts
import assert from 'node:assert/strict'
import test from 'node:test'
import { responseMessage, responseStatus } from '../src/api/errors.ts'

test('reads an Axios-compatible response', () => {
  const error = { response: { status: 403, data: { message: '租户已暂停' } } }
  assert.equal(responseStatus(error), 403)
  assert.equal(responseMessage(error), '租户已暂停')
})

test('does not invent fields for unknown errors', () => {
  assert.equal(responseStatus(new Error('network')), undefined)
  assert.equal(responseMessage(new Error('network')), undefined)
})
```

- [ ] **Step 2: Create workspace manifests**

`web/package.json`:

```json
{
  "name": "imaiplay-web",
  "private": true,
  "workspaces": ["shared", "admin", "pc", "h5"],
  "scripts": {
    "test:all": "npm run test -w @imaiplay/shared && npm run test -w imaiplay-admin && npm run test -w imaiplay-pc && npm run test -w imaiplay-h5",
    "build:all": "npm run build -w imaiplay-admin && npm run build -w imaiplay-pc && npm run build -w imaiplay-h5"
  }
}
```

`web/shared/package.json` exports `.`, `./api/errors`, and `./auth/sessionCore`; its scripts are `node --test tests/*.test.ts` and `tsc --noEmit -p tsconfig.json`. Use strict ES2020/Bundler/noEmit TypeScript settings.

- [ ] **Step 3: Verify the test fails, then implement errors**

```bash
cd web && npm install --ignore-scripts
npm run test -w @imaiplay/shared
```

Expected: FAIL before `errors.ts` exists. Implement:

```ts
type ResponseLike = { response?: { status?: unknown; data?: { message?: unknown } } }
export const responseStatus = (error: unknown) => {
  const value = (error as ResponseLike | null)?.response?.status
  return typeof value === 'number' ? value : undefined
}
export const responseMessage = (error: unknown) => {
  const value = (error as ResponseLike | null)?.response?.data?.message
  return typeof value === 'string' && value.trim() ? value : undefined
}
```

- [ ] **Step 4: Align compatible versions and produce one lockfile**

Set all apps to Axios `^1.8.4`, TypeScript `~5.7.2`, Vite `^6.2.4`, React `^18.3.1`, and matching React types. Add `"@imaiplay/shared": "0.1.0"`. Keep PC on React Router 6, admin/H5 on Router 7, and keep Ant Design versus Ant Design Mobile separate.

Delete app lockfiles and run `cd web && npm install --package-lock-only --ignore-scripts`.

- [ ] **Step 5: Update Docker installation**

```dockerfile
WORKDIR /src
COPY web/package*.json web/
COPY web/shared/package.json web/shared/
COPY web/admin/package.json web/admin/
COPY web/pc/package.json web/pc/
COPY web/h5/package.json web/h5/
RUN cd web && npm ci
COPY web/ web/
RUN cd web && npm run build:all
```

- [ ] **Step 6: Verify and commit**

```bash
cd web && npm ci
npm run test -w @imaiplay/shared
npm run build -w @imaiplay/shared
npm ls --workspaces --depth=0
git add Dockerfile web
git commit -m "build(web): introduce shared npm workspace"
```

---

### Task 3: Extract shared session refresh coordination

**Files:**
- Create: `web/shared/src/auth/sessionCore.ts`, `web/shared/tests/sessionCore.test.ts`
- Modify: `web/shared/src/index.ts`
- Modify: all three `src/api/authSession.ts` files and their tests

**Interfaces:**
- Produces: `StorageLike`, `RefreshSession`, `markSessionChanged`, `decodeJwtPayload`, and `createRefreshCoordinator(options): () => Promise<string>`.

- [ ] **Step 1: Write shared concurrency tests**

Use this contract in `sessionCore.test.ts`:

```ts
const refresh = createRefreshCoordinator({
  storage,
  accessTokenKey: 'access',
  refreshTokenKey: 'refresh',
  request: async () => response,
  validateAccessToken: (token) => token.startsWith('valid-'),
  supersededError: () => new Error('superseded'),
})
```

Test that concurrent calls issue one request, rotated tokens are stored, `markSessionChanged(storage)` blocks a stale refresh after logout, and invalid access tokens are rejected without storage.

- [ ] **Step 2: Run the failing test**

```bash
cd web && npm run test -w @imaiplay/shared
```

Expected: FAIL because the session exports do not exist.

- [ ] **Step 3: Implement the core types and algorithm**

```ts
export interface StorageLike {
  getItem(key: string): string | null
  setItem(key: string, value: string): void
  removeItem(key: string): void
}
export interface RefreshSession { token: string; refresh_token?: string }
export interface RefreshCoordinatorOptions {
  storage: StorageLike
  accessTokenKey: string
  refreshTokenKey: string
  request(refreshToken: string): Promise<RefreshSession>
  validateAccessToken(token: string): boolean
  supersededError(): Error
}
```

Use a module-level `WeakMap<object, number>`. Deduplicate by generation plus refresh-token text, recheck both before writing, rotate the refresh token when supplied, and clear only the matching pending promise in `finally`.

- [ ] **Step 4: Convert each app to a compatibility adapter**

Keep every exported app key, type, function, and custom superseded-error class. Replace JWT payload decoding, generation bookkeeping, and refresh coordination with shared functions. App writes and clears call `markSessionChanged` at the same points as the former generation increment.

- [ ] **Step 5: Verify and commit**

```bash
cd web
npm run test -w @imaiplay/shared
npm run test -w imaiplay-admin
npm run test -w imaiplay-pc
npm run test -w imaiplay-h5
git add shared admin/src/api/authSession.ts admin/tests/authSession.test.ts pc/src/api/authSession.ts pc/tests/authSession.test.ts h5/src/api/authSession.ts h5/tests/authSession.test.ts
git commit -m "refactor(web): share session refresh coordination"
```

---

### Task 4: Share API errors, contracts, and tenant theme primitives

**Files:**
- Modify: all three `src/api/client.ts` files and their auth/client tests
- Create: `web/shared/src/types/api.ts`, `web/shared/src/types/theme.ts`, `web/shared/src/theme/tenantTheme.ts`
- Create: `web/shared/tests/tenantTheme.test.ts`
- Modify: all three `src/api/theme.ts` files and their theme providers/tests

**Interfaces:**
- Consumes: `responseStatus` and `responseMessage` from `@imaiplay/shared/api/errors`.
- Produces: `ApiEnvelope<T>`, `PageResult<T>`, `TenantThemeContract`, `TenantPortalContract`, and `normalizePrimaryColor`; unchanged interceptors, redirects, events, retry behavior, and visible messages.

- [ ] **Step 1: Add adapter regression assertions**

```ts
assert.equal(responseStatus({ response: { status: 401 } }), 401)
assert.equal(responseStatus({ response: { status: 403 } }), 403)
assert.equal(responseMessage({ response: { data: { message: '请求失败' } } }), '请求失败')
```

Retain tests proving admin persistent auth clears on 401/403 and learner sessions never clear admin keys.

- [ ] **Step 2: Replace local response-shape parsing**

Import the shared helpers. Keep navigation, event dispatch, toast copy, retry flags, refresh calls, and tenant Headers inside each app.

- [ ] **Step 3: Add common API and theme contracts**

Define the shared contracts without presentation-specific fields:

```ts
export interface ApiEnvelope<T> { data: T }
export interface PageResult<T> { items?: T[]; list?: T[]; data?: T[]; total?: number }
export interface TenantThemeContract {
  primary_color: string
  logo_url?: string
  welcome_text?: string
  browser_title?: string
}
export interface TenantPortalContract extends TenantThemeContract {
  tenant_id: string
  code: string
  name: string
  default_portal_url: string
  custom_domain_url?: string
}
```

`normalizePrimaryColor(value, fallback)` returns a trimmed `#RRGGBB` value or the supplied fallback. Add tests for uppercase, whitespace, invalid hex, and empty input. Re-export shared contracts from each app's existing API modules so current imports remain valid; do not unify PC/H5 presentation-level `Course` models.

- [ ] **Step 4: Verify and commit**

```bash
cd web && npm run test:all
git add shared admin/src admin/tests pc/src pc/tests h5/src h5/tests
git commit -m "refactor(web): share API and theme primitives"
```

---

### Task 5: Split and flatten the PC stylesheet

**Files:**
- Modify: `web/pc/src/styles.css`, `web/pc/tests/learnerStyles.test.ts`
- Create: `web/pc/src/styles/base.css`, `layout.css`, `login.css`, `dashboard.css`, `course.css`, `player.css`, `responsive.css`
- Create: `web/pc/tests/styleSource.ts`

**Interfaces:**
- Consumes: current selectors and cascade order.
- Produces: identical computed styles through seven responsibility files and an import-only entrypoint.

- [ ] **Step 1: Make tests resolve CSS imports**

```ts
import { readFileSync } from 'node:fs'
export function readStyleBundle(entry: URL, seen = new Set<string>()): string {
  if (seen.has(entry.href)) return ''
  seen.add(entry.href)
  return readFileSync(entry, 'utf8').replace(
    /@import\s+["']([^"']+)["'];?/g,
    (_, path: string) => readStyleBundle(new URL(path, entry), seen),
  )
}
```

Update `learnerStyles.test.ts` to use `readStyleBundle(new URL('../src/styles.css', import.meta.url))`.

- [ ] **Step 2: Add a failing entrypoint test**

```ts
test('style entrypoint contains only ordered imports', () => {
  const entry = readFileSync(new URL('../src/styles.css', import.meta.url), 'utf8')
  assert.deepEqual(entry.trim().split('\n'), [
    "@import './styles/base.css';",
    "@import './styles/layout.css';",
    "@import './styles/login.css';",
    "@import './styles/dashboard.css';",
    "@import './styles/course.css';",
    "@import './styles/player.css';",
    "@import './styles/responsive.css';",
  ])
})
```

Run `cd web && npm run test -w imaiplay-pc -- --test-name-pattern='style entrypoint'`. Expected: FAIL.

- [ ] **Step 3: Move rules by ownership**

Move variables/resets/typography/animations/utilities to `base.css`; shell/navigation/container rules to `layout.css`; login selectors to `login.css`; summary/filter/home/card selectors to `dashboard.css`; detail/chapter/lesson/material selectors to `course.css`; video/document/progress selectors to `player.css`; and media queries to `responsive.css`. Set `styles.css` to the seven imports asserted above.

- [ ] **Step 4: Remove superseded layers**

For `.course-card`, `.course-cover`, `.login-page`, `.login-card`, `.lesson-row`, and `.course-grid`, retain final unified-coral/screenshot-matched declarations, merge non-overridden properties into the owner rule, and delete prior dark/design-system/simplified duplicates. Add a test requiring one top-level owner definition outside media queries.

- [ ] **Step 5: Verify and commit**

```bash
cd web
npm run test -w imaiplay-pc
npm run build -w imaiplay-pc
wc -l pc/src/styles.css pc/src/styles/*.css
git add pc/src/styles.css pc/src/styles pc/tests
git commit -m "refactor(pc): split and flatten learner styles"
```

---

### Task 6: Decompose the admin course detail page

**Files:**
- Modify: `web/admin/src/pages/CourseDetail.tsx`
- Create: `web/admin/src/pages/course-detail/CourseDetailPage.tsx`, `useCourseDetail.ts`, `courseDetailModel.ts`, `CourseSummary.tsx`, `EnrollmentManager.tsx`, `ChapterEditor.tsx`, `LessonEditor.tsx`, `CourseOutline.tsx`, `ResourcePreviewModal.tsx`
- Create: `web/admin/tests/courseDetailModel.test.ts`

**Interfaces:**
- Produces: `lessonPayload(values): Omit<Lesson, 'id'>`, `CourseDetailController`, and the unchanged default route export.

- [ ] **Step 1: Write a failing mapping test**

```ts
assert.deepEqual(lessonPayload({
  title: '视频课', content_type: 'video', content_url: 'stale',
  resource_id: 'resource-1', duration_seconds: 0, sort_order: 0,
}), {
  title: '视频课', content_type: 'video', content_url: '',
  resource_id: 'resource-1', duration_seconds: 0, sort_order: 0,
})
```

Add a text case that preserves `content_url` and makes `resource_id` undefined. Run admin tests and expect failure before the helper exists.

- [ ] **Step 2: Extract the model and controller**

Move `Editor`, `LessonForm`, and the exact save mapping to `courseDetailModel.ts`. Define:

```ts
export interface CourseDetailController {
  course?: Course
  loading: boolean
  saving: boolean
  officialMode: boolean
  instructor: boolean
  resources: Resource[]
  matchingResources: Resource[]
  enrollments: CourseEnrollment[]
  learners: User[]
  editor?: Editor
  selectedResource?: UploadedMedia
  previewTarget?: UploadedMedia
  preview?: ResourcePreview
  previewLoading: boolean
  enrollmentOpen: boolean
  form: FormInstance<LessonForm>
  enrollmentForm: FormInstance<{ user_id: string; assignment_type: AssignmentType }>
  reload(): Promise<void>
  edit(editor: Editor): void
  closeEditor(): void
  save(): Promise<void>
  uploadResource(file: File, onProgress: (percent: number) => void): Promise<Resource>
  previewResource(resource: UploadedMedia): Promise<void>
  closePreview(): void
  enroll(): Promise<void>
  changeAssignment(enrollmentID: string, assignmentType: AssignmentType): Promise<void>
  removeEnrollment(enrollmentID: string): Promise<void>
  removeChapter(chapterID: string): Promise<void>
  removeLesson(chapterID: string, lessonID: string): Promise<void>
  setEnrollmentOpen(open: boolean): void
}
```

The hook owns loading, resources, enrollment data, editor state, preview lifecycle, forms, and API actions with unchanged messages.

- [ ] **Step 3: Extract presentational sections**

Move summary, enrollment table, outline, edit modals, and preview modal into the listed components. Keep `CourseMaterialsManager` independent. Replace the old entry with:

```ts
export { default } from './course-detail/CourseDetailPage'
```

- [ ] **Step 4: Verify and commit**

```bash
cd web
npm run test -w imaiplay-admin
npm run build -w imaiplay-admin
test "$(wc -l < admin/src/pages/course-detail/CourseDetailPage.tsx | tr -d ' ')" -le 200
git add admin/src/pages/CourseDetail.tsx admin/src/pages/course-detail admin/tests/courseDetailModel.test.ts
git commit -m "refactor(admin): decompose course detail workspace"
```

---

### Task 7: Split route registration by domain

**Files:**
- Modify: `internal/server/server.go`, `internal/server/server_test.go`
- Create: `internal/server/routes.go`, `routes_auth.go`, `routes_admin.go`, `routes_courses.go`, `routes_learner.go`, `routes_infrastructure.go`

**Interfaces:**
- Produces: `newRouteHandlers`, `registerAuthRoutes`, `registerAdminRoutes`, `registerCourseRoutes`, `registerLearnerRoutes`, and `registerInfrastructureRoutes`.

- [ ] **Step 1: Add and run the route contract**

```go
func TestRouteContractKeepsCompleteRouteCount(t *testing.T) {
	router := New(config.Config{}, func() error { return nil }, Dependencies{})
	if got, want := len(router.Routes()), 114; got != want {
		t.Fatalf("registered routes = %d, want %d", got, want)
	}
}
```

Run `go test ./internal/server -count=1`. Expected: PASS before moving routes.

- [ ] **Step 2: Create one handler registry**

Define an unexported `routeHandlers` struct. `newRouteHandlers(cfg, deps)` constructs every existing handler once and preserves `ResourceHandler.WithLearnerAccess(...).WithPlaybackSecret(...)`.

- [ ] **Step 3: Move routes unchanged**

Put portal/auth routes in `routes_auth.go`; plans/tenants/users/dashboard/config/audit in `routes_admin.go`; courses/chapters/lessons/enrollments/resources/categories/materials in `routes_courses.go`; published learner routes/progress/playback in `routes_learner.go`; Swagger and optional infrastructure in `routes_infrastructure.go`. Preserve middleware creation and order byte-for-byte where practical.

- [ ] **Step 4: Verify and commit**

```bash
gofmt -w internal/server/*.go
go test ./internal/server -count=1
go test ./internal/api/... -count=1
git add internal/server
git commit -m "refactor(server): split route registration by domain"
```

---

### Task 8: Split the application composition root

**Files:**
- Modify: `cmd/server/main.go`
- Create: `cmd/server/dependencies.go`, `cmd/server/infrastructure.go`

**Interfaces:**
- Produces: `newRepositories(database) appRepositories`, `newInfrastructure(cfg) (appInfrastructure, error)`, and `buildServerDependencies(cfg, database, repos, infrastructure) (server.Dependencies, error)`.

- [ ] **Step 1: Record the build baseline**

```bash
go test ./cmd/server ./internal/server -count=1
go build ./cmd/server
```

Expected: PASS.

- [ ] **Step 2: Extract construction responsibilities**

Move repository constructors into `newRepositories`; SMS, local/runtime storage, and optional Baota construction into `newInfrastructure`; and service construction/setters/wrappers into `buildServerDependencies`. Keep config loading, secret validation, DB lifecycle, migrations, reserved-domain repair, and `server.Run` in `run()`.

Preserve existing contextual error messages and all `WithLearnerAccess`, `WithCourseCategories`, and Auth setter calls.

- [ ] **Step 3: Verify and commit**

```bash
gofmt -w cmd/server/*.go
go test ./cmd/server ./internal/server -count=1
go build ./cmd/server
git add cmd/server
git commit -m "refactor(server): split dependency composition"
```

---

### Task 9: Split authentication responsibilities

**Files:**
- Modify: `internal/service/auth.go`
- Create: `internal/service/auth_credentials.go`, `auth_login.go`, `auth_recovery.go`, `auth_tokens.go`

**Interfaces:**
- Produces: unchanged `AuthService`, `TokenPair`, `AuthUser`, `OrganizationOption`, `LoginOutcome`, constructors, setters, and public methods.

- [ ] **Step 1: Run security characterization tests**

```bash
go test ./internal/service -run '^TestAuthService' -count=1
go test ./internal/server -run 'Test(AuthMe|SelectTenant|BackendRoutes)' -count=1
```

- [ ] **Step 2: Keep the compatibility shell**

Leave the service struct, constructors, setters, DTOs, `presentAuthUser`, `CurrentUser`, and shared normalization helpers in `auth.go`.

- [ ] **Step 3: Move methods by responsibility**

- `auth_credentials.go`: registration, phone registration, bootstrap, credential support.
- `auth_login.go`: password login, begin/scoped/platform login, tenant selection, completion.
- `auth_recovery.go`: forgot/reset, SMS login, verification-code and phone helpers.
- `auth_tokens.go`: issue, refresh, logout, rotation, tenant availability.

Do not rename methods or alter receivers. Preserve password-before-organization disclosure, unknown-phone non-disclosure, and refresh revocation order.

- [ ] **Step 4: Verify and commit**

```bash
gofmt -w internal/service/auth*.go
go test ./internal/service -run '^TestAuthService' -count=1
go test ./internal/api -run '^TestAuth' -count=1
go test ./internal/server -run 'Test(AuthMe|SelectTenant|BackendRoutes)' -count=1
wc -l internal/service/auth*.go
git add internal/service/auth*.go
git commit -m "refactor(auth): split authentication responsibilities"
```

Expected: tests pass, no public signature changes, and responsibility files are approximately 350 lines or less.

---

### Task 10: Split domain-binding responsibilities

**Files:**
- Modify: `internal/service/domain_bind.go`
- Create: `internal/service/domain_bind_verifier.go`, `domain_bind_orchestrator.go`, `domain_bind_status.go`

**Interfaces:**
- Produces: unchanged `DomainPanel`, `DomainResolver`, `DomainAuditRecorder`, `DomainBindConfig`, `DomainBindStatus`, service constructor, and public methods.

- [ ] **Step 1: Run domain characterization tests**

```bash
go test ./internal/service -run 'Test(VerifyDomain|DomainBind|DomainUnbind)' -count=1
```

- [ ] **Step 2: Keep public contracts in `domain_bind.go`**

Keep exported interfaces/types, service struct, resolver, constructor, and public Verify/Bind/Status/Unbind plus superadmin variants. Public methods delegate to moved private operations.

- [ ] **Step 3: Move responsibilities**

- `domain_bind_verifier.go`: DNS/IP checks, validation, actors, ownership, reservation/release.
- `domain_bind_orchestrator.go`: private bind/unbind, panel operations, rollback, SSL wait/readiness.
- `domain_bind_status.go`: cache cloning, operation tracking, status mutation, portal decoration, audit.

All shared status-map access remains protected by the existing mutex; do not change goroutine or wait timing.

- [ ] **Step 4: Verify and commit**

```bash
gofmt -w internal/service/domain_bind*.go
go test ./internal/service -run 'Test(VerifyDomain|DomainBind|DomainUnbind)' -count=1
go test ./internal/api -run '^TestDomainBind' -count=1
go test ./internal/server -run 'TestBackendRoutes' -count=1
wc -l internal/service/domain_bind*.go
git add internal/service/domain_bind*.go
git commit -m "refactor(domain): split binding responsibilities"
```

Expected: verification, concurrency, rollback, SSL, ownership, portal metadata, unbind, and audit tests pass; responsibility files are approximately 350 lines or less.

---

### Task 11: Consolidated verification and final review

**Files:**
- Modify: dependency metadata only when normalization finds justified changes
- Modify: implementation files only to fix Critical or Important final-review findings

**Interfaces:**
- Consumes: Tasks 1–10.
- Produces: a clean, fully tested optimization branch.

- [ ] **Step 1: Normalize dependencies**

```bash
go mod tidy
cd web
npm install --package-lock-only --ignore-scripts
npm dedupe
```

Retain only changes explained by removed imports or the workspace.

- [ ] **Step 2: Run the complete automated suite**

```bash
go test ./... -count=1
cd web
npm ci
npm run test:all
npm run build:all
```

Expected: all backend, shared, admin, PC, and H5 tests and builds pass.

- [ ] **Step 3: Verify structural contracts**

```bash
git check-ignore server
git diff --check
test "$(wc -l < web/pc/src/styles.css | tr -d ' ')" -le 10
test "$(wc -l < web/admin/src/pages/course-detail/CourseDetailPage.tsx | tr -d ' ')" -le 200
find internal/service -maxdepth 1 \( -name 'auth*.go' -o -name 'domain_bind*.go' \) -print0 | xargs -0 wc -l
```

- [ ] **Step 4: Perform final-only visual QA**

Compare login, learner home, course detail, player, admin dashboard, and admin course detail at retained desktop/mobile baseline viewports. Record regressions only; add no process screenshots to Git.

- [ ] **Step 5: Conduct one consolidated review**

Review the complete diff once for API drift, auth/session regressions, domain concurrency changes, missing middleware, CSS cascade changes, duplicate shared logic, and accidental artifact deletion. Fix all Critical and Important findings, then rerun Steps 2–3.

- [ ] **Step 6: Commit justified final fixes and verify handoff**

```bash
git add go.mod go.sum Dockerfile internal web
git diff --cached --quiet || git commit -m "chore: finalize codebase optimization"
git status -sb
git log --oneline --decorate --max-count=12
```

Expected: working tree clean and every phase represented by a focused commit.
