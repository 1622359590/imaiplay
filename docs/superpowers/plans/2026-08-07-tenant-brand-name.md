# Tenant Brand Name Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let each tenant administrator persist and immediately display a custom admin-sidebar brand name without renaming the tenant.

**Architecture:** Extend the existing tenant theme record and API with an optional `brand_name` field, then consume that field through the existing admin theme context. Keep fallback selection in a small pure frontend utility so it can be tested without rendering React.

**Tech Stack:** Go, Gin, GORM, SQLite/PostgreSQL migrations, React 18, TypeScript, Ant Design, Node test runner.

## Global Constraints

- `brand_name` is tenant-scoped and independent of the registered tenant name.
- Trim surrounding whitespace and reject values longer than 50 characters.
- Empty `brand_name` falls back to the authenticated tenant name, then `ImaiPlay`.
- Saving must update the sidebar through the existing `tenant-theme-changed` event without a new login.
- Browser title remains controlled only by `browser_title`.

---

### Task 1: Persist the Tenant Brand Name

**Files:**
- Modify: `internal/domain/tenant.go`
- Modify: `internal/migration/migration.go`
- Modify: `internal/migration/migration_test.go`
- Modify: `internal/repository/tenant_gorm.go`
- Modify: `internal/repository/tenant_gorm_test.go`
- Modify: `internal/service/theme.go`
- Modify: `internal/service/theme_test.go`
- Modify: `internal/service/coverage_additional_test.go`

**Interfaces:**
- Produces: `domain.Tenant.BrandName string` with JSON name `brand_name`.
- Produces: `TenantThemeService.Update(ctx, primaryColor, logoURL, welcomeText, browserTitle, brandName string)`.
- Migration version 18 adds `tenants.brand_name` through GORM auto-migration.

- [ ] **Step 1: Write failing migration and service tests**

Add a migration assertion and update the expected migration count:

```go
if !database.Migrator().HasColumn(&domain.Tenant{}, "BrandName") {
    t.Fatal("AutoMigrate() did not create tenants.brand_name")
}
// The schema migration count must be 18 before and after the idempotency run.
```

Add service coverage using two tenants and an authenticated tenant-admin context:

```go
updated, err := service.Update(ctx, "#111111", "", "", "", "  Acme Academy  ")
if err != nil { t.Fatal(err) }
if updated.BrandName != "Acme Academy" { t.Fatalf("brand = %q", updated.BrandName) }

stored, err := repo.FindByID(context.Background(), first.ID)
if err != nil || stored.BrandName != "Acme Academy" {
    t.Fatalf("stored brand = %q, err=%v", stored.BrandName, err)
}
other, _ := repo.FindByID(context.Background(), second.ID)
if other.BrandName != "" { t.Fatalf("other tenant brand = %q", other.BrandName) }

_, err = service.Update(ctx, "#111111", "", "", "", strings.Repeat("界", 51))
if err == nil { t.Fatal("expected brand length validation error") }
```

- [ ] **Step 2: Run the focused tests and observe failure**

Run:

```bash
go test ./internal/migration ./internal/repository ./internal/service
```

Expected: compile failures for missing `BrandName` and the six-argument `Update`, or assertions that migration/persistence is missing.

- [ ] **Step 3: Implement the model, migration, repository, and service changes**

Add the model field:

```go
BrandName string `gorm:"size:50" json:"brand_name,omitempty"`
```

Register migration 18 and implement it:

```go
{Version: 18, Up: migrateV18},

func migrateV18(database *gorm.DB) error {
    return database.AutoMigrate(&domain.Tenant{})
}
```

Persist `brand_name` in `UpdateTheme`. Extend the service signature, trim all string values before validation, validate `len([]rune(brandName)) <= 50`, assign `tenant.BrandName`, and reuse the existing repository update.

- [ ] **Step 4: Run focused tests until green**

Run:

```bash
go test ./internal/migration ./internal/repository ./internal/service
```

Expected: all three packages pass.

- [ ] **Step 5: Commit backend persistence**

```bash
git add internal/domain/tenant.go internal/migration/migration.go internal/migration/migration_test.go internal/repository/tenant_gorm.go internal/repository/tenant_gorm_test.go internal/service/theme.go internal/service/theme_test.go internal/service/coverage_additional_test.go
git commit -m "feat(theme): persist tenant brand name"
```

---

### Task 2: Expose Brand Name Through the Theme API Contract

**Files:**
- Modify: `internal/api/theme.go`
- Create: `internal/api/theme_test.go`
- Modify: `web/shared/src/types/theme.ts`

**Interfaces:**
- Consumes: `TenantThemeService.Update(ctx, primaryColor, logoURL, welcomeText, browserTitle, brandName string)` from Task 1.
- Produces: theme GET/PUT JSON property `brand_name`.
- Produces: `TenantThemeContract.brand_name?: string` for all frontend consumers.

- [ ] **Step 1: Write failing handler contract tests**

Create a handler stub that records the final argument and returns a tenant:

```go
type themeServiceStub struct { brandName string }

func (stub *themeServiceStub) Get(context.Context) (*domain.Tenant, error) {
    return &domain.Tenant{PrimaryColor: "#4F46E5", BrandName: "Acme Academy"}, nil
}

func (stub *themeServiceStub) Update(_ context.Context, primary, logo, welcome, browser, brand string) (*domain.Tenant, error) {
    stub.brandName = brand
    return &domain.Tenant{PrimaryColor: primary, BrandName: brand}, nil
}
```

