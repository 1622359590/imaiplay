# 未完成工作统一收口设计

日期：2026-08-11
状态：已确认

## 1. 目标

在不覆盖主工作区现有 `.codex` 改动的前提下，补齐员工套餐上限在公开注册路径中的检查，确认旧门户工作已经进入主分支，清理冗余工作树，并将最终结果依次推送到 Gitee 和 GitHub。

## 2. 已确认现状

- `codex/remove-official-course-employee-limit` 比 `main` 多两个提交：隐藏官方课程导航、按套餐限制员工数。
- 旧 `codex/tenant-portal-login` 工作树共有 97 项未提交内容；94 个文件与已合入主分支的 `174ae6f` 完全一致，其余 3 个文件已被主分支后续修复取代。
- `codex/tenant-brand-name` 工作树中的两个带 ` 2.ts` 后缀文件与正式文件及主分支内容完全一致。
- `codex/simplify-learner-experience` 没有领先 `main` 的提交。
- 主工作区 `.codex/current-task.md` 和 `.codex/roadmap.md` 属于用户现有改动，不纳入本次提交。
- 本地存在重复的 Gitee 远端引用 `main 2`，它指向当前 `main` 的旧祖先提交，但非法空格文件名会导致部分遍历引用的 Git 命令失败。

## 3. 员工上限检查

### 3.1 共享检查器

将租户套餐容量判断保留在用户服务域内，通过一个小型接口向认证服务暴露：

```go
type EmployeeCapacityChecker interface {
    EnsureEmployeeCapacity(ctx context.Context, tenantID string) error
}
```

`UserService` 实现该接口。判断顺序固定为：租户 → `PlanID` → Plan → `MaxUsers` → 当前租户用户总数。

- `MaxUsers <= 0`：不限制。
- 已用人数 `>= MaxUsers`：返回 `errorsx.Forbidden("员工数已达套餐上限，请升级套餐")`。
- 无套餐：保持当前兼容行为，不限制。
- superadmin 初始化：不经过租户容量检查。

### 3.2 公开注册

`AuthService.RegisterWithPhone()` 在解析出当前租户后、创建用户前调用同一个检查器。生产依赖装配中将同一个 `UserService` 实例注入 `AuthService`，避免复制套餐判断逻辑。

新租户注册创建的第一个 tenant_admin 不额外拦截：此时租户用户数为 0；任何正数上限都允许首个管理员，`0` 本身代表不限。

## 4. 测试策略

- 先新增公开注册在达到套餐上限时返回 40300 的失败测试。
- 保留现有 `MaxUsers == 0` 不限制测试。
- 验证 `UserService` 单个创建与批量导入继续使用同一检查逻辑。
- 运行 `go build ./cmd/server/`、`go test ./...`。
- 运行 Admin、PC、H5 工作区测试及构建，确认合并没有破坏门户和导航。

## 5. Git 与清理策略

1. 将公开注册补丁 amend 到 `feat: enforce tenant employee limit from plan`，保持原功能提交边界。
2. 旧工作树在删除前执行包含未跟踪文件的 stash，并使用明确说明命名；stash 作为恢复点保留。
3. 仅在确认分支没有领先 `main` 的提交后移除工作树和本地冗余分支。
4. 将重复的 Gitee 引用移动到用户废纸篓，而不是直接不可恢复地删除。
5. 主分支采用 fast-forward 合并，避免生成无意义合并提交。
6. 推送顺序固定为 Gitee → GitHub；任一推送失败则停止并报告，不假定另一端成功。

## 6. 非目标

- 不恢复或覆盖已经被主分支后续版本替代的旧门户代码。
- 不提交主工作区现有 `.codex` 改动。
- 不改变 Plan CRUD、租户初始套餐选择或 superadmin 初始化逻辑。
- 不做与本次收口无关的 UI 或架构重构。

## 7. 完成标准

- 公开注册、后台创建和批量导入均受租户员工上限控制。
- superadmin 初始化和 `MaxUsers == 0` 行为保持不变。
- 后端与三端前端验证通过。
- `main` 包含最终提交，Gitee 与 GitHub 的 `main` 指向同一提交。
- 冗余工作树已清理且有 stash 恢复点，主工作区用户改动仍然保留。
