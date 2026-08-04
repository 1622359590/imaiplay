# Admin Coral UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert the admin login and authenticated admin shell from dark blue/purple styling to the approved light coral design with a white sidebar.

**Architecture:** Add a dedicated fixed admin palette that is independent of tenant portal theme data. Feed it into Ant Design tokens, switch the menu to light mode, and replace old blue/purple CSS values across the admin shell, login, tables, cards, uploads, and course placeholders without altering business components.

**Tech Stack:** React 18, TypeScript, Redux Toolkit, Ant Design 5, React Router 7, Vite, Node test runner

## Global Constraints

- Admin shell always uses `#ff5156` primary, `#e84349` hover, `#fff1f0` selection, and the approved neutral palette.
- Tenant theme APIs, forms, permissions, data models, and portal metadata remain unchanged.
- Sidebar is white with a subtle right divider; selected menu is soft coral with coral text.
- Semantic success, warning, and delete colors remain semantic.
- Preserve all admin routes, role-based menus, login, tenant selection, tables, forms, uploads, and responsive behavior.

---

### Task 1: Create and test the fixed admin palette

**Files:**
- Create: `web/admin/src/theme/adminPalette.ts`
- Create: `web/admin/tests/adminPalette.test.ts`
- Modify: `web/admin/package.json`

**Interfaces:**
- Produces: `ADMIN_PALETTE` readonly object and `applyAdminPalette(element?: HTMLElement): void`.
- Consumed by: `AdminThemeProvider.tsx` and `styles.css` through CSS custom properties.

- [ ] **Step 1: Write the failing palette test**

Create a Node test that imports `ADMIN_PALETTE`, asserts the exact values, and uses the same luminance/contrast helpers as learner palette tests:

```ts
assert.deepEqual(ADMIN_PALETTE, {
  accent: '#ff5156', accentHover: '#e84349', accentSoft: '#fff1f0',
  heading: '#262626', text: '#595959', muted: '#737373',
  page: '#fafafa', card: '#ffffff', line: '#eeeeee',
})
assert.ok(contrast(ADMIN_PALETTE.heading, ADMIN_PALETTE.card) >= 12)
assert.ok(contrast(ADMIN_PALETTE.text, ADMIN_PALETTE.card) >= 7)
assert.ok(contrast(ADMIN_PALETTE.muted, ADMIN_PALETTE.card) >= 4.5)
```

- [ ] **Step 2: Add the admin test command and verify failure**

Set `package.json` script to:

```json
"test": "node --test tests/*.test.ts"
```

Run: `cd web/admin && npm test`  
Expected: FAIL because `src/theme/adminPalette.ts` does not exist.

- [ ] **Step 3: Implement the palette module**

```ts
export const ADMIN_PALETTE = {
  accent: '#ff5156', accentHover: '#e84349', accentSoft: '#fff1f0',
  heading: '#262626', text: '#595959', muted: '#737373',
  page: '#fafafa', card: '#ffffff', line: '#eeeeee',
} as const

export function applyAdminPalette(element: HTMLElement = document.documentElement) {
  for (const [name, value] of Object.entries(ADMIN_PALETTE)) {
    element.style.setProperty(`--admin-${name.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)}`, value)
  }
}
```

- [ ] **Step 4: Run tests**

