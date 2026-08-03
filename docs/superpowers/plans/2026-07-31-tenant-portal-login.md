# Tenant Portal and Unified Login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give every tenant an immediately usable branded portal at `/t/{tenantCode}`, support safe organization discovery from the platform login, and keep custom domains as an optional alias over the same tenant-isolated data.

**Architecture:** Keep the existing tenant-scoped `users` rows and JWT `tenant_id` model. Add a public portal resolver, a database-backed one-time organization-selection challenge, explicit platform-host handling, and a JWT-to-portal tenant consistency check. Serve the responsive PC application at platform portal routes while retaining Admin, PC, and H5 compatibility paths.

**Tech Stack:** Go 1.22+, Gin, GORM, PostgreSQL/SQLite tests, React 18, TypeScript, React Router, Axios, Ant Design, Nginx, Docker Compose.

## Global Constraints

- `play.imai.work` is always the platform host and cannot be a tenant custom domain.
- Default tenant portal URL is exactly `https://play.imai.work/t/{tenantCode}`.
- Custom domains remain optional and resolve to the same immutable `tenant_id`.
- Business data isolation always uses the signed JWT `tenant_id`, never a client-supplied tenant ID.
- Organization names are disclosed only after successful credential verification.
- Organization-selection tokens expire after five minutes, are stored hashed, and are consumed once.
- Existing `users` rows are not merged or rewritten in this release.
- Existing `/admin/*`, `/pc/*`, and `/h5/*` links remain compatible.
- Do not modify `/Users/imaiwork/Documents/playedu-main/`.
- Preserve the unrelated untracked repository-root `server` file.

---

## File Structure

### Backend files to create

- `internal/domain/login_challenge.go`: persisted one-time organization-selection challenge.
- `internal/repository/login_challenge.go`: challenge repository interface.
- `internal/repository/login_challenge_gorm.go`: atomic create/consume implementation.
- `internal/repository/login_challenge_gorm_test.go`: repository expiry and replay tests.
- `internal/service/portal.go`: public portal metadata resolution and URL generation.
- `internal/service/portal_test.go`: code, custom-domain, inactive, and missing portal tests.
- `internal/api/portal.go`: public `GET /api/v1/portal` handler.
- `internal/api/portal_test.go`: public response and error status tests.
- `internal/middleware/tenant_match.go`: JWT tenant versus resolved portal guard.
- `internal/middleware/tenant_match_test.go`: matching, platform, mismatch, and superadmin tests.
- `internal/migration/reserved_domain.go`: safe cleanup for legacy `ADMIN_HOST` bindings.
- `internal/migration/reserved_domain_test.go`: confirms cleanup never invokes Baota.

### Backend files to modify

- `internal/migration/migration.go`: migration v13 and cross-tenant credential indexes.
- `internal/migration/migration_test.go`: table and index assertions.
- `internal/repository/user.go`: all-role cross-tenant credential lookup contract.
- `internal/repository/user_gorm.go`: remove tenant-admin-only filter.
- `internal/repository/user_gorm_test.go`: all-role lookup and normalization tests.
- `internal/service/auth.go`: login outcome, password-first discovery, selection flow.
- `internal/service/auth_test.go`: single/multiple/invalid/inactive login cases.
- `internal/api/auth.go`: new login response and selection endpoint.
- `internal/api/auth_test.go`: response contract and anti-enumeration tests.
- `internal/middleware/tenant.go`: platform-host-aware resolution and correct `X-Tenant-ID`.
- `internal/middleware/tenant_test.go`: base-host, code, ID, and custom-domain precedence.
- `internal/service/domain_bind.go`: reserved-domain unbind safety.
- `internal/service/domain_bind_test.go`: ensure platform site is never deleted.
- `internal/server/server.go`: wire portal, selection, resolver, and tenant-match routes.
- `internal/server/server_test.go`: route-level portal and mismatch tests.
- `cmd/server/main.go`: construct new dependencies and run safe legacy-domain cleanup.

### PC portal files to create

- `web/pc/src/api/portal.ts`: public portal API contract.
- `web/pc/src/api/portalSession.ts`: active portal code used by Axios.
- `web/pc/src/context/PortalContext.tsx`: resolve path/host portal and expose status.
- `web/pc/src/pages/OrganizationSelectPage.tsx`: enterprise selection UI.
- `web/pc/src/pages/PortalErrorPage.tsx`: missing, suspended, and expired states.
- `web/pc/src/utils/portalRouting.ts`: pure URL and redirect helpers.
- `web/pc/tests/portalRouting.test.ts`: route and role redirect tests.

### PC portal files to modify

- `web/pc/src/main.tsx`: install `PortalProvider`.
- `web/pc/src/router.tsx`: platform, tenant-path, custom-host, and compatibility routes.
- `web/pc/src/api/auth.ts`: login-outcome and tenant-selection contracts.
- `web/pc/src/api/authSession.ts`: application-scoped token keys and legacy migration.
- `web/pc/src/api/client.ts`: attach tenant code and refresh the portal session.
- `web/pc/src/context/AuthContext.tsx`: expose pending organization selection.
- `web/pc/src/context/TenantThemeContext.tsx`: consume portal metadata.
- `web/pc/src/pages/LoginPage.tsx`: scoped/unified login and brand presentation.
- `web/pc/src/components/AppLayout.tsx`: tenant logo/name and tenant-safe navigation.
- `web/pc/src/pages/CourseDetailPage.tsx`: router navigation without dropping portal prefix.
- `web/pc/src/styles.css`: organization selector and portal status states.
- `web/pc/tests/authSession.test.ts`: key migration and tenant checks.

### Admin and H5 files to modify

- `web/admin/src/api/auth.ts`: login-outcome and `selectTenant` support.
- `web/admin/src/api/authSession.ts`: admin-scoped keys and legacy migration.
- `web/admin/src/api/client.ts`: use admin refresh key.
- `web/admin/src/pages/Login.tsx`: replace manual tenant-code entry with organization choices.
- `web/admin/src/pages/Register.tsx`: show and copy the default portal URL.
- `web/admin/src/pages/DomainSettings.tsx`: label custom domain as optional and show default portal.
- `web/admin/tests/authSession.test.ts`: scoped-key migration tests.
- `web/h5/src/api/auth.ts`: portal-scoped session contract.
- `web/h5/src/api/client.ts`: tenant header, refresh, and 401 behavior.
- `web/h5/src/context/TenantThemeContext.tsx`: resolved tenant branding.
- `web/h5/src/pages/LoginPage.tsx`: tenant-aware login and error copy.
- `web/h5/src/components/ProtectedRoute.tsx`: reliable expiration redirect.

### Routing, documentation, and collaboration files

