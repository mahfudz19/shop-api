// Package http = Delivery layer untuk Product
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
)

// ProductHandler struct untuk handle HTTP request terkait Product
type ProductHandler struct {
	usecase domain.ProductUseCase
}

// NewProductHandler = Inisialisasi routes untuk Product
func NewProductHandler(r *gin.Engine, us domain.ProductUseCase) {
	handler := &ProductHandler{usecase: us}

	r.GET("/products/deals", handler.GetDeals)
	r.GET("/products/stats", handler.GetStats)
	r.GET("/products", handler.FetchAll)
	r.GET("/product/:id", handler.GetByID)
}

// FetchAll = List dengan pagination
func (h *ProductHandler) FetchAll(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse filter (sama seperti sebelumnya)
	filter := domain.ProductFilter{
		Search:      c.Query("search"),
		Location:    c.Query("location"),
		Marketplace: c.Query("marketplace"),
		SortBy:      c.Query("sort_by"),
		SortOrder:   c.Query("sort_order"),
	}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if val, err := strconv.ParseInt(minPrice, 10, 64); err == nil {
			filter.MinPrice = val
		}
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if val, err := strconv.ParseInt(maxPrice, 10, 64); err == nil {
			filter.MaxPrice = val
		}
	}
	if page := c.Query("page"); page != "" {
		if val, err := strconv.ParseInt(page, 10, 64); err == nil {
			filter.Page = val
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.ParseInt(limit, 10, 64); err == nil {
			filter.Limit = val
		}
	}

	// Panggil usecase
	resp, err := h.usecase.GetProductsWithFilter(ctx, filter)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}

	// Build pagination metadata
	pagination := response.Pagination{
		Page:       resp.Page,
		Limit:      resp.Limit,
		Total:      resp.Total,
		TotalPages: resp.TotalPages,
		HasNext:    resp.Page < resp.TotalPages,
		HasPrev:    resp.Page > 1,
	}

	// Response standar
	response.SuccessList(c, "Products retrieved successfully", resp.Data, pagination)
}

// GetByID = Single data (tambahan)
func (h *ProductHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	product, err := h.usecase.GetProductByID(ctx, id)
	if err != nil {
		response.ErrorNotFound(c, "Product")
		return
	}

	response.SuccessSingle(c, "Product retrieved successfully", product)
}

// GetDeals = Handler untuk Featured Deals
func (h *ProductHandler) GetDeals(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query params untuk limit (opsional, default diatur di usecase)
	var limit int64 = 10
	if limitQuery := c.Query("limit"); limitQuery != "" {
		if val, err := strconv.ParseInt(limitQuery, 10, 64); err == nil {
			limit = val
		}
	}

	products, err := h.usecase.GetDeals(ctx, limit)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}

	response.SuccessSingle(c, "Success fetch featured deals", products)
}

// GetStats = Handler untuk Trust Section
func (h *ProductHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.usecase.GetStats(ctx)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}

	response.SuccessSingle(c, "Success fetch product statistics", stats)
}
