# Codex 協作日誌

> 本文件由 Claude（AI 開發負責人）維護，供 Codex 在接手新任務前閱讀。
> 記錄任務指令、關鍵決策、評審反饋和當前狀態。

---

## 項目概覽

- **項目名稱**：ImaiPlay
- **定位**：多租戶企業培訓 SaaS 後端
- **語言**：Go 1.22+
- **技術棧**：Gin + Viper + PostgreSQL + GORM + JWT（待實現）
- **核心原則**：從零開發，不參考、不複製 PlayEdu 或其他現有系統代碼
- **倉庫**：`/Users/imaiwork/Documents/imaiplay-go/`

---

## 已完成的任務

### 任務 1：Go 項目骨架 + 租戶識別中間件 + /health 接口

**時間**：2026-07-27
**狀態**：✅ 已完成

**要求摘要**：
- 使用 Gin + Viper 建立項目骨架
- 實現基於子域名和 Header 的租戶識別中間件
- `/health` 返回 app_name、version、tenant 信息
- 編寫 Makefile 和單元測試

**關鍵實現**：
- `internal/middleware/tenant.go`：租戶識別中間件
- `internal/context/tenant.go`：租戶上下文
- `internal/server/server.go`：HTTP 服務器
- `internal/config/config.go`：配置加載

---

### 任務 1.5：修復配置路徑依賴與補充租戶識別測試

**時間**：2026-07-27
**狀態**：✅ 已完成

**修復內容**：
- 配置加載優先查找當前目錄 `.env`，找不到則查找可執行文件同目錄 `.env`
- 補充帶端口、IP 地址、一級域名的租戶識別測試
- 清理測試中的環境變量處理

---

### 任務 2：數據庫接入與 GORM 基礎

**時間**：2026-07-27
**狀態**：✅ 已完成（本地已提交，待 push）

**要求摘要**：
- 接入 PostgreSQL，使用 GORM
- 採用函數式健康檢查設計：`server.New(cfg, dbCheck func() error)`
- 建立 `Tenant` 模型與 Repository
- 自動遷移 `Tenant` 表
- `/health/db` 連通返回 HTTP 200，斷開返回 HTTP 503

**關鍵實現**：
- `internal/db/db.go`：PostgreSQL 連接、連接池、Ping/Close
- `internal/domain/tenant.go`：Tenant 模型，BeforeCreate 生成 UUID
- `internal/repository/tenant.go` + `tenant_gorm.go`：Repository 接口與實現
- `internal/migration/migration.go`：GORM 自動遷移
- `internal/server/server.go`：注入 `dbCheck`，新增 `/health/db`
- `cmd/server/main.go`：初始化數據庫、遷移、啟動服務

**關鍵決策**：
- 數據庫：PostgreSQL
- ORM：GORM
- Server 不直接依賴 GORM，通過 `func() error` 注入健康檢查
- Repository 測試使用 SQLite 內存數據庫

---

## 待完成的任務

按以下順序推進：

1. **任務 4：用戶認證與 JWT**
   - 用戶註冊、登錄
   - JWT Token 生成與驗證
   - 密碼哈希（bcrypt）
   - 中間件驗證 JWT

2. **任務 5：租戶模型與超級管理員接口**
   - Tenant CRUD 接口
   - 超級管理員權限控制

3. **任務 6：用戶管理與 RBAC**
   - 用戶 CRUD
   - 角色：superadmin / tenant_admin / instructor / learner
   - 權限控制

---

## 重要約定

### 代碼規範

- 不參考、不複製 PlayEdu 代碼
- 每個任務邊界清楚，一次只做一件事
- 模塊邊界清晰：`internal/domain`、`internal/repository`、`internal/service`、`internal/api`
- 協作記錄文件控制在 80 行以內（本文件除外）

### 協作記錄

每次完成任務後必須更新：
- `.codex/progress.md`：標記任務狀態
- `.codex/issues.md`：記錄新問題或標記已決定事項
- `.codex/decisions.md`：記錄架構決策
- `.codex/knowledge-graph.md`：更新模塊關係
- **本文件 `.codex/codex-log.md`**：新增任務記錄和評審反饋

### 測試要求

- `go test ./...` 必須通過
- `make build` 必須成功
- 關鍵功能必須有單元測試

---

## 當前狀態

- 任務 1、1.5、2 已完成
- 本地倉庫領先 `origin/main` 1 個 commit（`5fa245b feat: add PostgreSQL and GORM foundation`）
- 下一步建議：任務 4（用戶認證與 JWT）

---

## 如何閱讀本文件

Codex 在開始新任務前，請先閱讀本文件，了解：
1. 項目背景和技術棧
2. 已完成的任務和關鍵實現
3. 待完成的任務列表
4. 重要約定和規範
5. 當前狀態

然後再閱讀 Claude 給出的具體任務指令。

---

## Codex 執行記錄：任務 2

### 任務執行摘要

- 完成 PostgreSQL、GORM、Tenant Repository、自動遷移及 `/health/db`。
- 完成配置、SQLite CRUD、遷移、健康檢查與連接失敗測試。

### 關鍵修改

- 新增 `internal/db`、`internal/domain`、`internal/repository`、`internal/migration`。
- Server 以 `func() error` 注入數據庫健康檢查，main 完成連接與遷移接線。

### 評審反饋

- 評審指出 GORM `default:1` 會將 `Status int` 的零值視為未設置。
- 已確認保持 `Status int`，並在 `.codex/issues.md` 記錄此限制。

### 下一步建議

- 任務 4 實現用戶認證與 JWT 前，先明確用戶模型與租戶關聯。

---

## Codex 執行記錄：本地 PostgreSQL 安裝

### 任務執行摘要

- 透過 Homebrew 安裝並啟動 PostgreSQL 18.4。
- 建立 `postgres` 本地角色與 `imaiplay` 數據庫。

### 關鍵修改

- 本地 PostgreSQL 服務已設為登入後自動啟動。
- 使用項目默認配置完成連接、遷移及健康檢查驗證，未修改業務代碼。

### 評審反饋

- `go test -count=1 ./...`、`make build` 均通過。
- `/health/db` 返回 HTTP 200，且 `tenants` 表已成功建立。

### 下一步建議

- 開發環境可直接使用默認連接配置；部署時應單獨配置安全密碼。
