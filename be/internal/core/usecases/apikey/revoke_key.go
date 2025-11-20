package apikey

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// RevokeKeyUseCase handles revoking/deleting API keys
type RevokeKeyUseCase struct {
	repo   domain.APIKeyRepository
	logger *logger.Logger
}

// NewRevokeKeyUseCase creates a new instance of RevokeKeyUseCase
func NewRevokeKeyUseCase(repo domain.APIKeyRepository, log *logger.Logger) *RevokeKeyUseCase {
	return &RevokeKeyUseCase{
		repo:   repo,
		logger: log.WithPrefix("RevokeKeyUC"),
	}
}

// Execute revokes (deletes) an API key
// Only the owner of the API key can revoke it
func (uc *RevokeKeyUseCase) Execute(ctx context.Context, keyID string, owner string) error {
	// Validate inputs
	if keyID == "" {
		return errors.New(errors.ErrTypeValidation, "key ID is required")
	}
	if owner == "" {
		return errors.New(errors.ErrTypeValidation, "owner is required")
	}

	// Retrieve the API key
	apiKey, err := uc.repo.GetByID(ctx, keyID)
	if err != nil {
		return err
	}

	// Verify ownership
	if apiKey.Owner != owner {
		uc.logger.Warn("Unauthorized attempt to revoke API key", logger.Fields{
			"key_id": keyID,
			"owner":  owner,
			"actual_owner": apiKey.Owner,
		})
		return errors.New(errors.ErrTypeUnauthorized, "you are not authorized to revoke this API key")
	}

	// Delete the API key
	if err := uc.repo.Delete(ctx, keyID); err != nil {
		return err
	}

	uc.logger.Info("API key revoked successfully", logger.Fields{
		"id":    keyID,
		"name":  apiKey.Name,
		"owner": owner,
	})

	return nil
}
