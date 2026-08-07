# Configurable Selection Colors Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let each tenant independently configure persistent selected-state background, text, and icon colors across the academy admin, PC learner portal, and H5 learner portal.

**Architecture:** Store three validated color fields on the tenant and expose them through the authenticated theme and public portal contracts. A shared theme helper supplies legacy-safe defaults and CSS-variable values; each client applies those variables only to persistent selected states while keeping ordinary branded actions on the existing primary color.

**Tech Stack:** Go, Gin, GORM, React 18, TypeScript, Ant Design, Ant Design Mobile, CSS custom properties, Node test runner, Vitest, Vite

## Global Constraints

- Add independent `selected_background_color`, `selected_text_color`, and `selected_icon_color` values.
- All explicit values must be six-digit hexadecimal colors.
- Empty or invalid stored values fall back to primary background, highest-contrast black/white text, and matching icon color.
- Low-contrast explicit choices display a warning but remain saveable.
- Ordinary buttons, progress bars, hover, focus, and temporary pressed states continue using the primary color.
- Legacy tenants and cached portal responses must remain readable during deployment.

---

### Task 1: Persist and Default Tenant Selection Colors

**Files:**
- Modify: `internal/domain/tenant.go`
- Modify: `internal/repository/tenant_gorm.go`
- Modify: `internal/service/theme.go`
- Modify: `internal/service/theme_test.go`
- Test: `internal/service/coverage_additional_test.go`

**Interfaces:**
- Consumes: existing `TenantThemeService`, `TenantRepository.UpdateTheme`, and `migration.AutoMigrate`.
- Produces: `service.ThemeUpdate`, tenant selection-color fields, and `themeWithDefaults(*domain.Tenant) *domain.Tenant` with complete values.

- [ ] **Step 1: Write failing default, persistence, and validation tests**

Add tests that create a tenant without selection colors and assert:

```go
if theme.SelectedBackgroundColor != DefaultPrimaryColor ||
    theme.SelectedTextColor != "#ffffff" ||
    theme.SelectedIconColor != "#ffffff" {
    t.Fatalf("selection colors = %#v", theme)
}
```

Add an update test using independent colors:

```go
updated, err := themeService.Update(ctx, ThemeUpdate{
    PrimaryColor: "#3582E1",
    SelectedBackgroundColor: "#FFF1F0",
    SelectedTextColor: "#C5221F",
    SelectedIconColor: "#8C1D18",
})
```

Assert the returned values and a fresh repository read preserve all three fields. Add table cases for malformed values such as `"blue"`, `"#123"`, and `"#12345678"`, asserting the bad-request message names the invalid JSON field.

- [ ] **Step 2: Run the focused tests and verify failure**

Run: `go test ./internal/service -run 'TestTenantTheme' -count=1`

Expected: FAIL because the fields and `ThemeUpdate` do not exist.

- [ ] **Step 3: Add model fields and repository persistence**

Add to `domain.Tenant`:

```go
SelectedBackgroundColor string `gorm:"size:16" json:"selected_background_color,omitempty"`
SelectedTextColor       string `gorm:"size:16" json:"selected_text_color,omitempty"`
SelectedIconColor       string `gorm:"size:16" json:"selected_icon_color,omitempty"`
```

Include the three column names in `tenantGORMRepository.UpdateTheme`. `migration.AutoMigrate` will add nullable columns without a data backfill.

- [ ] **Step 4: Introduce a structured update and default calculation**

Define:

```go
type ThemeUpdate struct {
    PrimaryColor            string
    LogoURL                 string
    WelcomeText             string
    BrowserTitle            string
    SelectedBackgroundColor string
    SelectedTextColor       string
    SelectedIconColor       string
}
```

Change `TenantThemeService.Update` to accept `ThemeUpdate`. Validate each non-empty color with `hexColorPattern`, normalize valid values to uppercase, and use a helper that chooses `#000000` or `#ffffff` by WCAG contrast against the selected background. `themeWithDefaults` must return all three selection colors even for legacy records.

- [ ] **Step 5: Run service tests**

Run: `go test ./internal/service -count=1`

Expected: PASS.

- [ ] **Step 6: Commit the persistence layer**

```bash
git add internal/domain/tenant.go internal/repository/tenant_gorm.go internal/service/theme.go internal/service/theme_test.go internal/service/coverage_additional_test.go
git commit -m "feat: persist tenant selection colors"
```

