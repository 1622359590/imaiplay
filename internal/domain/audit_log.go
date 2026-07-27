package domain

import "time"

type AuditLog struct {
	BaseModel
	UserID       string `json:"user_id"`
	UserEmail    string `json:"user_email"`
	UserRole     string `json:"user_role"`
	Action       string `gorm:"index;not null" json:"action"`
	ResourceType string `gorm:"index" json:"resource_type"`
	ResourceID   string `gorm:"index" json:"resource_id"`
	Detail       string `gorm:"type:text" json:"detail"`
	IP           string `json:"ip"`
	RequestID    string `gorm:"index" json:"request_id"`
}

type AuditEvent struct {
	TenantID     string
	UserID       string
	UserEmail    string
	UserRole     string
	Action       string
	ResourceType string
	ResourceID   string
	Detail       string
	IP           string
	RequestID    string
	CreatedAt    time.Time
}