- `docker/nginx/conf/nginx.conf`: portal SPA routes and cache policy.
- `internal/server/nginx_config_test.go`: static Nginx contract checks.
- `internal/test/integration/core_flows_test.go`: end-to-end portal and organization login.
- `README.md`: default portal, unified login, optional domain instructions.
- `DESIGN.md`: tenant resolution and login architecture.
- `docs/域名配置指南.md`: optional-domain wording and protected platform host.
- `.codex/codex-log.md`: implementation summary, review feedback, and next step.

---

### Task 1: Make Tenant Resolution Platform-Host Aware

**Files:**
- Modify: `internal/middleware/tenant.go`
- Modify: `internal/middleware/tenant_test.go`
- Create: `internal/middleware/tenant_match.go`
- Create: `internal/middleware/tenant_match_test.go`

**Interfaces:**
- Produces: `TenantWithRepositoryForPlatformHost(tenants repository.TenantRepository, platformHost string) gin.HandlerFunc`
- Produces: `TenantMatch(tenants repository.TenantRepository) gin.HandlerFunc`
- Preserves: `tenantcontext.TenantFromContext(ctx) (code string, source string)`

- [ ] **Step 1: Write failing resolver tests**

```go
func TestPlatformHostIsUnknownEvenThoughItHasThreeLabels(t *testing.T) {
	router := gin.New()
	router.Use(TenantWithRepositoryForPlatformHost(repo, "play.imai.work"))
	router.GET("/", echoTenant)
	request := httptest.NewRequest(http.MethodGet, "https://play.imai.work/", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assertTenantJSON(t, response, tenantcontext.UnknownTenant, tenantcontext.SourceUnknown)
}

func TestTenantIDHeaderResolvesToTenantCode(t *testing.T) {
	request.Header.Set("X-Tenant-ID", tenant.ID)
	assertTenantJSON(t, serve(request), tenant.Code, tenantcontext.SourceHeaderID)
}
```

- [ ] **Step 2: Run the resolver tests and verify failure**

Run:

```bash
go test ./internal/middleware -run 'TestPlatformHostIsUnknown|TestTenantIDHeaderResolves'
```

Expected: FAIL because the generic subdomain parser returns `play` and the ID header is treated as a code.

- [ ] **Step 3: Implement explicit platform-host and header-ID resolution**

```go
func tenantFromRequestWithRepository(
	request *http.Request,
	tenants repository.TenantRepository,
	platformHost string,
) (string, string) {
	host := requestHost(request.Host)
	if host != "" && host != requestHost(platformHost) {
		if tenant, err := tenants.FindByCustomDomain(request.Context(), host); err == nil {
			return tenant.Code, tenantcontext.SourceCustomDomain
		}
	}
	if code := strings.TrimSpace(request.Header.Get("X-Tenant-Code")); code != "" {
		return code, tenantcontext.SourceHeaderCode
	}
	if id := strings.TrimSpace(request.Header.Get("X-Tenant-ID")); id != "" {
		if tenant, err := tenants.FindByID(request.Context(), id); err == nil {
			return tenant.Code, tenantcontext.SourceHeaderID
		}
	}
	if host == requestHost(platformHost) {
		return tenantcontext.UnknownTenant, tenantcontext.SourceUnknown
	}
	return tenantFromRequest(request)
}
```

- [ ] **Step 4: Add failing JWT-to-portal consistency tests**

```go
func TestTenantMatchRejectsDifferentPortalTenant(t *testing.T) {
	ctx := usercontext.WithUser(context.Background(), "user-1", "tenant-a", "a@example.com", "learner")
	ctx = tenantcontext.WithTenant(ctx, "bravo", tenantcontext.SourceHeaderCode)
	response := serveTenantMatch(t, ctx, tenants)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
```

- [ ] **Step 5: Implement and verify the tenant-match guard**

```go
func TenantMatch(tenants repository.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, tokenTenantID, _, role, ok := usercontext.UserFromContext(c.Request.Context())
		code, _ := tenantcontext.TenantFromContext(c.Request.Context())
		if !ok || role == "superadmin" || code == tenantcontext.UnknownTenant {
			c.Next()
			return
		}
		portal, err := tenants.FindByCode(c.Request.Context(), code)
		if err != nil || portal.ID != tokenTenantID {
			errorsx.GinResponse(c, errorsx.Forbidden("tenant context does not match session"))
			c.Abort()
			return
		}
		c.Next()
	}
}
```

Run:

```bash
go test ./internal/middleware
```

Expected: PASS.

- [ ] **Step 6: Commit Task 1**

```bash
git add internal/middleware/tenant.go internal/middleware/tenant_test.go internal/middleware/tenant_match.go internal/middleware/tenant_match_test.go
git commit -m "fix: make tenant resolution platform aware"
```

---

### Task 2: Add Public Portal Resolution

**Files:**
- Create: `internal/service/portal.go`
- Create: `internal/service/portal_test.go`
- Create: `internal/api/portal.go`
- Create: `internal/api/portal_test.go`
- Modify: `internal/server/server.go`

**Interfaces:**
- Consumes: `repository.TenantRepository`
- Produces: `PortalService.Resolve(ctx context.Context, tenantCode, host string) (*Portal, error)`
- Produces: `GET /api/v1/portal?tenant_code={code}`

- [ ] **Step 1: Write failing portal service tests**

```go
func TestPortalResolveByCode(t *testing.T) {
	portal, err := service.Resolve(context.Background(), "acme", "play.imai.work")
	if err != nil || portal.Code != "acme" {
		t.Fatalf("portal=%#v err=%v", portal, err)
	}
	if portal.DefaultPortalURL != "https://play.imai.work/t/acme" {
		t.Fatalf("default URL=%q", portal.DefaultPortalURL)
	}
}

func TestPortalResolveRejectsSuspendedTenant(t *testing.T) {
	_, err := service.Resolve(context.Background(), "suspended", "play.imai.work")
	assertAPIErrorStatus(t, err, http.StatusForbidden)
}
```

- [ ] **Step 2: Run tests and verify missing service**

Run:

```bash
go test ./internal/service -run TestPortalResolve
```

Expected: FAIL because `PortalService` is undefined.

- [ ] **Step 3: Implement portal metadata resolution**

