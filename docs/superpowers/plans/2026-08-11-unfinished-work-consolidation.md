# Unfinished Work Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the remaining public-registration employee-limit gap, verify the complete product, archive obsolete worktrees safely, fast-forward `main`, and publish the same commit to Gitee before GitHub.

**Architecture:** `UserService` remains the single owner of plan-capacity rules and implements a small `EmployeeCapacityChecker` interface. `AuthService` receives that checker through an optional setter and invokes it after resolving the tenant but before checking credentials or creating a user; production wiring injects the same `UserService` instance into both HTTP service dependencies.

**Tech Stack:** Go 1.24, GORM, SQLite service tests, React 18, TypeScript 5.7, Vite 6, npm workspaces, Git worktrees.

## Global Constraints

- Preserve `/Users/imaiwork/Documents/imaiplay-go/.codex/current-task.md` and `/Users/imaiwork/Documents/imaiplay-go/.codex/roadmap.md` exactly as user-owned changes.
- `MaxUsers <= 0` and tenants without a plan remain unlimited.
- `BootstrapSuperadmin` and initial tenant registration do not use the employee-capacity checker.
- Do not restore obsolete portal variants already superseded by `main`.
- Stash tracked and untracked content before removing any dirty worktree.
- Push Gitee first and GitHub second; stop and report if either push fails.

---

### Task 1: Enforce plan capacity in public tenant registration

**Files:**
- Modify: `internal/service/auth.go`
- Modify: `internal/service/auth_credentials.go`
- Modify: `internal/service/user.go`
- Modify: `internal/service/auth_test.go`
- Modify: `internal/service/user_import_test.go`
- Modify: `cmd/server/dependencies.go`

**Interfaces:**
- Produces: `type EmployeeCapacityChecker interface { EnsureEmployeeCapacity(context.Context, string) error }`
- Produces: `func (service *AuthService) SetEmployeeCapacityChecker(checker EmployeeCapacityChecker)`
- Produces: `func (service *UserService) EnsureEmployeeCapacity(ctx context.Context, tenantID string) error`
- Consumes: tenant, plan, and user repositories already held by `UserService`.

- [ ] **Step 1: Write the failing public-registration capacity test**

Add this regression test to `internal/service/auth_test.go`:

```go
func TestAuthServiceRegisterRejectsWhenTenantEmployeeLimitReached(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	planRepo := repository.NewPlanRepository(database)
	ctx := context.Background()
	plan := &domain.Plan{Name: "Single employee", MaxUsers: 1, Status: 1}
	if err := planRepo.Create(ctx, plan); err != nil {
		t.Fatal(err)
	}
	tenant := &domain.Tenant{Code: "registration-limit", Name: "Registration Limit", Status: 1, PlanID: &plan.ID}
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatal(err)
	}
	if err := userRepo.Create(ctx, &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Email: "existing@example.com", Password: "hash", Name: "Existing", Role: "learner", Status: 1,
	}); err != nil {
		t.Fatal(err)
	}
	auth := NewAuthService(userRepo, tenantRepo, "secret")
	auth.SetEmployeeCapacityChecker(NewUserService(userRepo, UserLimitRepositories{
		Tenants: tenantRepo,
		Plans: planRepo,
	}))
	tenantCtx := tenantcontext.WithTenant(ctx, tenant.Code, tenantcontext.SourceHeaderCode)

	_, err := auth.Register(tenantCtx, "new@example.com", "password123", "New", "learner")
	if errorCode(err) != 40300 || err.Error() != "员工数已达套餐上限，请升级套餐" {
		t.Fatalf("Register() error = %#v", err)
	}
}
```

- [ ] **Step 2: Run the new test and verify the red state**

Run:

```bash
go test -count=1 ./internal/service -run '^TestAuthServiceRegisterRejectsWhenTenantEmployeeLimitReached$'
```

Expected: compilation fails because `SetEmployeeCapacityChecker` is undefined, demonstrating that public registration has no capacity dependency.

- [ ] **Step 3: Expose the shared capacity checker from `UserService`**

In `internal/service/user.go`, add:

```go
type EmployeeCapacityChecker interface {
	EnsureEmployeeCapacity(ctx context.Context, tenantID string) error
}
```

Rename the existing private `ensureEmployeeCapacity` method to `EnsureEmployeeCapacity` and update `CreateWithPhone` to call the exported method without changing capacity behavior.

- [ ] **Step 4: Add optional capacity injection to `AuthService`**

In `internal/service/auth.go`, add the field and nil-safe setter:

