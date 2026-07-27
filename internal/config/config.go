package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort string
	AppName    string
	AppVersion string
}

func Load() (Config, error) {
	v := viper.New()
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("APP_NAME", "imaiplay")
	v.SetDefault("APP_VERSION", "0.1.0")
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !os.IsNotExist(err) {
			return Config{}, err
		}
	}

	return Config{
		ServerPort: v.GetString("SERVER_PORT"),
		AppName:    v.GetString("APP_NAME"),
		AppVersion: v.GetString("APP_VERSION"),
	}, nil
}
