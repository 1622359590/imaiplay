package domain

type DomainBindJob struct {
	BaseModel
	Domain         string `gorm:"uniqueIndex;not null" json:"domain"`
	State          string `gorm:"index;not null" json:"state"`
	Message        string `json:"message"`
	CurrentStep    int    `gorm:"not null;default:0" json:"current_step"`
	ErrorMessage   string `json:"error_message,omitempty"`
	ExternalSiteID int    `gorm:"not null;default:0" json:"external_site_id,omitempty"`
	AttemptCount   int    `gorm:"not null;default:0" json:"attempt_count"`
}
