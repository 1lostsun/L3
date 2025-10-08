package handler

import (
	"github.com/1lostsun/L3/tree/main/L3_2/internal/entity"
	"github.com/1lostsun/L3/tree/main/L3_2/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"net/http"
)

type Handler struct {
	uc *usecase.UseCase
	*ginext.Engine
}

func NewHandler(uc *usecase.UseCase) *Handler {
	return &Handler{
		uc:     uc,
		Engine: ginext.New(),
	}
}

func (h *Handler) InitRoutes() {
	v1 := h.Group("/short_url/api")
	{
		v1.POST("/shorten", h.CreateShortURL)
		v1.GET("/s/:short_url", h.GetShortURL)
		v1.GET("analytics/:short_url", h.GetAnalytics)
	}
}

func (h *Handler) CreateShortURL(c *gin.Context) {
	var req entity.Request

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortURL, err := h.uc.CreateShortURL(c, req.OriginalURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shortUrl": shortURL})
}

func (h *Handler) GetShortURL(c *gin.Context) {
	shortURL := c.Param("short_url")

	link, err := h.uc.GetShortURL(c, shortURL, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, link.OriginalURL)
}

func (h *Handler) GetAnalytics(c *gin.Context) {
	shortURL := c.Param("short_url")

	analytics, err := h.uc.GetAnalytics(c, shortURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}
