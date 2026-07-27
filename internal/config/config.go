package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort     string
	AppName        string
	AppVersion     string
	DBHost         string
	DBPort         int
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	DBMaxOpenConns int
	DBMaxIdleConns int
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
		ServerPort:     v.GetString("SERVER_PORT"),
		AppName:        v.GetString("APP_NAME"),
		AppVersion:     v.GetString("APP_VERSION"),
		DBHost:         v.GetString("DB_HOST"),
		DBPort:         v.GetInt("DB_PORT"),
		DBUser:         v.GetString("DB_USER"),
		DBPassword:     v.GetString("DB_PASSWORD"),
		DBName:         v.GetString("DB_NAME"),
		DBSSLMode:      v.GetString("DB_SSLMODE"),
		DBMaxOpenConns: v.GetInt("DB_MAX_OPEN_CONNS"),
		DBMaxIdleConns: v.GetInt("DB_MAX_IDLE_CONNS"),
	}, nil
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
