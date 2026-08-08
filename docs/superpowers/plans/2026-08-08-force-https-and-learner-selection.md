# Forced HTTPS and Learner Selection Synchronization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Require every successfully bound tenant domain to redirect HTTP to HTTPS and make PC learner course-filter selection colors match the tenant theme.

**Architecture:** Extend the domain panel abstraction with one explicit HTTPS-redirect operation, implement it through BaoTa's official `HttpToHttps` site action, and place it after certificate readiness but before tenant persistence. Independently tighten the PC Tabs selection selectors so the existing CSS variables win over Ant Design's active-state styles.

**Tech Stack:** Go, BaoTa form API, GORM-backed service tests, React 18, Ant Design 5, TypeScript, CSS, Node test runner.

## Global Constraints

- Strong HTTPS redirect is a required binding condition; failure must roll back the newly created BaoTa site.
- Do not enable BaoTa global HTTPS and do not add a handwritten Nginx redirect.
- Preserve the existing five-step domain binding status model.
- Preserve the existing learner Tabs component and filtering behavior.
- Publish to Gitee `main` before GitHub `main` after verification.

---

### Task 1: Require BaoTa HTTP-to-HTTPS Redirect

**Files:**
- Modify: `internal/baota/client_test.go`
- Modify: `internal/baota/site.go`
- Modify: `internal/service/domain_bind.go`
- Modify: `internal/service/domain_bind_test.go`
- Modify: `internal/service/domain_bind_orchestrator.go`

**Interfaces:**
- Consumes: existing `Client.request(ctx, "/site", form, true, false)` and existing certificate readiness polling.
- Produces: `DomainPanel.EnableHTTPSRedirect(domain string) error` and `(*baota.Client).EnableHTTPSRedirect(domain string) error`.

- [ ] **Step 1: Write the failing BaoTa request test**

Add a test in `internal/baota/client_test.go` that invokes `client.EnableHTTPSRedirect("academy.example.com")` and asserts this exact signed form:

```go
url.Values{
    "action":   {"HttpToHttps"},
    "siteName": {"academy.example.com"},
}
```

The request must use `POST /site`, and the mock response is `{"status":true,"msg":"设置成功"}`.

- [ ] **Step 2: Write failing domain orchestration assertions**

Extend `domainPanelStub` with:

```go
func (stub *domainPanelStub) EnableHTTPSRedirect(string) error {
    return stub.record("https_redirect")
}
```

Change the ready-flow expected call sequence to:

```go
[]string{"site", "proxy", "snippet", "ssl", "info", "https_redirect"}
```

Add `TestDomainBindHTTPSRedirectFailureRollsBackWithoutPersistingDomain`, using `failAt: "https_redirect"`, and assert:

```go
[]string{"site", "proxy", "snippet", "ssl", "info", "https_redirect", "delete"}
```

The final state must be `DomainStateSetupFailed`, the stored tenant's `CustomDomain` must remain `nil`, and no `domain.bind` audit event may be recorded.

- [ ] **Step 3: Run focused tests and verify they fail**

Run:

```bash
go test ./internal/baota ./internal/service -run 'TestEnableHTTPSRedirect|TestDomainBindFlowTransitionsToReadyAndPersistsDomain|TestDomainBindHTTPSRedirectFailureRollsBackWithoutPersistingDomain' -count=1
```

Expected: compilation or assertion failure because the production interface and BaoTa method do not yet exist and the orchestrator does not call the redirect operation.

- [ ] **Step 4: Implement the BaoTa operation**

Add this method to `internal/baota/site.go`:

```go
// EnableHTTPSRedirect forces all HTTP requests for a site to HTTPS.
func (client *Client) EnableHTTPSRedirect(domain string) error {
    _, err := client.request(context.Background(), "/site", url.Values{
        "action":   {"HttpToHttps"},
        "siteName": {domain},
    }, true, false)
    return err
}
```

Add the same signature to `DomainPanel` in `internal/service/domain_bind.go`.

- [ ] **Step 5: Enforce redirect before persistence**

In `runBind`, immediately after `waitForSSL(domainName)` succeeds and before `currentStep = 5`, set the progress message and call:

```go
service.setStatus(actor.tenantID, domainName, DomainStateConfiguring, "正在开启强制 HTTPS", 4)
if err := service.panel.EnableHTTPSRedirect(domainName); err != nil {
    fail("开启强制 HTTPS 失败", err)
    return
}
```

Do not move tenant persistence or success auditing ahead of this call.

- [ ] **Step 6: Run focused tests and verify they pass**

Run:

