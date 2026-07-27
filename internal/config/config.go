package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort string
	AppName    string
	AppVersion string
}

func Load() (Config, error) {
	return load(os.Executable)
}

func load(executablePath func() (string, error)) (Config, error) {
	v := viper.New()
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("APP_NAME", "imaiplay")
	v.SetDefault("APP_VERSION", "0.1.0")
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
		ServerPort: v.GetString("SERVER_PORT"),
		AppName:    v.GetString("APP_NAME"),
		AppVersion: v.GetString("APP_VERSION"),
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
