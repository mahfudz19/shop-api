// Package http = Delivery layer untuk User
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// UserHandler struct untuk handler
type UserHandler struct {
	usecase domain.UserUseCase
}

// NewUserHandler setup routes
func NewUserHandler(r *gin.Engine, us domain.UserUseCase) {
	handler := &UserHandler{usecase: us}

	// Auth routes
	r.POST("/auth/register", handler.Register)
	r.POST("/auth/login", handler.Login)

	// User routes (protected - nanti bisa tambah middleware)
	r.GET("/users/:id", handler.GetByID)
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
		response.ErrorBadRequest(c, "Invalid email or password")
		return
	}

	response.SuccessSingle(c, "Login successful", user)
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
