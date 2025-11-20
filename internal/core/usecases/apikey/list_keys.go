package apikey

import (
	"context"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// ListKeysUseCase handles retrieving API keys for a user
type ListKeysUseCase struct {
	repo   domain.APIKeyRepository
	logger *logger.Logger
}

// NewListKeysUseCase creates a new instance of ListKeysUseCase
func NewListKeysUseCase(repo domain.APIKeyRepository, log *logger.Logger) *ListKeysUseCase {
	return &ListKeysUseCase{
		repo:   repo,
		logger: log.WithPrefix("ListKeysUC"),
	}
}

// ListKeysResponse represents the response for listing API keys
type ListKeysResponse struct {
	Keys   []*domain.APIKey `json:"keys"`
	Total  int64            `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// Execute retrieves all API keys for a user with pagination
func (uc *ListKeysUseCase) Execute(ctx context.Context, owner string, limit, offset int) (*ListKeysResponse, error) {
	// Validate owner
	if owner == "" {
		return nil, errors.New(errors.ErrTypeValidation, "owner is required")
	}

	// Set default limit
	if limit <= 0 {
		limit = 50
	}

	// Retrieve API keys from repository
	keys, total, err := uc.repo.List(ctx, owner, limit, offset)
	if err != nil {
		return nil, err
	}

	uc.logger.Info("Retrieved API keys", logger.Fields{
		"owner":  owner,
		"count":  len(keys),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})

	// Mask the API keys in the response (show only last 8 characters)
	for _, key := range keys {
		if len(key.Key) > 8 {
			key.Key = "..." + key.Key[len(key.Key)-8:]
		}
	}

	return &ListKeysResponse{
		Keys:   keys,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
