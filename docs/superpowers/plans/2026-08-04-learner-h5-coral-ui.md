# Learner H5 Coral UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the learner H5 home, course detail, lesson player, and login screens use the approved compact light coral design from 320px upward.

**Architecture:** Keep the existing React and Ant Design Mobile page structure. Apply the fixed learner palette to Ant Mobile and consolidate the final H5 CSS cascade so all core screens share the same card, border, spacing, and typography rules.

**Tech Stack:** React 18, TypeScript, Ant Design Mobile 5, React Router 7, Vite, Node test runner, Vitest

## Global Constraints

- Use `#ff5156`, `#e84349`, `#fff1f0`, `#fafafa`, `#ffffff`, `#262626`, `#595959`, `#737373`, and `#eeeeee` exactly as approved.
- Preserve tenant logo, welcome copy, portal routing, login, course retrieval, lesson navigation, and progress behavior.
- Keep one-column mobile layouts with 14–16px side padding and at least 44px interactive targets.
- Do not add a phone frame, new asset, dark gradient, glass effect, or large shadow.

---

### Task 1: Fix Ant Design Mobile to the learner palette

**Files:**
- Modify: `web/h5/tests/learnerPalette.test.ts`
- Modify: `web/h5/src/context/TenantThemeContext.tsx`

**Interfaces:**
- Consumes: `LEARNER_PALETTE` and `applyLearnerPalette()` from `web/h5/src/theme/learnerPalette.ts`.
- Produces: `--adm-color-primary` fixed to `LEARNER_PALETTE.accent`, while context metadata and portal resolution remain unchanged.

- [ ] **Step 1: Add exact palette assertions**

```ts
assert.deepEqual(LEARNER_PALETTE, {
  accent: '#ff5156', accentHover: '#e84349', accentSoft: '#fff1f0',
  heading: '#262626', text: '#595959', muted: '#737373',
  page: '#fafafa', card: '#ffffff', line: '#eeeeee',
})
```

- [ ] **Step 2: Run the focused test**

Run: `cd web/h5 && node --test tests/learnerPalette.test.ts`  
Expected: PASS.

- [ ] **Step 3: Replace dynamic UI color application**

Change the effect to apply the fixed accent:

```ts
useEffect(() => {
  applyLearnerPalette()
  document.documentElement.style.setProperty('--brand-600', LEARNER_PALETTE.accent)
  document.documentElement.style.setProperty('--adm-color-primary', LEARNER_PALETTE.accent)
  document.title = portal ? `${portal.name} | 企业学习中心` : 'iMaiPlay 企业学习中心'
}, [portal])
```

Import `LEARNER_PALETTE`, remove `applyTheme`, and keep `theme.primary_color` in the context value for API compatibility.

- [ ] **Step 4: Run tests and build**

Run: `cd web/h5 && npm test && npm run build`  
Expected: all tests pass and Vite build completes.

- [ ] **Step 5: Commit**

```bash
git add web/h5/tests/learnerPalette.test.ts web/h5/src/context/TenantThemeContext.tsx
git commit -m "refactor(h5): fix learner UI to coral palette"
```

### Task 2: Consolidate compact H5 page layouts

**Files:**
- Modify: `web/h5/src/styles.css`
- Verify: `web/h5/src/pages/HomePage.tsx`
- Verify: `web/h5/src/pages/CourseDetailPage.tsx`
- Verify: `web/h5/src/pages/LessonPlayerPage.tsx`
- Verify: `web/h5/src/components/CourseCard.tsx`

**Interfaces:**
- Consumes: existing learner header, course card, detail, chapter, lesson, and player class names.
- Produces: consistent 16px content gutters, 12px card gaps, 16:9 covers, and no horizontal overflow at 320px.

- [ ] **Step 1: Capture baseline screenshots**

Using the selected browser, capture home, detail, and player at 390×844 and 320×568 under `.qa-artifacts/h5-before/`.

- [ ] **Step 2: Apply the compact surface rules**

Use the existing final `Simplified learner surfaces` section as the single winning cascade:

