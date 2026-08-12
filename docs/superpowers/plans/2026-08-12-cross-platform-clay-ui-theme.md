# ImaiPlay Cross-Platform Clay UI Theme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign Admin and H5, enhance the completed PC learner portal with tiered Claymorphism, and prove that tenant theme changes drive all three frontends without changing business behavior.

**Architecture:** Keep the existing Admin, PC, and H5 component trees and API flows. Add one pure shared Clay color derivation contract, expose its values through each frontend palette/provider, and implement device-appropriate CSS: full tactile Clay surfaces for learners and restrained depth for Admin dashboards/actions while tables and forms remain flat.

**Tech Stack:** React 18, TypeScript 5.7, Ant Design 5, Ant Design Mobile 5, Vite 6, Node test runner, Vitest, Go/Gin/GORM.

## Global Constraints

- Do not add npm dependencies or UI frameworks.
- Do not change authentication, authorization, routing, course APIs, resource loading, or learning heartbeat behavior.
- Color literals are allowed only in palette/default-theme definitions and palette tests; pages, components, and CSS consume variables or framework tokens.
- `primary_color` drives accent, hover, light, soft, strong, Clay surface, Clay contact shadow, and Clay atmosphere values.
- `selected_background_color`, `selected_text_color`, and `selected_icon_color` remain independent persisted values.
- Learner buttons use 6px resting depth, 2px hover compression, 6px active press, and `100ms ease-out` transitions.
- Admin tables, long forms, editors, and modal bodies must not receive 6–8px heavy Clay contact shadows.
- Preserve the user's uncommitted `.codex/` files and exclude them from every commit.

---

### Task 1: Shared Clay color derivation contract

**Files:**
- Modify: `web/shared/src/theme/tenantTheme.ts`
- Modify: `web/shared/src/index.ts`
- Modify: `web/shared/tests/tenantTheme.test.ts`

**Interfaces:**
- Consumes: normalized six-digit tenant `primary_color`.
- Produces: `deriveClayColors(primaryColor: string): { surface: string; shadow: string; atmosphere: string; highlight: string }` with uppercase hex surface/shadow and CSS color strings for atmosphere/highlight.

- [ ] **Step 1: Add failing shared derivation tests**

Add assertions showing the same input is deterministic, invalid input falls back to `#4F46E5`, the shadow differs from and is darker than the surface, and output keys are complete:

```ts
test('derives deterministic clay colors from tenant primary', () => {
  const first = deriveClayColors('#6366F1')
  const second = deriveClayColors('#6366F1')
  assert.deepEqual(first, second)
  assert.match(first.surface, /^#[0-9A-F]{6}$/)
  assert.match(first.shadow, /^#[0-9A-F]{6}$/)
  assert.notEqual(first.shadow, first.surface)
  assert.match(first.atmosphere, /^rgba\(/)
  assert.match(first.highlight, /^rgba\(/)
})

test('clay derivation normalizes invalid primary to the shared fallback', () => {
  assert.deepEqual(deriveClayColors('bad'), deriveClayColors('#4F46E5'))
})
```

- [ ] **Step 2: Run the shared tests and verify RED**

Run: `cd web/shared && node --test tests/tenantTheme.test.ts`

Expected: FAIL because `deriveClayColors` is not exported.

- [ ] **Step 3: Implement pure RGB/HSL derivation**

Implement `deriveClayColors()` next to the existing normalization helpers. Normalize the input, convert to HSL, derive the surface with bounded lightness, derive the contact shadow by decreasing lightness and increasing saturation, and return stable atmosphere/highlight opacity strings. Export it from `web/shared/src/index.ts`.

- [ ] **Step 4: Run shared tests and TypeScript verification**

Run: `cd web/shared && node --test tests/*.test.ts && npx tsc --noEmit`

Expected: all shared tests pass and TypeScript exits 0.

- [ ] **Step 5: Commit the shared contract**

```bash
git add web/shared/src/theme/tenantTheme.ts web/shared/src/index.ts web/shared/tests/tenantTheme.test.ts
git commit -m "feat(theme): derive shared clay surface colors"
```

