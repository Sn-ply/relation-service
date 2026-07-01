package config

import "github.com/spf13/viper"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL string
}

func Load() (*Config, error) {
	viper.SetDefault("SERVER_PORT", "8083")

	viper.AutomaticEnv()

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			URL: viper.GetString("DATABASE_URL"),
		},
	}

	if cfg.Database.URL == "" {
		cfg.Database.URL = "postgres://snaply:snaply_secret@localhost:5432/relations?sslmode=disable"
	}

	return cfg, nil
}
