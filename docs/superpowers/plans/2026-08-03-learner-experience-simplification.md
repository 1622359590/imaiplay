# Learner Experience Simplification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the PC and H5 learner dashboards with a minimal course list → course detail → lesson learning flow that uses only real data from existing APIs.

**Architecture:** Keep `web/pc` and `web/h5` independent and preserve their existing React, router, and Ant Design stacks. Both clients use the published-course list as the only learner home page, enrich course cards with real lesson counts through the existing detail endpoint, and leave progress display exclusively on the lesson player because the existing APIs do not expose aggregate course progress.

**Tech Stack:** React 18, TypeScript, Vite 6, React Router, Ant Design 5, Ant Design Mobile 5, Axios, Vitest.

## Global Constraints

- Do not modify backend APIs, database structures, the admin frontend, authentication behavior, or learning-progress business rules.
- Do not display or infer aggregate course progress, lesson completion state, learning time, certificate counts, weekly plans, or other unavailable data.
- Course cards show only cover, title, and a real lesson count; hide the count if detail enrichment fails instead of showing a fabricated zero.
- The only learner flow is `/` → `/courses/:id` → `/courses/:courseId/lessons/:lessonId`.
- Preserve tenant logo and primary-color theming.
- Preserve video progress retrieval and reporting and the existing document-resource behavior.
- Use the existing code style in each app: semicolons in `web/pc`, no semicolons in `web/h5`.
- Do not alter pre-existing unrelated worktree changes, especially the current package-lock edits, without reviewing their diff first.

---

## File Map

### Shared testing setup

- Modify `web/pc/package.json` and `web/pc/package-lock.json`: add the one-shot `test` script and Vitest.
- Modify `web/h5/package.json` and `web/h5/package-lock.json`: add the one-shot `test` script and Vitest.

### PC learner app

- Modify `web/pc/src/api/course.ts`: enrich list items with real lesson counts through `getCourse` and tolerate individual detail failures.
- Create `web/pc/src/api/course.test.ts`: cover lesson-count aggregation and partial enrichment failure.
- Modify `web/pc/src/router.tsx`: make `/` the only course-list page and redirect legacy list routes.
- Modify `web/pc/src/pages/HomePage.tsx`: render the minimal “我的课程” page.
- Delete `web/pc/src/pages/CoursesPage.tsx`: remove the duplicate list and search experience.
- Delete `web/pc/src/pages/RecentPage.tsx`: remove the separate recent-learning page.
- Modify `web/pc/src/components/AppLayout.tsx`: reduce the shell to tenant brand, logout, and content.
- Modify `web/pc/src/components/CourseCard.tsx`: remove metadata, progress, category, and CTA duplication.
- Modify `web/pc/src/components/CourseGrid.tsx`: use the simplified responsive grid and empty copy.
- Modify `web/pc/src/pages/CourseDetailPage.tsx`: remove aggregate progress assumptions and direct learners to individual lessons.
- Modify `web/pc/src/pages/LessonPlayerPage.tsx`: simplify page chrome while retaining resource and progress behavior.
- Modify `web/pc/src/styles.css`: replace dashboard-specific styling with the minimal shell, list, detail, and player styles while preserving login styles.

### H5 learner app

- Modify `web/h5/src/types/course.ts`: make lesson count optional until detail enrichment succeeds and remove unused presentation fields from card usage.
- Modify `web/h5/src/api/course.ts`: enrich list items with real lesson counts through `getCourse` and tolerate individual detail failures.
- Create `web/h5/src/api/course.test.ts`: cover lesson-count aggregation and partial enrichment failure.
- Modify `web/h5/src/App.tsx`: redirect legacy `/courses` and `/profile` routes to `/`.
- Modify `web/h5/src/components/PageShell.tsx`: remove bottom-tab composition.
- Delete `web/h5/src/components/AppTabBar.tsx`: remove learner bottom navigation.
- Modify `web/h5/src/pages/HomePage.tsx`: render tenant brand, logout, and a single course list.
- Delete `web/h5/src/pages/CoursesPage.tsx`: remove duplicate course browsing and filters.
- Delete `web/h5/src/pages/ProfilePage.tsx`: remove fabricated achievements and inactive service entries.
- Modify `web/h5/src/components/CourseCard.tsx`: show cover, title, and optional real lesson count only.
- Modify `web/h5/src/pages/CourseDetailPage.tsx`: remove aggregate progress and the unreliable continue action.
- Modify `web/h5/src/pages/LessonPlayerPage.tsx`: keep content and real lesson progress, with simpler chrome.
- Modify `web/h5/src/styles.css`: remove dashboard, bottom-tab, statistics, profile, and decorative detail styles while preserving login styles and mobile safe areas.

