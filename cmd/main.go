package main

import (
	"avito-test-quest/internal/app"
	"avito-test-quest/internal/config"
	"avito-test-quest/internal/logger"
	"context"
	"os"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	cfg, err := config.New()

	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Fatal(ctx, "failed to load config", zap.Error(err))
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Fatal(ctx, "failed to initialize application", zap.Error(err))
		os.Exit(1)
	}

	err = application.Run(ctx)
	if err != nil {
		logger.GetOrCreateLoggerFromCtx(ctx).Fatal(ctx, "failed with app", zap.Error(err))
	}
}
