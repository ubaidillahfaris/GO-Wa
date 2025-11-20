package whatsapp

import (
	"context"
	"fmt"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// DisconnectUseCase handles device disconnection logic
type DisconnectUseCase struct {
	manager domain.WhatsAppManagerInterface
	logger  *logger.Logger
}

// NewDisconnectUseCase creates a new DisconnectUseCase
func NewDisconnectUseCase(manager domain.WhatsAppManagerInterface) *DisconnectUseCase {
	return &DisconnectUseCase{
		manager: manager,
		logger:  logger.New("DisconnectUseCase"),
	}
}

// Execute disconnects a WhatsApp device
func (uc *DisconnectUseCase) Execute(ctx context.Context, deviceName string) error {
	uc.logger.WithField("device", deviceName).Info("Disconnecting device")

	// Get client
	client, exists := uc.manager.GetClient(deviceName)
	if !exists {
		return apperrors.NewNotFoundError(fmt.Sprintf("Device '%s'", deviceName))
	}

	// Check if already disconnected
	if !client.IsConnected() {
		uc.logger.WithField("device", deviceName).Warn("Device already disconnected")
		return nil
	}

	// Disconnect
	if err := client.Disconnect(ctx); err != nil {
		uc.logger.WithField("device", deviceName).Error("Failed to disconnect device: %v", err)
		return apperrors.NewConnectionError("Failed to disconnect device", err)
	}

	uc.logger.WithField("device", deviceName).Success("Device disconnected successfully")
	return nil
}