### Task 2: PC palette injection and Clay interaction contracts

**Files:**
- Modify: `web/pc/src/theme/learnerPalette.ts`
- Modify: `web/pc/tests/learnerPalette.test.ts`
- Modify: `web/pc/tests/learnerStyles.test.ts`

**Interfaces:**
- Consumes: `deriveClayColors()` from Task 1.
- Produces: `LearnerPalette.claySurface`, `clayShadow`, `clayAtmosphere`, `clayHighlight`, `clayWhiteShadow`, plus CSS properties `--learner-clay-*`.

- [ ] **Step 1: Add failing PC palette injection assertions**

Extend the tenant-primary test:

```ts
assert.equal(properties.get('--learner-clay-surface'), palette.claySurface)
assert.equal(properties.get('--learner-clay-shadow'), palette.clayShadow)
assert.equal(properties.get('--learner-clay-atmosphere'), palette.clayAtmosphere)
assert.equal(properties.get('--learner-clay-highlight'), palette.clayHighlight)
assert.equal(properties.get('--learner-clay-white-shadow'), palette.clayWhiteShadow)
```

Add style assertions for primary button resting/hover/active depth and `100ms ease-out`, and require colored Clay cards to reference `var(--learner-clay-shadow)`.

- [ ] **Step 2: Run PC contract tests and verify RED**

Run: `cd web/pc && node --test tests/learnerPalette.test.ts tests/learnerStyles.test.ts`

Expected: FAIL because Clay palette properties and style contracts are absent.

- [ ] **Step 3: Extend `LearnerPalette` and `applyLearnerPalette()`**

Call `deriveClayColors(accent)` in `createLearnerPalette()`, map all five values, and add them to `CSS_PROPERTIES`. Retain existing readable foreground and Ant Design token behavior.

- [ ] **Step 4: Run PC palette tests**

Run: `cd web/pc && node --test tests/learnerPalette.test.ts`

Expected: palette tests pass; style tests still fail until Task 3.

- [ ] **Step 5: Commit palette injection**

```bash
git add web/pc/src/theme/learnerPalette.ts web/pc/tests/learnerPalette.test.ts web/pc/tests/learnerStyles.test.ts
git commit -m "feat(learner): inject PC clay theme variables"
```

### Task 3: PC learner Clay enhancement

**Files:**
- Modify: `web/pc/src/styles/base.css`
- Modify: `web/pc/src/styles/layout.css`
- Modify: `web/pc/src/styles/dashboard.css`
- Modify: `web/pc/src/styles/course.css`
- Modify: `web/pc/src/styles/player.css`
- Modify: `web/pc/src/styles/login.css`
- Modify: `web/pc/src/styles/responsive.css`

**Interfaces:**
- Consumes: `--learner-clay-*` variables from Task 2 and existing page class names from commit `4636180`.
- Produces: full Clay learner surfaces without JSX, routing, or API changes.

- [ ] **Step 1: Implement the base Clay button system**

Define primary and secondary Ant Design buttons with top highlight, zero-blur contact shadow, atmosphere shadow, and the exact resting/hover/active transforms. Keep text/link table actions visually light through scoped exclusions.

- [ ] **Step 2: Enhance dashboard and course surfaces**

Apply 6px white contact depth to stat/course cards, 8px brand-family depth to the continue-learning Hero, molded icon containers, pill progress tracks, and snappy press feedback on clickable course cards.

- [ ] **Step 3: Enhance detail, player, and login surfaces**

Give course chapters and login forms white Clay depth, keep the video stage dark and focused, and apply Clay only to player controls/sidebar navigation rather than the video element itself.

- [ ] **Step 4: Add responsive and reduced-motion behavior**

Reduce contact depth and atmosphere blur on narrow screens. Under `prefers-reduced-motion`, remove translate transforms and retain immediate visual state changes.

- [ ] **Step 5: Run PC tests and build**

Run: `cd web/pc && npm test && npm run build`

Expected: Node 58+ tests, Vitest 60+ tests, and Vite build all pass; the Clay style contract from Task 2 is green.

