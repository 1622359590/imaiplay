package service

import (
	"context"
	"errors"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/middleware"
	"gorm.io/gorm"
)

func (service *DomainBindService) Status(
	ctx context.Context,
) (DomainBindStatus, error) {
	actor, err := domainBindActor(ctx)
	if err != nil {
		return DomainBindStatus{}, err
	}
	return service.statusForActor(ctx, actor)
}

func (service *DomainBindService) StatusForTenant(
	ctx context.Context,
	tenantID string,
) (DomainBindStatus, error) {
	actor, err := domainBindSuperadminActor(ctx, tenantID)
	if err != nil {
		return DomainBindStatus{}, err
	}
	return service.statusForActor(ctx, actor)
}

func (service *DomainBindService) statusForActor(
	ctx context.Context,
	actor domainBindIdentity,
) (DomainBindStatus, error) {
	tenant, err := service.tenants.FindByID(ctx, actor.tenantID)
	if err != nil {
		return DomainBindStatus{}, mapNotFound(err, "tenant not found")
	}
	current := service.status(actor.tenantID)
	if current.State != DomainStateNone {
		return service.withTenantPortal(current, tenant), nil
	}
	if tenant.CustomDomain != nil && *tenant.CustomDomain != "" {
		service.reservePersistedDomain(actor.tenantID, *tenant.CustomDomain)
		service.reserveJob(actor.tenantID, *tenant.CustomDomain, DomainStateReady, "域名已绑定", 5)
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
	if service.jobs != nil {
		errorMessage := ""
		if state == DomainStateSetupFailed || state == DomainStateVerificationFailed {
			errorMessage = message
		}
		_ = service.jobs.UpdateStatus(context.Background(), tenantID, state, message, step, errorMessage)
	}
	return cloneDomainStatus(current)
}

func (service *DomainBindService) startBind(
	tenantID, domainName string,
) (DomainBindStatus, error) {
	current := service.status(tenantID)
	service.mu.Lock()
	defer service.mu.Unlock()
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
	if service.jobs != nil {
		_ = service.jobs.IncrementAttempt(context.Background(), tenantID)
	}
	return service.setStatusLocked(
		tenantID, domainName, DomainStateCreatingSite,
		"正在创建宝塔站点", 2,
	), nil
}

func (service *DomainBindService) beginVerification(
	tenantID, domainName string,
) (DomainBindStatus, error) {
	if err := service.reserveJob(tenantID, domainName, DomainStatePendingVerification, "正在验证 DNS 解析", 1); err != nil {
		return DomainBindStatus{}, errorsx.Conflict("该域名正在被其他租户绑定")
	}
	current := service.status(tenantID)
	service.mu.Lock()
	defer service.mu.Unlock()
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
	if service.jobs != nil {
		job, err := service.jobs.FindByDomain(context.Background(), domainName)
		return err == nil && job.TenantID == tenantID
	}
	service.mu.RLock()
	defer service.mu.RUnlock()
	return service.owners[domainName] == tenantID
}

func (service *DomainBindService) status(tenantID string) DomainBindStatus {
	if service.jobs != nil {
		job, err := service.jobs.FindByTenant(context.Background(), tenantID)
		if err == nil {
			return DomainBindStatus{
				State: job.State, Domain: job.Domain, Message: job.Message,
				CurrentStep: job.CurrentStep, TotalSteps: 5,
				CNAMETarget: service.config.CNAMETarget, UpdatedAt: job.UpdatedAt,
			}
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return DomainBindStatus{State: DomainStateSetupFailed, Message: "读取域名配置任务失败", TotalSteps: 5, CNAMETarget: service.config.CNAMETarget}
		}
	}
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

func (service *DomainBindService) reserveJob(tenantID, domainName, state, message string, step int) error {
	if service.jobs == nil {
		return nil
	}
	return service.jobs.Reserve(context.Background(), &domain.DomainBindJob{
		BaseModel: domain.BaseModel{TenantID: tenantID}, Domain: domainName,
		State: state, Message: message, CurrentStep: step,
	})
}

func cloneDomainStatus(value DomainBindStatus) DomainBindStatus {
	value.History = append([]string(nil), value.History...)
	return value
}

type domainBindIdentity struct {
	userID, tenantID, email, role string
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
