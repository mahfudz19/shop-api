// Package http berisi handler HTTP untuk entitas Promotion
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
)

// PromotionHandler menangani request HTTP untuk entitas Promotion
type PromotionHandler struct {
	usecase domain.PromotionUseCase
}

// NewPromotionHandler membuat instance baru dari PromotionHandler dan mendaftarkan route
func NewPromotionHandler(public gin.IRouter, _ gin.IRouter, us domain.PromotionUseCase) {
	h := &PromotionHandler{usecase: us}

	group := public.Group("/promotions")
	{
		group.GET("", h.GetAll)
		group.GET("/:id", h.GetByID)
		group.POST("", h.Create)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}

// Create menangani request untuk membuat promotion baru
func (h *PromotionHandler) Create(c *gin.Context) {
	var req domain.PromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorValidation(c, err.Error())
		return
	}
	res, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}
	response.SuccessCreated(c, "Promotion created", res)
}

// GetAll menangani request untuk mengambil semua promotion, dengan opsi filter aktif
func (h *PromotionHandler) GetAll(c *gin.Context) {
	// Jika ada query param ?active=true, hanya ambil yang aktif
	activeOnly := c.Query("active") == "true"
	res, err := h.usecase.GetAll(c.Request.Context(), activeOnly)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}
	if res == nil {
		res = []domain.Promotion{}
	}
	response.SuccessSingle(c, "Promotions fetched", res)
}

// GetByID menangani request untuk mengambil promotion berdasarkan ID
func (h *PromotionHandler) GetByID(c *gin.Context) {
	res, err := h.usecase.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorNotFound(c, "Promotion")
		return
	}
	response.SuccessSingle(c, "Promotion fetched", res)
}

// Update menangani request untuk memperbarui promotion berdasarkan ID
func (h *PromotionHandler) Update(c *gin.Context) {
	var req domain.PromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorValidation(c, err.Error())
		return
	}
	res, err := h.usecase.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}
	response.SuccessSingle(c, "Promotion updated", res)
}

// Delete menangani request untuk menghapus promotion berdasarkan ID
func (h *PromotionHandler) Delete(c *gin.Context) {
	if err := h.usecase.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.ErrorInternal(c, err)
		return
	}
	response.SuccessSingle(c, "Promotion deleted", nil)
}
