package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL string
}

type KafkaConfig struct {
	Brokers []string
}

func Load() (*Config, error) {
	viper.SetDefault("SERVER_PORT", "8083")
	viper.SetDefault("KAFKA_BROKERS", "localhost:29092")

	viper.AutomaticEnv()

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			URL: viper.GetString("DATABASE_URL"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(viper.GetString("KAFKA_BROKERS"), ","),
		},
	}

	if cfg.Database.URL == "" {
		cfg.Database.URL = "postgres://snaply:snaply_secret@localhost:5432/relations?sslmode=disable"
	}

	return cfg, nil
}
