# Learner Light Coral Palette Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every simplified PC and H5 learner surface readable on a light background and match the approved coral-red reference palette.

**Architecture:** Keep tenant branding and login styling intact, and introduce learner-only palette variables in the final simplified CSS cascade. Add source-level regression tests that assert the fixed learner palette and the high-specificity text overrides needed to prevent the legacy dark-theme cascade from returning.

**Tech Stack:** React 18, TypeScript, Ant Design, Ant Design Mobile, CSS, Node.js test runner, Vite.

## Global Constraints

- Use `#ff5156` as the learner accent, `#e84349` as its hover state, and `#fff1f0` as its soft background.
- Use `#262626` for headings and course names, `#595959` for body text, and `#737373` for secondary text.
- Use `#fafafa` for page backgrounds, `#ffffff` for cards, and `#eeeeee` for borders.
- Apply the palette to PC and H5 learner surfaces only; preserve existing tenant branding on login and organization selection.
- Do not add statistics, filters, inferred progress, navigation items, routes, or backend changes.

---

### Task 1: PC Learner Palette

**Files:**
- Create: `web/pc/tests/learnerPalette.test.ts`
- Create: `web/pc/src/theme/learnerPalette.ts`
- Modify: `web/pc/src/context/TenantThemeContext.tsx`
- Modify: `web/pc/src/styles.css`

**Interfaces:**
- Consumes: existing `.app-shell`, `.course-home-heading`, `.course-card`, `.detail-*`, `.chapter-*`, and `.player-*` selectors.
- Produces: `LEARNER_PALETTE` and learner-only CSS variables `--learner-accent`, `--learner-accent-hover`, `--learner-accent-soft`, `--learner-heading`, `--learner-text`, `--learner-muted`, `--learner-page`, `--learner-card`, and `--learner-line`.

- [ ] **Step 1: Write the failing palette regression test**

Create `web/pc/tests/learnerPalette.test.ts` to import the real `LEARNER_PALETTE`, calculate WCAG contrast independently in the test, and assert minimum contrast for headings, body text, and secondary text against learner card/page backgrounds.

- [ ] **Step 2: Run the focused test and verify it fails**

Run: `node --test tests/learnerPalette.test.ts`

Expected: FAIL because `src/theme/learnerPalette.ts` does not exist.

- [ ] **Step 3: Add the PC learner palette**

Create `LEARNER_PALETTE`, expose its values as CSS custom properties from `TenantThemeContext`, then update the final `/* Simplified learner surfaces */` section in `web/pc/src/styles.css`. Replace learner page/card/border/accent colors and add selectors at least as specific as the legacy `.app-shell ...` dark-theme selectors:

```css
.app-shell .course-home-heading h1,
.app-shell .course-card h4.ant-typography,
.app-shell .detail-summary h1,
.app-shell .chapter-card > h2,
.app-shell .player-page > h2 { color: var(--learner-heading); }

.app-shell .course-home-heading .ant-typography-secondary,
.app-shell .course-card .ant-typography-secondary { color: var(--learner-muted); }
```

- [ ] **Step 4: Run the PC test suite and build**

Run: `npm test && npm run build`

Expected: all Node/Vitest tests pass and Vite production build exits 0.

- [ ] **Step 5: Commit the PC palette**

```bash
git add web/pc/src/theme/learnerPalette.ts web/pc/src/context/TenantThemeContext.tsx web/pc/src/styles.css web/pc/tests/learnerPalette.test.ts
git commit -m "fix(pc): improve learner palette contrast"
```

### Task 2: H5 Learner Palette

**Files:**
- Create: `web/h5/tests/learnerPalette.test.ts`
- Create: `web/h5/src/theme/learnerPalette.ts`
- Modify: `web/h5/src/context/TenantThemeContext.tsx`
- Modify: `web/h5/src/styles.css`

**Interfaces:**
- Consumes: existing simplified `#root`, `.learner-header`, `.course-heading`, `.course-card`, `.detail-*`, `.chapter-*`, and `.player-*` selectors.
- Produces: `LEARNER_PALETTE` plus the same learner-only variable names and color meanings as PC.

- [ ] **Step 1: Write the failing H5 palette regression test**

Create `web/h5/tests/learnerPalette.test.ts` to import the real palette and independently verify heading, body, and secondary-text contrast against the learner surfaces.

- [ ] **Step 2: Run the focused test and verify it fails**

Run: `node --test tests/learnerPalette.test.ts`

Expected: FAIL because `src/theme/learnerPalette.ts` does not exist.

- [ ] **Step 3: Add the H5 learner palette**

Create the H5 `LEARNER_PALETTE`, expose it through `TenantThemeContext`, and update the final simplified learner CSS to use those variables. Explicitly set readable heading/body/secondary text colors, use white cards with `#eeeeee` borders, and use coral for learner icons, empty covers, progress, active and focus states.

- [ ] **Step 4: Run the H5 test suite and build**

Run: `npm test && npm run build`

Expected: all Node/Vitest tests pass and Vite production build exits 0.

- [ ] **Step 5: Commit the H5 palette**

```bash
git add web/h5/src/theme/learnerPalette.ts web/h5/src/context/TenantThemeContext.tsx web/h5/src/styles.css web/h5/tests/learnerPalette.test.ts
git commit -m "fix(h5): align learner palette contrast"
```

### Task 3: Visual and Release Verification

**Files:**
- Modify only if QA finds a palette defect: `web/pc/src/styles.css`, `web/h5/src/styles.css`
- Create: `design-qa.md`

**Interfaces:**
- Consumes: built PC/H5 applications and the approved reference screenshot.
- Produces: a `design-qa.md` report with `final result: passed` or a concrete blocking reason.

- [ ] **Step 1: Run fresh complete verification**

Run PC and H5 tests/builds again plus `git diff --check`.

- [ ] **Step 2: Open the rendered learner pages**

Run the local frontend, inspect PC and H5 at their relevant viewport sizes, and compare title, secondary text, course name, course count, cover, card, border, and interaction colors with the approved screenshot.

- [ ] **Step 3: Record design QA**

Write `design-qa.md` with the reference, inspected routes/viewports, findings, remaining P3 notes if any, and the exact final line `final result: passed` only when no P0/P1/P2 issue remains.

- [ ] **Step 4: Commit QA evidence and any final corrections**

```bash
git add design-qa.md web/pc/src/styles.css web/h5/src/styles.css
git commit -m "test: verify learner palette design"
```
