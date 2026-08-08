# Domain Binding Navigation Guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在宝塔自动配置自定义域名期间显示不可关闭的进度弹窗，并阻止误触导航、刷新或关闭页面。

**Architecture:** 将“是否需要离开保护”和进度百分比提取为纯函数，供页面与 Node 单元测试共同使用。`DomainSettings` 根据服务端工作状态渲染 Ant Design 模态框，通过 React Router `useBlocker` 拦截 SPA 导航，并通过 `beforeunload` 保护刷新、关闭标签页及站外跳转。

**Tech Stack:** React 18、TypeScript 5.7、React Router 7、Ant Design 5、Node Test Runner、Vite 6

## Global Constraints

- 只在 `creating_site` 与 `configuring` 状态启用等待弹窗和离开保护。
- `pending_verification`、`ready`、`setup_failed` 不启用离开保护。
- 弹窗不可通过关闭按钮、遮罩或 Esc 关闭。
- 弹窗复用现有五步流程标题，并展示 `current_step / total_steps` 的百分比。
- 绑定结束后立即移除路由阻塞及 `beforeunload` 监听。
- 不修改服务端域名绑定流程，不增加取消按钮或任务中心。

---

### Task 1: 绑定保护状态模型

**Files:**
- Create: `web/admin/src/utils/domainBindingGuard.ts`
- Create: `web/admin/tests/domainBindingGuard.test.ts`

**Interfaces:**
- Consumes: `DomainBindState` 与 `DomainBindStatus`。
- Produces: `shouldGuardDomainBinding(state?: DomainBindState): boolean` 与 `domainBindingProgress(status?: Pick<DomainBindStatus, 'current_step' | 'total_steps'>): number`。

- [ ] **Step 1: Write the failing tests**

```ts
import assert from 'node:assert/strict'
import test from 'node:test'
import { domainBindingProgress, shouldGuardDomainBinding } from '../src/utils/domainBindingGuard.ts'

test('guards only long-running domain provisioning states', () => {
  assert.equal(shouldGuardDomainBinding('creating_site'), true)
  assert.equal(shouldGuardDomainBinding('configuring'), true)
  assert.equal(shouldGuardDomainBinding('pending_verification'), false)
  assert.equal(shouldGuardDomainBinding('ready'), false)
  assert.equal(shouldGuardDomainBinding('setup_failed'), false)
})

test('calculates a bounded whole-number binding progress percentage', () => {
  assert.equal(domainBindingProgress({ current_step: 3, total_steps: 5 }), 60)
  assert.equal(domainBindingProgress({ current_step: 8, total_steps: 5 }), 100)
})
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `npm test -- --test-name-pattern='domain provisioning states|binding progress'`

Expected: FAIL because `domainBindingGuard.ts` does not exist.

- [ ] **Step 3: Implement the pure helpers**

```ts
import type { DomainBindState, DomainBindStatus } from '../api/domain'

export function shouldGuardDomainBinding(state?: DomainBindState) {
  return state === 'creating_site' || state === 'configuring'
}

export function domainBindingProgress(
  status?: Pick<DomainBindStatus, 'current_step' | 'total_steps'>,
) {
  const total = Math.max(1, status?.total_steps || 5)
  const current = Math.max(0, Math.min(total, status?.current_step || 0))
  return Math.round((current / total) * 100)
}
```

- [ ] **Step 4: Run the focused test and verify GREEN**

Run: `npm test -- --test-name-pattern='domain provisioning states|binding progress'`

Expected: both tests PASS.

### Task 2: 等待弹窗与离开保护

**Files:**
- Modify: `web/admin/src/pages/DomainSettings.tsx`
- Test: `web/admin/tests/domainBindingGuard.test.ts`

**Interfaces:**
- Consumes: `shouldGuardDomainBinding`、`domainBindingProgress`、现有 `flowSteps` 和 `DomainBindStatus`。
- Produces: 工作状态下不可关闭的等待弹窗、SPA 路由拦截和浏览器离开确认。

- [ ] **Step 1: Import guard dependencies**

在 `DomainSettings.tsx` 中加入 Ant Design `Modal`、React Router `useBlocker`，以及两个 guard helper。

- [ ] **Step 2: Derive guarded state and install browser protection**

```ts
const bindingInProgress = shouldGuardDomainBinding(status?.state)
const bindingProgress = domainBindingProgress(status)
const navigationBlocker = useBlocker(bindingInProgress)

useEffect(() => {
  if (!bindingInProgress) return
  const preventUnload = (event: BeforeUnloadEvent) => {
    event.preventDefault()
    event.returnValue = ''
  }
  window.addEventListener('beforeunload', preventUnload)
  return () => window.removeEventListener('beforeunload', preventUnload)
}, [bindingInProgress])

useEffect(() => {
  if (navigationBlocker.state !== 'blocked') return
  message.warning('域名正在配置，请等待完成后再离开')
  navigationBlocker.reset()
}, [navigationBlocker])
```

- [ ] **Step 3: Render the non-dismissible progress modal**

在页面根节点内渲染：

```tsx
<Modal
  open={bindingInProgress}
  title="正在配置自定义域名"
  footer={null}
  closable={false}
  maskClosable={false}
  keyboard={false}
  centered
>
  <Space direction="vertical" size={16} style={{ width: '100%' }}>
    <Spin size="large" />
    <Typography.Paragraph style={{ marginBottom: 0 }}>
      系统正在创建站点、配置访问并申请 HTTPS 证书，请勿关闭、刷新或离开当前页面。
    </Typography.Paragraph>
    <Progress percent={bindingProgress} status="active" />
    <Typography.Text strong>
      当前步骤：{flowSteps[currentStep]?.title}
    </Typography.Text>
    <Typography.Text type="secondary">
      通常需要 1–3 分钟，完成后页面会自动更新。
    </Typography.Text>
  </Space>
</Modal>
```

- [ ] **Step 4: Run frontend tests and production build**

Run: `npm test && npm run build`

Expected: all Node tests pass; TypeScript and Vite build complete successfully.

### Task 3: Repository Verification and Delivery

**Files:**
- Verify all changed files from Tasks 1–2.

**Interfaces:**
- Consumes: completed frontend implementation.
- Produces: verified commit ready for Gitee-first and GitHub-second delivery.

- [ ] **Step 1: Run full repository verification**

Run from repository root: `go test ./... -count=1 && go vet ./... && git diff --check`

Expected: all commands exit 0.

- [ ] **Step 2: Review the final diff**

Run: `git status --short && git diff --stat && git diff`

Expected: only the plan, guard helper, guard tests, and `DomainSettings.tsx` are changed after the prior design commit.

- [ ] **Step 3: Commit implementation**

```bash
git add web/admin/src/utils/domainBindingGuard.ts web/admin/tests/domainBindingGuard.test.ts web/admin/src/pages/DomainSettings.tsx docs/superpowers/plans/2026-08-08-domain-binding-navigation-guard.md
git commit -m "feat: guard domain binding navigation"
```

- [ ] **Step 4: Push in project order**

先推送 Gitee `main`，确认成功后再推送 GitHub `main`，并检查 GitHub 部署流程最终结论。