Run: `cd web/admin && npm test`  
Expected: all existing and new admin tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/admin/package.json web/admin/tests/adminPalette.test.ts web/admin/src/theme/adminPalette.ts
git commit -m "test(admin): define accessible coral palette"
```

### Task 2: Decouple Ant Design admin shell colors from tenant theme

**Files:**
- Modify: `web/admin/src/components/AdminThemeProvider.tsx`
- Modify: `web/admin/src/layout/AdminLayout.tsx`

**Interfaces:**
- Consumes: `ADMIN_PALETTE` and `applyAdminPalette()`.
- Produces: fixed light Ant Design theme and a light-mode menu; tenant theme loading is no longer required to color the shell.

- [ ] **Step 1: Replace the provider implementation**

Remove `themeApi`, `TenantTheme`, `tokenRole`, state, and theme-change listeners. Apply the fixed palette once and configure:

```tsx
const themeConfig = {
  token: {
    colorPrimary: ADMIN_PALETTE.accent,
    colorInfo: ADMIN_PALETTE.accent,
    colorText: ADMIN_PALETTE.text,
    colorTextHeading: ADMIN_PALETTE.heading,
    colorBgLayout: ADMIN_PALETTE.page,
    colorBgContainer: ADMIN_PALETTE.card,
    colorBorderSecondary: ADMIN_PALETTE.line,
    borderRadius: 10,
  },
  components: {
    Table: { headerBg: '#fafafa' },
    Layout: { headerBg: '#ffffff', siderBg: '#ffffff', bodyBg: '#fafafa' },
    Menu: {
      itemBg: '#ffffff', itemColor: '#595959', itemHoverColor: '#ff5156',
      itemHoverBg: '#fff7f6', itemSelectedColor: '#ff5156', itemSelectedBg: '#fff1f0',
    },
  },
}
```

- [ ] **Step 2: Switch the sidebar menu to light mode**

Change only this prop in `AdminLayout.tsx`:

```tsx
<Menu theme="light" mode="inline" ... />
```

Keep role filtering, selected key logic, collapse behavior, navigation, and logout unchanged.

- [ ] **Step 3: Run tests and build**

Run: `cd web/admin && npm test && npm run build`  
Expected: zero test, TypeScript, and Vite failures.

- [ ] **Step 4: Commit**

```bash
git add web/admin/src/components/AdminThemeProvider.tsx web/admin/src/layout/AdminLayout.tsx
git commit -m "refactor(admin): use fixed light coral theme"
```

### Task 3: Restyle the authenticated admin shell and content surfaces

**Files:**
- Modify: `web/admin/src/styles.css`
- Verify: `web/admin/src/pages/Dashboard.tsx`
- Verify: `web/admin/src/pages/Courses.tsx`
- Verify: `web/admin/src/pages/Resources.tsx`
- Verify: `web/admin/src/pages/ThemeSettings.tsx`

**Interfaces:**
- Consumes: `--admin-*` custom properties applied by the provider and existing layout/page class names.
- Produces: white sidebar, soft-coral active menu, neutral cards/tables, and coral primary interactions.

- [ ] **Step 1: Capture the baseline**

Capture dashboard, courses, resources, and theme settings at 1440×900 and the collapsed shell at 1024×768 under `.qa-artifacts/admin-before/`.

- [ ] **Step 2: Replace root and shell rules**

```css
:root {
  --brand-600: var(--admin-accent, #ff5156);
  --gray-50: var(--admin-page, #fafafa);
  --gray-900: var(--admin-heading, #262626);
  --shadow-sm: 0 4px 14px rgb(38 38 38 / 4%);
}
.app-sider { border-right: 1px solid var(--admin-line); background: var(--admin-card); box-shadow: none; }
.brand { color: var(--admin-heading); }
.brand-mark { color: #fff; background: var(--admin-accent); }
.sider-version { color: var(--admin-muted); }
.top-header { border-bottom-color: var(--admin-line); background: var(--admin-card); }
.user-avatar { color: var(--admin-accent); background: var(--admin-accent-soft); }
.app-content { background: var(--admin-page); }
.ant-card { border-color: var(--admin-line); box-shadow: var(--shadow-sm); }
```

Override group labels, menu icons, hover, and selected states so no dark-menu text or purple background survives.

- [ ] **Step 3: Replace remaining blue/purple presentation colors**

Use `rg -n '#(1769e0|4F46E5|312E81|4338CA|3b82f6|8b5cf6|60a5fa|082f72|0f5ac7|2384ef)' web/admin/src` and replace presentation-only occurrences in `styles.css` and theme defaults with admin variables. Keep `#000` for the media player and semantic status colors.

Specifically update `.stat-icon`, `.course-cover`, `.detail-cover`, `.lesson-index`, `.media-uploader`, `.media-uploader-thumb`, progress controls, links, pagination, and focus rings to coral or soft coral.

- [ ] **Step 4: Build and inspect responsive shell behavior**

Run: `cd web/admin && npm run build`  
Expected: build passes. Check 1440×900, 1024×768, and 390×844; content is not covered by the fixed collapsed sidebar, tables remain reachable, and active navigation is clearly visible.

- [ ] **Step 5: Commit**

```bash
git add web/admin/src/styles.css
git commit -m "style(admin): apply light coral workspace"
```

### Task 4: Convert admin login and organization selection to light coral

**Files:**
- Modify: `web/admin/src/styles.css`
- Verify: `web/admin/src/pages/Login.tsx`

**Interfaces:**
- Consumes: existing `.admin-login-page`, `.admin-login-brand`, `.admin-login-card`, `.organization-picker`, and `.organization-card` classes.
- Produces: a centered light admin login and organization picker with unchanged authentication behavior.

- [ ] **Step 1: Replace dark login styles**

```css
.login-page.admin-login-page { min-height: 100vh; padding: 32px 20px; background: var(--admin-page); }
.admin-login-brand { margin-bottom: 24px; color: var(--admin-heading); }
.login-logo { color: #fff; background: var(--admin-accent); }
.admin-login-card { border: 1px solid var(--admin-line); background: var(--admin-card); box-shadow: 0 10px 30px rgb(38 38 38 / 6%); backdrop-filter: none; }
.admin-login-card .admin-login-title { color: var(--admin-heading); background: none; -webkit-text-fill-color: currentColor; }
.admin-login-card .ant-typography-secondary,
.admin-login-card .ant-form-item-label > label { color: var(--admin-muted); }
.admin-login-card .ant-input-affix-wrapper { color: var(--admin-heading); border-color: #d9d9d9; background: #fff; }
.admin-login-card .ant-input-affix-wrapper-focused { border-color: var(--admin-accent); box-shadow: 0 0 0 2px rgb(255 81 86 / 12%); }
.organization-card.ant-btn { color: var(--admin-text); border-color: var(--admin-line); background: #fff; }
.organization-card.ant-btn:hover { color: var(--admin-accent); border-color: var(--admin-accent); background: var(--admin-accent-soft); }
```

- [ ] **Step 2: Verify all login states**

Run: `cd web/admin && npm test && npm run build`  
Expected: zero failures. Verify default, focus, validation error, loading, organization selection, return-to-login, forgotten-password, and registration links at 1440×900 and 390×844.

- [ ] **Step 3: Compare with the reference**

Save `.qa-artifacts/admin-after/login.png`, compare beside the supplied coral learner reference for tone, and correct contrast, spacing, border, radius, and alignment mismatches.

- [ ] **Step 4: Commit**

```bash
git add web/admin/src/styles.css
git commit -m "style(admin): simplify management login"
```

### Task 5: Admin and cross-surface regression gate

**Files:**
- Modify: `design-qa.md`

**Interfaces:**
- Consumes: completed learner PC, learner H5, and admin work.
- Produces: final documented QA evidence before merge and push.

- [ ] **Step 1: Run every frontend gate**

```bash
cd web/pc && npm test && npm run build
cd ../h5 && npm test && npm run build
cd ../admin && npm test && npm run build
```

Expected: all tests, TypeScript checks, and Vite builds pass.

- [ ] **Step 2: Scan for obsolete presentation colors**

Run:

```bash
rg -n '#(4F46E5|312E81|4338CA|3b82f6|8b5cf6|60a5fa)' web/pc/src web/h5/src web/admin/src
```

Expected: no effective shell/login/card presentation rules use these colors; API fallback data may remain only where documented and visually inactive.

- [ ] **Step 3: Perform browser comparison QA**

In the selected browser, capture matching before/after viewports for learner detail, learner login, admin login, and admin courses. Place each implementation screenshot beside the supplied reference and inspect layout, crop, spacing, type weight, contrast, border, and radius. Fix visible mismatches and rerun the affected build.

- [ ] **Step 4: Complete the QA report**

Add an `Admin and unified regression` section to `design-qa.md` containing commands, results, routes, viewport sizes, screenshot paths, remaining intentional semantic colors, and confirmation that core navigation and forms work.

- [ ] **Step 5: Commit**

```bash
git add design-qa.md web/pc web/h5 web/admin
git commit -m "test: verify unified coral interface"
```