---

### Task 1: Add a Minimal Unit-Test Runner to Both Frontends

**Files:**
- Modify: `web/pc/package.json`
- Modify: `web/pc/package-lock.json`
- Modify: `web/h5/package.json`
- Modify: `web/h5/package-lock.json`

**Interfaces:**
- Consumes: existing Vite 6 and TypeScript configurations in each frontend.
- Produces: `npm test -- --run` in both frontend directories, backed by Vitest.

- [ ] **Step 1: Review existing lockfile changes before installing**

Run:

```bash
git diff -- web/pc/package-lock.json web/h5/package-lock.json
```

Expected: identify and preserve all pre-existing lockfile edits; do not overwrite or revert them.

- [ ] **Step 2: Add the test script and Vitest to the PC app**

Run:

```bash
cd web/pc
npm install --save-dev vitest
npm pkg set scripts.test="vitest run"
```

Expected: `package.json` contains `"test": "vitest run"`, and the lockfile retains its prior changes plus Vitest dependencies. Vite 6 and Node 26 satisfy Vitest's documented minimums.

- [ ] **Step 3: Add the test script and Vitest to the H5 app**

Run:

```bash
cd web/h5
npm install --save-dev vitest
npm pkg set scripts.test="vitest run"
```

Expected: `package.json` contains `"test": "vitest run"`, and the lockfile retains its prior changes plus Vitest dependencies.

- [ ] **Step 4: Verify both test runners start cleanly**

Run:

```bash
cd web/pc && npm test -- --passWithNoTests
cd web/h5 && npm test -- --passWithNoTests
```

Expected: both commands exit successfully with no tests found.

- [ ] **Step 5: Commit the testing foundation**

```bash
git add web/pc/package.json web/pc/package-lock.json web/h5/package.json web/h5/package-lock.json
git commit -m "test: add learner frontend unit test runners"
```

---

### Task 2: Enrich PC Course Cards With Real Lesson Counts

**Files:**
- Modify: `web/pc/src/api/course.ts`
- Create: `web/pc/src/api/course.test.ts`

**Interfaces:**
- Consumes: `getCourse(id: string): Promise<Course>` and `Course.chapters?: Chapter[]`.
- Produces: `countLessons(chapters?: Chapter[]): number` and `enrichLessonCounts(courses: Course[], loadDetail: (id: string) => Promise<Course>): Promise<Course[]>`.
- Produces: `getCourses(): Promise<Course[]>` where `lesson_count` is real when detail loading succeeds and remains `undefined` when it fails.

- [ ] **Step 1: Write failing aggregation and failure-tolerance tests**

Create `web/pc/src/api/course.test.ts`:

```ts
import { describe, expect, it, vi } from 'vitest';
import { countLessons, enrichLessonCounts, type Course } from './course';

describe('PC learner course presentation data', () => {
  it('counts lessons across chapters', () => {
    expect(countLessons([
      { id: 'chapter-1', title: '第一章', lessons: [
        { id: 'lesson-1', title: '课时一' },
        { id: 'lesson-2', title: '课时二' },
      ] },
      { id: 'chapter-2', title: '第二章', lessons: [
        { id: 'lesson-3', title: '课时三' },
      ] },
    ])).toBe(3);
  });

  it('keeps a course usable when detail enrichment fails', async () => {
    const courses: Course[] = [
      { id: 'ok', title: '可用课程' },
      { id: 'failed', title: '详情失败课程' },
    ];
    const loadDetail = vi.fn(async (id: string): Promise<Course> => {
      if (id === 'failed') throw new Error('detail failed');
      return { ...courses[0], chapters: [
        { id: 'chapter', title: '章节', lessons: [{ id: 'lesson', title: '课时' }] },
      ] };
    });

    await expect(enrichLessonCounts(courses, loadDetail)).resolves.toEqual([
      { id: 'ok', title: '可用课程', lesson_count: 1 },
      { id: 'failed', title: '详情失败课程' },
    ]);
  });
});
```

