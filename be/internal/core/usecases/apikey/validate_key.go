package apikey

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// ValidateKeyUseCase handles API key validation
type ValidateKeyUseCase struct {
	repo   domain.APIKeyRepository
	logger *logger.Logger
}

// NewValidateKeyUseCase creates a new instance of ValidateKeyUseCase
func NewValidateKeyUseCase(repo domain.APIKeyRepository, log *logger.Logger) *ValidateKeyUseCase {
	return &ValidateKeyUseCase{
		repo:   repo,
		logger: log.WithPrefix("ValidateKeyUC"),
	}
}

// Execute validates an API key and returns the associated API key entity
func (uc *ValidateKeyUseCase) Execute(ctx context.Context, key string) (*domain.APIKey, error) {
	// Validate input
	if key == "" {
		return nil, errors.New(errors.ErrTypeValidation, "API key is required")
	}

	// Retrieve API key from repository
	apiKey, err := uc.repo.GetByKey(ctx, key)
	if err != nil {
		// Don't expose that the key doesn't exist for security reasons
		if errors.IsNotFound(err) {
			uc.logger.Warn("Invalid API key attempt")
			return nil, errors.New(errors.ErrTypeUnauthorized, "invalid API key")
		}
		return nil, err
	}

	// Check if the key is active
	if !apiKey.IsActive() {
		uc.logger.Warn("Attempt to use inactive API key", logger.Fields{
			"key_id": apiKey.ID,
			"status": apiKey.Status,
		})

		if apiKey.IsExpired() {
			return nil, errors.New(errors.ErrTypeUnauthorized, "API key has expired")
		}

		return nil, errors.New(errors.ErrTypeUnauthorized, "API key is not active")
	}

	// Update last used timestamp asynchronously (don't block request)
	go func() {
		// Use a new context for background operation
		bgCtx := context.Background()
		if err := uc.repo.UpdateLastUsed(bgCtx, key); err != nil {
			uc.logger.Error("Failed to update last used timestamp", err, logger.Fields{
				"key_id": apiKey.ID,
			})
		}
	}()

	uc.logger.Debug("API key validated successfully", logger.Fields{
		"key_id": apiKey.ID,
		"owner":  apiKey.Owner,
	})

	return apiKey, nil
}

// ValidateWithPermission validates an API key and checks if it has permission for a specific action
func (uc *ValidateKeyUseCase) ValidateWithPermission(ctx context.Context, key, resource, action string) (*domain.APIKey, error) {
	// First, validate the key
	apiKey, err := uc.Execute(ctx, key)
	if err != nil {
		return nil, err
	}

	// Check permission
	if !apiKey.HasPermission(resource, action) {
		uc.logger.Warn("API key lacks required permission", logger.Fields{
			"key_id":   apiKey.ID,
			"resource": resource,
			"action":   action,
		})
		return nil, errors.New(errors.ErrTypeUnauthorized, "insufficient permissions for this operation")
	}

	return apiKey, nil
}
