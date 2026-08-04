# Learner PC Coral UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the learner PC home, course detail, lesson player, and login screens use the approved compact light coral design.

**Architecture:** Keep the existing React and Ant Design component structure. Make `LEARNER_PALETTE` the fixed UI palette while retaining tenant logo and copy data, then replace the final learner CSS cascade with compact responsive rules so old dark-theme declarations cannot win.

**Tech Stack:** React 18, TypeScript, Ant Design 5, React Router, Vite, Node test runner, Vitest

## Global Constraints

- Use `#ff5156` for the primary accent, `#e84349` for hover, and `#fff1f0` for soft accent backgrounds.
- Use `#fafafa` page, `#ffffff` card, `#262626` heading, `#595959` text, `#737373` muted, and `#eeeeee` line colors.
- Preserve existing routes, API calls, authentication, course browsing, lesson navigation, and progress behavior.
- Do not add a UI framework, new route, decorative asset, glass effect, blue-purple gradient, or large shadow.
- Main PC content width is approximately 1080px with at least 20px viewport gutters.

---

### Task 1: Lock the fixed learner palette into Ant Design

**Files:**
- Modify: `web/pc/tests/learnerPalette.test.ts`
- Modify: `web/pc/src/context/TenantThemeContext.tsx`

**Interfaces:**
- Consumes: `LEARNER_PALETTE` and `applyLearnerPalette()` from `web/pc/src/theme/learnerPalette.ts`.
- Produces: Ant Design `colorPrimary` and `colorInfo` fixed to `LEARNER_PALETTE.accent`; tenant `logo_url` and `welcome_text` remain available through `useTenantTheme()`.

- [ ] **Step 1: Add exact palette assertions**

Add these assertions to the existing palette test:

```ts
assert.deepEqual(LEARNER_PALETTE, {
  accent: '#ff5156',
  accentHover: '#e84349',
  accentSoft: '#fff1f0',
  heading: '#262626',
  text: '#595959',
  muted: '#737373',
  page: '#fafafa',
  card: '#ffffff',
  line: '#eeeeee',
})
```

- [ ] **Step 2: Run the focused test**

Run: `cd web/pc && node --test tests/learnerPalette.test.ts`  
Expected: PASS, proving the approved palette is unchanged before provider work.

- [ ] **Step 3: Fix the provider-facing UI color**

Import `LEARNER_PALETTE` beside `applyLearnerPalette`, remove the dynamic `--gradient-brand` assignment, and configure Ant Design as follows:

```tsx
<ConfigProvider
  theme={{
    token: {
      colorPrimary: LEARNER_PALETTE.accent,
      colorInfo: LEARNER_PALETTE.accent,
      colorText: LEARNER_PALETTE.text,
      colorTextHeading: LEARNER_PALETTE.heading,
      colorBgLayout: LEARNER_PALETTE.page,
      colorBgContainer: LEARNER_PALETTE.card,
      colorBorderSecondary: LEARNER_PALETTE.line,
      borderRadius: 10,
    },
  }}
>
  {children}
</ConfigProvider>
```

Keep `theme.primary_color` in the context value for compatibility, but do not use it to color the learner UI.

- [ ] **Step 4: Run tests and build**

Run: `cd web/pc && npm test && npm run build`  
Expected: all Node/Vitest tests pass and Vite emits `dist/` without TypeScript errors.

- [ ] **Step 5: Commit**

```bash
git add web/pc/tests/learnerPalette.test.ts web/pc/src/context/TenantThemeContext.tsx
git commit -m "refactor(pc): fix learner UI to coral palette"
```

### Task 2: Compact the learner content and course detail layout

**Files:**
- Modify: `web/pc/src/styles.css`
- Verify: `web/pc/src/components/AppLayout.tsx`
- Verify: `web/pc/src/pages/HomePage.tsx`
- Verify: `web/pc/src/pages/CourseDetailPage.tsx`
- Verify: `web/pc/src/pages/LessonPlayerPage.tsx`

**Interfaces:**
- Consumes: existing classes `.top-header`, `.main-content`, `.course-card`, `.detail-hero`, `.detail-cover`, `.chapter-card`, `.lesson-row`, and `.player-content`.
- Produces: one aligned 1080px desktop content column and a single-column detail layout below 900px.

- [ ] **Step 1: Record the current visual baseline**

In the user-selected browser, capture the authenticated home, detail, and lesson pages at 1440×900. Store screenshots under `.qa-artifacts/pc-before/` and note visible width, cover size, card padding, and overflow.

- [ ] **Step 2: Replace the final desktop geometry rules**

Update the final learner cascade to use these dimensions:

