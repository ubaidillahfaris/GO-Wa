package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/usecases/apikey"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
)

// handleError is a helper function to handle errors consistently in middleware
func handleError(c *gin.Context, err error) {
	if customErr, ok := err.(*errors.CustomError); ok {
		statusCode := http.StatusInternalServerError
		switch customErr.Type {
		case errors.ErrTypeValidation:
			statusCode = http.StatusBadRequest
		case errors.ErrTypeUnauthorized:
			statusCode = http.StatusUnauthorized
		case errors.ErrTypeNotFound:
			statusCode = http.StatusNotFound
		case errors.ErrTypeConflict:
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": customErr.Message})
		c.Abort()
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	c.Abort()
}

const (
	// APIKeyHeader is the header name for API key
	APIKeyHeader = "X-API-Key"

	// ContextKeyAPIKey is the context key for storing API key info
	ContextKeyAPIKey = "api_key"
)

// APIKeyMiddleware creates a middleware that validates API keys
func APIKeyMiddleware(validateUC *apikey.ValidateKeyUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for API key in header
		apiKeyHeader := c.GetHeader(APIKeyHeader)
		if apiKeyHeader == "" {
			// No API key provided - reject request
			handleError(c, errors.New(errors.ErrTypeUnauthorized, "API key is required"))
			return
		}

		// Validate the API key
		key, err := validateUC.Execute(c.Request.Context(), apiKeyHeader)
		if err != nil {
			handleError(c, err)
			return
		}

		// Store API key info in context
		c.Set(ContextKeyAPIKey, key)
		c.Set("username", key.Owner) // For compatibility with existing code

		c.Next()
	}
}

// APIKeyOrJWTMiddleware creates a middleware that accepts either API key or JWT
// This allows gradual migration from JWT-only to supporting both authentication methods
func APIKeyOrJWTMiddleware(validateUC *apikey.ValidateKeyUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for API key first
		apiKeyHeader := c.GetHeader(APIKeyHeader)
		if apiKeyHeader != "" {
			// Validate the API key
			key, err := validateUC.Execute(c.Request.Context(), apiKeyHeader)
			if err != nil {
				handleError(c, err)
				return
			}

			// Store API key info in context
			c.Set(ContextKeyAPIKey, key)
			c.Set("username", key.Owner)
			c.Next()
			return
		}

		// Check for JWT token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			// Use existing JWT middleware logic
			JWTAuthMiddleware()(c)
			return
		}

		// No authentication provided
		handleError(c, errors.New(errors.ErrTypeUnauthorized, "authentication required: provide either X-API-Key or Authorization header"))
	}
}

// APIKeyWithPermissionMiddleware creates a middleware that validates API keys with specific permissions
func APIKeyWithPermissionMiddleware(validateUC *apikey.ValidateKeyUseCase, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for API key in header
		apiKeyHeader := c.GetHeader(APIKeyHeader)
		if apiKeyHeader == "" {
			handleError(c, errors.New(errors.ErrTypeUnauthorized, "API key is required"))
			return
		}

		// Validate the API key with permission check
		key, err := validateUC.ValidateWithPermission(c.Request.Context(), apiKeyHeader, resource, action)
		if err != nil {
			handleError(c, err)
			return
		}

		// Store API key info in context
		c.Set(ContextKeyAPIKey, key)
		c.Set("username", key.Owner)

		c.Next()
	}
}

// GetAPIKeyFromContext retrieves the API key from the Gin context
func GetAPIKeyFromContext(c *gin.Context) (*domain.APIKey, bool) {
	key, exists := c.Get(ContextKeyAPIKey)
	if !exists {
		return nil, false
	}
	apiKey, ok := key.(*domain.APIKey)
	return apiKey, ok
}