- [ ] **Step 6: Commit PC Clay visuals**

```bash
git add web/pc/src/styles web/pc/tests/learnerStyles.test.ts
git commit -m "feat(learner): add tactile clay surfaces to PC portal"
```

### Task 4: H5 palette injection and style contracts

**Files:**
- Modify: `web/h5/src/theme/learnerPalette.ts`
- Modify: `web/h5/src/context/TenantThemeContext.tsx`
- Modify: `web/h5/tests/learnerPalette.test.ts`
- Modify: `web/h5/tests/selectionTheme.test.ts`
- Create: `web/h5/tests/learnerStyles.test.ts`

**Interfaces:**
- Consumes: `deriveClayColors()`, `TenantSelectionColors`, portal theme fields.
- Produces: the same semantic `--learner-*` and `--learner-clay-*` CSS variable family as PC, plus Ant Design Mobile `--adm-color-primary`.

- [ ] **Step 1: Add failing H5 palette and stylesheet tests**

Test tenant primary propagation to accent/light/soft/strong and Clay variables, selection color persistence, absence of color literals in `src/styles.css`, 44px touch targets, and presence of reduced-motion overrides.

- [ ] **Step 2: Run H5 tests and verify RED**

Run: `cd web/h5 && node --test tests/*.test.ts`

Expected: FAIL on missing Clay variables and hardcoded stylesheet colors.

- [ ] **Step 3: Replace the H5 palette with the shared derivation contract**

Make `createLearnerPalette(primaryColor)` return the complete semantic palette, update `applyLearnerPalette(element, palette, selectionColors)`, and change `TenantThemeContext` to memoize the palette before injection. Keep portal resolution and `tenant-theme-changed` listeners unchanged.

- [ ] **Step 4: Run H5 palette tests**

Run: `cd web/h5 && node --test tests/learnerPalette.test.ts tests/selectionTheme.test.ts`

Expected: palette and selection tests pass; stylesheet test remains red until Task 5.

- [ ] **Step 5: Commit H5 palette support**

```bash
git add web/h5/src/theme/learnerPalette.ts web/h5/src/context/TenantThemeContext.tsx web/h5/tests
git commit -m "feat(learner): inject H5 clay theme variables"
```

### Task 5: H5 learner UI redesign

**Files:**
- Modify: `web/h5/src/pages/HomePage.tsx`
- Modify: `web/h5/src/components/CourseCard.tsx`
- Modify: `web/h5/src/pages/CourseDetailPage.tsx`
- Modify: `web/h5/src/pages/LessonPlayerPage.tsx`
- Modify: `web/h5/src/pages/LoginPage.tsx`
- Modify: `web/h5/src/pages/ForgotPasswordPage.tsx`
- Modify: `web/h5/src/components/CourseMaterials.tsx`
- Modify: `web/h5/src/components/PageShell.tsx`
- Modify: `web/h5/src/styles.css`

**Interfaces:**
- Consumes: existing H5 course/auth/progress APIs, theme context route helpers, complete H5 CSS variables.
- Produces: mobile-first Clay home, course, player, login, and recovery experiences.

- [ ] **Step 1: Restructure the home page without changing data flow**

Add a branded header, continue-learning card sourced from the first course with progress, three compact summary values derived from the loaded course array, filter/section heading, and tactile course cards. Preserve logout and navigation callbacks.

- [ ] **Step 2: Redesign course detail and materials**

Keep `getCourse()` and chapter navigation unchanged. Add Clay cover treatment, metadata/progress summary, molded content-type icons, expandable chapter cards, and a fixed primary learning action with safe-area padding.

- [ ] **Step 3: Redesign the lesson player**

Preserve every lifecycle callback and report function. Add a dark media stage, Clay top navigation, thin pill progress, lesson metadata, and mobile outline navigation sourced from the loaded course.

- [ ] **Step 4: Redesign login and forgot-password pages**

Use a brand-colored Clay header and opaque white form card. Keep validation, submission, redirects, SMS reset states, and error Toasts unchanged.

- [ ] **Step 5: Replace historical H5 CSS with semantic Clay CSS**

