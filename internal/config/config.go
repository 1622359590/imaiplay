package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const DefaultJWTSecret = "imaiplay-dev-secret-change-in-production"
const DefaultAllowedOrigins = "http://localhost:5173,http://localhost:5174,http://localhost:5175,http://127.0.0.1:5173,http://127.0.0.1:5174,http://127.0.0.1:5175"

type Config struct {
	ServerPort                 string
	AppName                    string
	AppVersion                 string
	DBHost                     string
	DBPort                     int
	DBUser                     string
	DBPassword                 string
	DBName                     string
	DBSSLMode                  string
	DBMaxOpenConns             int
	DBMaxIdleConns             int
	JWTSecret                  string
	SuperadminBootstrapSecret  string
	AllowedOrigins             string
	AdminHost                  string
	StorageDriver              string
	StorageLocalRoot           string
	StorageLocalURL            string
	BaotaPanelURL              string
	BaotaAPIKey                string
	BaotaServerIP              string
	BaotaProxyTarget           string
	BaotaTLSInsecureSkipVerify bool
	AuthRateLimit              int
	AuthRateWindowSeconds      int
	LogLevel                   string
	LogFormat                  string
	SMSConfigFile              string
	StorageConfigFile          string
	SwaggerEnabled             bool
}

func Load() (Config, error) {
	return load(os.Executable)
}

func load(executablePath func() (string, error)) (Config, error) {
	v := viper.New()
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("APP_NAME", "imaiplay")
	v.SetDefault("APP_VERSION", "0.1.0")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "")
	v.SetDefault("DB_NAME", "imaiplay")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 25)
	v.SetDefault("JWT_SECRET", DefaultJWTSecret)
	v.SetDefault("SUPERADMIN_BOOTSTRAP_SECRET", "")
	v.SetDefault("ALLOWED_ORIGINS", DefaultAllowedOrigins)
	v.SetDefault("ADMIN_HOST", "play.imai.work")
	v.SetDefault("STORAGE_DRIVER", "local")
	v.SetDefault("STORAGE_LOCAL_ROOT", "./uploads")
	v.SetDefault("STORAGE_LOCAL_URL", "http://localhost:8080/uploads")
	v.SetDefault("BAOTA_PANEL_URL", "")
	v.SetDefault("BAOTA_API_KEY", "")
	v.SetDefault("BAOTA_SERVER_IP", "")
	v.SetDefault("BAOTA_PROXY_TARGET", "http://127.0.0.1:18080")
	v.SetDefault("BAOTA_TLS_INSECURE_SKIP_VERIFY", false)
	v.SetDefault("AUTH_RATE_LIMIT", 10)
	v.SetDefault("AUTH_RATE_WINDOW_SECONDS", 60)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
	v.SetDefault("SMS_CONFIG_FILE", ".imaiplay-sms.json")
	v.SetDefault("STORAGE_CONFIG_FILE", ".imaiplay-storage.json")
	v.SetDefault("SWAGGER_ENABLED", true)
	v.SetConfigType("env")
	v.AutomaticEnv()

	loaded, err := readConfig(v, ".env")
	if err != nil {
		return Config{}, err
	}
	if !loaded {
		path, err := executablePath()
		if err != nil {
			return Config{}, fmt.Errorf("resolve executable path: %w", err)
		}
		if _, err := readConfig(v, filepath.Join(filepath.Dir(path), ".env")); err != nil {
			return Config{}, err
		}
	}

	return Config{
		ServerPort:                 v.GetString("SERVER_PORT"),
		AppName:                    v.GetString("APP_NAME"),
		AppVersion:                 v.GetString("APP_VERSION"),
		DBHost:                     v.GetString("DB_HOST"),
		DBPort:                     v.GetInt("DB_PORT"),
		DBUser:                     v.GetString("DB_USER"),
		DBPassword:                 v.GetString("DB_PASSWORD"),
		DBName:                     v.GetString("DB_NAME"),
		DBSSLMode:                  v.GetString("DB_SSLMODE"),
		DBMaxOpenConns:             v.GetInt("DB_MAX_OPEN_CONNS"),
		DBMaxIdleConns:             v.GetInt("DB_MAX_IDLE_CONNS"),
		JWTSecret:                  v.GetString("JWT_SECRET"),
		SuperadminBootstrapSecret:  v.GetString("SUPERADMIN_BOOTSTRAP_SECRET"),
		AllowedOrigins:             v.GetString("ALLOWED_ORIGINS"),
		AdminHost:                  v.GetString("ADMIN_HOST"),
		StorageDriver:              v.GetString("STORAGE_DRIVER"),
		StorageLocalRoot:           v.GetString("STORAGE_LOCAL_ROOT"),
		StorageLocalURL:            v.GetString("STORAGE_LOCAL_URL"),
		BaotaPanelURL:              v.GetString("BAOTA_PANEL_URL"),
		BaotaAPIKey:                v.GetString("BAOTA_API_KEY"),
		BaotaServerIP:              v.GetString("BAOTA_SERVER_IP"),
		BaotaProxyTarget:           v.GetString("BAOTA_PROXY_TARGET"),
		BaotaTLSInsecureSkipVerify: v.GetBool("BAOTA_TLS_INSECURE_SKIP_VERIFY"),
		AuthRateLimit:              v.GetInt("AUTH_RATE_LIMIT"),
		AuthRateWindowSeconds:      v.GetInt("AUTH_RATE_WINDOW_SECONDS"),
		LogLevel:                   v.GetString("LOG_LEVEL"), LogFormat: v.GetString("LOG_FORMAT"),
		SMSConfigFile: v.GetString("SMS_CONFIG_FILE"), StorageConfigFile: v.GetString("STORAGE_CONFIG_FILE"),
		SwaggerEnabled: v.GetBool("SWAGGER_ENABLED"),
	}, nil
}

func ValidateRuntimeSecrets(cfg Config) error {
	if strings.TrimSpace(cfg.DBPassword) == "" {
		return errors.New("DB_PASSWORD must be configured")
	}
	jwtSecret := strings.TrimSpace(cfg.JWTSecret)
	if jwtSecret == DefaultJWTSecret || len(jwtSecret) < 32 {
		return errors.New("JWT_SECRET must contain at least 32 characters and must not use the development default")
	}
	bootstrapSecret := strings.TrimSpace(cfg.SuperadminBootstrapSecret)
	if bootstrapSecret != "" && len(bootstrapSecret) < 32 {
		return errors.New("SUPERADMIN_BOOTSTRAP_SECRET must contain at least 32 characters when enabled")
	}
	return nil
}

func readConfig(v *viper.Viper, path string) (bool, error) {
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) || os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
