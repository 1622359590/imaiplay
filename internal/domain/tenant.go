package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tenant struct {
	ID        string `gorm:"primaryKey"`
	Code      string `gorm:"uniqueIndex;not null"`
	Name      string `gorm:"not null"`
	Status    int    `gorm:"default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tenant *Tenant) BeforeCreate(_ *gorm.DB) error {
	if tenant.ID == "" {
		tenant.ID = uuid.NewString()
	}
	return nil
}