- [ ] **Step 2: Run the PC test and verify it fails**

Run:

```bash
cd web/pc && npm test -- src/api/course.test.ts
```

Expected: FAIL because `countLessons` and `enrichLessonCounts` are not exported.

- [ ] **Step 3: Implement count enrichment without fake fallbacks**

Add to `web/pc/src/api/course.ts`:

```ts
export function countLessons(chapters: Chapter[] = []): number {
  return chapters.reduce((total, chapter) => total + chapter.lessons.length, 0);
}

export async function enrichLessonCounts(
  courses: Course[],
  loadDetail: (id: string) => Promise<Course>,
): Promise<Course[]> {
  return Promise.all(courses.map(async (course) => {
    try {
      const detail = await loadDetail(course.id);
      return { ...course, lesson_count: countLessons(detail.chapters) };
    } catch {
      return course;
    }
  }));
}
```

Then change `getCourses` to normalize the list first and return:

```ts
return enrichLessonCounts(normalizeList(response.data), getCourse);
```

Do not set `progress`, duration, instructor, category, or a zero lesson count in list normalization.

- [ ] **Step 4: Run the focused test and PC build**

Run:

```bash
cd web/pc && npm test -- src/api/course.test.ts && npm run build
```

Expected: both tests PASS and the production build completes.

- [ ] **Step 5: Commit PC data enrichment**

```bash
git add web/pc/src/api/course.ts web/pc/src/api/course.test.ts
git commit -m "feat(pc): load real course lesson counts"
```

---

### Task 3: Replace the PC Dashboard With the Minimal Course Flow

**Files:**
- Modify: `web/pc/src/router.tsx`
- Modify: `web/pc/src/pages/HomePage.tsx`
- Delete: `web/pc/src/pages/CoursesPage.tsx`
- Delete: `web/pc/src/pages/RecentPage.tsx`
- Modify: `web/pc/src/components/AppLayout.tsx`
- Modify: `web/pc/src/components/CourseCard.tsx`
- Modify: `web/pc/src/components/CourseGrid.tsx`
- Modify: `web/pc/src/pages/CourseDetailPage.tsx`
- Modify: `web/pc/src/pages/LessonPlayerPage.tsx`
- Modify: `web/pc/src/styles.css`

**Interfaces:**
- Consumes: `getCourses(): Promise<Course[]>`, `getCourse(id)`, `Course.lesson_count?: number`, existing auth logout, tenant theme, and lesson progress APIs.
- Produces: legacy route redirects, a single PC learner home page, a chapter-based lesson chooser, and the unchanged lesson-progress data flow.

- [ ] **Step 1: Capture the current PC build baseline**

Run:

```bash
cd web/pc && npm run build
```

Expected: PASS before UI edits. If it fails, record the pre-existing error and do not attribute it to this task.

- [ ] **Step 2: Collapse PC routing to one list page**

In `web/pc/src/router.tsx`:

- Remove lazy imports for `CoursesPage` and `RecentPage`.
- Keep `HomePage`, `CourseDetailPage`, and `LessonPlayerPage`.
- Replace the old list routes with redirects:

```tsx
{ index: true, element: <HomePage /> },
{ path: 'courses', element: <Navigate to="/" replace /> },
{ path: 'recent', element: <Navigate to="/" replace /> },
{ path: 'courses/:courseId', element: <CourseDetailPage /> },
{ path: 'courses/:courseId/lessons/:lessonId', element: <LessonPlayerPage /> },
```

Delete `CoursesPage.tsx` and `RecentPage.tsx` after no imports remain.

- [ ] **Step 3: Reduce the PC shell to brand and logout**

In `AppLayout.tsx`:

- Remove `Menu`, nav items, avatar, role label, footer, and their icons.
- Keep the tenant logo or `ReadOutlined` fallback, tenant welcome text or `iMaiPlay`, and a text logout button.
- Render only `<Header>` and `<Content><Outlet /></Content>`.
- Give the logo image `alt="租户 Logo"` and keep the brand button navigating to `/`.

