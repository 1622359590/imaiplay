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
	"github.com/1622359590/imaiplay/internal/security"
	"gorm.io/gorm"
)

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
	reset := &domain.PasswordReset{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Phone: phone, Purpose: "password_reset", CodeHash: hash, ExpiresAt: time.Now().UTC().Add(5 * time.Minute)}
	if err := service.passwordResets.Create(ctx, reset); err != nil {
		return errorsx.Internal("create password reset failed")
	}
	if err := service.smsSender.Send(ctx, phone, "", map[string]string{"code": code}); err != nil {
		return errorsx.Internal("send verification code failed")
	}
	_ = user
	return nil
}

func (service *AuthService) SendLoginCode(ctx context.Context, phone string) error {
	if service.passwordResets == nil {
		return errorsx.Internal("login code is not configured")
	}
	phone = normalizePhone(phone)
	tenant, err := service.currentTenant(ctx)
	if err != nil {
		return err
	}
	if !validPhone(phone) {
		return errorsx.BadRequest("invalid phone")
	}
	if latest, latestErr := service.passwordResets.FindLatestForPurpose(ctx, tenant.ID, phone, "login_code"); latestErr == nil && time.Since(latest.CreatedAt) < time.Minute {
		return errorsx.Conflict("please wait before requesting another code")
	}
	user, findErr := service.users.FindByPhoneAndTenant(ctx, phone, tenant.ID)
	if errors.Is(findErr, gorm.ErrRecordNotFound) {
		return nil
	}
	if findErr != nil {
		return errorsx.Internal("find user failed")
	}
	if user.Status != 1 {
		return nil
	}
	code, err := verificationCode()
	if err != nil {
		return errorsx.Internal("generate verification code failed")
	}
	reset := &domain.PasswordReset{BaseModel: domain.BaseModel{TenantID: tenant.ID}, Phone: phone, Purpose: "login_code", CodeHash: hashVerificationCode(code), ExpiresAt: time.Now().UTC().Add(5 * time.Minute)}
	if err := service.passwordResets.Create(ctx, reset); err != nil {
		return errorsx.Internal("create login code failed")
	}
	if err := service.smsSender.Send(ctx, phone, "", map[string]string{"code": code}); err != nil {
		return errorsx.Internal("send verification code failed")
	}
	return nil
}

func (service *AuthService) LoginWithCode(ctx context.Context, phone, code string) (*LoginOutcome, error) {
	if service.passwordResets == nil {
		return nil, errorsx.Internal("login code is not configured")
	}
	phone = normalizePhone(phone)
	if !validPhone(phone) {
		return nil, errorsx.Unauthorized("invalid verification code")
	}
	tenant, err := service.currentTenant(ctx)
	if err != nil {
		return nil, err
	}
	reset, err := service.passwordResets.FindLatestForPurpose(ctx, tenant.ID, phone, "login_code")
	if errors.Is(err, gorm.ErrRecordNotFound) || err != nil || reset.Used || reset.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errorsx.Unauthorized("invalid verification code")
	}
	if reset.Attempts >= 5 {
		return nil, errorsx.Unauthorized("too many verification attempts")
	}
	if subtle.ConstantTimeCompare([]byte(reset.CodeHash), []byte(hashVerificationCode(code))) != 1 {
		_ = service.passwordResets.IncrementAttempts(ctx, reset.ID)
		return nil, errorsx.Unauthorized("invalid verification code")
	}
	user, err := service.users.FindByPhoneAndTenant(ctx, phone, tenant.ID)
	if err != nil || user.Status != 1 {
		return nil, errorsx.Unauthorized("invalid verification code")
	}
	if err := service.passwordResets.MarkUsed(ctx, reset.ID); err != nil {
		return nil, errorsx.Internal("consume login code failed")
	}
	return service.completeTenantLogin(ctx, user, tenant)
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
