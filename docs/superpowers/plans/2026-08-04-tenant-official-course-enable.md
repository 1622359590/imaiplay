# 学院启用官方课程实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 让学院管理员从课程管理页发现、启用和停用平台已发布的官方课程，并正确回显本学院状态。

**Architecture:** 后端在现有官方课程列表上按当前学院关联 `tenant_official_courses`，只向学院管理员暴露已发布课程及 `enabled` 字段；超级管理员仍看到完整官方课程。前端复用同一 API，在课程管理页提供快捷弹窗，并保留独立的官方课程菜单页。

**Tech Stack:** Go、Gin、GORM、React 18、TypeScript、Ant Design、Node test

## Global Constraints

- 复用现有珊瑚红后台主题和 Ant Design 组件。
- 官方课程保持平台共享引用，不复制到学院课程表。
- 学院管理员不可编辑官方课程内容。
- 不提交或推送，直到用户明确说“提交”。

---

### Task 1: 后端返回学院启用状态

**Files:**
- Modify: `internal/domain/course.go`
- Modify: `internal/repository/course_gorm.go`
- Test: `internal/repository/course_gorm_test.go`

**Interfaces:**
- Consumes: 当前请求上下文中的 `tenant_id` 与角色。
- Produces: `domain.Course.Enabled bool`，学院官方课程列表只含已发布课程并回显当前学院启用状态。

- [x] **Step 1: 写失败的仓储测试**

创建草稿、已发布未启用、已发布已启用三门官方课程；断言学院列表只有两门已发布课程，并分别返回 `enabled=false/true`，超级管理员仍能看到三门。

- [x] **Step 2: 验证测试因缺少状态关联而失败**

Run: `go test ./internal/repository -run TestOfficialCourseListIncludesTenantActivation -count=1`

Expected: FAIL，学院列表包含草稿或已启用课程的 `Enabled` 仍为 false。

- [x] **Step 3: 实现最小查询变更**

在 `Course` 增加非持久化 JSON 字段 `Enabled`。`FindOfficial` 对学院管理员增加 `status = 1` 条件，并用当前 `tenant_id` 查询启用状态；超级管理员保持原查询。

- [x] **Step 4: 验证仓储测试通过**

Run: `go test ./internal/repository -run TestOfficialCourseListIncludesTenantActivation -count=1`

Expected: PASS

### Task 2: 管理端课程选择模型与菜单入口

**Files:**
- Create: `web/admin/src/utils/officialCourses.ts`
- Create: `web/admin/tests/officialCourses.test.ts`
- Modify: `web/admin/src/api/course.ts`
- Modify: `web/admin/src/layout/AdminLayout.tsx`

**Interfaces:**
- Consumes: `Course[]` 与切换目标课程 ID。
- Produces: `updateOfficialCourseEnabled(items, id, enabled)`，以及学院菜单中的 `/official-courses` 入口。

- [x] **Step 1: 写失败的前端状态测试**

断言更新指定课程启用状态时不改变其他课程，并返回新数组。

- [x] **Step 2: 验证测试因工具函数不存在而失败**

Run: `npm test -- --test-name-pattern="updates one official course"`

Expected: FAIL，模块或导出不存在。

- [x] **Step 3: 实现纯状态函数和类型字段**

为 `Course` 增加 `enabled?: boolean`，实现不可变更新函数，并在学院菜单加入“官方课程”。

- [x] **Step 4: 验证前端状态测试通过**

Run: `npm test -- --test-name-pattern="updates one official course"`

Expected: PASS

### Task 3: 课程管理页添加官方课程弹窗

**Files:**
- Create: `web/admin/src/components/OfficialCoursePicker.tsx`
- Modify: `web/admin/src/pages/Courses.tsx`
- Modify: `web/admin/src/pages/OfficialCourses.tsx`

**Interfaces:**
- Consumes: `officialCourseApi.list()` 与 `officialCourseApi.enable(id, enabled)`。
- Produces: “添加官方课程”按钮、可加载和切换的官方课程弹窗、独立页面一致的乐观状态与错误恢复。

- [x] **Step 1: 用现有 API 和纯状态函数实现选择器**

弹窗展示课程名称、简介和启用开关；加载时显示表格 loading，空列表显示“暂无已发布的官方课程”。

- [x] **Step 2: 在课程管理页接入快捷入口**

右上角按“添加官方课程”“新建课程”排列；关闭弹窗后保持课程管理列表不变。

- [x] **Step 3: 修复独立官方课程页的状态交互**

使用返回的 `enabled` 字段和同一不可变更新逻辑，切换成功保持状态，失败恢复并提示。

- [x] **Step 4: 运行管理端测试与构建**

Run: `npm test && npm run build`

Expected: 全部 PASS，TypeScript 无错误，Vite 构建成功。

### Task 4: 全量验证与视觉检查

**Files:**
- Modify: `design-qa.md`
- Create: `.qa-artifacts/official-course-picker/` 截图文件

**Interfaces:**
- Consumes: 已运行的管理端与测试数据。
- Produces: 通过的自动化测试、同视口截图和 `final result: passed` 的设计 QA 报告。

- [x] **Step 1: 运行 Go 全量测试**

Run: `go test ./...`

Expected: PASS

- [x] **Step 2: 在桌面和窄屏检查主要交互**

验证菜单入口、弹窗打开、开关状态、空状态、关闭弹窗以及无控制台错误。

- [x] **Step 3: 对照现有后台截图完成设计 QA**

确认按钮层级、珊瑚红色调、表格间距、弹窗宽度和响应式布局一致；更新 `design-qa.md` 为 `final result: passed`。
