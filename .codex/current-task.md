# 当前任务：SaaS 自助开通流程

> 本文件由 Claude 生成，Codex 请直接阅读并执行。
> 执行前请先阅读 `DESIGN.md` 和 `.codex/codex-log.md` 了解完整设计和项目状态。
> 当前分支：`codex/first-core-features`。

## 目标

完成 ImaiPlay 任务 14：SaaS 自助开通流程，让客户无需 superadmin 介入即可注册租户并开始使用。

## 已确认设计决策

- **开通方式**：完全自助开通，客户访问注册页填写信息即可
- **租户代码**：基于组织名称拼音/英文生成（如 `acme`），创建后不可修改
- **演示数据**：自动创建示例课程、学员、讲师、资源
- **演示数据清除**：提供「清除演示数据」功能
- **权限**：开通者自动成为 `tenant_admin`
- **邮箱规则**：同一邮箱可在不同租户注册，但同一租户内唯一
- **自动登录**：注册成功后自动签发 JWT，直接进入管理后台

## 后端要求

### 1. 注册接口

```
POST /api/v1/tenants/register
```

请求体：
```json
{
  "organization_name": "Acme 公司",
  "admin_email": "admin@acme.com",
  "admin_name": "管理员",
  "password": "password123"
}
```

响应：
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "tenant": {
      "id": "uuid",
      "code": "acme",
      "name": "Acme 公司"
    },
    "user": {
      "id": "uuid",
      "email": "admin@acme.com",
      "name": "管理员",
      "role": "tenant_admin"
    },
    "token": "jwt-token"
  }
}
```

### 2. 租户代码生成规则

- 从 `organization_name` 提取拼音或英文，转为小写
- 移除特殊字符，只保留字母和数字
- 如果为空或已存在，追加随机后缀（如 `acme-1`、`acme-2`）
- 最大长度 32 字符
- 示例：
  - `Acme 公司` → `acme`
  - `Acme Inc` → `acme-inc`
  - `测试` → `test`（或随机 `t-xxxxxx`）

### 3. 演示数据初始化

注册成功后自动创建：

**课程**：
- 标题：`新员工入职培训`
- 状态：`1`（已发布）
- 章节 1：`第一章：公司介绍`
  - 课时 1：`欢迎视频`（video，示例 URL）
  - 课时 2：`企业文化手册`（document，示例 URL）
- 章节 2：`第二章：规章制度`
  - 课时 3：`考勤制度`（document，示例 URL）
  - 课时 4：`安全须知`（text，示例内容）

**用户**：
- 学员 1：`learner1@example.com`，learner 角色
- 学员 2：`learner2@example.com`，learner 角色
- 讲师 1：`instructor@example.com`，instructor 角色

**课程指派**：
- 将示例课程指派给 2 个示例学员

**资源**：
- 示例图片 1 张（可生成 1x1 PNG）
- 示例文档 1 个（可生成简单 PDF）

### 4. 清除演示数据

```
DELETE /backend/v1/tenants/demo-data
```

- 仅 `tenant_admin` 可调用
- 删除当前租户下所有标记为演示的数据（通过名称或固定 ID 识别）
- 演示数据识别方式：课程标题为 `新员工入职培训`、用户邮箱为 `learner1@example.com` 等

### 5. 权限与校验

- 注册接口无需认证，但需通过租户中间件（tenant code 为 `unknown` 时走注册逻辑）
- 检查同一租户下邮箱是否已存在
- 密码 bcrypt 哈希
- 注册成功后自动登录，返回 JWT

### 6. 文件结构

- `internal/api/tenant_register.go`：注册 Handler
- `internal/service/tenant_register.go`：注册 Service
- `internal/service/demo_data.go`：演示数据初始化
- `internal/api/demo_data.go`：清除演示数据 Handler
- `internal/service/demo_data.go`：清除演示数据 Service

## 前端要求

### 管理后台注册页

更新 `web/admin/src/pages/Login.tsx`：
- 添加「开通租户」链接，跳转到 `/register`

新增 `web/admin/src/pages/Register.tsx`：
- 表单字段：组织名称、管理员邮箱、姓名、密码、确认密码
- 调用 `POST /api/v1/tenants/register`
- 成功后自动登录并跳转 Dashboard
- 显示错误信息（如邮箱已存在）

新增 `web/admin/src/api/tenant.ts` 或更新：
```typescript
export interface RegisterTenantPayload {
  organization_name: string
  admin_email: string
  admin_name: string
  password: string
}

export interface RegisterTenantResponse {
  tenant: { id: string; code: string; name: string }
  user: { id: string; email: string; name: string; role: string }
  token: string
}
```

### 管理后台清除演示数据

在 Dashboard 或设置页添加「清除演示数据」按钮：
- 调用 `DELETE /backend/v1/tenants/demo-data`
- 二次确认弹窗
- 成功后刷新页面

## 测试

- 注册接口测试：成功、邮箱重复、组织名称为空、密码过短
- 租户代码生成测试：各种组织名称输入
- 演示数据初始化测试：验证课程、用户、资源已创建
- 清除演示数据测试：验证数据已删除
- 前端注册页测试（可选）

## 协作记录

- `.codex/progress.md`：任务 14 标记为已完成
- `.codex/issues.md`：记录任何新问题
- `.codex/decisions.md`：更新租户代码生成策略和演示数据清除决策
- `.codex/knowledge-graph.md`：补充注册流程模块
- `.codex/codex-log.md`：追加本次任务记录和评审反馈

## 不要做

- 不要参考、复制或改写 PlayEdu 的代码。
- 不要修改 `/Users/imaiwork/Documents/playedu-main/` 下的任何文件。
- 不要实现邮箱验证（后续任务）。
- 不要实现支付/套餐选择（后续任务）。
- 不要把记录文件写得太长，每个文件控制在 100 行以内（`.codex/codex-log.md` 和 `DESIGN.md` 除外）。

## 验收标准

### 后端

1. `go test ./...` 全部通过。
2. `make build` 能生成可执行文件。
3. `POST /api/v1/tenants/register` 能成功创建租户、管理员和演示数据。
4. 注册返回有效 JWT。
5. `DELETE /backend/v1/tenants/demo-data` 能清除演示数据。
6. 同一租户下重复邮箱注册返回 409。

### 前端

1. `web/admin` 能 `npm install && npm run build` 成功。
2. 注册页能正常提交并跳转 Dashboard。
3. 清除演示数据按钮能正常工作。

### 协作记录

1. `.codex/progress.md` 更新
2. `.codex/issues.md` 记录新问题
3. `.codex/decisions.md` 更新决策
4. `.codex/knowledge-graph.md` 补充模块
5. `.codex/codex-log.md` 追加任务记录

## Codex 完成后需要返回

- 修改文件列表
- git diff 或 GitHub commit 链接
- 测试命令和结果（后端 `go test ./...`，前端 `npm run build`）
- 遇到的问题
- `.codex/` 记录文件的更新摘要

---

## 备注

- 完整系统设计见 `DESIGN.md`
- 协作历史见 `.codex/codex-log.md`
- 有任何设计疑问请先暂停并返回问题清单，不要猜测实现