The shell should have this structure:

```tsx
<Layout className="app-shell">
  <Header className="top-header">
    <button className="brand" type="button" onClick={() => navigate('/')}>{/* tenant brand */}</button>
    <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>退出</Button>
  </Header>
  <Content className="main-content"><Outlet /></Content>
</Layout>
```

- [ ] **Step 4: Replace the PC home page with the only course list**

In `HomePage.tsx`, remove the hero, statistics, learning filters, theme welcome copy, and CTA. Render:

```tsx
<section className="course-home page-section">
  <div className="course-home-heading">
    <div>
      <Typography.Title level={1}>我的课程</Typography.Title>
      <Typography.Text type="secondary">选择课程开始学习</Typography.Text>
    </div>
    {!loading && <Typography.Text type="secondary">共 {courses.length} 门</Typography.Text>}
  </div>
  <CourseGrid courses={courses} loading={loading} emptyText="暂无可学习课程" />
</section>
```

- [ ] **Step 5: Simplify PC course cards and the grid**

In `CourseCard.tsx`:

- Keep cover, title, optional real `lesson_count`, and whole-card navigation.
- Remove description, instructor/category metadata, duration, student count, progress, and the button.
- When `lesson_count === undefined`, render no count label; when it is `0`, render `0 课时` because detail enrichment proved the value is real.

Use this conditional:

```tsx
{course.lesson_count !== undefined && (
  <Typography.Text type="secondary">{course.lesson_count} 课时</Typography.Text>
)}
```

In `CourseGrid.tsx`, use `xs={24} sm={12} lg={8}` and keep the existing skeleton and empty-state behavior.

- [ ] **Step 6: Simplify PC course detail without fake progress**

In `CourseDetailPage.tsx`:

- Replace the breadcrumb with one `返回我的课程` link.
- Keep cover, title, description, and the real lesson count computed from `chapters`.
- Remove category, total duration, student count, aggregate progress, `learned` icons, and the start/continue button.
- Make each lesson row itself a `Link` to its player and show only sequence, title, and optional real duration.
- For a missing course, render “课程不存在或暂时无法访问” plus a button linking to `/`.
- For no chapters or lessons, render “暂无可学习课时”.

Calculate the count locally from the already-loaded detail:

```ts
const lessonCount = (course.chapters ?? []).reduce(
  (total, chapter) => total + chapter.lessons.length,
  0,
);
```

- [ ] **Step 7: Keep the PC player behavior and remove decorative card chrome**

In `LessonPlayerPage.tsx`, retain `getLessonProgress`, `getResourceFile`, `reportLessonProgress`, video event handlers, document link, resource-missing state, and the 0–100 progress display. Replace the outer decorative card with a plain content section and keep a visible `返回课程目录` action.

- [ ] **Step 8: Replace PC dashboard CSS with minimal responsive styles**

In `styles.css`:

- Preserve login-page selectors and base token imports.
- Remove `.hero-*`, `.stat-*`, `.top-nav`, `.user-actions`, `.app-footer`, category tag, course metadata, and course-progress selectors that no longer have markup.
- Set `.app-shell` to a plain `#f7f7f7` background with no radial gradient.
- Set `.top-header` to two columns (`1fr auto`), 64px height, white background, and one bottom border.
- Set `.main-content` to `min(1120px, calc(100% - 40px))`.
- Give `.course-home` 32px top/bottom padding and `.course-home-heading` a simple flex layout.
- Give `.course-card` a 1px neutral border, no large shadow, and a restrained hover border change.
- Use `aspect-ratio: 16 / 9` for `.course-cover`.
- At `max-width: 576px`, reduce page/header padding and stack the course heading.
- Keep focus-visible outlines on the brand, course cards, links, and buttons.

- [ ] **Step 9: Verify the PC flow**

Run:

```bash
cd web/pc && npm test && npm run build
```

Manual checks at `/pc/`:

- `/pc/` shows only tenant brand, logout, title, count, and course cards.
- `/pc/courses` and `/pc/recent` redirect to `/pc/`.
- A failed detail enrichment hides lesson count rather than displaying `0`.
- Course detail shows no progress or completion state and every lesson link works.
- Video progress loads and reports; document links still open; missing resources show the existing message.
- Keyboard focus is visible and the layout works at 375px and 1440px widths.

