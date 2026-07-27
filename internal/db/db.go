package db

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/1622359590/imaiplay/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(cfg config.Config) (*gorm.DB, error) {
	database, err := gorm.Open(
		postgres.Open(connectionString(cfg)),
		&gorm.Config{DisableAutomaticPing: true, TranslateError: true},
	)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}

	sqlDatabase, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("get SQL database: %w", err)
	}
	sqlDatabase.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDatabase.SetMaxIdleConns(cfg.DBMaxIdleConns)

	if err := sqlDatabase.Ping(); err != nil {
		_ = sqlDatabase.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}
	return database, nil
}

func Ping(database *gorm.DB) error {
	sqlDatabase, err := database.DB()
	if err != nil {
		return fmt.Errorf("get SQL database: %w", err)
	}
	return sqlDatabase.Ping()
}

func Close(database *gorm.DB) error {
	sqlDatabase, err := database.DB()
	if err != nil {
		return fmt.Errorf("get SQL database: %w", err)
	}
	return sqlDatabase.Close()
}

func connectionString(cfg config.Config) string {
	user := url.User(cfg.DBUser)
	if cfg.DBPassword != "" {
		user = url.UserPassword(cfg.DBUser, cfg.DBPassword)
	}
	connectionURL := url.URL{
		Scheme: "postgres",
		User:   user,
		Host:   net.JoinHostPort(cfg.DBHost, strconv.Itoa(cfg.DBPort)),
		Path:   cfg.DBName,
	}
	query := connectionURL.Query()
	query.Set("sslmode", cfg.DBSSLMode)
	connectionURL.RawQuery = query.Encode()
	return connectionURL.String()
}
