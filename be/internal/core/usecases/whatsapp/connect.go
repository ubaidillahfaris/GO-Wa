package whatsapp

import (
	"context"
	"fmt"

	"github.com/ubaidillahfaris/whatsapp.git/internal/core/domain"
	apperrors "github.com/ubaidillahfaris/whatsapp.git/internal/pkg/errors"
	"github.com/ubaidillahfaris/whatsapp.git/internal/pkg/logger"
)

// ConnectUseCase handles device connection logic
type ConnectUseCase struct {
	manager domain.WhatsAppManagerInterface
	logger  *logger.Logger
}

// NewConnectUseCase creates a new ConnectUseCase
func NewConnectUseCase(manager domain.WhatsAppManagerInterface) *ConnectUseCase {
	return &ConnectUseCase{
		manager: manager,
		logger:  logger.New("ConnectUseCase"),
	}
}

// Execute connects a WhatsApp device
func (uc *ConnectUseCase) Execute(ctx context.Context, deviceName string) error {
	uc.logger.WithField("device", deviceName).Info("Connecting device")

	// Get client
	client, exists := uc.manager.GetClient(deviceName)
	if !exists {
		return apperrors.NewNotFoundError(fmt.Sprintf("Device '%s'", deviceName))
	}

	// Check if already connected
	if client.IsConnected() {
		uc.logger.WithField("device", deviceName).Warn("Device already connected")
		return apperrors.New(apperrors.ErrorTypeConflict, "Device is already connected")
	}

	// Connect
	if err := client.Connect(ctx); err != nil {
		uc.logger.WithField("device", deviceName).Error("Failed to connect device: %v", err)
		return apperrors.NewConnectionError("Failed to connect device", err)
	}

	uc.logger.WithField("device", deviceName).Success("Device connected successfully")
	return nil
}
