package config

import (
	"avito-test-quest/internal/postgres"
	"fmt"
	"os"

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

	// Ищем config.yaml в нескольких местах для совместимости с тестами и Docker
	configPaths := []string{
		"./configs/config.yaml",
		"configs/config.yaml",
		"../configs/config.yaml",
		"../../configs/config.yaml",
	}

	var configPath string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}

	if configPath == "" {
		return &Config{}, fmt.Errorf("config file not found")
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return &Config{}, fmt.Errorf("error reading config: %w", err)
	}
	return &cfg, nil
}
