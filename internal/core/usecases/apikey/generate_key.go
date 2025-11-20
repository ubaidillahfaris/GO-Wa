package apikey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// GenerateKeyUseCase handles the generation of new API keys
type GenerateKeyUseCase struct {
	repo   domain.APIKeyRepository
	logger *logger.Logger
}

// NewGenerateKeyUseCase creates a new instance of GenerateKeyUseCase
func NewGenerateKeyUseCase(repo domain.APIKeyRepository, log *logger.Logger) *GenerateKeyUseCase {
	return &GenerateKeyUseCase{
		repo:   repo,
		logger: log.WithPrefix("GenerateKeyUC"),
	}
}

// Execute generates a new API key for a user
func (uc *GenerateKeyUseCase) Execute(ctx context.Context, owner string, req *domain.CreateAPIKeyRequest) (*domain.APIKey, error) {
	// Validate owner
	if owner == "" {
		return nil, errors.New(errors.ErrTypeValidation, "owner is required")
	}

	// Generate a secure random API key
	key, err := generateSecureKey(64) // 64 bytes = 128 hex characters
	if err != nil {
		uc.logger.Error("Failed to generate API key", err)
		return nil, errors.Wrap(err, errors.ErrTypeInternal, "failed to generate API key")
	}

	// Calculate expiration time
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		expiry := time.Now().AddDate(0, 0, req.ExpiresIn)
		expiresAt = &expiry
	}

	// Set default permissions if not provided
	permissions := req.Permissions
	if len(permissions) == 0 {
		// Default: full access to all resources
		permissions = []domain.APIKeyPermission{
			{
				Resource: "*",
				Actions:  []string{"*"},
			},
		}
	}

	// Create API key entity
	apiKey := &domain.APIKey{
		Key:         key,
		Name:        req.Name,
		Owner:       owner,
		Permissions: permissions,
		Status:      domain.APIKeyStatusActive,
		RateLimit:   req.RateLimit,
		ExpiresAt:   expiresAt,
	}

	// Save to repository
	if err := uc.repo.Create(ctx, apiKey); err != nil {
		return nil, err
	}

	uc.logger.Info("API key generated successfully", logger.Fields{
		"id":    apiKey.ID,
		"name":  apiKey.Name,
		"owner": owner,
	})

	return apiKey, nil
}

// generateSecureKey generates a cryptographically secure random key
func generateSecureKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
