package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tenant struct {
	ID                      string     `gorm:"primaryKey" json:"id"`
	Code                    string     `gorm:"uniqueIndex;not null" json:"code"`
	Name                    string     `gorm:"not null" json:"name"`
	Status                  int        `gorm:"default:1" json:"status"`
	PrimaryColor            string     `gorm:"size:16" json:"primary_color,omitempty"`
	SelectedBackgroundColor string     `gorm:"size:16" json:"selected_background_color,omitempty"`
	SelectedTextColor       string     `gorm:"size:16" json:"selected_text_color,omitempty"`
	SelectedIconColor       string     `gorm:"size:16" json:"selected_icon_color,omitempty"`
	LogoURL                 string     `gorm:"size:500" json:"logo_url,omitempty"`
	WelcomeText             string     `gorm:"size:255" json:"welcome_text,omitempty"`
	BrowserTitle            string     `gorm:"size:255" json:"browser_title,omitempty"`
	PlanID                  *string    `gorm:"index" json:"plan_id,omitempty"`
	LifecycleStatus         string     `gorm:"size:16;index" json:"lifecycle_status"`
	TrialEndsAt             *time.Time `json:"trial_ends_at,omitempty"`
	CustomDomain            *string    `gorm:"uniqueIndex;size:255" json:"custom_domain,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

func (tenant *Tenant) BeforeCreate(_ *gorm.DB) error {
	if tenant.ID == "" {
		tenant.ID = uuid.NewString()
	}
	return nil
}
