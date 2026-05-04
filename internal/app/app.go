package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stannisl/image-processor/internal/config"
	"github.com/stannisl/image-processor/internal/infrastructure/postgres"
	"golang.org/x/sync/errgroup"
)

type App struct {
	log         *slog.Logger
	cfg         *config.Config
	server      *http.Server
	diContainer *diContainer
}

func NewApp(ctx context.Context, cfg *config.Config, logger *slog.Logger) *App {
	a := &App{
		cfg:         cfg,
		log:         logger,
		diContainer: newDIContainer(ctx, logger, cfg),
	}

	a.server = &http.Server{
		Handler: a.diContainer.Handler(),
		Addr:    a.cfg.Server.Addr(),
	}

	return a
}

func (a *App) Start(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	errg, ctx := errgroup.WithContext(ctx)

	errg.Go(func() error {
		return a.diContainer.ImageProcessor().Start(ctx, a.cfg.Kafka.Concurrency, a.cfg.Kafka.MaxRetries)
	})

	errg.Go(func() error {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	errg.Go(func() error {
		<-ctx.Done()
		return a.Close(context.Background())
	})

	return errg.Wait()
}

func (a *App) Close(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	if err := a.diContainer.EventBus().Close(); err != nil {
		return err
	}

	if err := postgres.Close(ctx, a.diContainer.DB()); err != nil {
		return err
	}

	return nil
}
