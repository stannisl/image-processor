package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stannisl/image-processor/internal/transport/http/handlers"
)

type Router interface {
	http.Handler
}

func NewRouter(
	imageHandler *handlers.ImageHandler,
	uiHandler *handlers.UIHandler,
) Router {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.POST("/upload", imageHandler.UploadImage)
	router.GET("/image/:id", imageHandler.DownloadProcessedImage)
	router.HEAD("/image/:id", imageHandler.GetInfoAboutPhoto)
	router.DELETE("/image/:id", imageHandler.DeleteImage)
	router.GET("/image/:id/original", imageHandler.DownloadOriginalImage)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/", uiHandler.ServeUI)

	return router
}
