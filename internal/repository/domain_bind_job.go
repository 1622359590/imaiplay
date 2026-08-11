package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type DomainBindJobRepository interface {
	FindByTenant(ctx context.Context, tenantID string) (*domain.DomainBindJob, error)
	FindByDomain(ctx context.Context, domainName string) (*domain.DomainBindJob, error)
	Reserve(ctx context.Context, job *domain.DomainBindJob) error
	UpdateStatus(ctx context.Context, tenantID, state, message string, step int, errorMessage string) error
	IncrementAttempt(ctx context.Context, tenantID string) error
	Delete(ctx context.Context, tenantID string) error
}
