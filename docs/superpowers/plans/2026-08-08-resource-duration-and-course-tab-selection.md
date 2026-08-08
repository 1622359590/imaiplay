# Resource Duration and Course Tab Selection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist uploaded video duration for reuse, present unknown durations honestly, and make course-detail tab selection obey tenant selection colors.

**Architecture:** Video metadata remains browser-derived before upload, is sent as multipart metadata, and is stored on the resource so later lesson selection can reuse it without downloading the video. Course lessons continue owning their saved duration; the admin editor copies resource duration into the lesson form. The PC tab fix wins the Ant Design cascade explicitly and removes the redundant ink bar.

**Tech Stack:** Go, Gin, GORM migrations, React, TypeScript, Ant Design, Node test runner, Vitest.

## Global Constraints

- Existing resources and lessons remain valid with `duration_seconds = 0`.
- Never download a large existing video merely to inspect metadata.
- Tenant `selected_background_color`, `selected_text_color`, and `selected_icon_color` remain independent.
- Unknown duration is displayed as “时长未设置”, not `00:00`.

---

### Task 1: Persist Resource Video Duration

**Files:**
- Modify: `internal/domain/resource.go`
- Modify: `internal/migration/migration.go`
- Modify: `internal/api/resource.go`
- Modify: `internal/service/resource.go`
- Modify: `internal/repository/resource_gorm.go`
- Test: `internal/api/resource_test.go`
- Test: `internal/service/resource_test.go`
- Test: `internal/migration/migration_test.go`

**Interfaces:**
- Consumes multipart `duration_seconds` for video uploads.
- Produces `domain.Resource.DurationSeconds int` as JSON `duration_seconds`.

- [ ] Add failing handler/service/migration tests proving duration is accepted, stored, returned, and migrated.
- [ ] Run the focused Go tests and confirm failures are caused by the missing resource field and upload contract.
- [ ] Add the field, migration v21, optional duration-aware upload interface, validation, and repository persistence.
- [ ] Run focused Go tests until green.

### Task 2: Carry Duration Through Admin Upload and Resource Reuse

**Files:**
- Modify: `web/admin/src/api/resource.ts`
- Modify: `web/admin/src/components/MediaUploader.tsx`
- Modify: `web/admin/src/pages/course-detail/LessonEditor.tsx`
- Test: `web/admin/tests/videoDuration.test.ts`
- Test: `web/admin/tests/courseDetailModel.test.ts`

**Interfaces:**
- Consumes browser-derived whole-second duration.
- Produces multipart `duration_seconds` and copies an existing resource duration into the lesson form.

- [ ] Add failing tests for upload metadata construction and choosing a resource with saved duration.
- [ ] Run Admin tests and confirm expected failures.
- [ ] Send duration with upload requests, expose it on `Resource`, and centralize the lesson-duration selection rule in the course detail model.
- [ ] Run Admin tests until green.

### Task 3: Present Unknown Duration and Correct Course Tab Selection

**Files:**
- Modify: `web/pc/src/pages/CourseDetailPage.tsx`
- Modify: `web/pc/src/styles/responsive.css`
- Test: `web/pc/src/utils/learnerCourses.test.ts`
- Test: `web/pc/tests/selectionTheme.test.ts`

**Interfaces:**
- Consumes lesson duration and tenant selection CSS variables.
- Produces an honest unknown-duration label and accessible selected tab colors.

- [ ] Add failing tests that zero duration renders as unknown and that the tenant text color wins the Ant Design active-tab cascade.
- [ ] Run focused PC tests and confirm expected failures.
- [ ] Render “时长未设置” for zero duration, strengthen active-tab selectors, and hide the duplicate ink bar.
- [ ] Run PC tests until green.

### Task 4: Verify and Publish

**Files:**
- Verify all files changed in Tasks 1–3.

- [ ] Run `go test ./internal/...`.
- [ ] Run Admin tests and production build.
- [ ] Run PC tests and production build.
- [ ] Run `git diff --check` and inspect the final diff.
- [ ] Commit, rebase onto the latest `origin/main`, rerun verification, and push the fast-forward result to GitHub and Gitee `main`.