Remove literal colors and glass-card rules, use full-pill progress, 20–24px learner cards, 44px controls, safe areas, no more than two low-cost background blobs, and reduced-motion overrides.

- [ ] **Step 6: Run H5 tests and build**

Run: `cd web/h5 && npm test && npm run build`

Expected: all Node/Vitest tests and Vite build pass, and `learnerStyles.test.ts` reports no color literals.

- [ ] **Step 7: Commit H5 redesign**

```bash
git add web/h5/src web/h5/tests
git commit -m "feat(learner): redesign H5 portal with clay UI"
```

### Task 6: Admin palette and global visual contracts

**Files:**
- Modify: `web/admin/src/theme/adminPalette.ts`
- Modify: `web/admin/src/components/AdminThemeProvider.tsx`
- Modify: `web/admin/src/context/AdminThemeContext.tsx`
- Modify: `web/admin/tests/adminPalette.test.ts`
- Create: `web/admin/tests/adminStyles.test.ts`

**Interfaces:**
- Consumes: shared Clay derivation and persisted Admin theme context.
- Produces: `--admin-accent-*`, `--admin-clay-*`, surface/status/shadow variables, and Ant Design tokens for the full Admin application.

- [ ] **Step 1: Add failing Admin palette/style tests**

Require the indigo fallback, all Clay variables, tenant selection variables, primary button contact shadow, Dashboard Clay cards, and explicit flat table/form surface rules. Scan `src/styles.css` for color literals.

- [ ] **Step 2: Run Admin tests and verify RED**

Run: `cd web/admin && node --test tests/adminPalette.test.ts tests/adminStyles.test.ts`

Expected: FAIL because the old coral palette lacks Clay variables and Admin CSS contains hardcoded colors.

- [ ] **Step 3: Extend Admin palette and provider tokens**

Use the shared derivation function, align neutral/status defaults with the cross-platform specification, inject every semantic property, and configure Ant Design Button/Card/Menu/Layout/Table/Input/Modal radius and colors. Keep superadmin fallback behavior and `tenant-theme-changed` reload unchanged.

- [ ] **Step 4: Run Admin palette tests**

Run: `cd web/admin && node --test tests/adminPalette.test.ts`

Expected: palette tests pass; stylesheet tests remain red until Tasks 7–8.

- [ ] **Step 5: Commit Admin theme foundation**

```bash
git add web/admin/src/theme/adminPalette.ts web/admin/src/components/AdminThemeProvider.tsx web/admin/src/context/AdminThemeContext.tsx web/admin/tests
git commit -m "feat(admin): add tenant-aware clay theme foundation"
```

### Task 7: Admin shell, login, and shared component redesign

**Files:**
- Modify: `web/admin/src/layout/AdminLayout.tsx`
- Modify: `web/admin/src/components/PageHeader.tsx`
- Modify: `web/admin/src/pages/Login.tsx`
- Modify: `web/admin/src/pages/ForgotPassword.tsx`
- Modify: `web/admin/src/components/MediaUploader.tsx`
- Modify: `web/admin/src/components/UserImportModal.tsx`
- Modify: `web/admin/src/components/CourseMaterialsManager.tsx`
- Modify: `web/admin/src/components/OfficialCoursePicker.tsx`
- Modify: `web/admin/src/components/RouteErrorPage.tsx`
- Modify: `web/admin/src/styles.css`

**Interfaces:**
- Consumes: Admin theme variables from Task 6 and existing role-based navigation config.
- Produces: responsive Admin shell and reusable surface classes used by every Admin page.

- [ ] **Step 1: Refine the Admin layout structure**

Change the desktop Sider width to 220px, add top-header page context derived from the active menu label, retain 76px collapse and 960px Drawer behavior, and preserve role filtering and logout.

- [ ] **Step 2: Redesign authentication surfaces**

Use a restrained brand panel and tactile primary action for Login/ForgotPassword. Preserve organization selection, validation, login navigation, and password-reset workflow.

- [ ] **Step 3: Standardize shared components**

