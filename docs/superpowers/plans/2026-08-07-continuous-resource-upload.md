# Continuous Resource Upload Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep the resource-list upload box visible after every successful upload while preserving single-file preview behavior elsewhere.

**Architecture:** Treat the Resources page uploader as a stateless create action: never bind the uploaded resource back into `MediaUploader.value`, and refresh the list after success. Leave the shared uploader and its Logo/course/lesson usages unchanged.

**Tech Stack:** TypeScript, React, Ant Design, Node test runner.

## Global Constraints

- The resource upload box must remain visible after success.
- The uploaded resource must appear in the refreshed resource list.
- Theme Logo, course cover, and lesson uploaders must retain preview, replace, and remove behavior.
- Do not change resource storage or API contracts.

---

### Task 1: Make the Resources uploader continuous

**Files:**
- Modify: `web/admin/src/pages/Resources.tsx`
- Create: `web/admin/tests/continuousResourceUpload.test.ts`

**Interfaces:**
- Consumes: existing `MediaUploader` `onChange(resource)` callback
- Produces: a Resources page that never passes a completed resource back as `value`

- [ ] **Step 1: Write a failing regression test**

Read `Resources.tsx` and assert it has no `useState<UploadedMedia>`, no `value={uploaded}`, and still calls both `message.success('资源上传成功')` and `void load()` when `resource` is defined. Also assert `MediaUploader.tsx` retains its `value`-based preview branch so other screens remain unaffected.

- [ ] **Step 2: Run Admin tests and verify failure**

Run: `npm test --prefix web/admin`

Expected: FAIL because Resources currently stores and binds `uploaded`.

- [ ] **Step 3: Remove the single-value state from Resources**

Remove the `UploadedMedia` import and `uploaded` state, omit the `value` prop, do not call `setUploaded`, and remove the deletion-time uploaded-state cleanup. Keep success notification and list reload inside `if (resource)`.

- [ ] **Step 4: Run Admin tests and build**

Run: `npm test --prefix web/admin && npm run build --prefix web/admin`

Expected: all tests and the production build pass.

- [ ] **Step 5: Commit continuous upload behavior**

```bash
git add web/admin/src/pages/Resources.tsx web/admin/tests/continuousResourceUpload.test.ts
git commit -m "fix: keep resource upload box available"
```
