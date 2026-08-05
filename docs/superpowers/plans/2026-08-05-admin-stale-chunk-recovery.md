# Admin Stale Chunk Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Recover admin tabs automatically when a deployment removes a lazy-loaded route chunk that the open tab still references.

**Architecture:** A small route-import wrapper owns stale-chunk detection, one-shot session state, and reload behavior. React Router owns the persistent-failure UI through a Chinese `errorElement`.

**Tech Stack:** React 18, React Router 7, TypeScript 5.7, Vite 6, Node 26 built-in test runner

## Global Constraints

- Reload automatically only for known dynamic-import or chunk-load failures.
- Reload at most once for a single unresolved failure.
- Clear recovery state after a successful import.
- Add no package dependency and preserve existing user changes.

---

### Task 1: One-shot stale chunk recovery

**Files:**
- Create: `web/admin/src/utils/lazyWithReload.ts`
- Test: `web/admin/src/utils/lazyWithReload.test.ts`
- Modify: `web/admin/tsconfig.app.json`

**Interfaces:**
- Produces: `loadWithReload<T>(importer, runtime): Promise<T>` and `lazyWithReload<T>(importer): LazyExoticComponent<T>`.

- [x] **Step 1: Write failing tests** for success cleanup, first chunk failure reload, second chunk failure rethrow, and unrelated failure rethrow.
- [x] **Step 2: Run `node --test src/utils/lazyWithReload.test.ts`** and confirm failure because the production module does not exist.
- [x] **Step 3: Implement the minimal recovery helper** with injected storage/reload runtime and a browser-facing React lazy wrapper; exclude Node-only test files from the browser TypeScript project.
- [x] **Step 4: Re-run the focused test** and confirm all cases pass.

### Task 2: Route integration and persistent error UI

**Files:**
- Create: `web/admin/src/components/RouteErrorPage.tsx`
- Modify: `web/admin/src/routes.tsx`
- Modify: `web/admin/src/styles.css`

**Interfaces:**
- Consumes: `lazyWithReload` from Task 1.
- Produces: all lazy routes use one-shot recovery; route failures render a Chinese recovery page.

- [x] **Step 1: Replace `React.lazy` route imports** with `lazyWithReload`.
- [x] **Step 2: Add `RouteErrorPage`** using `useRouteError`, a concise Chinese message, and `window.location.reload()` action.
- [x] **Step 3: Register the error element** at the router root and add scoped styling.
- [x] **Step 4: Run `npm run build`** and confirm TypeScript and Vite complete successfully.
- [x] **Step 5: Run the focused unit test and inspect the build output** to confirm route chunks remain lazy-loaded.
