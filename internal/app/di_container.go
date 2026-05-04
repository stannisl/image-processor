package app

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stannisl/image-processor/internal/config"
	"github.com/stannisl/image-processor/internal/infrastructure/kafka"
	"github.com/stannisl/image-processor/internal/infrastructure/postgres"
	"github.com/stannisl/image-processor/internal/infrastructure/s3"
	imagerepo "github.com/stannisl/image-processor/internal/repositories/postgres"
	imageservice "github.com/stannisl/image-processor/internal/service/image"
	httptransport "github.com/stannisl/image-processor/internal/transport/http"
	"github.com/stannisl/image-processor/internal/transport/http/handlers"
)

type lazy[T any] struct {
	once  sync.Once
	value T
}

func (l *lazy[T]) Init(init func() T) T {
	l.once.Do(func() {
		l.value = init()
	})
	return l.value
}

type diContainer struct {
	ctx context.Context
	log *slog.Logger
	cfg *config.Config

	db       lazy[*pgxpool.Pool]
	eventBus lazy[*kafka.EventBus]
	storage  lazy[*s3.MinIOStorage]

	imageRepository lazy[*imagerepo.ImageRepository]

	imageService   lazy[*imageservice.Service]
	imageEditor    lazy[*imageservice.Editor]
	imageProcessor lazy[*imageservice.Processor]

	imageHandler lazy[*handlers.ImageHandler]
	uiHandler    lazy[*handlers.UIHandler]

	router lazy[httptransport.Router]
}

func newDIContainer(ctx context.Context, log *slog.Logger, cfg *config.Config) *diContainer {
	return &diContainer{
		ctx: ctx,
		log: log,
		cfg: cfg,
	}
}

func (d *diContainer) DB() *pgxpool.Pool {
	return d.db.Init(func() *pgxpool.Pool {
		pool, err := postgres.Open(d.cfg.Postgres.DSN())
		if err != nil {
			d.log.Error("Error opening database connection", "error", err)
			panic(err)
		}
		return pool
	})
}

func (d *diContainer) ImageProcessor() *imageservice.Processor {
	return d.imageProcessor.Init(func() *imageservice.Processor {
		return imageservice.NewProcessor(d.ImageRepository(), d.EventBus(), d.Storage(), d.ImageEditor())
	})
}

func (d *diContainer) ImageEditor() *imageservice.Editor {
	return d.imageEditor.Init(func() *imageservice.Editor {
		e, err := imageservice.NewEditor(d.cfg.Image.WatermarkPath)
		if err != nil {
			d.log.Error("Error opening image editor", "error", err)
			panic(err)
		}
		return e
	})
}

func (d *diContainer) EventBus() *kafka.EventBus {
	return d.eventBus.Init(func() *kafka.EventBus {
		client, err := kafka.NewClient(d.cfg, d.log)
		if err != nil {
			d.log.Error("Error creating event bus", "error", err)
			panic(err)
		}
		return kafka.NewEventBus(client, d.log)
	})
}

func (d *diContainer) Storage() *s3.MinIOStorage {
	return d.storage.Init(func() *s3.MinIOStorage {
		c, err := s3.NewMinIO(
			d.log,
			d.cfg.Storage.Endpoint,
			d.cfg.Storage.AccessKey,
			d.cfg.Storage.SecretKey,
		)
		if err != nil {
			d.log.Error("Error creating minio storage", "error", err)
			panic(err)
		}
		return c
	})
}

func (d *diContainer) ImageRepository() *imagerepo.ImageRepository {
	return d.imageRepository.Init(func() *imagerepo.ImageRepository {
		return imagerepo.NewImageRepository(d.DB())
	})
}

func (d *diContainer) ImageService() *imageservice.Service {
	return d.imageService.Init(func() *imageservice.Service {
		return imageservice.NewService(d.ImageRepository(), d.EventBus(), d.Storage())
	})
}

func (d *diContainer) ImageHandler() *handlers.ImageHandler {
	return d.imageHandler.Init(func() *handlers.ImageHandler {
		return handlers.NewImageHandler(d.ImageService(), d.log)
	})
}

func (d *diContainer) Handler() http.Handler {
	return d.router.Init(func() httptransport.Router {
		return httptransport.NewRouter(
			d.ImageHandler(),
			d.UIHandler(),
		)
	})
}

func (d *diContainer) UIHandler() *handlers.UIHandler {
	return d.uiHandler.Init(func() *handlers.UIHandler {
		return &handlers.UIHandler{}
	})
}
