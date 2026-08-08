# Compact Learner Filter Tabs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the oversized learner course-filter selection with a compact solid capsule and restrained sliding animation.

**Architecture:** Reuse Ant Design Tabs' existing animated ink bar as the moving capsule instead of adding component state or JavaScript. Keep the tab labels above the indicator, preserve tenant selection color variables, and add explicit reduced-motion and narrow-screen rules.

**Tech Stack:** React 18, Ant Design Tabs, CSS, Node test runner, TypeScript, Vite.

## Global Constraints

- The selected capsule is approximately 36px high with an approximately 9px radius.
- Selection colors continue to use tenant theme variables and readable foreground colors.
- The old separate underline is removed.
- Motion is approximately 240ms with a restrained ease-out curve and no bounce.
- `prefers-reduced-motion: reduce` disables movement and press scaling.
- Course data, filtering behavior, routes, and the category selector remain unchanged.
- No new dependency is introduced.

---

### Task 1: Synchronize the feature branch with the current main branch

**Files:**
- Integrate: `origin/main`

**Interfaces:**
- Consumes: the latest learner selection-theme variables and regression tests from `origin/main`
- Produces: a clean feature branch containing both the existing work and current learner UI baseline

- [ ] **Step 1: Confirm the worktree is clean and fetch remotes**

Run: `git status --short && git fetch origin`

Expected: no uncommitted files before integration.

- [ ] **Step 2: Rebase local commits onto current main**

Run: `git rebase origin/main`

Expected: rebase completes without discarding local commits; resolve only overlapping documentation or learner-style changes if Git reports conflicts.

- [ ] **Step 3: Verify the updated baseline**

Run: `git status --short --branch && npm test --prefix web/pc`

Expected: clean branch and all PC tests pass before the new style regression is added.

### Task 2: Specify the compact capsule behavior with a failing regression test

**Files:**
- Modify: `web/pc/tests/selectionTheme.test.ts`

**Interfaces:**
- Consumes: `readStyleBundle()` and existing tenant selection variable assertions
- Produces: regression coverage for capsule geometry, layering, motion, removed underline, and reduced motion

- [ ] **Step 1: Replace the old underline expectation with capsule expectations**

Assert that the bundled stylesheet contains:

```ts
assert.match(stylesheet, /\.learner-filter-tabs \.ant-tabs-ink-bar[^\{]*\{[^}]*height:\s*36px[^}]*border-radius:\s*9px[^}]*background:\s*var\(--tenant-selected-background\)/s)
assert.match(stylesheet, /\.learner-filter-tabs \.ant-tabs-ink-bar-animated[^\{]*\{[^}]*transition:[^}]*240ms[^}]*cubic-bezier\(0\.22,\s*1,\s*0\.36,\s*1\)/s)
assert.match(stylesheet, /@media\s*\(prefers-reduced-motion:\s*reduce\)[\s\S]*\.learner-filter-tabs \.ant-tabs-ink-bar-animated[^\{]*\{[^}]*transition:\s*none/s)
assert.doesNotMatch(stylesheet, /\.learner-filter-tabs \.ant-tabs-tab-active[^\{]*\{[^}]*background:\s*var\(--tenant-selected-background\)/s)
```

- [ ] **Step 2: Run the targeted test and verify it fails for the expected CSS mismatch**

Run: `node --test web/pc/tests/selectionTheme.test.ts`

Expected: FAIL because the current ink bar is still a 3px underline and the active tab still owns the background.

### Task 3: Implement the compact animated capsule

**Files:**
- Modify: `web/pc/src/styles/course.css`
- Modify: `web/pc/src/styles/player.css`

**Interfaces:**
- Consumes: `--tenant-selected-background`, `--tenant-selected-text`, and Ant Design `.ant-tabs-ink-bar-animated`
- Produces: compact desktop capsule plus narrow-screen spacing and reduced-motion behavior

- [ ] **Step 1: Move selection fill from the active tab to the animated ink bar**

Set tabs to a 36px minimum height with compact horizontal padding and `z-index: 1`. Remove the active tab background. Style the ink bar as a 36px-high, 9px-radius selection surface with `var(--tenant-selected-background)` and place it beneath the labels.

- [ ] **Step 2: Add restrained movement and press feedback**

Set `.ant-tabs-ink-bar-animated` width/left/right transitions to `240ms cubic-bezier(0.22, 1, 0.36, 1)`. Add a `180ms` transform transition to the tab button and scale it to `0.97` only while the tab is actively pressed.

- [ ] **Step 3: Add reduced-motion and narrow-screen rules**

Within `@media (prefers-reduced-motion: reduce)`, set the indicator and tab-button transitions to `none` and remove press transforms. Retain compact 36px controls at the existing narrow-screen breakpoint, reduce gaps, keep labels on one line, and allow the Tabs navigation area to handle overflow without clipping text.

- [ ] **Step 4: Run targeted and complete PC verification**

Run: `node --test web/pc/tests/selectionTheme.test.ts && npm test --prefix web/pc && npm run build --prefix web/pc`

Expected: the targeted regression, all PC tests, TypeScript checks, and Vite production build pass.

- [ ] **Step 5: Commit the implementation**

```bash
git add web/pc/tests/selectionTheme.test.ts web/pc/src/styles/course.css web/pc/src/styles/player.css
git commit -m "fix: compact learner filter selection"
```
