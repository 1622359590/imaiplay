# Learner Selection and Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Unify PC learner selected controls at a readable 40px size and replace the completed recent-course action with a clear completion badge.

**Architecture:** Keep tenant selection colors as the single color source and normalize only control geometry in the existing stylesheet layers. Put recent-course completion semantics in a small pure learner utility, then let `RecentCourseCard` render either the existing continue link or a non-interactive completed badge.

**Tech Stack:** React 19, TypeScript, Ant Design 5, CSS, Vitest, Node test runner, Vite

## Global Constraints

- Selected controls are exactly `40px` high with `10px` corner radius and at least `18px` horizontal padding.
- Course detail tabs have a minimum width of `96px` and centered text.
- Tenant-selected background, text, and icon CSS variables remain unchanged.
- A recent course is completed when `progressPercent >= 100`.
- Completed state uses visible “已完成” text plus a green check icon and is not an action.
- No backend, database, H5, or tenant theme contract changes.

---

### Task 1: Normalize Selected Control Geometry

**Files:**
- Modify: `web/pc/tests/selectionTheme.test.ts`
- Modify: `web/pc/src/styles/course.css`
- Modify: `web/pc/src/styles/responsive.css`
- Verify: `web/pc/src/styles/player.css`

**Interfaces:**
- Consumes: existing CSS variables `--tenant-selected-background`, `--tenant-selected-text`, and `--tenant-selected-icon`
- Produces: a shared visual contract of `40px` height, `10px` radius, and stable padding for top navigation, learner filters, and course detail tabs

- [ ] **Step 1: Write failing geometry assertions**

Update the compact filter assertion and add assertions for the other selected controls:

```ts
test('learner selected controls share one readable geometry', () => {
  const stylesheet = readStyleBundle(new URL('../src/styles.css', import.meta.url))

  assert.match(stylesheet, /\.learner-top-nav-link\s*\{[^}]*height:\s*40px[^}]*padding:\s*0\s+18px[^}]*border-radius:\s*10px/s)
  assert.match(stylesheet, /\.learner-filter-tabs \.ant-tabs-tab\s*\{[^}]*min-height:\s*40px[^}]*padding:\s*0\s+18px[^}]*border-radius:\s*10px/s)
  assert.match(stylesheet, /\.course-experience-tabs \.ant-tabs-tab\s*\{[^}]*min-width:\s*96px[^}]*min-height:\s*40px[^}]*padding:\s*0\s+20px[^}]*border-radius:\s*10px/s)
  assert.equal(resolvedLearnerInkBarHeight(stylesheet), '40px')
})
```

Change existing expectations from `36px`/`9px` to `40px`/`10px`.

- [ ] **Step 2: Run the focused style test and verify RED**

Run:

```bash
cd web/pc
node --test --test-name-pattern="selected controls|filter selection" tests/selectionTheme.test.ts
```

Expected: FAIL because current top navigation is `height: 100%`, filters are `36px`, and course tabs have no `96px` minimum width.

- [ ] **Step 3: Implement stable geometry without layout shift**

In `course.css`, give `.learner-top-nav-link` base-state geometry rather than adding padding only when active:

```css
.learner-top-nav-link {
  height: 40px;
  align-items: center;
  justify-content: center;
  padding: 0 18px;
  border-radius: 10px;
}

.learner-top-nav-link.active {
  background: var(--tenant-selected-background);
  color: var(--tenant-selected-text);
}
```

Set filter tabs and the animated ink bar to `40px` high with `10px` radius and `18px` horizontal padding. In `responsive.css`, set course experience tabs to `min-width: 96px`, `min-height: 40px`, `padding: 0 20px`, `border-radius: 10px`, and center the tab button. Preserve the narrow-screen font-size overrides in `player.css`; do not reduce the 40px touch target.

- [ ] **Step 4: Run the focused style test and verify GREEN**

Run the same focused command. Expected: PASS with both geometry tests green.

- [ ] **Step 5: Commit the geometry change**

```bash
git add web/pc/tests/selectionTheme.test.ts web/pc/src/styles/course.css web/pc/src/styles/responsive.css
git commit -m "fix(learner): unify selected control sizing"
```

