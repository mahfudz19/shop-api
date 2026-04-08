package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
)

type ProductHandler struct {
	usecase domain.ProductUseCase
}

func NewProductHandler(r *gin.Engine, us domain.ProductUseCase) {
	handler := &ProductHandler{usecase: us}

	r.GET("/products", handler.FetchAll)
	r.GET("/products/:id", handler.GetByID) // Tambah endpoint by ID
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
