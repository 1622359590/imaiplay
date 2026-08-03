package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginChallenge struct {
	ID               string     `gorm:"primaryKey" json:"id"`
	TokenHash        string     `gorm:"uniqueIndex;not null" json:"-"`
	CandidateUserIDs string     `gorm:"type:text;not null" json:"-"`
	ExpiresAt        time.Time  `gorm:"index;not null" json:"expires_at"`
	ConsumedAt       *time.Time `gorm:"index" json:"consumed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (challenge *LoginChallenge) BeforeCreate(_ *gorm.DB) error {
	if challenge.ID == "" {
		challenge.ID = uuid.NewString()
	}
	return nil
}
