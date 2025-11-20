package whatsapp

import (
	"context"
	"fmt"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// ListGroupsUseCase handles listing groups logic
type ListGroupsUseCase struct {
	manager domain.WhatsAppManagerInterface
	logger  *logger.Logger
}

// NewListGroupsUseCase creates a new ListGroupsUseCase
func NewListGroupsUseCase(manager domain.WhatsAppManagerInterface) *ListGroupsUseCase {
	return &ListGroupsUseCase{
		manager: manager,
		logger:  logger.New("ListGroupsUseCase"),
	}
}

// Execute retrieves groups from a WhatsApp device
func (uc *ListGroupsUseCase) Execute(ctx context.Context, deviceName string) ([]domain.WhatsAppGroup, error) {
	uc.logger.WithField("device", deviceName).Info("Listing groups")

	// Get client
	client, exists := uc.manager.GetClient(deviceName)
	if !exists {
		return nil, apperrors.NewNotFoundError(fmt.Sprintf("Device '%s'", deviceName))
	}

	// Check if connected
	if !client.IsConnected() {
		return nil, apperrors.New(apperrors.ErrorTypeConnection,
			fmt.Sprintf("Device '%s' is not connected", deviceName))
	}

	// Get groups
	groups, err := client.GetGroups(ctx)
	if err != nil {
		uc.logger.WithField("device", deviceName).Error("Failed to get groups: %v", err)
		return nil, apperrors.NewWhatsAppError("Failed to retrieve groups", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"device": deviceName,
		"count":  len(groups),
	}).Success("Groups retrieved")

	return groups, nil
}
