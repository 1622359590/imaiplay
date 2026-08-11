package main

import (
	"fmt"
	"log"
	"log/slog"

	_ "github.com/1622359590/imaiplay/docs"
	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/db"
	"github.com/1622359590/imaiplay/internal/migration"
	"github.com/1622359590/imaiplay/internal/server"
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
	if err := config.ValidateRuntimeSecrets(cfg); err != nil {
		return fmt.Errorf("validate runtime secrets: %w", err)
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
	repairedDomains, err := migration.ClearReservedTenantDomain(
		database,
		cfg.AdminHost,
	)
	if err != nil {
		return fmt.Errorf("repair reserved tenant domain: %w", err)
	}
	if repairedDomains > 0 {
		slog.Warn(
			"cleared reserved tenant domain bindings",
			"count",
			repairedDomains,
			"domain",
			cfg.AdminHost,
		)
	}
	repos := newRepositories(database)
	infra, err := newInfrastructure(cfg)
	if err != nil {
		return err
	}
	deps := buildServerDependencies(cfg, database, repos, infra)
	if err := server.Run(
		cfg,
		func() error { return db.Ping(database) },
		deps,
	); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}
