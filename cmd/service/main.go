package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/stannisl/image-processor/internal/app"
	"github.com/stannisl/image-processor/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("fail to load config file", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	application := app.NewApp(ctx, cfg, logger)

	if err := application.Start(ctx); err != nil {
		logger.Error("fail to start application", "err", err)
		os.Exit(1)
	}
}
