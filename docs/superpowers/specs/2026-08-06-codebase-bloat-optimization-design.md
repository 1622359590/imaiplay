# 代码库瘦身与结构优化设计

日期：2026-08-06

## 背景

当前仓库的运行时代码规模仍然可控，但维护成本已经集中暴露在几个位置：三套前端重复实现认证、请求和主题基础设施；PC 样式文件通过多轮末尾覆盖维持最终视觉；后台课程详情、认证服务、域名绑定服务和路由注册承担过多职责；过程截图、历史实施计划和本地构建产物增加仓库体积与搜索噪音。

基线数据如下：

- Git 跟踪文件 482 个，约 5.28 MB。
- Go 生产代码约 15,810 行，Go 测试约 17,041 行。
- 前端生产代码约 14,163 行，前端测试约 3,169 行。
- Markdown 文档约 11,482 行。
- `.qa-artifacts` 约 2.69 MB，占当前跟踪内容约一半。
- PC 主样式文件约 2,010 行。
- 根目录存在约 36 MB、未跟踪且未忽略的 `server` 本地二进制。

## 目标

1. 降低重复维护、样式级联和大文件带来的理解成本。
2. 让前端公共基础设施只有一个可信实现，同时保留三端不同的交互和 UI。
3. 将后端路由、认证和域名绑定拆为职责清晰、可独立测试的组件。
4. 清理不会影响产品运行的过程产物，减少仓库体积和自动化上下文噪音。
5. 在整个重构过程中保持 API、数据库、视觉和用户操作流程不变。

## 非目标

- 不增加或删除产品功能。
- 不修改公开 API 路径、请求参数、响应格式或错误码。
- 不修改数据库表、迁移版本或已有数据。
- 不重新设计页面，不改变视觉基准或交互顺序。
- 不合并 PC 与 H5 的页面组件。
- 不重写 Git 历史，也不在本轮执行强制性的仓库压缩。

## 实施原则

采用渐进式全量重构。每个阶段都必须处于可构建、可测试状态；前一阶段验证通过后才进入下一阶段。优先建立行为特征测试，再移动或抽取实现。结构调整通过兼容外壳完成，调用方不需要一次性迁移。

## 一、仓库瘦身

### 本地构建产物

- 在根 `.gitignore` 中加入根目录 `/server`，不影响同名源码目录。
- 删除当前工作区中可重新构建的 `server` 二进制。
- 保留 Dockerfile 的 `-trimpath -ldflags="-s -w"` 发布构建方式。

### QA 产物

- 更新 `design-qa.md`，使其只引用最终基准截图，再保留这些有效引用。
- 保留文件名明确包含 `final` 或代表最终桌面、移动端状态的基准图。
- 删除 `before`、`comparison`、`source-normalized` 等过程截图，并同步删除对应的过程说明和失效引用。
- 被删除文件仍可从 Git 历史恢复，不执行历史重写。

### 设计与计划文档

- 保留 `docs/superpowers/specs` 中的设计规格。
- 删除已经完成、且需求与决策已沉淀到规格或当前代码中的历史实施计划。
- 保留本轮优化计划和仍然有效的部署、运维文档。
- `.codex` 当前状态文件不纳入自动删除范围。

## 二、前端公共层

建立以 `web` 为根的 npm workspace，包含 `admin`、`pc`、`h5` 和 `shared` 四个包，并使用一份 `web/package-lock.json`。各应用继续在自己的 `package.json` 声明专属依赖，workspace 只负责统一安装、锁定兼容版本和解析公共包。

新增 `web/shared` 包，共享没有 UI 倾向的代码：

```text
web/shared/
  api/
    clientCore.ts
    errors.ts
  auth/
    sessionCore.ts
    refreshCoordinator.ts
  theme/
    tenantTheme.ts
  types/
    auth.ts
    course.ts
    theme.ts
  test/
    sessionFixtures.ts
```

公共层不依赖 Ant Design、Ant Design Mobile、React Router 或具体 Store。各应用通过适配器提供：

- Token 的读取、写入与清除方式。
- 登录页面和无权限页面的跳转策略。
- 租户 Header 与平台 Host 规则。
- 当前应用使用的提示和日志回调。

PC、H5 和后台继续保留各自的 API 文件作为稳定入口，入口内部改为调用 `@imaiplay/shared`。这样可以逐步迁移，不要求一次性修改所有页面 import。

三个前端包的 React、Axios、Vite、TypeScript 和测试工具版本在兼容范围内对齐。Ant Design 与 Ant Design Mobile 保持独立。Docker 构建改为在 `web` 根目录执行一次 `npm ci`，再分别运行三个应用的构建脚本；部署输出路径保持不变。

## 三、前端热点拆分

### PC 样式

将 `web/pc/src/styles.css` 改为有序入口：

```text
web/pc/src/styles/
  base.css
  layout.css
  login.css
  dashboard.css
  course.css
  player.css
  responsive.css
```

迁移时先保持级联顺序，再逐组消除被最终主题完全覆盖的旧规则。相同选择器应在所属功能文件中只有一个主定义；响应式调整只放在 `responsive.css`。最终入口只负责导入，禁止继续在文件末尾追加“最终覆盖层”。

