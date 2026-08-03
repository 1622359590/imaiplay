# 租户默认门户与统一登录设计

日期：2026-07-31
状态：已确认

## 1. 背景

当前系统主要通过请求域名识别租户。`play.imai.work` 同时被错误写入某租户的
`custom_domain`，但它又是平台保留的总后台域名。后端在该域名上将认证请求
视为无租户请求，只允许 superadmin 或唯一的 tenant_admin 登录，因此 learner
和 instructor 即使密码正确也无法登录。

租户隔离本身已经以 JWT 中的 `tenant_id` 和 Repository 查询条件为基础，问题
集中在登录前如何安全、低摩擦地确定租户。

## 2. 目标

- 客户注册后无需配置 DNS，立即获得可分享的独立品牌门户。
- 客户可选绑定自己的域名，域名变更不影响用户、课程或学习数据。
- 用户在主域名登录时自动识别所属企业，多企业账号通过选择页确定企业。
- 所有业务数据继续以可信 JWT 中的 `tenant_id` 隔离。
- 修复平台保留域名被租户绑定以及危险解绑的问题。
- 保持现有 Admin、PC、H5 链接在迁移期可用。

## 3. 非目标

- 本阶段不把现有 `users` 拆分为全局账号和 `tenant_memberships`。
- 本阶段不引入第三方身份平台、SSO、Cloudflare for SaaS 或数据库 RLS。
- 本阶段不改变课程、资源、报名和学习进度的数据模型。
- 本阶段不允许客户端通过任意 `tenant_id` 越权切换租户。

## 4. 选定方案

### 4.1 三种入口

| 入口 | 示例 | 租户来源 |
| --- | --- | --- |
| 默认门户 | `https://play.imai.work/t/acme` | URL 中的租户编码 |
| 自定义域名 | `https://learn.acme.com` | 已验证的 `custom_domain` |
| 统一登录 | `https://play.imai.work/login` | 凭证验证后的企业归属 |

默认门户是每个租户开通后立即可用的正式入口。自定义域名是可选品牌增强，
两者最终解析为同一个租户记录和 `tenant_id`。

### 4.2 URL 结构

PC 门户使用以下路由：

```text
/login
/select-organization
/t/:tenantCode
/t/:tenantCode/login
/t/:tenantCode/courses
/t/:tenantCode/courses/:courseId
/t/:tenantCode/courses/:courseId/lessons/:lessonId
/t/:tenantCode/recent
```

自定义域名使用相同页面，但保留简洁路径：

```text
/
/login
/courses
/courses/:courseId
/courses/:courseId/lessons/:lessonId
/recent
```

`/pc/*` 和 `/h5/*` 作为兼容入口保留。旧 PC 登录页重定向到统一登录页；
已有有效 token 的旧页面继续工作。

## 5. 租户解析

新增统一的 Portal Resolver，解析优先级如下：

1. 已验证的自定义域名；
2. 平台路径 `/t/:tenantCode` 对应的显式租户编码；
3. 兼容的 `X-Tenant-Code`；
4. 兼容的 `X-Tenant-ID`，先按 ID 查询再转换为租户编码；
5. 未提供租户信息时标记为平台上下文。

`play.imai.work` 必须始终是平台上下文，不能再被当成 code=`play` 的子域名。
请求中的 Host、路径或 Header 只用于登录前定位门户。登录后以 JWT 的
`tenant_id` 为数据隔离依据。

对于已经解析出门户租户的受保护请求，新增租户一致性校验：

- JWT `tenant_id` 等于门户租户：继续请求；
- 不相等：返回 HTTP 403；
- 平台上下文：允许使用 JWT 中已签发的租户；
- superadmin：仅按现有平台权限规则处理。

## 6. 门户与品牌

新增公开接口：

```text
GET /api/v1/portal
GET /api/v1/portal?tenant_code=acme
```

自定义域名请求可只使用 Host；默认路径门户传 `tenant_code`。响应仅包含公开
信息：

```json
{
  "code": "acme",
  "name": "Acme 学院",
  "logo_url": "https://cdn.example.com/logo.png",
  "primary_color": "#4F46E5",
  "welcome_text": "欢迎来到 Acme 学院",
  "default_portal_url": "https://play.imai.work/t/acme",
  "custom_domain_url": "https://learn.acme.com"
}
```

PC 门户首屏在渲染登录卡片前加载该配置，并同步更新：

- Logo、企业名称、欢迎语和主色；
- 页面标题；
- 登录页及学习中心品牌；
- 租户不存在、停用或试用到期的独立错误状态。

租户管理后台显示默认门户链接和复制按钮。自定义域名区域明确标记为
“可选”，不再让客户误以为绑定域名后才能使用。

## 7. 统一登录

### 7.1 有明确门户上下文

用户从默认门户或自定义域名登录时，只在该租户内查询账号。验证成功后签发
带该租户 `tenant_id` 的 access token 和 refresh token。

### 7.2 平台统一登录

用户从 `/login` 登录时：

1. 规范化邮箱或手机号；
2. 查询所有租户中的匹配用户，覆盖 tenant_admin、instructor 和 learner；
3. 对每条候选记录验证密码，密码验证前不返回企业信息；
4. 过滤停用用户和不可用租户；
5. 没有匹配：统一返回“账号或密码错误”；
6. 只有一个匹配：直接签发该租户 token；
7. 多个匹配：返回短期企业选择凭证和企业列表。

多企业响应：

```json
{
  "requires_tenant_selection": true,
  "selection_token": "random-one-time-token",
  "organizations": [
    {
      "code": "acme",
      "name": "Acme 学院",
      "logo_url": "https://cdn.example.com/acme.png",
      "role": "learner"
    }
  ]
}
```

