package domain

type ResourceCategory struct {
	BaseModel
	Name     string  `gorm:"not null" json:"name"`
	ParentID *string `json:"parent_id,omitempty"`
}
