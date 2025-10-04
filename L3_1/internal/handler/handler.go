package handler

import (
	"github.com/1lostsun/L3/internal/entity"
	"github.com/1lostsun/L3/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"net/http"
)

type Handler struct {
	uc     *usecase.UseCase
	Engine *ginext.Engine
}

func New(uc *usecase.UseCase) *Handler {
	return &Handler{
		uc:     uc,
		Engine: ginext.New(),
	}
}

func (h *Handler) InitRoutes() {
	v1 := h.Engine.Group("/notification/api")
	{
		v1.POST("/notify", h.CreateNotification)
		v1.GET("/notify/:id", h.GetNotification)
		v1.DELETE("/notify/:id", h.CancelNotification)
	}
}

func (h *Handler) CreateNotification(c *ginext.Context) {
	var req entity.Notification

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{Err: err.Error()})
		return
	}

	if err := h.uc.CreateNotification(c, req); err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{Err: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Notification created"})
}

func (h *Handler) GetNotification(c *ginext.Context) {
	id := c.Param("id")

	n, err := h.uc.GetNotification(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{Err: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": n})
}

func (h *Handler) CancelNotification(c *ginext.Context) {
	id := c.Param("id")

	if err := h.uc.CancelNotification(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{Err: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification canceled"})
}