```go
type Portal struct {
	Code              string `json:"code"`
	Name              string `json:"name"`
	LogoURL           string `json:"logo_url"`
	PrimaryColor      string `json:"primary_color"`
	WelcomeText       string `json:"welcome_text"`
	DefaultPortalURL  string `json:"default_portal_url"`
	CustomDomainURL   string `json:"custom_domain_url,omitempty"`
}

func (s *PortalService) Resolve(ctx context.Context, code, host string) (*Portal, error) {
	var tenant *domain.Tenant
	var err error
	if strings.TrimSpace(code) != "" {
		tenant, err = s.tenants.FindByCode(ctx, strings.ToLower(strings.TrimSpace(code)))
	} else {
		tenant, err = s.tenants.FindByCustomDomain(ctx, normalizeHost(host))
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.NotFound("tenant portal not found")
	}
	if err != nil {
		return nil, errorsx.Internal("resolve tenant portal failed")
	}
	if ok, reason := TenantAccessible(tenant, time.Now().UTC()); !ok {
		return nil, errorsx.Forbidden(reason)
	}
	return portalFromTenant(tenant, s.platformHost), nil
}
```

- [ ] **Step 4: Write and implement the API handler**

```go
type PortalResolver interface {
	Resolve(context.Context, string, string) (*service.Portal, error)
}

func (h *PortalHandler) Get(c *gin.Context) {
	portal, err := h.service.Resolve(
		c.Request.Context(),
		c.Query("tenant_code"),
		c.Request.Host,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, portal)
}
```

Register `GET /api/v1/portal` without authentication or tenant middleware.

- [ ] **Step 5: Run service, API, and server tests**

Run:

```bash
go test ./internal/service ./internal/api ./internal/server
```

Expected: PASS.

- [ ] **Step 6: Commit Task 2**

```bash
git add internal/service/portal.go internal/service/portal_test.go internal/api/portal.go internal/api/portal_test.go internal/server/server.go
git commit -m "feat: add public tenant portal resolver"
```

---

### Task 3: Persist One-Time Organization Selection Challenges

**Files:**
- Create: `internal/domain/login_challenge.go`
- Create: `internal/repository/login_challenge.go`
- Create: `internal/repository/login_challenge_gorm.go`
- Create: `internal/repository/login_challenge_gorm_test.go`
- Modify: `internal/migration/migration.go`
- Modify: `internal/migration/migration_test.go`

**Interfaces:**
- Produces: `LoginChallengeRepository.Create(ctx, challenge) error`
- Produces: `LoginChallengeRepository.Consume(ctx, tokenHash string, now time.Time) (*domain.LoginChallenge, error)`
- Produces: schema migration version 13.

- [ ] **Step 1: Write the failing repository tests**

```go
func TestLoginChallengeConsumeIsSingleUse(t *testing.T) {
	challenge := &domain.LoginChallenge{
		TokenHash: "hash",
		CandidateUserIDs: `["user-a","user-b"]`,
		ExpiresAt: time.Now().UTC().Add(time.Minute),
	}
	require.NoError(t, repository.Create(ctx, challenge))
	_, err := repository.Consume(ctx, "hash", time.Now().UTC())
	require.NoError(t, err)
	_, err = repository.Consume(ctx, "hash", time.Now().UTC())
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestLoginChallengeConsumeRejectsExpired(t *testing.T) {
	now := time.Now().UTC()
	challenge := &domain.LoginChallenge{
		TokenHash:       "expired-hash",
		CandidateUserIDs: `["user-a"]`,
		ExpiresAt:       now.Add(-time.Second),
	}
	require.NoError(t, repository.Create(ctx, challenge))
	_, err := repository.Consume(ctx, "expired-hash", now)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/repository -run TestLoginChallenge
```

Expected: FAIL because the model and repository do not exist.

- [ ] **Step 3: Implement the challenge model and interface**

```go
type LoginChallenge struct {
	ID               string     `gorm:"primaryKey"`
	TokenHash        string     `gorm:"uniqueIndex;not null"`
	CandidateUserIDs string     `gorm:"type:text;not null"`
	ExpiresAt        time.Time  `gorm:"index;not null"`
	ConsumedAt       *time.Time `gorm:"index"`
	CreatedAt        time.Time
}

type LoginChallengeRepository interface {
	Create(context.Context, *domain.LoginChallenge) error
	Consume(context.Context, string, time.Time) (*domain.LoginChallenge, error)
}
```

- [ ] **Step 4: Implement atomic challenge consumption**

```go
func (r *loginChallengeGORMRepository) Consume(
	ctx context.Context,
	hash string,
	now time.Time,
) (*domain.LoginChallenge, error) {
	var challenge domain.LoginChallenge
	err := r.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(
			"token_hash = ? AND consumed_at IS NULL AND expires_at > ?",
			hash, now,
		).First(&challenge).Error; err != nil {
			return err
		}
		result := tx.Model(&domain.LoginChallenge{}).
			Where("id = ? AND consumed_at IS NULL", challenge.ID).
			Update("consumed_at", now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	return &challenge, err
}
```

- [ ] **Step 5: Add migration v13 and credential indexes**

```go
func migrateV13(database *gorm.DB) error {
	if err := database.AutoMigrate(&domain.LoginChallenge{}); err != nil {
		return err
	}
	for _, statement := range []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email_lookup ON users (LOWER(email))",
		"CREATE INDEX IF NOT EXISTS idx_users_phone_lookup ON users (phone)",
	} {
		if err := database.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 6: Run migration and repository tests**

Run:

```bash
go test ./internal/migration ./internal/repository
```

Expected: PASS.

- [ ] **Step 7: Commit Task 3**

```bash
git add internal/domain/login_challenge.go internal/repository/login_challenge.go internal/repository/login_challenge_gorm.go internal/repository/login_challenge_gorm_test.go internal/migration/migration.go internal/migration/migration_test.go
git commit -m "feat: persist organization login challenges"
```

---

### Task 4: Implement Password-First Cross-Tenant Login

**Files:**
- Modify: `internal/repository/user.go`
- Modify: `internal/repository/user_gorm.go`
- Modify: `internal/repository/user_gorm_test.go`
- Modify: `internal/service/auth.go`
- Modify: `internal/service/auth_test.go`
- Modify: `internal/api/auth.go`
- Modify: `internal/api/auth_test.go`
- Modify: `internal/server/server.go`
- Modify: `cmd/server/main.go`

**Interfaces:**
- Consumes: `LoginChallengeRepository`
- Produces: `AuthService.BeginLogin(ctx, identifier, password) (*LoginOutcome, error)`;
  an explicit tenant in `ctx` stays tenant-scoped, while platform context performs
  password-first organization discovery.
- Produces: `AuthService.SelectTenant(ctx, selectionToken, tenantCode) (*LoginOutcome, error)`
- Produces: `POST /api/v1/auth/select-tenant`

- [ ] **Step 1: Expand repository tests to all roles**

```go
func TestFindByCredentialAcrossTenantsReturnsAllRoles(t *testing.T) {
	createUser(t, "tenant-a", "shared@example.com", "tenant_admin")
	createUser(t, "tenant-b", "shared@example.com", "instructor")
	createUser(t, "tenant-c", "shared@example.com", "learner")
	users, err := repository.FindByCredentialAcrossTenants(ctx, "SHARED@example.com")
	if err != nil || len(users) != 3 {
		t.Fatalf("users=%#v err=%v", users, err)
	}
}
```

- [ ] **Step 2: Remove the role filter and run repository tests**

```go
query := repository.database.WithContext(ctx).
	Where("tenant_id IS NOT NULL")
