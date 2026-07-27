package sms

import (
	"context"
	"log/slog"

	"github.com/1622359590/imaiplay/internal/middleware"
)

type LogSender struct{ logger *slog.Logger }

func NewLogSender(logger *slog.Logger) *LogSender { return &LogSender{logger: logger} }

func (sender *LogSender) Send(ctx context.Context, phone, templateCode string, params map[string]string) error {
	sender.logger.InfoContext(ctx, "sms send",
		slog.String("request_id", middleware.RequestIDFromContext(ctx)),
		slog.String("provider", "log"), slog.String("phone", phone),
		slog.String("template_code", templateCode), slog.Any("params", params),
	)
	return nil
}

func NewSender(config Config, logger *slog.Logger) SMSSender {
	if config.Provider == "aliyun" && config.AccessKeyID != "" && config.AccessKeySecret != "" && config.SignName != "" && config.TemplateCode != "" {
		return NewAliyunSender(config)
	}
	return NewLogSender(logger)
}