Add reusable CSS classes for page headings, upload surfaces, empty/error states, modal sections, material rows, picker rows, flat data cards, and light Clay action cards. Remove inline visual styles only where a named class can express the same layout.

- [ ] **Step 4: Rewrite the Admin global CSS cascade**

Consolidate historical overrides into one semantic cascade: restrained 3D on primary buttons, selected nav, metric cards, and quick actions; flat bordered surfaces for tables/forms/modals; consistent focus, responsive behavior, and reduced motion. Eliminate color literals.

- [ ] **Step 5: Run Admin stylesheet contracts and build**

Run: `cd web/admin && node --test tests/adminStyles.test.ts && npm run build`

Expected: style contracts and Vite build pass.

- [ ] **Step 6: Commit shell and shared UI**

```bash
git add web/admin/src/layout web/admin/src/components web/admin/src/pages/Login.tsx web/admin/src/pages/ForgotPassword.tsx web/admin/src/styles.css web/admin/tests/adminStyles.test.ts
git commit -m "feat(admin): redesign shared shell with restrained clay UI"
```

### Task 8: Admin Dashboard, data pages, settings, and course workspace

**Files:**
- Modify: `web/admin/src/pages/Dashboard.tsx`
- Modify: `web/admin/src/pages/ThemeSettings.tsx`
- Modify: `web/admin/src/pages/Courses.tsx`
- Modify: `web/admin/src/pages/OfficialCourses.tsx`
- Modify: `web/admin/src/pages/CourseCategories.tsx`
- Modify: `web/admin/src/pages/Users.tsx`
- Modify: `web/admin/src/pages/Resources.tsx`
- Modify: `web/admin/src/pages/ResourceCategories.tsx`
- Modify: `web/admin/src/pages/Tenants.tsx`
- Modify: `web/admin/src/pages/CreateTenant.tsx`
- Modify: `web/admin/src/pages/Plans.tsx`
- Modify: `web/admin/src/pages/AuditLogs.tsx`
- Modify: `web/admin/src/pages/DomainSettings.tsx`
- Modify: `web/admin/src/pages/SMSConfig.tsx`
- Modify: `web/admin/src/pages/StorageSettings.tsx`
- Modify: `web/admin/src/pages/course-detail/CourseDetailPage.tsx`
- Modify: `web/admin/src/pages/course-detail/CourseSummary.tsx`
- Modify: `web/admin/src/pages/course-detail/CourseOutline.tsx`
- Modify: `web/admin/src/pages/course-detail/LessonEditor.tsx`
- Modify: `web/admin/src/pages/course-detail/EnrollmentManager.tsx`
- Modify: `web/admin/tests/themeSettingsSelectionColors.test.ts`

**Interfaces:**
- Consumes: shared Admin shell/surface classes and unchanged page APIs.
- Produces: consistent visual structure across every routed Admin workflow and a live cross-platform theme preview.

- [ ] **Step 1: Upgrade Dashboard hierarchy**

Apply shallow Clay depth to metric cards and quick actions, molded icon containers, pill progress, and flat ranking/resource lists. Keep all role-specific data models, comparison labels, demo cleanup, and navigation unchanged.

- [ ] **Step 2: Standardize table/data pages**

Wrap page filters/actions in named toolbars, give Table cards flat semantic surfaces, normalize course/user/resource identity cells, status tags, pagination spacing, and responsive overflow. Do not add Clay shadows to table rows.

- [ ] **Step 3: Standardize settings and creation forms**

Use section cards and two-column desktop grids for Domain/SMS/Storage/CreateTenant/Plans while retaining exact form names, validation, API payloads, secret handling, and test-connection behavior.

- [ ] **Step 4: Standardize the course-detail workspace**

Use a restrained Clay summary/header, flat chapter editor and enrollment table, consistent modal/editor sections, and shared material/upload surfaces. Preserve drag/order/edit/resource-preview behavior.

- [ ] **Step 5: Rebuild Theme Settings as four semantic sections**

Create Brand Basics, Color System, Brand Assets, and Live Preview cards. Preview an Admin selected nav item, primary button, learner Hero, progress bar, and Clay contact shadow using local state. Keep `syncSelectionColorsForPrimaryChange()`, contrast warning, upload, reset, API update, and post-save event behavior.