```

Run:

```bash
go test ./internal/repository -run 'TestFindByCredentialAcrossTenants'
```

Expected: PASS.

- [ ] **Step 3: Write failing service tests for login outcomes**

```go
func TestBeginLoginWrongPasswordDoesNotRevealOrganizations(t *testing.T) {
	fixture := newAuthLoginFixture(t)
	fixture.addTenantUser("acme", "shared@example.com", "password-a", "learner", 1)
	fixture.addTenantUser("bravo", "shared@example.com", "password-b", "instructor", 1)

	outcome, err := fixture.service.BeginLogin(
		context.Background(), "shared@example.com", "wrong-password",
	)
	require.Nil(t, outcome)
	require.Equal(t, http.StatusUnauthorized, appErrorStatus(err))
	require.Empty(t, fixture.challengeRepository.created)
}

func TestBeginLoginMultipleMatchingPasswordsReturnsOrganizations(t *testing.T) {
	fixture := newAuthLoginFixture(t)
	fixture.addTenantUser("acme", "shared@example.com", "same-password", "learner", 1)
	fixture.addTenantUser("bravo", "shared@example.com", "same-password", "instructor", 1)

	outcome, err := fixture.service.BeginLogin(
		context.Background(), "SHARED@example.com", "same-password",
	)
	require.NoError(t, err)
	require.True(t, outcome.RequiresTenantSelection)
	require.Len(t, outcome.Organizations, 2)
	require.NotEmpty(t, outcome.SelectionToken)
	require.Nil(t, outcome.Pair)
}
```

Add the fixture helper in `auth_test.go`, backed by the existing in-memory repositories,
then add concrete cases for:

- one learner match returns a token whose claims contain that tenant ID;
- an explicit portal context only searches that tenant;
- platform superadmin login still works and never produces organization choices;
- disabled users and suspended/expired tenants are excluded;
- expired, replayed, tampered, and unlisted selections return 401;
- refresh preserves the selected tenant ID.

Each case must assert HTTP-equivalent error status, role, selected `tenant_id`, and that no
organization list or login challenge is created for a wrong password.

- [ ] **Step 4: Add explicit login outcome types**

```go
type OrganizationOption struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	LogoURL string `json:"logo_url,omitempty"`
	Role    string `json:"role"`
}

type LoginOutcome struct {
	User                    *domain.User
	Tenant                  *Portal
	Pair                    *TokenPair
	RequiresTenantSelection bool
	SelectionToken          string
	Organizations           []OrganizationOption
}
```

`BeginLogin` first follows the existing tenant-scoped path when the request context resolves
to a tenant. In platform context it checks the tenantless superadmin record, then calls
`matchingTenantUsers`; this preserves superadmin access without exposing tenant candidates.

- [ ] **Step 5: Implement password-first candidate filtering**

```go
func (s *AuthService) matchingTenantUsers(
	ctx context.Context,
	identifier string,
	password string,
) ([]domain.User, []OrganizationOption, error) {
	candidates, err := s.users.FindByCredentialAcrossTenants(ctx, identifier)
	if err != nil {
		return nil, nil, errorsx.Internal("find user failed")
	}
	matches := make([]domain.User, 0, len(candidates))
	options := make([]OrganizationOption, 0, len(candidates))
	for _, candidate := range candidates {
		if !security.CheckPassword(password, candidate.Password) || candidate.Status != 1 {
			continue
		}
		tenant, err := s.tenants.FindByID(ctx, candidate.TenantID)
		if err != nil {
			continue
		}
		if ok, _ := TenantAccessible(tenant, time.Now().UTC()); !ok {
			continue
		}
		matches = append(matches, candidate)
		options = append(options, organizationOption(tenant, candidate.Role))
	}
	return matches, options, nil
}
```

- [ ] **Step 6: Implement challenge creation and selection**

```go
plain, hash, err := security.GenerateRefreshToken()
candidateJSON, err := json.Marshal(candidateIDs(matches))
challenge := &domain.LoginChallenge{
	TokenHash: hash,
	CandidateUserIDs: string(candidateJSON),
	ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
}
```

`SelectTenant` must consume the hash, decode candidate IDs, fetch the selected user by ID,
verify its tenant code and current status, then call `issueTokens`.

- [ ] **Step 7: Update API response contracts**

```go
func (handler *AuthHandler) Login(c *gin.Context) {
	outcome, err := handler.service.BeginLogin(c.Request.Context(), identifier, request.Password)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, loginOutcomeResponse(outcome))
}

