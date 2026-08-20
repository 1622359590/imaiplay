package domain

import "time"

type LearnerEngagementState struct {
	BaseModel
	UserID                 string `gorm:"index;not null"`
	FirstLoginAt           *time.Time
	WelcomeSeenAt          *time.Time
	LastDailyPromptDate    string `gorm:"size:10;not null;default:''"`
	PendingPromptKey       string `gorm:"size:80;not null;default:''"`
	PendingPromptKind      string `gorm:"size:24;not null;default:''"`
	PendingPromptDate      string `gorm:"size:10;not null;default:''"`
	PendingPromptExpiresAt *time.Time
}