Assert GET returns `"brand_name":"Acme Academy"`, and an authenticated PUT body containing `"brand_name":"Sales School"` reaches the stub and returns the same property.

- [ ] **Step 2: Run the API test and observe failure**

Run:

```bash
go test ./internal/api -run Theme -count=1
```

Expected: the stub does not match the current service interface and/or `brand_name` is absent.

- [ ] **Step 3: Implement the handler and shared contract**

Extend the handler request and both response maps:

```go
BrandName string `json:"brand_name"`
```

Pass `request.BrandName` as the final service argument. Extend the shared contract:

```ts
export interface TenantThemeContract {
  primary_color: string
  brand_name?: string
  logo_url?: string
  welcome_text?: string
  browser_title?: string
}
```

- [ ] **Step 4: Run handler and shared tests**

Run:

```bash
go test ./internal/api -run Theme -count=1
npm test --prefix web/shared
```

Expected: both commands pass.

- [ ] **Step 5: Commit the API contract**

```bash
git add internal/api/theme.go internal/api/theme_test.go web/shared/src/types/theme.ts
git commit -m "feat(theme): expose tenant brand name"
```

---

### Task 3: Edit and Apply the Brand Name in Tenant Admin

**Files:**
- Create: `web/admin/src/utils/adminBrandName.ts`
- Create: `web/admin/tests/adminBrandName.test.ts`
- Modify: `web/admin/src/context/AdminThemeContext.tsx`
- Modify: `web/admin/src/pages/ThemeSettings.tsx`

**Interfaces:**
- Consumes: `TenantThemeContract.brand_name?: string` from Task 2.
- Produces: `resolveAdminBrandName(brandName: unknown, tenantName: unknown): string`.
- Existing `AdminThemeValue.brandName` continues to drive desktop and drawer sidebar labels and logo alt text.

- [ ] **Step 1: Write a failing fallback test**

```ts
import assert from 'node:assert/strict'
import test from 'node:test'
import { resolveAdminBrandName } from '../src/utils/adminBrandName.ts'

test('tenant brand name falls back without coupling to browser title', () => {
  assert.equal(resolveAdminBrandName(' Acme Academy ', 'Acme Ltd'), 'Acme Academy')
  assert.equal(resolveAdminBrandName('   ', ' Acme Ltd '), 'Acme Ltd')
  assert.equal(resolveAdminBrandName(undefined, undefined), 'ImaiPlay')
})
```

- [ ] **Step 2: Run the test and observe failure**

Run:

```bash
npm test --prefix web/admin -- --test-name-pattern="tenant brand name"
```

Expected: module-not-found failure for `adminBrandName.ts`.

- [ ] **Step 3: Implement the resolver, theme context, and form field**

Create the resolver:

```ts
export function resolveAdminBrandName(brandName: unknown, tenantName: unknown): string {
  const configured = typeof brandName === 'string' ? brandName.trim() : ''
  if (configured) return configured
  const tenant = typeof tenantName === 'string' ? tenantName.trim() : ''
  return tenant || 'ImaiPlay'
}
```

In `AdminThemeContext`, resolve the sidebar name with:

```ts
brandName: resolveAdminBrandName(
  data.brand_name,
  localStorage.getItem(ADMIN_TENANT_NAME_KEY),
),
```

Do not change the browser-title expression. Add this form item before the logo field:

```tsx
<Form.Item
  label="品牌名称"
  name="brand_name"
  rules={[{ max: 50, message: '品牌名称不能超过 50 个字符' }]}
>
  <Input maxLength={50} showCount placeholder="例如：小明科技学习中心" />
</Form.Item>
```

Update the page description to mention the brand name. The existing save payload and `tenant-theme-changed` event already carry and reload the new contract field.

- [ ] **Step 4: Run admin tests and production build**

Run:

```bash
npm test --prefix web/admin
npm run build --prefix web/admin
```

Expected: all admin tests pass and Vite produces the production bundle.

- [ ] **Step 5: Commit the admin UI**

```bash
git add web/admin/src/utils/adminBrandName.ts web/admin/tests/adminBrandName.test.ts web/admin/src/context/AdminThemeContext.tsx web/admin/src/pages/ThemeSettings.tsx
git commit -m "feat(admin): customize tenant brand name"
```

---

### Task 4: Full Regression Verification

**Files:**
- Verify only; modify a file only if a regression is found.

**Interfaces:**
- Consumes all deliverables from Tasks 1–3.
- Produces a release-ready branch with no uncommitted changes.

- [ ] **Step 1: Run all backend tests**

```bash
go test ./...
```

Expected: all Go packages pass.

- [ ] **Step 2: Run all web workspace tests and builds**

```bash
npm run test:all --prefix web
npm run build:all --prefix web
```

Expected: every configured workspace test and build passes.

- [ ] **Step 3: Inspect scope and whitespace**

```bash
git diff origin/main...HEAD --check
git status --short
git log --oneline origin/main..HEAD
```

Expected: no whitespace errors, a clean worktree, and only the design plus tenant-brand-name commits.
