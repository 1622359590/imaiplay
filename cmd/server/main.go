package main

import (
	"fmt"
	"log"

	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/db"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/server"
	"github.com/1622359590/imaiplay/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	database, err := db.New(cfg)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer func() {
		if err := db.Close(database); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	if err := migration.AutoMigrate(database); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	tenantRepo := repository.NewTenantRepository(database)
	userRepo := repository.NewUserRepository(database)
	deps := server.Dependencies{
		AuthService:   service.NewAuthService(userRepo, tenantRepo, cfg.JWTSecret),
		TenantService: service.NewTenantService(tenantRepo),
		UserService:   service.NewUserService(userRepo),
	}
	if err := server.Run(
		cfg,
		func() error { return db.Ping(database) },
		deps,
	); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}
