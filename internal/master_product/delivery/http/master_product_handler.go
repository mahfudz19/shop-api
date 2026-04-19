// Package http = Delivery layer untuk Master Product
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
)

// MasterProductHandler struct untuk handle HTTP request terkait Master Product
type MasterProductHandler struct {
	usecase domain.MasterProductUseCase
}

// NewMasterProductHandler = Inisialisasi routes untuk Master Product
func NewMasterProductHandler(public gin.IRouter, protected gin.IRouter, us domain.MasterProductUseCase) {
	handler := &MasterProductHandler{usecase: us}

	// Endpoint khusus untuk detail master product
	public.GET("/master-product/:id", handler.GetDetailByID)

	// Rute Protected (Wajib Login)
	protected.GET("/master-product/:id/test", handler.TestAuth)
}

// GetDetailByID = Handler untuk mendapatkan detail master product beserta penawaran terbaiknya
func (h *MasterProductHandler) GetDetailByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	detail, err := h.usecase.GetDetailByID(ctx, id)
	if err != nil {
		response.ErrorNotFound(c, "Master Product")
		return
	}

	response.SuccessSingle(c, "Master product detail retrieved successfully", detail)
}

// TestAuth = Handler testing
func (h *MasterProductHandler) TestAuth(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	response.SuccessSingle(c, "Akses master product berhasil", gin.H{
		"master_product_id": id,
		"accessed_by":       userID,
		"current_role":      role,
	})
}
