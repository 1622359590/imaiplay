package sms

import (
	"context"
	"log/slog"
	"sync"
)

type SMSSender interface {
	Send(context.Context, string, string, map[string]string) error
}

type Config struct {
	Provider        string `json:"provider"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"-"`
	SignName        string `json:"sign_name"`
	TemplateCode    string `json:"template_code"`
}

type ConfigurableSender struct {
	mu     sync.RWMutex
	config Config
	logger *slog.Logger
	inner  SMSSender
}

func NewConfigurableSender(config Config, logger *slog.Logger) *ConfigurableSender {
	sender := &ConfigurableSender{logger: logger}
	sender.Update(config)
	return sender
}

func (sender *ConfigurableSender) Update(config Config) {
	sender.mu.Lock()
	defer sender.mu.Unlock()
	sender.config = config
	sender.inner = NewSender(config, sender.logger)
}

func (sender *ConfigurableSender) Config() Config {
	sender.mu.RLock()
	defer sender.mu.RUnlock()
	return sender.config
}

func (sender *ConfigurableSender) Send(ctx context.Context, phone, templateCode string, params map[string]string) error {
	sender.mu.RLock()
	inner, config := sender.inner, sender.config
	sender.mu.RUnlock()
	if templateCode == "" {
		templateCode = config.TemplateCode
	}
	return inner.Send(ctx, phone, templateCode, params)
}
