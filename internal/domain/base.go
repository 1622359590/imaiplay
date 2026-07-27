package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	TenantID  string    `gorm:"index;not null" json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (model *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if model.ID == "" {
		model.ID = uuid.NewString()
	}
	return nil
}
