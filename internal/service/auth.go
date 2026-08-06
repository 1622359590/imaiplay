package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"log/slog"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/sms"
	"gorm.io/gorm"
)

type AuthService struct {
	users           repository.UserRepository
	tenants         repository.TenantRepository
	jwtSecret       string
	refreshTokens   repository.RefreshTokenRepository
	passwordResets  repository.PasswordResetRepository
	loginChallenges repository.LoginChallengeRepository
	portals         *PortalService
	smsSender       sms.SMSSender
}

func NewAuthService(
	users repository.UserRepository,
	tenants repository.TenantRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		users:     users,
		tenants:   tenants,
		jwtSecret: jwtSecret,
		portals:   NewPortalService(tenants, ""),
		smsSender: sms.NewLogSender(slog.Default()),
	}
}

func (service *AuthService) SetPasswordResetRepository(resets repository.PasswordResetRepository) {
	service.passwordResets = resets
}
func (service *AuthService) SetSMSSender(sender sms.SMSSender) {
	if sender != nil {
		service.smsSender = sender
	}
}

func (service *AuthService) SetLoginChallengeRepository(
	challenges repository.LoginChallengeRepository,
) {
	service.loginChallenges = challenges
}

func (service *AuthService) SetPortalService(portals *PortalService) {
	if portals != nil {
		service.portals = portals
	}
}

func NewAuthServiceWithRefreshTokens(users repository.UserRepository, tenants repository.TenantRepository, refreshTokens repository.RefreshTokenRepository, jwtSecret string) *AuthService {
	service := NewAuthService(users, tenants, jwtSecret)
	service.refreshTokens = refreshTokens
	return service
}

type TokenPair struct {
	AccessToken  string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type AuthUser struct {
	ID       string  `json:"id"`
	TenantID string  `json:"tenant_id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Phone    *string `json:"phone"`
	Role     string  `json:"role"`
}

type OrganizationOption struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	LogoURL string `json:"logo_url,omitempty"`
	Role    string `json:"role"`
}

type LoginOutcome struct {
	User                    *AuthUser            `json:"user,omitempty"`
	Tenant                  *Portal              `json:"tenant,omitempty"`
	Pair                    *TokenPair           `json:"-"`
	RequiresTenantSelection bool                 `json:"requires_tenant_selection"`
	SelectionToken          string               `json:"selection_token,omitempty"`
	Organizations           []OrganizationOption `json:"organizations,omitempty"`
}

func presentAuthUser(user *domain.User) *AuthUser {
	if user == nil {
		return nil
	}
	return &AuthUser{
		ID: user.ID, TenantID: user.TenantID, Name: user.Name,
		Email: user.Email, Phone: user.Phone, Role: user.Role,
	}
}

func (service *AuthService) CurrentUser(ctx context.Context) (AuthUser, error) {
	userID, tenantID, _, role, ok := tenantcontext.UserFromContext(ctx)
	if !ok || userID == "" {
		return AuthUser{}, errorsx.Unauthorized("authentication required")
	}
	if role == "superadmin" {
		if tenantID != "" {
			return AuthUser{}, errorsx.Forbidden("permission denied")
		}
	} else if !isTenantRole(role) || tenantID == "" {
		return AuthUser{}, errorsx.Forbidden("permission denied")
	}
	user, err := service.users.FindByID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if mismatched, mismatchErr := service.users.FindByIDAcrossTenants(
			ctx, userID,
		); mismatchErr == nil && mismatched != nil {
			return AuthUser{}, errorsx.Forbidden("permission denied")
		}
		return AuthUser{}, errorsx.Unauthorized("authentication required")
	}
	if err != nil {
		return AuthUser{}, errorsx.Internal("find user failed")
	}
	if user.Status != 1 {
		return AuthUser{}, errorsx.Forbidden("user is disabled")
	}
	if user.ID != userID || user.TenantID != tenantID || user.Role != role ||
		(role == "superadmin" && user.TenantID != "") ||
		(role != "superadmin" && user.TenantID == "") {
		return AuthUser{}, errorsx.Forbidden("permission denied")
	}
	return *presentAuthUser(user), nil
}

func normalizeLoginCredential(identifier string) string {
	identifier = strings.TrimSpace(identifier)
	if strings.Contains(identifier, "@") {
		return strings.ToLower(identifier)
	}
	return normalizePhone(identifier)
}
