# 学员批量导入 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为站长提供安全的 CSV/XLSX 学员批量导入、部分成功反馈和无密码错误明细下载。

**Architecture:** 后端把文件解析、逐行业务导入和 HTTP 上传处理分离；逐行业务导入复用现有 `CreateWithPhone`，确保租户隔离、唯一性和密码哈希一致。前端用独立纯函数生成模板及错误 CSV，用弹窗承载上传和结果反馈，并在成功创建后刷新用户列表。

**Tech Stack:** Go 1.22、Gin、GORM、`github.com/xuri/excelize/v2`、React 18、TypeScript、Ant Design、Axios、Node test runner

## Global Constraints

- 支持 `.xlsx` 和 UTF-8 `.csv`；单次最多 1000 条非空数据。
- 固定中文表头：`姓名,邮箱,手机号（可选）,角色（可选）,初始密码`。
- 角色留空默认 `learner`；只允许学员/`learner`、讲师/`instructor`。
- 姓名和邮箱必填，邮箱合法，密码至少 8 位；手机号沿用现有规则。
- 逐行部分成功；文件级格式错误整份拒绝。
- 结果响应、错误明细、日志均不得包含初始密码。
- 只有 `tenant_admin` 可导入；客户端不能指定租户。
- 不改动工作区中与本功能无关的既有修改。

---

### Task 1: CSV/XLSX 文件解析器

**Files:**
- Create: `internal/service/user_import_file.go`
- Create: `internal/service/user_import_file_test.go`
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Consumes: `io.Reader` 和上传文件名。
- Produces: `ParseUserImportFile(filename string, reader io.Reader) ([]UserImportRow, error)`；`UserImportRow` 包含 `Row int`、`Name`、`Email`、`Phone`、`Role`、`Password`。

- [ ] **Step 1: 写 CSV 解析失败测试**

在 `internal/service/user_import_file_test.go` 用手写 CSV 字符串断言：正确表头返回原始行号 2 和 4（中间空行忽略）；错误表头、错误扩展名、1001 条数据均返回 `errorsx.BadRequest`。

```go
func TestParseUserImportCSVPreservesSourceRows(t *testing.T) {
    input := "姓名,邮箱,手机号（可选）,角色（可选）,初始密码\n张三,ZHANG@example.com,,学员,password1\n,,,,\n李四,li@example.com,13800138000,instructor,password2\n"
    rows, err := ParseUserImportFile("users.csv", strings.NewReader(input))
    if err != nil || len(rows) != 2 || rows[0].Row != 2 || rows[1].Row != 4 {
        t.Fatalf("rows=%#v err=%v", rows, err)
    }
}
```

- [ ] **Step 2: 运行 CSV 测试并确认因解析器不存在而失败**

Run: `go test ./internal/service -run 'TestParseUserImportCSV' -count=1`

Expected: FAIL，提示 `ParseUserImportFile` 或 `UserImportRow` 未定义。

- [ ] **Step 3: 最小实现 CSV 解析**

在 `user_import_file.go` 定义固定表头、1000 行上限、空白行判断、列数补齐与 `encoding/csv` 解析。扩展名用 `strings.ToLower(filepath.Ext(filename))` 判断；CSV 先用 `utf8.Valid` 拒绝非 UTF-8 内容；UTF-8 BOM 仅允许出现在第一个表头单元格并在比较前移除。

```go
type UserImportRow struct {
    Row      int
    Name     string
    Email    string
    Phone    string
    Role     string
    Password string
}

func ParseUserImportFile(filename string, reader io.Reader) ([]UserImportRow, error)
```

- [ ] **Step 4: 运行 CSV 测试并确认通过**

Run: `go test ./internal/service -run 'TestParseUserImportCSV' -count=1`

Expected: PASS。

- [ ] **Step 5: 写 XLSX 解析失败测试**

测试内用 `excelize.NewFile()` 构造工作簿，写入标准表头、空白行和两条数据；断言使用第一个工作表、保留原始行号，并对无工作表/错误表头返回 40000。

- [ ] **Step 6: 运行 XLSX 测试并确认因不支持 XLSX 而失败**

Run: `go test ./internal/service -run 'TestParseUserImportXLSX' -count=1`