- [ ] **Step 10: Commit the PC simplification**

```bash
git add web/pc/src/router.tsx web/pc/src/pages web/pc/src/components web/pc/src/styles.css
git commit -m "feat(pc): simplify learner course flow"
```

---

### Task 4: Enrich H5 Course Cards With Real Lesson Counts

**Files:**
- Modify: `web/h5/src/types/course.ts`
- Modify: `web/h5/src/api/course.ts`
- Create: `web/h5/src/api/course.test.ts`

**Interfaces:**
- Consumes: `getCourse(id: string): Promise<Course>` and `Course.chapters?: Chapter[]`.
- Produces: `countLessons(chapters?: Chapter[]): number`, `enrichLessonCounts(courses, loadDetail)`, and optional `Course.lessonCount?: number`.

- [ ] **Step 1: Write failing H5 aggregation and failure-tolerance tests**

Create `web/h5/src/api/course.test.ts`:

```ts
import { describe, expect, it, vi } from 'vitest'
import { countLessons, enrichLessonCounts } from './course'
import type { Course } from '../types/course'

describe('H5 learner course presentation data', () => {
  it('counts lessons across chapters', () => {
    expect(countLessons([
      { id: 'chapter-1', title: '第一章', lessons: [
        { id: 'lesson-1', title: '课时一', duration: 2 },
        { id: 'lesson-2', title: '课时二', duration: 3 },
      ] },
    ])).toBe(2)
  })

  it('does not invent a count when detail loading fails', async () => {
    const courses: Course[] = [
      { id: 'ok', title: '可用课程', description: '', cover: '', instructor: '', progress: 0, duration: 0, category: '' },
      { id: 'failed', title: '详情失败课程', description: '', cover: '', instructor: '', progress: 0, duration: 0, category: '' },
    ]
    const loadDetail = vi.fn(async (id: string): Promise<Course> => {
      if (id === 'failed') throw new Error('detail failed')
      return { ...courses[0], chapters: [
        { id: 'chapter', title: '章节', lessons: [{ id: 'lesson', title: '课时', duration: 1 }] },
      ] }
    })

    const result = await enrichLessonCounts(courses, loadDetail)
    expect(result[0].lessonCount).toBe(1)
    expect(result[1].lessonCount).toBeUndefined()
  })
})
```

- [ ] **Step 2: Run the H5 test and verify it fails**

Run:

```bash
cd web/h5 && npm test -- src/api/course.test.ts
```

Expected: FAIL because the helpers are not exported and `lessonCount` is currently required.

- [ ] **Step 3: Implement optional real lesson counts**

In `types/course.ts`, change `lessonCount: number` to `lessonCount?: number`.

In `api/course.ts`, add:

```ts
export function countLessons(chapters: Chapter[] = []): number {
  return chapters.reduce((total, chapter) => total + chapter.lessons.length, 0)
}

export async function enrichLessonCounts(
  courses: Course[],
  loadDetail: (id: string) => Promise<Course>,
): Promise<Course[]> {
  return Promise.all(courses.map(async (course) => {
    try {
      const detail = await loadDetail(course.id)
      return { ...course, lessonCount: countLessons(detail.chapters) }
    } catch {
      const { lessonCount: _ignored, ...withoutCount } = course
      return withoutCount
    }
  }))
}
```

Import `Chapter`, remove the hardcoded `lessonCount: 0` from `mapCourse`, and return enriched items from `getCourses`:

```ts
return { items: await enrichLessonCounts(payload.items.map(mapCourse), getCourse), total: payload.total }
```

Keep H5's current `cover` mapping because it provides either a real URL background or a brand fallback, but do not render the hardcoded instructor, progress, duration, or category fields in simplified pages.

- [ ] **Step 4: Run the focused H5 test and build**

Run:

```bash
cd web/h5 && npm test -- src/api/course.test.ts && npm run build
```

Expected: tests PASS and the production build completes.

- [ ] **Step 5: Commit H5 data enrichment**

```bash
git add web/h5/src/types/course.ts web/h5/src/api/course.ts web/h5/src/api/course.test.ts
git commit -m "feat(h5): load real course lesson counts"
```

---