```go
employeeCapacity EmployeeCapacityChecker

func (service *AuthService) SetEmployeeCapacityChecker(checker EmployeeCapacityChecker) {
	service.employeeCapacity = checker
}
```

In `internal/service/auth_credentials.go`, immediately after `currentTenant` succeeds, add:

```go
if service.employeeCapacity != nil {
	if err := service.employeeCapacity.EnsureEmployeeCapacity(ctx, tenant.ID); err != nil {
		return nil, err
	}
}
```

- [ ] **Step 5: Wire one `UserService` instance into authentication and dependencies**

In `cmd/server/dependencies.go`, construct once and reuse:

```go
userService := service.NewUserService(repos.user, service.UserLimitRepositories{
	Tenants: repos.tenant,
	Plans: repos.plan,
})
auth.SetEmployeeCapacityChecker(userService)
```

Set `UserService: userService` in `server.Dependencies`.

- [ ] **Step 6: Format and run focused tests**

Add this regression test to `internal/service/user_import_test.go` to prove each imported row uses the same capacity rule:

```go
func TestUserImportStopsCreatingUsersAtTenantEmployeeLimit(t *testing.T) {
	database, tenantRepo, userRepo := serviceRepositories(t)
	planRepo := repository.NewPlanRepository(database)
	ctx := context.Background()
	plan := &domain.Plan{Name: "Single employee", MaxUsers: 1, Status: 1}
	if err := planRepo.Create(ctx, plan); err != nil {
		t.Fatal(err)
	}
	tenant := &domain.Tenant{Code: "import-limit", Name: "Import Limit", Status: 1, PlanID: &plan.ID}
	if err := tenantRepo.Create(ctx, tenant); err != nil {
		t.Fatal(err)
	}
	users := NewUserService(userRepo, UserLimitRepositories{Tenants: tenantRepo, Plans: planRepo})
	admin := usercontext.WithUser(ctx, "admin", tenant.ID, "admin@example.com", "tenant_admin")
	result, err := users.Import(admin, []UserImportRow{
		{Row: 2, Name: "One", Email: "one@example.com", Password: "password1"},
		{Row: 3, Name: "Two", Email: "two@example.com", Password: "password2"},
	})
	if err != nil || result.Succeeded != 1 || result.Failed != 1 ||
		len(result.Errors) != 1 || result.Errors[0].Reason != "员工数已达套餐上限，请升级套餐" {
		t.Fatalf("Import() = %#v, %v", result, err)
	}
}
```

Add `domain` and `repository` imports to that test file, then run:

```bash
gofmt -w internal/service/auth.go internal/service/auth_credentials.go internal/service/user.go internal/service/auth_test.go internal/service/user_import_test.go cmd/server/dependencies.go
go test -count=1 ./internal/service -run 'Test(AuthServiceRegisterRejectsWhenTenantEmployeeLimitReached|UserServiceCreateRejectsWhenTenantEmployeeLimitReached|UserServiceCreateAllowsUnlimitedEmployeesWhenMaxUsersIsZero|UserImportStopsCreatingUsersAtTenantEmployeeLimit)$'
```

Expected: all four tests pass.

- [ ] **Step 7: Fold code into the existing employee-limit feature commit**

```bash
git add internal/service/auth.go internal/service/auth_credentials.go internal/service/user.go internal/service/auth_test.go internal/service/user_import_test.go cmd/server/dependencies.go
git commit --fixup=f1f8b35
GIT_SEQUENCE_EDITOR=: git rebase -i --autosquash main
```

Expected: employee-limit code remains one functional commit, followed by documentation commits.

### Task 2: Verify the backend and all frontend workspaces

**Files:**
- Verify: `internal/service/...`
- Verify: `cmd/server/...`
- Verify: `web/shared/...`
- Verify: `web/admin/...`
- Verify: `web/pc/...`
- Verify: `web/h5/...`

**Interfaces:**
- Consumes: the complete branch after Task 1.
- Produces: reproducible build and test evidence.

- [ ] **Step 1: Check patch hygiene and status**

```bash
git diff --check
git status --short
```

- [ ] **Step 2: Build and test Go**

```bash
go build ./cmd/server/
go test -count=1 ./...
```

Expected: both commands exit 0.

- [ ] **Step 3: Test and build all web workspaces**

```bash
npm run test:all --prefix web
npm run build:all --prefix web
```

Expected: shared, Admin, PC, and H5 tests pass; all three production builds complete.

- [ ] **Step 4: Commit this plan**

```bash
git add docs/superpowers/plans/2026-08-11-unfinished-work-consolidation.md
git commit -m "docs: plan unfinished work consolidation"
```

