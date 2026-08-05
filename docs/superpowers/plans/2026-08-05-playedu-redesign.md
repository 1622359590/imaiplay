# PlayEdu 风格界面重设计 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改变业务接口、权限和路由行为的前提下，将 Admin 和 PC 学员端改造成浅色 PlayEdu 风格界面。

**Architecture:** 沿用现有 React 组件、Ant Design、React Router 和 CSS token。Admin 只调整布局、页面结构和样式；PC 只调整布局、课程展示和样式；缺失的统计数据由前端安全展示为 `--` 或空状态，不新增后端接口。

**Tech Stack:** React 18、TypeScript、Vite、Ant Design 5、Ant Design Icons、原生 CSS。

## Global Constraints

- 不安装新的 npm 依赖。
- 不修改 API 调用逻辑、权限校验、路由守卫、租户主题和数据模型。
- 主色沿用 `#ff5156`、`#e84349`、`#fff1f0`。
- 桌面多列布局在小屏降级为单列或横向可用布局。
- Admin 阶段完成后必须通过 `cd web/admin && npm run build`。
- PC 阶段完成后必须通过 `cd web/pc && npm run build`。

---

### Task 1: Admin layout and brand shell

**Files:**
- Modify: `web/admin/src/layout/AdminLayout.tsx`
- Modify: `web/admin/src/styles.css`

**Interfaces:**
- Consumes existing `profile`, `tokenRole`, `visibleMenuItems`, `Outlet`, `logout`.
- Produces the same role-filtered menu keys and logout behavior with a non-collapsible PlayEdu shell.

- [ ] Replace the collapsible `Sider` behavior with a fixed 220–240px light sidebar; keep `superadminMenus` and `tenantAdminMenus` keys unchanged.
- [ ] Replace the text `I` mark with an existing icon-library play icon plus `iMaiPlay`; keep tenant/admin logo semantics intact.
- [ ] Keep the current top-right profile/dropdown, remove the collapse button and footer slogan, and preserve mobile usability with the existing breakpoint behavior.
- [ ] Add CSS for white sidebar, `#fff1f0` selected menu, `#ff5156` active text/icon, light top bar and responsive content padding.
- [ ] Run `cd web/admin && npm run build`; expected: TypeScript and Vite build pass.

### Task 2: Admin dashboard composition

**Files:**
- Modify: `web/admin/src/pages/Dashboard.tsx`
- Modify: `web/admin/src/styles.css`

**Interfaces:**
- Consumes existing `dashboardApi.get()` and `DashboardStats` fields.
- Produces PlayEdu-style platform and tenant dashboards without changing API requests.

- [ ] Keep the existing `loading` and unavailable-stat states.
- [ ] Render platform statistics as a five-slot responsive row, using real platform counts where available and `--` for unavailable department/category/admin/resource values.
- [ ] Render tenant statistics as course/user/learning cards and retain the existing plan usage and demo-data controls.
- [ ] Add a quick-actions block that navigates only to existing routes (`/users`, `/courses`, `/resources`) and does not create new APIs.
- [ ] Add CSS-only resource summary and learning summary blocks that show empty/unknown states when the API has no corresponding fields; do not add a chart dependency.
- [ ] Run `cd web/admin && npm run build`; expected: pass.

### Task 3: Admin login and visual verification

**Files:**
- Modify: `web/admin/src/pages/Login.tsx`
- Modify: `web/admin/src/styles.css`

**Interfaces:**
- Consumes existing login form, tenant-code fallback and error handling.
- Produces a light centered login card with the same submit behavior.

- [ ] Preserve all current form fields/conditional tenant-code behavior and only replace the visual shell.
- [ ] Use a light `#fafafa` page, centered white card, coral play brand mark, readable labels/placeholders and coral primary button.
- [ ] Preserve mobile width and keyboard/focus states.
- [ ] Run `cd web/admin && npm run build`; expected: pass.

### Task 4: PC learner shell and navigation

**Files:**
- Modify: `web/pc/src/components/AppLayout.tsx`
- Modify: `web/pc/src/styles.css`

**Interfaces:**
- Consumes existing portal/theme/auth contexts and `portalRoutePath`.
- Produces navigation to existing home/recent routes while preserving tenant logo, tenant welcome text and logout.

- [ ] Build a single-line 64px white header with brand, `首页`/`最近学习` links, and the existing user/logout area.
- [ ] Keep `theme.logo_url`, portal name, tenant route prefix and `handleLogout` unchanged semantically.
- [ ] Add active coral underline and responsive header collapse/overflow behavior without adding a dependency.
- [ ] Run `cd web/pc && npm run build`; expected: pass.

### Task 5: PC home summaries and filters

**Files:**
- Modify: `web/pc/src/pages/HomePage.tsx`
- Modify: `web/pc/src/styles.css`

**Interfaces:**
- Consumes `useCourseList(getCourses)` and existing `Course` fields.
- Produces client-side filter state and PlayEdu-style progress/course sections.

- [ ] Preserve the current course-loading and empty states.
- [ ] Add two summary cards using available course/progress data; show `--` instead of inventing learning-time data.
- [ ] Add filter state for `全部`, `必修课`, `选修课`, `已学完`, `未学完`; use the existing optional `category` field and treat missing category as all-compatible.
- [ ] Add a category select derived from the loaded course list; filtering must be client-side only.
- [ ] Keep course navigation paths unchanged.
- [ ] Run `cd web/pc && npm run build`; expected: pass.

### Task 6: PC course card and login

**Files:**
- Modify: `web/pc/src/components/CourseCard.tsx`
- Modify: `web/pc/src/components/CourseGrid.tsx`
- Modify: `web/pc/src/pages/LoginPage.tsx`
- Modify: `web/pc/src/styles.css`

**Interfaces:**
- Consumes existing `Course` data, portal routing and login/auth context.
- Produces accessible PlayEdu-style course cards and a light login page without changing APIs.

- [ ] Change course cards to a 3-column responsive grid with cover, title, category tag, progress track and `未学习`/percentage state.
- [ ] Keep the existing keyboard activation and `portalRoutePath` navigation.
- [ ] Use a neutral fallback cover with the existing icon when no cover exists.
- [ ] Convert the PC login shell to a light centered card while retaining portal resolution, credentials and error handling.
- [ ] Run `cd web/pc && npm run build`; expected: pass.

### Task 7: Final verification

**Files:**
- Test: `web/admin` and `web/pc` build outputs.
- Inspect: all files changed in Tasks 1–6.

- [ ] Run `cd web/admin && npm run build`.
- [ ] Run `cd web/pc && npm run build`.
- [ ] Run `go test ./...` to ensure UI-only changes did not break the repository.
- [ ] Check `git diff --check` only for changed implementation files; ignore pre-existing whitespace in `.codex/*`.
- [ ] Confirm no files under `web/h5`, `internal/`, API clients or migrations were changed.
- [ ] Commit implementation files with a focused message after all checks pass.

