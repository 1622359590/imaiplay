package migration

import (
	"strings"

	"github.com/1622359590/imaiplay/internal/domain"
	"gorm.io/gorm"
)

func ClearReservedTenantDomain(
	database *gorm.DB,
	reserved string,
) (int64, error) {
	reserved = normalizeReservedDomain(reserved)
	if reserved == "" {
		return 0, nil
	}

	var tenants []domain.Tenant
	if err := database.
		Where("custom_domain IS NOT NULL").
		Find(&tenants).Error; err != nil {
		return 0, err
	}
	ids := make([]string, 0)
	for _, tenant := range tenants {
		if tenant.CustomDomain != nil &&
			normalizeReservedDomain(*tenant.CustomDomain) == reserved {
			ids = append(ids, tenant.ID)
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	result := database.Model(&domain.Tenant{}).
		Where("id IN ?", ids).
		Update("custom_domain", nil)
	return result.RowsAffected, result.Error
}

func normalizeReservedDomain(value string) string {
	return strings.ToLower(
		strings.TrimSuffix(strings.TrimSpace(value), "."),
	)
}
