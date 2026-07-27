# 当前任务：统计看板

> 本文件由 Claude 生成，Codex 请直接阅读并执行。
> 执行前请先阅读 `DESIGN.md` 和 `.codex/codex-log.md` 了解完整设计和项目状态。
> 当前分支：`codex/first-core-features`。

## 目标

完成 ImaiPlay 任务 10：统计看板，为管理后台提供租户级数据统计接口和首页展示。

## 统计口径（第一阶段）

1. **今日学习人数**：按当天更新过 `lesson_progress.updated_at` 的去重 `user_id` 数计算
2. **学习时长**：按各课时 `last_position_seconds` 求和（粗略估计，非真实观看时长）
3. **课程完成率**：按「全部课时完成的有效报名人数 ÷ 有效报名人数」计算

## 后端要求

### 1. 统计指标

租户级 Dashboard 接口返回以下数据：

```json
{
  "user_count": 100,           // 学员总数
  "course_count": 10,          // 课程总数
  "published_course_count": 8, // 已发布课程数
  "today_new_user_count": 5,   // 今日新增学员数
  "today_learning_user_count": 12, // 今日学习人数（去重）
  "total_learning_seconds": 36000, // 总学习时长（秒）
  "course_completion_rate": 0.75   // 课程完成率（0-1）
}
```

### 2. API 接口

```
GET /backend/v1/dashboard
```

- 权限：仅 `tenant_admin` / `instructor` 可访问
- 按 JWT `tenant_id` 隔离

### 3. 数据查询

**基础统计**：
- 学员总数：`users` 表中当前 tenant 且 `status = 1` 的记录数
- 课程总数：`courses` 表中当前 tenant 的记录数
- 已发布课程数：`courses` 表中当前 tenant 且 `status = 1` 的记录数
- 今日新增学员：`users` 表中当前 tenant 且 `created_at` 为今天的记录数

**今日学习人数**：
```sql
SELECT COUNT(DISTINCT user_id)
FROM lesson_progress
WHERE tenant_id = ? AND DATE(updated_at) = CURRENT_DATE
```

**总学习时长**：
```sql
SELECT COALESCE(SUM(last_position_seconds), 0)
FROM lesson_progress
WHERE tenant_id = ?
```

**课程完成率**：
```sql
-- 全部课时完成的有效报名人数
SELECT COUNT(DISTINCT ce.user_id)
FROM course_enrollments ce
WHERE ce.tenant_id = ? AND ce.status = 1
  AND NOT EXISTS (
      SELECT 1 FROM course_lessons cl
      JOIN course_chapters cc ON cl.chapter_id = cc.id
      WHERE cc.course_id = ce.course_id
        AND NOT EXISTS (
            SELECT 1 FROM lesson_progress lp
            WHERE lp.user_id = ce.user_id
              AND lp.lesson_id = cl.id
              AND lp.status = 2
        )
  )

-- 有效报名人数
SELECT COUNT(DISTINCT user_id)
FROM course_enrollments
WHERE tenant_id = ? AND status = 1
```

注意：如果有效报名人数为 0，完成率返回 0。

### 4. 实现结构

- `internal/api/dashboard.go`：DashboardHandler
- `internal/service/dashboard.go`：DashboardService
- `internal/repository/dashboard.go`：DashboardRepository（可自定义查询）
- `internal/server/server.go`：注册路由
- `cmd/server/main.go`：组装依赖

### 5. 测试

- Service 测试覆盖各统计指标计算
- API Handler 测试覆盖权限和数据返回

## 前端要求

### 管理后台 Dashboard 页面

更新 `web/admin/src/pages/Dashboard.tsx`：

- 调用 `/backend/v1/dashboard` 获取统计数据
- 使用 Ant Design `Statistic` 和 `Card` 组件展示
- 布局参考 PlayEdu 后台首页风格
- 展示指标：
  - 学员总数
  - 课程总数 / 已发布课程
  - 今日新增学员
  - 今日学习人数
  - 总学习时长（转换为小时显示）
  - 课程完成率（百分比显示）

### API 客户端

新增 `web/admin/src/api/dashboard.ts`：
```typescript
export interface DashboardStats {
  user_count: number
  course_count: number
  published_course_count: number
  today_new_user_count: number
  today_learning_user_count: number
  total_learning_seconds: number
  course_completion_rate: number
}
```

## 协作记录

- `.codex/progress.md`：任务 10 标记为已完成
- `.codex/issues.md`：记录任何新问题
- `.codex/decisions.md`：记录统计口径决策
- `.codex/knowledge-graph.md`：补充 dashboard 模块
- `.codex/codex-log.md`：追加本次任务记录和评审反馈

## 不要做

- 不要参考、复制或改写 PlayEdu 的代码。
- 不要修改 `/Users/imaiwork/Documents/playedu-main/` 下的任何文件。
- 不要实现复杂图表库（ECharts/Ant Design Charts），先用数字卡片即可。
- 不要实现 superadmin 平台级统计，只做租户级。
- 不要把记录文件写得太长，每个文件控制在 100 行以内（`.codex/codex-log.md` 和 `DESIGN.md` 除外）。

## 验收标准

### 后端

1. `go test ./...` 全部通过。
2. `make build` 能生成可执行文件。
3. `GET /backend/v1/dashboard` 返回正确的租户统计数据。
4. 无权限访问返回 403。

### 前端

1. `web/admin` 能 `npm install && npm run build` 成功。
2. 管理后台首页能展示统计数据卡片。
3. 数据与后端接口返回一致。

### 协作记录

1. `.codex/progress.md` 更新
2. `.codex/issues.md` 记录新问题
3. `.codex/decisions.md` 记录统计口径决策
4. `.codex/knowledge-graph.md` 补充新模块
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
