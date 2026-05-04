package image

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"

	imagedomain "github.com/stannisl/image-processor/internal/domain/image"
)

type Editor struct {
	watermark []byte
}

func NewEditor(watermarkPath string) (*Editor, error) {
	if watermarkPath == "" {
		return nil, errors.New("watermarkPath cannot be empty")
	}

	if _, err := os.Stat(watermarkPath); os.IsNotExist(err) {
		return nil, errors.New("watermarkPath does not exist")
	}

	watermark, err := os.ReadFile(watermarkPath)
	if err != nil {
		return nil, err
	}

	return &Editor{
		watermark: watermark,
	}, nil
}

func (e *Editor) ApplyWatermark(original []byte, mimeType imagedomain.MimeType) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(original))
	if err != nil {
		return nil, fmt.Errorf("decode original: %w", err)
	}

	wm, err := png.Decode(bytes.NewReader(e.watermark))
	if err != nil {
		return nil, fmt.Errorf("decode watermark: %w", err)
	}

	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Src)

	wmBounds := wm.Bounds()
	offset := image.Point{
		X: src.Bounds().Max.X - wmBounds.Max.X - 20,
		Y: src.Bounds().Max.Y - wmBounds.Max.Y - 20,
	}
	draw.Draw(dst, wmBounds.Add(offset), wm, image.Point{}, draw.Over)

	var buf bytes.Buffer
	switch mimeType {
	case "image/png":
		err = png.Encode(&buf, dst)
	default:
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90})
	}
	if err != nil {
		return nil, fmt.Errorf("encode result: %w", err)
	}

	return buf.Bytes(), nil
}
