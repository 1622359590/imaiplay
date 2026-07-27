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

---

## 協作流程更新

### 2026-07-27

為減少人工轉發成本，任務指令統一寫入 `.codex/current-task.md`。

**Claude 的工作方式**：
- 規劃任務並將完整指令寫入 `.codex/current-task.md`
- 更新 `.codex/codex-log.md` 記錄決策與狀態
- 評審 Codex 完成後的代碼與記錄

**用戶的工作方式**：
- 告訴 Codex：「請閱讀 `.codex/current-task.md`、`DESIGN.md` 和 `.codex/codex-log.md`，然後執行任務。」
- Codex 完成後，告訴 Claude：「检查一下」

**Codex 的工作方式**：
- 開始前閱讀 `.codex/current-task.md`、`DESIGN.md`、`.codex/codex-log.md`
- 執行任務
- 完成後更新 `.codex/progress.md`、`.codex/issues.md`、`.codex/decisions.md`、`.codex/knowledge-graph.md`、`.codex/codex-log.md`
- 在 `.codex/current-task.md` 頂部標記任務狀態與執行摘要
- 返回執行結果給用戶

---

## Codex 執行記錄：第一批核心功能

### 任務執行摘要

- 完成統一錯誤、bcrypt、JWT、認證中間件及用戶上下文。
- 完成認證、租戶、用戶的 Repository、Service、Handler 與路由。

### 關鍵修改

- 公開註冊限制為租戶角色，`superadmin` 返回 HTTP 400 / `40000`。
- User 查詢以 JWT `tenant_id` 隔離，並建立 Tenant RESTRICT 外鍵。
- 新增統一 API 響應、User 遷移及租戶郵箱唯一索引。

### 評審反饋

- 修復 DB 故障誤報 401、缺少外鍵、Handler CRUD 測試不足等問題。
- 復審無 Critical/Important，結論 Ready；真實 PostgreSQL 遷移通過。

### 下一步建議

- 確定 superadmin 初始化與登入機制，再進入課程、章節及課時管理。

---

## 設計變更：增加前端

### 2026-07-27

用戶明確需求：ImaiPlay 需要完整的培訓平台，包含前端學員端和管理後台，不是純後端 API 項目。

**決策**：
- 前端技術棧：React 18 + TypeScript + Vite + Ant Design 5
- H5 學員端：Ant Design Mobile 5
- 前端放在同一倉庫：`web/admin`、`web/pc`、`web/h5`
- 管理後台和學員端一起做
- UI 完全模仿 PlayEdu 的界面風格和佈局，但不複製其代碼
- 狀態管理：Redux Toolkit
- API 客戶端：axios，Token 存 localStorage

**調整後規劃**：
- 後端繼續完善課程/章節/課時 API
- 前端同步搭建三端骨架和基礎頁面
- 第二批再實現視頻播放、學習進度、資源存儲

---

## Codex 執行記錄：Tenant code 不可修改

### 任務執行摘要

- 修復 Tenant Repository 更新時可修改 code 的問題。

### 關鍵修改

- Update 僅寫入 name 與 status；Repository、Service、API 測試均斷言 code 不變。

### 評審反饋

- 根因位於 Repository Updates map；Service 與 API 簽名原本已符合要求。

### 下一步建議

- 等待 Claude 評審後合併第一批核心功能。

---

## Codex 執行記錄：任務 7

### 任務執行摘要

- 完成課程、章節、課時管理 API 與學員已發佈課程 API。
- 完成管理後台、PC 學員端及 H5 學員端 React 骨架。

### 關鍵修改

- Repository 以 JWT `tenant_id` 隔離，講師只管理自己建立的課程。
- 課程詳情組裝章節與課時；刪除課程或章節時以事務清理子內容。
- 新增三端登入、課程頁面、API 攔截器及最小 CORS。

### 評審反饋

- 修復學員 API 角色邊界、孤兒內容、前後端字段與路由契約問題。
- 移除會被後端忽略的表單選項及 H5 錯誤時的演示資料回退。

### 下一步建議

- 任務 8 實現課程指派、學習進度與最近學習，再按路由拆分前端包。
