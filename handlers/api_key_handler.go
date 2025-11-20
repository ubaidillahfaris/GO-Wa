package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/core/usecases/apikey"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
)

// APIKeyHandler handles API key management requests
type APIKeyHandler struct {
	generateUC *apikey.GenerateKeyUseCase
	listUC     *apikey.ListKeysUseCase
	revokeUC   *apikey.RevokeKeyUseCase
	updateUC   *apikey.UpdateKeyUseCase
}

// NewAPIKeyHandler creates a new instance of APIKeyHandler
func NewAPIKeyHandler(
	generateUC *apikey.GenerateKeyUseCase,
	listUC *apikey.ListKeysUseCase,
	revokeUC *apikey.RevokeKeyUseCase,
	updateUC *apikey.UpdateKeyUseCase,
) *APIKeyHandler {
	return &APIKeyHandler{
		generateUC: generateUC,
		listUC:     listUC,
		revokeUC:   revokeUC,
		updateUC:   updateUC,
	}
}

// GenerateKey handles POST /api-keys - Generate a new API key
func (h *APIKeyHandler) GenerateKey(c *gin.Context) {
	// Get username from context (set by JWT middleware)
	username, exists := c.Get("username")
	if !exists {
		handleError(c, errors.New(errors.ErrTypeUnauthorized, "user not authenticated"))
		return
	}

	owner := username.(string)

	// Parse request body
	var req domain.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, errors.Wrap(err, errors.ErrTypeValidation, "invalid request body"))
		return
	}

	// Execute use case
	apiKey, err := h.generateUC.Execute(c.Request.Context(), owner, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(201, gin.H{
		"message": "API key generated successfully",
		"data":    apiKey,
	})
}

// ListKeys handles GET /api-keys - List all API keys for the authenticated user
func (h *APIKeyHandler) ListKeys(c *gin.Context) {
	// Get username from context
	username, exists := c.Get("username")
	if !exists {
		handleError(c, errors.New(errors.ErrTypeUnauthorized, "user not authenticated"))
		return
	}

	owner := username.(string)

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Execute use case
	response, err := h.listUC.Execute(c.Request.Context(), owner, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "API keys retrieved successfully",
		"data":    response,
	})
}

// GetKey handles GET /api-keys/:id - Get a specific API key by ID
func (h *APIKeyHandler) GetKey(c *gin.Context) {
	// Get username from context
	username, exists := c.Get("username")
	if !exists {
		handleError(c, errors.New(errors.ErrTypeUnauthorized, "user not authenticated"))
		return
	}

	owner := username.(string)
	keyID := c.Param("id")

	// For now, we'll use the list use case with filtering
	// In a production system, you might want a dedicated GetByID use case
	response, err := h.listUC.Execute(c.Request.Context(), owner, 1, 0)
	if err != nil {
		handleError(c, err)
		return
	}

	// Find the specific key
	var foundKey *domain.APIKey
	for _, key := range response.Keys {
		if key.ID == keyID {
			foundKey = key
			break
		}
	}

	if foundKey == nil {
		handleError(c, errors.New(errors.ErrTypeNotFound, "API key not found"))
		return
	}

	c.JSON(200, gin.H{
		"message": "API key retrieved successfully",
		"data":    foundKey,
	})
}

// UpdateKey handles PUT /api-keys/:id - Update an API key
func (h *APIKeyHandler) UpdateKey(c *gin.Context) {
	// Get username from context
	username, exists := c.Get("username")
	if !exists {
		handleError(c, errors.New(errors.ErrTypeUnauthorized, "user not authenticated"))
		return
	}

	owner := username.(string)
	keyID := c.Param("id")

	// Parse request body
	var req domain.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, errors.Wrap(err, errors.ErrTypeValidation, "invalid request body"))
		return
	}

	// Execute use case
	apiKey, err := h.updateUC.Execute(c.Request.Context(), keyID, owner, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "API key updated successfully",
		"data":    apiKey,
	})
}

// RevokeKey handles DELETE /api-keys/:id - Revoke (delete) an API key
func (h *APIKeyHandler) RevokeKey(c *gin.Context) {
	// Get username from context
	username, exists := c.Get("username")
	if !exists {
		handleError(c, errors.New(errors.ErrTypeUnauthorized, "user not authenticated"))
		return
	}

	owner := username.(string)
	keyID := c.Param("id")

	// Execute use case
	if err := h.revokeUC.Execute(c.Request.Context(), keyID, owner); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "API key revoked successfully",
	})
}
