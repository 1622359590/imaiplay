package domain

import "time"

type RefreshToken struct {
	BaseModel
	UserID    string    `gorm:"index;not null" json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	Revoked   bool      `gorm:"default:false;index" json:"revoked"`
}
