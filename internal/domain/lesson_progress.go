package domain

import "time"

type LessonProgress struct {
	BaseModel
	UserID              string     `gorm:"index;not null" json:"user_id"`
	LessonID            string     `gorm:"index;not null" json:"lesson_id"`
	ProgressPercent     int        `gorm:"default:0" json:"progress_percent"`
	Status              int        `gorm:"default:0" json:"status"`
	LastPositionSeconds int        `gorm:"default:0" json:"last_position_seconds"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
}

func (LessonProgress) TableName() string {
	return "lesson_progress"
}
