package image

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProcessType string

const (
	ProcessTypeResize    ProcessType = "resize"
	ProcessTypeWatermark ProcessType = "watermark"
	ProcessTypeMiniature ProcessType = "miniature"
)

func (p ProcessType) String() string {
	return string(p)
}

type MimeType string

func (m MimeType) String() string {
	return string(m)
}

func ParseMimeType(mimeType string) (MimeType, bool) {
	switch mimeType {
	case "image/png", "image/jpeg", "image/gif":
		return MimeType(mimeType), true
	case "gif":
		return TypeGIF, true
	case "png":
		return TypePNG, true
	case "jpeg", "jpg":
		return TypeJPEG, true
	default:
		return "", false
	}
}

const (
	TypeJPEG MimeType = "image/jpeg"
	TypePNG  MimeType = "image/png"
	TypeGIF  MimeType = "image/gif"
)

type Status string

func (p Status) String() string {
	return string(p)
}

const (
	StatusFailed     Status = "failed"
	StatusNew        Status = "new"
	StatusProcessing Status = "processing"
	StatusProcessed  Status = "processed"
)

var allowedTransition = map[Status][]Status{
	StatusNew:        {StatusProcessing, StatusFailed},
	StatusProcessing: {StatusProcessed, StatusFailed},
}

func (i *Image) ProcessStatus(next Status) error {
	for _, s := range allowedTransition[i.Status] {
		if s == next {
			now := time.Now()
			i.Status = next
			i.UpdatedAt = now

			if s == StatusProcessed {
				i.ProcessedAt = &now
			}
			return nil
		}
	}
	return fmt.Errorf("image: invalid status transition %s -> %s", i.Status, next)
}

type Image struct {
	ID uuid.UUID

	Filename string
	Height   int
	Width    int
	Size     int64

	ProcessType ProcessType
	MimeType    MimeType

	Status      Status
	UpdatedAt   time.Time
	CreatedAt   time.Time
	ProcessedAt *time.Time
}
