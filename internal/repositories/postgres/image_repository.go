package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	imagedomain "github.com/stannisl/image-processor/internal/domain/image"
)

type ImageRepository struct {
	pool *pgxpool.Pool
}

func NewImageRepository(pool *pgxpool.Pool) *ImageRepository {
	return &ImageRepository{pool}
}

func (r *ImageRepository) GetByID(ctx context.Context, ID uuid.UUID) (*imagedomain.Image, error) {
	const q = `SELECT id, filename, height, width, size, process_type, mime_type, status, updated_at, created_at, processed_at
			   FROM images
			   WHERE id = $1`

	var img imagedomain.Image
	if err := r.pool.QueryRow(ctx, q, ID).Scan(
		&img.ID,
		&img.Filename,
		&img.Height,
		&img.Width,
		&img.Size,
		&img.ProcessType,
		&img.MimeType,
		&img.Status,
		&img.UpdatedAt,
		&img.CreatedAt,
		&img.ProcessedAt,
	); err != nil {
		return nil, fmt.Errorf("GetByID: error querying photo: %w", err)
	}

	return &img, nil
}

func (r *ImageRepository) UpdateImageStatus(ctx context.Context, image *imagedomain.Image) error {
	const q = `UPDATE images SET status = $1 WHERE id = $2` // TODO
	if _, err := r.pool.Exec(ctx, q, image.Status, image.ID); err != nil {
		return fmt.Errorf("UpdateImageStatus: error updating image: %w", err)
	}
	return nil
}

func (r *ImageRepository) CreateImage(ctx context.Context, image *imagedomain.Image) (*imagedomain.Image, error) {
	const q = `INSERT INTO images (filename, height,  width, size, process_type, mime_type)
			   VALUES ($1, $2, $3, $4, $5, $6)
			   RETURNING id, filename, height, width, size, process_type, mime_type, status, updated_at, created_at, processed_at`

	var img imagedomain.Image
	if err := r.pool.QueryRow(ctx, q,
		image.Filename,
		image.Height,
		image.Width,
		image.Size,
		image.ProcessType.String(),
		image.MimeType.String(),
	).
		Scan(
			&img.ID,
			&img.Filename,
			&img.Height,
			&img.Width,
			&img.Size,
			&img.ProcessType,
			&img.MimeType,
			&img.Status,
			&img.UpdatedAt,
			&img.CreatedAt,
			&img.ProcessedAt,
		); err != nil {
		return nil, fmt.Errorf("CreateImage: error creating photo in db: %w", err)
	}

	return &img, nil
}

func (r *ImageRepository) DeleteImage(ctx context.Context, ID uuid.UUID) error {
	const q = `DELETE FROM images WHERE id = $1`
	if _, err := r.pool.Exec(ctx, q, ID); err != nil {
		return fmt.Errorf("DeleteImage: error deleting image: %w", err)
	}
	return nil
}
