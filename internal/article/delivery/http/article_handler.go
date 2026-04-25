// Package http mengimplementasikan handler HTTP untuk artikel
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
)

// ArticleHandler menangani permintaan HTTP terkait artikel
type ArticleHandler struct {
	usecase domain.ArticleUseCase
}

// NewArticleHandler membuat instance baru dari ArticleHandler dan mendaftarkan rute
func NewArticleHandler(public gin.IRouter, _ gin.IRouter, us domain.ArticleUseCase) {
	h := &ArticleHandler{usecase: us}
	group := public.Group("/articles")
	{
		group.GET("", h.GetAll)
		group.GET("/slug/:slug", h.GetBySlug) // URL SEO: /articles/slug/5-tenda-camping-terbaik
		group.POST("", h.Create)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}

// Create Handler untuk setiap endpoint
func (h *ArticleHandler) Create(c *gin.Context) {
	var req domain.ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorValidation(c, err.Error())
		return
	}
	res, err := h.usecase.Create(c.Request.Context(), req)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}
	response.SuccessCreated(c, "Article created", res)
}

// GetAll mendukung query parameter ?published=true untuk hanya mengambil artikel yang sudah dipublikasikan
func (h *ArticleHandler) GetAll(c *gin.Context) {
	onlyPublished := c.Query("published") == "true"
	res, err := h.usecase.GetAll(c.Request.Context(), onlyPublished)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}
	if res == nil {
		res = []domain.Article{}
	}
	response.SuccessSingle(c, "Articles fetched", res)
}

// GetBySlug menggunakan slug sebagai parameter untuk mengambil artikel, lebih SEO-friendly
func (h *ArticleHandler) GetBySlug(c *gin.Context) {
	res, err := h.usecase.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		response.ErrorNotFound(c, "Article")
		return
	}
	response.SuccessSingle(c, "Article fetched", res)
}

// Update menggunakan ID untuk mengidentifikasi artikel yang akan diupdate
func (h *ArticleHandler) Update(c *gin.Context) {
	var req domain.ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorValidation(c, err.Error())
		return
	}
	res, err := h.usecase.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}
	response.SuccessSingle(c, "Article updated", res)
}

// Delete menggunakan ID untuk mengidentifikasi artikel yang akan dihapus
func (h *ArticleHandler) Delete(c *gin.Context) {
	if err := h.usecase.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.ErrorInternal(c, err)
		return
	}
	response.SuccessSingle(c, "Article deleted", nil)
}
