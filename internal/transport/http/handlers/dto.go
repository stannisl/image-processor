package handlers

import (
	"time"

	"github.com/google/uuid"
	imagedomain "github.com/stannisl/image-processor/internal/domain/image"
)

type UploadPhotoResponse struct {
	ID uuid.UUID `json:"id"`

	Filename string `json:"filename"`
	Height   int    `json:"height"`
	Width    int    `json:"width"`
	Size     int64  `json:"size"`

	MimeType    string `json:"mime_type"`
	ProcessType string `json:"process_type"`
	Status      string `json:"status"`

	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func UploadResponseFromDomain(d *imagedomain.Image) UploadPhotoResponse {
	return UploadPhotoResponse{
		ID:          d.ID,
		Filename:    d.Filename,
		Height:      d.Height,
		Width:       d.Width,
		Size:        d.Size,
		MimeType:    d.MimeType.String(),
		Status:      d.Status.String(),
		CreatedAt:   d.CreatedAt,
		ProcessType: d.ProcessType.String(),
		UpdatedAt:   d.UpdatedAt,
		ProcessedAt: d.ProcessedAt,
	}
}
