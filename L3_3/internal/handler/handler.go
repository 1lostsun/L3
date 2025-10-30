package handler

import (
	"fmt"
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/entity"
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/usecase"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"net/http"
	"strconv"
	"time"
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

	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	v1 := h.Group("/tree_comments/api")
	{
		v1.POST("/comments", h.AddComment)
		v1.GET("/comments", h.GetComments)
		v1.DELETE("/comments/:id", h.DeleteComment)
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
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	sort := c.Query("sort")
	search := c.Query("search")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		fmt.Println(page, sort, search)
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 5
	}

	var parentPtr *string

	if parent != "" {
		parentPtr = &parent
	}

	comments, err := h.uc.GetCommentsTree(parentPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if search != "" {
		comments, err = h.uc.SearchComments(search)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
	}

	if sort != "" {
		comments, err = h.uc.GetPagedComments(page, limit, usecase.SortOrder(sort))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

func (h *Handler) DeleteComment(c *gin.Context) {
	id := c.Param("id")

	if err := h.uc.DeleteCommentsTree(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment and all his child comments was deleted"})
}
