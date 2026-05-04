package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	imagedomain "github.com/stannisl/image-processor/internal/domain/image"
	imageservice "github.com/stannisl/image-processor/internal/service/image"
)

type ImageHandler struct {
	service *imageservice.Service
	log     *slog.Logger
}

func NewImageHandler(service *imageservice.Service, log *slog.Logger) *ImageHandler {
	return &ImageHandler{
		service: service,
		log:     log,
	}
}

func (h *ImageHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file"})
		return
	}

	img, err := h.service.UploadPhoto(c.Request.Context(), imageservice.UploadInput{
		Filename:  header.Filename,
		Data:      data,
		Operation: imagedomain.ProcessTypeWatermark,
	})
	if err != nil {
		h.log.Error("handler: failed to upload photo", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"image": UploadResponseFromDomain(img)})
}

func (h *ImageHandler) DownloadProcessedImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	img, data, err := h.service.GetProcessedPhoto(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, img.Filename))
	c.Data(http.StatusOK, img.MimeType.String(), data)
}

func (h *ImageHandler) GetInfoAboutPhoto(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	img, err := h.service.GetPhotoMetadata(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image": UploadResponseFromDomain(img)})
}

func (h *ImageHandler) DownloadOriginalImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	img, data, err := h.service.GetOriginalPhoto(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, img.Filename))
	c.Data(http.StatusOK, img.MimeType.String(), data)
}

func (h *ImageHandler) DeleteImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteImage(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
