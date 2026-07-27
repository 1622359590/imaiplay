package db

import (
	"testing"

	"github.com/1622359590/imaiplay/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPingAndClose(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := Ping(database); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if err := Close(database); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := Ping(database); err == nil {
		t.Fatal("Ping() after Close() error = nil, want error")
	}
}

func TestNewReturnsErrorWhenPostgresUnavailable(t *testing.T) {
	_, err := New(config.Config{
		DBHost:         "127.0.0.1",
		DBPort:         1,
		DBUser:         "postgres",
		DBName:         "imaiplay",
		DBSSLMode:      "disable",
		DBMaxOpenConns: 25,
		DBMaxIdleConns: 25,
	})
	if err == nil {
		t.Fatal("New() error = nil, want unavailable PostgreSQL error")
	}
}
