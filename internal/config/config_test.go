package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	inDirectory(t, t.TempDir())
	unsetEnvironment(t, "SERVER_PORT", "APP_NAME", "APP_VERSION")

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
	unsetEnvironment(t, "SERVER_PORT", "APP_NAME", "APP_VERSION")

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

func TestLoadDotEnvBesideExecutable(t *testing.T) {
	inDirectory(t, t.TempDir())
	unsetEnvironment(t, "SERVER_PORT", "APP_NAME", "APP_VERSION")

	executableDir := t.TempDir()
	content := []byte("SERVER_PORT=6060\nAPP_NAME=executable-app\nAPP_VERSION=3.0.0\n")
	if err := os.WriteFile(filepath.Join(executableDir, ".env"), content, 0o600); err != nil {
		t.Fatalf("write executable .env: %v", err)
	}

	got, err := load(func() (string, error) {
		return filepath.Join(executableDir, "imaiplay"), nil
	})
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	want := Config{ServerPort: "6060", AppName: "executable-app", AppVersion: "3.0.0"}
	if got != want {
		t.Fatalf("load() = %#v, want %#v", got, want)
	}
}

func TestLoadPrefersCurrentDirectoryDotEnv(t *testing.T) {
	currentDir := t.TempDir()
	inDirectory(t, currentDir)
	unsetEnvironment(t, "SERVER_PORT", "APP_NAME", "APP_VERSION")

	currentContent := []byte("SERVER_PORT=5050\nAPP_NAME=current-app\nAPP_VERSION=4.0.0\n")
	if err := os.WriteFile(filepath.Join(currentDir, ".env"), currentContent, 0o600); err != nil {
		t.Fatalf("write current .env: %v", err)
	}

	executableDir := t.TempDir()
	executableContent := []byte("SERVER_PORT=4040\nAPP_NAME=executable-app\nAPP_VERSION=3.0.0\n")
	if err := os.WriteFile(filepath.Join(executableDir, ".env"), executableContent, 0o600); err != nil {
		t.Fatalf("write executable .env: %v", err)
	}

	got, err := load(func() (string, error) {
		return filepath.Join(executableDir, "imaiplay"), nil
	})
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	want := Config{ServerPort: "5050", AppName: "current-app", AppVersion: "4.0.0"}
	if got != want {
		t.Fatalf("load() = %#v, want %#v", got, want)
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

func unsetEnvironment(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		value, exists := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			if exists {
				if err := os.Setenv(key, value); err != nil {
					t.Errorf("restore %s: %v", key, err)
				}
				return
			}
			if err := os.Unsetenv(key); err != nil {
				t.Errorf("clear %s: %v", key, err)
			}
		})
	}
}
