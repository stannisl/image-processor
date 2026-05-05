package image

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	imagedomain "github.com/stannisl/image-processor/internal/domain/image"
	"github.com/stannisl/image-processor/internal/infrastructure/kafka"
	"github.com/stannisl/image-processor/internal/infrastructure/s3"
	imagerepo "github.com/stannisl/image-processor/internal/repositories/postgres"
)

type Processor struct {
	repo     *imagerepo.ImageRepository
	eventBus *kafka.EventBus
	storage  *s3.MinIOStorage
	editor   *Editor
}

func NewProcessor(repo *imagerepo.ImageRepository,
	eventBus *kafka.EventBus,
	storage *s3.MinIOStorage,
	editor *Editor,
) *Processor {
	return &Processor{
		repo:     repo,
		eventBus: eventBus,
		storage:  storage,
		editor:   editor,
	}
}

func (p *Processor) Start(
	ctx context.Context,
	groupID string,
	workers int,
	maxRetries int,
) error {
	p.eventBus.SubscribeUploaded(ctx, groupID, workers, maxRetries, func(
		ctx context.Context, event kafka.ImageUploadedEvent) error {
		return p.processImage(ctx, event)
	})
	return nil
}

func (p *Processor) processImage(ctx context.Context, event kafka.ImageUploadedEvent) error {
	img, err := p.storage.GetOriginal(ctx, event.ImageID)
	if err != nil {
		return fmt.Errorf("processImage: failed to get original: %w", err)
	}

	id, err := uuid.Parse(event.ImageID)
	if err != nil {
		return fmt.Errorf("processImage: failed to parse UUID from event: %w", err)

	}

	imgDomain, err := p.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("processImage: failed to get domain by ID: %w", err)
	}

	if err := imgDomain.ProcessStatus(imagedomain.StatusProcessing); err != nil {
		return fmt.Errorf("processImage: failed to update domain status: %w", err)
	}

	if err := p.repo.UpdateImageStatus(ctx, imgDomain); err != nil {
		return fmt.Errorf("processImage: failed to update image status in db: %w", err)
	}

	processedImg, err := p.editor.ApplyWatermark(img, imgDomain.MimeType)
	if err != nil {
		return fmt.Errorf("processImage: failed to apply watermark: %w", err)
	}

	if err := p.storage.SaveProcessed(ctx, event.ImageID, imgDomain.Filename, processedImg); err != nil {
		return fmt.Errorf("processImage: failed to save processed image: %w", err)
	}

	if err := imgDomain.ProcessStatus(imagedomain.StatusProcessed); err != nil {
		return fmt.Errorf("processImage: failed to update domain status: %w", err)
	}

	if err := p.repo.UpdateImageStatus(ctx, imgDomain); err != nil {
		return fmt.Errorf("processImage: failed to update image status in db: %w", err)
	}

	return nil
}
