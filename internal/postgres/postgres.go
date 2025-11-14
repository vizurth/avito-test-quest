package postgres

import (
	"avito-test-quest/internal/logger"
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"time"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	MaxConns int32  `yaml:"max_conns" env:"MAX_CONNS" env-default:"10"`
	MinConns int32  `yaml:"min_conns" env:"MIN_CONNS" env-default:"5"`
}

func New(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	connString := cfg.GetConnString()
	connString += fmt.Sprintf("&pool_max_conns=%d&pool_min_conns=%d",
		cfg.MaxConns,
		cfg.MinConns,
	)

	conn, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return conn, nil
}

func Migrate(ctx context.Context, cfg Config) error {
	connString := cfg.GetConnString()
	log := logger.GetOrCreateLoggerFromCtx(ctx)

	m, err := migrate.New("file://migrations", connString)

	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	retries := 5
	for i := 0; i < retries; i++ {
		err = m.Up()
		if err == nil {
			break
		}
		log.Info(ctx, "migration failed, retrying...", zap.Error(err))
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.GetOrCreateLoggerFromCtx(ctx).Info(ctx, "migrated successfully")
	return nil
}

func (c *Config) GetConnString() string {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
	return connString
}
