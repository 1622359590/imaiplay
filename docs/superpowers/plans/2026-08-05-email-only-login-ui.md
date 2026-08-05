# Email-Only Login UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove all visible phone-login guidance from the admin, PC learner, and H5 learner login forms while retaining existing backend and password-recovery capabilities.

**Architecture:** Keep the existing `identifier` request field and authentication APIs unchanged. Limit the change to the three login form configurations: email-only labels/placeholders plus client-side email validation, then verify the actual rendered pages and existing application builds.

**Tech Stack:** React 18, Ant Design, Ant Design Mobile, TypeScript, Vite

## Global Constraints

- Backend phone-login code remains unchanged.
- Password recovery, registration, user profile, and SMS settings remain unchanged.
- Admin, PC, and H5 login pages must not display phone-login wording or a phone-number example.
- All three forms must reject non-email identifiers before submitting.

---

### Task 1: Make all login forms email-only

**Files:**
- Modify: `web/admin/src/pages/Login.tsx`
- Modify: `web/pc/src/pages/LoginPage.tsx`
- Modify: `web/h5/src/pages/LoginPage.tsx`

**Interfaces:**
- Consumes: existing `identifier: string` login payload used by every client.
- Produces: unchanged login API requests containing a validated email in `identifier`.

- [x] **Step 1: Record the failing rendered behavior**

Open the admin, PC, and H5 login pages and verify that each currently renders at least one of `手机号`, `13800138000`, or `手机号或邮箱`. This is the failing behavior the change removes.

- [x] **Step 2: Update the admin form**

Use the following field configuration in `web/admin/src/pages/Login.tsx`:

```tsx
<Form.Item
  label="邮箱"
  name="identifier"
  rules={[
    { required: true, message: '请输入邮箱' },
    { type: 'email', message: '请输入有效邮箱' },
  ]}
>
  <Input
    prefix={<MailOutlined />}
    placeholder="name@company.com"
    autoComplete="username"
  />
</Form.Item>
```

- [x] **Step 3: Update the PC learner form**

Use the same label, messages, email rule, and `name@company.com` placeholder in `web/pc/src/pages/LoginPage.tsx`, preserving its existing classes, size, icon, and autofocus behavior.

- [x] **Step 4: Update the H5 learner form**

Keep the existing icon label and input styling in `web/h5/src/pages/LoginPage.tsx`, but use these rules and placeholder:

```tsx
rules={[
  { required: true, message: '请输入邮箱' },
  { type: 'email', message: '请输入有效邮箱' },
]}
```

```tsx
<Input className="dark-input" placeholder="邮箱" clearable />
```

- [x] **Step 5: Verify the rendered behavior**

Open all three login pages again. Confirm each shows only email wording, contains no phone-number example, accepts `name@company.com`, and displays `请输入有效邮箱` for `13800138000` without sending the login request.

- [x] **Step 6: Run regression tests and production builds**

Run:

```bash
cd web/admin && npm test && npm run build
cd ../pc && npm test && npm run build
cd ../h5 && npm test && npm run build
```

Expected: every command exits with status `0`.

- [x] **Step 7: Commit**

```bash
git add web/admin/src/pages/Login.tsx web/pc/src/pages/LoginPage.tsx web/h5/src/pages/LoginPage.tsx
git commit -m "fix: show email-only login fields"
```
