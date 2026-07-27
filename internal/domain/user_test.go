package domain

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserBeforeCreateGeneratesUUIDAndOmitsPasswordFromJSON(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.AutoMigrate(&User{}); err != nil {
		t.Fatalf("migrate user: %v", err)
	}
	user := &User{
		BaseModel: BaseModel{TenantID: "tenant-1"},
		Email:     "admin@example.com",
		Password:  "secret",
		Name:      "Admin",
		Role:      "tenant_admin",
	}
	if err := database.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := uuid.Parse(user.ID); err != nil {
		t.Fatalf("ID = %q, want UUID: %v", user.ID, err)
	}
	body, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal user: %v", err)
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	if _, exists := fields["password"]; exists {
		t.Fatalf("password exposed in JSON: %s", body)
	}
}