- [ ] **Step 6: Extend Theme Settings tests**

Assert that primary changes only replace recommended selection values, custom values remain stable, and the page source contains preview hooks for Admin nav, learner Hero, button, and progress.

- [ ] **Step 7: Run the full Admin suite and build**

Run: `cd web/admin && npm test && npm run build`

Expected: all Node tests and Vite build pass; style scan is clean.

- [ ] **Step 8: Commit all Admin routed pages**

```bash
git add web/admin/src/pages web/admin/tests/themeSettingsSelectionColors.test.ts
git commit -m "feat(admin): unify management workflows and theme preview"
```

### Task 9: Backend theme contract regression coverage

**Files:**
- Modify: `internal/service/theme_test.go`
- Modify: `internal/api/theme_test.go`
- Modify: `internal/api/portal_test.go`
- Modify only if a failing contract test proves necessary: `internal/service/theme.go`
- Modify only if a failing contract test proves necessary: `internal/api/theme.go`

**Interfaces:**
- Consumes: persisted tenant theme fields.
- Produces: evidence that GET/PUT/portal responses preserve the fields consumed by all three frontends.

- [ ] **Step 1: Add cross-client contract cases**

Add table-driven tests proving update normalization and response inclusion for `primary_color`, three selection colors, `logo_url`, `welcome_text`, `browser_title`, and `brand_name`; verify empty optional values and invalid colors retain current error behavior.

- [ ] **Step 2: Run focused Go tests and verify behavior**

Run: `go test ./internal/service ./internal/api ./internal/repository`

Expected: tests pass without production changes because the backend contract already supports these fields. If a test fails, implement only the smallest contract correction and re-run.

- [ ] **Step 3: Commit backend regression tests**

```bash
git add internal/service/theme_test.go internal/api/theme_test.go internal/api/portal_test.go internal/service/theme.go internal/api/theme.go
git commit -m "test(theme): cover cross-platform tenant theme contract"
```

### Task 10: Final cross-platform verification and handoff

**Files:**
- Verify all files changed in Tasks 1–9.
- Do not modify `.codex/current-task.md`, `.codex/mockups/`, or `.codex/roadmap.md`.

**Interfaces:**
- Consumes: complete implementation.
- Produces: fresh proof for tests, builds, color-variable compliance, and commit scope.

- [ ] **Step 1: Run all frontend test suites**

```bash
cd web/shared && node --test tests/*.test.ts && npx tsc --noEmit
cd ../pc && npm test
cd ../h5 && npm test
cd ../admin && npm test
```

Expected: zero failed tests.

- [ ] **Step 2: Run all frontend production builds**

```bash
cd web/pc && npm run build
cd ../h5 && npm run build
cd ../admin && npm run build
```

Expected: all three Vite builds exit 0; chunk-size warnings are recorded but are not build failures.

- [ ] **Step 3: Run backend regression tests**

Run: `go test ./...`

Expected: all Go packages pass.

- [ ] **Step 4: Run source and diff audits**

```bash
git diff --check
rg -n --glob '*.css' '#[0-9a-fA-F]{3,8}\b|rgba?\(|hsla?\(' web/pc/src/styles web/h5/src/styles.css web/admin/src/styles.css
git status --short
git diff --stat HEAD~9..HEAD
```

Expected: no CSS color literals, no whitespace errors, and only scoped source/test/doc changes plus the user's untouched `.codex/` state.

- [ ] **Step 5: Perform requirement-by-requirement review**

Confirm Admin, PC, and H5 each expose tenant primary and selection variables; learner Clay surfaces use derived contact shadows; Admin data surfaces remain flat; theme settings save/broadcast behavior remains intact; learner heartbeat/API/router code is unchanged except structure required to display existing data.

- [ ] **Step 6: Create a final integration commit only if verification produced fixes**

```bash
git add web internal
git commit -m "fix(theme): close cross-platform clay UI verification gaps"
```

Skip this commit when there are no post-verification changes.