### Task 3: Archive obsolete worktrees and repair the duplicate Gitee ref

**Files:**
- Archive: `/Users/imaiwork/Documents/playedu/imaiplay-release`
- Archive: `/Users/imaiwork/Documents/imaiplay-go/.worktrees/learner-station-workspaces`
- Remove: `/Users/imaiwork/Documents/imaiplay-go/.worktrees/simplify-learner-experience`
- Move: `/Users/imaiwork/Documents/imaiplay-go/.git/refs/remotes/gitee/main 2`

**Interfaces:**
- Consumes: byte-for-byte audit recorded in the design specification.
- Produces: named stash recovery points and valid local remote refs.

- [ ] **Step 1: Reconfirm obsolete branches are not ahead of `main`**

```bash
git rev-list --count main..codex/tenant-portal-login
git rev-list --count main..codex/tenant-brand-name
git rev-list --count main..codex/simplify-learner-experience
```

Expected: each prints `0`.

- [ ] **Step 2: Stash dirty obsolete snapshots**

```bash
git -C /Users/imaiwork/Documents/playedu/imaiplay-release stash push -u -m "archive obsolete tenant portal worktree before consolidation 2026-08-11"
git -C /Users/imaiwork/Documents/imaiplay-go/.worktrees/learner-station-workspaces stash push -u -m "archive duplicate admin files before consolidation 2026-08-11"
```

Then confirm both worktrees are clean with `git status --short`.

- [ ] **Step 3: Remove obsolete worktrees and local branches**

Run from `/Users/imaiwork/Documents/imaiplay-go`:

```bash
git worktree remove /Users/imaiwork/Documents/playedu/imaiplay-release
git worktree remove /Users/imaiwork/Documents/imaiplay-go/.worktrees/learner-station-workspaces
git worktree remove /Users/imaiwork/Documents/imaiplay-go/.worktrees/simplify-learner-experience
git branch -d codex/tenant-portal-login codex/tenant-brand-name codex/simplify-learner-experience
```

- [ ] **Step 4: Move the stale duplicate Gitee ref to Trash**

```bash
DUPLICATE_REF=$(sed -n '1p' '.git/refs/remotes/gitee/main 2')
git merge-base --is-ancestor "$DUPLICATE_REF" main
mv '.git/refs/remotes/gitee/main 2' '/Users/imaiwork/.Trash/imaiplay-gitee-main-2-ref-20260811'
git for-each-ref --format='%(refname) %(objectname)' refs/remotes/gitee
```

Expected: the duplicate points to a commit already contained in `main`, remains recoverable from Trash, and Git enumerates Gitee refs without errors.

### Task 4: Fast-forward `main` and publish Gitee before GitHub

**Files:**
- Preserve: `/Users/imaiwork/Documents/imaiplay-go/.codex/current-task.md`
- Preserve: `/Users/imaiwork/Documents/imaiplay-go/.codex/roadmap.md`

**Interfaces:**
- Consumes: verified task branch.
- Produces: identical `main` heads locally, on Gitee, and on GitHub.

- [ ] **Step 1: Confirm root contains only user-owned changes**

```bash
git -C /Users/imaiwork/Documents/imaiplay-go status --short
```

Expected: only the two `.codex` paths are listed.

- [ ] **Step 2: Fast-forward and push**

```bash
git -C /Users/imaiwork/Documents/imaiplay-go merge --ff-only codex/remove-official-course-employee-limit
git -C /Users/imaiwork/Documents/imaiplay-go push gitee main
git -C /Users/imaiwork/Documents/imaiplay-go push origin main
```

- [ ] **Step 3: Verify remote heads**

```bash
LOCAL=$(git -C /Users/imaiwork/Documents/imaiplay-go rev-parse main)
GITEE=$(git -C /Users/imaiwork/Documents/imaiplay-go ls-remote gitee refs/heads/main | awk '{print $1}')
GITHUB=$(git -C /Users/imaiwork/Documents/imaiplay-go ls-remote origin refs/heads/main | awk '{print $1}')
test "$LOCAL" = "$GITEE" && test "$LOCAL" = "$GITHUB"
```

- [ ] **Step 4: Remove the completed task worktree and branch**

```bash
git -C /Users/imaiwork/Documents/imaiplay-go worktree remove /Users/imaiwork/Documents/imaiplay-go/.worktrees/remove-official-course-employee-limit
git -C /Users/imaiwork/Documents/imaiplay-go branch -d codex/remove-official-course-employee-limit
```

Expected: all hashes match and only the root worktree remains for this task.