### 后台课程详情

将课程详情拆为页面控制器、数据 Hook 和领域组件：

```text
web/admin/src/pages/course-detail/
  CourseDetailPage.tsx
  useCourseDetail.ts
  CourseInfoSection.tsx
  ChapterEditor.tsx
  LessonEditor.tsx
  MaterialsSection.tsx
```

页面控制器只负责路由参数、页面级加载状态和组件组合。保存、删除、排序和上传行为保持原有接口及提示。目标是控制器约 200 行以内，单个子组件原则上不超过约 250 行。

## 四、后端组合层

`internal/server/server.go` 保留 Engine 创建、全局中间件和健康检查。业务路由按领域拆分：

```text
internal/server/
  server.go
  routes_auth.go
  routes_admin.go
  routes_courses.go
  routes_learner.go
  routes_infrastructure.go
```

拆分前生成或测试完整路由清单。拆分后必须保持现有 114 条路由、HTTP 方法、路径、中间件顺序和可选服务条件一致。

`cmd/server/main.go` 按模块拆分依赖构造函数，但继续作为唯一组合根。数据库连接、配置读取和服务生命周期不改变。

## 五、认证服务拆分

现有 `AuthService` 类型和公开方法保留，内部按职责拆分：

```text
internal/service/
  auth.go
  auth_credentials.go
  auth_login.go
  auth_recovery.go
  auth_tokens.go
```

- `auth_credentials.go`：注册、密码验证、用户查找。
- `auth_login.go`：平台登录、租户登录、租户选择和登录结果组装。
- `auth_recovery.go`：短信验证码、忘记密码和重置密码。
- `auth_tokens.go`：访问令牌、刷新令牌轮换和退出。
- `auth.go`：兼容外壳、依赖装配和公共数据类型。

拆分不改变限流位置、密码验证时机、租户信息暴露规则、刷新令牌轮换、错误码和审计语义。

## 六、域名绑定服务拆分

现有 `DomainBindService` 类型和公开方法保留，内部拆分为：

```text
internal/service/
  domain_bind.go
  domain_bind_verifier.go
  domain_bind_orchestrator.go
  domain_bind_status.go
```

- verifier：域名格式、保留域名、DNS 与 IP 验证。
- orchestrator：宝塔站点、证书申请、绑定与解绑流程。
- status：状态存储、并发任务占用、失败状态与审计。
- domain_bind.go：兼容外壳和对外入口。

现有锁粒度、异步任务行为、SSL 等待策略、租户所有权校验和审计记录必须保持。任何并发模型调整都需要独立测试证明等价，本轮默认不调整。

## 错误处理与兼容性

- 后端继续使用现有 `errorsx` 错误码和本地化消息。
- 前端公共层只分类错误，不决定具体 UI 文案。
- 401、403、租户停用、组织选择和登录失败的现有跳转行为保持不变。
- 公共核心不能吞掉 HTTP 非成功状态、刷新失败或页面卸载上报失败。
- 兼容外壳使现有 Handler 和测试可以分阶段迁移。

## 验证策略

### 特征测试

重构前补充或固定以下行为：

- 完整后端路由清单及关键中间件要求。
- 密码登录、短信登录、租户选择、刷新轮换、退出和找回密码。
- 域名验证、重复绑定、并发绑定、SSL 失败、解绑和审计。
- 三端 401/403、刷新并发、Token 清除和租户主题解析。
- PC 样式导入顺序、禁止历史覆盖层重新出现。
- 后台课程详情保存、章节课时管理和附件管理的数据映射。

### 阶段验证

每个阶段运行受影响模块测试。最终统一运行：

- `go test ./... -count=1`
- 后台测试与生产构建
- PC 测试与生产构建
- H5 测试与生产构建
- `git diff --check`
- 路由清单对比

视觉验证使用保留的最终基准图，检查桌面和移动端关键页面。重构不得以测试通过替代必要的视觉检查。

## 完成标准

- 外部 API、数据库、页面视觉和操作流程无变化。
- PC 样式入口只包含有序导入，历史末尾覆盖层被移除。
- 三端不再各自维护完整的 Token 刷新和标准错误处理实现。
- 后台课程详情页面控制器约 200 行以内。
- Auth 与 DomainBind 的职责文件原则上不超过约 350 行。
- 后端路由按领域可独立阅读，现有路由全部保留。
- 过程截图、过期计划和根目录本地二进制不再占用主工作区。
- 全量测试、生产构建、差异检查和最终复审全部通过。

## 风险与缓解

- CSS 删除可能造成隐蔽视觉回归：先移动保持级联，再按选择器逐组去重，并进行桌面与移动端基准检查。
- 共享会话代码可能混淆三端差异：通过无 UI 核心和应用适配器隔离。
- Auth 拆分可能改变安全边界：保留公共入口并先固定安全特征测试。
- DomainBind 拆分可能改变并发时序：本轮不改变并发模型，只移动职责。
- 大范围重构难以定位回归：按阶段形成小提交，每个提交保持绿色。
