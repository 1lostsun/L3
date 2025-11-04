package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"backend/internal/entity"
	"backend/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"io"
	"net/http"
)

type Handler struct {
	uc     *usecase.UseCase
	Engine *ginext.Engine
}

func New(uc *usecase.UseCase) *Handler {
	return &Handler{
		uc:     uc,
		Engine: ginext.New(""),
	}
}

func (h *Handler) InitRoutes() {
	v1 := h.Engine.Group("/img_compressor/v1")
	{
		v1.POST("/upload", h.Upload)
		v1.GET("/image/:image_id", h.Download)
		v1.DELETE("/image/:image_id", h.Delete)
	}
}

func (h *Handler) Upload(c *ginext.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to open file: %w", err).Error()})
		return
	}
	defer openedFile.Close()

	filedata, err := io.ReadAll(openedFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to read file: %w", err).Error()})
		return
	}

	operationsStr := c.PostForm("operations")
	if operationsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operations is empty"})
		return
	}

	var operations entity.Operations
	if err := json.Unmarshal([]byte(operationsStr), &operations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to unmarshal operations: %w", err).Error()})
		return
	}

	imageID, err := h.uc.Upload(c, filedata, &operations)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": fmt.Errorf("failed to upload image: %w", err).Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image_id": imageID})
}

func (h *Handler) Download(c *ginext.Context) {
	imageID := c.Param("image_id")

	fileData, status, err := h.uc.Download(c, imageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": fmt.Errorf("failed to download image: %w", err).Error(),
				"status": status,
			})
		return
	}

	if status != usecase.Done {
		c.JSON(http.StatusOK, gin.H{"status": status})
		return
	}

	contentType := http.DetectContentType(fileData)

	c.DataFromReader(http.StatusOK,
		int64(len(fileData)),
		contentType,
		bytes.NewReader(fileData),
		map[string]string{"Content-Disposition": fmt.Sprintf("inline; filename=\"%s\"", imageID)},
	)
}

func (h *Handler) Delete(c *ginext.Context) {
	imageID := c.Param("image_id")

	if err := h.uc.Delete(c, imageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to delete image: %w", err).Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("image %s deleted", imageID)})
}
