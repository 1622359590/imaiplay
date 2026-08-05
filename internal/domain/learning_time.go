package domain

type LearningDailyStat struct {
	BaseModel
	UserID          string `gorm:"index;not null" json:"user_id"`
	StudyDate       string `gorm:"size:10;index;not null" json:"study_date"`
	DurationSeconds int64  `gorm:"not null;default:0" json:"duration_seconds"`
}

type LearningTimeReport struct {
	BaseModel
	UserID              string `gorm:"index;not null"`
	LessonID            string `gorm:"index;not null"`
	ReportID            string `gorm:"not null"`
	WatchedSecondsDelta int    `gorm:"not null"`
}
