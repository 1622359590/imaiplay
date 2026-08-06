package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"gorm.io/gorm"
)

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
	return service.verify(ctx, actor, value)
}

func (service *DomainBindService) VerifyForTenant(
	ctx context.Context,
	tenantID, value string,
) (DomainBindStatus, error) {
	actor, err := domainBindSuperadminActor(ctx, tenantID)
	if err != nil {
		return DomainBindStatus{}, err
	}
	return service.verify(ctx, actor, value)
}

func (service *DomainBindService) verify(
	ctx context.Context,
	actor domainBindIdentity,
	value string,
) (DomainBindStatus, error) {
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

func domainBindActor(ctx context.Context) (domainBindIdentity, error) {
	userID, tenantID, email, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" || tenantID == "" {
		return domainBindIdentity{}, errorsx.Forbidden("permission denied")
	}
	return domainBindIdentity{
		userID: userID, tenantID: tenantID, email: email, role: role,
	}, nil
}

func domainBindSuperadminActor(ctx context.Context, tenantID string) (domainBindIdentity, error) {
	userID, _, email, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "superadmin" || strings.TrimSpace(tenantID) == "" {
		return domainBindIdentity{}, errorsx.Forbidden("permission denied")
	}
	return domainBindIdentity{
		userID: userID, tenantID: strings.TrimSpace(tenantID), email: email, role: role,
	}, nil
}
