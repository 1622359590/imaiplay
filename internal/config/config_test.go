package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	inDirectory(t, t.TempDir())
	t.Setenv("SERVER_PORT", "")
	t.Setenv("APP_NAME", "")
	t.Setenv("APP_VERSION", "")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := Config{ServerPort: "8080", AppName: "imaiplay", AppVersion: "0.1.0"}
	if got != want {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadEnvironment(t *testing.T) {
	inDirectory(t, t.TempDir())
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("APP_NAME", "training")
	t.Setenv("APP_VERSION", "1.2.3")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := Config{ServerPort: "9090", AppName: "training", AppVersion: "1.2.3"}
	if got != want {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	inDirectory(t, dir)
	t.Setenv("SERVER_PORT", "")
	t.Setenv("APP_NAME", "")
	t.Setenv("APP_VERSION", "")

	content := []byte("SERVER_PORT=7070\nAPP_NAME=dotenv-app\nAPP_VERSION=2.0.0\n")
	if err := os.WriteFile(filepath.Join(dir, ".env"), content, 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := Config{ServerPort: "7070", AppName: "dotenv-app", AppVersion: "2.0.0"}
	if got != want {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func inDirectory(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}
