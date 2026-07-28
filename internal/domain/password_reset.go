package domain

import "time"

type PasswordReset struct {
	BaseModel
	Phone     string    `gorm:"index;not null" json:"phone"`
	Purpose   string    `gorm:"size:16;index;default:'password_reset'" json:"-"`
	CodeHash  string    `gorm:"not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Used      bool      `gorm:"default:false;index" json:"used"`
	Attempts  int       `gorm:"default:0" json:"attempts"`
}