```css
.top-header { padding-inline: max(20px, calc((100vw - 1080px) / 2)); }
.main-content { width: min(1080px, calc(100% - 40px)); }
.page-section { padding: 28px 0 44px; }
.detail-hero {
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 28px;
  padding: 24px;
  border: 1px solid var(--learner-line);
  border-radius: 12px;
  box-shadow: 0 6px 20px rgb(38 38 38 / 4%);
}
.detail-cover { min-height: 180px; aspect-ratio: 16 / 9; }
.chapter-card { margin-top: 16px; padding: 22px 24px; box-shadow: none; }
.lesson-row { min-height: 48px; padding: 9px 16px; }
```

Set home cards, player card, skeletons, and headings to the approved palette and remove inherited gradients, backdrop filters, glows, and hover shadows from the final cascade.

- [ ] **Step 3: Add responsive rules**

Use this breakpoint behavior:

```css
@media (max-width: 900px) {
  .detail-hero { grid-template-columns: 1fr; gap: 18px; }
  .detail-cover { width: min(100%, 520px); }
}
@media (max-width: 576px) {
  .main-content { width: calc(100% - 28px); }
  .detail-hero, .chapter-card, .player-content { padding: 16px; }
  .detail-cover { width: 100%; min-height: 0; }
}
```

Ensure long titles wrap, the chapter count remains right-aligned on desktop, and lesson links retain at least 44px click height.

- [ ] **Step 4: Build and inspect**

Run: `cd web/pc && npm run build`  
Expected: build passes. Inspect at 1440×900, 1024×768, and 390×844; no horizontal overflow, cover remains 16:9, and detail copy no longer leaves a large empty column.

- [ ] **Step 5: Commit**

```bash
git add web/pc/src/styles.css
git commit -m "style(pc): compact learner layouts"
```

### Task 3: Convert the learner PC login to the light coral surface

**Files:**
- Modify: `web/pc/src/styles.css`
- Verify: `web/pc/src/pages/LoginPage.tsx`

**Interfaces:**
- Consumes: existing `.login-page`, `.login-container`, `.login-brand`, `.login-brand-mark`, `.login-card`, `.dark-input`, `.btn-primary`, and `.login-footer` classes.
- Produces: centered 400–420px accessible light login card with unchanged submit and navigation behavior.

- [ ] **Step 1: Add the final login rules**

```css
.login-page {
  min-height: 100vh;
  padding: 32px 20px;
  color: var(--learner-text);
  background: var(--learner-page);
}
.login-container { width: min(100%, 410px); padding: 0; }
.login-brand { margin-bottom: 24px; color: var(--learner-heading); }
.login-brand-mark { color: #fff; border: 0; background: var(--learner-accent); box-shadow: none; }
.login-card {
  border: 1px solid var(--learner-line);
  background: var(--learner-card);
  box-shadow: 0 10px 30px rgb(38 38 38 / 6%);
  backdrop-filter: none;
}
.login-card .ant-card-body { padding: 32px; }
.login-card .ant-input-affix-wrapper { color: var(--learner-heading); background: #fff; border-color: #d9d9d9; }
.login-card .ant-input::placeholder { color: #8c8c8c; }
.login-card .ant-input-affix-wrapper-focused { border-color: var(--learner-accent); box-shadow: 0 0 0 2px rgb(255 81 86 / 12%); }
```

Replace `.gradient-text` and primary button gradient styling in the login scope with solid approved heading and accent colors.

- [ ] **Step 2: Verify login states**

Run: `cd web/pc && npm test && npm run build`  
Expected: tests and build pass. In the selected browser verify default, focused input, validation error, loading/submit, tenant logo, and forgotten-password link states at 1440×900 and 390×844.

- [ ] **Step 3: Compare against the reference**

Create `.qa-artifacts/pc-after/login.png`, place it beside the supplied coral reference at the same viewport, and correct any mismatched contrast, padding, radius, or alignment before continuing.

- [ ] **Step 4: Commit**

```bash
git add web/pc/src/styles.css
git commit -m "style(pc): simplify learner login"
```

### Task 4: PC regression gate

**Files:**
- Modify: `design-qa.md`

**Interfaces:**
- Consumes: completed PC UI and browser screenshots.
- Produces: recorded PC viewport evidence and a passing regression gate for integration.

- [ ] **Step 1: Run the complete PC gate**

Run: `cd web/pc && npm test && npm run build`  
Expected: zero failures.

- [ ] **Step 2: Check required page states**

Verify home with one and multiple courses, detail with and without cover, expanded/collapsed chapters, lesson player, login, empty course list, and a long title. Check 1440×900, 1024×768, and 390×844.

- [ ] **Step 3: Document evidence**

Add a `Learner PC` section to `design-qa.md` listing routes, viewport sizes, screenshot paths, test commands, and confirmation that text contrast and horizontal overflow passed.

- [ ] **Step 4: Commit**

```bash
git add design-qa.md
git commit -m "test(pc): verify unified coral learner UI"
```

