package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigStoreEncryptsSecretAndHidesIt(t *testing.T) {
	file := filepath.Join(t.TempDir(), "storage.json")
	store, err := NewConfigStore(file, "jwt-secret")
	if err != nil {
		t.Fatal(err)
	}
	config := Config{Driver: "s3", S3: S3Config{Endpoint: "http://minio:9000", Bucket: "training", AccessKey: "key", SecretKey: "secret", Region: "cn-test-1", Prefix: "uploads"}}
	if err := store.Save(config); err != nil {
		t.Fatal(err)
	}
	if got := store.Get().S3.SecretKey; got != "" {
		t.Fatalf("secret was returned: %q", got)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"secret_key": "secret"`) {
		t.Fatalf("secret was persisted in plaintext: %s", data)
	}
	reloaded, err := NewConfigStore(file, "jwt-secret")
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.internal().S3.SecretKey; got != "secret" {
		t.Fatalf("decrypted secret = %q", got)
	}
}

func TestS3TestConnectionSignsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodHead || request.Header.Get("Authorization") == "" || request.Header.Get("x-amz-date") == "" {
			t.Fatalf("request was not signed: %#v", request)
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	s3, err := NewS3(S3Config{Endpoint: server.URL, Bucket: "bucket", AccessKey: "access", SecretKey: "secret", Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s3.Test(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestLocalGet(t *testing.T) {
	local, err := NewLocal(LocalConfig{Root: t.TempDir(), URL: "http://localhost/uploads"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := local.Put(context.Background(), "tenant/file.txt", strings.NewReader("data"), 4); err != nil {
		t.Fatal(err)
	}
	body, err := local.Get(context.Background(), "tenant/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	buffer := make([]byte, 4)
	if _, err := body.Read(buffer); err != nil {
		t.Fatal(err)
	}
	if string(buffer) != "data" {
		t.Fatalf("body = %q", buffer)
	}
}
