// Package http = Delivery layer untuk User
package http

import (
	"os"

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
func NewUserHandler(public gin.IRouter, protected gin.IRouter, us domain.UserUseCase) {
	handler := &UserHandler{usecase: us}

	// Auth routes
	public.POST("/auth/register", handler.Register)
	public.POST("/auth/login", handler.Login)
	public.POST("/auth/logout", handler.Logout)

	// User routes
	public.GET("/users/:id", handler.GetByID)

	// Rute Protected (Wajib Login)
	protected.GET("/auth/my", handler.GetMyProfile)
}

// RegisterRequest struct untuk request body
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`
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
		Role:     domain.UserRole(req.Role),
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
		response.ErrorUnauthorized(c, "Invalid email or password")
		return
	}

	tokenString, err := util.GenerateToken(user.ID.Hex(), user.Email, user.Role)
	if err != nil {
		response.ErrorInternal(c, err)
		return
	}

	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	isSecure := os.Getenv("APP_ENV") == "production"

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