---

### Task 2: Present Completed Recent Courses Clearly

**Files:**
- Modify: `web/pc/src/utils/learnerCourses.test.ts`
- Modify: `web/pc/src/utils/learnerCourses.ts`
- Modify: `web/pc/src/components/RecentCourseCard.tsx`
- Modify: `web/pc/src/styles/course.css`
- Modify: `web/pc/tests/learnerStyles.test.ts`

**Interfaces:**
- Produces: `recentCourseCompleted(progressPercent: number): boolean`
- Consumes: `RecentLearningItem.progressPercent`
- Produces: `.recent-complete-status` visual badge with icon and “已完成” text

- [ ] **Step 1: Write failing completion semantics tests**

Add to `learnerCourses.test.ts`:

```ts
describe('recent course completion', () => {
  it.each([
    [0, false],
    [99, false],
    [100, true],
    [120, true],
    [Number.NaN, false],
  ])('maps %s percent to completed=%s', (percent, expected) => {
    expect(recentCourseCompleted(percent)).toBe(expected)
  })
})
```

Add a stylesheet assertion to `learnerStyles.test.ts` requiring `.recent-complete-status` to use inline-flex alignment, a circular green icon, and visible text.

- [ ] **Step 2: Run focused tests and verify RED**

Run:

```bash
cd web/pc
npx vitest run src/utils/learnerCourses.test.ts -t "recent course completion"
node --test --test-name-pattern="completed recent-course" tests/learnerStyles.test.ts
```

Expected: FAIL because `recentCourseCompleted` and `.recent-complete-status` do not exist.

- [ ] **Step 3: Implement the pure completion predicate**

Add to `learnerCourses.ts`:

```ts
export function recentCourseCompleted(progressPercent: number): boolean {
  return Number.isFinite(progressPercent) && progressPercent >= 100
}
```

- [ ] **Step 4: Render one mutually exclusive footer status**

In `RecentCourseCard.tsx`, import `CheckCircleFilled` and `recentCourseCompleted`, derive `completed`, then replace the unconditional continue link with:

```tsx
{completed ? (
  <span className="recent-complete-status" aria-label="课程已完成">
    <CheckCircleFilled aria-hidden="true" />
    <span>已完成</span>
  </span>
) : (
  <Link className="recent-continue-link" to={continuePath}>
    <PlayCircleOutlined aria-hidden="true" />
    <span>继续学习</span>
  </Link>
)}
```

Style the badge as a compact green status with a soft green background, readable dark-green text, and a circular filled check icon. Keep it the same vertical rhythm as the existing continue link.

- [ ] **Step 5: Run focused tests and verify GREEN**

Run both focused commands again. Expected: PASS.

- [ ] **Step 6: Commit the completed-state change**

```bash
git add web/pc/src/utils/learnerCourses.test.ts web/pc/src/utils/learnerCourses.ts web/pc/src/components/RecentCourseCard.tsx web/pc/src/styles/course.css web/pc/tests/learnerStyles.test.ts
git commit -m "fix(learner): mark completed recent courses"
```

---

### Task 3: Full Regression and Production Verification

**Files:**
- Verify only: `web/pc`, `web/shared`, `web/admin`, `web/h5`

**Interfaces:**
- Consumes: the completed Task 1 and Task 2 changes
- Produces: evidence that the whole frontend workspace remains buildable

- [ ] **Step 1: Run all frontend tests**

```bash
cd web
npm run test:all
```

Expected: all shared, admin, PC, and H5 tests pass with zero failures.

- [ ] **Step 2: Run all frontend production builds**

```bash
cd web
npm run build:all
```

Expected: admin, PC, and H5 TypeScript checks and Vite builds exit successfully. Existing large-chunk warnings are informational.

- [ ] **Step 3: Verify repository hygiene**

```bash
git diff --check
git status --short
```

Expected: no whitespace errors and only intentional plan/implementation commits.

- [ ] **Step 4: Record final verification**

Do not create a code change for this step. Report test counts, build results, branch name, and the fact that no push occurred unless the user separately requests submission.
