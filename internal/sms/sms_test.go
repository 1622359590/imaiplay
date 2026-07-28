package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1622359590/imaiplay/internal/middleware"
)

func TestConfigStoreEncryptsSecretAndSwitchesSender(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sms.json")
	store, err := NewConfigStore(path, "jwt-secret", slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Sender().inner.(*LogSender); !ok {
		t.Fatalf("default sender = %T", store.Sender().inner)
	}
	config := Config{Provider: "aliyun", AccessKeyID: "key", AccessKeySecret: "secret-value", SignName: "ImaiPlay", TemplateCode: "SMS_1"}
	if err := store.Save(config); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "secret-value") {
		t.Fatal("sms secret was stored in plaintext")
	}
	if got := store.Get(); got.AccessKeySecret != "" {
		t.Fatal("sms secret was returned")
	}
	if _, ok := store.Sender().inner.(*AliyunSender); !ok {
		t.Fatalf("configured sender = %T", store.Sender().inner)
	}
	reloaded, err := NewConfigStore(path, "jwt-secret", slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Sender().Config().AccessKeySecret != "secret-value" {
		t.Fatal("encrypted secret did not reload")
	}
}

func TestLogSenderIncludesRequestIDAndCode(t *testing.T) {
	var output bytes.Buffer
	sender := NewLogSender(slog.New(slog.NewJSONHandler(&output, nil)))
	ctx := middleware.WithRequestID(context.Background(), "request-1")
	if err := sender.Send(ctx, "13800138000", "", map[string]string{"code": "123456"}); err != nil {
		t.Fatal(err)
	}
	var record map[string]interface{}
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["request_id"] != "request-1" || record["phone"] != "13800138000" {
		t.Fatalf("log = %v", record)
	}
	if !strings.Contains(output.String(), "123456") {
		t.Fatal("verification code was not logged")
	}
}
