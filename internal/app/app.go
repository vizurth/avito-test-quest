package app

import (
	"avito-test-quest/internal/config"
	"avito-test-quest/internal/handler"
	"avito-test-quest/internal/logger"
	"avito-test-quest/internal/postgres"
	"avito-test-quest/internal/repository"
	"avito-test-quest/internal/service"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"net/http"
)

// App представляет основное приложение
type App struct {
	config *config.Config
	log    *logger.Logger
	pool   *pgxpool.Pool
	server *http.Server
}

// New инициализирует приложение
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	ctx, _, err := logger.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	log := logger.GetLoggerFromCtx(ctx)

	pool, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to init postgres: %w", err)
	}

	err = postgres.Migrate(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate postgres: %w", err)
	}

	prRepo := repository.NewPrRepository(pool)
	prService := service.NewPrService(prRepo)

	router := gin.Default()

	httpHandler := handler.NewPrHandler(prService, router)

	httpHandler.InitRoutes()

	srv := &http.Server{
		Addr:              ":" + cfg.PR.Port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,  // общее время на чтение запроса
		ReadHeaderTimeout: 5 * time.Second,   // время на чтение заголовков
		WriteTimeout:      30 * time.Second,  // время на запись ответа
		IdleTimeout:       120 * time.Second, // время простоя соединения
	}

	return &App{
		config: cfg,
		log:    log,
		pool:   pool,
		server: srv,
	}, nil
}

// Shutdown выполняет graceful shutdown приложения
func (a *App) Shutdown(ctx context.Context) error {
	a.log.Info(ctx, "starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.log.Error(ctx, "failed to shutdown HTTP server", zap.Error(err))
		return fmt.Errorf("http server shutdown failed: %w", err)
	}
	a.log.Info(ctx, "HTTP server shutdown successfully")

	a.pool.Close()
	a.log.Info(ctx, "database pool closed successfully")

	a.log.Info(ctx, "graceful shutdown completed")
	return nil
}

// Run запускает HTTP сервер
func (a *App) Run(ctx context.Context) error {
	a.log.Info(ctx, "starting HTTP server", zap.String("addr", fmt.Sprintf(":%s", a.config.PR.Port)))

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error(ctx, "HTTP server error", zap.Error(err))
		}
	}()

	a.log.Info(ctx, "HTTP server started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	a.log.Info(ctx, "shutdown signal received")

	return a.Shutdown(ctx)
}
