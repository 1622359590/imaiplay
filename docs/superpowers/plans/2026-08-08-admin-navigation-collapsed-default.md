# Admin Navigation Collapsed Default Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every grouped admin sidebar section start collapsed after page load or refresh, without route-driven expansion or persisted open state.

**Architecture:** Keep Ant Design Menu's controlled `openKeys` state inside `AdminLayout`, but initialize it from a navigation helper that always returns a fresh empty array. Remove session storage reads/writes and route effects so only `onOpenChange` can change the state during the current component lifetime.

**Tech Stack:** React 18, TypeScript, Ant Design Menu, Node's built-in test runner, Vite.

## Global Constraints

- All grouped menus start collapsed on mount and refresh.
- Manual expansion remains multi-open and lasts only for the current component lifetime.
- Route changes do not open, close, or reset groups.
- Existing selected menu state, role filtering, compact sidebar, and mobile drawer behavior remain unchanged.
- Add no dependency and make no visual-style change.
- Preserve unrelated dirty-worktree files.

---

### Task 1: Replace Persisted Route-Driven State with Empty Local State

**Files:**
- Modify: `web/admin/src/config/adminNavigation.ts`
- Modify: `web/admin/src/layout/AdminLayout.tsx`
- Test: `web/admin/tests/adminNavigation.test.ts`

**Interfaces:**
- Produces: `initialOpenGroups(): string[]`, returning a new empty array for each caller.
- Removes: `requiredOpenGroups(path, role)`, because routes no longer drive group expansion.
- `AdminLayout` consumes `initialOpenGroups` as the `useState<string[]>` initializer and assigns `onOpenChange` keys directly.

- [ ] **Step 1: Write the failing navigation-state test**

Replace the route-expansion assertion with an observable initialization contract:

```ts
import { initialOpenGroups } from '../src/config/adminNavigation.ts'

test('admin navigation starts every page with all groups collapsed', () => {
  const first = initialOpenGroups()
  const second = initialOpenGroups()
  assert.deepEqual(first, [])
  assert.deepEqual(second, [])
  assert.notEqual(first, second)
})
```

The separate array assertion prevents future callers from sharing and mutating a module-level array.

- [ ] **Step 2: Run the focused test and verify RED**

Run: `npm --prefix web/admin test`

Expected: TypeScript execution fails because `initialOpenGroups` is not exported.

- [ ] **Step 3: Add the initializer and simplify `AdminLayout`**

In `adminNavigation.ts`, remove `requiredOpenGroups` and add:

```ts
export function initialOpenGroups(): string[] {
  return []
}
```

In `AdminLayout.tsx`:

```ts
import {
  initialOpenGroups,
  navigationForRole,
  roleLabel,
  type NavigationIcon,
} from '../config/adminNavigation'

const [openKeys, setOpenKeys] = useState<string[]>(initialOpenGroups)

const changeOpenKeys = (next: string[]) => {
  setOpenKeys(next)
}
```

Delete `requiredGroups`, `storageKey`, the session storage initializer, the route-change `useEffect`, and the session storage write. Keep the viewport resize `useEffect` and every other menu property unchanged.

- [ ] **Step 4: Run test and build verification**

Run:

```bash
npm --prefix web/admin test
npm --prefix web/admin run build
```

Expected: all 45 navigation/admin tests pass and the production build exits 0.

- [ ] **Step 5: Check scope and commit**

Run:

```bash
git diff --check
git status --short
git diff -- web/admin/src/config/adminNavigation.ts web/admin/src/layout/AdminLayout.tsx web/admin/tests/adminNavigation.test.ts
git add web/admin/src/config/adminNavigation.ts web/admin/src/layout/AdminLayout.tsx web/admin/tests/adminNavigation.test.ts
git commit -m "fix(admin): collapse navigation groups by default"
```

### Task 2: Final Verification

**Files:**
- Verify only; no planned code changes.

**Interfaces:**
- Verifies the admin behavior and ensures shared backend/other frontend work remains unaffected.

- [ ] **Step 1: Re-run the management app checks from a clean committed tree**

Run:

```bash
npm --prefix web/admin test
npm --prefix web/admin run build
```

Expected: tests and build exit 0. Existing Vite chunk-size warnings are allowed; TypeScript errors and test failures are not.

- [ ] **Step 2: Verify repository state**

Run:

```bash
git diff --check HEAD^..HEAD
git show --stat --oneline HEAD
git status --short
```

Expected: the commit contains only the three planned admin files; pre-existing unrelated working-tree changes remain unstaged and uncommitted.