Expected: FAIL，错误为不支持 `.xlsx` 或测试编译时缺少 excelize。

- [ ] **Step 7: 添加 XLSX 依赖并实现解析**

Run: `go get github.com/xuri/excelize/v2`

使用 `excelize.OpenReader`、`GetSheetList`、`Rows` 读取首个工作表，并把每行交给与 CSV 共用的表头/空行/上限处理函数。确保 `defer workbook.Close()` 和 `defer rows.Close()`。

- [ ] **Step 8: 运行解析器测试和依赖整理**

Run: `go test ./internal/service -run 'TestParseUserImport(CSV|XLSX)' -count=1 && go mod tidy`

Expected: PASS，且 `go.mod` 只新增解析 XLSX 所需的直接/间接依赖。

- [ ] **Step 9: 提交解析器**

```bash
git add internal/service/user_import_file.go internal/service/user_import_file_test.go go.mod go.sum
git commit -m "feat: parse user import files"
```

### Task 2: 逐行用户导入服务

**Files:**
- Create: `internal/service/user_import.go`
- Create: `internal/service/user_import_test.go`
- Modify: `internal/service/user.go`

**Interfaces:**
- Consumes: Task 1 的 `[]UserImportRow` 和现有 `CreateWithPhone`。
- Produces: `(*UserService).Import(ctx context.Context, rows []UserImportRow) UserImportResult`；结果含 `Total`、`Succeeded`、`Failed`、`Errors []UserImportError`，错误不含密码。

- [ ] **Step 1: 写部分成功和默认角色的失败测试**

构造站长上下文，导入合法学员、合法讲师、重复邮箱和短密码四行。断言前两行创建成功，空角色变成 `learner`，后两行带原始行号失败，结果 JSON 序列化后不含任何密码文本。

```go
result := users.Import(admin, []UserImportRow{
    {Row: 2, Name: "张三", Email: "ZHANG@example.com", Password: "password1"},
    {Row: 3, Name: "李四", Email: "li@example.com", Role: "讲师", Password: "password2"},
    {Row: 4, Name: "重复", Email: "zhang@example.com", Password: "password3"},
    {Row: 5, Name: "弱密码", Email: "weak@example.com", Password: "short"},
})
```

- [ ] **Step 2: 运行测试并确认因 Import 不存在而失败**

Run: `go test ./internal/service -run 'TestUserImport' -count=1`

Expected: FAIL，提示 `Import`/结果类型未定义。

- [ ] **Step 3: 实现最小逐行导入**

在 `user_import.go` 定义 JSON 类型和角色映射函数。逐行 `TrimSpace` 姓名/邮箱/手机号/角色，邮箱转小写并用 `net/mail.ParseAddress` 且 `parsed.Address == email` 验证；密码按原字符串计算长度，不写入错误结构。合法行调用 `CreateWithPhone`，错误原因使用 `errorsx.LocalizeMessage(err.Error())`。

```go
type UserImportError struct {
    Row    int    `json:"row"`
    Name   string `json:"name"`
    Email  string `json:"email"`
    Phone  string `json:"phone"`
    Role   string `json:"role"`
    Reason string `json:"reason"`
}

type UserImportResult struct {
    Total     int               `json:"total"`
    Succeeded int               `json:"succeeded"`
    Failed    int               `json:"failed"`
    Errors    []UserImportError `json:"errors"`
}
```

- [ ] **Step 4: 运行服务测试并确认通过**

Run: `go test ./internal/service -run 'TestUserImport' -count=1`

Expected: PASS。

- [ ] **Step 5: 补充权限、字段和租户隔离失败测试**

表驱动覆盖 learner/superadmin 无权限、空姓名、非法邮箱、非法手机号、站长角色和未知角色；另建 tenant-2 同邮箱后断言 tenant-1 仍可导入。每个断言检查明确的行号、失败计数和数据库实际状态。

- [ ] **Step 6: 运行服务包测试**

Run: `go test ./internal/service -count=1`

Expected: PASS。

- [ ] **Step 7: 提交导入服务**

```bash
git add internal/service/user.go internal/service/user_import.go internal/service/user_import_test.go
git commit -m "feat: import tenant users in batches"
```

### Task 3: 上传 API 与路由

