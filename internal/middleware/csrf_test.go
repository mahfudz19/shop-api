// Package middleware test untuk CSRF middleware
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCSRFProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "GET method should pass without CSRF header",
			method:         http.MethodGet,
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HEAD method should pass without CSRF header",
			method:         http.MethodHead,
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OPTIONS method should pass without CSRF header",
			method:         http.MethodOptions,
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "TRACE method should pass without CSRF header",
			method:         http.MethodTrace,
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST method should pass with X-Requested-With header",
			method: http.MethodPost,
			headers: map[string]string{
				"X-Requested-With": "XMLHttpRequest",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST method should pass with X-CSRF-Token header",
			method: http.MethodPost,
			headers: map[string]string{
				"X-CSRF-Token": "some-csrf-token",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST method should pass with both headers",
			method: http.MethodPost,
			headers: map[string]string{
				"X-Requested-With": "XMLHttpRequest",
				"X-CSRF-Token":     "some-csrf-token",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST method should fail without CSRF headers",
			method:         http.MethodPost,
			headers:        map[string]string{},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "PUT method should fail without CSRF headers",
			method:         http.MethodPut,
			headers:        map[string]string{},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "DELETE method should fail without CSRF headers",
			method:         http.MethodDelete,
			headers:        map[string]string{},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "PATCH method should fail without CSRF headers",
			method:         http.MethodPatch,
			headers:        map[string]string{},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "PUT method should pass with X-Requested-With header",
			method: http.MethodPut,
			headers: map[string]string{
				"X-Requested-With": "XMLHttpRequest",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "DELETE method should pass with X-CSRF-Token header",
			method: http.MethodDelete,
			headers: map[string]string{
				"X-CSRF-Token": "some-token",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "PATCH method should pass with X-Requested-With header",
			method: http.MethodPatch,
			headers: map[string]string{
				"X-Requested-With": "XMLHttpRequest",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			req, _ := http.NewRequest(tt.method, "/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			c.Request = req

			// Call middleware
			CSRFProtection()(c)

			// Assert response status
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