---

### Task 2: Expose Selection Colors Through Theme and Portal APIs

**Files:**
- Modify: `internal/api/theme.go`
- Modify: `internal/api/theme_test.go`
- Modify: `internal/service/portal.go`
- Modify: `internal/service/portal_test.go`
- Modify: `internal/api/portal_test.go`

**Interfaces:**
- Consumes: `service.ThemeUpdate` and defaulted fields on `domain.Tenant`.
- Produces: JSON properties `selected_background_color`, `selected_text_color`, and `selected_icon_color` from both `/backend/v1/theme` and `/api/v1/portal`.

- [ ] **Step 1: Write failing API and portal tests**

Extend the theme handler test request with:

```json
{
  "primary_color": "#3582E1",
  "selected_background_color": "#FFF1F0",
  "selected_text_color": "#C5221F",
  "selected_icon_color": "#8C1D18"
}
```

Assert the handler passes a `service.ThemeUpdate` containing the same values and returns them. Extend `TestPortalResolveByTenantCode` and `TestPortalHandlerReturnsPublicMetadata` to assert all three public fields.

- [ ] **Step 2: Run focused tests and verify failure**

Run: `go test ./internal/api ./internal/service -run 'Theme|Portal' -count=1`

Expected: FAIL because the handlers and portal DTO omit the fields.

- [ ] **Step 3: Update handler and portal contracts**

Extend `TenantThemeService` to:

```go
Update(context.Context, service.ThemeUpdate) (*domain.Tenant, error)
```

Bind the seven theme values into `service.ThemeUpdate`, and centralize the response map in a `themeResponse` helper so GET and PUT cannot drift. Add three fields to `service.Portal` and populate them from `themeWithDefaults(tenant)`.

- [ ] **Step 4: Run API and portal tests**

Run: `go test ./internal/api ./internal/service -run 'Theme|Portal' -count=1`

Expected: PASS.

- [ ] **Step 5: Run the complete Go suite and commit**

Run: `go test ./... && go build ./...`

Expected: both commands exit `0`.

```bash
git add internal/api/theme.go internal/api/theme_test.go internal/service/portal.go internal/service/portal_test.go internal/api/portal_test.go
git commit -m "feat: expose tenant selection colors"
```

---

### Task 3: Add Shared Selection Theme Utilities

**Files:**
- Modify: `web/shared/src/types/theme.ts`
- Modify: `web/shared/src/theme/tenantTheme.ts`
- Modify: `web/shared/src/index.ts`
- Modify: `web/shared/tests/tenantTheme.test.ts`

**Interfaces:**
- Consumes: optional selection colors from API payloads.
- Produces: `TenantSelectionColors`, `recommendedSelectionColors(primaryColor)`, `normalizeSelectionColors(theme, fallbackPrimary)`, `contrastRatio(foreground, background)`, and the extended `TenantThemeContract`.

- [ ] **Step 1: Write failing utility tests**

Add assertions for:

```ts
assert.deepEqual(recommendedSelectionColors('#3582E1'), {
  selected_background_color: '#3582E1',
  selected_text_color: '#FFFFFF',
  selected_icon_color: '#FFFFFF',
})

assert.deepEqual(normalizeSelectionColors({
  primary_color: '#3582E1',
  selected_background_color: '#fff1f0',
  selected_text_color: '#c5221f',
  selected_icon_color: '#8c1d18',
}, '#4F46E5'), {
  selected_background_color: '#FFF1F0',
  selected_text_color: '#C5221F',
  selected_icon_color: '#8C1D18',
})
```

Also assert invalid or missing values fall back independently and that `contrastRatio('#FFFFFF', '#FFFFFF') === 1`.

- [ ] **Step 2: Run shared tests and verify failure**

Run: `cd web/shared && npm test`

Expected: FAIL because the exports do not exist.

- [ ] **Step 3: Implement shared types and helpers**

Extend `TenantThemeContract` with optional selection fields for rollout compatibility. Normalize valid hex strings to uppercase. `recommendedSelectionColors` must choose black or white using the larger WCAG contrast ratio against the normalized background.

- [ ] **Step 4: Run shared tests and build**

Run: `cd web/shared && npm test && npm run build`

Expected: PASS.