```bash
go test ./internal/baota ./internal/service -run 'TestEnableHTTPSRedirect|TestDomainBindFlowTransitionsToReadyAndPersistsDomain|TestDomainBindHTTPSRedirectFailureRollsBackWithoutPersistingDomain' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit the backend change**

```bash
git add internal/baota/client_test.go internal/baota/site.go internal/service/domain_bind.go internal/service/domain_bind_test.go internal/service/domain_bind_orchestrator.go
git commit -m "feat: force https for tenant domains"
```

### Task 2: Synchronize Learner Filter Selection Colors

**Files:**
- Modify: `web/pc/tests/selectionTheme.test.ts`
- Modify: `web/pc/src/styles/course.css`

**Interfaces:**
- Consumes: `--tenant-selected-background`, `--tenant-selected-text`, and `--tenant-selected-icon` set by `applyLearnerPalette`.
- Produces: stable selected-state styling for `.learner-filter-tabs` that overrides Ant Design's active tab text color.

- [ ] **Step 1: Strengthen the failing stylesheet regression test**

In `web/pc/tests/selectionTheme.test.ts`, assert all three course-filter mappings:

```ts
assert.match(stylesheet, /\.learner-filter-tabs \.ant-tabs-tab\.ant-tabs-tab-active \.ant-tabs-tab-btn[^\{]*\{[^}]*color:\s*var\(--tenant-selected-text\)\s*!important/s)
assert.match(stylesheet, /\.learner-filter-tabs \.ant-tabs-ink-bar[^\{]*\{[^}]*background:\s*var\(--tenant-selected-icon\)/s)
```

Keep the existing assertion that the active tab container consumes `--tenant-selected-background` and `--tenant-selected-text`.

- [ ] **Step 2: Run the focused PC test and verify it fails**

Run from `web/pc`:

```bash
npm test -- --test-name-pattern='PC persistent navigation and tabs consume selection variables'
```

Expected: FAIL because the active button selector is not specific enough and the ink bar still uses `--learner-accent`.

- [ ] **Step 3: Implement the minimal CSS correction**

In `web/pc/src/styles/course.css`, replace the active button rule with:

```css
.learner-filter-tabs .ant-tabs-tab.ant-tabs-tab-active .ant-tabs-tab-btn {
  color: var(--tenant-selected-text) !important;
}
```

Keep the active container background rule, and change the course-filter ink bar to:

```css
.learner-filter-tabs .ant-tabs-ink-bar {
  height: 3px;
  border-radius: 999px;
  background: var(--tenant-selected-icon);
}
```

- [ ] **Step 4: Run PC tests and build**

Run from `web/pc`:

```bash
npm test
npm run build
```

Expected: all tests pass and Vite production build succeeds.

- [ ] **Step 5: Commit the learner change**

```bash
git add web/pc/tests/selectionTheme.test.ts web/pc/src/styles/course.css
git commit -m "fix: sync learner filter selection colors"
```

### Task 3: Full Verification and Delivery

**Files:**
- Verify only; no planned source changes.

**Interfaces:**
- Consumes: Tasks 1 and 2 commits.
- Produces: verified commits on Gitee and GitHub `main`, with GitHub deployment workflow status inspected.

- [ ] **Step 1: Run backend verification**

```bash
go test ./... -count=1
go vet ./...
```

Expected: PASS with no vet diagnostics.

- [ ] **Step 2: Run affected frontend workspace verification**

Run each command from its workspace:

```bash
cd web/shared && npm test && npm run build
cd ../pc && npm test && npm run build
cd ../admin && npm test && npm run build
```

Expected: all tests and production builds pass.

- [ ] **Step 3: Check repository integrity**

```bash
git diff --check
git status --short --branch
git log -5 --oneline
```

Expected: no unstaged changes, no whitespace errors, and the design, backend, and learner commits are visible.

- [ ] **Step 4: Integrate remote changes safely**

```bash
git fetch gitee main
git fetch origin main
git merge --no-edit gitee/main
git merge --no-edit origin/main
```

Expected: clean fast-forward or merge with no unresolved conflicts. If either remote has new code, rerun Steps 1–3 before pushing.

- [ ] **Step 5: Push in project order**

```bash
git push gitee HEAD:main
git push origin HEAD:main
```

Expected: Gitee succeeds first; GitHub succeeds second and starts the deployment workflow.

- [ ] **Step 6: Confirm GitHub deployment**

```bash
gh run list --repo 1622359590/imaiplay --limit 3
IMAIPLAY_RUN_ID=$(gh run list --repo 1622359590/imaiplay --limit 1 --json databaseId --jq '.[0].databaseId')
gh run watch --repo 1622359590/imaiplay "$IMAIPLAY_RUN_ID" --exit-status
```

Expected: the workflow triggered by the final GitHub push completes successfully.
