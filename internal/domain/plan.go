package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Plan struct {
	ID                string    `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"not null" json:"name"`
	StorageQuotaBytes int64     `gorm:"not null;default:0" json:"storage_quota_bytes"`
	MaxUsers          int       `gorm:"not null;default:0" json:"max_users"`
	MaxCourses        int       `gorm:"not null;default:0" json:"max_courses"`
	Features          string    `gorm:"type:text" json:"features"`
	IsDefault         bool      `gorm:"not null;default:false" json:"is_default"`
	Status            int       `gorm:"not null;default:1" json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (plan *Plan) BeforeCreate(_ *gorm.DB) error {
	if plan.ID == "" {
		plan.ID = uuid.NewString()
	}
	return nil
}
