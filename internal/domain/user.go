package domain

type User struct {
	BaseModel
	Tenant   Tenant  `gorm:"foreignKey:TenantID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Email    string  `gorm:"not null" json:"email"`
	Phone    *string `gorm:"index" json:"phone,omitempty"`
	Password string  `gorm:"not null" json:"-"`
	Name     string  `gorm:"not null" json:"name"`
	Role     string  `gorm:"not null;default:'learner'" json:"role"`
	Status   int     `gorm:"default:1" json:"status"`
}