**Files:**
- Modify: `internal/api/user.go`
- Modify: `internal/api/user_test.go`
- Modify: `internal/server/server.go`
- Modify: `internal/server/server_test.go`

**Interfaces:**
- Consumes: Task 1 的 `ParseUserImportFile`、Task 2 的 `Import`。
- Produces: `POST /backend/v1/users/import` multipart 接口，字段名 `file`，返回 `UserImportResult`。

- [ ] **Step 1: 写 multipart API 失败测试**

在 `internal/api/user_test.go` 增加创建 multipart 请求的测试辅助函数，上传包含一条合法和一条非法记录的 CSV；断言 HTTP 200、`succeeded=1`、`failed=1`，响应正文不包含 `password1`/`short`。另用 learner 路由断言 403，用缺少文件和 `.txt` 文件断言 400。

- [ ] **Step 2: 运行 API 测试并确认路由/处理器不存在而失败**

Run: `go test ./internal/api -run 'TestUserImportHandler' -count=1`

Expected: FAIL，提示 `Import` handler 不存在或返回 404。

- [ ] **Step 3: 扩展接口并实现 handler**

在 `UserService` 接口加入：

```go
Import(ctx context.Context, rows []service.UserImportRow) service.UserImportResult
```

`UserHandler.Import` 先调用 `requireHandlerRole(c, "tenant_admin")`，再用 `http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)` 把请求限制为 10 MiB，随后用 `c.FormFile("file")`/`Open()` 获取流，调用解析器和服务，文件级错误走 `errorsx.GinResponse`，结果用 `success` 返回。

- [ ] **Step 4: 运行 API 测试并确认通过**

Run: `go test ./internal/api -run 'TestUserImportHandler' -count=1`

Expected: PASS。

- [ ] **Step 5: 写并验证路由注册测试**

在 `internal/server/server_test.go` 断言 `POST /backend/v1/users/import` 出现在 `router.Routes()`；先运行看到缺路由失败，再在 `server.go` 注册 `backend.POST("/users/import", userHandler.Import)`，注意必须放在 `GET /users/:id` 等参数路由之前。

Run: `go test ./internal/server -run 'TestUserImportRouteRegistered' -count=1`

Expected: RED 后 GREEN。

- [ ] **Step 6: 运行后端相关测试**

Run: `go test ./internal/service ./internal/api ./internal/server -count=1`

Expected: PASS。

- [ ] **Step 7: 提交 API**

```bash
git add internal/api/user.go internal/api/user_test.go internal/server/server.go internal/server/server_test.go
git commit -m "feat: expose bulk user import API"
```

### Task 4: 前端导入契约与安全 CSV 下载

**Files:**
- Create: `web/admin/src/utils/userImport.ts`
- Create: `web/admin/tests/userImport.test.ts`
- Modify: `web/admin/src/api/user.ts`

**Interfaces:**
- Consumes: 浏览器 `File`/`Blob`，Axios client，后端 `UserImportResult`。
- Produces: `validateUserImportFile`、`userImportTemplateCSV`、`userImportErrorsCSV`、`downloadUserImportCSV` 和 `userApi.import(file)`。

- [ ] **Step 1: 写纯函数失败测试**

测试 `.csv`/`.xlsx` 大小写扩展名通过、其他扩展名拒绝；模板 CSV 以 BOM 开头并包含固定表头；错误 CSV 正确转义逗号、双引号和换行，且不包含 `password`、`初始密码` 或任意输入密码。

```ts
assert.equal(
  userImportErrorsCSV([{ row: 3, name: '张,三', email: 'a@example.com', phone: '', role: 'learner', reason: '邮箱"重复' }]),
  '\uFEFF行号,姓名,邮箱,手机号,角色,失败原因\r\n3,"张,三",a@example.com,,learner,"邮箱""重复"',
)
```

- [ ] **Step 2: 运行前端测试并确认模块不存在而失败**

Run: `cd web/admin && node --test tests/userImport.test.ts`

Expected: FAIL，提示找不到 `userImport.ts`。

- [ ] **Step 3: 实现纯函数与 API 类型**

定义 `UserImportError`/`UserImportResult`，CSV 转义规则为包含逗号、双引号或换行时加双引号并把内部双引号加倍。`downloadUserImportCSV` 创建 Blob、对象 URL、临时 `<a>`，点击后撤销 URL。`userApi.import` 构造 `FormData` 并 POST `/backend/v1/users/import`。

