package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/middleware"
)

func (service *DomainBindService) Bind(
	ctx context.Context,
	value string,
) (DomainBindStatus, error) {
	actor, err := domainBindActor(ctx)
	if err != nil {
		return DomainBindStatus{}, err
	}
	return service.bind(ctx, actor, value)
}

func (service *DomainBindService) BindForTenant(
	ctx context.Context,
	tenantID, value string,
) (DomainBindStatus, error) {
	actor, err := domainBindSuperadminActor(ctx, tenantID)
	if err != nil {
		return DomainBindStatus{}, err
	}
	return service.bind(ctx, actor, value)
}

func (service *DomainBindService) bind(
	ctx context.Context,
	actor domainBindIdentity,
	value string,
) (DomainBindStatus, error) {
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

func (service *DomainBindService) Unbind(
	ctx context.Context,
) (DomainBindStatus, error) {
	actor, err := domainBindActor(ctx)
	if err != nil {
		return DomainBindStatus{}, err
	}
	return service.unbind(ctx, actor)
}

func (service *DomainBindService) UnbindForTenant(
	ctx context.Context,
	tenantID string,
) (DomainBindStatus, error) {
	actor, err := domainBindSuperadminActor(ctx, tenantID)
	if err != nil {
		return DomainBindStatus{}, err
	}
	return service.unbind(ctx, actor)
}

func (service *DomainBindService) unbind(
	ctx context.Context,
	actor domainBindIdentity,
) (DomainBindStatus, error) {
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
