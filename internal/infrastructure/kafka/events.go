package kafka

import "time"

const (
	TopicUploaded  = "image.uploaded"
	TopicProcessed = "image.processed"
	TopicFailed    = "image.failed" // DLQ
)

type ImageUploadedEvent struct {
	ImageID      string            `json:"image_id"`
	OriginalPath string            `json:"original_path"`
	OutputDir    string            `json:"output_dir"`
	Operation    string            `json:"operation"` // ["resize","thumbnail","watermark"]
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

type ImageProcessedEvent struct {
	ImageID          string            `json:"image_id"`
	Result           string            `json:"results"` // {"resize":"/path/...", ...}
	ProcessingTimeMs int64             `json:"processing_time_ms"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	ProcessedAt      time.Time         `json:"processed_at"`
}

type ImageFailedEvent struct {
	ImageID       string             `json:"image_id"`
	Error         string             `json:"error"`
	OriginalEvent ImageUploadedEvent `json:"original_event"`
	Attempt       int                `json:"attempt"`
	FailedAt      time.Time          `json:"failed_at"`
}