- [ ] **Step 4: 运行前端测试并确认通过**

Run: `cd web/admin && node --test tests/userImport.test.ts`

Expected: PASS。

- [ ] **Step 5: 提交前端契约**

```bash
git add web/admin/src/api/user.ts web/admin/src/utils/userImport.ts web/admin/tests/userImport.test.ts
git commit -m "feat: add user import client helpers"
```

### Task 5: 批量导入弹窗与页面入口

**Files:**
- Create: `web/admin/src/components/UserImportModal.tsx`
- Modify: `web/admin/src/pages/Users.tsx`

**Interfaces:**
- Consumes: Task 4 的 API、校验和下载函数。
- Produces: `UserImportModal`，props 为 `open: boolean`、`onClose(): void`、`onImported(): void`。

- [ ] **Step 1: 写可测试的结果状态失败测试**

在 `web/admin/tests/userImport.test.ts` 增加对 `importResultSummary(result)` 的行为测试，断言全成功、部分成功和全失败分别生成明确的成功/警告状态与数量；该函数放在 `userImport.ts`，避免依赖 DOM 测试库。

- [ ] **Step 2: 运行测试并确认函数不存在而失败**

Run: `cd web/admin && node --test tests/userImport.test.ts`

Expected: FAIL，提示 `importResultSummary` 未导出。

- [ ] **Step 3: 实现结果状态和弹窗**

先最小实现 `importResultSummary` 通过测试，再创建弹窗：使用 Ant Design `Upload.Dragger` 且 `beforeUpload={() => false}`，只保留一个文件；提交时再次校验文件，调用 API；请求中禁用上传和确认按钮；成功后保存结果并调用 `onImported`；失败交由现有 Axios 拦截器提示。结果区展示总数、成功数、失败数，失败时提供错误 CSV 下载。

- [ ] **Step 4: 接入 Users 页面**

站长页头 `extra` 改为 `Space`，顺序为“批量导入”次按钮、“新增用户”主按钮；总管理员仍无按钮。管理 `importOpen` 状态，关闭不影响单条新增弹窗，`onImported` 调用 `load()`。

- [ ] **Step 5: 运行测试和 TypeScript 生产构建**

Run: `cd web/admin && npm test && npm run build`

Expected: 所有测试 PASS，TypeScript 与 Vite 构建 exit 0。

- [ ] **Step 6: 提交 UI**

```bash
git add web/admin/src/components/UserImportModal.tsx web/admin/src/pages/Users.tsx web/admin/src/utils/userImport.ts web/admin/tests/userImport.test.ts
git commit -m "feat: add bulk import flow to users page"
```

### Task 6: 全量验证与验收

**Files:**
- Verify only; fix only failures caused by this feature in the files above.

**Interfaces:**
- Consumes: Tasks 1–5 的完整功能。
- Produces: 可复现的测试、构建和差异审查证据。

- [ ] **Step 1: 运行 Go 全量测试**

Run: `go test ./... -count=1`

Expected: PASS，0 failures。

- [ ] **Step 2: 运行管理端全量测试与构建**

Run: `cd web/admin && npm test && npm run build`

Expected: PASS，构建 exit 0。

- [ ] **Step 3: 审查差异和敏感字段**

Run: `git diff --check && git diff --stat && rg -n 'Password|password|初始密码' internal/service/user_import.go internal/api/user.go web/admin/src/utils/userImport.ts web/admin/src/components/UserImportModal.tsx`

逐项确认密码只存在输入/校验路径，不在 `UserImportError`、JSON 响应、错误 CSV 或日志中；确认未包含无关工作区文件。

- [ ] **Step 4: 验收接口行为**

使用 API 测试中的 multipart 场景复核：合法行进入当前租户；错误行未创建；中文/英文角色和默认学员生效；响应成功/失败数匹配；总管理员与 learner 均为 403。

- [ ] **Step 5: 提交验证阶段必要修正（仅在存在修正时）**

仅暂存 Tasks 1–5 列出的、确实在验证阶段修正过的文件，然后提交 `fix: harden bulk user import`；若无需修正则不创建空提交。
