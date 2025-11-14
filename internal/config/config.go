package config

import (
	"avito-test-quest/internal/postgres"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// PRConfig содержит настройки для сервера PR
type PRConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

// Config содержит общие настройки приложения
type Config struct {
	Postgres postgres.Config `yaml:"postgres"`
	PR       PRConfig        `yaml:"pr"`
}

// New загружает конфигурацию из файла и возвращает Config
func New() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig("./configs/config.yaml", &cfg); err != nil {
		return &Config{}, fmt.Errorf("error reading config: %w", err)
	}
	return &cfg, nil
}
