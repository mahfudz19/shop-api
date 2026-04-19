// Package http mengatur routing dan handler untuk API kategori produk
package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
)

// CategoryHandler adalah struct yang menangani HTTP request untuk kategori produk
type CategoryHandler struct {
	usecase domain.CategoryUseCase
}

// NewCategoryHandler membuat instance baru dari CategoryHandler dan mengatur routing untuk kategori
func NewCategoryHandler(public gin.IRouter, admin gin.IRouter, us domain.CategoryUseCase) {
	handler := &CategoryHandler{usecase: us}

	// Grouping routes untuk API categories
	categoryRoutes := public.Group("/categories")
	{
		categoryRoutes.POST("/sync", handler.SyncCategories)

		categoryRoutes.GET("", handler.GetAll)
		categoryRoutes.GET("/:id", handler.GetByID)

		categoryRoutes.POST("", handler.Create)
		categoryRoutes.PUT("/:id", handler.Update)
		categoryRoutes.DELETE("/:id", handler.Delete)
	}
}

// Create = Handler untuk setiap endpoint API kategori produk
func (h *CategoryHandler) Create(c *gin.Context) {
	var req domain.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorValidation(c, err.Error())
		return
	}

	cat, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}

	response.SuccessCreated(c, "Category created successfully", cat)
}

// GetAll = Handler untuk setiap endpoint API kategori produk
func (h *CategoryHandler) GetAll(c *gin.Context) {
	categories, err := h.usecase.GetAll(c.Request.Context())
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}

	if categories == nil {
		categories = []domain.Category{}
	}

	response.SuccessSingle(c, "Success fetch categories", categories)
}

// GetByID = Handler untuk setiap endpoint API kategori produk
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	cat, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "category not found" {
			response.ErrorNotFound(c, "Category")
			return
		}
		response.ErrorInternal(c, err)
		return
	}

	response.SuccessSingle(c, "Success fetch category", cat)
}

// Update = Handler untuk setiap endpoint API kategori produk
func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req domain.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorValidation(c, err.Error())
		return
	}

	cat, err := h.usecase.Update(c.Request.Context(), id, req)
	if err != nil {
		if err.Error() == "category not found" {
			response.ErrorNotFound(c, "Category")
			return
		}
		response.ErrorInternal(c, err)
		return
	}

	response.SuccessSingle(c, "Category updated successfully", cat)
}

// Delete = Handler untuk setiap endpoint API kategori produk
func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.usecase.Delete(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "category not found" {
			response.ErrorNotFound(c, "Category")
			return
		}
		response.ErrorInternal(c, err)
		return
	}

	response.SuccessSingle(c, "Category deleted successfully", nil)
}

// SyncCategories = Handler khusus untuk mensinkronisasi data kategori dari products
func (h *CategoryHandler) SyncCategories(c *gin.Context) {
	count, err := h.usecase.SyncCategories(c.Request.Context())
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}

	message := "Sync completed successfully. No new categories found."
	if count > 0 {
		message = fmt.Sprintf("Sync completed successfully. Added %d new categories.", count)
	}

	response.SuccessSingle(c, message, map[string]int64{
		"inserted_count": count,
	})
}
