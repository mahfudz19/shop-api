// internal/master_product/delivery/http/master_product_handler.go
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
)

type MasterProductHandler struct {
	usecase domain.MasterProductUseCase
}

func NewMasterProductHandler(r *gin.Engine, us domain.MasterProductUseCase) {
	handler := &MasterProductHandler{usecase: us}

	// Endpoint khusus untuk detail master product
	r.GET("/master-products/:id", handler.GetDetailByID)
}

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