package image

import (
	"bytes"
	"context"
	"fmt"
	"image"

	"github.com/google/uuid"
	imagedomain "github.com/stannisl/image-processor/internal/domain/image"
	"github.com/stannisl/image-processor/internal/infrastructure/kafka"
	"github.com/stannisl/image-processor/internal/infrastructure/s3"
	imagerepo "github.com/stannisl/image-processor/internal/repositories/postgres"
)

type Service struct {
	repo     *imagerepo.ImageRepository
	eventBus *kafka.EventBus
	storage  *s3.MinIOStorage
}

func NewService(repo *imagerepo.ImageRepository, eventBus *kafka.EventBus, storage *s3.MinIOStorage) *Service {
	return &Service{
		repo:     repo,
		eventBus: eventBus,
		storage:  storage,
	}
}

type UploadInput struct {
	Filename  string
	Data      []byte
	Operation imagedomain.ProcessType
}

func (s *Service) UploadPhoto(ctx context.Context, input UploadInput) (*imagedomain.Image, error) {
	id := uuid.New()

	cfg, format, err := image.DecodeConfig(bytes.NewReader(input.Data))
	if err != nil {
		return nil, fmt.Errorf("UploadPhoto: decode image config: %w", err)
	}

	if err := s.storage.SaveOriginal(ctx, id.String(), input.Filename, input.Data); err != nil {
		return nil, err
	}

	mime, ok := imagedomain.ParseMimeType(format)
	if !ok {
		return nil, fmt.Errorf("UploadPhoto: unknown format")
	}

	img := &imagedomain.Image{
		ID:       id,
		Filename: input.Filename,
		Height:   cfg.Height,
		Width:    cfg.Width,
		Size:     int64(len(input.Data)),
		MimeType: mime,
	}

	img, err = s.repo.CreateImage(ctx, img)
	if err != nil {
		return nil, err
	}

	if err := s.eventBus.PublishUploaded(ctx, kafka.ImageUploadedEvent{
		ImageID:      img.ID.String(),
		OriginalPath: img.ID.String(),
		OutputDir:    img.ID.String(),
		Operation:    img.ProcessType.String(),
	}); err != nil {
		return nil, fmt.Errorf("UploadPhoto: publish uploaded event: %w", err)
	}

	return img, nil
}

func (s *Service) GetProcessedPhoto(ctx context.Context, ID uuid.UUID) (*imagedomain.Image, []byte, error) {
	img, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, nil, fmt.Errorf("GetPhoto: getting from db: %w", err)
	}

	content, err := s.storage.GetProcessed(ctx, img.ID.String())
	if err != nil {
		return nil, nil, fmt.Errorf("GetPhoto: getting image from storage: %w", err)
	}

	return img, content, nil
}

func (s *Service) GetOriginalPhoto(ctx context.Context, ID uuid.UUID) (*imagedomain.Image, []byte, error) {
	img, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, nil, fmt.Errorf("GetPhoto: getting from db: %w", err)
	}

	content, err := s.storage.GetOriginal(ctx, img.ID.String())
	if err != nil {
		return nil, nil, fmt.Errorf("GetPhoto: getting image from storage: %w", err)
	}

	return img, content, nil
}

func (s *Service) DeleteImage(ctx context.Context, ID uuid.UUID) error {
	if err := s.repo.DeleteImage(ctx, ID); err != nil {
		return fmt.Errorf("DeleteImage: deleting image: %w", err)
	}
	return nil
}
