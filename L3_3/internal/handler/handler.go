package handler

import (
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/entity"
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"net/http"
)

type Handler struct {
	uc *usecase.UseCase
	*ginext.Engine
}

func New(uc *usecase.UseCase) *Handler {
	return &Handler{
		uc:     uc,
		Engine: ginext.New(""),
	}
}

func (h *Handler) InitRoutes() {
	v1 := h.Group("/tree_comments/api")
	{
		v1.POST("/comments", h.AddComment)
		v1.GET("/comments", h.GetComments)
		//v1.DELETE("/comments/:id ")
	}
}

func (h *Handler) AddComment(c *gin.Context) {
	var req entity.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.uc.AddComment(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func (h *Handler) GetComments(c *gin.Context) {
	parent := c.Param("parent")
	var parentPtr *string

	if parent != "" {
		parentPtr = &parent
	}

	comments, err := h.uc.GetCommentsTree(parentPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}
