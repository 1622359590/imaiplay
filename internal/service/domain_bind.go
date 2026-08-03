package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/middleware"
	"github.com/1622359590/imaiplay/internal/repository"
	"gorm.io/gorm"
)

const (
	DomainStateNone                = "none"
	DomainStatePendingVerification = "pending_verification"
	DomainStateVerified            = "verified"
	DomainStateCreatingSite        = "creating_site"
	DomainStateConfiguring         = "configuring"
	DomainStateReady               = "ready"
	DomainStateVerificationFailed  = "verification_failed"
	DomainStateSetupFailed         = "setup_failed"
)

const disableAdminSnippet = `# ImaiPlay tenant admin guard
location ^~ /admin {
    return 404;
}`

type DomainPanel interface {
	AddSite(domain string) (int, error)
	AddReverseProxy(domain, proxyName, target string) error
	ApplyLetsEncrypt(domain string) error
	AddNginxSnippet(domain, snippet string) error
	DeleteSite(domain string) error
	GetSiteInfo(domain string) (map[string]interface{}, error)
}

type DomainResolver interface {
	LookupCNAME(host string) (string, error)
	LookupIP(host string) ([]net.IP, error)
}

type DomainAuditRecorder interface {
	Record(context.Context, domain.AuditEvent) error
}

type DomainBindConfig struct {
	ExpectedIP     string
	ReservedDomain string
	CNAMETarget    string
	ProxyTarget    string
	PollInterval   time.Duration
	MaxPolls       int
}

