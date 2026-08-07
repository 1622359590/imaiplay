# User-Facing Chinese Errors Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ensure Admin, PC, and H5 show understandable Chinese errors without exposing Axios, HTTP, or internal English messages.

**Architecture:** Add a pure shared `userFacingErrorMessage` translator that preserves Chinese business messages, maps known backend errors, handles transport/status failures, and uses a caller-provided Chinese fallback. Route all UI error presentation through it while retaining status/code inspection for control flow.

**Tech Stack:** TypeScript, Axios-compatible error objects, React, Ant Design, Ant Design Mobile, Node test runner.

## Global Constraints

- Preserve explicit Chinese business messages returned by the backend.
- Never display unknown English, Axios default messages, stack traces, or internal details to end users.
- Keep raw errors available to programmatic control flow and developer logging.
- Cover Admin, PC, and H5.

---

### Task 1: Shared user-facing error translator

**Files:**
- Modify: `web/shared/src/api/errors.ts`
- Modify: `web/shared/tests/errors.test.ts`

**Interfaces:**
- Produces: `userFacingErrorMessage(error: unknown, fallback?: string): string`
- Preserves: `responseStatus(error)` and `responseMessage(error)`

- [ ] **Step 1: Write failing table-driven tests**

Add cases for Chinese passthrough; known English messages such as `invalid email or password`, `permission denied`, and `resource is in use`; HTTP 401/403/404/409/413/429/500; Axios `ERR_NETWORK`; timeout `ECONNABORTED`; plain `Network Error`; and an unknown English error returning the supplied fallback.

- [ ] **Step 2: Run the shared tests and verify failure**

Run: `npm test --prefix web/shared`

Expected: FAIL because `userFacingErrorMessage` is not exported.

- [ ] **Step 3: Implement the minimal pure translator**

Implement the precedence `Chinese response message -> known English mapping -> transport mapping -> status mapping -> Chinese Error.message -> fallback`, defaulting the fallback to `请求失败，请稍后重试`. Detect Chinese with `/[\u3400-\u9fff]/` and normalize known messages using trimmed lowercase text.

- [ ] **Step 4: Run shared tests and type checking**

Run: `npm test --prefix web/shared && npm run build --prefix web/shared`

Expected: all tests and TypeScript checks pass.

- [ ] **Step 5: Commit the shared translator**

```bash
git add web/shared/src/api/errors.ts web/shared/tests/errors.test.ts
git commit -m "feat: translate user-facing request errors"
```

### Task 2: Adopt the translator in Admin

**Files:**
- Modify: `web/admin/src/api/client.ts`
- Modify: `web/admin/src/pages/Login.tsx`
- Modify: `web/admin/src/components/CourseMaterialsManager.tsx`
- Modify: `web/admin/src/components/RouteErrorPage.tsx`
- Create: `web/admin/tests/routeErrorPage.test.ts`

**Interfaces:**
- Consumes: `userFacingErrorMessage(error, fallback)` from `@imaiplay/shared/api/errors`

- [ ] **Step 1: Add a failing source-level route error regression test**

Assert that `RouteErrorPage.tsx` does not render `<code>`, `error.message`, or the label `错误详情`, while still rendering the refresh action and readable recovery text.

- [ ] **Step 2: Run Admin tests and verify failure**

Run: `npm test --prefix web/admin`

Expected: FAIL because the route page still exposes raw error details.

- [ ] **Step 3: Replace raw Admin error presentation**

Use `userFacingErrorMessage` in the Axios response interceptor, login/tenant-selection catches, and course-material failures. Remove the raw detail block from the route error page. Keep the existing auth endpoint rejection behavior and status checks unchanged so special login routing continues to work.

- [ ] **Step 4: Run Admin tests and build**

Run: `npm test --prefix web/admin && npm run build --prefix web/admin`

Expected: all tests, type checks, and Vite build pass.

- [ ] **Step 5: Commit Admin adoption**

```bash
git add web/admin/src/api/client.ts web/admin/src/pages/Login.tsx web/admin/src/components/CourseMaterialsManager.tsx web/admin/src/components/RouteErrorPage.tsx web/admin/tests/routeErrorPage.test.ts
git commit -m "fix: show readable Chinese admin errors"
```

### Task 3: Adopt the translator in PC and H5

**Files:**
- Modify: `web/pc/src/api/client.ts`
- Modify: `web/pc/src/pages/LoginPage.tsx`
- Modify: `web/pc/src/pages/OrganizationSelectPage.tsx`
- Modify: `web/h5/src/api/client.ts`
- Modify: `web/h5/src/pages/LoginPage.tsx`
- Modify: `web/h5/src/pages/ForgotPasswordPage.tsx`

**Interfaces:**
- Consumes: `userFacingErrorMessage(error, fallback)` from `@imaiplay/shared/api/errors`

- [ ] **Step 1: Replace direct `error.message` and raw response presentation**

Apply operation-specific Chinese fallbacks: `登录失败，请稍后重试`, `企业选择失败，请重新登录`, and `请求失败，请稍后重试`. Preserve current redirect and session-clearing behavior.

- [ ] **Step 2: Scan for remaining unsafe presentation paths**

Run: `rg -n "message\.error\(error\.message|error instanceof Error \? error\.message|responseMessage\(error\)" web/admin/src web/pc/src web/h5/src`

Expected: no user-facing raw error presentation remains; helper internals and non-display control flow are allowed.

- [ ] **Step 3: Run PC and H5 tests and builds**

Run: `npm test --prefix web/pc && npm run build --prefix web/pc && npm test --prefix web/h5 && npm run build --prefix web/h5`

Expected: all tests, type checks, and builds pass.

- [ ] **Step 4: Commit learner client adoption**

```bash
git add web/pc/src/api/client.ts web/pc/src/pages/LoginPage.tsx web/pc/src/pages/OrganizationSelectPage.tsx web/h5/src/api/client.ts web/h5/src/pages/LoginPage.tsx web/h5/src/pages/ForgotPasswordPage.tsx
git commit -m "fix: localize learner-facing errors"
```

