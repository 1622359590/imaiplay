package service

import (
	"context"
	"errors"
	"time"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/security"
	"gorm.io/gorm"
)

func (service *AuthService) IssueTokens(ctx context.Context, user *domain.User) (*TokenPair, error) {
	return service.issueTokens(ctx, user)
}

func (service *AuthService) Refresh(ctx context.Context, raw string) (*TokenPair, error) {
	if service.refreshTokens == nil || raw == "" {
		return nil, errorsx.Unauthorized("invalid refresh token")
	}
	token, err := service.refreshTokens.FindValidByHash(ctx, security.HashRefreshToken(raw))
	if err != nil || token.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errorsx.Unauthorized("invalid refresh token")
	}
	role := ""
	if token.TenantID == "" {
		role = "superadmin"
	}
	userCtx := tenantcontext.WithUser(ctx, token.UserID, token.TenantID, "", role)
	user, err := service.users.FindByID(userCtx, token.UserID)
	if err != nil || user.Status != 1 {
		return nil, errorsx.Unauthorized("invalid refresh token")
	}
	if user.Role != "superadmin" {
		tenant, findErr := service.tenants.FindByID(ctx, user.TenantID)
		if findErr != nil {
			return nil, errorsx.Unauthorized("invalid refresh token")
		}
		if accessible, _ := TenantAccessible(tenant, time.Now().UTC()); !accessible {
			return nil, errorsx.Unauthorized("invalid refresh token")
		}
	}
	if err := service.refreshTokens.Revoke(ctx, token.TokenHash); err != nil {
		return nil, errorsx.Internal("revoke refresh token failed")
	}
	return service.issueTokens(userCtx, user)
}

func (service *AuthService) Logout(ctx context.Context, raw string) error {
	if service.refreshTokens == nil {
		return nil
	}
	userID, _, _, _, ok := tenantcontext.UserFromContext(ctx)
	if !ok {
		return errorsx.Unauthorized("authentication required")
	}
	if raw != "" {
		token, err := service.refreshTokens.FindValidByHash(ctx, security.HashRefreshToken(raw))
		if err != nil || token.UserID != userID {
			return errorsx.Unauthorized("invalid refresh token")
		}
		if err := service.refreshTokens.Revoke(ctx, token.TokenHash); err != nil {
			return errorsx.Unauthorized("invalid refresh token")
		}
		return nil
	}
	return service.refreshTokens.RevokeAllForUser(ctx, userID)
}

func (service *AuthService) issueTokens(ctx context.Context, user *domain.User) (*TokenPair, error) {
	access, err := security.GenerateToken(user.ID, user.TenantID, user.Email, user.Role, service.jwtSecret)
	if err != nil {
		return nil, errorsx.Internal("generate token failed")
	}
	pair := &TokenPair{AccessToken: access, ExpiresAt: time.Now().UTC().Add(24 * time.Hour)}
	if service.refreshTokens == nil {
		return pair, nil
	}
	plain, hash, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, errorsx.Internal("generate refresh token failed")
	}
	if err := service.refreshTokens.Create(ctx, &domain.RefreshToken{BaseModel: domain.BaseModel{TenantID: user.TenantID}, UserID: user.ID, TokenHash: hash, ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour)}); err != nil {
		return nil, errorsx.Internal("store refresh token failed")
	}
	pair.RefreshToken = plain
	return pair, nil
}

func (service *AuthService) currentTenant(
	ctx context.Context,
) (*domain.Tenant, error) {
	code, _ := tenantcontext.TenantFromContext(ctx)
	tenant, err := service.tenants.FindByCode(ctx, code)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.NotFound("tenant not found")
	}
	if err != nil {
		return nil, errorsx.Internal("find tenant failed")
	}
	if accessible, reason := TenantAccessible(tenant, time.Now().UTC()); !accessible {
		return nil, errorsx.Forbidden(reason)
	}
	return tenant, nil
}

func isTenantRole(role string) bool {
	return role == "tenant_admin" || role == "instructor" || role == "learner"
}
