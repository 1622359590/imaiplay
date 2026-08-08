package main

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/1622359590/imaiplay/internal/baota"
	"github.com/1622359590/imaiplay/internal/config"
	"github.com/1622359590/imaiplay/internal/service"
	"github.com/1622359590/imaiplay/internal/sms"
	"github.com/1622359590/imaiplay/internal/storage"
)

type appInfrastructure struct {
	sms            *sms.ConfigStore
	storage        *storage.Runtime
	domainPanel    service.DomainPanel
	expectedIP     string
	reservedDomain string
	cnameTarget    string
	proxyTarget    string
}

func newInfrastructure(cfg config.Config) (appInfrastructure, error) {
	smsConfig, err := sms.NewConfigStore(cfg.SMSConfigFile, cfg.JWTSecret, slog.Default())
	if err != nil {
		return appInfrastructure{}, fmt.Errorf("initialize sms config: %w", err)
	}
	local, err := storage.NewLocal(storage.LocalConfig{Root: cfg.StorageLocalRoot, URL: cfg.StorageLocalURL})
	if err != nil {
		return appInfrastructure{}, fmt.Errorf("initialize local storage: %w", err)
	}
	store, err := storage.NewConfigStore(cfg.StorageConfigFile, cfg.JWTSecret, storage.Config{
		Driver: cfg.StorageDriver,
		Local:  storage.LocalConfig{Root: cfg.StorageLocalRoot, URL: cfg.StorageLocalURL},
	})
	if err != nil {
		return appInfrastructure{}, fmt.Errorf("initialize storage config: %w", err)
	}
	runtime, err := storage.NewRuntime(local, store, cfg.StorageDriver)
	if err != nil {
		return appInfrastructure{}, fmt.Errorf("initialize storage: %w", err)
	}
	var panel service.DomainPanel
	if strings.TrimSpace(cfg.BaotaPanelURL) != "" && strings.TrimSpace(cfg.BaotaAPIKey) != "" {
		panelURL := strings.TrimSpace(cfg.BaotaPanelURL)
		panel = &baota.Client{
			PanelURL:              panelURL,
			APIKey:                strings.TrimSpace(cfg.BaotaAPIKey),
			TLSInsecureSkipVerify: cfg.BaotaTLSInsecureSkipVerify,
		}
		if cfg.BaotaTLSInsecureSkipVerify {
			slog.Warn(
				"baota TLS certificate verification disabled",
				"panel_url", panelURL,
			)
		}
	}
	return appInfrastructure{
		sms: smsConfig, storage: runtime, domainPanel: panel,
		expectedIP: strings.TrimSpace(cfg.BaotaServerIP), reservedDomain: strings.TrimSpace(cfg.AdminHost),
		cnameTarget: strings.TrimSpace(cfg.AdminHost), proxyTarget: strings.TrimSpace(cfg.BaotaProxyTarget),
	}, nil
}