type DomainBindStatus struct {
	State            string    `json:"state"`
	Domain           string    `json:"domain,omitempty"`
	Message          string    `json:"message,omitempty"`
	CurrentStep      int       `json:"current_step"`
	TotalSteps       int       `json:"total_steps"`
	CNAMETarget      string    `json:"cname_target"`
	TenantCode       string    `json:"tenant_code,omitempty"`
	DefaultPortalURL string    `json:"default_portal_url,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
	History          []string  `json:"-"`
}

type DomainBindService struct {
	tenants  repository.TenantRepository
	panel    DomainPanel
	resolver DomainResolver
	audit    DomainAuditRecorder
	config   DomainBindConfig

	mu       sync.RWMutex
	statuses map[string]DomainBindStatus
	owners   map[string]string
}

type netDomainResolver struct{}

func (netDomainResolver) LookupCNAME(host string) (string, error) {
	return net.LookupCNAME(host)
}

func (netDomainResolver) LookupIP(host string) ([]net.IP, error) {
	return net.LookupIP(host)
}

func NewDomainBindService(
	tenants repository.TenantRepository,
	panel DomainPanel,
	resolver DomainResolver,
	audit DomainAuditRecorder,
	config DomainBindConfig,
) *DomainBindService {
	if resolver == nil {
		resolver = netDomainResolver{}
	}
	if config.ReservedDomain == "" {
		config.ReservedDomain = "play.imai.work"
	}
	if config.CNAMETarget == "" {
		config.CNAMETarget = config.ReservedDomain
	}
	if config.ProxyTarget == "" {
		config.ProxyTarget = "http://127.0.0.1:18080"
	}
	if config.PollInterval <= 0 {
		config.PollInterval = 5 * time.Second
	}
	if config.MaxPolls <= 0 {
		config.MaxPolls = 24
	}
	return &DomainBindService{
		tenants: tenants, panel: panel, resolver: resolver, audit: audit,
		config: config, statuses: make(map[string]DomainBindStatus),
		owners: make(map[string]string),
	}
}

func VerifyDomain(domainName, expectedIP string) (bool, error) {
	return verifyDomainWithResolver(domainName, expectedIP, netDomainResolver{})
}

func verifyDomainWithResolver(
	domainName, expectedIP string,
	resolver DomainResolver,
) (bool, error) {
	if resolver == nil || net.ParseIP(strings.TrimSpace(expectedIP)) == nil {
		return false, errors.New("域名绑定服务尚未配置服务器公网 IP")
	}
	expected := net.ParseIP(strings.TrimSpace(expectedIP))
	expectedIP = expected.String()
	if cname, err := resolver.LookupCNAME(domainName); err == nil {
		cname = strings.TrimSuffix(strings.TrimSpace(cname), ".")
		if cnameIP := net.ParseIP(cname); cnameIP != nil && cnameIP.Equal(expected) {
			return true, nil
		}
		if ips, lookupErr := resolver.LookupIP(cname); lookupErr == nil && containsIP(ips, expected) {
			return true, nil
		}
	}
	if ips, err := resolver.LookupIP(domainName); err == nil && containsIP(ips, expected) {
		return true, nil
	}
	return false, fmt.Errorf(
		"域名 %s 未指向服务器 %s，请先配置 CNAME 或 A 记录",
		domainName, expectedIP,
	)
}

func containsIP(values []net.IP, expected net.IP) bool {
	for _, value := range values {
		if value.Equal(expected) {
			return true
		}
	}
	return false
}

func (service *DomainBindService) Verify(
	ctx context.Context,
	value string,
) (DomainBindStatus, error) {
	actor, err := domainBindActor(ctx)
	if err != nil {
		return DomainBindStatus{}, err
	}
	domainName, err := service.validateDomain(ctx, actor.tenantID, value)
	if err != nil {
		return DomainBindStatus{}, err
	}
	if _, err := service.beginVerification(actor.tenantID, domainName); err != nil {
		return DomainBindStatus{}, err
	}
	ok, verifyErr := verifyDomainWithResolver(domainName, service.config.ExpectedIP, service.resolver)
	if verifyErr != nil || !ok {
		if verifyErr == nil {
			verifyErr = errors.New("域名 DNS 验证失败")
		}
		service.finishVerification(
			actor.tenantID,
			domainName,
			DomainStateVerificationFailed,
			verifyErr.Error(),
		)
		return service.status(actor.tenantID), errorsx.BadRequest(verifyErr.Error())
	}
	if !service.finishVerification(
		actor.tenantID,
		domainName,
		DomainStateVerified,
		"DNS 验证通过",
	) {
		return DomainBindStatus{}, errorsx.Conflict("域名状态已发生变化，请重新验证")
	}
	return service.status(actor.tenantID), nil
}

func (service *DomainBindService) Bind(
	ctx context.Context,
	value string,
) (DomainBindStatus, error) {
	actor, err := domainBindActor(ctx)
	if err != nil {
		return DomainBindStatus{}, err
	}
	if service.panel == nil || strings.TrimSpace(service.config.ExpectedIP) == "" {
		return DomainBindStatus{}, errorsx.BadRequest("域名自动绑定服务尚未完成配置")
	}
	domainName, err := service.validateDomain(ctx, actor.tenantID, value)
	if err != nil {
		return DomainBindStatus{}, err
	}
	current := service.status(actor.tenantID)
	if current.State != DomainStateVerified || current.Domain != domainName {
		return DomainBindStatus{}, errorsx.BadRequest("请先验证该域名的 DNS 解析")
	}
	ok, verifyErr := verifyDomainWithResolver(domainName, service.config.ExpectedIP, service.resolver)
	if verifyErr != nil || !ok {
		if verifyErr == nil {
			verifyErr = errors.New("域名 DNS 验证失败")
		}
		service.setStatus(actor.tenantID, domainName, DomainStateVerificationFailed, verifyErr.Error(), 1)
		return service.status(actor.tenantID), errorsx.BadRequest(verifyErr.Error())
	}
	started, err := service.startBind(actor.tenantID, domainName)
	if err != nil {
		return DomainBindStatus{}, err
	}
	background := context.WithoutCancel(ctx)
	go service.runBind(background, actor, domainName)
	return started, nil
}

func (service *DomainBindService) Status(
	ctx context.Context,
) (DomainBindStatus, error) {
	actor, err := domainBindActor(ctx)
	if err != nil {
		return DomainBindStatus{}, err
	}
	tenant, err := service.tenants.FindByID(ctx, actor.tenantID)
	if err != nil {
		return DomainBindStatus{}, mapNotFound(err, "tenant not found")
	}
	service.mu.RLock()
	_, exists := service.statuses[actor.tenantID]
	service.mu.RUnlock()
	if exists {
		return service.withTenantPortal(service.status(actor.tenantID), tenant), nil
	}
	if tenant.CustomDomain != nil && *tenant.CustomDomain != "" {
		service.reservePersistedDomain(actor.tenantID, *tenant.CustomDomain)
		return service.withTenantPortal(service.setStatus(
			actor.tenantID, *tenant.CustomDomain, DomainStateReady,
			"域名已绑定", 5,
		), tenant), nil
	}
	return service.withTenantPortal(service.setStatus(
		actor.tenantID, "", DomainStateNone, "尚未绑定域名", 0,
	), tenant), nil
}

func (service *DomainBindService) withTenantPortal(
	status DomainBindStatus,
	tenant *domain.Tenant,
) DomainBindStatus {
	status.TenantCode = tenant.Code
	platformHost := normalizePortalHost(service.config.ReservedDomain)
	if platformHost == "" {
		platformHost = "play.imai.work"
	}
	status.DefaultPortalURL = "https://" + platformHost + "/t/" + tenant.Code
	return status
}

func (service *DomainBindService) Unbind(
	ctx context.Context,
) (DomainBindStatus, error) {
	actor, err := domainBindActor(ctx)
	if err != nil {
		return DomainBindStatus{}, err
	}
	if service.active(actor.tenantID) {
		return DomainBindStatus{}, errorsx.Conflict("域名正在配置中，请稍候")
	}
	tenant, err := service.tenants.FindByID(ctx, actor.tenantID)
	if err != nil {
		return DomainBindStatus{}, mapNotFound(err, "tenant not found")
	}
	domainName := ""
	persisted := tenant.CustomDomain != nil && *tenant.CustomDomain != ""
	if persisted {
		domainName = *tenant.CustomDomain
	} else {
		current := service.status(actor.tenantID)
		if current.State == DomainStateSetupFailed {
			domainName = current.Domain
		}
	}
	if domainName == "" {
		return service.setStatus(actor.tenantID, "", DomainStateNone, "尚未绑定域名", 0), nil
	}
	reserved := strings.ToLower(strings.TrimSuffix(
		strings.TrimSpace(service.config.ReservedDomain),
		".",
	))
	isReserved := strings.ToLower(strings.TrimSuffix(
		strings.TrimSpace(domainName),
		".",
	)) == reserved
	if !isReserved &&
		(persisted || service.ownsDomain(actor.tenantID, domainName)) {
		if service.panel == nil {
			return DomainBindStatus{}, errorsx.BadRequest("域名自动绑定服务尚未完成配置")
		}
		if err := service.panel.DeleteSite(domainName); err != nil && !missingSiteError(err) {
			slog.Error(
				"delete baota site during domain unbind",
				"tenant_id", actor.tenantID,
				"domain", domainName,
				"request_id", middleware.RequestIDFromContext(ctx),
				"error", err,
			)
			service.setStatus(actor.tenantID, domainName, DomainStateSetupFailed, "删除站点失败，请稍后重试或联系平台管理员", 0)
			return service.status(actor.tenantID), errorsx.Internal("delete baota site failed")
		}
	}
	if persisted {
		tenant.CustomDomain = nil
		if err := service.tenants.Update(ctx, tenant); err != nil {
			service.setStatus(actor.tenantID, domainName, DomainStateSetupFailed, "清除租户域名失败", 0)
			return service.status(actor.tenantID), errorsx.Internal("clear custom domain failed")
		}
	}
	service.releaseDomain(actor.tenantID, domainName)
	result := service.setStatus(actor.tenantID, "", DomainStateNone, "域名已解绑", 0)
	service.recordAudit(ctx, actor, "domain.unbind", domainName, DomainStateNone)
	return result, nil
}

func missingSiteError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not found") ||
		strings.Contains(message, "does not exist") ||
		strings.Contains(message, "不存在")
}

func (service *DomainBindService) runBind(
	ctx context.Context,
	actor domainBindIdentity,
	domainName string,
) {
	siteCreated := false
	currentStep := 2
	fail := func(message string, err error) {
		rollbackSucceeded := !siteCreated
		var rollbackErr error
		if siteCreated {
			rollbackErr = service.panel.DeleteSite(domainName)
			rollbackSucceeded = rollbackErr == nil || missingSiteError(rollbackErr)
		}
		if rollbackSucceeded {
			service.releaseDomain(actor.tenantID, domainName)
		}
		slog.Error(
			"tenant domain provisioning failed",
			"tenant_id", actor.tenantID,
			"domain", domainName,
			"step", currentStep,
			"request_id", middleware.RequestIDFromContext(ctx),
			"error", err,
			"rollback_error", rollbackErr,
		)
		service.recordAudit(ctx, actor, "domain.bind_failed", domainName, DomainStateSetupFailed)
		service.setStatus(
			actor.tenantID,
			domainName,
			DomainStateSetupFailed,
			message+"，请稍后重试或联系平台管理员",
			currentStep,
		)
	}

	if _, err := service.panel.AddSite(domainName); err != nil {
		fail("创建宝塔站点失败", err)
		return
	}
	siteCreated = true
	currentStep = 3
	service.setStatus(actor.tenantID, domainName, DomainStateConfiguring, "正在配置反向代理", 3)
	if err := service.panel.AddReverseProxy(domainName, "imaiplay", service.config.ProxyTarget); err != nil {
		fail("配置反向代理失败", err)
		return
	}
	if err := service.panel.AddNginxSnippet(domainName, disableAdminSnippet); err != nil {
		fail("配置管理后台访问限制失败", err)
		return
	}
	currentStep = 4
	service.setStatus(actor.tenantID, domainName, DomainStateConfiguring, "正在申请 Let's Encrypt 证书", 4)
	if err := service.panel.ApplyLetsEncrypt(domainName); err != nil {
		fail("申请 Let's Encrypt 证书失败", err)
		return
	}
	service.setStatus(actor.tenantID, domainName, DomainStateConfiguring, "正在等待 HTTPS 证书生效", 4)
	if err := service.waitForSSL(domainName); err != nil {
		fail("HTTPS 证书未就绪", err)
		return
	}
	currentStep = 5
	tenant, err := service.tenants.FindByID(ctx, actor.tenantID)
	if err != nil {
		fail("读取租户失败", err)
		return
	}
	tenant.CustomDomain = &domainName
	if err := service.tenants.Update(ctx, tenant); err != nil {
		fail("保存租户域名失败", err)
		return
	}
	service.recordAudit(ctx, actor, "domain.bind", domainName, DomainStateReady)
	service.setStatus(actor.tenantID, domainName, DomainStateReady, "域名绑定完成", 5)
}

func (service *DomainBindService) waitForSSL(domainName string) error {
	var lastErr error
	for attempt := 0; attempt < service.config.MaxPolls; attempt++ {
		info, err := service.panel.GetSiteInfo(domainName)
		if err == nil && sslReady(info) {
			return nil
		}
		if err != nil {
			lastErr = err
		}
		if attempt+1 < service.config.MaxPolls {
			time.Sleep(service.config.PollInterval)
		}
	}
	if lastErr != nil {
		return fmt.Errorf("宝塔未确认 Let's Encrypt 证书已生效：%w", lastErr)
	}
	return errors.New("宝塔未确认 Let's Encrypt 证书已生效")
}

func sslReady(info map[string]interface{}) bool {
	for _, key := range []string{"ssl_status", "sslStatus", "ssl", "certificate"} {
		value, exists := info[key]
		if !exists {
			continue
		}
		switch current := value.(type) {
		case bool:
			if current {
				return true
			}
		case string:
			switch strings.ToLower(strings.TrimSpace(current)) {
			case "ready", "valid", "active", "enabled", "success", "true":
				return true
			}
		case float64:
			if current > 0 {
				return true
			}
		case int:
			if current > 0 {
				return true
			}
		case map[string]interface{}:
			if valid, ok := current["valid"].(bool); ok && valid {
				return true
			}
		}
	}
	return false
}

func (service *DomainBindService) validateDomain(
	ctx context.Context,
	tenantID, value string,
) (string, error) {
	domainName := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "."))
	if domainName == "" || strings.Contains(domainName, "*") || !validCustomDomain(domainName) {
		return domainName, errorsx.BadRequest("请输入合法的自定义域名")
	}
	reserved := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(service.config.ReservedDomain), "."))
	if reserved != "" && domainName == reserved {
		return domainName, errorsx.BadRequest("总后台域名不能绑定给租户")
	}
	current, err := service.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return domainName, mapNotFound(err, "tenant not found")
	}
	if current.CustomDomain != nil && strings.TrimSpace(*current.CustomDomain) != "" {
		return domainName, errorsx.Conflict("当前租户已有绑定域名，请先解绑")
	}
	other, err := service.tenants.FindByCustomDomain(ctx, domainName)
	if err == nil && other.ID != tenantID {
		return domainName, errorsx.Conflict("该域名已被其他租户绑定")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return domainName, errorsx.Internal("查询自定义域名失败")
	}
	return domainName, nil
}

func (service *DomainBindService) setStatus(
	tenantID, domainName, state, message string,
	step int,
) DomainBindStatus {
	service.mu.Lock()
	defer service.mu.Unlock()
	return service.setStatusLocked(tenantID, domainName, state, message, step)
}

func (service *DomainBindService) setStatusLocked(
	tenantID, domainName, state, message string,
	step int,
) DomainBindStatus {
	current := service.statuses[tenantID]
	if current.Domain != domainName {
		current.History = nil
	}
	if len(current.History) == 0 || current.History[len(current.History)-1] != state {
		current.History = append(current.History, state)
	}
	current.State, current.Domain, current.Message = state, domainName, message
	current.CurrentStep, current.TotalSteps = step, 5
	current.CNAMETarget = service.config.CNAMETarget
	current.UpdatedAt = time.Now().UTC()
	service.statuses[tenantID] = current
	return cloneDomainStatus(current)
}

func (service *DomainBindService) startBind(
	tenantID, domainName string,
) (DomainBindStatus, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	current := service.statuses[tenantID]
	if current.State == DomainStateCreatingSite || current.State == DomainStateConfiguring {
		return DomainBindStatus{}, errorsx.Conflict("域名正在配置中，请稍候")
	}
	if current.State != DomainStateVerified || current.Domain != domainName {
		return DomainBindStatus{}, errorsx.BadRequest("请先验证该域名的 DNS 解析")
	}
	if owner, exists := service.owners[domainName]; exists && owner != tenantID {
		return DomainBindStatus{}, errorsx.Conflict("该域名正在被其他租户绑定")
	}
	service.owners[domainName] = tenantID
	return service.setStatusLocked(
		tenantID, domainName, DomainStateCreatingSite,
		"正在创建宝塔站点", 2,
	), nil
}

func (service *DomainBindService) beginVerification(
	tenantID, domainName string,
) (DomainBindStatus, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	current := service.statuses[tenantID]
	if current.State == DomainStateCreatingSite || current.State == DomainStateConfiguring {
		return DomainBindStatus{}, errorsx.Conflict("域名正在配置中，请稍候")
	}
	if current.State == DomainStateSetupFailed &&
		current.Domain != "" &&
		service.owners[current.Domain] == tenantID {
		return DomainBindStatus{}, errorsx.Conflict("上次绑定的站点尚未清理，请先解绑")
	}
	return service.setStatusLocked(
		tenantID,
		domainName,
		DomainStatePendingVerification,
		"正在验证 DNS 解析",
		1,
	), nil
}

func (service *DomainBindService) finishVerification(
	tenantID, domainName, state, message string,
) bool {
	service.mu.Lock()
	defer service.mu.Unlock()
	current := service.statuses[tenantID]
	if current.State != DomainStatePendingVerification || current.Domain != domainName {
		return false
	}
	service.setStatusLocked(tenantID, domainName, state, message, 1)
	return true
}

func (service *DomainBindService) active(tenantID string) bool {
	current := service.status(tenantID)
	return current.State == DomainStateCreatingSite || current.State == DomainStateConfiguring
}

func (service *DomainBindService) reservePersistedDomain(
	tenantID, domainName string,
) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if _, exists := service.owners[domainName]; !exists {
		service.owners[domainName] = tenantID
	}
}

func (service *DomainBindService) releaseDomain(
	tenantID, domainName string,
) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if service.owners[domainName] == tenantID {
		delete(service.owners, domainName)
	}
}

func (service *DomainBindService) ownsDomain(
	tenantID, domainName string,
) bool {
	service.mu.RLock()
	defer service.mu.RUnlock()
	return service.owners[domainName] == tenantID
}

func (service *DomainBindService) status(tenantID string) DomainBindStatus {
	service.mu.RLock()
	defer service.mu.RUnlock()
	current, ok := service.statuses[tenantID]
	if !ok {
		return DomainBindStatus{
			State: DomainStateNone, TotalSteps: 5,
			CNAMETarget: service.config.CNAMETarget,
		}
	}
	return cloneDomainStatus(current)
}

func cloneDomainStatus(value DomainBindStatus) DomainBindStatus {
	value.History = append([]string(nil), value.History...)
	return value
}

type domainBindIdentity struct {
	userID, tenantID, email, role string
}

func domainBindActor(ctx context.Context) (domainBindIdentity, error) {
	userID, tenantID, email, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" || tenantID == "" {
		return domainBindIdentity{}, errorsx.Forbidden("permission denied")
	}
	return domainBindIdentity{
		userID: userID, tenantID: tenantID, email: email, role: role,
	}, nil
}

func (service *DomainBindService) recordAudit(
	ctx context.Context,
	actor domainBindIdentity,
	action, domainName, state string,
) {
	if service.audit == nil {
		return
	}
	_ = service.audit.Record(ctx, domain.AuditEvent{
		TenantID: actor.tenantID, UserID: actor.userID,
		UserEmail: actor.email, UserRole: actor.role,
		Action: action, ResourceType: "tenant_domain",
		ResourceID: actor.tenantID,
		RequestID:  middleware.RequestIDFromContext(ctx),
		Detail: AuditDetail(map[string]interface{}{
			"domain": domainName, "state": state,
		}),
	})
}
