package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	tenantcontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/security"
	"github.com/1622359590/imaiplay/internal/sms"
	"gorm.io/gorm"
	"log/slog"
)

type AuthService struct {
	users          repository.UserRepository
	tenants        repository.TenantRepository
	jwtSecret      string
	refreshTokens  repository.RefreshTokenRepository
	passwordResets repository.PasswordResetRepository
	smsSender      sms.SMSSender
}

func NewAuthService(
	users repository.UserRepository,
	tenants repository.TenantRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{users: users, tenants: tenants, jwtSecret: jwtSecret, smsSender: sms.NewLogSender(slog.Default())}
}

func (service *AuthService) SetPasswordResetRepository(resets repository.PasswordResetRepository) {
	service.passwordResets = resets
}
func (service *AuthService) SetSMSSender(sender sms.SMSSender) {
	if sender != nil {
		service.smsSender = sender
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

func (service *AuthService) Register(
	ctx context.Context,
	email, password, name, role string,
) (*domain.User, error) {
	return service.RegisterWithPhone(ctx, email, "", password, name, role)
}

func (service *AuthService) RegisterWithPhone(ctx context.Context, email, phone, password, name, role string) (*domain.User, error) {
	if role == "superadmin" {
		return nil, errorsx.BadRequest("superadmin 不可通過公開註冊創建")
	}
	if !isTenantRole(role) {
		return nil, errorsx.BadRequest("invalid role")
	}
	tenant, err := service.currentTenant(ctx)
	if err != nil {
		return nil, err
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := service.users.FindByEmailAndTenant(ctx, email, tenant.ID); err == nil {
		return nil, errorsx.Conflict("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorsx.Internal("find user failed")
	}
	phone = normalizePhone(phone)
	if phone != "" {
		if !validPhone(phone) {
			return nil, errorsx.BadRequest("invalid phone")
		}
		if _, err := service.users.FindByPhoneAndTenant(ctx, phone, tenant.ID); err == nil {
			return nil, errorsx.Conflict("phone already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorsx.Internal("find user failed")
		}
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, errorsx.Internal("hash password failed")
	}
	user := &domain.User{
		BaseModel: domain.BaseModel{TenantID: tenant.ID},
		Email:     email, Phone: nullablePhone(phone), Password: hash, Name: name, Role: role, Status: 1,
	}
	if err := service.users.Create(ctx, user); err != nil {
		return nil, mapCreateError(err, "email already exists", "create user failed")
	}
	return user, nil
}

func (service *AuthService) BootstrapSuperadmin(ctx context.Context, email, name, password string) (*domain.User, *TokenPair, error) {
	users, _, err := service.users.FindByTenant(ctx, "", 0, 100000)
	if err != nil {
		return nil, nil, errorsx.Internal("find superadmin failed")
	}
	for _, user := range users {
		if user.Role == "superadmin" {
			return nil, nil, errorsx.Conflict("superadmin already initialized")
		}
	}
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if email == "" || name == "" || len(password) < 8 {
		return nil, nil, errorsx.BadRequest("email, name and password are required")
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, nil, errorsx.Internal("hash password failed")
	}
	user := &domain.User{BaseModel: domain.BaseModel{TenantID: ""}, Email: email, Password: hash, Name: name, Role: "superadmin", Status: 1}
	if err := service.users.Create(ctx, user); err != nil {
		return nil, nil, errorsx.Internal("create superadmin failed")
	}
	pair, err := service.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	return user, pair, nil
}

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
	return service.issueTokens(ctx, user)
}

func (service *AuthService) ForgotPassword(ctx context.Context, phone string) error {
	if service.passwordResets == nil {
		return errorsx.Internal("password reset is not configured")
	}
	phone = normalizePhone(phone)
	tenant, err := service.currentTenant(ctx)
	if err != nil {
		return err
	}
	user, err := service.users.FindByPhoneAndTenant(ctx, phone, tenant.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return errorsx.Internal("find user failed")
	}
	latest, latestErr := service.passwordResets.FindLatest(ctx, tenant.ID, phone)
	if latestErr == nil && time.Since(latest.CreatedAt) < time.Minute {
		return errorsx.Conflict("please wait before requesting another code")
	}
	code, err := verificationCode()
	if err != nil {
		return errorsx.Internal("generate verification code failed")
	}
	hash := hashVerificationCode(code)
	reset := &domain.PasswordReset{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Phone: phone, CodeHash: hash, ExpiresAt: time.Now().UTC().Add(5 * time.Minute)}
	if err := service.passwordResets.Create(ctx, reset); err != nil {
		return errorsx.Internal("create password reset failed")
	}
	if err := service.smsSender.Send(ctx, phone, "", map[string]string{"code": code}); err != nil {
		return errorsx.Internal("send verification code failed")
	}
	_ = user
	return nil
}

func (service *AuthService) ResetPassword(ctx context.Context, phone, code, newPassword string) error {
	if service.passwordResets == nil {
		return errorsx.Internal("password reset is not configured")
	}
	if len(newPassword) < 8 {
		return errorsx.BadRequest("password must be at least 8 characters")
	}
	phone = normalizePhone(phone)
	tenant, err := service.currentTenant(ctx)
	if err != nil {
		return err
	}
	reset, err := service.passwordResets.FindLatest(ctx, tenant.ID, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) || err != nil {
		return errorsx.BadRequest("invalid or expired verification code")
	}
	if reset.Used || reset.ExpiresAt.Before(time.Now().UTC()) {
		return errorsx.BadRequest("invalid or expired verification code")
	}
	if reset.Attempts >= 5 {
		return errorsx.BadRequest("too many verification attempts")
	}
	if subtle.ConstantTimeCompare([]byte(reset.CodeHash), []byte(hashVerificationCode(code))) != 1 {
		if err := service.passwordResets.IncrementAttempts(ctx, reset.ID); err != nil {
			return errorsx.Internal("update verification attempts failed")
		}
		return errorsx.BadRequest("invalid verification code")
	}
	user, err := service.users.FindByPhoneAndTenant(ctx, phone, tenant.ID)
	if err != nil {
		return errorsx.BadRequest("invalid verification code")
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return errorsx.Internal("hash password failed")
	}
	user.Password = hash
	userCtx := tenantcontext.WithUser(ctx, user.ID, user.TenantID, user.Email, user.Role)
	if err := service.users.Update(userCtx, user); err != nil {
		return errorsx.Internal("update password failed")
	}
	if err := service.passwordResets.MarkUsed(ctx, reset.ID); err != nil {
		return errorsx.Internal("consume password reset failed")
	}
	if service.refreshTokens != nil {
		if err := service.refreshTokens.RevokeAllForUser(ctx, user.ID); err != nil {
			return errorsx.Internal("revoke refresh tokens failed")
		}
	}
	return nil
}

func verificationCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}
func hashVerificationCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return fmt.Sprintf("%x", hash[:])
}
func normalizePhone(phone string) string { return strings.TrimSpace(phone) }
func validPhone(phone string) bool {
	if len(phone) != 11 || phone[0] != '1' {
		return false
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
func nullablePhone(phone string) *string {
	if phone == "" {
		return nil
	}
	return &phone
}

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
	userCtx := tenantcontext.WithUser(ctx, token.UserID, token.TenantID, "", "")
	user, err := service.users.FindByID(userCtx, token.UserID)
	if err != nil || user.Status != 1 {
		return nil, errorsx.Unauthorized("invalid refresh token")
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
	if tenant.Status != 1 {
		return nil, errorsx.Forbidden("tenant is disabled")
	}
	return tenant, nil
}

func isTenantRole(role string) bool {
	return role == "tenant_admin" || role == "instructor" || role == "learner"
}