### Task 5: Replace the H5 Dashboard With the Minimal Course Flow

**Files:**
- Modify: `web/h5/src/App.tsx`
- Modify: `web/h5/src/components/PageShell.tsx`
- Delete: `web/h5/src/components/AppTabBar.tsx`
- Modify: `web/h5/src/pages/HomePage.tsx`
- Delete: `web/h5/src/pages/CoursesPage.tsx`
- Delete: `web/h5/src/pages/ProfilePage.tsx`
- Modify: `web/h5/src/components/CourseCard.tsx`
- Modify: `web/h5/src/pages/CourseDetailPage.tsx`
- Modify: `web/h5/src/pages/LessonPlayerPage.tsx`
- Modify: `web/h5/src/styles.css`

**Interfaces:**
- Consumes: `getCourses(): Promise<CourseList>`, optional `lessonCount`, tenant theme, auth logout, course detail, resources, and lesson progress APIs.
- Produces: one H5 home list, legacy redirects, a chapter-based lesson chooser, and unchanged lesson-player progress behavior.

- [ ] **Step 1: Capture the current H5 build baseline**

Run:

```bash
cd web/h5 && npm run build
```

Expected: PASS before UI edits. Record any pre-existing failure separately.

- [ ] **Step 2: Remove duplicate H5 routes and bottom navigation**

In `App.tsx`:

- Remove lazy imports for `CoursesPage` and `ProfilePage`.
- Replace those route elements with redirects:

```tsx
<Route index element={<HomePage />} />
<Route path="/courses" element={<Navigate to="/" replace />} />
<Route path="/profile" element={<Navigate to="/" replace />} />
```

In `PageShell.tsx`, remove `AppTabBar`, theme-only data attributes, and bottom padding assumptions:

```tsx
export function PageShell() {
  return <main className="page-content"><Outlet /></main>
}
```

Delete `AppTabBar.tsx`, `CoursesPage.tsx`, and `ProfilePage.tsx` once no imports remain.

- [ ] **Step 3: Build the single H5 course home**

In `HomePage.tsx`:

- Keep course loading from `getCourses`.
- Use tenant Logo/welcome text only for brand identity, not a decorative welcome hero.
- Import `logout` from `api/auth` and navigate to `/login` with `{ replace: true }` after logout.
- Render a compact header with logo/name and logout, followed by title, optional real total, and the course list.
- Remove greeting, notification, weekly plan, score, statistics, and “继续学习” decorative section.
- For load failure or no items, show “暂无可学习课程”.

Required state rendering:

```tsx
{loading ? (
  <div className="loading-state"><DotLoading color="primary" /> 正在加载课程</div>
) : courses.length ? (
  <div className="course-list">
    {courses.map((course) => <CourseCard key={course.id} course={course} />)}
  </div>
) : (
  <div className="empty-state">暂无可学习课程</div>
)}
```

- [ ] **Step 4: Simplify H5 course cards**

In `CourseCard.tsx`:

- Render the whole card as a semantic `<button type="button">` and keep click navigation and the course cover.
- Keep only the title and optional real lesson count.
- Remove category, cover mark, instructor, duration, progress bar, status, and chevron.
- Add an accessible label containing the course title; native button keyboard behavior supplies Enter/Space activation without custom key handlers.

Render the count only when known:

```tsx
{course.lessonCount !== undefined && <p>{course.lessonCount} 课时</p>}
```

- [ ] **Step 5: Simplify H5 course detail**

In `CourseDetailPage.tsx`:

- Keep `NavBar`, title, description, cover, real lesson count, chapters, and lesson navigation.
- Remove category, instructor, “随时学习”, aggregate progress, completion icons, and the fixed bottom start/continue action.
- Make every lesson row a button or keyboard-accessible link containing sequence, title, and optional real duration.
- Show “暂无可学习课时” when the flattened lesson list is empty.
- Keep the existing “课程不可访问” error state and add a visible return-to-home action.

- [ ] **Step 6: Keep H5 lesson progress and simplify chrome**

In `LessonPlayerPage.tsx`, retain resource loading/revocation, video resume position, 5% reporting cadence, forced pause reporting, completion reporting, document opening, and resource-missing state. Keep `NavBar`, learning content, and the real lesson progress bar; remove any decorative wrapper that does not aid learning.

