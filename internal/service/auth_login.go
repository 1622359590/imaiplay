package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/security"
	"gorm.io/gorm"
)

func (service *AuthService) Login(
	ctx context.Context,
	identifier, password string,
) (string, error) {
	pair, err := service.LoginWithRefresh(ctx, identifier, password)
	if err != nil {
		return "", err
	}
	return pair.AccessToken, nil
}

func (service *AuthService) LoginWithRefresh(ctx context.Context, identifier, password string) (*TokenPair, error) {
	outcome, err := service.BeginLogin(ctx, identifier, password)
	if err != nil {
		return nil, err
	}
	if outcome.RequiresTenantSelection {
		return nil, errorsx.Conflict("account_exists_multiple_tenants")
	}
	return outcome.Pair, nil
}

func (service *AuthService) BeginLogin(
	ctx context.Context,
	identifier, password string,
) (*LoginOutcome, error) {
	code, _ := tenantcontext.TenantFromContext(ctx)
	if code == "" || code == tenantcontext.UnknownTenant {
		return service.beginPlatformLogin(ctx, identifier, password)
	}
	return service.beginScopedLogin(ctx, identifier, password)
}

func (service *AuthService) beginScopedLogin(
	ctx context.Context,
	identifier, password string,
) (*LoginOutcome, error) {
	tenant, err := service.currentTenant(ctx)
	if err != nil {
		return nil, err
	}
	var user *domain.User
	if strings.Contains(identifier, "@") {
		user, err = service.users.FindByEmailAndTenant(ctx, strings.ToLower(strings.TrimSpace(identifier)), tenant.ID)
	} else {
		user, err = service.users.FindByPhoneAndTenant(ctx, normalizePhone(identifier), tenant.ID)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || err == nil && !security.CheckPassword(password, user.Password) {
		return nil, errorsx.Unauthorized("invalid email or password")
	}
	if err != nil {
		return nil, errorsx.Internal("find user failed")
	}
	if user.Status != 1 {
		return nil, errorsx.Forbidden("user is disabled")
	}
	return service.completeTenantLogin(ctx, user, tenant)
}

func (service *AuthService) beginPlatformLogin(
	ctx context.Context,
	identifier, password string,
) (*LoginOutcome, error) {
	credential := normalizeLoginCredential(identifier)
	superadmin, err := service.findPlatformSuperadmin(ctx, credential)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.Internal("find user failed")
	}
	if err == nil && superadmin.Role == "superadmin" &&
		security.CheckPassword(password, superadmin.Password) {
		if superadmin.Status != 1 {
			return nil, errorsx.Forbidden("user is disabled")
		}
		pair, issueErr := service.issueTokens(ctx, superadmin)
		if issueErr != nil {
			return nil, issueErr
		}
		return &LoginOutcome{User: presentAuthUser(superadmin), Pair: pair}, nil
	}

	users, err := service.users.FindByCredentialAcrossTenants(ctx, credential)
	if err != nil {
		return nil, errorsx.Internal("find user failed")
	}
	type candidate struct {
		user   *domain.User
		tenant *domain.Tenant
	}
	candidates := make([]candidate, 0, len(users))
	for index := range users {
		user := &users[index]
		if !security.CheckPassword(password, user.Password) ||
			user.Status != 1 {
			continue
		}
		tenant, findErr := service.tenants.FindByID(ctx, user.TenantID)
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			continue
		}
		if findErr != nil {
			return nil, errorsx.Internal("find tenant failed")
		}
		if accessible, _ := TenantAccessible(
			tenant,
			time.Now().UTC(),
		); !accessible {
			continue
		}
		candidates = append(candidates, candidate{user: user, tenant: tenant})
	}
	if len(candidates) == 0 {
		return nil, errorsx.Unauthorized("invalid email or password")
	}
	if len(candidates) == 1 {
		return service.completeTenantLogin(
			ctx,
			candidates[0].user,
			candidates[0].tenant,
		)
	}
	if service.loginChallenges == nil {
		return nil, errorsx.Conflict("account_exists_multiple_tenants")
	}

	candidateIDs := make([]string, 0, len(candidates))
	organizations := make([]OrganizationOption, 0, len(candidates))
	for _, item := range candidates {
		candidateIDs = append(candidateIDs, item.user.ID)
		portal := service.portals.portalFromTenant(item.tenant)
		organizations = append(organizations, OrganizationOption{
			Code:    item.tenant.Code,
			Name:    item.tenant.Name,
			LogoURL: portal.LogoURL,
			Role:    item.user.Role,
		})
	}
	encodedIDs, err := json.Marshal(candidateIDs)
	if err != nil {
		return nil, errorsx.Internal("create organization selection failed")
	}
	selectionToken, tokenHash, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, errorsx.Internal("create organization selection failed")
	}
	if err := service.loginChallenges.Create(ctx, &domain.LoginChallenge{
		TokenHash:        tokenHash,
		CandidateUserIDs: string(encodedIDs),
		ExpiresAt:        time.Now().UTC().Add(5 * time.Minute),
	}); err != nil {
		return nil, errorsx.Internal("create organization selection failed")
	}
	return &LoginOutcome{
		RequiresTenantSelection: true,
		SelectionToken:          selectionToken,
		Organizations:           organizations,
	}, nil
}

