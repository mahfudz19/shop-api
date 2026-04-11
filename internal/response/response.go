// Package response = Helper untuk format response API yang konsisten
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BaseResponse = Struct dasar untuk semua response
type BaseResponse struct {
	Success bool        `json:"success"`
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Meta    MetaData    `json:"meta"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
}

// MetaData = Informasi tambahan
type MetaData struct {
	Timestamp  string      `json:"timestamp"`
	RequestID  string      `json:"request_id"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination = Metadata untuk list data
type Pagination struct {
	Page       int64 `json:"page"`
	Limit      int64 `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// ErrorData = Detail error
type ErrorData struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Helper: Generate Request ID
func generateRequestID() string {
	return "req_" + uuid.New().String()[:12]
}

// Helper: Current timestamp ISO8601
func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ============================================
// RESPONSE SUCCESS
// ============================================

// SuccessList = Response untuk data array (dengan pagination)
func SuccessList(c *gin.Context, message string, data interface{}, pagination Pagination) {
	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Status:  http.StatusOK,
		Message: message,
		Meta: MetaData{
			Timestamp:  nowISO(),
			RequestID:  generateRequestID(),
			Pagination: &pagination,
		},
		Data: data,
	})
}

// SuccessSingle = Response untuk single data (by ID)
func SuccessSingle(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Status:  http.StatusOK,
		Message: message,
		Meta: MetaData{
			Timestamp: nowISO(),
			RequestID: generateRequestID(),
		},
		Data: data,
	})
}

// SuccessCreated = Response untuk create success (201)
func SuccessCreated(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, BaseResponse{
		Success: true,
		Status:  http.StatusCreated,
		Message: message,
		Meta: MetaData{
			Timestamp: nowISO(),
			RequestID: generateRequestID(),
		},
		Data: data,
	})
}

// ============================================
// RESPONSE ERROR
// ============================================

// ErrorNotFound = 404
func ErrorNotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, BaseResponse{
		Success: false,
		Status:  http.StatusNotFound,
		Message: resource + " not found",
		Error: &ErrorData{
			Code:    "NOT_FOUND",
			Details: "The requested " + resource + " does not exist",
		},
		Meta: MetaData{
			Timestamp: nowISO(),
			RequestID: generateRequestID(),
		},
	})
}

// ErrorBadRequest = 400
func ErrorBadRequest(c *gin.Context, details string) {
	c.JSON(http.StatusBadRequest, BaseResponse{
		Success: false,
		Status:  http.StatusBadRequest,
		Message: "Bad request",
		Error: &ErrorData{
			Code:    "BAD_REQUEST",
			Details: details,
		},
		Meta: MetaData{
			Timestamp: nowISO(),
			RequestID: generateRequestID(),
		},
	})
}

// ErrorInternal = 500
func ErrorInternal(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, BaseResponse{
		Success: false,
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
		Error: &ErrorData{
			Code:    "INTERNAL_ERROR",
			Details: err.Error(),
		},
		Meta: MetaData{
			Timestamp: nowISO(),
			RequestID: generateRequestID(),
		},
	})
}

// ErrorValidation = 422
func ErrorValidation(c *gin.Context, details string) {
	c.JSON(http.StatusUnprocessableEntity, BaseResponse{
		Success: false,
		Status:  http.StatusUnprocessableEntity,
		Message: "Validation error",
		Error: &ErrorData{
			Code:    "VALIDATION_ERROR",
			Details: details,
		},
		Meta: MetaData{
			Timestamp: nowISO(),
			RequestID: generateRequestID(),
		},
	})
}

// ErrorUnauthorized = 401
func ErrorUnauthorized(c *gin.Context, details string) {
	c.JSON(http.StatusUnauthorized, BaseResponse{
		Success: false,
		Status:  http.StatusUnauthorized,
		Message: "Unauthorized",
		Error: &ErrorData{
			Code:    "UNAUTHORIZED",
			Details: details,
		},
		Meta: MetaData{
			Timestamp: nowISO(),
			RequestID: generateRequestID(),
		},
	})
}
