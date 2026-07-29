package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	inDirectory(t, t.TempDir())
	unsetConfigEnvironment(t)

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := defaultConfig()
	if got != want {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadEnvironment(t *testing.T) {
	inDirectory(t, t.TempDir())
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("APP_NAME", "training")
	t.Setenv("APP_VERSION", "1.2.3")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "imaiplay_user")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "training")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("DB_MAX_OPEN_CONNS", "50")
	t.Setenv("DB_MAX_IDLE_CONNS", "10")
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	t.Setenv("STORAGE_DRIVER", "local")
	t.Setenv("STORAGE_LOCAL_ROOT", "/var/lib/imaiplay/uploads")
	t.Setenv("STORAGE_LOCAL_URL", "https://cdn.example.com/uploads")

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := Config{
		ServerPort:       "9090",
		AppName:          "training",
		AppVersion:       "1.2.3",
		DBHost:           "db.internal",
		DBPort:           5433,
		DBUser:           "imaiplay_user",
		DBPassword:       "secret",
		DBName:           "training",
		DBSSLMode:        "require",
		DBMaxOpenConns:   50,
		DBMaxIdleConns:   10,
		JWTSecret:        "test-jwt-secret",
		AllowedOrigins:   DefaultAllowedOrigins,
		AdminHost:        "play.imai.work",
		StorageDriver:    "local",
		StorageLocalRoot: "/var/lib/imaiplay/uploads",
		StorageLocalURL:  "https://cdn.example.com/uploads",
		AuthRateLimit:    10, AuthRateWindowSeconds: 60, LogLevel: "info", LogFormat: "json", SMSConfigFile: ".imaiplay-sms.json", StorageConfigFile: ".imaiplay-storage.json",
	}
	if got != want {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	inDirectory(t, dir)
	unsetConfigEnvironment(t)

	content := []byte(
		"SERVER_PORT=7070\nAPP_NAME=dotenv-app\nAPP_VERSION=2.0.0\n" +
			"DB_HOST=postgres.local\nDB_PORT=5434\nDB_USER=dotenv-user\n" +
			"DB_PASSWORD=dotenv-pass\nDB_NAME=dotenv-db\nDB_SSLMODE=verify-full\n" +
			"DB_MAX_OPEN_CONNS=40\nDB_MAX_IDLE_CONNS=12\nJWT_SECRET=dotenv-secret\n",
	)
	if err := os.WriteFile(filepath.Join(dir, ".env"), content, 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := Config{
		ServerPort:       "7070",
		AppName:          "dotenv-app",
		AppVersion:       "2.0.0",
		DBHost:           "postgres.local",
		DBPort:           5434,
		DBUser:           "dotenv-user",
		DBPassword:       "dotenv-pass",
		DBName:           "dotenv-db",
		DBSSLMode:        "verify-full",
		DBMaxOpenConns:   40,
		DBMaxIdleConns:   12,
		JWTSecret:        "dotenv-secret",
		AllowedOrigins:   DefaultAllowedOrigins,
		AdminHost:        "play.imai.work",
		StorageDriver:    "local",
		StorageLocalRoot: "./uploads",
		StorageLocalURL:  "http://localhost:8080/uploads",
		AuthRateLimit:    10, AuthRateWindowSeconds: 60, LogLevel: "info", LogFormat: "json", SMSConfigFile: ".imaiplay-sms.json", StorageConfigFile: ".imaiplay-storage.json",
	}
	if got != want {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadDotEnvBesideExecutable(t *testing.T) {
	inDirectory(t, t.TempDir())
	unsetConfigEnvironment(t)

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

	want := defaultConfig()
	want.ServerPort = "6060"
	want.AppName = "executable-app"
	want.AppVersion = "3.0.0"
	if got != want {
		t.Fatalf("load() = %#v, want %#v", got, want)
	}
}

func TestLoadPrefersCurrentDirectoryDotEnv(t *testing.T) {
	currentDir := t.TempDir()
	inDirectory(t, currentDir)
	unsetConfigEnvironment(t)

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

	want := defaultConfig()
	want.ServerPort = "5050"
	want.AppName = "current-app"
	want.AppVersion = "4.0.0"
	if got != want {
		t.Fatalf("load() = %#v, want %#v", got, want)
	}
}

func defaultConfig() Config {
	return Config{
		ServerPort:       "8080",
		AppName:          "imaiplay",
		AppVersion:       "0.1.0",
		DBHost:           "localhost",
		DBPort:           5432,
		DBUser:           "postgres",
		DBPassword:       "",
		DBName:           "imaiplay",
		DBSSLMode:        "disable",
		DBMaxOpenConns:   25,
		DBMaxIdleConns:   25,
		JWTSecret:        "imaiplay-dev-secret-change-in-production",
		AllowedOrigins:   DefaultAllowedOrigins,
		AdminHost:        "play.imai.work",
		StorageDriver:    "local",
		StorageLocalRoot: "./uploads",
		StorageLocalURL:  "http://localhost:8080/uploads",
		AuthRateLimit:    10, AuthRateWindowSeconds: 60, LogLevel: "info", LogFormat: "json", SMSConfigFile: ".imaiplay-sms.json", StorageConfigFile: ".imaiplay-storage.json",
	}
}

func unsetConfigEnvironment(t *testing.T) {
	t.Helper()
	unsetEnvironment(t,
		"SERVER_PORT", "APP_NAME", "APP_VERSION",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"DB_SSLMODE", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
		"JWT_SECRET", "ALLOWED_ORIGINS",
		"ADMIN_HOST",
		"STORAGE_DRIVER", "STORAGE_LOCAL_ROOT", "STORAGE_LOCAL_URL",
		"AUTH_RATE_LIMIT", "AUTH_RATE_WINDOW_SECONDS",
		"LOG_LEVEL", "LOG_FORMAT",
		"SMS_CONFIG_FILE",
	)
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
