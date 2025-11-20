package apikey

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// UpdateKeyUseCase handles updating API key properties
type UpdateKeyUseCase struct {
	repo   domain.APIKeyRepository
	logger *logger.Logger
}

// NewUpdateKeyUseCase creates a new instance of UpdateKeyUseCase
func NewUpdateKeyUseCase(repo domain.APIKeyRepository, log *logger.Logger) *UpdateKeyUseCase {
	return &UpdateKeyUseCase{
		repo:   repo,
		logger: log.WithPrefix("UpdateKeyUC"),
	}
}

// Execute updates an API key's properties
func (uc *UpdateKeyUseCase) Execute(ctx context.Context, keyID string, owner string, req *domain.UpdateAPIKeyRequest) (*domain.APIKey, error) {
	// Validate inputs
	if keyID == "" {
		return nil, errors.New(errors.ErrTypeValidation, "key ID is required")
	}
	if owner == "" {
		return nil, errors.New(errors.ErrTypeValidation, "owner is required")
	}

	// Retrieve the existing API key
	apiKey, err := uc.repo.GetByID(ctx, keyID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if apiKey.Owner != owner {
		uc.logger.Warn("Unauthorized attempt to update API key", logger.Fields{
			"key_id":       keyID,
			"owner":        owner,
			"actual_owner": apiKey.Owner,
		})
		return nil, errors.New(errors.ErrTypeUnauthorized, "you are not authorized to update this API key")
	}

	// Apply updates
	if req.Name != nil {
		apiKey.Name = *req.Name
	}

	if req.RateLimit != nil {
		apiKey.RateLimit = *req.RateLimit
	}

	if req.Permissions != nil {
		apiKey.Permissions = req.Permissions
	}

	if req.Status != nil {
		// Validate status transition
		if err := uc.validateStatusTransition(apiKey.Status, *req.Status); err != nil {
			return nil, err
		}
		apiKey.Status = *req.Status
	}

	// Update in repository
	if err := uc.repo.Update(ctx, apiKey); err != nil {
		return nil, err
	}

	uc.logger.Info("API key updated successfully", logger.Fields{
		"id":    keyID,
		"owner": owner,
	})

	return apiKey, nil
}

// validateStatusTransition validates if a status transition is allowed
func (uc *UpdateKeyUseCase) validateStatusTransition(current, new domain.APIKeyStatus) error {
	// Cannot change from expired status
	if current == domain.APIKeyStatusExpired {
		return errors.New(errors.ErrTypeValidation, "cannot modify expired API key")
	}

	// Cannot change to expired status manually (only system can do this)
	if new == domain.APIKeyStatusExpired {
		return errors.New(errors.ErrTypeValidation, "cannot manually set key to expired status")
	}

	return nil
}
