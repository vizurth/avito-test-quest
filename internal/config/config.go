package config

import (
	"avito-test-quest/internal/postgres"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type PRConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}
type Config struct {
	Postgres postgres.Config `yaml:"postgres"`
	PR       PRConfig        `yaml:"pr"`
}

func New() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig("./configs/config.yaml", &cfg); err != nil {
		return &Config{}, fmt.Errorf("error reading config: %w", err)
	}
	return &cfg, nil
}
