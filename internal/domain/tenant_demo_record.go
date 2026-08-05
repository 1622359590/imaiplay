package domain

type TenantDemoRecord struct {
	BaseModel
	BatchID    string `gorm:"index;not null"`
	RecordType string `gorm:"not null"`
	RecordID   string `gorm:"not null"`
}