- [ ] **Step 7: Replace H5 dashboard CSS with minimal mobile styles**

In `styles.css`:

- Preserve base tokens, login selectors, lesson media selectors, safe-area padding, and the centered 560px development frame.
- Remove `.app-tabbar`, `.hero-card`, `.hero-score`, `.quick-stats`, `.section-heading` eyebrow styles, `.progress-row`, `.profile-*`, `.detail-progress`, `.detail-action`, decorative category tags, and fixed bottom-action rules.
- Use a white `.learner-header` with one bottom border, tenant mark on the left, and logout on the right.
- Use `.home-page { padding: 0 16px 24px; }` and a compact heading with no negative margins.
- Use a single-column `.course-list` with 12px gaps.
- Use a two-column course card (`96px 1fr`), a 1px neutral border, no large shadow, and at least a 44px interactive height.
- Make detail sections plain white surfaces separated by 12px spacing and neutral borders.
- Ensure the player and document surfaces honor `env(safe-area-inset-bottom)` without a fixed tab bar.
- Preserve visible `:focus-visible` outlines.

- [ ] **Step 8: Verify the H5 flow**

Run:

```bash
cd web/h5 && npm test && npm run build
```

Manual checks at `/h5/`:

- `/h5/` contains brand, logout, one heading, and course cards only.
- `/h5/courses` and `/h5/profile` redirect to `/h5/`.
- No bottom navigation, notification, weekly plan, statistics, achievements, certificates, or inactive service links remain.
- Failed detail enrichment hides the lesson count instead of showing `0`.
- Course detail shows no aggregate progress or fake completion state and each lesson link works.
- Video resume/report behavior and document opening remain unchanged.
- The layout works at 320px, 375px, and 560px widths, including bottom safe areas.

- [ ] **Step 9: Commit the H5 simplification**

```bash
git add web/h5/src/App.tsx web/h5/src/components web/h5/src/pages web/h5/src/styles.css
git commit -m "feat(h5): simplify learner course flow"
```

---

### Task 6: Cross-App Regression and Scope Verification

**Files:**
- Verify only; modify only files already listed if verification finds a task-scoped defect.

**Interfaces:**
- Consumes: the completed PC and H5 learner flows.
- Produces: evidence that both builds, tests, legacy redirects, and learning behaviors satisfy the approved design without backend changes.

- [ ] **Step 1: Run all learner frontend tests and builds**

Run:

```bash
cd web/pc && npm test && npm run build
cd web/h5 && npm test && npm run build
```

Expected: all four commands PASS.

- [ ] **Step 2: Verify no removed learner concepts remain**

Run:

```bash
rg -n "本周学习计划|获得证书|我的证书|最近学习|全部课程|正在学习|已完成课程|LEARNING FOR GROWTH|COURSE OUTLINE|CONTINUE" web/pc/src web/h5/src
```

Expected: no matches in active learner UI. A match inside a deliberately retained login message must be reviewed and justified rather than deleted blindly.

- [ ] **Step 3: Verify no aggregate progress is rendered outside players**

Run:

```bash
rg -n "course\.progress|progress-row|detail-progress|course-progress" web/pc/src web/h5/src
```

Expected: no matches in home, course-card, or course-detail code. Progress references remain only in lesson-player code/API types where required for real lesson progress.

- [ ] **Step 4: Verify backend and admin scope stayed untouched**

Run:

```bash
git diff --name-only 4a97aa2..HEAD
```

Expected: only `web/pc`, `web/h5`, and the already-approved docs files appear; no `internal/`, `cmd/`, database, or `web/admin/` source changes.

- [ ] **Step 5: Perform the final visual walkthrough**

Start both apps in separate terminals:

```bash
cd web/pc && npm run dev
cd web/h5 && npm run dev
```

Verify logged-in empty, populated, course-with-no-lessons, video, document, missing-resource, and logout scenarios. Capture one desktop and one 375px mobile screenshot for review, confirming that the layout matches the approved A direction.

- [ ] **Step 6: Commit any verification-only corrections**

If verification required task-scoped corrections, stage only those explicit files and commit:

```bash
git commit -m "fix: complete learner simplification verification"
```

If no correction was required, do not create an empty commit.