func (service *AuthService) SelectTenant(
	ctx context.Context,
	selectionToken, tenantCode string,
) (*LoginOutcome, error) {
	selectionToken = strings.TrimSpace(selectionToken)
	tenantCode = strings.ToLower(strings.TrimSpace(tenantCode))
	if service.loginChallenges == nil ||
		selectionToken == "" || tenantCode == "" {
		return nil, errorsx.Unauthorized(
			"invalid or expired organization selection",
		)
	}
	challenge, err := service.loginChallenges.Consume(
		ctx,
		security.HashRefreshToken(selectionToken),
		time.Now().UTC(),
	)
	if err != nil {
		return nil, errorsx.Unauthorized(
			"invalid or expired organization selection",
		)
	}
	var candidateIDs []string
	if err := json.Unmarshal(
		[]byte(challenge.CandidateUserIDs),
		&candidateIDs,
	); err != nil {
		return nil, errorsx.Unauthorized(
			"invalid or expired organization selection",
		)
	}
	for _, userID := range candidateIDs {
		user, findErr := service.users.FindByIDAcrossTenants(ctx, userID)
		if findErr != nil {
			continue
		}
		tenant, findErr := service.tenants.FindByID(ctx, user.TenantID)
		if findErr != nil || tenant.Code != tenantCode ||
			user.Status != 1 {
			continue
		}
		if accessible, _ := TenantAccessible(
			tenant,
			time.Now().UTC(),
		); !accessible {
			continue
		}
		return service.completeTenantLogin(ctx, user, tenant)
	}
	return nil, errorsx.Unauthorized(
		"invalid or expired organization selection",
	)
}

func (service *AuthService) findPlatformSuperadmin(
	ctx context.Context,
	credential string,
) (*domain.User, error) {
	if strings.Contains(credential, "@") {
		return service.users.FindByEmailAndTenant(ctx, credential, "")
	}
	return service.users.FindByPhoneAndTenant(ctx, credential, "")
}

func (service *AuthService) completeTenantLogin(
	ctx context.Context,
	user *domain.User,
	tenant *domain.Tenant,
) (*LoginOutcome, error) {
	userCtx := tenantcontext.WithUser(
		ctx,
		user.ID,
		user.TenantID,
		user.Email,
		user.Role,
	)
	pair, err := service.issueTokens(userCtx, user)
	if err != nil {
		return nil, err
	}
	return &LoginOutcome{
		User:   presentAuthUser(user),
		Tenant: service.portals.portalFromTenant(tenant),
		Pair:   pair,
	}, nil
}