- [ ] **Step 5: Commit shared contracts**

```bash
git add web/shared/src/types/theme.ts web/shared/src/theme/tenantTheme.ts web/shared/src/index.ts web/shared/tests/tenantTheme.test.ts
git commit -m "feat: share selection theme colors"
```

---

### Task 4: Add Admin Controls, Preview, and Selected-State Variables

**Files:**
- Modify: `web/admin/src/context/AdminThemeContext.tsx`
- Modify: `web/admin/src/components/AdminThemeProvider.tsx`
- Modify: `web/admin/src/theme/adminPalette.ts`
- Modify: `web/admin/src/pages/ThemeSettings.tsx`
- Modify: `web/admin/src/styles.css`
- Modify: `web/admin/tests/adminPalette.test.ts`
- Create: `web/admin/tests/themeSettingsSelectionColors.test.ts`

**Interfaces:**
- Consumes: shared `normalizeSelectionColors`, `recommendedSelectionColors`, and `contrastRatio`.
- Produces: admin context values `selectedBackgroundColor`, `selectedTextColor`, `selectedIconColor` and CSS variables `--tenant-selected-background`, `--tenant-selected-text`, `--tenant-selected-icon`.

- [ ] **Step 1: Write failing palette and settings-source tests**

Extend the palette test to apply explicit selection colors and assert:

```ts
assert.equal(properties.get('--tenant-selected-background'), '#FFF1F0')
assert.equal(properties.get('--tenant-selected-text'), '#C5221F')
assert.equal(properties.get('--tenant-selected-icon'), '#8C1D18')
assert.equal(properties.get('--tenant-primary'), '#3582E1')
```

Add a source-level settings test that requires all three Chinese labels, a selection preview class, contrast warning copy `当前配色对比度较低，可能看不清`, and independent request properties.

- [ ] **Step 2: Run admin tests and verify failure**

Run: `cd web/admin && npm test`

Expected: FAIL on missing controls and variables.

- [ ] **Step 3: Extend context and theme application**

Normalize the three values when loading the theme. Pass them into `createAdminThemeTokens` and `applyAdminPalette`. Keep `primary`, `info`, hover, focus, buttons, and progress derived from `primaryColor`; use selection colors only for Ant Menu selected tokens and selection CSS variables.

- [ ] **Step 4: Build the editor controls and live preview**

Keep four controlled color states: primary and three selection colors. Render labeled `ColorPicker` controls, plus a preview containing one normal row and one selected row whose background, text, and icon use the controlled values. Show an Ant Design `Alert` when text/background or icon/background contrast is below `3`.

When primary changes, compare each selection color to the previous recommended value; update only fields still matching recommendations. Reset derives all three recommendations from `#4F46E5`. Save all fields through `themeApi.update`.

- [ ] **Step 5: Replace conflicting selected-state CSS**

Consolidate repeated selected-menu rules so desktop and drawer use:

```css
color: var(--tenant-selected-text) !important;
background: var(--tenant-selected-background) !important;
```

and selected `.anticon` uses `var(--tenant-selected-icon) !important`. Remove obsolete `--tenant-selected` fallbacks that force the icon to text color.

- [ ] **Step 6: Run admin tests and build**

Run: `cd web/admin && npm test && npm run build`

Expected: PASS.

- [ ] **Step 7: Commit admin configuration**

```bash
git add web/admin/src/context/AdminThemeContext.tsx web/admin/src/components/AdminThemeProvider.tsx web/admin/src/theme/adminPalette.ts web/admin/src/pages/ThemeSettings.tsx web/admin/src/styles.css web/admin/tests/adminPalette.test.ts web/admin/tests/themeSettingsSelectionColors.test.ts
git commit -m "feat: configure admin selection colors"
```

---

### Task 5: Apply Selection Colors to the PC Learner Portal

**Files:**
- Modify: `web/pc/src/context/TenantThemeContext.tsx`
- Modify: `web/pc/src/theme/learnerPalette.ts`
- Modify: `web/pc/src/styles/course.css`
- Modify: `web/pc/src/styles/responsive.css`
- Create: `web/pc/tests/selectionTheme.test.ts`

**Interfaces:**
- Consumes: selection fields on `TenantPortalContract` and shared normalization helpers.
- Produces: the three selection CSS variables and PC current-navigation/tab styles.