func (handler *AuthHandler) SelectTenant(c *gin.Context) {
	var request struct {
		SelectionToken string `json:"selection_token" binding:"required"`
		TenantCode     string `json:"tenant_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorsx.GinResponse(c, errorsx.BadRequest("invalid request"))
		return
	}
	outcome, err := handler.service.SelectTenant(
		c.Request.Context(), request.SelectionToken, request.TenantCode,
	)
	if err != nil {
		errorsx.GinResponse(c, err)
		return
	}
	success(c, loginOutcomeResponse(outcome))
}
```

- [ ] **Step 8: Wire the repository and route**

```go
loginChallenges := repository.NewLoginChallengeRepository(database)
authService.SetLoginChallengeRepository(loginChallenges)
auth.POST("/select-tenant", limiter.Handler(), authHandler.SelectTenant)
```

- [ ] **Step 9: Run auth and server tests**

Run:

```bash
go test ./internal/repository ./internal/service ./internal/api ./internal/server
```

Expected: PASS.

- [ ] **Step 10: Commit Task 4**

```bash
git add internal/repository/user.go internal/repository/user_gorm.go internal/repository/user_gorm_test.go internal/service/auth.go internal/service/auth_test.go internal/api/auth.go internal/api/auth_test.go internal/server/server.go cmd/server/main.go
git commit -m "feat: add unified organization-aware login"
```

---

### Task 5: Repair Reserved-Domain Bindings Safely

**Files:**
- Create: `internal/migration/reserved_domain.go`
- Create: `internal/migration/reserved_domain_test.go`
- Modify: `internal/service/domain_bind.go`
- Modify: `internal/service/domain_bind_test.go`
- Modify: `cmd/server/main.go`

**Interfaces:**
- Produces: `migration.ClearReservedTenantDomain(db *gorm.DB, reserved string) (int64, error)`
- Changes: `DomainBindService.Unbind` skips `DeleteSite` when persisted domain equals `ReservedDomain`.

- [ ] **Step 1: Write failing cleanup tests**

```go
func TestClearReservedTenantDomainOnlyClearsMatchingTenant(t *testing.T) {
	createTenant(t, "bad", ptr("play.imai.work"))
	createTenant(t, "good", ptr("learn.example.com"))
	affected, err := ClearReservedTenantDomain(db, "play.imai.work")
	if err != nil || affected != 1 {
		t.Fatalf("affected=%d err=%v", affected, err)
	}
	assertTenantDomain(t, "bad", nil)
	assertTenantDomain(t, "good", ptr("learn.example.com"))
}

func TestUnbindReservedDomainDoesNotDeletePanelSite(t *testing.T) {
	_, err := service.Unbind(ctx)
	if err != nil || panel.deleteCalls != 0 {
		t.Fatalf("err=%v deleteCalls=%d", err, panel.deleteCalls)
	}
}
```

- [ ] **Step 2: Run tests and verify current dangerous behavior**

Run:

```bash
go test ./internal/migration ./internal/service -run 'ReservedDomain|UnbindReserved'
```

Expected: FAIL because cleanup is missing and Unbind calls `DeleteSite`.

- [ ] **Step 3: Implement database-only cleanup**

```go
func ClearReservedTenantDomain(database *gorm.DB, reserved string) (int64, error) {
	reserved = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(reserved), "."))
	if reserved == "" {
		return 0, nil
	}
	result := database.Model(&domain.Tenant{}).
		Where("LOWER(custom_domain) = ?", reserved).
		Update("custom_domain", nil)
	return result.RowsAffected, result.Error
}
```

- [ ] **Step 4: Guard the Unbind panel deletion**

```go
reserved := normalizeDomain(service.config.ReservedDomain)
isReserved := normalizeDomain(domainName) == reserved
if !isReserved && (persisted || service.ownsDomain(actor.tenantID, domainName)) {
	if err := service.panel.DeleteSite(domainName); err != nil && !missingSiteError(err) {
		return service.status(actor.tenantID), errorsx.Internal("delete baota site failed")
	}
}
```

- [ ] **Step 5: Invoke cleanup after migrations**

```go
affected, err := migration.ClearReservedTenantDomain(database, cfg.AdminHost)
if err != nil {
	log.Fatalf("repair reserved tenant domain: %v", err)
}
if affected > 0 {
	slog.Warn("cleared reserved tenant domain bindings", "count", affected, "domain", cfg.AdminHost)
}
```

- [ ] **Step 6: Run focused and full backend tests**

Run:

```bash
go test ./internal/migration ./internal/service
go test ./...
```

Expected: PASS.

- [ ] **Step 7: Commit Task 5**

```bash
git add internal/migration/reserved_domain.go internal/migration/reserved_domain_test.go internal/service/domain_bind.go internal/service/domain_bind_test.go cmd/server/main.go
git commit -m "fix: protect platform host during domain cleanup"
```

---

### Task 6: Build the PC Portal Routing and Context Foundation

**Files:**
- Create: `web/pc/src/api/portal.ts`
- Create: `web/pc/src/api/portalSession.ts`
- Create: `web/pc/src/context/PortalContext.tsx`
- Create: `web/pc/src/pages/PortalErrorPage.tsx`
- Create: `web/pc/src/utils/portalRouting.ts`
- Create: `web/pc/tests/portalRouting.test.ts`
- Modify: `web/pc/src/main.tsx`
- Modify: `web/pc/src/router.tsx`
- Modify: `web/pc/src/api/client.ts`
- Modify: `web/pc/src/context/TenantThemeContext.tsx`
- Modify: `web/pc/src/components/AppLayout.tsx`
- Modify: `web/pc/src/pages/CourseDetailPage.tsx`

**Interfaces:**
- Consumes: `GET /api/v1/portal`
- Produces: `PortalContextValue { portal, tenantCode, mode, loading, error }`
- Produces: `portalPath(tenantCode, childPath) string`
- Produces: Axios `X-Tenant-Code` injection for `/t/:tenantCode`.

- [ ] **Step 1: Write failing pure routing tests**

```ts
test('extracts tenant code from default portal path', () => {
  assert.equal(tenantCodeFromPath('/t/acme/courses'), 'acme')
})

test('builds tenant-safe course path', () => {
  assert.equal(portalPath('acme', '/courses/42'), '/t/acme/courses/42')
})

test('does not infer tenant from platform login', () => {
  assert.equal(tenantCodeFromPath('/login'), undefined)
})
```

- [ ] **Step 2: Run tests and verify missing helpers**

Run:

```bash
cd web/pc && node --test tests/portalRouting.test.ts
```

Expected: FAIL because `portalRouting.ts` does not exist.

- [ ] **Step 3: Implement route helpers and active portal session**

```ts
export function tenantCodeFromPath(pathname: string): string | undefined {
  const match = pathname.match(/^\/t\/([^/]+)(?:\/|$)/)
  return match ? decodeURIComponent(match[1]).toLowerCase() : undefined
}

export function portalPath(code: string, child = '/'): string {
  const suffix = child === '/' ? '' : `/${child.replace(/^\/+/, '')}`
  return `/t/${encodeURIComponent(code)}${suffix}`
}
```

`portalSession.ts` keeps the active code in module state and `sessionStorage`, with
`setActivePortalCode`, `getActivePortalCode`, and `clearActivePortalCode`.

- [ ] **Step 4: Implement portal API and provider**

```ts
export interface Portal {
  code: string
  name: string
  logo_url: string
  primary_color: string
  welcome_text: string
  default_portal_url: string
  custom_domain_url?: string
}

export async function resolvePortal(code?: string): Promise<Portal> {
  const response = await apiClient.get<Portal>(
    '/api/v1/portal',
    code ? { params: { tenant_code: code } } : undefined,
  )
  return response.data
}
```

The provider must resolve by path code first and by Host when running outside
`play.imai.work`; it must not silently replace 404/403 with platform defaults.

- [ ] **Step 5: Replace basename routing with explicit portal routes**

```tsx
export const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  { path: '/select-organization', element: <OrganizationSelectPage /> },
  { path: '/t/:tenantCode/login', element: <LoginPage /> },
  {
    path: '/t/:tenantCode',
    element: <ProtectedRoute />,
    children: portalChildren,
  },
  { path: '/pc/login', element: <Navigate to="/login" replace /> },
  { path: '/pc/*', element: <LegacyPortalRedirect /> },
  { path: '/', element: <CustomDomainOrPlatformEntry /> },
])
```

Keep Vite `base: '/pc/'`; it controls asset URLs, not the browser route table.

- [ ] **Step 6: Attach the portal code to API requests**

```ts
apiClient.interceptors.request.use((config) => {
  const tenantCode = getActivePortalCode()
  if (tenantCode) config.headers['X-Tenant-Code'] = tenantCode
  const token = readPortalAccessToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})
```

- [ ] **Step 7: Make navigation preserve the portal prefix**

Replace hardcoded `window.location.assign('/courses/...')` and root-relative links with
React Router navigation generated by `portalPath(portal.code, childPath)`.

- [ ] **Step 8: Run PC tests and build**

Run:

```bash
cd web/pc
node --test tests/*.test.ts
npm run build
```

Expected: all tests and TypeScript/Vite build pass.

- [ ] **Step 9: Commit Task 6**

```bash
git add web/pc/src web/pc/tests
git commit -m "feat: add default tenant portal routing"
```

---

### Task 7: Add Unified Login and Organization Selection to PC

**Files:**
- Create: `web/pc/src/pages/OrganizationSelectPage.tsx`
- Modify: `web/pc/src/api/auth.ts`
- Modify: `web/pc/src/api/authSession.ts`
- Modify: `web/pc/src/context/AuthContext.tsx`
- Modify: `web/pc/src/pages/LoginPage.tsx`
- Modify: `web/pc/src/styles.css`
- Modify: `web/pc/tests/authSession.test.ts`
- Modify: `web/pc/tests/portalRouting.test.ts`

**Interfaces:**
- Consumes: `POST /api/v1/auth/login`
- Consumes: `POST /api/v1/auth/select-tenant`
- Produces: role-aware redirect to `/t/{code}` or `/admin/`.

- [ ] **Step 1: Write failing session namespace tests**

```ts
test('migrates legacy learner token to portal key', () => {
  storage.setItem('imaiplay_token', learnerToken)
  migrateLegacySession(storage)
  assert.equal(storage.getItem(PORTAL_ACCESS_TOKEN_KEY), learnerToken)
  assert.equal(storage.getItem('imaiplay_token'), null)
})

test('does not migrate admin token into portal session', () => {
  storage.setItem('imaiplay_token', tenantAdminToken)
  migrateLegacySession(storage)
  assert.equal(storage.getItem(PORTAL_ACCESS_TOKEN_KEY), null)
})
```

- [ ] **Step 2: Implement scoped keys and decoded session metadata**

```ts
export const PORTAL_ACCESS_TOKEN_KEY = 'imaiplay_portal_access_token'
export const PORTAL_REFRESH_TOKEN_KEY = 'imaiplay_portal_refresh_token'

export interface SessionClaims {
  user_id: string
  tenant_id: string
  role: 'learner' | 'instructor' | 'tenant_admin' | 'superadmin'
  exp: number
}
```

- [ ] **Step 3: Add login and selection response types**

```ts
export interface OrganizationOption {
  code: string
  name: string
  logo_url?: string
  role: string
}

export type LoginResult =
  | { requires_tenant_selection: true; selection_token: string; organizations: OrganizationOption[] }
  | { requires_tenant_selection?: false; token: string; refresh_token?: string; user: AuthUser; tenant: Portal }
```

- [ ] **Step 4: Implement role-aware session storage**

```ts
export function persistLogin(result: AuthenticatedLoginResult): string {
  const claims = decodeClaims(result.token)
  if (claims.role === 'learner') {
    writePortalSession(result)
    return portalPath(result.tenant.code)
  }
  writeAdminSession(result)
  return '/admin/'
}
```

- [ ] **Step 5: Implement organization selection UI**

The page must read a non-persisted selection state from `AuthContext`, render logo, name,
and role for each option, disable double submission, and call:

```ts
await selectTenant({
  selection_token: pending.selectionToken,
  tenant_code: organization.code,
})
```

Direct navigation without a pending selection redirects to `/login`.

- [ ] **Step 6: Apply tenant branding to login and portal chrome**

Use `PortalContext` for logo, tenant name, welcome text, primary color, title, and status.
The platform `/login` remains generic until authentication identifies an organization.

- [ ] **Step 7: Run PC tests and build**

Run:

```bash
cd web/pc
node --test tests/*.test.ts
npm run build
```

Expected: PASS.

- [ ] **Step 8: Commit Task 7**

```bash
git add web/pc/src web/pc/tests
git commit -m "feat: add organization-aware portal login"
```

---

### Task 8: Update Admin Login, Session Keys, and Tenant Portal UX

**Files:**
- Modify: `web/admin/src/api/auth.ts`
- Modify: `web/admin/src/api/authSession.ts`
- Modify: `web/admin/src/api/client.ts`
- Modify: `web/admin/src/pages/Login.tsx`
- Modify: `web/admin/src/pages/Register.tsx`
- Modify: `web/admin/src/pages/DomainSettings.tsx`
- Modify: `web/admin/src/styles.css`
- Modify: `web/admin/tests/authSession.test.ts`

**Interfaces:**
- Consumes: unified login outcome and `POST /api/v1/auth/select-tenant`.
- Produces: `ADMIN_ACCESS_TOKEN_KEY` and `ADMIN_REFRESH_TOKEN_KEY`.
- Produces: visible/copyable `https://play.imai.work/t/{tenantCode}`.

- [ ] **Step 1: Write failing admin session migration tests**

```ts
test('migrates legacy staff token into admin key', () => {
  storage.setItem('imaiplay_token', tenantAdminToken)
  migrateLegacyAdminSession(storage)
  assert.equal(storage.getItem(ADMIN_ACCESS_TOKEN_KEY), tenantAdminToken)
})

test('leaves learner token for portal migration', () => {
  storage.setItem('imaiplay_token', learnerToken)
  migrateLegacyAdminSession(storage)
  assert.equal(storage.getItem(ADMIN_ACCESS_TOKEN_KEY), null)
})
```

- [ ] **Step 2: Implement scoped admin session keys**

```ts
export const ADMIN_ACCESS_TOKEN_KEY = 'imaiplay_admin_access_token'
export const ADMIN_REFRESH_TOKEN_KEY = 'imaiplay_admin_refresh_token'
```

Update the Axios request and refresh interceptors to use only these keys.

- [ ] **Step 3: Replace manual tenant-code fallback with organization cards**

```tsx
{pendingSelection ? (
  <OrganizationPicker
    organizations={pendingSelection.organizations}
    onSelect={handleOrganizationSelect}
  />
) : (
  <LoginForm onFinish={submit} />
)}
```

The admin page must reject a selected learner role with “请前往学习门户登录”, while
tenant_admin and instructor sessions continue to `/admin/`.

- [ ] **Step 4: Show the permanent default portal after registration**

```tsx
const portalURL = `${window.location.origin}/t/${tenant.code}`
<Input value={portalURL} readOnly addonAfter={<CopyButton value={portalURL} />} />
```

- [ ] **Step 5: Reframe custom domains as optional**

At the top of `DomainSettings`, render the default portal as “立即可用”. Rename the custom
domain card to “自定义品牌域名（可选）” and retain the existing verify/bind flow.

- [ ] **Step 6: Run Admin tests and build**

Run:

```bash
cd web/admin
node --test tests/*.test.ts
npm run build
```

Expected: PASS.

- [ ] **Step 7: Commit Task 8**

```bash
git add web/admin/src web/admin/tests
git commit -m "feat: simplify tenant portal onboarding"
```

---

### Task 9: Align H5 Authentication and Branding

**Files:**
- Modify: `web/h5/src/api/auth.ts`
- Modify: `web/h5/src/api/client.ts`
- Modify: `web/h5/src/api/theme.ts`
- Modify: `web/h5/src/context/TenantThemeContext.tsx`
- Modify: `web/h5/src/pages/LoginPage.tsx`
- Modify: `web/h5/src/components/ProtectedRoute.tsx`
- Create: `web/h5/src/api/authSession.ts`
- Create: `web/h5/tests/authSession.test.ts`

**Interfaces:**
- Consumes: portal-scoped token keys and `X-Tenant-Code`.
- Produces: the same expiry, refresh, logout, and role checks as PC.

- [ ] **Step 1: Write failing H5 session tests**

```ts
test('rejects expired portal token', () => {
  assert.equal(isValidPortalSession(expiredLearnerToken), false)
})

test('rejects tenant mismatch on an explicit portal', () => {
  assert.equal(isValidPortalSession(acmeToken, 'bravo-id'), false)
})
```

- [ ] **Step 2: Implement H5 auth session helpers**

Reuse the exact portal key strings and claim validation rules from the PC contract:

```ts
export const PORTAL_ACCESS_TOKEN_KEY = 'imaiplay_portal_access_token'
export const PORTAL_REFRESH_TOKEN_KEY = 'imaiplay_portal_refresh_token'
```

- [ ] **Step 3: Implement refresh and reliable 401 logout**

Use a single-flight refresh promise. On refresh failure, clear both portal keys, dispatch
`imaiplay:portal-session-expired`, and redirect to the current tenant-aware login route.

- [ ] **Step 4: Load and render resolved portal branding**

The H5 login page must use resolved logo, tenant name, welcome text, and primary color.
The forgot-password link must remain within a tenant-capable route rather than pointing to
the custom-domain-blocked `/admin/forgot-password`.

- [ ] **Step 5: Run H5 tests and build**

Run:

```bash
cd web/h5
node --test tests/*.test.ts
npm run build
```

Expected: PASS.

- [ ] **Step 6: Commit Task 9**

```bash
git add web/h5/src web/h5/tests
git commit -m "fix: align h5 tenant sessions and branding"
```

---

### Task 10: Serve Portal Routes and Prevent Stale SPA Entrypoints

**Files:**
- Modify: `docker/nginx/conf/nginx.conf`
- Modify: `internal/server/nginx_config_test.go`

**Interfaces:**
- Consumes: PC assets under `/usr/share/nginx/html/pc`.
- Produces: `/login`, `/select-organization`, `/t/*`, and custom-domain `/` SPA fallback.

- [ ] **Step 1: Write failing Nginx contract tests**

```go
func TestNginxServesTenantPortalRoutesFromPCApp(t *testing.T) {
	config := readNginxConfig(t)
	for _, fragment := range []string{
		"location = /login",
		"location = /select-organization",
		"location ^~ /t/",
		"/usr/share/nginx/html/pc/index.html",
	} {
		if !strings.Contains(config, fragment) {
			t.Fatalf("missing %q", fragment)
		}
	}
}

```

Implement the second test as:

```go
func TestNginxDisablesCacheForEverySPAEntry(t *testing.T) {
	config := readNginxConfig(t)
	for _, entry := range []string{
		"/admin/index.html",
		"/pc/index.html",
		"/h5/index.html",
		"location = /login",
		"location = /select-organization",
	} {
		block := nginxLocationBlock(t, config, entry)
		if !strings.Contains(block, `Cache-Control "no-cache, no-store, must-revalidate"`) {
			t.Fatalf("%s does not disable cache:\n%s", entry, block)
		}
	}
}
```

- [ ] **Step 2: Run tests and verify missing routes**

Run:

```bash
go test ./internal/server -run Nginx
```

Expected: FAIL.

- [ ] **Step 3: Add explicit PC entry locations**

```nginx
location = /login {
    alias /usr/share/nginx/html/pc/index.html;
    add_header Cache-Control "no-cache, no-store, must-revalidate" always;
    expires -1;
}
location = /select-organization {
    alias /usr/share/nginx/html/pc/index.html;
    add_header Cache-Control "no-cache, no-store, must-revalidate" always;
    expires -1;
}
location ^~ /t/ {
    try_files $uri /pc/index.html;
    add_header Cache-Control "no-cache, no-store, must-revalidate" always;
}
```

Keep `/api/`, `/backend/`, `/admin/`, `/pc/`, and `/h5/` ordering ahead of the generic
custom-domain root fallback.

- [ ] **Step 4: Add immutable asset caching and entry revalidation**

```nginx
location ~* ^/(admin|pc|h5)/assets/.*\.[a-z0-9]+$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

Add no-cache handling for `/pc/index.html` and `/h5/index.html`, matching the existing
Admin entry policy.

- [ ] **Step 5: Run Nginx contract tests**

Run:

```bash
go test ./internal/server -run Nginx
```

Expected: PASS.

- [ ] **Step 6: Commit Task 10**

```bash
git add docker/nginx/conf/nginx.conf internal/server/nginx_config_test.go
git commit -m "fix: serve tenant portal routes safely"
```

---

### Task 11: Add End-to-End Tenant Portal Integration Coverage

**Files:**
- Modify: `internal/test/integration/core_flows_test.go`
- Modify: `internal/server/server_test.go`

**Interfaces:**
- Consumes: portal, login, selection, JWT, and student APIs from Tasks 1-10.
- Produces: regression coverage for tenant isolation and complete login flows.

- [ ] **Step 1: Add a single-tenant default portal flow**

```go
func TestDefaultPortalLearnerLoginAndDashboardFlow(t *testing.T) {
	fx := newFixture(t)
	tenant, learner := fx.seedTenantUser(
		t, "acme", "learner@acme.test", "password123", "learner",
	)
	portal := fx.requestWithHost(
		http.MethodGet, "/api/v1/portal?tenant_code=acme", nil, "play.imai.work",
	)
	requireStatus(t, portal, http.StatusOK)
	login := fx.requestWithTenant(
		http.MethodPost, "/api/v1/auth/login",
		map[string]any{"identifier": learner.Email, "password": "password123"},
		"", tenant.Code,
	)
	token := responseToken(t, login)
	claims := requireTokenClaims(t, token, integrationSecret)
	if claims.TenantID != tenant.ID {
		t.Fatalf("tenant_id=%q want=%q", claims.TenantID, tenant.ID)
	}
	courses := fx.requestWithTenant(
		http.MethodGet, "/api/v1/courses", nil, token, tenant.Code,
	)
	requireStatus(t, courses, http.StatusOK)
	requireOnlyTenantOrOfficialCourses(t, courses, tenant.ID)
}
```

- [ ] **Step 2: Add a multi-tenant selection flow**

```go
func TestPlatformLoginSelectsOrganizationWithoutLeakingOnWrongPassword(t *testing.T) {
	fx := newFixture(t)
	acme, _ := fx.seedTenantUser(t, "acme", "shared@test", "same-pass", "learner")
	fx.seedTenantUser(t, "bravo", "shared@test", "same-pass", "instructor")

	wrong := fx.requestWithHost(
		http.MethodPost, "/api/v1/auth/login",
		map[string]any{"identifier": "shared@test", "password": "wrong"},
		"play.imai.work",
	)
	requireStatus(t, wrong, http.StatusUnauthorized)
	requireJSONFieldAbsent(t, wrong, "data.organizations")

	login := fx.requestWithHost(
		http.MethodPost, "/api/v1/auth/login",
		map[string]any{"identifier": "shared@test", "password": "same-pass"},
		"play.imai.work",
	)
	selectionToken := requireSelectionResponse(t, login, 2)
	selected := fx.requestWithHost(
		http.MethodPost, "/api/v1/auth/select-tenant",
		map[string]any{"selection_token": selectionToken, "tenant_code": "acme"},
		"play.imai.work",
	)
	token := responseToken(t, selected)
	if claims := requireTokenClaims(t, token, integrationSecret); claims.TenantID != acme.ID {
		t.Fatalf("tenant_id=%q want=%q", claims.TenantID, acme.ID)
	}
	replay := fx.requestWithHost(
		http.MethodPost, "/api/v1/auth/select-tenant",
		map[string]any{"selection_token": selectionToken, "tenant_code": "acme"},
		"play.imai.work",
	)
	requireStatus(t, replay, http.StatusUnauthorized)
}
```

- [ ] **Step 3: Add portal mismatch and custom-domain alias flow**

```go
func TestCustomDomainAndDefaultPortalShareTenantButRejectForeignToken(t *testing.T) {
	fx := newFixture(t)
	acme := fx.seedTenant(t, "acme", ptr("learn.acme.test"))
	bravo := fx.seedTenant(t, "bravo", nil)
	acmeToken := fx.learnerToken(t, acme)
	bravoToken := fx.learnerToken(t, bravo)

	byCode := fx.requestWithHost(
		http.MethodGet, "/api/v1/portal?tenant_code=acme", nil, "play.imai.work",
	)
	byDomain := fx.requestWithHost(
		http.MethodGet, "/api/v1/portal", nil, "learn.acme.test",
	)
	requireSamePortal(t, byCode, byDomain)

	foreign := fx.requestWithTenant(
		http.MethodGet, "/api/v1/courses", nil, bravoToken, acme.Code,
	)
	requireStatus(t, foreign, http.StatusForbidden)
	own := fx.requestWithTenant(
		http.MethodGet, "/api/v1/courses", nil, acmeToken, acme.Code,
	)
	requireStatus(t, own, http.StatusOK)
}
```

Add the named fixture helpers next to the existing request/decode helpers. They must create
real GORM records and send requests through the real Gin engine; do not replace the
integration assertions with service mocks.

- [ ] **Step 4: Run integration and full backend tests**

Run:

```bash
go test ./internal/test/integration/... ./internal/server/...
go test ./...
```

Expected: PASS.

- [ ] **Step 5: Commit Task 11**

```bash
git add internal/test/integration/core_flows_test.go internal/server/server_test.go
git commit -m "test: cover tenant portal login flows"
```

---

### Task 12: Update Documentation, Collaboration Log, and Final Verification

**Files:**
- Modify: `README.md`
- Modify: `DESIGN.md`
- Modify: `docs/域名配置指南.md`
- Modify: `.codex/codex-log.md`

**Interfaces:**
- Documents the shipped URLs, login behavior, data isolation, and optional custom domain.

- [ ] **Step 1: Update user-facing documentation**

Document these exact entry points:

```text
统一登录：https://play.imai.work/login
默认门户：https://play.imai.work/t/{tenantCode}
管理后台：https://play.imai.work/admin/
自定义域名：可选，绑定后与默认门户同时有效
```

Remove wording that implies a custom domain is required for a branded portal.

- [ ] **Step 2: Update the architecture design**

Add the final resolver order, JWT tenant consistency rule, organization-selection
challenge, and the explicit decision to defer `accounts + tenant_memberships`.

- [ ] **Step 3: Append the required collaboration log entry**

Append concise sections:

```markdown
### 租户默认门户与统一登录
- 任務執行摘要：完成默认租户门户、统一登录、多企业选择和可选自定义域名。
- 關鍵修改：新增 Portal Resolver、一次性选择凭证、租户一致性校验与三端会话隔离。
- 評審反饋：平台主域名不得作为租户域名；密码验证成功前不得返回企业信息。
- 下一步建議：观察登录选择率与旧链接迁移情况，再评估全局账号模型。
```

Keep `.codex/codex-log.md` concise and do not rewrite prior entries.

- [ ] **Step 4: Run fresh final verification**

Run:

```bash
go test ./...
go build -o bin/imaiplay ./cmd/server
(cd web/admin && node --test tests/*.test.ts && npm run build)
(cd web/pc && node --test tests/*.test.ts && npm run build)
(cd web/h5 && node --test tests/*.test.ts && npm run build)
git diff --check
git status --short
```

Expected:

- All Go tests pass.
- The server binary builds.
- Admin, PC, and H5 tests/builds pass.
- `git diff --check` prints no errors.
- Only intended changes plus the pre-existing untracked `server` are present.

- [ ] **Step 5: Commit documentation and final fixes**

```bash
git add README.md DESIGN.md docs/域名配置指南.md .codex/codex-log.md
git commit -m "docs: document tenant portals and unified login"
```

- [ ] **Step 6: Review the complete branch diff**

Run:

```bash
git diff --stat main...HEAD
git log --oneline main..HEAD
```

Confirm that the branch contains the approved design, implementation, tests, and docs,
and does not include the unrelated `server` file.
