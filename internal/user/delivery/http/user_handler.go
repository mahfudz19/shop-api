// Package http = Delivery layer untuk User
package http

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
	"github.com/username/shop-api/internal/util"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// UserHandler struct untuk handler
type UserHandler struct {
	usecase domain.UserUseCase
}

// NewUserHandler setup routes
func NewUserHandler(public gin.IRouter, protected gin.IRouter, admin gin.IRouter, us domain.UserUseCase) {
	handler := &UserHandler{usecase: us}

	// Auth routes
	public.POST("/auth/register", handler.Register)
	public.POST("/auth/login", handler.Login)

	// Logout dipindahkan ke protectedRoutes agar memerlukan CSRF protection
	protected.POST("/auth/logout", handler.Logout)
	protected.GET("/auth/my", handler.GetMyProfile)

	// User routes (Admin only - dengan CSRF protection)
	admin.GET("/users", handler.GetAll)
	admin.GET("/users/:id", handler.GetByID)
	admin.PUT("/users/:id", handler.Update)
	admin.DELETE("/users/:id", handler.Delete)
}

// RegisterRequest struct untuk request body
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
}

// Register handler
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, err.Error())
		return
	}

	user := domain.User{
		ID:       bson.NewObjectID(),
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Role:     domain.UserRole("user"),
	}

	if err := h.usecase.Register(c.Request.Context(), user); err != nil {
		response.ErrorBadRequest(c, err.Error())
		return
	}

	response.SuccessCreated(c, "User registered successfully", gin.H{
		"id":    user.ID.Hex(),
		"email": user.Email,
		"name":  user.Name,
	})
}

// LoginRequest struct
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login handler
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, err.Error())
		return
	}

	user, err := h.usecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Gunakan ErrorUnauthorized untuk login gagal
		response.ErrorUnauthorized(c, err.Error())
		return
	}

	tokenString, err := util.GenerateToken(user.ID.Hex(), user.Email, user.Role)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}

	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	isSecure := os.Getenv("APP_ENV") == "production"

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", tokenString, 3600*24, "/", cookieDomain, isSecure, true)

	response.SuccessSingle(c, "Login successful", gin.H{
		"id":    user.ID.Hex(),
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	})
}

// Logout handler
func (h *UserHandler) Logout(c *gin.Context) {
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	isSecure := os.Getenv("APP_ENV") == "production"

	c.SetCookie("auth_token", "", -1, "/", cookieDomain, isSecure, true)

	response.SuccessSingle(c, "Logout successful", nil)
}

// GetByID handler
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.usecase.GetUserByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "User")
		return
	}

	response.SuccessSingle(c, "User retrieved successfully", user)
}

// GetMyProfile handler
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(string)

	// Ambil data dari Usecase
	user, err := h.usecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorNotFound(c, "User")
		return
	}

	response.SuccessSingle(c, "Berhasil memuat profil", user)
}

// GetAll handler untuk mendapatkan semua user dengan pagination dan filter
func (h *UserHandler) GetAll(c *gin.Context) {
	// Parse filter dari query params
	filter := domain.UserFilter{
		BaseQuery: domain.BaseQuery{
			Search:    c.Query("search"),
			SortBy:    c.Query("sort_by"),
			SortOrder: c.Query("sort_order"),
		},
	}

	if role := c.Query("role"); role != "" {
		userRole := domain.UserRole(role)
		if !userRole.IsValid() {
			response.ErrorBadRequest(c, "invalid role")
			return
		}
		filter.Role = userRole
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
	resp, err := h.usecase.GetAllUsers(c.Request.Context(), filter)
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
	response.SuccessList(c, "Users retrieved successfully", resp.Data, pagination)
}

// UpdateRequest struct untuk request update
type UpdateRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// Update handler untuk update user
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, err.Error())
		return
	}

	updateReq := domain.UpdateUserRequest{
		Email:  req.Email,
		Name:   req.Name,
		Role:   req.Role,
		Status: req.Status,
	}

	user, err := h.usecase.UpdateUser(c.Request.Context(), id, updateReq)
	if err != nil {
		response.ErrorBadRequest(c, err.Error())
		return
	}

	response.SuccessSingle(c, "User updated successfully", user)
}

// Delete handler untuk hapus user
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.usecase.DeleteUser(c.Request.Context(), id)
	if err != nil {
		response.ErrorBadRequest(c, err.Error())
		return
	}

	response.SuccessSingle(c, "User deleted successfully", nil)
}