新增接口：

```text
POST /api/v1/auth/select-tenant
```

请求携带 `selection_token` 和 `tenant_code`。服务端再次检查候选用户、用户
状态和租户状态，原子消费凭证后签发 token。选择凭证：

- 使用密码学安全随机值；
- 数据库只保存哈希；
- 五分钟过期；
- 只能消费一次；
- 只允许选择凭证中记录的候选用户。

登录成功后的跳转规则固定为：

- learner：进入 `/t/{tenantCode}`；
- tenant_admin、instructor：进入 `/admin/`；
- superadmin：进入 `/admin/`；
- 通过自定义域名登录的 learner：留在当前域名。

平台路径与 `/admin/` 仍在同一 Origin，token 不需要跨域传递。统一登录页按
最终角色把 token 写入对应的 Admin 或 Portal 会话 key。

## 8. 当前用户模型的兼容策略

本阶段继续把现有 `users` 行视为“租户内成员账号”。同一邮箱在不同租户可有
不同密码：

- 只有一个候选密码匹配时，自动进入对应企业；
- 多个候选使用相同密码并匹配时，显示企业选择页；
- 不自动合并历史用户，不改写报名、进度和审计关联。

未来需要“一套账号加入多个企业”时，再新增全局 `accounts` 与
`tenant_memberships`。在邮箱未验证且历史密码可能不同的情况下，当前版本
直接合并账号存在账号接管风险，因此不在本次实施。

## 9. 前端会话

Admin、PC、H5 当前共用 `imaiplay_token`，会互相覆盖。本次拆分为应用级 key：

```text
imaiplay_admin_access_token
imaiplay_admin_refresh_token
imaiplay_portal_access_token
imaiplay_portal_refresh_token
```

门户 token 还必须校验：

- 角色允许进入学习端；
- token 未过期；
- 在明确门户上下文中，token `tenant_id` 与门户租户一致。

PC 与 H5 统一 401、refresh、退出和过期跳转行为。旧 key 首次读取后按角色
迁移到新 key，再删除旧 key。

## 10. 域名安全修复

- `ADMIN_HOST` 永远不能写入任何租户的 `custom_domain`。
- 服务启动时检测与 `ADMIN_HOST` 相同的历史错误绑定，只清除数据库字段并
 记录告警，绝不调用宝塔删除站点。
- 解绑保留域名时同样只清理错误数据，不调用 `DeleteSite`。
- 普通自定义域名继续执行现有 DNS 验证、宝塔建站、代理、证书和解绑流程。
- 自定义域名与默认门户同时有效，不设置“二选一”的主入口状态。

## 11. Nginx 与缓存

容器 Nginx 增加：

- `/login`、`/select-organization`、`/t/*` 返回 PC `index.html`；
- 自定义域名 `/` 返回 PC `index.html`；
- `/admin/*` 保持 Admin SPA；
- 旧 `/pc/*`、`/h5/*` 保持兼容；
- 所有 SPA 入口 HTML 使用 `no-cache, no-store, must-revalidate`；
- 带 hash 的静态资源使用长期缓存。

宝塔主站继续代理到 `127.0.0.1:18080`。本方案的默认门户不需要为每个租户
创建宝塔站点或证书。

## 12. 数据库变更

新增 `login_challenges` 表，至少包含：

```text
id
token_hash
candidate_user_ids
expires_at
consumed_at
created_at
```

新增跨租户登录查询索引：

```text
LOWER(email)
phone
```

不删除或重写现有用户数据。

## 13. 错误与状态

| 场景 | 行为 |
| --- | --- |
| 门户编码不存在 | 404，显示“企业空间不存在” |
| 租户暂停或到期 | 403，显示明确状态，不显示密码错误 |
| 账号或密码错误 | 401，统一文案，不泄露企业 |
| 多企业 | 200，进入企业选择流程，不签发访问 token |
| 选择凭证过期/重放 | 401，返回统一登录页 |
| token 与门户不匹配 | 403，清理门户会话并回到当前门户登录页 |
| 保留域名历史绑定 | 安全清理数据库，不删除平台站点 |

## 14. 测试

后端必须覆盖：

- 平台 Host 不再解析为租户 `play`；
- 默认门户编码、自定义域名和 Header 映射到同一租户；
- 单企业、多企业、无企业和错误密码；
- learner、instructor、tenant_admin、superadmin；
- 禁用用户、暂停租户和过期租户；
- 企业信息只在密码成功后返回；
- 选择凭证篡改、过期、重放和越权选择；
- JWT 与门户租户不一致返回 403；
- 保留域名修复与解绑不会调用宝塔删除站点；
- refresh token 始终保持已选择的 `tenant_id`。

前端必须覆盖：

- 默认门户、统一登录和企业选择；
- 品牌加载、404、停用和到期状态；
- 应用级 token 隔离和旧 token 迁移；
- PC/H5 401、refresh 和退出；
- 旧 `/pc/*`、`/h5/*` 兼容导航；
- 三端生产构建。

验收流程：

```text
租户注册
→ 复制默认门户
→ learner 直接登录
→ 加载本租户品牌和课程
→ 同邮箱多企业登录并选择企业
→ 绑定自定义域名后从新域名登录
→ 两个入口数据完全一致且互不串租户
```

## 15. 发布顺序

1. 发布后端数据库迁移、门户解析和新登录接口；
2. 发布 PC/H5/Admin 前端与 Nginx 路由；
3. 启动时安全修复生产库中的保留域名错误绑定；
4. 验证默认门户和统一登录；
5. 验证原有自定义域名；
6. 保留旧入口至少一个发布周期后再评估下线。

回滚时保留新增表和索引，旧认证接口与旧入口仍可继续工作。