```css
#root { max-width: 560px; color: var(--learner-text); background: var(--learner-page); box-shadow: none; }
.home-page { padding: 0 16px calc(24px + env(safe-area-inset-bottom)); }
.learner-header { height: 60px; border-bottom: 1px solid var(--learner-line); background: var(--learner-card); }
.course-heading { padding: 22px 0 16px; }
.course-list { gap: 12px; }
.course-card { min-height: 92px; border: 1px solid var(--learner-line); border-radius: 10px; box-shadow: none; }
.detail-content { padding: 14px; }
.detail-cover { width: 100%; aspect-ratio: 16 / 9; border-radius: 10px; }
.detail-summary, .chapter-section { margin-top: 12px; padding: 16px; border: 1px solid var(--learner-line); box-shadow: none; }
.lesson-item { min-height: 48px; padding: 12px 0; }
.mobile-player-progress { margin: 14px; padding: 16px; box-shadow: none; }
```

Remove effective blue gradients, dark tab/nav backgrounds, glow layers, and glass filters from the final cascade. Retain a dark background only behind actual video media.

- [ ] **Step 3: Verify responsive behavior**

Run: `cd web/h5 && npm run build`  
Expected: build passes. At 320×568 and 390×844 verify no horizontal scrolling, no clipped long title, 16:9 covers, safe-area padding, and 44px minimum controls.

- [ ] **Step 4: Commit**

```bash
git add web/h5/src/styles.css
git commit -m "style(h5): compact learner layouts"
```

### Task 3: Convert H5 login to the light coral surface

**Files:**
- Modify: `web/h5/src/styles.css`
- Verify: `web/h5/src/pages/LoginPage.tsx`

**Interfaces:**
- Consumes: existing `.login-page`, `.login-container`, `.login-brand`, `.login-card`, `.login-panel-title`, `.dark-input`, and `.login-help` classes.
- Produces: a safe-area-aware light login form with unchanged tenant and authentication behavior.

- [ ] **Step 1: Add the final login styling**

```css
.login-page { min-height: 100vh; padding: 28px 0; color: var(--learner-text); background: var(--learner-page); }
.login-container { width: 100%; max-width: 420px; padding: 0 20px; }
.login-brand { margin-bottom: 24px; color: var(--learner-heading); }
.brand-logo { color: #fff; border: 0; background: var(--learner-accent); box-shadow: none; }
.login-card { padding: 28px 20px 20px; border: 1px solid var(--learner-line); background: var(--learner-card); box-shadow: 0 8px 24px rgb(38 38 38 / 5%); }
.login-panel-title h2 { color: var(--learner-heading); background: none; -webkit-text-fill-color: currentColor; }
.login-card .adm-list { border-color: #d9d9d9; background: #fff; }
.login-help a, .input-icon { color: var(--learner-accent); }
```

- [ ] **Step 2: Verify form states and build**

Run: `cd web/h5 && npm test && npm run build`  
Expected: zero failures. Inspect default, focus, validation, loading, tenant logo, and forgotten-password states at 320×568 and 390×844.

- [ ] **Step 3: Compare with the reference**

Save `.qa-artifacts/h5-after/login.png`, compare beside the coral reference, and correct visible contrast, spacing, border, radius, and alignment mismatches.

- [ ] **Step 4: Commit**

```bash
git add web/h5/src/styles.css
git commit -m "style(h5): simplify learner login"
```

### Task 4: H5 regression gate

**Files:**
- Modify: `design-qa.md`

**Interfaces:**
- Consumes: completed H5 UI and screenshots.
- Produces: documented H5 evidence for integration.

- [ ] **Step 1: Run the complete H5 gate**

Run: `cd web/h5 && npm test && npm run build`  
Expected: zero failures.

- [ ] **Step 2: Check required states**

Verify login, home, empty course list, detail with and without cover, expanded/collapsed chapters, lesson player, long text, and safe-area behavior at 320×568 and 390×844.

- [ ] **Step 3: Document evidence**

Add a `Learner H5` section to `design-qa.md` with routes, viewport sizes, screenshot paths, test commands, contrast result, and overflow result.

- [ ] **Step 4: Commit**

```bash
git add design-qa.md
git commit -m "test(h5): verify unified coral learner UI"
```