- [ ] **Step 1: Write a failing PC theme test**

Assert the provider/palette helper maps independent portal values to all three variables and that the PC CSS references them for `.learner-top-nav-link.active`, `.learner-filter-tabs .ant-tabs-tab-active`, and `.course-experience-tabs .ant-tabs-tab-active`. Assert primary button/progress selectors continue using learner accent variables.

- [ ] **Step 2: Run PC tests and verify failure**

Run: `cd web/pc && npm test`

Expected: FAIL because selection variables are absent.

- [ ] **Step 3: Apply normalized variables and selected styles**

Add the normalized selection values to the context theme and set the three variables on `document.documentElement`. Style persistent active navigation and tabs with selected background/text; nested icons use selected icon. Do not change hover-only selectors, focus outlines, progress bars, or primary buttons.

- [ ] **Step 4: Run PC tests and build**

Run: `cd web/pc && npm test && npm run build`

Expected: PASS.

- [ ] **Step 5: Commit PC theme support**

```bash
git add web/pc/src/context/TenantThemeContext.tsx web/pc/src/theme/learnerPalette.ts web/pc/src/styles/course.css web/pc/src/styles/responsive.css web/pc/tests/selectionTheme.test.ts
git commit -m "feat: apply PC selection colors"
```

---

### Task 6: Apply Selection Colors to the H5 Learner Portal

**Files:**
- Modify: `web/h5/src/context/TenantThemeContext.tsx`
- Modify: `web/h5/src/styles.css`
- Create: `web/h5/tests/selectionTheme.test.ts`

**Interfaces:**
- Consumes: selection fields on `TenantPortalContract` and shared normalization helpers.
- Produces: the three selection CSS variables and H5 persistent TabBar/tab selected styles.

- [ ] **Step 1: Write a failing H5 theme test**

Assert the provider source sets all three variables from normalized portal values. Assert `.app-tabbar .adm-tab-bar-item-active` uses selected background and text while its icon uses selected icon. Assert `.adm-button-primary` still uses `--learner-accent`.

- [ ] **Step 2: Run H5 tests and verify failure**

Run: `cd web/h5 && npm test`

Expected: FAIL because selection variables are absent.

- [ ] **Step 3: Apply normalized variables and selected styles**

Add all three values to `TenantThemeContextValue`, normalize legacy payloads, and set the CSS variables when applying the learner palette. Add persistent active TabBar/tab selectors only; do not reuse the colors for `:active` press feedback.

- [ ] **Step 4: Run H5 tests and build**

Run: `cd web/h5 && npm test && npm run build`

Expected: PASS.

- [ ] **Step 5: Commit H5 theme support**

```bash
git add web/h5/src/context/TenantThemeContext.tsx web/h5/src/styles.css web/h5/tests/selectionTheme.test.ts
git commit -m "feat: apply H5 selection colors"
```

---

### Task 7: Full Regression and Browser Verification

**Files:**
- Verify only; modify earlier files only if a discovered defect requires a focused fix and regression test.

**Interfaces:**
- Consumes: completed backend, shared, admin, PC, and H5 implementations.
- Produces: evidence that the configuration persists and affects only persistent selected states.

- [ ] **Step 1: Run all backend checks**

Run: `go test ./... && go build ./...`

Expected: PASS with zero failures.

- [ ] **Step 2: Run all frontend checks**

Run:

```bash
cd web/shared && npm test && npm run build
cd ../admin && npm test && npm run build
cd ../pc && npm test && npm run build
cd ../h5 && npm test && npm run build
```

Expected: every command exits `0`.

- [ ] **Step 3: Verify the rendered workflow**

In a local tenant session:

1. Open academy theme settings.
2. Set background `#FFF1F0`, text `#C5221F`, and icon `#8C1D18`.
3. Confirm the preview shows three independent colors and save.
4. Reload admin, PC, and H5 pages and inspect persistent selected states.
5. Confirm primary buttons and progress bars retain the primary color.
6. Select a low-contrast combination and confirm the warning appears but save remains enabled.
7. Reset and confirm recommended values return.

- [ ] **Step 4: Confirm a clean diff and final commit if needed**

Run: `git diff --check && git status --short`

Expected: no whitespace errors and only intentional changes. If browser verification required a correction, commit only that correction and its regression test with `git commit -m "fix: finalize selection theme colors"`.
